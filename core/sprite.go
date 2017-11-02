package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/sirupsen/logrus"
	"math"
	"strconv"
	"sync"
	"time"
)

var (
	SpriteCounter                = 0
	FORCE_COEFFICIENT    float64 = 1000
	DISTANCE_COEFFICIENT float64 = 100
)

func init() {

}

type ShapeType int

const (
	_ ShapeType = iota
	CircleShape
	PolygonShape
	LineShape
)

/**
 * Sprite type is a complex 32 bit field.  The top 4 bits only are used to represent
 * the core type of the sprite (kind).   The remaining bits are used to represents various
 * variations or states of a sprite, specific for each type of sprite.
 *
 */

/*
 * Ship
 *
 *   Ship states:   Jets On, Shields On, Phantom Mode
 *
 */

//type SpriteType uint32

const (
	SPRITE_KIND uint32 = 0xFF0000
	SHIP_STATE  uint32 = 0x0000FF
	PRIZE_TYPE  uint32 = 0x00FF00
	PRIZE_VALUE uint32 = 0x0000FF
)

const (
	SHIP           uint32 = 0x010000
	LARGE_ASTEROID uint32 = 0x020000
	SMALL_ASTEROID uint32 = 0x030000
	BULLET         uint32 = 0x040000
	BLACKHOLE      uint32 = 0x050000
	STAR           uint32 = 0x060000
	PRIZE          uint32 = 0x070000
	PLANET         uint32 = 0x080000
)

/**
 * For Ship Sprites:    0x000000FF = State
 */

type ShipState uint32

const (
	_              ShipState = iota
	JETS_ON                  = 0x000001
	SHIELDS_ACTIVE           = 0x000002
	PHANTOM_MODE             = 0x000004
	CLOAK_MODE               = 0x000008
	TRACTOR_ACTIVE           = 0x000010
)

/**
 * For Prize Sprites:    0x000000FF = Prize Type,  0x0000FF00 = Prize Value
 */

type PrizeType uint32

const (
	_          PrizeType = iota
	SHIELD               = 0x000100
	BOOSTER              = 0x000200
	HYPERSPACE           = 0x000400
	LIFEENERGY           = 0x000800
	CLOAK                = 0x001000
	TRACTOR              = 0x002000
)

/*
var SpriteTypes = [...]string{
	"None",
	"Ship", "Asteroid", "Bullet",
	"Blackhole",
	"Explosion",
}

func (s SpriteType) String() string {
	return SpriteTypes[s]
}

func SpriteTypeFromString(s string) SpriteType {
	var r SpriteType
	for i, t := range SpriteTypes {
		if t == s {
			r = SpriteType(i)
			break
		}
	}
	return r
}
*/

type IdPair struct {
	a int
	b int
}

func NewPair(a int, b int) IdPair {
	if a < b {
		return IdPair{a, b}
	} else {
		return IdPair{b, a}
	}
}

type Force struct {
	Typ      ForceType
	Dir      float64
	Mag      float64
	duration int64
	start    int64
	ActionId int32
	angular  bool
}

/*
func NewAngularForce(value float64, duration int64) {
	f := new(Force)
	f.Dir = dir
	fir.Mag = value
	f.duration = duration
	f.start = time.Now().UnixNano()
	f.angular = true
}
*/

func NewLinearForce(typ ForceType, dir float64, value float64, duration int64) *Force {
	f := new(Force)
	f.Typ = typ
	f.Dir = dir
	f.Mag = value
	f.duration = duration
	f.start = time.Now().UnixNano()
	f.angular = false
	return f
}

func NewAngularForce(typ ForceType, value float64, duration int64) *Force {
	f := new(Force)
	f.Typ = typ
	f.Mag = value
	f.duration = duration
	f.start = time.Now().UnixNano()
	f.angular = true
	return f
}

func (f Force) String() string {
	return fmt.Sprintf("Force: Dir %f, Mag %f, dur: %d, start %d", f.Dir, f.Mag, f.duration, f.start)
}

