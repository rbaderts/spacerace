package core

import (
	_ "bytes"
	_ "fmt"
	_ "math"
	"time"
)

func TestCollision(p1 CollisionShape, p2 CollisionShape) bool {

	if p1.GetShapeType() == PolygonShape {
		switch p2.GetShapeType() {
		case PolygonShape:
			return PolyPolyIntersects(p1.(*Polygon), p2.(*Polygon))
		case CircleShape:
			return PolyCircleIntersects(p1.(*Polygon), p2.(*Circle))
		case LineShape:
			return LinePolyIntersects(p2.(*LineSegment), p1.(*Polygon))
		}

	} else if p1.GetShapeType() == CircleShape {
		switch p2.GetShapeType() {
		case PolygonShape:
			return PolyCircleIntersects(p2.(*Polygon), p1.(*Circle))
		case CircleShape:
			return CircleCircleIntersects(p2.(*Circle), p1.(*Circle))
		case LineShape:
			return LineCircleIntersects(p2.(*LineSegment), p1.(*Circle))
		}

	} else if p1.GetShapeType() == LineShape {
		switch p2.GetShapeType() {
		case PolygonShape:
			return LinePolyIntersects(p1.(*LineSegment), p2.(*Polygon))
		case CircleShape:
			return LineCircleIntersects(p1.(*LineSegment), p2.(*Circle))
		case LineShape:
			return LineLineIntersects(p2.(*LineSegment), p1.(*LineSegment))
		}
	}

	return false

}

/**
 * This uses SAT collision detection
 *  @see <a href="https://en.wikipedia.org/wiki/Hyperplane_separation_theorem">Separating Axis Theorm</a>
 */
func PolyPolyIntersects(p1 *Polygon, p2 *Polygon) bool {

	defer TimeCall(time.Now(), "PolyPolyIntersect")

	// here we use a Point, X represents the min projection, Y represents the max projection

	// The min/max Projection of poly1 and poly2 onto the X-Axis
	P1_X := Point{10000.00, -100000.00}
	P2_X := Point{10000.00, -100000.00}

	// The min/max Projection of poly1 and poly2 onto the Y-Axis
	P1_Y := Point{10000.00, -100000.00}
	P2_Y := Point{10000.00, -100000.00}

	for _, p := range p1.vertices {
		if p.x <= P1_X.x {
			P1_X.x = p.x
		}
		if p.x >= P1_X.y {
			P1_X.y = p.x
		}
		if p.y <= P1_Y.x {
			P1_Y.x = p.y
		}
		if p.y >= P1_Y.x {
			P1_Y.y = p.y
		}
	}

	//	fmt.Printf("p1 xmin/max: %d, %d :   p2 xmin/xmax)"

	for _, p := range p2.vertices {
		if p.GetX() <= P2_X.GetX() {
			P2_X.x = p.GetX()
		}
		if p.GetX() >= P2_X.GetY() {
			P2_X.y = p.GetY()
		}
		if p.GetY() <= P2_Y.GetX() {
			P2_Y.x = p.GetX()
		}
		if p.GetY() >= P2_Y.GetY() {
			P2_Y.y = p.GetY()
		}
	}

	if P1_X.GetY() < P2_X.GetX() || P2_X.GetY() < P1_X.GetX() {
		return false
	}

	if P1_Y.GetY() < P2_Y.GetX() || P2_Y.GetY() < P1_Y.GetX() {
		return false
	}

	return true

}

func LineLineIntersects(p1 *LineSegment, p2 *LineSegment) bool {
	return p1.intersects(p2)
}
func LineCircleIntersects(p1 *LineSegment, p2 *Circle) bool {

	dis := p1.start.Distance(p2.Center)
	if dis <= p2.Radius {
		return true
	}
	dis = p1.end.Distance(p2.Center)
	if dis <= p2.Radius {
		return true
	}
	return false
}
func LinePolyIntersects(p1 *LineSegment, p2 *Polygon) bool {
	for i, _ := range p2.vertices {
		j := i + 1
		if i == len(p2.vertices) {
			j = 0
		}
		seg := LineSegment{p2.vertices[i], p2.vertices[j]}

		if p1.intersects(&seg) {
			return true
		}
	}
	return false

}

/**
 * Checks if the distance between any poly vertex and the circle center is smaller than the circle radius
 */
func PolyCircleIntersects(p1 *Polygon, p2 *Circle) bool {

	defer TimeCall(time.Now(), "PolyCirleIntersect")

	start := len(p1.vertices) - 1
	end := 0

	seg := &LineSegment{p1.vertices[start], p1.vertices[end]}
	//seg := NewSegment(p1.vertices[start], p1.vertices[end])
	//seg := &Segment2{p1.vertices[start], p1.vertices[end]}

	if seg.distanceFrom(&(p2.Center)) < p2.Radius {
		return true
	}

	for i := 1; i < len(p1.vertices); i++ {
		seg = &LineSegment{p1.vertices[i-1], p1.vertices[i]}
		if seg.distanceFrom(&(p2.Center)) < p2.Radius {
			return true
		}
	}

	if p1.Contains(&p2.Center) {
		return true
	}

	return false

	//fmt.Println("PolyCircleIntersects\n")

	/*
		for _, v := range p1.vertices {
			if v.Distance(&p2.Center) <= p2.Radius {
				return true
			}
		}

		if p1.Contains(&p2.Center) {
			return true
		}
		return false
	*/
}

/**
 * Checks if the distance between the two circle centers is smaller than the sum of their radiuses */
func CircleCircleIntersects(p1 *Circle, p2 *Circle) bool {

	//fmt.Println("CircleCircleIntersects\n")
	dis := p1.Center.Distance(p2.Center)
	if dis <= (p1.Radius + p2.Radius) {
		return true
	}
	return false
}
