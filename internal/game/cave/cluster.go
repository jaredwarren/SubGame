package cave

import (
	"math"
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
)

// TilePos represents a 2D integer tile coordinate.
type TilePos struct {
	X, Y int
}

// Dist returns Euclidean distance between two tile positions.
func (p TilePos) Dist(other TilePos) float64 {
	dx := float64(p.X - other.X)
	dy := float64(p.Y - other.Y)
	return math.Hypot(dx, dy)
}

// WallFace identifies which solid face an open tile is adjacent to.
type WallFace struct {
	Pos       TilePos
	Direction string // "floor", "ceiling", "left", "right"
}

// FindFloorTiles returns all open tiles that have a solid tile directly beneath them.
func FindFloorTiles(grid [][]bool) []TilePos {
	gridW := len(grid)
	if gridW == 0 {
		return nil
	}
	gridH := len(grid[0])
	var floors []TilePos

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			if !grid[tx][ty] && grid[tx][ty+1] {
				floors = append(floors, TilePos{X: tx, Y: ty})
			}
		}
	}
	return floors
}

// FindFloorRuns groups floor tiles into horizontally connected segments (ledges).
func FindFloorRuns(grid [][]bool) [][]TilePos {
	floors := FindFloorTiles(grid)
	if len(floors) == 0 {
		return nil
	}

	// Index by Y row
	rowMap := make(map[int][]int)
	for _, f := range floors {
		rowMap[f.Y] = append(rowMap[f.Y], f.X)
	}

	var runs [][]TilePos
	for y, xs := range rowMap {
		if len(xs) == 0 {
			continue
		}
		// Sort xs
		for i := 0; i < len(xs)-1; i++ {
			for j := i + 1; j < len(xs); j++ {
				if xs[i] > xs[j] {
					xs[i], xs[j] = xs[j], xs[i]
				}
			}
		}

		currentRun := []TilePos{{X: xs[0], Y: y}}
		for i := 1; i < len(xs); i++ {
			if xs[i] == xs[i-1]+1 {
				currentRun = append(currentRun, TilePos{X: xs[i], Y: y})
			} else {
				if len(currentRun) > 0 {
					runs = append(runs, currentRun)
				}
				currentRun = []TilePos{{X: xs[i], Y: y}}
			}
		}
		if len(currentRun) > 0 {
			runs = append(runs, currentRun)
		}
	}
	return runs
}

// FindOpenWaterTiles returns open tiles with a clear buffer around them (ideal for swimming shoals).
func FindOpenWaterTiles(grid [][]bool, buffer int) []TilePos {
	gridW := len(grid)
	if gridW == 0 {
		return nil
	}
	gridH := len(grid[0])
	var openTiles []TilePos

	for tx := buffer; tx < gridW-buffer; tx++ {
		for ty := buffer; ty < gridH-buffer; ty++ {
			if grid[tx][ty] {
				continue
			}
			isOpen := true
			for dx := -buffer; dx <= buffer && isOpen; dx++ {
				for dy := -buffer; dy <= buffer; dy++ {
					if grid[tx+dx][ty+dy] {
						isOpen = false
						break
					}
				}
			}
			if isOpen {
				openTiles = append(openTiles, TilePos{X: tx, Y: ty})
			}
		}
	}
	return openTiles
}

// FindWallFaces returns all adjacent solid faces for open tiles.
func FindWallFaces(grid [][]bool) []WallFace {
	gridW := len(grid)
	if gridW == 0 {
		return nil
	}
	gridH := len(grid[0])
	var faces []WallFace

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			if grid[tx][ty] {
				continue
			}
			pos := TilePos{X: tx, Y: ty}
			if grid[tx][ty+1] {
				faces = append(faces, WallFace{Pos: pos, Direction: "floor"})
			}
			if grid[tx][ty-1] {
				faces = append(faces, WallFace{Pos: pos, Direction: "ceiling"})
			}
			if grid[tx-1][ty] {
				faces = append(faces, WallFace{Pos: pos, Direction: "left"})
			}
			if grid[tx+1][ty] {
				faces = append(faces, WallFace{Pos: pos, Direction: "right"})
			}
		}
	}
	return faces
}

