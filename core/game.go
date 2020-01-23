/**
 * A game is the runtime representation of an active game session
 * There may be 0 or more Players attached to the Game
 */
package core

import (
	"encoding/json"
	"fmt"
	"github.com/rbaderts/spacerace/core/messages"

	"github.com/sirupsen/logrus"
	_ "github.com/vmihailenco/msgpack"
	"math"
	"math/rand"
	_ "os"
	"sync"
	"time"
)

var (
	MAX_QUEUE = 400
)

var FPS int64 = 60 // Ticks per second
var random = rand.New(rand.NewSource(time.Now().Unix()))

type GameStateType int

const (
	_ GameStateType = iota
	Running
	Waiting
	Shutdown
)

var GameStateTypes = [...]string{
	"None",
	"Running",
	"Waiting",
	"Shutdown",
}

func (b GameStateType) String() string {
	return GameStateTypes[b]
}

type Game struct {
	Id             int64
	Players        map[Player]*Sprite
	Sprites        map[*Sprite]bool
	PlayerCommands chan PlayerCommandHolder

	sendRoutineQuit     chan bool
	receiveRoutineQuit  chan bool
	gameLoopRoutineQuit chan bool
	width               int
	height              int
	state               GameStateType
	T                   int64
	frame        int32
	mutex        *sync.Mutex
	spritesMutex *sync.Mutex
	bulletSpeed  int
	Race         *Race
	Frozen       bool
	spriteCounts map[SpriteKind]int
	blackhole    *Sprite
}

func (this *Game) IsRunning() bool {
	return this.state == Running
}

func (this *Game) Start() {
	this.T = 0
	if this.state != Running && len(this.Players) >= 1 {
		go this.GameLoop()
		this.state = Running
	}
}

func (this *Game) Stop() {

	fmt.Printf("Stop\n")

	this.spritesMutex.Lock()
	for s, _ := range this.Sprites {
		delete(this.Sprites, s)
	}
	this.spritesMutex.Unlock()
	this.state = Waiting

	err := this.Race.UpdateRaceStatus(Environment.DB, RaceComplete)
	if err != nil {
		fmt.Printf("err = %v\n", err)
	}
	this.gameLoopRoutineQuit <- true

}

func (this *Game) incrementSpriteCount(typ SpriteKind, cnt int) {

	//	this.mutex.Lock()
	_, present := this.spriteCounts[typ]
	if !present {
		this.spriteCounts[typ] = cnt
	} else {
		this.spriteCounts[typ] = this.spriteCounts[typ] + cnt
	}
	//	this.mutex.Unlock()
}

func (this *Game) decrementSpriteCount(typ SpriteKind, cnt int) {
	//	this.mutex.Lock()
	_, present := this.spriteCounts[typ]
	if !present {
		return
	} else {
		this.spriteCounts[typ] = this.spriteCounts[typ] - cnt
	}
	if this.spriteCounts[typ] < 0 {
		this.spriteCounts[typ] = 0
	}
	//	this.mutex.Unlock()
}

func NewGame(id int64) *Game {
	game := new(Game)
	game.state = Waiting
	game.Id = id
	game.Players = make(map[Player]*Sprite)
	game.Sprites = make(map[*Sprite]bool)
	game.PlayerCommands = make(chan PlayerCommandHolder, MAX_QUEUE)
	game.sendRoutineQuit = make(chan bool)
	game.receiveRoutineQuit = make(chan bool)
	game.gameLoopRoutineQuit = make(chan bool)
	game.mutex = new(sync.Mutex)
	game.spritesMutex = new(sync.Mutex)

//	game.updates = make(chan ServerMessage, 10)
//	game.sounds = make(chan ServerMessage, 10)
	game.spriteCounts = make(map[SpriteKind]int)
	game.width = 5000
	game.height = 5000
	for i := 0; i < 12; i++ {
		game.newLargeAsteroid(i % 4)
	}
	for i := 0; i < 2; i++ {
		game.newAIPlayer()
	}

	game.newBlackhole(9e+13)
	game.newBlackhole(9e+13)
	game.newBlackhole(9e+13)

	game.newSpacestation()
	game.newSpacestation()
	game.newSpacestation()

	game.newEndToken()

	game.newStar(24, 2e+13)
	game.newStar(30, 2e+13)
	game.newPlanet(40, 2e+11)
	game.frame = 1

	Log.WithFields(logrus.Fields{"gameId": id}).Info("New Game")

	game.bulletSpeed = 170
	game.Frozen = false
	return game
}

