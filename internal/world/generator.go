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

	const (
		caveW  = 60
		caveH  = 120
		splitY = 60
	)

	tileType := w.OverworldMap[tx][ty]
	if tileType == TileWreckage {
		r := rand.New(rand.NewSource(w.Seed + int64(tx*73) + int64(ty*31)))
		cave := GenerateWreckageCave(caveW, caveH, r)
		w.Caves[key] = cave
		return cave
	}

	if tileType != TileTrench {
		dist := w.DistanceToLand(tx, ty)
		floorY := 6 + int(dist*2.2)
		if floorY < 6 {
			floorY = 6
		}
		if floorY > 60 {
			floorY = 60
		}

		hasLeftWater := tx-1 >= 0 && w.OverworldMap[tx-1][ty] == TileWater
		hasRightWater := tx+1 < w.Width && w.OverworldMap[tx+1][ty] == TileWater

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

			// Apply slope to the left edge if the neighbor is not water
			if !hasLeftWater && x < 15 {
				t := float64(x) / 15.0
				t = math.Sin(t * math.Pi / 2.0)
				blendY := 4.0 + (float64(colFloorY)-4.0)*t
				colFloorY = int(blendY)
			}

			// Apply slope to the right edge if the neighbor is not water
			if !hasRightWater && x >= caveW-15 {
				t := float64(caveW-1-x) / 15.0
				t = math.Sin(t * math.Pi / 2.0)
				blendY := 4.0 + (float64(colFloorY)-4.0)*t
				colFloorY = int(blendY)
			}

			for y := 0; y < caveH; y++ {
				isLeftBorderSolid := !hasLeftWater && x == 0
				isRightBorderSolid := !hasRightWater && x == caveW-1
				if isLeftBorderSolid || isRightBorderSolid || y >= colFloorY {
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

// GenerateWreckageCave generates a grid-aligned ship hull structure with central vertical elevator shaft, corridors, and rooms.
func GenerateWreckageCave(w, h int, r *rand.Rand) [][]bool {
	grid := make([][]bool, w)
	for x := 0; x < w; x++ {
		grid[x] = make([]bool, h)
		for y := 0; y < h; y++ {
			grid[x][y] = true
		}
	}

	// 1. Central elevator shaft
	shaftX1 := w/2 - 3 // 27
	shaftX2 := w/2 + 2 // 32
	for y := 0; y < h-4; y++ {
		for x := shaftX1; x <= shaftX2; x++ {
			grid[x][y] = false
		}
	}

	// 2. Horizontal corridors (decks)
	deckYs := []int{24, 52, 80, 108}
	deckHeight := 4

	for _, dy := range deckYs {
		for y := dy; y < dy+deckHeight; y++ {
			for x := 4; x < w-4; x++ {
				grid[x][y] = false
			}
		}
	}

	carveRoom := func(x1, y1, x2, y2 int) {
		for x := x1; x <= x2; x++ {
			for y := y1; y <= y2; y++ {
				grid[x][y] = false
			}
		}
	}

	carveDoor := func(doorX, y1, y2 int) {
		for y := y1; y <= y2; y++ {
			grid[doorX][y] = false
			grid[doorX+1][y] = false
		}
	}

	// 3. Generate rooms branching off corridors
	bays := []struct {
		yMin int
		yMax int
	}{
		{4, deckYs[0] - 1},
		{deckYs[0] + deckHeight, deckYs[1] - 1},
		{deckYs[1] + deckHeight, deckYs[2] - 1},
		{deckYs[2] + deckHeight, deckYs[3] - 1},
	}

	for _, bay := range bays {
		bayH := bay.yMax - bay.yMin + 1
		if bayH < 6 {
			continue
		}

		leftXMin := 4
		leftXMax := shaftX1 - 2
		rightXMin := shaftX2 + 2
		rightXMax := w - 5

		generateRoomsInBay := func(xMin, xMax int, yMin, yMax int, doorToY int) {
			width := xMax - xMin + 1
			if width < 8 {
				return
			}

			numRooms := 2
			if width >= 18 && r.Float64() < 0.6 {
				numRooms = 3
			}

			roomWidth := width / numRooms
			for i := 0; i < numRooms; i++ {
				rx1 := xMin + i*roomWidth + 1
				rx2 := rx1 + roomWidth - 3
				if i == numRooms-1 {
					rx2 = xMax - 1
				}

				ry1 := yMin + 1
				ry2 := yMax - 1

				if rx2 > rx1 && ry2 > ry1 {
					carveRoom(rx1, ry1, rx2, ry2)

					doorX := (rx1 + rx2) / 2
					if doorToY > ry2 {
						carveDoor(doorX, ry2+1, doorToY)
					} else {
						carveDoor(doorX, doorToY, ry1-1)
					}
				}
			}
		}

		if bayH >= 18 {
			midY := (bay.yMin + bay.yMax) / 2
			// Upper half
			generateRoomsInBay(leftXMin, leftXMax, bay.yMin, midY-1, bay.yMin-1)
			generateRoomsInBay(rightXMin, rightXMax, bay.yMin, midY-1, bay.yMin-1)
			// Lower half
			generateRoomsInBay(leftXMin, leftXMax, midY+1, bay.yMax, bay.yMax+1)
			generateRoomsInBay(rightXMin, rightXMax, midY+1, bay.yMax, bay.yMax+1)
		} else {
			doorToY := bay.yMax + 1
			generateRoomsInBay(leftXMin, leftXMax, bay.yMin, bay.yMax, doorToY)
			generateRoomsInBay(rightXMin, rightXMax, bay.yMin, bay.yMax, doorToY)
		}
	}

	// 4. Ensure entrance at top center is open for diving player
	for y := 0; y < 5; y++ {
		for x := w/2 - 3; x <= w/2+3; x++ {
			grid[x][y] = false
		}
	}

	return grid
}

