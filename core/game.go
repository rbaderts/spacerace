/**
 * A game is the runtime representation of an active game session
 * There may be 0 or more Players attached to the Game
 */
package core

import (
	"encoding/json"
	"fmt"
	//	jsonpb "github.com/golang/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
	"math"
	"math/rand"
	_ "os"
	"sync"
	"time"
)

var (
	MAX_QUEUE = 100
)

var TickRate int64 = 40 // Ticks per second
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
	Id             int
	Players        map[Player]*Sprite
	Sprites        map[*Sprite]bool
	PlayerCommands chan PlayerCommandHolder
	//	broadcast          chan ServerMessage
	updates            chan ServerMessage
	sounds             chan ServerMessage
	sendRoutineQuit    chan bool
	receiveRoutineQuit chan bool
	loopRoutineQuit    chan bool
	width              int
	height             int
	state              GameStateType
	T                  int64
	lastFrame          int64
	frame              int32
	mutex              *sync.Mutex
	bulletSpeed        int
	Race               *Race
	Frozen             bool
	spriteCounts       map[uint32]int
	blackhole          *Sprite
}

func (this *Game) IsRunning() bool {
	return this.state == Running
}

func (this *Game) Start() {
	this.T = 0
	if this.state != Running && len(this.Players) >= 1 {
		go this.Loop()
		this.state = Running
	}
}

func (this *Game) Stop() {

	fmt.Printf("Stop\n")
	for s, _ := range this.Sprites {
		delete(this.Sprites, s)
	}
	this.state = Waiting

	err := this.Race.UpdateRaceStatus(DB, RaceComplete)
	if err != nil {
		fmt.Printf("err = %v\n", err)
	}

}

func (this *Game) incrementSpriteCount(typ uint32, cnt int) {

	//	this.mutex.Lock()
	_, present := this.spriteCounts[typ]
	if !present {
		this.spriteCounts[typ] = cnt
	} else {
		this.spriteCounts[typ] = this.spriteCounts[typ] + cnt
	}
	//	this.mutex.Unlock()
}