//var _ msgpack.CustomEncoder = (*Game)(nil)

//func (this *Game) EncodeMsgpack(enc *msgpack.Encoder) error {

//sprites := this.CopySprites();
//	return enc.Encode(this.Sprites)
//}

func (this Game) MarshalJSON() ([]byte, error) {

	//sprites := this.kkk/CopySprites();
	sprites := this.CopySprites()

	players := make([]Player, 0, len(this.Players))
	for p := range this.Players {
		players = append(players, p)
	}

	b, err := json.Marshal(map[string]interface{}{
		"sprites": sprites,
		"players": players,
	})

	if err != nil {
		panic(err)
	}
	return b, err
}

func (this *Game) RemoveSprite(s *Sprite) {

	this.spritesMutex.Lock()
	if _, ok := this.Sprites[s]; !ok {
		return
	}

	kind := s.GetKind()
	fmt.Printf("RemoveSprite: kind = %v\n", kind)

    if kind == messages.SpriteKindShip {
		if s.player.HasResource(messages.PlayerResourceTypeLife) >= 5 {
			s.player.DepleteResource(messages.PlayerResourceTypeLife, 10)
			s.HealthPoints = 100
			loc := &Point{float64(this.width / 2), float64(this.height / 2)}
			s.Position = *loc
			s.Rotation = 0
			s.Velocity = Vector{0, 0}

			this.UpdatePlayer(s.player)
		}
		this.spritesMutex.Unlock()
		return
	}
	delete(this.Sprites, s)
	this.spritesMutex.Unlock()

	if kind == messages.SpriteKindLargeAsteroid {
		this.newSmallAsteroids(s.Position)
		this.decrementSpriteCount(messages.SpriteKindLargeAsteroid, 1)
	} else if kind == messages.SpriteKindSmallAsteroid {
		this.decrementSpriteCount(messages.SpriteKindSmallAsteroid, 1)
	} else if kind == messages.SpriteKindPrize {
		this.decrementSpriteCount(messages.SpriteKindPrize, 1)
	} else if kind == messages.SpriteKindAiShip {
//		s.player.Update(NewPlayerDead(s.player))
	}
}

func (this *Game) Join(p Player) {
	Log.WithFields(logrus.Fields{"player": p.GetName(), "game": this.Id}).Info("Player joined game")
	ship := this.newShip(p, nil, messages.SpriteKindShip)
	this.Players[p] = ship

	p.UpdateWithBytes(NewInitializePlayer(int32(p.GetPlayerId()), int32(ship.Id)))
	//	this.UpdatePlayer(p)
	//	this.Start()
}

func (this *Game) Quit(player Player) {
	if _, ok := this.Players[player]; ok {
		//this.RemoveSprite(this.Players[player])
		delete(this.Sprites, player.GetShip())
		delete(this.Players, player)
	}

	var humanPlayerCount int = 0
	for p := range this.Players {
		if p.GetPlayerType() == HUMAN_PLAYER {
			humanPlayerCount += 1
			break
		}
	}

	if humanPlayerCount <= 0 {
		this.Stop()
		this.state = Shutdown
	}
}

func (this *Game) Complete(player Player) {
	this.Stop()
}

