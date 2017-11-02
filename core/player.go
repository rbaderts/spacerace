/**
 * A Player is the runtime representation of a single user playing a game
 *   A Player encapsulates a websocket connection to a browser window
 *      displaying the game
 *
 */
package core

import (
	"encoding/json"
	_ "fmt"
	_ "github.com/golang/protobuf/proto"
	_ "github.com/gorilla/websocket"

	_ "math/rand"
	_ "os"
	_ "strconv"
	_ "sync"
	_ "time"
)

var MessageNumber int32 = 0
var PlayerCount int = 0

type PlayerCommandHolder struct {
	Cmd    PlayerCommandMessage
	Player Player
}

type PlayerType int

const (
	_ PlayerType = iota
	HUMAN_PLAYER
	AI_PLAYER
)

type Player interface {
	Update(msg ServerMessage)
	GetPlayerId() int

	GetActionId() int
	SetActionId(id int)

	SetShip(s *Sprite)
	GetShip() *Sprite

	GetName() string

	GetInventory() map[PlayerResourceType]*PlayerInventory
	HasResource(typ PlayerResourceType) int
	DepleteResource(typ PlayerResourceType, amount int) int
	AddResource(typ PlayerResourceType, amount int)

	GetPlayerType() PlayerType
}

type InitData struct {
	PlayerId int
}

type PlayerInventory struct {
	player Player
	Typ    PlayerResourceType
	Amount int
}

func (this PlayerInventory) MarshalJSON() ([]byte, error) {

	b, err := json.Marshal(map[string]interface{}{
		"typ":    this.Typ,
		"amount": this.Amount,
	})

	if err != nil {
		panic("error marshall Sprite\n")
	}
	return b, err
}

type DrawCmds struct {
	Cmds []string `json: "cmds"`
}
