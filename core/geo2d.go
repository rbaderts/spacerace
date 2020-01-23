package core

import (
	"bytes"
	"fmt"
	"math"
)

/**
 *  Set of 2D Plane(Euclidiean) Geometry primitives
 */

type Point struct {
	x float64
	y float64
}

type Bound struct {
	topLeft     Point
	bottomRight Point
}

func (this *Bound) Contains(p *Point) bool {

	if p.y < this.topLeft.y || this.bottomRight.y < p.y {
		return false
	}

	if p.x < this.topLeft.x || this.bottomRight.x < p.x {
		return false
	}

	return true
}

func (this *Point) GetX() float64 {
	return this.x
}

func (this *Point) GetY() float64 {
	return this.y
}

func (this Point) Distance(other Point) float64 {
	dX := this.x - other.x
	dY := this.y - other.y
	result := math.Sqrt(math.Pow(dX, 2) + math.Pow(dY, 2))
	return result
}

func (this *Point) Vector() *Vector {
	return &Vector{this.x, this.y}
}

func (this Point) Direction(other Point) float64 {
	dir := math.Atan2((other.y - this.y), (other.x - this.x))
	//	if dir < 0 {
	//		dir = 2*math.Pi - dir
	//	}
	fmt.Printf("Direction.  src = %v, dest = %v, dir = %v\n", this, other, dir)
	return dir
}

type LineSegment struct {
	start Point
	end   Point
}

func (this LineSegment) String() string {
	return fmt.Sprintf("start: %v, end %v", this.start, this.end)
}

func (this LineSegment) GetShapeType() ShapeType {
	return LineShape
}

//type Vector struct {
//	x float64
//	y float64
//angle     float64
//magnitude float64

//GetX() float64
//GetY() float64
//GetAngle() float64
//GetMagnitude() float64
//}

type Vector struct {
	x float64
	y float64
}

//type PolarVector struct {
//	angle     float64
//	magnitude float64
//}

/*
func (v PolarVector) String() string {
	return fmt.Sprintf("PolarVector:  angle: %f, magnitude: %f", v.angle, v.magnitude)
}

func (this *PolarVector) GetX() float64 {
	return this.Vector().x
}

func (this *PolarVector) GetY() float64 {
	return this.Vector().y
}

*/

func UnitVector(angle float64) *Vector {
	var x = math.Cos(angle)
	var y = math.Sin(angle)
	return &Vector{x, y}
}

func (this *Vector) NonZero() bool {
	if this.x != 0 || this.y != 0 {
		return true
	}
	return false
}

func (v Vector) String() string {
	return fmt.Sprintf("Vector:  x: %f, y: %f", v.x, v.y)
}

func (this Vector) GetX() float64 {
	return this.x
}

func (this Vector) GetY() float64 {
	return this.y
}

func (this Vector) Length() float64 {
	return this.GetMagnitude()
}

func (this Vector) GetMagnitude() float64 {
	result := math.Sqrt(math.Pow(this.x, 2) + math.Pow(this.y, 2))
	return result
}

func (this Vector) GetAngle() float64 {
	result := math.Atan2(this.y, this.x)
	return result
}

func (this Vector) Add(v Vector) Vector {
	return Vector{this.x + v.GetX(), this.y + v.GetY()}
}

func (this Vector) Sub(v Vector) Vector {
	return Vector{this.x - v.GetX(), this.y - v.GetY()}
}

func (this Vector) Mul(s float64) Vector {
	return Vector{this.x * s, this.y * s}
}

func (this Vector) Inverse() Vector {
	//x := -this.x
	//y := -this.y
	return Vector{0 - this.x, 0 - this.y}
}

func (this Point) Rotate(angle float64, origin *Point) *Point {
	xOffset := 0 - origin.GetX()
	yOffset := 0 - origin.GetY()

	tmp := Point{this.GetX() - origin.GetX(), this.GetY() + origin.GetY()}

	return &Point{tmp.x*math.Cos(angle) - tmp.y*math.Sin(angle) - xOffset,
		tmp.y*math.Sin(angle) + tmp.y*math.Cos(angle) - yOffset}
}

func (this Point) Translate(v Vector) *Point {
	return &Point{this.x + v.GetX(), this.y + v.GetY()}
}