func (this *Game) fire(source *Sprite, val float64) {
	bullet := this.newBullet(source, val)
	bullet.VelocityLimit = 500
	//    this.addSprite(bullet)
}

func (this *Game) phaser(source *Sprite, val float64) {
	//	phaser := this.newPhaser(source, val)
	//	this.Sprites[phaser] = true
}

func (this *Game) addSprite(s *Sprite) {
	this.spritesMutex.Lock()
	this.Sprites[s] = true
	this.spritesMutex.Unlock()
}

func (this *Game) newBullet(source *Sprite, velocity float64) *Sprite {

	velocityVec := UnitVector(source.Rotation).Mul(velocity)
	netVelocity := velocityVec.Add(source.Velocity)

	var x float64 = float64(source.Width) / 2
	var y float64 = 0

	newX := source.Position.x + x*math.Cos(source.Rotation) - y*math.Sin(source.Rotation)
	newY := source.Position.y + x*math.Sin(source.Rotation) + y*math.Cos(source.Rotation)

	sprite := NewSprite2(this, int32(messages.SpriteKindBullet), Point{newX, newY}, 6, 6, netVelocity, 1, 10, 100, 40)
	sprite.Parent = source

	this.addSprite(sprite)

	return sprite
}

/*
func (this *Game) newExplosion(location Point, size int) *Sprite {

	s := NewSprite(this, Explosion, location, size, size, Vector{0, 0}, 0, 1)
	s.Parent = nil
	this.Sprites[s] = true
	return s
}
*/

func (this *Game) newStar(radius int, mass float64) *Sprite {

	//	randX := float64(random.Intn(400))
	//	randY := float64(random.Intn(400))

	sprite := NewSprite(this, int32(messages.SpriteKindStar),
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		radius*2, radius*2, Vector{0, 0}, mass, 0, 5000)
	sprite.SetCollisionCircle(float64(radius) - (.20 * float64(radius)))
	sprite.Parent = nil
	this.addSprite(sprite)
	return sprite
}

func (this *Game) newPlanet(radius int, mass float64) *Sprite {

	sprite := NewSprite(this, int32(messages.SpriteKindPlanet),
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		radius*2, radius*2, Vector{0, 0}, mass, 0, 500)
	sprite.SetCollisionCircle(float64(radius) - (.20 * float64(radius)))
	sprite.Parent = nil
	this.addSprite(sprite)
	return sprite
}

func (this *Game) newBlackhole(mass float64) *Sprite {

	sprite := NewSprite(this, int32(messages.SpriteKindBlackhole),
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		40, 40, Vector{0, 0}, mass, 0, 100000)
	sprite.SetCollisionCircle(1)
	sprite.Parent = nil
	this.addSprite(sprite)
	this.blackhole = sprite
	return sprite
}

func (this *Game) newAIPlayer() {

	player := NewAIPlayer(this)
	sprite := this.newShip(player,
		&Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		messages.SpriteKindAiShip)

	this.Players[player] = sprite
	//this.addSprite(sprite);
	player.SetShip(sprite)

}

func (this *Game) newLargeAsteroid(quadrant int) *Sprite {
	velocity := Vector{float64(random.Intn(20) - 10), float64(random.Intn(20) - 10)}

	sprite := NewSprite(this, int32(messages.SpriteKindLargeAsteroid),
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		30, 30, velocity, 5e+9, 0, 14)

	f := NewAngularForce(messages.ForceTypeInitialForce, 50*(rand.Float64()-0.5), 100*1000*1000)
	sprite.AddForce(f)

	sprite.SetCollisionCircle(15)
	this.addSprite(sprite)
	this.incrementSpriteCount(messages.SpriteKindLargeAsteroid, 1)
	return sprite
}