type Sprite struct {
	Id              int
	typeInfo        uint32
	Position        Point // Center
	Velocity        Vector
	AngularVelocity float64
	AngularAccel    float64
	Accel           Vector
	Height          int
	Width           int

	Rotation float64 // (radians)/second

	player Player
	Age    int64 // in Nanoseconds

	Mass     float64
	Lifespan float64
	Vmax     int

	//lastMoved int64
	Parent *Sprite
	game   *Game
	Forces map[*Force]bool

	collisionShapeType ShapeType
	collisionRadius    float64
	//collisionInset     Inset
	//collisionCircle    Circle
	//collisionRect      Rectangle

	VelocityLimit float64

	// current (ship) states (Shield On,
	States map[SpriteStatus]int

	//prize *Prize

	// sprite type specific data
	//Properties map[string]string
	mutex *sync.Mutex

	Tractored *Sprite
}

func (this Sprite) GetKind() uint32 {
	kind := this.typeInfo & SPRITE_KIND
	return kind
}

func (this *Sprite) GetMutex() *sync.Mutex {
	return this.mutex
}

/*
func (this *Sprite) RemoveState(state SpriteStatus) {
	delete(this.States, state)
}

func (this *Sprite) HasState(state SpriteStatus) bool {
	for s, _ := range this.States {
		if s == state {
			return true
		}
	}
	return false
}
*/

func (this Sprite) GetTypeInfo() uint32 {
	return this.typeInfo
}

func (this *Sprite) HasState(state ShipState) bool {

	res := this.typeInfo & SHIP_STATE
	if res != 0 {
		return true
	}
	return false

}

func (this *Sprite) ClearState(state ShipState) {
	this.mutex.Lock()
	this.typeInfo &^= uint32(state)
	this.mutex.Unlock()
}

func (this *Sprite) SetState(state ShipState, ttlMillis int) {

	this.mutex.Lock()
	this.typeInfo |= uint32(state)
	this.mutex.Unlock()
	time.AfterFunc(time.Duration(int64(ttlMillis)*int64(time.Millisecond)),
		func() {
			this.ClearState(state)
		})

}

//func (this *Sprite) GetPolygonBody() *olygonBody {

//	NewPolygonBody(

//}

func NewSprite(g *Game, typ uint32, position Point, height int, width int, velocity Vector, mass float64, lifespan int32) *Sprite {

	SpriteCounter += 1
	s := new(Sprite)
	s.game = g
	s.Id = SpriteCounter
	s.typeInfo = typ
	s.Position = position
	s.Velocity = velocity
	s.Height = height
	s.Width = width
	s.Rotation = 0
	s.Age = 0
	s.Mass = mass
	s.Vmax = 200
	s.Parent = nil
	s.Lifespan = float64(lifespan)
	s.Forces = make(map[*Force]bool)
	s.collisionShapeType = PolygonShape
	s.collisionRadius = -1
	s.AngularVelocity = 0

	s.VelocityLimit = 300
	s.States = make(map[SpriteStatus]int)

	//jhifmt.Printf("NewSprite %d(%x), position=(%v, %v)\n", s.Id, s.typeInfo, s.Position.x, s.Position.y)
	//Log.WithFields(logrus.Fields{"id": s.Id, "typeInfo": fmt.Sprintf("0x%x", s.typeInfo)}).Info("NEW SPRITE")

	s.mutex = new(sync.Mutex)
	return s
}

/*
func (this *Sprite) addProperty(name string, value string) {
	this.Properties[name] = value
}

func (this *Sprite) addPropertyWithTtl(name string, value string, ttlMillis int) {

	this.addProperty(name, value)
	time.AfterFunc(time.Duration(int64(ttlMillis)*int64(time.Millisecond)),
		func() {
			this.removeProperty(name)
		})
}

func (this *Sprite) hasProperty(name string) bool {
	_, present := this.Properties[name]
	return present
}

func (this *Sprite) getProperty(name string) (string, error) {
	v, present := this.Properties[name]

	if present {
		return v, nil
	}
	return "", errors.New("Property not present")
}

func (this *Sprite) removeProperty(name string) {
	delete(this.Properties, name)
}
*/

func (this *Sprite) isPrize() bool {
	if (this.typeInfo & SPRITE_KIND) == PRIZE {
		//if this.Type == PrizeSprite {
		return true
	}
	return false
}