func (this *Game) decrementSpriteCount(typ uint32, cnt int) {
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

func NewGame(id int) *Game {
	game := new(Game)
	game.state = Waiting
	game.Id = id
	game.Players = make(map[Player]*Sprite)
	game.Sprites = make(map[*Sprite]bool)
	game.PlayerCommands = make(chan PlayerCommandHolder, MAX_QUEUE)
	game.sendRoutineQuit = make(chan bool)
	game.receiveRoutineQuit = make(chan bool)
	game.loopRoutineQuit = make(chan bool)
	game.mutex = new(sync.Mutex)

	game.updates = make(chan ServerMessage, 10)
	game.sounds = make(chan ServerMessage, 10)
	game.spriteCounts = make(map[uint32]int)
	game.width = 5000
	game.height = 5000
	game.newLargeAsteroid(0)
	game.newLargeAsteroid(2)
	game.newAIPlayer()
	game.newAIPlayer()
	game.newAIPlayer()
	game.newBlackhole(9e+13)

	game.newStar(24, 2e+13)
	game.newStar(30, 2e+13)
	game.newPlanet(44, 2e+11)
	game.frame = 1

	Log.WithFields(logrus.Fields{"gameId": id}).Info("New Game")

	game.bulletSpeed = 170
	game.Frozen = false
	return game
}

var _ msgpack.CustomEncoder = (*Game)(nil)

func (this *Game) EncodeMsgpack(enc *msgpack.Encoder) error {
	sprites := make([]*Sprite, 0, len(this.Sprites))
	for s := range this.Sprites {
		sprites = append(sprites, s)
	}
	return enc.Encode(sprites)
}

func (this Game) MarshalJSON() ([]byte, error) {

	this.mutex.Lock()
	sprites := make([]*Sprite, 0, len(this.Sprites))
	for s := range this.Sprites {
		sprites = append(sprites, s)
	}
	this.mutex.Unlock()

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

	kind := s.GetKind()

	fmt.Printf("RemoveSprite: typeInfo = 0x%x, kind = 0x%x\n", s.GetTypeInfo(), kind)

	//	this.mutex.Lock()
	if kind == LARGE_ASTEROID {
		this.newSmallAsteroids(s.Position)
		this.decrementSpriteCount(LARGE_ASTEROID, 1)
	} else if kind == SMALL_ASTEROID {
		this.decrementSpriteCount(SMALL_ASTEROID, 1)
	} else if kind == PRIZE {
		this.decrementSpriteCount(PRIZE, 1)
	} else if kind == SHIP {
		if s.player.GetPlayerType() == HUMAN_PLAYER &&
			s.player.HasResource(LifeEnergyResource) >= 5 {
			newShip := this.newShip(s.player, nil)
			this.Players[s.player] = newShip
			s.player.DepleteResource(LifeEnergyResource, 10)
			this.updatePlayer(s.player)
		} else {
			s.player.Update(NewPlayerDead(s.player))
		}
	}
	delete(this.Sprites, s)
	//	this.mutex.Unlock()
}

func (this *Game) Join(p Player) {
	Log.WithFields(logrus.Fields{"player": p.GetName(), "game": this.Id}).Info("Player joined game")
	ship := this.newShip(p, nil)
	this.Players[p] = ship

	p.Update(NewInitializePlayer(int32(p.GetPlayerId()), int32(ship.Id)))
	this.updatePlayer(p)
	this.Start()
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

func (this *Game) fire(source *Sprite, val float64) {
	bullet := this.newBullet(source, val)
	bullet.VelocityLimit = 500
	this.Sprites[bullet] = true
}

func (this *Game) phaser(source *Sprite, val float64) {
	//	phaser := this.newPhaser(source, val)
	//	this.Sprites[phaser] = true
}

func (this *Game) newBullet(source *Sprite, velocity float64) *Sprite {

	velocityVec := UnitVector(source.Rotation).Mul(velocity)
	netVelocity := velocityVec.Add(source.Velocity)

	var x float64 = float64(source.Width) / 2
	var y float64 = 0

	newX := source.Position.x + x*math.Cos(source.Rotation) - y*math.Sin(source.Rotation)
	newY := source.Position.y + x*math.Sin(source.Rotation) + y*math.Cos(source.Rotation)

	sprite := NewSprite(this, BULLET, Point{newX, newY}, 6, 6, *netVelocity, 1, 10)
	sprite.Parent = source
	this.Sprites[sprite] = true
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

	sprite := NewSprite(this, STAR,
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		radius*2, radius*2, Vector{0, 0}, mass, 0)
	sprite.SetCollisionCircle(float64(radius) - (.20 * float64(radius)))
	sprite.Parent = nil
	this.Sprites[sprite] = true
	return sprite
}

func (this *Game) newPlanet(radius int, mass float64) *Sprite {

	sprite := NewSprite(this, PLANET,
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		radius*2, radius*2, Vector{0, 0}, mass, 0)
	sprite.SetCollisionCircle(float64(radius) - (.20 * float64(radius)))
	sprite.Parent = nil
	this.Sprites[sprite] = true
	return sprite
}

func (this *Game) newBlackhole(mass float64) *Sprite {

	sprite := NewSprite(this, BLACKHOLE,
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		40, 40, Vector{0, 0}, mass, 0)
	sprite.SetCollisionCircle(1)
	sprite.Parent = nil
	this.Sprites[sprite] = true
	this.blackhole = sprite
	return sprite
}

func (this *Game) newAIPlayer() {

	player := NewAIPlayer(this)
	sprite := this.newShip(player, &Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))})

	this.Players[player] = sprite
	this.Sprites[sprite] = true
	player.SetShip(sprite)

}

func (this *Game) newLargeAsteroid(quadrant int) *Sprite {
	velocity := Vector{float64(random.Intn(20) - 10), float64(random.Intn(20) - 10)}

	sprite := NewSprite(this, LARGE_ASTEROID,
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		30, 30, velocity, 5e+9, 0)

	f := NewAngularForce(InitialForce, 50*rand.Float64()-0.5, 100*1000*1000)
	sprite.AddForce(f)

	sprite.SetCollisionCircle(15)
	this.Sprites[sprite] = true
	this.incrementSpriteCount(LARGE_ASTEROID, 1)
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
	s1 := NewSprite(this, SMALL_ASTEROID, p1,
		14, 14, velocity, 5e+5, 0)

	f := NewAngularForce(InitialForce, 50*rand.Float64()-0.5, 100*1000*1000)
	s1.AddForce(f)

	s1.SetCollisionCircle(7)
	this.Sprites[s1] = true

	velocity = Vector{float64(-(xRand - 10)), float64(-(yRand - 10))}
	p2 := pos
	p2.x = p2.x + float64(-(xRand - 10))
	p2.y = p2.y + float64(-(yRand - 10))
	s2 := NewSprite(this, SMALL_ASTEROID, p2,
		14, 14, velocity, 5e+5, 0)

	f = NewAngularForce(InitialForce, 50*rand.Float64()-0.5, 100*1000*1000)
	s2.AddForce(f)

	s2.SetCollisionCircle(7)
	this.Sprites[s2] = true

	this.incrementSpriteCount(SMALL_ASTEROID, 2)

	return s2
}