func (this *Game) newSmallAsteroids(pos Point) *Sprite {

	// := float64(this.width / 2)
	//y := float64(this.height / 2)
	xRand := random.Intn(20)
	yRand := random.Intn(20)
	velocity := Vector{float64(xRand - 10), float64(yRand - 10)}
	p1 := pos
	p1.x = p1.x + float64(xRand-10)
	p1.y = p1.y + float64(yRand-10)
	s1 := NewSprite(this, int32(messages.SpriteKindSmallAsteroid), p1,
		14, 14, velocity, 5e+5, 0, 10)

	f := NewAngularForce(messages.ForceTypeInitialForce, 50*(rand.Float64()-0.5), 100*1000*1000)
	s1.AddForce(f)

	s1.SetCollisionCircle(7)
	this.addSprite(s1)

	velocity = Vector{float64(-(xRand - 10)), float64(-(yRand - 10))}
	p2 := pos
	p2.x = p2.x + float64(-(xRand - 10))
	p2.y = p2.y + float64(-(yRand - 10))
	s2 := NewSprite(this, int32(messages.SpriteKindSmallAsteroid), p2,
		14, 14, velocity, 5e+5, 0, 10)

	f = NewAngularForce(messages.ForceTypeInitialForce, 50*(rand.Float64()-0.5), 100*1000*1000)
	s2.AddForce(f)

	s2.SetCollisionCircle(7)
	this.addSprite(s2)

	this.incrementSpriteCount(messages.SpriteKindSmallAsteroid, 2)

	return s2
}

func (this *Game) newEndToken() *Sprite {

	spriteType := messages.SpriteKindEndToken

	s := NewSprite(
		this, int32(spriteType),
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		42, 42, Vector{0, 0}, 2e+2, 0, 2)

	s.SetCollisionCircle(15)
	this.addSprite(s)

	return s
}

func (this *Game) newPrizeSprite(prizeType PlayerResourceType, value int) *Sprite {

	spriteType := uint32(messages.SpriteKindPrize)
	spriteType |= uint32(prizeType)
	spriteType |= uint32(value)

	s := NewSprite(
		this, int32(spriteType),
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		48, 48, Vector{0, 0}, 2e+2, 0, 2)

	//p := &PrizeTokenSprite{Sprite: *s, value: value}
	/*
		p := "Unknown"
		switch resource {
		case Life:
			p = "LifeEnergy"
			break
		case Shield:
			p = "Shield"
			break
		case Hyperdrive:
			p = "Hyperspect"
			break
		case Booster:
			p = "Booster"
			break

		}
	*/
	//	s.addProperty("Prize", fmt.Sprintf("%v", resource))
	//	s.addProperty("PrizeValue", fmt.Sprintf("%v", value))

	s.SetCollisionCircle(15)
	this.addSprite(s)

	this.incrementSpriteCount(messages.SpriteKindPrize, 1)

	//	prize := &Prize{resource, value}
	//	s.prize = prize
	return s

}

func (this *Game) newSpacestation() *Sprite {

	sprite := NewSprite(this, int32(messages.SpriteKindSpaceStation),
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		130, 130, Vector{0, 0}, 9e+8, 0, 100)

	f := NewAngularForce(messages.ForceTypeInitialForce, 50*(rand.Float64()-0.5), 100*1000*1000)
	sprite.AddForce(f)

	sprite.SetCollisionCircle(65)
	sprite.Parent = nil
	this.addSprite(sprite)
	return sprite
}

func (this *Game) newAIShip(player Player, loc *Point) *Sprite {
	if loc == nil {
		loc = &Point{float64(this.width / 2), float64(this.height / 2)}
	}
	s := NewSprite(this, int32(messages.SpriteKindAiShip), *loc, 26, 40, Vector{0, 0}, 2e+2, 0, 100)
	s.Rotation = 3 * (math.Pi / 2)
	s.player = player
	player.SetShip(s)
	this.addSprite(s)
	s.SetCollisionCircle(20)
	return s
}