func (this *Sprite) GetCollisionShape() CollisionShape {
	switch this.collisionShapeType {

	case CircleShape:
		return &Circle{this.Position, this.collisionRadius}

	case PolygonShape:
		poly := this.polygon()
		poly.Rotate(this.Rotation, &this.Position)
		return poly
	}
	return nil
}

func (this *Sprite) SetCollisionCircle(radius float64) {
	this.collisionShapeType = CircleShape
	this.collisionRadius = radius
}

/*
func (this *Sprite) GetCollisionCircle() *Circle {
	radius := this.collisionRadius
	if radius == -1 {
		radius = math.Max(float64(this.Height), float64(this.Width)) / 2
	}
	return &Circle{this.Position, radius}
}
*/

/*
func (this *Sprite) GetCollisionRectangle() *Rectangle {
	r := this.rectangle()
	return r.Inset(this.collisionInset)
}
*/

func (this *Sprite) SetCollisionRectangle() {
	this.collisionRadius = -1
	this.collisionShapeType = PolygonShape
}

func (this *Sprite) Intersects(line *LineSegment) bool {

	c := Circle{this.Position, this.collisionRadius}
	return LineCircleIntersects(line, &c)
}

func (this Sprite) String() string {
	str := new(bytes.Buffer)
	str.WriteString(fmt.Sprintf("id: %d, type: %x, pos: %v, vel: %v\n", this.Id, this.typeInfo, this.Position, this.Velocity))
	str.WriteString(fmt.Sprintf(" collisionShape: %v\n", this.GetCollisionShape()))
	return str.String()
}

/**
 * gets bounds as Bounds (TopLeft + BottomRight)
 */
/*
func (this *Sprite) bounds() *Bounds {
	return &Bounds{
		Point{this.Position.x - float64(this.Width)/2, this.Position.y - float64(this.Height)/2},
		Point{this.Position.x + float64(this.Width)/2, this.Position.y + float64(this.Height)/2}}

}
*/

/**
 * gets bounds as Rectangle (topLeft, + H/W)
 */

func (this *Sprite) polygon() *Polygon {
	return &Polygon{
		[]Point{
			Point{this.Position.x - float64(this.Width)/2, this.Position.y - float64(this.Height)/2},
			Point{this.Position.x + float64(this.Width)/2, this.Position.y - float64(this.Height)/2},
			Point{this.Position.x + float64(this.Width)/2, this.Position.y + float64(this.Height)/2},
			Point{this.Position.x - float64(this.Width)/2, this.Position.y + float64(this.Height)/2}}}
}

/*
func (this *Sprite) rectangle() *Rectangle {
	return &Rectangle{
		Point{this.Position.x - float64(this.Width)/2, this.Position.y - float64(this.Height)/2},
		float64(this.Width), float64(this.Height)}

}

*/
func (this *Sprite) intersects(other *Sprite) bool {

	res := TestCollision(this.GetCollisionShape(), other.GetCollisionShape())

	if res == true {
		fmt.Printf("sprite %d(%x) x sprte %d(%x) collides = %v\n", this.Id, this.typeInfo, other.Id, other.typeInfo, res)
	}

	return res
}

func (this *Sprite) rotate(val float64) {

	newVal := this.Rotation + val
	if newVal < 0 {
		this.Rotation = (math.Pi*2 + newVal)
	} else if newVal > 2*math.Pi {
		this.Rotation = newVal - (math.Pi * 2)
	} else {
		this.Rotation = newVal
	}
}

func (this *Sprite) accelerate(v *Vector) {
	this.Accel = *(this.Accel.Add(*v))
}

func (this *Sprite) accelerateWithRotation(val float64) {

	f := NewLinearForce(ThrustForce, this.Rotation, val, 0)
	this.AddForce(f)

}

func (this *Sprite) isDead() bool {

	if this.Lifespan == 0 {
		return false
	}
	if this.Age > int64(this.Lifespan)*int64(time.Millisecond) {
		return true
	}
	return false
}

func (this *Sprite) warp() *Sprite {
	x := random.Intn(this.game.width)
	y := random.Intn(this.game.height)
	this.Position = Point{float64(x), float64(y)}
	this.Velocity = Vector{0, 0}
	return this
}

func (this *Sprite) AddForce(force *Force) {
	this.Forces[force] = true
}

