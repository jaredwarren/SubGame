package gvec

import (
	"math"
)

// TileRange calculates the tile index range spanned by a bounding box.
// Subtracts a small epsilon of 0.001 from the maximum bounds to prevent flush boundaries probing an extra tile.
func TileRange(pos Vec2, dims Vec2, tileSize int) (x1, x2, y1, y2 int) {
	d := float64(tileSize)
	x1 = int(math.Floor(pos.X / d))
	x2 = int(math.Floor((pos.X + dims.X - 0.001) / d))
	y1 = int(math.Floor(pos.Y / d))
	y2 = int(math.Floor((pos.Y + dims.Y - 0.001) / d))
	return
}

// MoveAxisSeparated updates pos and vel on both X and Y axes, resolving collisions against a solid-checking function.
// If a collision occurs on an axis, the velocity on that axis is zeroed and the corresponding impact callback is triggered.
func MoveAxisSeparated(pos *Vec2, vel *Vec2, dims Vec2, isSolid func(Vec2) bool, onImpactX, onImpactY func()) {
	newX := pos.X + vel.X
	if isSolid(Vec2{X: newX, Y: pos.Y}) {
		if onImpactX != nil {
			onImpactX()
		}
		vel.X = 0
	} else {
		pos.X = newX
	}

	newY := pos.Y + vel.Y
	if isSolid(Vec2{X: pos.X, Y: newY}) {
		if onImpactY != nil {
			onImpactY()
		}
		vel.Y = 0
	} else {
		pos.Y = newY
	}
}