func (this *Game) newShip(player Player, loc *Point, kind SpriteKind) *Sprite {
	var health int32 = 100
	if kind == messages.SpriteKindAiShip {
		health = 20
	}

	if loc == nil {
		loc = &Point{float64(this.width / 2), float64(this.height / 2)}
	}
	s := NewSprite(this, int32(kind), *loc, 26, 40, Vector{0, 0}, 2e+2, 0, health)
	s.Rotation = 3 * (math.Pi / 2)
	s.player = player
	player.SetShip(s)
	this.addSprite(s)
	s.SetCollisionCircle(20)

	if player.GetPlayerType() == HUMAN_PLAYER {
		s.SetState(PHANTOM_MODE, 4000)
	}
	//s.SetState(PhantomMode, 4000)
	//	s.addProperty("PhantomMode", "on")
	return s
}

func (this *Game) PlaySound(typ SoundType, vol float32, player Player) {
	msg := NewPlaySoundMessage(typ, float64(vol))
	if player == nil {
		this.BroadcastMessage(msg)
	} else {
		player.UpdateWithBytes(msg)
	}
}

func (this *Game) BroadcastMessage(msg []byte) {
	for p := range this.Players {
		p.UpdateWithBytes(msg)
	}
}

func (this *Game) freezeDrawing() {
//	for p := range this.Players {
//		p.Update(NewFreezeDrawingMessage())
//		this.Frozen = true
//	}
}

func (this *Game) draw(cmds []string) {
//	for p := range this.Players {
//		m := NewDrawMessage(cmds)
//		p.Update(m)
//	}
}

func (this *Game) shakeSprite(sprite *Sprite) {
	msg := NewShakeMessage(sprite.Id, 14)
	this.BroadcastMessage(msg)
}

func (this *Game) UpdatePlayer(p Player) {
	p.UpdateWithBytes(NewPlayerUpdate(p))
}

func (this *Game) GameLoop() {

	var counter int = 1
	lastFrame := time.Now().UnixNano()
	randomEventTime := time.Now().UnixNano()

	for {

		startTime := time.Now().UnixNano()
		//if this.T == 0 {
		//	this.T = startTime
		//	this.lastFrame = startTime
		//}

		if counter%100 == 0 {
			Log.WithFields(logrus.Fields{"spriteCount": len(this.Sprites)}).Info("Game Tick")
			counter = 0
		}
		counter += 1
		fmt.Printf("FRAME = %v\n", this.frame)

		elapsed := startTime - lastFrame

		this.processInput()

		delta := float64(elapsed) / float64(time.Second)
		//	fmt.Printf("GameLoop start = %vns, frameTime = %vns\n", startTime, frameTime)

		//this.T += frameTime
		this.frame += 1

		this.compute(delta)

		sprites := this.CopySprites()
		//pupdate := NewPhysicsUpdate(this.frame, sprites)

		//		this.mutex.Unlock()

		for player := range this.Players {

			if (player.GetPlayerType() == AI_PLAYER) {
				continue;
			}
			physicsUpdate := NewPhysicsUpdate(
				this.frame, startTime, player.GetActionId(), sprites, player)

			player.UpdateWithBytes(physicsUpdate)
		}

		if (startTime - randomEventTime) > (20 * (int64(time.Second))) {
			this.randomEvent()
			randomEventTime = time.Now().UnixNano()
		}

		sleepTime := int64(((1 / float64(FPS)) * float64(time.Second))) - (time.Now().UnixNano() - startTime)
		lastFrame = startTime

		fmt.Printf("sleep time = %d\n", sleepTime)
		time.Sleep(time.Duration(sleepTime))

	}

}

func (this *Game) processInput() {

	for {

		select {

		case cmdHolder := <-this.PlayerCommands:
			this.HandlePlayerCommand(cmdHolder.Player, cmdHolder.Cmd)
			break
		default:
			return
		}
	}

}