func (this *Sprite) applyForce(force *Force) {

	// F = m * a :     a = F/m
	//x := math.Sin(force.Dir) / this.Mass
	//y := math.Cos(force.Dir) / this.Mass

	if force.angular == true {
		this.AngularAccel += force.Mag
	} else {
		x := math.Cos(force.Dir) / this.Mass
		y := math.Sin(force.Dir) / this.Mass

		this.Accel.x += x * FORCE_COEFFICIENT * float64(force.Mag)
		this.Accel.y += y * FORCE_COEFFICIENT * float64(force.Mag)
	}
}

func (this *Sprite) move(delta float64) *Sprite {

	this.Accel.x = 0
	this.Accel.y = 0
	this.AngularAccel = 0

	now := time.Now().UnixNano()
	for force, b := range this.Forces {
		if b == false {
			continue
		}
		this.applyForce(force)
		if now-force.start > force.duration {
			delete(this.Forces, force)
		}
	}

	if this.Accel.NonZero() {
		this.Velocity = *(this.Velocity.Add(*(this.Accel.Mul(delta))))
		this.limitVelocity()
	}
	//	this.applyDrag(0.10, delta)

	if this.Velocity.NonZero() {
		this.Position = *(this.Position.Add(this.Velocity.Mul(delta)))
	}

	if this.AngularAccel != 0 {
		this.AngularVelocity += this.AngularAccel * delta
	}
	if this.AngularVelocity != 0 {
		this.rotate(this.AngularVelocity * delta)
	}

	this.Age += int64(delta * 1000 * 1000)

	return this
}

func (this *Sprite) limitVelocity() {
	if this.Velocity.x > this.VelocityLimit {
		this.Velocity.x = this.VelocityLimit
	} else if this.Velocity.x < (-this.VelocityLimit) {
		this.Velocity.x = -this.VelocityLimit
	}
	if this.Velocity.y > this.VelocityLimit {
		this.Velocity.y = this.VelocityLimit
	} else if this.Velocity.y < (-this.VelocityLimit) {
		this.Velocity.y = -this.VelocityLimit
	}
}

func (this *Sprite) applyDrag(factor float64, delta float64) {

	//fmt.Printf("drag:  start V = %v\n", this.Velocity)
	if this.Velocity.x > 0 {
		this.Velocity.x -= (factor * this.Velocity.x) * delta
		if this.Velocity.x < 0 {
			this.Velocity.x = 0
		}
	} else if this.Velocity.x < 0 {
		this.Velocity.x -= (factor * this.Velocity.x) * delta
		if this.Velocity.x > 0 {
			this.Velocity.x = 0
		}
	}

	if this.Velocity.y > 0 {
		this.Velocity.y -= (factor * this.Velocity.y) * delta
		if this.Velocity.y < 0 {
			this.Velocity.y = 0
		}
	} else if this.Velocity.y < 0 {
		this.Velocity.y -= (factor * this.Velocity.y) * delta
		if this.Velocity.y > 0 {
			this.Velocity.y = 0
		}
	}

	//fmt.Printf("drag:  end V = %v\n", this.Velocity)
}

func (this *Sprite) Resize(width int, height int) {

	//	offset := &Vector{float64((this.Width - width) / 2), float64((this.Height - height) / 2)}
	this.Height = height
	this.Width = width
	//	this.Position = *(this.Position.Add(offset))

}

func (this Sprite) MarshalJSON() ([]byte, error) {

	var playerId int = 0
	var actionId int = 0

	if this.player != nil {
		playerId = this.player.GetPlayerId()
		actionId = this.player.GetActionId()
	}

	b, err := json.Marshal(map[string]interface{}{
		"id":       this.Id,
		"typeInfo": this.typeInfo,
		"x":        this.Position.x,
		"y":        this.Position.y,
		"vx":       this.Velocity.x,
		"vy":       this.Velocity.y,
		"ax":       this.Accel.x,
		"ay":       this.Accel.y,
		"height":   this.Height,
		"width":    this.Width,
		"rotation": this.Rotation,
		"playerId": playerId,
		"actionId": actionId,
	})

	if err != nil {
		panic(err)
	}
	return b, err
}

