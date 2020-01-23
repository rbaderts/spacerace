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
	_ "github.com/golang/protobuf/proto"
	_ "github.com/gorilla/websocket"
	"github.com/rbaderts/spacerace/core/messages"

	"math/rand"
	_ "os"
	"strconv"
	"sync"
	"time"
)

const (
	tickPeriod = 50 * time.Millisecond
)

type AIPlayer struct {
	Name     string
	PlayerId int
	game     *Game
	Ship     *Sprite
	mutex    *sync.Mutex

	// map of item type name to PlayerInventoryRecord
	Inventory     map[PlayerResourceType]*PlayerInventory
	UserId        int
	actionId      int
	actionCounter int
	Random        *rand.Rand

	//spriteState []*SpriteState
	target *Sprite
}

func NewAIPlayer(s *Game) Player {
	PlayerCount += 1

	name := "AIPlayer_" + strconv.Itoa(PlayerCount)

	p := new(AIPlayer)
	p.Name = name
	p.PlayerId = PlayerCount
	p.game = s
	//	p.ws = con
	//	p.Send = make(chan ServerMessage, 10)
	p.mutex = new(sync.Mutex)
	p.Inventory = make(map[PlayerResourceType]*PlayerInventory)
	p.actionId = 0
	p.actionCounter = 0

	p.AddResource(messages.PlayerResourceTypeBooster, 2)
	p.AddResource(messages.PlayerResourceTypeShield, 10)
	p.AddResource(messages.PlayerResourceTypeHyperdrive, 2)
	p.AddResource(messages.PlayerResourceTypeLife, 100)
	p.AddResource(messages.PlayerResourceTypeCloak, 5)
	p.Random = rand.New(rand.NewSource(time.Now().Unix()))

	go p.loop()
	return p
}

func (this *AIPlayer) SetShip(s *Sprite) {
	this.Ship = s
}

func (this *AIPlayer) GetShip() *Sprite {
	return this.Ship
}

func (this *AIPlayer) chooseTarget() *Sprite {

	var minDistance float64 = 10000000
	var target *Sprite

	sprites := this.game.CopySprites()

	for _, s := range sprites {
		kind := SpriteKind(s.typeInfo & SPRITE_KIND)
		if (kind == messages.SpriteKindShip) &&
			(s.typeInfo&SHIP_STATE) != CLOAK_MODE {
			p := &Point{s.Position.x, s.Position.y}
			dis := this.Ship.Position.Distance(*p)
			if dis < minDistance {
				target = s
				minDistance = dis
			}
		}

	}
	if minDistance > 1200 {
		return nil
	}
	return target
}

func (this *AIPlayer) loop() {

	ticker := time.NewTicker(tickPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			target := this.chooseTarget()
			if target != nil {
				this.attack(target)
			}
		}
	}
}

func (this *AIPlayer) attack(s *Sprite) {

	dir := this.Ship.Position.Direction(s.Position)
	this.Ship.Rotation = dir

	cmd := NewPlayerCommand(this, messages.CommandTypeThrust, 20, int32(this.actionCounter))
	this.actionCounter += 1

	fmt.Printf("AIPlayer attacking\n")
	this.game.PlayerCommands <- PlayerCommandHolder{cmd, this}

}

/*
func (this *AIPlayer) Update(msg ServerMessage) {

	/*
		switch msg.Typ {
		case MessageType_PhysicsUpdate:
			this.spriteState = msg.Update.Sprites
		}
	*/


func (this *AIPlayer) UpdateWithBytes(bytes []byte) {

	/*
		switch msg.Typ {
		case PhysicsUpdate:
			this.spriteState = msg.Update.Sprites
		}
	*/
}

func (this *AIPlayer) SetActionId(id int) {
	this.actionId = id
}

func (this *AIPlayer) GetActionId() int {
	return this.actionId
}

func (this *AIPlayer) GetPlayerId() int {
	return this.PlayerId
}

func (this *AIPlayer) GetName() string {
	return this.Name
}

func (this *AIPlayer) GetMutex() *sync.Mutex {
	return this.mutex
}

func (this AIPlayer) MarshalJSON() ([]byte, error) {

	b, err := json.Marshal(map[string]interface{}{
		"name":     this.Name,
		"playerId": this.PlayerId,
	})

	if err != nil {
		panic("error marshall Sprite\n")
	}
	return b, err
}

func (this *AIPlayer) sendRoutine() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			fmt.Printf("AIPlayer Tick\n")
		}
	}
}

func (this *AIPlayer) ProcessCommands() {
	//this.game.PlayerCommands <- PlayerCommandHolder{msg.Cmd, this}

}

func (this *AIPlayer) GetInventory() map[PlayerResourceType]*PlayerInventory {
	return this.Inventory
}

func (this *AIPlayer) AddResource(typ PlayerResourceType, amount int) {
	i, present := this.Inventory[typ]
	if present {
		i.Amount += amount
	} else {
		this.Inventory[typ] = &PlayerInventory{this, typ, amount}
	}
}

func (this *AIPlayer) DepleteResource(typ PlayerResourceType, amount int) int {
	i, present := this.Inventory[typ]
	if present {
		i.Amount -= amount
		//		if i.Amount <= 0 {
		//			delete(this.Inventory, typ)
		//		}
	}
	return i.Amount
}

func (this *AIPlayer) HasResource(typ PlayerResourceType) int {
	i, present := this.Inventory[typ]
	if present {
		return i.Amount
	}
	return 0
}
func (this *AIPlayer) GetPlayerType() PlayerType {
	return AI_PLAYER
}