func (this Point) Delta(v *Point) *Vector {
	x := this.x - v.x
	y := this.y - v.y
	return &Vector{x, y}
}

func (this *Point) Sub(v Vector) *Point {
	return &Point{this.x - v.x, this.y - v.y}
}

func (this *Point) Add(v Vector) *Point {
	return &Point{this.x + v.x, this.y + v.y}
}

func (this Point) String() string {
	return fmt.Sprintf("{%f, %f}", this.x, this.y)
}

func (this *LineSegment) Side(p *Point) int {
	val := (this.end.x-this.start.x)*(p.y-this.end.y) - (this.end.y-this.start.y)*(p.x-this.end.x)

	if val < 0 {
		return 1 // right
	} else if val > 0 {
		return -1 // left
	}

	return 0 // collinear

}

func (this *LineSegment) Bound() *Bound {
	return &Bound{Point{math.Max(this.start.x, this.end.x), math.Min(this.start.x, this.end.x)},
		Point{math.Max(this.start.y, this.end.y), math.Min(this.start.y, this.end.y)}}
}

func (this *LineSegment) intersects(other *LineSegment) bool {

	s1 := this.Side(&other.start)
	s2 := this.Side(&other.end)
	s3 := other.Side(&this.start)
	s4 := other.Side(&this.end)

	if s1 != s2 && s3 != s4 {
		return true
	}

	// Special Cases
	// l1 and l2.a collinear, check if l2.a is on l1
	lBound := this.Bound()
	if s1 == 0 && lBound.Contains(&other.start) {
		return true
	}

	// l1 and l2.b collinear, check if l2.b is on l1
	if s2 == 0 && lBound.Contains(&other.end) {
		return true
	}

	// TODO: are these next two tests redudant give the test above.
	// Thinking yes if there is round off magic.

	// l2 and l1.a collinear, check if l1.a is on l2
	lineBound := other.Bound()
	if s3 == 0 && lineBound.Contains(&this.start) {
		return true
	}

	// l2 and l1.b collinear, check if l1.b is on l2
	if s4 == 0 && lineBound.Contains(&this.end) {
		return true
	}

	return false

}

func (this *LineSegment) distanceFrom(p *Point) float64 {

	A := p.x - this.start.x
	B := p.y - this.start.y
	C := this.end.x - this.start.x
	D := this.end.y - this.start.y

	dot := A*C + B*D
	len_sq := C*C + D*D
	param := -1.0
	if len_sq != 0 { //in case of 0 length line
		param = dot / len_sq
	}

	var xx float64
	var yy float64

	if param < 0 {
		xx = this.start.x
		yy = this.start.y
	} else if param > 1 {
		xx = this.end.x
		yy = this.end.y
	} else {
		xx = this.start.x + param*C
		yy = this.start.y + param*D
	}

	dx := p.x - xx
	dy := p.y - yy
	return math.Sqrt(dx*dx + dy*dy)
}

func (this *LineSegment) rayIntersect(point *Point) bool {
	// Always ensure that the the first point
	// has a y coordinate that is less than the second point

	start := this.start
	end := this.end

	if start.y > end.y {
		// Switch the points if otherwise.
		start, end = end, start

	}

	// Move the point's y coordinate
	// outside of the bounds of the testing region
	// so we can start drawing a ray
	for point.y == start.y || point.y == end.y {
		newLng := math.Nextafter(point.y, math.Inf(1))
		point = &Point{point.x, newLng}
	}

	// If we are outside of the polygon, indicate so.
	if point.y < start.y || point.y > end.y {
		return false
	}

	if start.x > end.x {
		if point.x > start.x {
			return false
		}
		if point.x < end.x {
			return true
		}

	} else {
		if point.x > end.x {
			return false
		}
		if point.x < start.x {
			return true
		}
	}

	raySlope := (point.y - start.y) / (point.x - start.x)
	diagSlope := (end.y - start.y) / (end.x - start.x)

	return raySlope >= diagSlope
}

/**
 * Cicle
 */

type Circle struct {
	Center Point
	Radius float64
}

func (this Circle) String() string {
	return fmt.Sprintf("radius: %v\n", this.Radius)
}

