package core

import (
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"time"
)

/*
type ServerMessageHolder struct {
	Cmd    ServerMessage
	Player *Player
}
*/

func NewFreezeDrawingMessage() ServerMessage {
	return ServerMessage{Typ: FreezeDrawing, Update: nil, Initialize: nil, Players: nil, Sound: nil, Draw: nil, Dead: nil}
}

func NewInitializePlayer(id int32, shipId int32) ServerMessage {

	init := new(ServerMessage_InitializePlayerData)
	init.PlayerId = id
	init.ShipId = shipId

	return ServerMessage{Typ: PlayerInitialize, Update: nil, Initialize: init, Players: nil, Sound: nil, Draw: nil, Dead: nil}

}

func NewPlayerUpdate(p Player) ServerMessage {

	playerUpdate := new(ServerMessage_PlayerUpdateData)
	playerState := new(PlayerState)
	playerState.Id = int32(p.GetPlayerId())
	playerState.Name = p.GetName()
	if p.GetShip() != nil {
		playerState.ShipId = int32(p.GetShip().Id)
	}

	//	fmt.Printf("NewPlayerUpdate: # of inventories = %d, ship = %d\n", len(p.Inventory), playerState.ShipId)
	inventories := make([]*PlayerState_Inventory, 0)
	for itemType, inv := range p.GetInventory() {
		fmt.Printf("NewPlayerUpdate: itemType = %s\n", itemType)
		inventories = append(inventories, &PlayerState_Inventory{
			itemType, int32(inv.Amount), nil})
	}

	playerState.Inventory = inventories
	playerUpdate.Player = playerState

	return ServerMessage{Typ: PlayerUpdate, Update: nil, Initialize: nil, Players: playerUpdate, Sound: nil, Draw: nil, Dead: nil}
}

func NewDrawMessage(cmds []string) ServerMessage {

	draw := new(ServerMessage_DrawData)
	draw.Cmds = make([]string, len(cmds))
	copy(draw.Cmds, cmds)

	return ServerMessage{Typ: DrawMessage, Update: nil, Initialize: nil, Players: nil, Sound: nil, Draw: draw, Dead: nil}

}

func NewPlaySoundMessage(soundType SoundType, volume float64) ServerMessage {

	msg := new(ServerMessage_PlaySoundData)
	msg.SoundType = soundType
	msg.Volume = *(proto.Float64(volume))
	return ServerMessage{Typ: PlaySound, Update: nil, Initialize: nil, Players: nil, Sound: msg, Draw: nil, Dead: nil}
}

func (this *ServerMessage_PhysicsUpdateData) SetActionid(actionId int) {
	this.Actionid = *(proto.Int32(int32(actionId)))
}

func NewPlayerDead(p Player) ServerMessage {
	msg := new(ServerMessage_PlayerDeadData)
	msg.PlayerId = int32(p.GetPlayerId())
	return ServerMessage{Typ: PlayerDead, Update: nil, Initialize: nil, Players: nil, Sound: nil, Draw: nil, Dead: msg}
}

func NewPhysicsUpdate(frame int32, sprites map[*Sprite]bool) ServerMessage {

	update := new(ServerMessage_PhysicsUpdateData)
	update.TimeNanos = *(proto.Int64(time.Now().UnixNano()))
	update.Frame = *(proto.Int32(frame))
	//	update.Sprites = new([]*SpriteState)

	for s, b := range sprites {
		if b == false {
			continue
		}
		//spriteTypeNum := SpriteType_value[s.Type.String()]

		var playerId int32
		if s.player != nil {
			playerId = int32(s.player.GetPlayerId())
		}

		/*
			forces := make([]SpriteState_Force, 0)
			for f, active := range s.Forces {
				if active {
					forces = append(forces, SpriteState_Force{f.Typ, f.Dir, f.Mag, f.ActionId, nil})
				}
			}
		*/

		/*
			states := make([]SpriteStatus, 0)
			for state, _ := range s.States {
				states = append(states, state)
			}
		*/

		//var prize *SpriteState_Prize = nil
		//if s.prize != nil {
		//j	prize = &SpriteState_Prize{s.prize.resource, int32(s.prize.value), nil}
		//	}

		//		props := make([]Property, 0)
		//		for n, v := range s.Properties {
		//			props = append(props, Property{n, v, nil})
		//		}

		s.GetMutex().Lock()

		state := &SpriteState{
			Id:  *proto.Int32(int32(s.Id)),
			Typ: *proto.Uint32(uint32(s.typeInfo)),
			X:   *(proto.Float64(s.Position.x)),
			Y:   *(proto.Float64(s.Position.y)),
			//Vx:       *(proto.Float64(s.Velocity.X)),
			//Vy:       *(proto.Float64(s.Velocity.Y)),
			//Ax:       *(proto.Float64(s.Accel.X)),
			//Ay:       *(proto.Float64(s.Accel.Y)),
			Height:   *proto.Int32(int32(s.Height)),
			Width:    *proto.Int32(int32(s.Width)),
			Rotation: *(proto.Float64(s.Rotation)),
			Mass:     *(proto.Float64(s.Mass)),
			PlayerId: *proto.Int32(playerId),
			//Forces:   forces,
			//	States:   states,
			//			Prize:    prize,
			//Properties: props,
		}
		s.GetMutex().Unlock()
		update.Sprites = append(update.Sprites, state)
	}

	return ServerMessage{Typ: PhysicsUpdate, Initialize: nil, Update: update, Players: nil, Sound: nil, Draw: nil, Dead: nil}

}