func (this *Game) newPrizeSprite(prizeType PrizeType, value int) *Sprite {

	spriteType := PRIZE
	spriteType |= uint32(prizeType)
	spriteType |= uint32(value)

	s := NewSprite(
		this, spriteType,
		Point{float64(random.Intn(this.width)), float64(random.Intn(this.height))},
		26, 26, Vector{0, 0}, 2e+2, 0)

	//p := &PrizeTokenSprite{Sprite: *s, value: value}
	/*
		p := "Unknown"
		switch resource {
		case LifeEnergyResource:
			p = "LifeEnergy"
			break
		case ShieldResource:
			p = "Shield"
			break
		case HyperspaceResource:
			p = "Hyperspect"
			break
		case BoosterResource:
			p = "Booster"
			break

		}
	*/
	//	s.addProperty("Prize", fmt.Sprintf("%v", resource))
	//	s.addProperty("PrizeValue", fmt.Sprintf("%v", value))

	s.SetCollisionCircle(15)
	this.Sprites[s] = true

	this.incrementSpriteCount(PRIZE, 1)

	//	prize := &Prize{resource, value}
	//	s.prize = prize
	return s

}

func (this *Game) newShip(player Player, loc *Point) *Sprite {

	if loc == nil {
		loc = &Point{float64(this.width / 2), float64(this.height / 2)}
	}
	s := NewSprite(this, SHIP, *loc, 26, 40, Vector{0, 0}, 2e+2, 0)
	s.Rotation = 3 * (math.Pi / 2)
	s.player = player
	//	player.Ship = s
	player.SetShip(s)
	this.Sprites[s] = true
	//	s.SetCollisionInset(Inset{2, 2, 6, 0})
	//s.SetCollisionRectangle()
	s.SetCollisionCircle(20)

	if player.GetPlayerType() == HUMAN_PLAYER {
		s.SetState(PHANTOM_MODE, 4000)
	}
	//s.SetState(PhantomMode, 4000)
	//	s.addProperty("PhantomMode", "on")
	return s
}

func (this *Game) playSound(typ SoundType, vol float32, player Player) {
	for p := range this.Players {
		p.Update(NewPlaySoundMessage(typ, float64(vol)))
	}
}

func (this *Game) freezeDrawing() {
	for p := range this.Players {
		p.Update(NewFreezeDrawingMessage())
		this.Frozen = true
	}
}

func (this *Game) draw(cmds []string) {
	for p := range this.Players {
		m := NewDrawMessage(cmds)
		p.Update(m)
	}
}

func (this *Game) updatePlayer(p Player) {
	p.Update(NewPlayerUpdate(p))
}

const NANOS_PER_FRAME = 16666666

func (this *Game) Loop() {

	ticker := time.NewTicker(time.Duration(int64(time.Second) / TickRate))
	defer ticker.Stop()

	randomEvents := time.NewTicker(time.Duration(20 * (int64(time.Second) / TickRate)))
	defer randomEvents.Stop()

	counter := 0
	quit := false
	for {

		if this.Frozen {
			break
		}

		select {

		case cmdHolder := <-this.PlayerCommands:
			this.HandlePlayerCommand(cmdHolder.Player, &cmdHolder.Cmd)

		case <-randomEvents.C:

			r := random.Intn(40)
			switch r {
			case 1:
				//				this.mutex.Lock()
				astCount, astPresent := this.spriteCounts[LARGE_ASTEROID]
				prizeCount, _ := this.spriteCounts[PRIZE]
				if !astPresent || astCount <= prizeCount {
					this.newLargeAsteroid(random.Intn(3))
				} else {
					rt := random.Intn(4)
					switch rt {
					case 1:
						_ = this.newPrizeSprite(SHIELD, 2)
						break
					case 2:
						_ = this.newPrizeSprite(HYPERSPACE, 2)
						break
					case 3:
						_ = this.newPrizeSprite(BOOSTER, 2)
						break
					case 4:
						_ = this.newPrizeSprite(LIFEENERGY, 2)
						break
					}
				}
				//				this.mutex.Unlock()
			default:
				break
			}

		case <-ticker.C:

			if counter%100 == 0 {
				Log.WithFields(logrus.Fields{"spriteCount": len(this.Sprites)}).Info("Game Tick")
				counter = 0
			}
			counter += 1

			timestamp := time.Now().UnixNano()
			if this.T == 0 {
				this.T = timestamp
				this.lastFrame = timestamp
			}

			frameTime := timestamp - this.lastFrame
			this.lastFrame = timestamp

			delta := float64(frameTime) / float64(time.Second)
			this.compute(delta)

			this.T += frameTime
			this.frame += 1

			pupdate := NewPhysicsUpdate(this.frame, this.Sprites)
			for player := range this.Players {
				if pupdate.Update != nil {
					pupdate.Update.SetActionid(player.GetActionId())
				}
				//				pupdate.Update.OriginX = player.Ship.Position.x + (float64(this.width) / 2)
				//				pupdate.Update.OriginY = player.Ship.Position.y + (float64(this.height) / 2)
				player.Update(pupdate)
			}

		case <-this.loopRoutineQuit:
			quit = true

		default:
			if quit == true {
				return
			}

		}
	}

}