func (this *Circle) GetShapeType() ShapeType {
	return CircleShape
}

func (this *Circle) intersects(other *Circle) bool {

	dis := this.Center.Distance(other.Center)
	//	fmt.Printf("Circle intersect:  c1: %v, radius = %v,  c2: %v, radius = %v,   Distance: %v\n", this.Center, this.Radius, other.Center, other.Radius, dis)
	if dis <= (this.Radius + other.Radius) {
		fmt.Printf("testing circle insersects circle: true\n")
		return true
	}
	return false
}

/*
func (this *Circle) intersectsRect(rect *Rectangle) bool {

	dx := math.Abs(this.Center.X - rect.TopLeft.X - (rect.Width / 2))
	dy := math.Abs(this.Center.Y - rect.TopLeft.Y - (rect.Height / 2))

	if dx > (rect.Width/2 + this.Radius) {
		return false
	}
	if dy > (rect.Height/2 + this.Radius) {
		return false
	}

	if dx <= (rect.Width / 2) {
		return true
	}
	if dy <= (rect.Height / 2) {
		return true
	}

	var ddx = dx - rect.Width/2
	var ddy = dy - rect.Height/2
	return (ddx*ddx+ddy*ddy <= (this.Radius * this.Radius))

}
*/

/**
 *  Polygon
 */

type Polygon struct {
	vertices []Point
}

func (this *Polygon) Rotate(angle float64, origin *Point) *Polygon {
	points := make([]Point, len(this.vertices))
	for i, v := range this.vertices {
		points[i] = *(v.Rotate(angle, origin))
	}
	return &Polygon{points}
}

func (this *Polygon) Contains(p *Point) bool {

	start := len(this.vertices) - 1
	end := 0

	seg := &LineSegment{this.vertices[start], this.vertices[end]}
	contains := seg.rayIntersect(p)

	for i := 1; i < len(this.vertices); i++ {
		seg = &LineSegment{this.vertices[i-1], this.vertices[i]}
		if seg.rayIntersect(p) {
			contains = !contains
		}
	}

	return contains

}

func (this *Polygon) GetShapeType() ShapeType {
	return PolygonShape
}

func (this Polygon) String() string {
	str := new(bytes.Buffer)
	for _, p := range this.vertices {
		str.WriteString(fmt.Sprintf(" v(%v,%v) : ", p.x, p.y))
	}
	return str.String()
}

func (this *Polygon) GetNormals() []Vector {
	//Shape.prototype.getNormals = function () {

	normals := make([]Vector, 2*len(this.vertices))
	for i, _ := range this.vertices {

		p := this.vertices[i]
		var v Point

		if i == len(this.vertices)-1 {
			v = this.vertices[0]
		} else {
			v = this.vertices[i+1]
		}

		//edge :=

		x1 := v.y - p.y
		y1 := -(v.x - p.x)
		l := math.Sqrt(x1*x1 + y1*y1)
		normals[i] = Vector{x1 / l, y1 / l}
		normals[i+len(this.vertices)] = Vector{-(x1 / l), -(y1 / l)}
	}

	return normals
}

//func (this *Polygon) Project(v Vector)

/*
double min = axis.dot(shape.vertices[0]);
double max = min;
for (int i = 1; i < shape.vertices.length; i++) {
  // NOTE: the axis must be normalized to get accurate projections
  double p = axis.dot(shape.vertices[i]);
  if (p < min) {
    min = p;
  } else if (p > max) {
    max = p;
  }
}
Projection proj = new Projection(min, max);
return proj;
*/

type CollisionShape interface {
	GetShapeType() ShapeType
	String() string
}

/*
type Inset struct {
	top    float64
	bottom float64
	left   float64
	right  float64
}

// assumes rectangle
func (this *Polygon) Inset(i Inset) *Polygon {

	this.vertices[0].X = this.vertices[0].X + ileft
	this.vertices[0] = this.vertices[0].X + ileft
	var r Rectangle
	r.TopLeft.X = this.TopLeft.X + i.left
	r.TopLeft.Y = this.TopLeft.Y + i.top
	r.Height = this.Height - i.top - i.bottom
	r.Width = this.Width - i.left - i.right
	return &r
}
*/
