package world

import (
	"fmt"
	"math"
	"math/rand"
)

// TileType represents the type of a map tile.
type TileType int

const (
	TileWater TileType = iota
	TileLand
	TileTrench
	TileWreckage
)

// World orchestrates procedural generation of overworld and caves.
type World struct {
	OverworldMap  [][]TileType
	LandDist      [][]int              // Precomputed BFS distance from each tile to nearest land
	Caves         map[string][][]bool  // Key: "trenchX_trenchY" -> Cave grid
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

	// Scatter wreckage locations
	wreckageCount := 0
	attempts = 0
	for wreckageCount < 3 && attempts < 2000 {
		tx := r.Intn(w.Width-10) + 5
		ty := r.Intn(w.Height-10) + 5

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
			w.OverworldMap[tx][ty] = TileWreckage
			wreckageCount++
		}
		attempts++
	}

	// Precompute BFS distance-to-land map for fast per-tile lookups
	w.buildLandDistMap()
}

// buildLandDistMap computes BFS distance from every tile to the nearest land tile.
// Result is stored in w.LandDist[x][y] (0 for land tiles, increasing outward).
func (w *World) buildLandDistMap() {
	w.LandDist = make([][]int, w.Width)
	for x := 0; x < w.Width; x++ {
		w.LandDist[x] = make([]int, w.Height)
		for y := 0; y < w.Height; y++ {
			w.LandDist[x][y] = -1 // unvisited
		}
	}

	type pos struct{ x, y int }
	queue := make([]pos, 0, w.Width*w.Height/4)

	// Seed BFS with all land tiles
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			if w.OverworldMap[x][y] == TileLand {
				w.LandDist[x][y] = 0
				queue = append(queue, pos{x, y})
			}
		}
	}

	// BFS expansion (4-directional Chebyshev would also work, but Manhattan is fine)
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if nx >= 0 && nx < w.Width && ny >= 0 && ny < w.Height && w.LandDist[nx][ny] == -1 {
				w.LandDist[nx][ny] = w.LandDist[cur.x][cur.y] + 1
				queue = append(queue, pos{nx, ny})
			}
		}
	}
}

// DistanceToLand returns the BFS distance (in tiles) from (tx, ty) to the nearest land tile.
func (w *World) DistanceToLand(tx, ty int) float64 {
	if tx < 0 || tx >= w.Width || ty < 0 || ty >= w.Height {
		return 999.0
	}
	return float64(w.LandDist[tx][ty])
}

// GetCave returns a procedurally generated cave linked to the trench position.
func (w *World) GetCave(tx, ty int) [][]bool {
	// Clamp inputs to safe overworld boundaries
	if tx < 0 {
		tx = 0
	}
	if tx >= w.Width {
		tx = w.Width - 1
	}
	if ty < 0 {
		ty = 0
	}
	if ty >= w.Height {
		ty = w.Height - 1
	}

	key := fmt.Sprintf("%d_%d", tx, ty)
	if cave, exists := w.Caves[key]; exists {
		return cave
	}

	isTrench := w.OverworldMap[tx][ty] == TileTrench

	const (
		caveW  = 60
		caveH  = 120
		splitY = 60
	)

	if !isTrench {
		dist := w.DistanceToLand(tx, ty)
		floorY := 6 + int(dist*2.2)
		if floorY < 6 {
			floorY = 6
		}
		if floorY > 60 {
			floorY = 60
		}

		// Create a local random generator seeded by coordinates to make every seabed layout unique
		r := rand.New(rand.NewSource(w.Seed + int64(tx*73) + int64(ty*31)))
		freq1 := 0.15 + r.Float64()*0.2
		freq2 := 0.05 + r.Float64()*0.1
		amp1 := 2.0 + r.Float64()*4.0
		amp2 := 1.0 + r.Float64()*3.0

		cave := make([][]bool, caveW)
		for x := 0; x < caveW; x++ {
			cave[x] = make([]bool, caveH)
			colFloorY := floorY + int(math.Sin(float64(x)*freq1)*amp1+math.Cos(float64(x)*freq2)*amp2)
			if colFloorY < 6 {
				colFloorY = 6
			}
			for y := 0; y < caveH; y++ {
				if x == 0 || x == caveW-1 || y >= colFloorY {
					cave[x][y] = true
				} else {
					cave[x][y] = false
				}
			}
		}
		w.Caves[key] = cave
		return cave
	}

	// Generate a new cave if it doesn't exist yet
	r := rand.New(rand.NewSource(w.Seed + int64(tx*73) + int64(ty*31)))
	
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
