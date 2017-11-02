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

	"math/rand"
	_ "os"
	"strconv"
	"sync"
	"time"
)

const (
	tickPeriod = 100 * time.Millisecond
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

	spriteState []*SpriteState
	target      *Sprite
}

//func NewPlayer(s *Game, con *websocket.Conn) *AIPlayer {
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

	p.AddResource(BoosterResource, 2)
	p.AddResource(ShieldResource, 10)
	p.AddResource(HyperspaceResource, 2)
	p.AddResource(LifeEnergyResource, 100)
	p.AddResource(CloakResource, 5)
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

func (this *AIPlayer) chooseTarget() *SpriteState {

	//spriteState []*SpriteState
	var minDistance float64 = 10000000
	var target *SpriteState

	this.mutex.Lock()
	for _, s := range this.spriteState {
		kind := s.Typ & SPRITE_KIND

		if (kind == SHIP) && (int(s.Id) != this.Ship.Id) &&
			(s.Typ&SHIP_STATE) != CLOAK_MODE {
			p := &Point{s.X, s.Y}
			dis := this.Ship.Position.Distance(*p)
			if dis < minDistance {
				target = s
				minDistance = dis
			}
		}

	}
	this.mutex.Unlock()
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

func (this *AIPlayer) attack(s *SpriteState) {

	targetPosition := &Point{s.X, s.Y}
	dir := this.Ship.Position.Direction(*targetPosition)

	//fmt.Printf("direction to target ship = %v\n", dir)
	//fmt.Printf("Ships current direction = %v\n", this.Ship.Rotation)

	this.Ship.Rotation = dir

	thrustCommand := PlayerCommandMessage{Thrust, 20, int32(this.actionCounter), nil}
	this.actionCounter += 1

	this.game.PlayerCommands <- PlayerCommandHolder{thrustCommand, this}

	/*	diffDir := dir - this.Ship.Rotation
		fmt.Printf("AIPlayer currentDir: %v, targetDir: %v, diff: %v\n", this.Ship.Rotation, dir, diffDir)

		if diffDir != 0 {
			rotateCommand := PlayerCommandMessage{Rotate, diffDir, int32(this.actionCounter), nil}
			this.actionCounter += 1
			this.game.PlayerCommands <- PlayerCommandHolder{rotateCommand, this}
		}

		thrustCommand := PlayerCommandMessage{Thrust, 20, int32(this.actionCounter), nil}
		this.actionCounter += 1

		this.game.PlayerCommands <- PlayerCommandHolder{thrustCommand, this}
	*/

}

func (this *AIPlayer) Update(msg ServerMessage) {
	//	this.SendUpdates <- msg

	switch msg.Typ {

	case PhysicsUpdate:
		this.mutex.Lock()
		this.spriteState = msg.Update.Sprites
		this.mutex.Unlock()
	}

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