func (this *Sprite) radius() float64 {
	return math.Min(float64(this.Height), float64(this.Width)) / float64(2)
}

func (this *Sprite) center() *Point {
	return &Point{this.Position.x, this.Position.y}
}

func (this *Sprite) distanceVector(other *Sprite) *Vector {
	c2 := this.center()
	c1 := other.center()
	dX := c1.x - c2.x
	dY := c1.y - c2.y

	return &Vector{dX, -dY}
}

func (this *Sprite) boundaryBounce(normal float64) {

	angle := math.Atan2(this.Velocity.y, this.Velocity.x)
	angle = 2*normal - math.Pi - angle
	mag := 0.9 * math.Hypot(this.Velocity.x, this.Velocity.y)

	//newX = math.Cos(angle) * mag
	//newY = math.Sin(angle) * mag
	this.Velocity.x = math.Cos(angle) * mag
	this.Velocity.y = math.Sin(angle) * mag

	if this.Position.x <= 15 {
		this.Position.x += 5
	}
	if this.Position.x >= float64(this.game.width)-15 {
		this.Position.x -= 5
	}
	if this.Position.y <= 15 {
		this.Position.y += 5
	}
	if this.Position.y >= float64(this.game.height)-15 {
		this.Position.y -= 5
	}

}

func (this *Sprite) BounceOff(other *Sprite) {

	s1Center := this.center()
	s2Center := other.center()

	collision_angle := math.Atan2((s2Center.y - s1Center.y), (s2Center.x - s1Center.x))
	s1Speed := this.Velocity.Length()
	s2Speed := other.Velocity.Length()
	s1Direction := this.Velocity.GetAngle()
	s2Direction := other.Velocity.GetAngle()

	newS1x := s1Speed * math.Cos(s1Direction-collision_angle)
	newS1y := s1Speed * math.Sin(s1Direction-collision_angle)

	newS2x := s2Speed * math.Cos(s2Direction-collision_angle)
	newS2y := s2Speed * math.Sin(s2Direction-collision_angle)

	finalS1_Vx := ((float64(this.Mass-other.Mass))*newS1x + (float64(other.Mass+other.Mass))*newS2x) / (float64(this.Mass + other.Mass))
	finalS2_Vx := ((float64(this.Mass+this.Mass))*newS1x + (float64(other.Mass-this.Mass))*newS2x) / (float64(this.Mass + other.Mass))
	finalS1_Vy := newS1y
	finalS2_Vy := newS2y

	cosAngle := math.Cos(collision_angle)
	sinAngle := math.Sin(collision_angle)

	newV1x := cosAngle*finalS1_Vx - sinAngle*finalS1_Vy
	newV1y := sinAngle*finalS1_Vx + cosAngle*finalS1_Vy
	newV2x := cosAngle*finalS2_Vx - sinAngle*finalS2_Vy
	newV2y := sinAngle*finalS2_Vx + cosAngle*finalS2_Vy

	posDiff := s1Center.Delta(s2Center)
	dis := posDiff.Length()

	mtd := posDiff.Mul(((this.radius() - other.radius() - dis) / dis))
	im1 := float64(1) / this.Mass
	im2 := float64(1) / other.Mass

	//	pos1 := *(s1Center.Sub(mtd.Mul((im1 / (im1 + im2)))))

	v := mtd.Mul((im1 / (im1 + im2)))
	pos1 := *(s1Center.Add(v.Inverse()))

	///	pos1 := *(s1Center.Sub(mtd.Mul((im1 / (im1 + im2)))))
	pos2 := *(s2Center.Add(mtd.Mul((im2 / (im1 + im2)))))

	this.Position.x = pos1.x
	this.Position.y = pos1.y
	other.Position.x = pos2.x
	other.Position.y = pos2.y

	this.Velocity.x = newV1x
	this.Velocity.y = newV1y
	other.Velocity.x = newV2x
	other.Velocity.y = newV2y
}

