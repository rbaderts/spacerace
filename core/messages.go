package core

import (
	"fmt"
	"sort"

	//proto "github.com/golang/protobuf/proto"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/rbaderts/spacerace/core/messages"

	"time"
)

/*
type UpdateHolder struct {
	Cmd    Update
	Player *Player
}
*/

type PlayerResourceType int32
type CommandType int32
type SpriteKind int32
type SoundType int32
type ForceType int32
type SpriteStatus int32


func NewShakeMessage(spriteId int, mag int32) []byte {

	builder := flatbuffers.NewBuilder(0)

	messages.ShakeStart(builder)
	messages.ShakeAddSpriteId(builder, int32(spriteId))
	messages.ShakeAddMagnitude(builder, mag)

	shakeOffset := messages.ShakeEnd(builder)

	messages.UpdateStart(builder)
	messages.UpdateAddMessageType(builder, messages.UpdateMessagePlaySound)
	messages.UpdateAddMessage(builder, shakeOffset)
	offset := messages.UpdateEnd(builder)


	builder.Finish(offset)
	buf := builder.FinishedBytes() // Of type `byte[]`.
	return buf
}

/*

func NewFreezeDrawingMessage() ServerMessage {
	return ServerMessage{Typ: MessageType_FreezeDrawing, Update: nil, Initialize: nil, Players: nil, Sound: nil, Draw: nil, Dead: nil, Shake: nil}
}
 */

func NewPlaySoundMessage(soundtype SoundType, vol float64) []byte {

	builder := flatbuffers.NewBuilder(0)

	messages.PlaySoundStart(builder)
	messages.PlaySoundAddSoundType(builder, int8(soundtype))
	messages.PlaySoundAddVolume(builder, vol)

	playSoundOffset := messages.PlaySoundEnd(builder)

	messages.UpdateStart(builder)
	messages.UpdateAddMessageType(builder, messages.UpdateMessagePlaySound)
	messages.UpdateAddMessage(builder, playSoundOffset)
	offset := messages.UpdateEnd(builder)

	builder.Finish(offset)
	buf := builder.FinishedBytes() // Of type `byte[]`.
	return buf
}

func NewInitializePlayer(id int32, shipId int32) []byte {

	builder := flatbuffers.NewBuilder(0)
	messages.InitializePlayerStart(builder)
	messages.InitializePlayerAddPlayerId(builder, id)
	messages.InitializePlayerAddShipId(builder, shipId)
	initOffset := messages.InitializePlayerEnd(builder)

	messages.UpdateStart(builder)
	messages.UpdateAddMessageType(builder, messages.UpdateMessageInitializePlayer)
	messages.UpdateAddMessage(builder, initOffset)
	offset := messages.UpdateEnd(builder)

	builder.Finish(offset)
	buf := builder.FinishedBytes()

	return buf

}

func newInventoryItem(
	builder *flatbuffers.Builder,
	itemType PlayerResourceType,
	amount int) flatbuffers.UOffsetT {

	messages.InventoryStart(builder)
	messages.InventoryAddResourceType(builder, uint32(itemType))
	messages.InventoryAddValue(builder, int32(amount))
	return messages.InventoryEnd(builder)
}

func NewPlayerUpdate(p Player) []byte {
	builder := flatbuffers.NewBuilder(0)

	inventoryOffsets := make([]flatbuffers.UOffsetT, 0)

	imap := p.GetInventory()
	keys := make([]int, 0, len(imap))
	for k := range imap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, k := range keys {
		inv := imap[PlayerResourceType(k)]
		offset := newInventoryItem(
			builder, PlayerResourceType(k), inv.Amount)
		inventoryOffsets = append(inventoryOffsets, offset)
	}

	/*
	for itemType, inv := range p.GetInventory() {
		fmt.Printf("NewPlayerUpdate: itemType = %s\n", itemType)

		offset := newInventoryItem(builder, itemType, inv.Amount)
		inventoryOffsets = append(inventoryOffsets, offset)
	}
	 */

	playerNameOffset := builder.CreateString(p.GetName())

	messages.PlayerUpdateStartInventoryVector(builder, len(inventoryOffsets))
	for i := len(inventoryOffsets) - 1; i >= 0; i-- {
		offset := inventoryOffsets[i]
		builder.PrependUOffsetT(offset)

	}
	/*
	for _, offset := range inventoryOffsets {
		builder.PrependUOffsetT(offset)
	}
	 */
	inventoryItems := builder.EndVector(len(inventoryOffsets))

	messages.PlayerUpdateStart(builder)
	messages.PlayerUpdateAddId(builder, int32(p.GetPlayerId()))
	messages.PlayerUpdateAddShipId(builder, int32(p.GetShip().Id))
	messages.PlayerUpdateAddName(builder, playerNameOffset)
	messages.PlayerUpdateAddInventory(builder, inventoryItems)

	playerUpdate := messages.PlayerUpdateEnd(builder)

	messages.UpdateStart(builder)
	messages.UpdateAddMessageType(builder, messages.UpdateMessagePlayerUpdate)
	messages.UpdateAddMessage(builder, playerUpdate)
	offset := messages.UpdateEnd(builder)

	builder.Finish(offset)
	buf := builder.FinishedBytes() // Of type `byte[]`.

	return buf
}