// SelectClusterSeeds selects N seed positions from candidates such that each seed is at least minSpacing apart.
func SelectClusterSeeds(candidates []TilePos, count int, minSpacing float64, r *rand.Rand) []TilePos {
	if len(candidates) == 0 || count <= 0 {
		return nil
	}
	shuffled := make([]TilePos, len(candidates))
	copy(shuffled, candidates)
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	var seeds []TilePos
	for _, c := range shuffled {
		if len(seeds) >= count {
			break
		}
		tooClose := false
		for _, s := range seeds {
			if c.Dist(s) < minSpacing {
				tooClose = true
				break
			}
		}
		if !tooClose {
			seeds = append(seeds, c)
		}
	}

	// If strict minSpacing resulted in fewer than desired seeds, fill from candidates
	if len(seeds) < count && len(candidates) > len(seeds) {
		for _, c := range shuffled {
			if len(seeds) >= count {
				break
			}
			already := false
			for _, s := range seeds {
				if s == c {
					already = true
					break
				}
			}
			if !already {
				seeds = append(seeds, c)
			}
		}
	}
	return seeds
}

// FloraClusterConfig configures a flora grove (e.g. kelp forest or shatterbulb bed).
type FloraClusterConfig struct {
	NumClusters    int     // Number of groves to attempt
	MinPerGrove    int     // Minimum plants per grove
	MaxPerGrove    int     // Maximum plants per grove
	MinSpacing     float64 // Minimum distance between grove seeds (in tiles)
	MinHeight      float64
	MaxHeight      float64
	BubbleO2Chance float64 // Chance for an individual stalk in the grove to have an O2 bubble (ShatterBulb)
	VarySize       bool    // If true, dynamically varies groves between small patches, medium groves, and dense forests
}