func (this *Sprite) DrawCollisionShape() []string {

	cmds := new([]string)
	*cmds = append(*cmds, fmt.Sprintf("lineStyle(1,4,1)"))

	if this.collisionShapeType == PolygonShape {
		/*
			p := NewPolygonBody(this.GetCollisionRectangle(), this.Rotation, &this.Position)
			//r := this.rectangle()
			//r = r.Inset(this.collisionInset)
			// [new PIXI.Point(x, y),new PIXI.Point(x,y)]

			cmdBuf := new(bytes.Buffer)
			cmdBuf.WriteString("drawPolygon(")
			var firstVertex *Point = nil
			for _, v := range p.vertices {
				if firstVertex == nil {
					firstVertex = v
				}
				c := fmt.Sprintf("new PIXI.Point(%v,%v),", v.x, v.y)
				cmdBuf.WriteString(c)
			}
			c := fmt.Sprintf("new PIXI.Point(%v,%v))", firstVertex.x, firstVertex.y)
			cmdBuf.WriteString(c)

			cmd := cmdBuf.String()
			fmt.Printf("draw poly command: %v\n", cmd)
			*cmds = append(*cmds, cmd)
		*/
	} else {
		*cmds = append(*cmds, fmt.Sprintf("drawCircle(%v,%v,%v)", this.Position.x, this.Position.y, this.collisionRadius))
	}
	return *cmds

}

func (this *Sprite) TractorOff() {
	this.ClearState(TRACTOR_ACTIVE)
	this.Tractored = nil
}

func (this *Sprite) Tractor(power float64) {

	this.SetState(TRACTOR_ACTIVE, 5000)

	var x float64 = float64(this.Width) / 2
	var y float64 = 0

	startX := this.Position.x + x*math.Cos(this.Rotation) - y*math.Sin(this.Rotation)
	startY := this.Position.y + x*math.Sin(this.Rotation) + y*math.Cos(this.Rotation)
	start := Point{startX, startY}

	v := UnitVector(this.Rotation)
	t := v.Mul(power)
	//func (this *Point) Translate(v Vector) *Point {

	end := start.Translate(*t)
	//	end := v.Add(&t)

	tractorLine := LineSegment{start, *end}
	//	var target *Sprite

	for spr, v := range this.game.Sprites {
		if this != spr && v == true {
			if spr.Intersects(&tractorLine) {
				//target = spr
				this.Tractored = spr
			}
		}
	}

}

/**
 *
 *     F = G * m(a) * m(b) / d2
 *     G = 6.673 x 10-11 N m2/kg2
 *
 */
func (this *Sprite) PullOn(other *Sprite) {

	dis := other.distanceVector(this)
	direction := math.Atan2(-dis.y, dis.x)
	//magnitude := dis.Length()

	G := 6.673 * math.Pow10(-11)
	F := FORCE_COEFFICIENT * ((G * this.Mass * other.Mass) / math.Pow((DISTANCE_COEFFICIENT*(dis.Length())), 2))

	if other.GetKind() == SHIP {
		//		fmt.Printf("PullOn: F = %v, old way %v\n", F, magnitude)
	}

	f := NewLinearForce(Gravitation, direction, F, 0)
	other.AddForce(f)

}

func Ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (this *Sprite) gobble(token *Sprite) {

	if this.GetKind() != SHIP || token.GetKind() != PRIZE {
		fmt.Printf("gobble type error \n")
		return
	}

	prizeType := token.typeInfo & PRIZE_TYPE
	value := token.typeInfo & PRIZE_VALUE

	var resource PlayerResourceType
	switch prizeType {
	case SHIELD:
		resource = ShieldResource
	case LIFEENERGY:
		resource = LifeEnergyResource
	case HYPERSPACE:
		resource = HyperspaceResource
	case BOOSTER:
		resource = BoosterResource
	case CLOAK:
		resource = CloakResource
	case TRACTOR:
		resource = TractorResource
	default:
	}

	this.player.AddResource(resource, int(value))

	//resource, _ := token.getProperty("Prize")
	//value, _ := token.getProperty("PrizeValue")

	//fmt.Printf("gobble: resource = %s, value = %s\n", resource, value)
	//	r := PlayerResourceType_value[resource]
	//	rType := PlayerResourceType(r)
	//	iValue, _ := strconv.Atoi(value)
	//	this.Player.AddResource(rType, iValue)
	//this.Player.AddResource(token.prize.resource, token.prize.value)

}

//type Prize struct {
//	resource PlayerResourceType
//	value    int
//}
