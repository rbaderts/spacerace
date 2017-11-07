/**
 * A Player is the runtime representation of a single user playing a game
 *   A Player encapsulates a websocket connection to a browser window
 *      displaying the game
 *
 */
package core

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"

	"math/rand"
	_ "os"
	"strconv"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type HumanPlayer struct {
	Name        string
	PlayerId    int
	game        *Game
	Ship        *Sprite
	ws          *websocket.Conn
	SendUpdates chan ServerMessage
	mutex       *sync.Mutex

	// map of item type name to PlayerInventoryRecord
	Inventory map[PlayerResourceType]*PlayerInventory
	UserId    int
	actionId  int
	Random    *rand.Rand
}

//func NewPlayer(s *Game, con *websocket.Conn) *HumanPlayer {
func NewHumanPlayer(s *Game) Player {
	PlayerCount += 1

	name := "Player_" + strconv.Itoa(PlayerCount)

	p := new(HumanPlayer)
	p.Name = name
	p.PlayerId = PlayerCount
	p.game = s
	//	p.ws = con
	//	p.Send = make(chan ServerMessage, 10)
	p.SendUpdates = make(chan ServerMessage, 10)
	p.mutex = new(sync.Mutex)
	p.Inventory = make(map[PlayerResourceType]*PlayerInventory)
	p.actionId = 0

	p.AddResource(BoosterResource, 2)
	p.AddResource(ShieldResource, 10)
	p.AddResource(HyperspaceResource, 2)
	p.AddResource(LifeEnergyResource, 100)
	p.AddResource(CloakResource, 5)
	p.Random = rand.New(rand.NewSource(time.Now().Unix()))
	return p
}

func (this *HumanPlayer) SetShip(s *Sprite) {
	this.Ship = s
}

func (this *HumanPlayer) GetShip() *Sprite {
	return this.Ship
}

func (this *HumanPlayer) Update(msg ServerMessage) {
	this.SendUpdates <- msg
}

func (this *HumanPlayer) SetActionId(id int) {
	this.actionId = id
}

func (this *HumanPlayer) GetActionId() int {
	return this.actionId
}

func (this *HumanPlayer) GetPlayerId() int {
	return this.PlayerId
}

func (this *HumanPlayer) GetName() string {
	return this.Name
}

func (this *HumanPlayer) setWebsocket(ws *websocket.Conn) {
	fmt.Printf("Player - setWebsocket\n")
	this.ws = ws
	//this.Send = make(chan ServerMessage)
	//	go this.ping(ws)
	go this.sendRoutine()

	//	this.Send <- NewServerMessage(InitMessage, this, nil)
}

func (this *HumanPlayer) GetMutex() *sync.Mutex {
	return this.mutex
}

func (this HumanPlayer) MarshalJSON() ([]byte, error) {

	b, err := json.Marshal(map[string]interface{}{
		"name":     this.Name,
		"playerId": this.PlayerId,
	})

	if err != nil {
		panic("error marshall Sprite\n")
	}
	return b, err
}

/*
func (c *HumanPlayer) ping(ws *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//			c.GetMutex().Lock()
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				fmt.Println("ping:", err)
			}
			//			c.GetMutex().Unlock()
			break
		}
	}
}
*/
func (this *HumanPlayer) sendRoutine() {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		this.ws.Close()
	}()
	for {
		select {
		case message, ok := <-this.SendUpdates:

			this.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				this.ws.WriteMessage(websocket.CloseMessage, []byte{})
				this.game.Quit(this)
			}

			out, err := proto.Marshal(&message)
			if err != nil {
				return
			}

			err = this.ws.WriteMessage(websocket.BinaryMessage, out)
			if err != nil {
				fmt.Printf("ws.WriteMessage error  %v\n", err)
				return
			}
		case <-pingTicker.C:
			fmt.Printf("Ping\n")
			this.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := this.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				fmt.Printf("Ping error\n")
				this.game.Quit(this)
				return
			}
		}

	}
}

func (c *HumanPlayer) WriteJSON(data interface{}) {
	c.ws.WriteJSON(data)
}

func (this *HumanPlayer) ProcessCommands() {
	defer func() {
		this.ws.Close()
	}()
	this.ws.SetReadLimit(maxMessageSize)
	this.ws.SetReadDeadline(time.Now().Add(pongWait))
	this.ws.SetPongHandler(func(string) error { this.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {

		mtype, messageBytes, err := this.ws.ReadMessage()

		if err != nil {
			fmt.Printf("ReadMessage error %v\n", err)
			this.ws.Close()
			this.game.Quit(this)
			break
		}

		if mtype == websocket.BinaryMessage {

			msgs := &ClientMessages{}
			err = proto.Unmarshal(messageBytes, msgs)

			if err != nil {
				fmt.Println("unmarshall error")
				return
			}

			for _, msg := range msgs.Messages {

				switch msg.Typ {
				case PlayerCommand:

					this.game.PlayerCommands <- PlayerCommandHolder{msg.Cmd, this}
					//this.game.HandlePlayerCommand(this, msg.Cmd)
					break
				default:
				}

			}
		} else if mtype == -1 {
			fmt.Printf("Received close message")
			this.ws.Close()
			break
		}

	}

}

func (this *HumanPlayer) GetInventory() map[PlayerResourceType]*PlayerInventory {
	return this.Inventory
}

func (this *HumanPlayer) AddResource(typ PlayerResourceType, amount int) {
	i, present := this.Inventory[typ]
	if present {
		i.Amount += amount
	} else {
		this.Inventory[typ] = &PlayerInventory{this, typ, amount}
	}
}

func (this *HumanPlayer) DepleteResource(typ PlayerResourceType, amount int) int {
	i, present := this.Inventory[typ]
	if present {
		i.Amount -= amount
		//		if i.Amount <= 0 {
		//			delete(this.Inventory, typ)
		//		}
	}
	return i.Amount
}

func (this *HumanPlayer) HasResource(typ PlayerResourceType) int {
	i, present := this.Inventory[typ]
	if present {
		return i.Amount
	}
	return 0
}

func (this *HumanPlayer) GetPlayerType() PlayerType {
	return HUMAN_PLAYER
}
