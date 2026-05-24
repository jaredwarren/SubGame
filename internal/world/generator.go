package world

import (
	"fmt"
	"math/rand"
)

// TileType represents the type of a map tile.
type TileType int

const (
	TileWater TileType = iota
	TileLand
	TileTrench
)

// World orchestrates procedural generation of overworld and caves.
type World struct {
	OverworldMap  [][]TileType
	Caves         map[string][][]bool // Key: "trenchX_trenchY" -> Cave grid
	Width, Height int
	Seed          int64
}

// NewWorld creates and procedurally initializes a new World.
func NewWorld(seed int64) *World {
	w := &World{
		Width:  100,
		Height: 100,
		Caves:  make(map[string][][]bool),
		Seed:   seed,
	}
	w.generateOverworld()
	return w
}

// generateOverworld builds the top-down sea and islands.
func (w *World) generateOverworld() {
	w.OverworldMap = make([][]TileType, w.Width)
	for x := 0; x < w.Width; x++ {
		w.OverworldMap[x] = make([]TileType, w.Height)
	}

	noise := NewNoise2D(w.Seed)

	// Populate islands and oceans using FBM noise
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			nx := float64(x) / 12.0
			ny := float64(y) / 12.0
			val := noise.FBM(nx, ny, 3)

			// Land threshold
			if val > 0.62 {
				w.OverworldMap[x][y] = TileLand
			} else {
				w.OverworldMap[x][y] = TileWater
			}
		}
	}

	// Scatter Trench/Sinkhole diving locations in deep ocean tiles
	r := rand.New(rand.NewSource(w.Seed + 13))
	trenchCount := 0
	attempts := 0

	for trenchCount < 6 && attempts < 2000 {
		tx := r.Intn(w.Width-10) + 5
		ty := r.Intn(w.Height-10) + 5

		// Check if the area is ocean water
		isOcean := true
		for dx := -2; dx <= 2; dx++ {
			for dy := -2; dy <= 2; dy++ {
				if w.OverworldMap[tx+dx][ty+dy] != TileWater {
					isOcean = false
					break
				}
			}
		}

		if isOcean {
			w.OverworldMap[tx][ty] = TileTrench
			trenchCount++
		}
		attempts++
	}
}

// GetCave returns a procedurally generated cave linked to the trench position.
func (w *World) GetCave(tx, ty int) [][]bool {
	key := fmt.Sprintf("%d_%d", tx, ty)
	if cave, exists := w.Caves[key]; exists {
		return cave
	}

	// Generate a new cave if it doesn't exist yet
	r := rand.New(rand.NewSource(w.Seed + int64(tx*73) + int64(ty*31)))
	
	const (
		caveW = 60
		caveH = 120
		splitY = 60
	)

	// 1. Generate upper shallow cave (Cellular Automata)
	shallowCave := GenerateCellularCave(caveW, splitY, 0.42, 4, r)

	// 2. Generate lower deep crevice cave (Drunkard's Walk)
	deepCave := GenerateDrunkardCave(caveW, caveH-splitY, r)

	// 3. Instantiate full cave grid
	cave := make([][]bool, caveW)
	for x := 0; x < caveW; x++ {
		cave[x] = make([]bool, caveH)
	}

	// 4. Merge upper and lower caves
	for x := 0; x < caveW; x++ {
		for y := 0; y < caveH; y++ {
			if y < splitY {
				cave[x][y] = shallowCave[x][y]
			} else {
				cave[x][y] = deepCave[x][y-splitY]
			}
		}
	}

	// 5. Connect the two halves at the split boundary
	// Carve a vertical connecting shaft in the middle to ensure pathability
	const shaftHalfWidth = 2
	for y := splitY - 8; y < splitY+8; y++ {
		for x := (caveW / 2) - shaftHalfWidth; x <= (caveW / 2)+shaftHalfWidth; x++ {
			if x > 0 && x < caveW-1 && y > 0 && y < caveH-1 {
				cave[x][y] = false // Carve path
			}
		}
	}

	// 6. Ensure entrance at top center is open for diving player
	for y := 0; y < 5; y++ {
		for x := (caveW / 2) - 3; x <= (caveW / 2)+3; x++ {
			if x > 0 && x < caveW-1 && y < caveH-1 {
				cave[x][y] = false
			}
		}
	}

	w.Caves[key] = cave
	return cave
}
