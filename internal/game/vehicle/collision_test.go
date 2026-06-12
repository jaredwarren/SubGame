package vehicle

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestSolidAt(t *testing.T) {
	// A mock query function that returns true only for a set of specific coordinates.
	// We can use a map keyed by [2]int to check if a specific tile is solid.
	solids := map[[2]int]bool{
		{-1, -1}: true,
		{0, -1}:  true,
		{-2, -2}: true,
		{1, 1}:   true,
	}

	query := func(tx, ty int) bool {
		return solids[[2]int{tx, ty}]
	}

	tests := []struct {
		name     string
		pos      gvec.Vec2
		dims     gvec.Vec2
		expected bool
	}{
		{
			name:     "origin offset overlap",
			pos:      gvec.Vec2{X: -1, Y: -1},
			dims:     gvec.Vec2{X: 10, Y: 10},
			expected: true, // overlaps with {-1, -1}
		},
		{
			name:     "positive overlap",
			pos:      gvec.Vec2{X: 63, Y: 63},
			dims:     gvec.Vec2{X: 10, Y: 10},
			expected: true, // overlaps with {1, 1} since 63+10 = 73 which is in tile 1 (from 64 to 127)
		},
		{
			name:     "flush against positive boundary without overlap",
			pos:      gvec.Vec2{X: 0, Y: 0},
			dims:     gvec.Vec2{X: 64, Y: 64}, // exactly 0 to 64
			expected: false, // does not overlap with {1, 1}
		},
		{
			name:     "negative bounds -63 overlap with -1 tile",
			pos:      gvec.Vec2{X: -63, Y: -10},
			dims:     gvec.Vec2{X: 10, Y: 20},
			expected: true, // X is in [-63, -53] which is in tile -1. Y is in [-10, 10] which overlaps with Y=-1.
		},
		{
			name:     "no overlap completely out",
			pos:      gvec.Vec2{X: 200, Y: 200},
			dims:     gvec.Vec2{X: 10, Y: 10},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := solidAt(query, tc.pos, tc.dims)
			if got != tc.expected {
				t.Errorf("solidAt expected %v, got %v", tc.expected, got)
			}
		})
	}
}