var computeCount int = 0

func (this *Game) hitBoundary(s *Sprite) *float64 {

	adjusted := s.Position.Add(&Vector{s.Rotation, 10})
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

func (this *Game) compute(delta float64) {

	defer TimeCall(time.Now(), "game.compute()")
	computeCount += 1
	for sprite, _ := range this.Sprites {
		if sprite.isDead() {
			this.RemoveSprite(sprite)
			continue
		}
		sprite.move(delta)
	}

	var pairs map[[2]int]bool = make(map[[2]int]bool)

	for spriteA, _ := range this.Sprites {

		normal := this.hitBoundary(spriteA)
		if normal != nil {
			spriteA.boundaryBounce(*normal)
			break
		}

		for spriteB, _ := range this.Sprites {

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

			if kindA == BLACKHOLE || kindA == STAR {
				spriteA.PullOn(spriteB)
			} else if kindB == BLACKHOLE || kindB == STAR {
				spriteB.PullOn(spriteA)
			}

			if spriteA.Parent == spriteB || spriteB.Parent == spriteA {
				continue
			}

			// If they're bullets from the same guy

			if (kindA == BULLET && kindB == BULLET) &&
				(spriteA.Parent == spriteB.Parent) {
				continue
			}

			if spriteA.intersects(spriteB) {
				result := this.handleCollision(spriteA, spriteB)

				switch result {
				case false:
					continue
				case true:
					return
				}

			}

		}

	}

}

func (this *Game) handleCollision(spriteA *Sprite, spriteB *Sprite) bool {

	kindA := spriteA.GetKind()
	kindB := spriteB.GetKind()

	//fmt.Printf("COLLISION: spriteA : %v\n", spriteA)
	//fmt.Printf("           spriteB : %v\n", spriteB)

	if kindA == SHIP {

		shipState := spriteA.typeInfo & SHIP_STATE

		//if spriteA.HasState(PhantomMode) {
		//if spriteA.hasProperty("PhantomMode") {
		if shipState&PHANTOM_MODE != 0 {
			return false
		}
		//if spriteA.HasState(ShieldActive) {
		///if spriteA.hasProperty("ShieldActive") {
		if shipState&SHIELDS_ACTIVE != 0 {
			spriteA.BounceOff(spriteB)
			this.playSound(BoingSound, 0.5, nil)
			return false
		}
		if spriteB.isPrize() {
			spriteA.gobble(spriteB)
			this.playSound(BloopSound, 0.5, nil)
			this.RemoveSprite(spriteB)
			this.updatePlayer(spriteA.player)
			return false
		}
		if kindB == BLACKHOLE {
			spriteA.warp()
			return false
		}
	}

	if kindB == SHIP {

		shipState := spriteB.typeInfo & SHIP_STATE

		//if spriteA.HasState(PhantomMode) {
		//if spriteA.hasProperty("PhantomMode") {
		if shipState&PHANTOM_MODE != 0 {
			return false
		}
		//if spriteA.HasState(ShieldActive) {
		//if spriteA.hasProperty("ShieldActive") {
		if shipState&SHIELDS_ACTIVE != 0 {
			spriteA.BounceOff(spriteB)
			this.playSound(BoingSound, 0.5, nil)
			return false
		}
		if spriteA.isPrize() {
			spriteB.gobble(spriteA)
			this.RemoveSprite(spriteA)
			this.updatePlayer(spriteB.player)
			return true
		}
		if kindA == BLACKHOLE {
			spriteB.warp()
			return false
		}
	}

	result := false
	if kindA != BLACKHOLE && kindA != STAR {
		this.RemoveSprite(spriteA)
		result = true
	}
	if kindB != BLACKHOLE && kindB != STAR {
		this.RemoveSprite(spriteB)
	}
	return result
}

/**
* This routine pulls messages off the broadcast channel and sends it to each registered player
 */

func (this *Game) Communicate() {

	//ticker := time.NewTicker(time.Second / 60)
	//defer ticker.Stop()

	quit := false
	for {
		//quit := false
		select {

		case <-this.receiveRoutineQuit:
			quit = true

		default:
			if quit == true {
				return
			}

		}

	}
}

func (this *Game) SendRoutine() {

	//ticker := time.NewTicker(time.Second / 60)
	//defer ticker.Stop()

	quit := false
	for {

		select {

		case message := <-this.sounds:

			for player := range this.Players {
				//if message.Update != nil {
				//		message.Update.SetActionid(player.ActionId)
				//	}
				//	player.GetMutex().Lock()
				player.Update(message)
				//	player.GetMutex().Unlock()
			}

		case message := <-this.updates:

			for player := range this.Players {
				if message.Update != nil {
					message.Update.SetActionid(player.GetActionId())
				}
				player.Update(message)
			}

		case <-this.sendRoutineQuit:
			Log.Info("Game SendRoutine quit")
			quit = true
		default:
			if quit == true {
				return
			}
		}

	}
}

func (this *Game) HandlePlayerCommand(player Player, cmd *PlayerCommandMessage) {

	sprite := this.Players[player]
	val := cmd.Value
	player.SetActionId(int(cmd.ActionId))

	//fmt.Printf("handling CmdType = %s,  ActionId  %v for player %v\n", CommandType_name[int32(cmd.Cmd)], cmd.ActionId, player.GetPlayerId())

	switch cmd.Cmd {

	case Thrust:

		f := NewLinearForce(ThrustForce, sprite.Rotation, val, 100*1000*1000)
		sprite.AddForce(f)
		sprite.SetState(JETS_ON, 1000)
		//sprite.addPropertyWithTtl("JetsOn", "true", 1000)
		//sprite.AddTempState("Thrusting", "true", 1000)
		break

	case Hyperspace:

		if player.HasResource(HyperspaceResource) >= 1 {
			sprite.warp()
			_ = player.DepleteResource(HyperspaceResource, 1)
			this.updatePlayer(player)
			sprite.SetState(PHANTOM_MODE, 4000)
			//	sprite.addPropertyWithTtl("PhantomMode", "true", 4000)
		} else {
		}
		break

	case Cloak:

		if player.HasResource(CloakResource) >= 1 {
			_ = player.DepleteResource(CloakResource, 1)
			this.updatePlayer(player)
			sprite.SetState(CLOAK_MODE, 7000)
			//	sprite.addPropertyWithTtl("PhantomMode", "true", 4000)
		} else {
		}
		break

	case Booster:

		if player.HasResource(BoosterResource) >= 1 {
			f := NewLinearForce(ThrustForce, sprite.Rotation, val, 100*1000*1000)
			sprite.AddForce(f)
			sprite.SetState(JETS_ON, 1000)
			//sprite.addPropertyWithTtl("JetsOn", "true", 1000)
			_ = player.DepleteResource(BoosterResource, 1)
			this.updatePlayer(player)
		} else {
		}
		break

	case Rotate:
		sprite.rotate(val)
		break

	case Fire:
		this.fire(sprite, float64(this.bulletSpeed))
		break

	case Phaser:
		this.phaser(sprite, val)
		break

	case ShieldOn:
		if player.HasResource(ShieldResource) >= 1 {
			sprite.Resize(44, 44)
			sprite.SetCollisionCircle(22)
			sprite.SetState(SHIELDS_ACTIVE, 1000)
			//sprite.addPropertyWithTtl("ShieldActive", "true", 1000)
			_ = player.DepleteResource(ShieldResource, 1)
			this.updatePlayer(player)

		}
		break

	case ShieldOff:
		sprite.Resize(40, 26)
		//	sprite.SetCollisionRectangle()
		sprite.SetCollisionCircle(20)
		//sprite.RemoveState(ShieldActive)
		sprite.ClearState(SHIELDS_ACTIVE)
		//sprite.removeProperty("ShieldActive")
		break

	case SetBulletSpeed:
		this.bulletSpeed = int(val)
		break

	case SetBlackholeMass:
		this.blackhole.Mass = float64(val)
		break

	case TractorOn:
		//fmt.Printf("TractorOn\n")
		sprite.Tractor(10)
		break

	case TractorOff:
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
