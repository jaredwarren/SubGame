package resource

import (
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/sliceutil"
)

// ResourceSpawnEntry holds weighted resource spawning properties.
type ResourceSpawnEntry struct {
	Type   NodeType
	Weight float64
}

func selectWeightedResource(entries []ResourceSpawnEntry, roll float64) NodeType {
	if len(entries) == 0 {
		return NodeTitanium
	}
	var total float64
	for _, e := range entries {
		total += e.Weight
	}
	if total <= 0 {
		return entries[0].Type
	}
	target := roll * total
	var current float64
	for _, e := range entries {
		current += e.Weight
		if target <= current {
			return e.Type
		}
	}
	return entries[len(entries)-1].Type
}

// ResourceTier defines configuration for resource spawning at a specific depth range.
type ResourceTier struct {
	MaxDepth    int                  // The maximum depth (exclusive threshold, e.g. ty < MaxDepth)
	SpawnChance float64              // The density/chance of spawning a resource on an exposed rock tile
	Entries     []ResourceSpawnEntry // Weighted ore types for this depth band
}

// ResourceGenConfig holds the configuration parameters for resource generation.
type ResourceGenConfig struct {
	FallbackSpawnChance float64              // Fallback spawn chance if no tier matches
	BaseHitsToMine      int                  // Base health/hits to mine a node
	HitsDepthScale      int                  // Scaling factor for health: depth / HitsDepthScale
	Tiers               []ResourceTier       // List of depth-based configuration tiers (ordered by MaxDepth ascending)
	WreckageSpawnChance float64              // Spawn chance for wreckage resources on wreckage floor tiles
	WreckageEntries     []ResourceSpawnEntry // Weighted scrap / e-waste for wreckage floors
}

// DefaultGenConfig represents the default generation settings matching the original game balance.
var DefaultGenConfig = ResourceGenConfig{
	FallbackSpawnChance: 0.05,
	BaseHitsToMine:      3,
	HitsDepthScale:      30,
	Tiers: []ResourceTier{
		{
			MaxDepth:    30,
			SpawnChance: 0.04,
			Entries: []ResourceSpawnEntry{
				{Type: NodeTitanium, Weight: 70.0},
				{Type: NodeCopper, Weight: 30.0},
			},
		},
		{
			MaxDepth:    60,
			SpawnChance: 0.055,
			Entries: []ResourceSpawnEntry{
				{Type: NodeTitanium, Weight: 28.0},
				{Type: NodeCopper, Weight: 32.0},
				{Type: NodeQuartz, Weight: 18.0},
				{Type: NodeNickel, Weight: 12.0},
				{Type: NodeTungsten, Weight: 10.0},
			},
		},
		{
			MaxDepth:    90,
			SpawnChance: 0.07,
			Entries: []ResourceSpawnEntry{
				{Type: NodeTitanium, Weight: 22.0},
				{Type: NodeCopper, Weight: 22.0},
				{Type: NodeQuartz, Weight: 22.0},
				{Type: NodeNickel, Weight: 12.0},
				{Type: NodeTungsten, Weight: 12.0},
				{Type: NodeAbyssalOre, Weight: 10.0},
			},
		},
		{
			MaxDepth:    999999, // Catch-all for super deep zones
			SpawnChance: 0.085,
			Entries: []ResourceSpawnEntry{
				{Type: NodeTitanium, Weight: 12.0},
				{Type: NodeCopper, Weight: 12.0},
				{Type: NodeQuartz, Weight: 26.0},
				{Type: NodeNickel, Weight: 12.0},
				{Type: NodeTungsten, Weight: 13.0},
				{Type: NodeAbyssalOre, Weight: 25.0},
			},
		},
	},
	WreckageSpawnChance: 0.08,
	WreckageEntries: []ResourceSpawnEntry{
		{Type: NodeScrapMetal, Weight: 65.0},
		{Type: NodeElectronicWaste, Weight: 35.0},
	},
}

// GenConfig is the active resource generation configuration.
// It can be adjusted at runtime to easily change spawning behavior.
var GenConfig = DefaultGenConfig