func NewPlayerCommand(p Player, cmd CommandType, value float64, actionId int32) *messages.PlayerCommandMessage {
	builder := flatbuffers.NewBuilder(64)
	messages.PlayerCommandMessageStart(builder)
	messages.PlayerCommandMessageAddCmd(builder, int8(cmd))
	messages.PlayerCommandMessageAddValue(builder, value)
	messages.PlayerCommandMessageAddActionId(builder, actionId)
	cmdOffset := messages.PlayerCommandMessageEnd(builder)
	builder.Finish(cmdOffset)
	buf := builder.FinishedBytes() // Of type `byte[]`.
	fmt.Printf("PlayerCommandMessage buf size = %d\n", len(buf))
//	fmt.Printf("PlayerCommandMessage = %x\n", buf)
	return messages.GetRootAsPlayerCommandMessage(buf, 0)
}


/*
func NewDrawMessage(cmds []string) ServerMessage {

	draw := new(DrawData)
	draw.Cmds = make([]string, len(cmds))
	copy(draw.Cmds, cmds)

	return ServerMessage{Typ: MessageType_DrawMessage, Update: nil, Initialize: nil, Players: nil, Sound: nil, Draw: draw, Dead: nil, Shake: nil}

}

/*
func (this *messages.PhysicsUpdate) SetActionId(actionId int) {
	this.ActionId = int32(actionId)
}

func NewPlayerDead(p Player) ServerMessage {
	msg := new(PlayerDeadData)
	msg.PlayerId = int32(p.GetPlayerId())
	return ServerMessage{Typ: MessageType_PlayerDead, Update: nil, Initialize: nil, Players: nil, Sound: nil, Draw: nil, Dead: msg, Shake: nil}
}
*/

/*
var updateMessage Update = nil;
var PhysicUpdate Update = return Update{Typ: PhysicsUpdate, Initialize: nil, Update: update, Players: nil, Sound: nil, Draw: nil, Dead: nil}
func getUpdateMessage() Update {

	if (updateMessage == nil) {
		update := new(PhsyicsUpdateData)
		updateMessage = &Update{Typ: PhysicsUpdate, Initialize: nil, Update: update, Players: nil, Sound: nil, Draw: nil, Dead: nil}
	}

	return updateMessage;
}
*/

var counter int = 0


func NewPhysicsUpdate(frame int32, frameTime int64, actionId int, sprites []*Sprite, player Player) []byte {

	builder := flatbuffers.NewBuilder(64)

	counter += 1
	spriteOffsets := make([]flatbuffers.UOffsetT, 0)

	for _, s := range sprites {
		offset := newSpriteState(builder, s)
		spriteOffsets = append(spriteOffsets, offset)
	}

	messages.PhysicsUpdateStartSpritesVector(builder, len(spriteOffsets))
	for i := len(spriteOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(spriteOffsets[i])
	}
	spriteRecords := builder.EndVector(len(spriteOffsets))

	messages.PhysicsUpdateStart(builder)
	messages.PhysicsUpdateAddTimeNanos(builder, int64(time.Now().UnixNano()))
	messages.PhysicsUpdateAddFrame(builder, frame)
	messages.PhysicsUpdateAddFrameTime(builder, frameTime)
	messages.PhysicsUpdateAddActionId(builder, int32(actionId))
	messages.PhysicsUpdateAddSprites(builder, spriteRecords)

	physicsOffset := messages.PhysicsUpdateEnd(builder)

	messages.UpdateStart(builder)
	messages.UpdateAddMessageType(builder, messages.UpdateMessagePhysicsUpdate)
	messages.UpdateAddMessage(builder, physicsOffset)
	offset := messages.UpdateEnd(builder)

	builder.Finish(offset)
	buf := builder.FinishedBytes() // Of type `byte[]`.

	fmt.Printf("Update buf size = %d\n", len(buf))

	return buf

}

func newSpriteState(builder *flatbuffers.Builder, sprite *Sprite) flatbuffers.UOffsetT {

	playerId := 0
	var playerNameOffset flatbuffers.UOffsetT = 0

	if (sprite.player != nil) {
		playerId = sprite.player.GetPlayerId()
		playerNameOffset = builder.CreateString(sprite.player.GetName())
	}

	messages.SpriteStateStart(builder)
	messages.SpriteStateAddId(builder, int32(sprite.Id))
	messages.SpriteStateAddTyp(builder, sprite.typeInfo)
	messages.SpriteStateAddX(builder, sprite.Position.x)
	messages.SpriteStateAddY(builder, sprite.Position.y)
	messages.SpriteStateAddVx(builder, sprite.Velocity.x)
	messages.SpriteStateAddVy(builder, sprite.Velocity.y)
	messages.SpriteStateAddMass(builder, sprite.Mass)
	messages.SpriteStateAddRotation(builder, sprite.Rotation)
	messages.SpriteStateAddWidth(builder, int32(sprite.Width))
	messages.SpriteStateAddHeight(builder, int32(sprite.Height))
	messages.SpriteStateAddHealthpoints(builder, sprite.HealthPoints)
	messages.SpriteStateAddPlayerId(builder, int32(playerId));
	messages.SpriteStateAddPlayerName(builder, playerNameOffset);
	return messages.SpriteStateEnd(builder)
}

