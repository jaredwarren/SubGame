package world

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/jaredwarren/SubGame/internal/game/cave"
)

// TileType represents the type of a map tile.
type TileType int

const (
	TileWater TileType = iota
	TileLand
	TileTrench
	TileWreckage
	TileShockKelpCave
)

// World orchestrates procedural generation of overworld and caves.
type World struct {
	OverworldMap  [][]TileType
	LandDist      [][]int             // Precomputed BFS distance from each tile to nearest land
	WaterDist     [][]int             // Precomputed BFS distance from each tile to nearest water
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

	// Scatter features using the tile type registry
	r := rand.New(rand.NewSource(w.Seed + 13))
	var scatterTypes []TileType
	for tt, info := range AllTileInfos() {
		if info.ScatterCount > 0 {
			scatterTypes = append(scatterTypes, tt)
		}
	}
	sort.Slice(scatterTypes, func(i, j int) bool {
		return scatterTypes[i] < scatterTypes[j]
	})
	for _, tt := range scatterTypes {
		info := GetTileInfo(tt)
		w.scatterFeature(r, tt, info.ScatterCount)
	}

	// Precompute BFS distance maps for fast per-tile lookups
	w.buildLandDistMap()
	w.buildWaterDistMap()
}

// isOceanArea checks if a 5x5 area centered at (tx, ty) consists entirely of TileWater.
func (w *World) isOceanArea(tx, ty int) bool {
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			if w.OverworldMap[tx+dx][ty+dy] != TileWater {
				return false
			}
		}
	}
	return true
}

// scatterFeature scatters a specific tile type in deep ocean areas.
func (w *World) scatterFeature(r *rand.Rand, tileType TileType, count int) {
	featureCount := 0
	attempts := 0
	for featureCount < count && attempts < 2000 {
		tx := r.Intn(w.Width-10) + 5
		ty := r.Intn(w.Height-10) + 5

		if w.isOceanArea(tx, ty) {
			w.OverworldMap[tx][ty] = tileType
			featureCount++
		}
		attempts++
	}
}

// buildDistMap computes a BFS distance map from every tile to the nearest tile satisfying matches predicate.
func (w *World) buildDistMap(matches func(TileType) bool) [][]int {
	dist := make([][]int, w.Width)
	for x := 0; x < w.Width; x++ {
		dist[x] = make([]int, w.Height)
		for y := 0; y < w.Height; y++ {
			dist[x][y] = -1 // unvisited
		}
	}

	type pos struct{ x, y int }
	queue := make([]pos, 0, w.Width*w.Height/4)

	// Seed BFS with matching tiles
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			if matches(w.OverworldMap[x][y]) {
				dist[x][y] = 0
				queue = append(queue, pos{x, y})
			}
		}
	}

	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if nx >= 0 && nx < w.Width && ny >= 0 && ny < w.Height && dist[nx][ny] == -1 {
				dist[nx][ny] = dist[cur.x][cur.y] + 1
				queue = append(queue, pos{nx, ny})
			}
		}
	}
	return dist
}

// buildLandDistMap computes BFS distance from every tile to the nearest land tile.
func (w *World) buildLandDistMap() {
	w.LandDist = w.buildDistMap(func(t TileType) bool {
		return t == TileLand
	})
}

// buildWaterDistMap computes BFS distance from every tile to the nearest water tile.
func (w *World) buildWaterDistMap() {
	w.WaterDist = w.buildDistMap(func(t TileType) bool {
		info := GetTileInfo(t)
		return info != nil && info.IsWater
	})
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
	tx = max(0, min(tx, w.Width-1))
	ty = max(0, min(ty, w.Height-1))

	key := fmt.Sprintf("%d_%d", tx, ty)
	if caveGrid, exists := w.Caves[key]; exists {
		return caveGrid
	}

	tileType := w.OverworldMap[tx][ty]
	seed := w.Seed + int64(tx*73) + int64(ty*31)
	r := rand.New(rand.NewSource(seed))

	var caveGrid [][]bool

	info := GetTileInfo(tileType)
	if info != nil && info.GenerateGrid != nil {
		caveGrid = info.GenerateGrid(r)
	} else {
		dist := w.DistanceToLand(tx, ty)
		hasLeftWater := tx-1 >= 0 && w.OverworldMap[tx-1][ty] == TileWater
		hasRightWater := tx+1 < w.Width && w.OverworldMap[tx+1][ty] == TileWater
		caveGrid = cave.GenerateShallowSeabedGrid(r, dist, hasLeftWater, hasRightWater)
	}

	w.Caves[key] = caveGrid
	return caveGrid
}