// GenerateWreckageResources spawns scrap metal and electronic waste nodes on room floors in wreckage caves,
// and also spawns appropriate recipe blueprints depending on the shipIndex (0, 1, or 2).
func GenerateWreckageResources(grid [][]bool, seed int64, shipIndex int) []Resource {
	nodes := []Resource{}
	if grid == nil {
		return nodes
	}
	gridW := len(grid)
	gridH := len(grid[0])
	r := rand.New(rand.NewSource(seed))

	// Find room floor tiles (open space above a solid tile, not in the central elevator shaft)
	var upperFloors [][2]int
	var lowerFloors [][2]int

	for tx := 1; tx < gridW-1; tx++ {
		// Central elevator shaft is tx 27..32
		if tx >= 27 && tx <= 32 {
			continue
		}
		for ty := 1; ty < gridH-2; ty++ {
			if !grid[tx][ty] { // open tile
				if grid[tx][ty+1] { // solid tile below (floor)
					if ty <= 51 {
						upperFloors = append(upperFloors, [2]int{tx, ty})
					} else {
						lowerFloors = append(lowerFloors, [2]int{tx, ty})
					}

					if r.Float64() < GenConfig.WreckageSpawnChance {
						kind := selectWeightedResource(GenConfig.WreckageEntries, r.Float64())
						node := NewNode(kind, tx, ty)
						// Scale hits with depth
						node.SetHitsToMine(GenConfig.BaseHitsToMine + (ty / GenConfig.HitsDepthScale))
						nodes = append(nodes, node)
					}
				}
			}
		}
	}

	// Spawn Blueprints
	t1Recipes := []string{
		"Ultra High Capacity O2 Tank",
		"Scout Sub Kit",
		"Solar Array MKII Module",
		"Storage Vault MKII Module",
		"Sonar Amplifier",
		"Thermal Generator",
		"Surface Sonar Module",
	}
	t2Recipes := []string{
		"Heavy Mech Kit",
		"Escape Rocket",
	}

	var selected []string
	if shipIndex == 0 {
		shuffled := sliceutil.Shuffle(t1Recipes, r)
		numToSpawn := 3 + r.Intn(2) // 3 or 4
		if numToSpawn > len(shuffled) {
			numToSpawn = len(shuffled)
		}
		selected = shuffled[:numToSpawn]
	} else if shipIndex == 1 {
		allRecipes := append([]string{}, t1Recipes...)
		allRecipes = append(allRecipes, t2Recipes...)
		shuffled := sliceutil.Shuffle(allRecipes, r)
		numToSpawn := 4 + r.Intn(2) // 4 or 5
		if numToSpawn > len(shuffled) {
			numToSpawn = len(shuffled)
		}
		selected = shuffled[:numToSpawn]
	} else if shipIndex == 2 {
		selected = append([]string{}, t2Recipes...)
	}

	// Helper to check if a tile is already occupied by a spawned node
	isOccupied := func(tx, ty int) bool {
		for _, n := range nodes {
			ntx, nty := n.GetTilePos()
			if ntx == tx && nty == ty {
				return true
			}
		}
		return false
	}

	for _, recipeName := range selected {
		// Determine tier
		isTier2 := false
		for _, name := range t2Recipes {
			if name == recipeName {
				isTier2 = true
				break
			}
		}

		var floorList *[][2]int
		if isTier2 {
			floorList = &lowerFloors
		} else {
			floorList = &upperFloors
		}

		if len(*floorList) > 0 {
			// Find a non-occupied random floor tile
			shuffledIndices := r.Perm(len(*floorList))
			var chosenTile [2]int
			found := false
			for _, idx := range shuffledIndices {
				tile := (*floorList)[idx]
				if !isOccupied(tile[0], tile[1]) {
					chosenTile = tile
					found = true
					// Remove the chosen tile from the list to avoid duplicate blueprint placement
					*floorList = append((*floorList)[:idx], (*floorList)[idx+1:]...)
					break
				}
			}

			if found {
				bpNode := NewBlueprintNode(chosenTile[0], chosenTile[1], recipeName)
				nodes = append(nodes, bpNode)
			}
		}
	}

	return nodes
}

// GenerateResourceNodes scans the cave tile grid and generates mineral nodes on exposed wall surfaces.
func GenerateResourceNodes(grid [][]bool, seed int64) []Resource {
	return GenerateResourceNodesWithBiome(grid, seed, nil)
}

// GenerateResourceNodesWithBiome scans the cave tile grid and generates mineral nodes using optional biome spawn weights.
func GenerateResourceNodesWithBiome(grid [][]bool, seed int64, mineralSpawns []ResourceSpawnEntry) []Resource {
	nodes := []Resource{}
	if grid == nil {
		return nodes
	}
	gridW := len(grid)
	gridH := len(grid[0])

	r := rand.New(rand.NewSource(seed))

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			// Place nodes in open (water) tiles that are adjacent to solid walls
			if !grid[tx][ty] {
				// Check which cardinal neighbors are solid blocks
				var possibleDirs []AttachDirection
				if grid[tx][ty-1] {
					possibleDirs = append(possibleDirs, AttachTop)
				}
				if grid[tx][ty+1] {
					possibleDirs = append(possibleDirs, AttachBottom)
				}
				if grid[tx-1][ty] {
					possibleDirs = append(possibleDirs, AttachLeft)
				}
				if grid[tx+1][ty] {
					possibleDirs = append(possibleDirs, AttachRight)
				}

				if len(possibleDirs) > 0 {
					spawnRoll := r.Float64()
					kind := NodeTitanium
					var spawnChance = GenConfig.FallbackSpawnChance

					// Find the matching tier based on depth ty
					var activeTier *ResourceTier
					for i := range GenConfig.Tiers {
						if ty < GenConfig.Tiers[i].MaxDepth {
							activeTier = &GenConfig.Tiers[i]
							break
						}
					}
					if activeTier != nil {
						spawnChance = activeTier.SpawnChance
					}

					if len(mineralSpawns) > 0 {
						kind = selectWeightedResource(mineralSpawns, r.Float64())
					} else if activeTier != nil && len(activeTier.Entries) > 0 {
						kind = selectWeightedResource(activeTier.Entries, r.Float64())
					}

					if spawnRoll < spawnChance {
						// Pick one of the adjacent solid wall directions to attach to
						attachDir := possibleDirs[r.Intn(len(possibleDirs))]
						node := NewNode(kind, tx, ty)
						node.SetAttachDir(attachDir)
						// Scale node hits (health) with depth: base + depth / scale
						node.SetHitsToMine(GenConfig.BaseHitsToMine + (ty / GenConfig.HitsDepthScale))
						nodes = append(nodes, node)
					}
				}
			}
		}
	}

	return nodes
}
