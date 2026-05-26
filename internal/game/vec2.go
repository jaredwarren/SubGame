package game

import "math"

// Vec2 represents a 2D vector.
type Vec2 struct {
	X, Y float64
}

// Add returns the vector sum of v and other.
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Scale returns a new vector scaled by s.
func (v Vec2) Scale(s float64) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}

// Length returns the Euclidean length of the vector.
func (v Vec2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Distance returns the Euclidean distance between two vectors.
func (v Vec2) Distance(other Vec2) float64 {
	return math.Hypot(v.X-other.X, v.Y-other.Y)
}