func (this *Game) randomEvent() {

	r := random.Intn(40)
	this.mutex.Lock()
	switch r {
	case 1:
		astCount, astPresent := this.spriteCounts[messages.SpriteKindLargeAsteroid]
		prizeCount, _ := this.spriteCounts[messages.SpriteKindPrize]
		if !astPresent || astCount <= prizeCount {
			this.newLargeAsteroid(random.Intn(3))
		} else {
			rt := random.Intn(4)
			switch rt {
			case 1:
				_ = this.newPrizeSprite(messages.PlayerResourceTypeShield, 2)
				break
			case 2:
				_ = this.newPrizeSprite(messages.PlayerResourceTypeHyperdrive, 2)
				break
			case 3:
				_ = this.newPrizeSprite(messages.PlayerResourceTypeBooster, 2)
				break
			case 4:
				_ = this.newPrizeSprite(messages.PlayerResourceTypeLife, 2)
				break
			}
		}
	default:
		break
	}
	this.mutex.Unlock()
}

var computeCount int = 0

func (this *Game) hitBoundary(s *Sprite) *float64 {

	adjusted := s.Position.Add(Vector{s.Rotation, 10})
	var normal *float64
	if adjusted.x < 5 {
		n := math.Pi
		normal = &n
	} else if adjusted.x > float64(this.width)-5 {
		n := float64(0)
		normal = &n
	} else if adjusted.y < 5 {
		n := math.Pi / 2
		normal = &n
	} else if adjusted.y > float64(this.height)-5 {
		n := 3 * (math.Pi / 2)
		normal = &n
	}
	return normal
}

func (this *Game) CopySprites() []*Sprite {
	this.spritesMutex.Lock()

	var sprites []*Sprite = make([]*Sprite, 0, 0)

	for k, _ := range this.Sprites {
		if !k.IsDead() {
			sprites = append(sprites, k)
		}
	}
	this.spritesMutex.Unlock()

	return sprites
}

func (this *Game) compute(delta float64) {

	defer TimeCall(time.Now(), "game.compute()")
	computeCount += 1

	//this.spritesMutex.Lock()

	for sprite, _ := range this.Sprites {
		if sprite.IsDead() {
			fmt.Printf("compute: found Dead sprite %v, deleting\n", sprite)

			if (sprite.GetKind() == messages.SpriteKindAiShip) {
				  this.newAIPlayer()
			}
			this.RemoveSprite(sprite)

			continue
		}
		sprite.move(delta)
	}

	//this.spritesMutex.Unlock()

	sprites := this.CopySprites()

	pairs := make(map[[2]int]bool)

	for _, spriteA := range sprites {

		if spriteA.IsDead() {
			continue
		}

		normal := this.hitBoundary(spriteA)
		if normal != nil {
			spriteA.boundaryBounce(*normal)
			break
		}

		for _, spriteB := range sprites {

			if spriteA.IsDead() {
				break
			}
			if spriteB.IsDead() {
				continue
			}

			if spriteA == spriteB {
				continue
			}

			pair := [...]int{
				int(math.Min(float64(spriteA.Id), float64(spriteB.Id))),
				int(math.Max(float64(spriteA.Id), float64(spriteB.Id))),
			}

			_, present := pairs[pair]

			if present {
				continue
			}
			pairs[pair] = true

			kindA := spriteA.GetKind()
			kindB := spriteB.GetKind()

			if spriteA.Parent == spriteB || spriteB.Parent == spriteA {
				continue
			}
			if (kindA == messages.SpriteKindBullet && kindB == messages.SpriteKindBullet) &&
				(spriteA.Parent == spriteB.Parent) {
				continue
			}

			spriteA.PullOn(spriteB)

			if spriteA.intersects(spriteB) {
				result := this.handleCollision(spriteA, spriteB)

				switch result {
				case true:
					continue
				case false:
					return
				}

			}

		}

	}

}