// SpawnFloraGroves spawns organic clusters of floor flora along contiguous floor ledges.
func SpawnFloraGroves(
	grid [][]bool,
	runs [][]TilePos,
	cfg FloraClusterConfig,
	floraID FloraID,
	r *rand.Rand,
) []entity.CaveEntity {
	if len(runs) == 0 || cfg.NumClusters <= 0 {
		return nil
	}

	// Build a list of candidate run centers
	type runInfo struct {
		center TilePos
		run    []TilePos
	}
	var candidates []runInfo
	for _, run := range runs {
		if len(run) >= 2 {
			mid := run[len(run)/2]
			candidates = append(candidates, runInfo{center: mid, run: run})
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	var chosen []runInfo
	for _, c := range candidates {
		if len(chosen) >= cfg.NumClusters {
			break
		}
		tooClose := false
		for _, prev := range chosen {
			if c.center.Dist(prev.center) < cfg.MinSpacing {
				tooClose = true
				break
			}
		}
		if !tooClose {
			chosen = append(chosen, c)
		}
	}
	if len(chosen) == 0 && len(candidates) > 0 {
		chosen = append(chosen, candidates[0])
	}

	var entities []entity.CaveEntity
	occupied := make(map[TilePos]bool)

	for _, c := range chosen {
		groveMinH := cfg.MinHeight
		groveMaxH := cfg.MaxHeight
		count := cfg.MinPerGrove
		if cfg.VarySize {
			roll := r.Float64()
			if roll < 0.35 {
				// Small sprout patch
				count = 2 + r.Intn(3) // 2 to 4 stalks
				groveMinH = cfg.MinHeight * 0.75
				groveMaxH = cfg.MaxHeight * 0.70
			} else if roll < 0.80 {
				// Medium standard grove
				count = 5 + r.Intn(4) // 5 to 8 stalks
				groveMinH = cfg.MinHeight
				groveMaxH = cfg.MaxHeight
			} else {
				// Giant towering forest
				count = 9 + r.Intn(6) // 9 to 14 stalks
				groveMinH = cfg.MinHeight * 1.15
				groveMaxH = cfg.MaxHeight * 1.25
			}
		} else if cfg.MaxPerGrove > cfg.MinPerGrove {
			count = cfg.MinPerGrove + r.Intn(cfg.MaxPerGrove-cfg.MinPerGrove+1)
		}
		if count > len(c.run) {
			count = len(c.run)
		}

		// Find midpoint of this run to crawl outward (Snail Swarm spread)
		midIdx := len(c.run) / 2
		indices := make([]int, 0, len(c.run))
		indices = append(indices, midIdx)
		for step := 1; len(indices) < len(c.run); step++ {
			if midIdx-step >= 0 {
				indices = append(indices, midIdx-step)
			}
			if midIdx+step < len(c.run) {
				indices = append(indices, midIdx+step)
			}
		}

		placedInGrove := 0
		for _, idx := range indices {
			if placedInGrove >= count {
				break
			}
			tile := c.run[idx]
			if occupied[tile] {
				continue
			}

			// Height variation: taller toward the center of the grove
			distFromCenter := math.Abs(float64(idx - midIdx))
			decay := math.Max(0.7, 1.0-distFromCenter*0.08)
			minH := groveMinH * decay
			maxH := groveMaxH * decay
			if maxH <= minH {
				maxH = minH + 10.0
			}
			h := minH + r.Float64()*(maxH-minH)

			plantType := floraID
			if cfg.BubbleO2Chance > 0 && r.Float64() < cfg.BubbleO2Chance {
				plantType = FloraShatterBulb
			}

			if ent := SpawnFlora(plantType, tile.X, tile.Y, h, r); ent != nil {
				entities = append(entities, ent)
				occupied[tile] = true
				placedInGrove++
			}
		}
	}
	return entities
}

// ShoalConfig defines parameters for a fish shoal cluster.
type ShoalConfig struct {
	NumShoals    int
	MinPerShoal  int
	MaxPerShoal  int
	MinSpacing   float64
	RadiusPixels float64
}

// SpawnFishShoals spawns organic clusters (schools) of swimming fauna in open water.
func SpawnFishShoals(
	grid [][]bool,
	cfg ShoalConfig,
	faunaID FaunaID,
	r *rand.Rand,
) []entity.CaveEntity {
	openTiles := FindOpenWaterTiles(grid, 1)
	if len(openTiles) == 0 || cfg.NumShoals <= 0 {
		return nil
	}

	seeds := SelectClusterSeeds(openTiles, cfg.NumShoals, cfg.MinSpacing, r)
	var entities []entity.CaveEntity

	ts := float64(config.TileSize)
	for _, seed := range seeds {
		count := cfg.MinPerShoal
		if cfg.MaxPerShoal > cfg.MinPerShoal {
			count = cfg.MinPerShoal + r.Intn(cfg.MaxPerShoal-cfg.MinPerShoal+1)
		}

		baseX := float64(seed.X)*ts + ts/2.0
		baseY := float64(seed.Y)*ts + ts/2.0

		// Synchronized base heading for the school with slight individual deviation
		schoolHeading := r.Float64() * math.Pi * 2
		schoolFacingLeft := r.Float64() < 0.5

		rad := cfg.RadiusPixels
		if rad <= 0 {
			rad = 28.0
		}

		for i := 0; i < count; i++ {
			// Offset from school center (polar jitter)
			angle := r.Float64() * math.Pi * 2
			dist := r.Float64() * rad
			px := baseX + math.Cos(angle)*dist
			py := baseY + math.Sin(angle)*dist

			// Clamp inside cave
			tx := int(px) / config.TileSize
			ty := int(py) / config.TileSize
			if tx <= 0 || tx >= len(grid)-1 || ty <= 0 || ty >= len(grid[0])-1 || grid[tx][ty] {
				px, py = baseX, baseY
			}

			var ent entity.CaveEntity
			switch faunaID {
			case FaunaPassiveFish:
				ent = entity.NewPassiveFish(
					px-10.0, py-6.0,
					schoolFacingLeft,
					schoolHeading+(r.Float64()-0.5)*0.5,
				)
			case FaunaLanternfish:
				ent = entity.NewLanternfish(
					px-9.0, py-6.0,
					schoolFacingLeft,
					schoolHeading+(r.Float64()-0.5)*0.5,
				)
			default:
				ent = SpawnFauna(faunaID, tx, ty, grid, r)
			}

			if ent != nil {
				entities = append(entities, ent)
			}
		}
	}
	return entities
}

// BenthicClusterConfig defines parameters for crawling floor fauna (e.g. crabs, hermits).
type BenthicClusterConfig struct {
	NumClusters int
	MinPerPatch int
	MaxPerPatch int
	MinSpacing  float64
}

// SpawnBenthicFaunaClusters spawns benthic creatures in small localized foraging patches on the seabed.
func SpawnBenthicFaunaClusters(
	grid [][]bool,
	runs [][]TilePos,
	cfg BenthicClusterConfig,
	faunaID FaunaID,
	r *rand.Rand,
) []entity.CaveEntity {
	if len(runs) == 0 || cfg.NumClusters <= 0 {
		return nil
	}

	var candidates []TilePos
	for _, run := range runs {
		for _, t := range run {
			candidates = append(candidates, t)
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	seeds := SelectClusterSeeds(candidates, cfg.NumClusters, cfg.MinSpacing, r)
	var entities []entity.CaveEntity
	occupied := make(map[TilePos]bool)

	for _, seed := range seeds {
		count := cfg.MinPerPatch
		if cfg.MaxPerPatch > cfg.MinPerPatch {
			count = cfg.MinPerPatch + r.Intn(cfg.MaxPerPatch-cfg.MinPerPatch+1)
		}

		// Find contiguous floor tiles around seed within radius of 3 tiles
		var patch []TilePos
		for dx := -3; dx <= 3; dx++ {
			nx := seed.X + dx
			if nx >= 1 && nx < len(grid)-1 {
				ny := seed.Y
				if !grid[nx][ny] && ny+1 < len(grid[0]) && grid[nx][ny+1] {
					patch = append(patch, TilePos{X: nx, Y: ny})
				}
			}
		}
		if len(patch) == 0 {
			patch = append(patch, seed)
		}

		r.Shuffle(len(patch), func(i, j int) {
			patch[i], patch[j] = patch[j], patch[i]
		})

		placed := 0
		for _, tile := range patch {
			if placed >= count {
				break
			}
			if occupied[tile] {
				continue
			}
			if ent := SpawnFauna(faunaID, tile.X, tile.Y, grid, r); ent != nil {
				entities = append(entities, ent)
				occupied[tile] = true
				placed++
			}
		}
	}
	return entities
}

// CoralColonyConfig defines parameters for decorative coral clusters.
type CoralColonyConfig struct {
	NumColonies  int
	MinPerColony int
	MaxPerColony int
	MinSpacing   float64
	VarySize     bool // If true, dynamically varies colonies between small outcroppings, medium clusters, and large reef banks
}

// SpawnCoralColonies spawns clustered colonies of corals attached to ledges and rock walls.
func SpawnCoralColonies(
	grid [][]bool,
	cfg CoralColonyConfig,
	biome int,
	variantCount int,
	r *rand.Rand,
) []entity.CaveEntity {
	faces := FindWallFaces(grid)
	if len(faces) == 0 || cfg.NumColonies <= 0 {
		return nil
	}

	seen := make(map[TilePos]bool)
	var uniquePos []TilePos
	for _, f := range faces {
		if !seen[f.Pos] {
			seen[f.Pos] = true
			uniquePos = append(uniquePos, f.Pos)
		}
	}

	seeds := SelectClusterSeeds(uniquePos, cfg.NumColonies, cfg.MinSpacing, r)
	var entities []entity.CaveEntity
	occupied := make(map[TilePos]bool)

	for _, seed := range seeds {
		count := cfg.MinPerColony
		searchRadius := 2
		if cfg.VarySize {
			roll := r.Float64()
			if roll < 0.35 {
				// Small outcropping
				count = 1 + r.Intn(2) // 1 to 2 corals
				searchRadius = 1
			} else if roll < 0.80 {
				// Medium colony
				count = 3 + r.Intn(3) // 3 to 5 corals
				searchRadius = 2
			} else {
				// Massive reef bank
				count = 6 + r.Intn(4) // 6 to 9 corals
				searchRadius = 3
			}
		} else if cfg.MaxPerColony > cfg.MinPerColony {
			count = cfg.MinPerColony + r.Intn(cfg.MaxPerColony-cfg.MinPerColony+1)
		}

		// Find adjacent wall face tiles within Manhattan distance searchRadius
		var colonyTiles []TilePos
		for dx := -searchRadius; dx <= searchRadius; dx++ {
			for dy := -searchRadius; dy <= searchRadius; dy++ {
				nx, ny := seed.X + dx, seed.Y + dy
				if nx >= 1 && nx < len(grid)-1 && ny >= 1 && ny < len(grid[0])-1 {
					pt := TilePos{X: nx, Y: ny}
					if !grid[nx][ny] && (grid[nx+1][ny] || grid[nx-1][ny] || grid[nx][ny+1] || grid[nx][ny-1]) {
						colonyTiles = append(colonyTiles, pt)
					}
				}
			}
		}

		r.Shuffle(len(colonyTiles), func(i, j int) {
			colonyTiles[i], colonyTiles[j] = colonyTiles[j], colonyTiles[i]
		})

		placed := 0
		for _, tile := range colonyTiles {
			if placed >= count {
				break
			}
			if occupied[tile] {
				continue
			}
			newEntities := MaybeSpawnCoral(nil, grid, tile.X, tile.Y, 1.0, biome, variantCount, r)
			if len(newEntities) > 0 {
				entities = append(entities, newEntities...)
				occupied[tile] = true
				placed++
			}
		}
	}
	return entities
}

// GenerateClusteredBiomeEntities implements the Snail Swarm clustering model for shallow and biome caves.
func GenerateClusteredBiomeEntities(
	biome *CaveBiomeSpec,
	grid [][]bool,
	coralBiome int,
	r *rand.Rand,
) []entity.CaveEntity {
	if grid == nil {
		return nil
	}
	gridW := len(grid)
	if gridW == 0 {
		return nil
	}
	gridH := len(grid[0])

	var entities []entity.CaveEntity
	runs := FindFloorRuns(grid)

	// 1. Flora Groves (Lush, varied forests: 7-10 groves, 2-15 stalks each)
	primaryFlora := FloraKelp

	kelpCfg := FloraClusterConfig{
		NumClusters:    7 + r.Intn(4), // slightly increased: 7 to 10 groves
		MinPerGrove:    6,
		MaxPerGrove:    12,
		MinSpacing:     4.5,
		MinHeight:      32.0,
		MaxHeight:      76.0,
		BubbleO2Chance: 0.12, // only 10-15% of stalks have an O2 bubble
		VarySize:       true, // varied: small sprout patches, medium groves, giant thickets
	}
	entities = append(entities, SpawnFloraGroves(grid, runs, kelpCfg, primaryFlora, r)...)

	// Standalone rare O2 bubble plant: 1 small pocket of 1-2 bulbs in a quiet nook
	rareBulbCfg := FloraClusterConfig{
		NumClusters:    1,
		MinPerGrove:    1,
		MaxPerGrove:    2,
		MinSpacing:     14.0,
		MinHeight:      42.0,
		MaxHeight:      56.0,
		BubbleO2Chance: 1.0,
	}
	entities = append(entities, SpawnFloraGroves(grid, runs, rareBulbCfg, FloraShatterBulb, r)...)

	// Shock Kelp: if biome features shock kelp, spawn 3-4 varied shock groves
	if biome != nil && (biome.ID == "kelp_forest" || biome.ID == "abyssal_blue") {
		shockCfg := FloraClusterConfig{
			NumClusters:    3 + r.Intn(2),
			MinPerGrove:    4,
			MaxPerGrove:    8,
			MinSpacing:     7.0,
			MinHeight:      36.0,
			MaxHeight:      64.0,
			VarySize:       true,
		}
		entities = append(entities, SpawnFloraGroves(grid, runs, shockCfg, FloraShockKelp, r)...)
	}

	// 2. Open Water Fauna Shoals
	shoalFauna := FaunaPassiveFish
	if biome != nil && biome.ID == "abyssal_blue" {
		shoalFauna = FaunaLanternfish
	}
	shoalCfg := ShoalConfig{
		NumShoals:    2 + r.Intn(2), // 2 to 3 schools of fish
		MinPerShoal:  3,
		MaxPerShoal:  6,
		MinSpacing:   12.0,
		RadiusPixels: 28.0,
	}
	entities = append(entities, SpawnFishShoals(grid, shoalCfg, shoalFauna, r)...)

	// 3. Benthic Fauna (Crabs / Hermits)
	hasCrabs := false
	hasViper := false
	if biome != nil {
		for _, fs := range biome.FaunaSpawns {
			if fs.Type == FaunaPassiveCrab || fs.Type == FaunaScrapHermitCrab {
				hasCrabs = true
			}
			if fs.Type == FaunaSandViper {
				hasViper = true
			}
		}
	} else {
		hasCrabs = true
	}

	if hasCrabs {
		crabType := FaunaPassiveCrab
		if biome != nil && biome.ID == "thermal_barrens" {
			crabType = FaunaScrapHermitCrab
		}
		crabCfg := BenthicClusterConfig{
			NumClusters: 1 + r.Intn(2), // 1 to 2 foraging patches
			MinPerPatch: 2,
			MaxPerPatch: 4,
			MinSpacing:  12.0,
		}
		entities = append(entities, SpawnBenthicFaunaClusters(grid, runs, crabCfg, crabType, r)...)
	}

	if hasViper {
		viperCfg := BenthicClusterConfig{
			NumClusters: 1,
			MinPerPatch: 1,
			MaxPerPatch: 2,
			MinSpacing:  15.0,
		}
		entities = append(entities, SpawnBenthicFaunaClusters(grid, runs, viperCfg, FaunaSandViper, r)...)
	}

	// 4. Coral Colonies (Slightly increased count: 6-9 colonies with varied sizes)
	coralVariants := entity.CoralShallowVariantCount
	if coralBiome == entity.CoralBiomeTrench {
		coralVariants = entity.CoralBiomeVariantCount
	}
	coralCfg := CoralColonyConfig{
		NumColonies:  6 + r.Intn(4), // 6 to 9 coral colonies
		MinPerColony: 2,
		MaxPerColony: 5,
		MinSpacing:   5.0,
		VarySize:     true, // varied: small outcroppings, medium clusters, large reef banks
	}
	entities = append(entities, SpawnCoralColonies(grid, coralCfg, coralBiome, coralVariants, r)...)

	// 5. Ink Squid Guarantee
	hasSquidInBiome := false
	if biome != nil {
		for _, s := range biome.FaunaSpawns {
			if s.Type == FaunaInkSquid {
				hasSquidInBiome = true
				break
			}
		}
	} else {
		hasSquidInBiome = true
	}

	if hasSquidInBiome {
		squidCount := 0
		for _, ent := range entities {
			if _, ok := ent.(*entity.InkSquid); ok {
				squidCount++
			}
		}
		for squidCount < 1 {
			found := false
			for attempts := 0; attempts < 100; attempts++ {
				tx := 2 + r.Intn(gridW-4)
				ty := 2 + r.Intn(gridH-4)
				if !grid[tx][ty] && !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1] {
					entities = append(entities, entity.NewInkSquid(
						float64(tx*config.TileSize)+float64(config.TileSize-22)/2.0,
						float64(ty*config.TileSize)+float64(config.TileSize-16)/2.0,
						r.Float64() < 0.5,
					))
					squidCount++
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
	}

	return entities
}

