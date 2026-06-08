package cave

import (
	"math/rand"
)

// GenerateDrunkardCave carves winding vertical shafts in a solid grid using biased random walkers.
// Returns a 2D boolean grid where true is a solid wall and false is open water.
func GenerateDrunkardCave(width, height int, r *rand.Rand) [][]bool {
	// Initialize grid: starts completely solid
	grid := make([][]bool, width)
	for x := 0; x < width; x++ {
		grid[x] = make([]bool, height)
		for y := 0; y < height; y++ {
			grid[x][y] = true
		}
	}

	// Helper to carve an empty brush (false) around (cx, cy)
	carve := func(cx, cy, brushSize int) {
		for dx := -brushSize; dx <= brushSize; dx++ {
			for dy := -brushSize; dy <= brushSize; dy++ {
				nx, ny := cx+dx, cy+dy
				// Leave at least a 1-tile solid border at edges
				if nx > 0 && nx < width-1 && ny > 0 && ny < height-1 {
					grid[nx][ny] = false
				}
			}
		}
	}

	// Run multiple walkers to create intersecting vertical crevices
	const numWalkers = 3
	for i := 0; i < numWalkers; i++ {
		// Start at top center with slight variation
		cx := width/2 + r.Intn(6) - 3
		cy := 2

		// Carve initial entrance pocket
		carve(cx, cy, 2)

		// Walk until reaching near the bottom border
		for cy < height-3 {
			roll := r.Float64()
			switch {
			case roll < 0.52:
				cy++ // Down (biased)
			case roll < 0.73:
				cx-- // Left
			case roll < 0.94:
				cx++ // Right
			default:
				cy-- // Up
			}

			// Clamp to maintain solid boundaries
			if cx < 2 {
				cx = 2
			}
			if cx > width-3 {
				cx = width - 3
			}
			if cy < 2 {
				cy = 2
			}

			// Carve a 2x2 brush area at the current walker position
			carve(cx, cy, 1)
		}
	}

	return grid
}