func (this *Game) handleCollision(spriteA *Sprite, spriteB *Sprite) bool {


	actionA := this.GetAction(spriteA, spriteB)
	actionB := this.GetAction(spriteB, spriteA)

	switch actionA {

	case Damage:
		spriteA.inflictDamage(spriteB)
		this.shakeSprite(spriteB)
		break

	case Eat:
		spriteA.gobble(spriteB)
		break

	case Annihilate:
		spriteA.annihilate(spriteB)
		break

	case Transport:
		spriteA.transport(spriteB)
		break

	case Bounce:
		if (actionA != Bounce) {
		     spriteA.BounceOff(spriteB)
		}
		break

	}

	switch actionB {

	case Damage:
		spriteB.inflictDamage(spriteA)
		this.shakeSprite(spriteA)
		break

	case Eat:
		spriteB.gobble(spriteA)
		break

	case Annihilate:
		spriteB.annihilate(spriteA)
		break

	case Transport:
		spriteB.transport(spriteA)
		break

	case Bounce:
		spriteB.BounceOff(spriteA)
		break
	}
	return true
}

func (this *Game) GetAction(spriteA *Sprite, spriteB *Sprite) CollisionResult {

	x := (spriteA.GetKind() >> 16) - 1
	y := (spriteB.GetKind() >> 16) - 1
	actionA := ActionMatrix[x][y]
	//actionB := ActionMatrix[spriteA][spriteB]

	if spriteA.GetKind() == messages.SpriteKindShip {
		shipState := spriteA.typeInfo & SHIP_STATE

		if shipState&PHANTOM_MODE != 0 {
			return None
		}
		if shipState&SHIELDS_ACTIVE != 0 {
			return Bounce
		}
	}
	if spriteB.GetKind() == messages.SpriteKindShip {
		shipState := spriteB.typeInfo & SHIP_STATE
		if shipState&PHANTOM_MODE != 0 {
			return None
		}
		if shipState&SHIELDS_ACTIVE != 0 {
			return Bounce
		}
	}

	return actionA

}

var ActionMatrix = [][]CollisionResult{
	{Damage, Damage, Damage, Annihilate, None, None, Eat, Damage, Damage, Eat, Damage},
	{Damage, Damage, Damage, Annihilate, None, None, Pass, Damage, Damage, None, Damage},
	{Damage, Damage, Damage, Annihilate, None, None, Pass, Damage, Damage, None, Damage},
	{Damage, Damage, Damage, Annihilate, None, None, Pass, Damage, Damage, None, Damage},
	{Transport, Annihilate, Transport, Annihilate, Annihilate, Transport, Transport, Transport, Transport, Transport, Transport},
	{Annihilate, Annihilate, Annihilate, Annihilate, None, Annihilate, Annihilate, Annihilate, Annihilate, None, Annihilate},
	{None, None, None, None, None, None, None, None, None, None, None},
	{Damage, Damage, Damage, Annihilate, None, Damage, Pass, None, None, None, Damage},
	{Damage, Damage, Damage, Annihilate, None, Damage, Pass, None, None, None, Damage},
	{None, None, None, None, None, None, None, None, None, None, None},
	{Damage, Damage, Damage, Annihilate, None, None, None, Damage, Damage, None, Damage}}

/**
* This routine pulls messages off the broadcast channel and sends it to each registered player
 */


