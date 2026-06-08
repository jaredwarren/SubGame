package cave

import (
	"math/rand"
)

// GenerateCellularCave generates organic bubble caves using Cellular Automata rules.
// Returns a 2D boolean grid where true represents a solid cave wall and false represents open water.
func GenerateCellularCave(width, height int, wallChance float64, steps int, r *rand.Rand) [][]bool {
	// Initialize grid: true is solid wall, false is empty path
	grid := make([][]bool, width)
	for x := 0; x < width; x++ {
		grid[x] = make([]bool, height)
		for y := 0; y < height; y++ {
			// Boundaries are always solid walls
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				grid[x][y] = true
			} else {
				grid[x][y] = r.Float64() < wallChance
			}
		}
	}

	// Helper to count solid neighbors in a 3x3 grid around (cx, cy)
	countNeighbors := func(g [][]bool, cx, cy int) int {
		count := 0
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := cx+dx, cy+dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height {
					if g[nx][ny] {
						count++
					}
				} else {
					// Out-of-bounds cells behave as solid boundaries
					count++
				}
			}
		}
		return count
	}

	// Run cellular automata simulation iterations
	for step := 0; step < steps; step++ {
		newGrid := make([][]bool, width)
		for x := 0; x < width; x++ {
			newGrid[x] = make([]bool, height)
			for y := 0; y < height; y++ {
				// Retain solid borders
				if x == 0 || x == width-1 || y == 0 || y == height-1 {
					newGrid[x][y] = true
					continue
				}

				neighbors := countNeighbors(grid, x, y)
				if neighbors > 4 {
					newGrid[x][y] = true
				} else if neighbors < 4 {
					newGrid[x][y] = false
				} else {
					newGrid[x][y] = grid[x][y]
				}
			}
		}
		grid = newGrid
	}

	return grid
}
