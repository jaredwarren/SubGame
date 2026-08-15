package world

import (
	"fmt"
	"math"
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
	TileThermoCave
)

// World orchestrates procedural generation of overworld and caves.
type World struct {
	OverworldMap  [][]TileType
	BiomeMap      [][]BiomeID
	LandDist      [][]int             // Precomputed BFS distance from each tile to nearest land
	WaterDist     [][]int             // Precomputed BFS distance from each tile to nearest water
	Caves         map[string][][]bool // Key: "trenchX_trenchY" -> Cave grid
	Width, Height int
	Seed          int64
}

// NewWorld creates and procedurally initializes a new World.
func NewWorld(seed int64) *World {
	w := &World{
		Width:  500,
		Height: 500,
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

	// Precompute BFS distance maps for fast per-tile lookups
	w.buildLandDistMap()
	w.buildWaterDistMap()
	w.generateBiomes()

	// Scatter global features (e.g. TileWreckage) using the tile type registry.
	// Seed+13 previously also scattered 6 trenches before wreckage; consume that
	// RNG the same way so wreckage sites stay put for existing world seeds.
	r := rand.New(rand.NewSource(w.Seed + 13))
	w.scatterFeature(r, TileTrench, 6)
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
	w.clearTiles(TileTrench)

	// Biome-local caves use a dedicated stream so later spawn-rate tweaks
	// cannot shift wreckage (or other Seed+13 features).
	w.scatterBiomeFeatures(rand.New(rand.NewSource(w.Seed + 17)))
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

// clearTiles converts every tile of the given type back to water.
func (w *World) clearTiles(tileType TileType) {
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			if w.OverworldMap[x][y] == tileType {
				w.OverworldMap[x][y] = TileWater
			}
		}
	}
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

// scatterBiomeFeatures scatters biome-specific special cave tiles based on BiomeSpec.
func (w *World) scatterBiomeFeatures(r *rand.Rand) {
	// Sort biome IDs for deterministic generation
	biomeIDs := []BiomeID{
		BiomeShallowReef,
		BiomeKelpForest,
		BiomeThermalBarrens,
		BiomeAbyssalBlue,
	}

	type pos struct{ x, y int }

	for _, bID := range biomeIDs {
		spec := GetBiomeInfo(bID)
		if spec == nil || spec.SpecialCaveTile == TileWater || spec.SpecialCaveSpawnChance <= 0 || spec.SpecialCaveMaxCount <= 0 {
			continue
		}

		var candidates []pos
		for x := 5; x < w.Width-5; x++ {
			for y := 5; y < w.Height-5; y++ {
				if w.OverworldMap[x][y] == TileWater && w.BiomeMap[x][y] == bID && w.isOceanArea(x, y) && w.LandDist[x][y] >= 3 {
					candidates = append(candidates, pos{x, y})
				}
			}
		}

		if len(candidates) == 0 {
			continue
		}

		r.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		var spawned []pos
		for _, c := range candidates {
			if len(spawned) >= spec.SpecialCaveMaxCount {
				break
			}

			if r.Float64() < spec.SpecialCaveSpawnChance {
				// Check min distance against already placed caves of this type
				tooClose := false
				if spec.SpecialCaveMinDist > 0 {
					for _, sp := range spawned {
						dx := float64(c.x - sp.x)
						dy := float64(c.y - sp.y)
						if math.Hypot(dx, dy) < spec.SpecialCaveMinDist {
							tooClose = true
							break
						}
					}
				}

				if !tooClose {
					w.OverworldMap[c.x][c.y] = spec.SpecialCaveTile
					spawned = append(spawned, c)
				}
			}
		}

		// Ensure at least 1 spawns if candidates exist and maxCount > 0
		if len(spawned) == 0 && spec.SpecialCaveMaxCount > 0 && len(candidates) > 0 {
			first := candidates[0]
			w.OverworldMap[first.x][first.y] = spec.SpecialCaveTile
		}
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

// FindLifepodSpawn finds the Shallow Coral Reef water tile nearest the center of the world map.
// If no ShallowReef water tile is available, it falls back to the nearest water tile to the center.
func (w *World) FindLifepodSpawn() (spawnTX, spawnTY int) {
	centerX := float64(w.Width) / 2.0
	centerY := float64(w.Height) / 2.0

	bestDist := math.MaxFloat64
	bestTX, bestTY := w.Width/2, w.Height/2
	foundShallow := false

	// Primary pass: look for Shallow Coral Reef water tiles
	for x := 5; x < w.Width-5; x++ {
		for y := 5; y < w.Height-5; y++ {
			if w.OverworldMap[x][y] == TileWater && w.BiomeMap[x][y] == BiomeShallowReef {
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				dist := math.Hypot(dx, dy)
				if dist < bestDist {
					bestDist = dist
					bestTX, bestTY = x, y
					foundShallow = true
				}
			}
		}
	}

	if foundShallow {
		return bestTX, bestTY
	}

	// Fallback pass: any water tile nearest center
	bestDist = math.MaxFloat64
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			if w.OverworldMap[x][y] == TileWater {
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				dist := math.Hypot(dx, dy)
				if dist < bestDist {
					bestDist = dist
					bestTX, bestTY = x, y
				}
			}
		}
	}

	return bestTX, bestTY
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

// generateBiomes initializes the biome map across the overworld.
func (w *World) generateBiomes() {
	w.BiomeMap = make([][]BiomeID, w.Width)
	for x := 0; x < w.Width; x++ {
		w.BiomeMap[x] = make([]BiomeID, w.Height)
	}

	biomeNoise := NewNoise2D(w.Seed + 27)

	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			// Noise-based biome assignment
			nx := float64(x) / 25.0
			ny := float64(y) / 25.0
			val1 := biomeNoise.FBM(nx, ny, 2)

			nx2 := float64(x+50) / 20.0
			ny2 := float64(y+50) / 20.0
			val2 := biomeNoise.FBM(nx2, ny2, 2)

			if val1 < 0.40 {
				w.BiomeMap[x][y] = BiomeShallowReef
			} else if val1 < 0.65 {
				if val2 > 0.50 {
					w.BiomeMap[x][y] = BiomeKelpForest
				} else {
					w.BiomeMap[x][y] = BiomeThermalBarrens
				}
			} else {
				w.BiomeMap[x][y] = BiomeAbyssalBlue
			}
		}
	}
}

// GetSmoothedWaterOffset calculates the neighborhood averaged water color offset for tile (tx, ty).
func (w *World) GetSmoothedWaterOffset(tx, ty int) ColorOffset {
	var totalR, totalG, totalB float64
	var count float64

	radius := BiomeBlendRadius
	if radius < 0 {
		radius = 0
	}
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			nx := tx + dx
			ny := ty + dy

			if nx < 0 {
				nx = 0
			} else if nx >= w.Width {
				nx = w.Width - 1
			}
			if ny < 0 {
				ny = 0
			} else if ny >= w.Height {
				ny = w.Height - 1
			}

			bID := w.BiomeMap[nx][ny]
			spec := GetBiomeInfo(bID)
			totalR += spec.WaterColorOffset.R
			totalG += spec.WaterColorOffset.G
			totalB += spec.WaterColorOffset.B
			count++
		}
	}

	if count == 0 {
		return ColorOffset{}
	}
	return ColorOffset{
		R: (totalR / count) * BiomeTransitionIntensity,
		G: (totalG / count) * BiomeTransitionIntensity,
		B: (totalB / count) * BiomeTransitionIntensity,
	}
}