func (this *Game) HandlePlayerCommand(player Player, cmd *messages.PlayerCommandMessage) {

	fmt.Printf("HandlePlayerCommand: %v for player %d\n",
		cmd.Cmd(), player.GetPlayerId());

	sprite := this.Players[player]
	val := cmd.Value()
	player.SetActionId(int(cmd.ActionId()))

	//fmt.Printf("handling CmdType = %s,  ActionId  %v for player %v\n", CommandType_name[int32(cmd.Cmd)], cmd.ActionId, player.GetPlayerId())
	playerType := player.GetPlayerType()

	switch cmd.Cmd() {


	case messages.CommandTypeThrust:

		f := NewLinearForce(messages.ForceTypeThrustForce, sprite.Rotation, val, 100*1000*1000)
		if (playerType == HUMAN_PLAYER) {
			fmt.Printf("Thrust command: %v\n", val)
			fmt.Printf("Command Force: %v\n", f)
		}
		sprite.AddForce(f)
		sprite.SetState(JETS_ON, 1000)
		//sprite.addPropertyWithTtl("JetsOn", "true", 1000)
		//sprite.AddTempState("Thrusting", "true", 1000)
		break

	case messages.CommandTypeHyperspace:

		if player.HasResource(messages.PlayerResourceTypeHyperdrive) >= 1 {
			sprite.warp()
			_ = player.DepleteResource(messages.PlayerResourceTypeHyperdrive, 1)
			this.UpdatePlayer(player)
			sprite.SetState(PHANTOM_MODE, 4000)
			//	sprite.addPropertyWithTtl("PhantomMode", "true", 4000)
		} else {
		}
		break

	case messages.CommandTypeCloakShip:

		if player.HasResource(messages.PlayerResourceTypeCloak) >= 1 {
			_ = player.DepleteResource(messages.PlayerResourceTypeCloak, 1)
			this.UpdatePlayer(player)
			sprite.SetState(CLOAK_MODE, 7000)
			//	sprite.addPropertyWithTtl("PhantomMode", "true", 4000)
		} else {
		}
		break

	case messages.CommandTypeBoost:

		if player.HasResource(messages.PlayerResourceTypeBooster) >= 1 {
			f := NewLinearForce(messages.ForceTypeThrustForce, sprite.Rotation, val, 100*1000*1000)
			sprite.AddForce(f)
			sprite.SetState(JETS_ON, 1000)
			//sprite.addPropertyWithTtl("JetsOn", "true", 1000)
			_ = player.DepleteResource(messages.PlayerResourceTypeBooster, 1)
			this.UpdatePlayer(player)
		} else {
		}
		break

	case messages.CommandTypeRotate:
		sprite.rotate(val)
		break

	case messages.CommandTypeFire:
		this.fire(sprite, float64(this.bulletSpeed))
		break

	case messages.CommandTypePhaser:
		this.phaser(sprite, val)
		break

	case messages.CommandTypeShieldOn:
		if player.HasResource(messages.PlayerResourceTypeShield) >= 1 {
			sprite.Resize(44, 44)
			sprite.SetCollisionCircle(22)
			sprite.SetState(SHIELDS_ACTIVE, 1000)
			//sprite.addPropertyWithTtl("ShieldActive", "true", 1000)
			_ = player.DepleteResource(messages.PlayerResourceTypeShield, 1)
			this.UpdatePlayer(player)

		}
		break

	case messages.CommandTypeShieldOff:
		sprite.Resize(40, 26)
		//	sprite.SetCollisionRectangle()
		sprite.SetCollisionCircle(20)
		//sprite.RemoveState(ShieldActive)
		sprite.ClearState(SHIELDS_ACTIVE)
		//sprite.removeProperty("ShieldActive")
		break

	case messages.CommandTypeSetBulletSpeed:
		this.bulletSpeed = int(val)
		break

	case messages.CommandTypeSetBlackholeMass:
		this.blackhole.Mass = float64(val)
		break

	case messages.CommandTypeTractorOn:
		//fmt.Printf("TractorOn\n")
		sprite.Tractor(10)
		break

	case messages.CommandTypeTractorOff:
		sprite.TractorOff()
		break

	default:

	}
}

func (this *Game) String() string {
	return fmt.Sprintf("Id: %v", this.Id)
}

func (this *Game) random(max int, by int, avoid ...int) int {
	for {
		i := random.Intn(max)
		ok := true
		for _, a := range avoid {
			if math.Abs(float64(i-a)) < float64(by) {
				ok = false
			}
		}
		if ok {
			return i
		}
	}
}
