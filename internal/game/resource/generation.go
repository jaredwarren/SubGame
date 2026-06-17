package resource

import (
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/sliceutil"
)

// ResourceTier defines configuration for resource spawning at a specific depth range.
type ResourceTier struct {
	MaxDepth         int     // The maximum depth (exclusive threshold, e.g. ty < MaxDepth)
	SpawnChance      float64 // The density/chance of spawning a resource on an exposed rock tile
	TitaniumWeight   float64 // Relative weight for Titanium spawning
	CopperWeight     float64 // Relative weight for Copper spawning
	QuartzWeight     float64 // Relative weight for Quartz spawning
	NickelWeight     float64 // Relative weight for Nickel spawning
	AbyssalOreWeight float64 // Relative weight for Abyssal Ore spawning
}

// ResourceGenConfig holds the configuration parameters for resource generation.
type ResourceGenConfig struct {
	FallbackSpawnChance float64        // Fallback spawn chance if no tier matches
	BaseHitsToMine      int            // Base health/hits to mine a node
	HitsDepthScale      int            // Scaling factor for health: depth / HitsDepthScale
	Tiers               []ResourceTier // List of depth-based configuration tiers (ordered by MaxDepth ascending)
	WreckageSpawnChance float64        // Spawn chance for wreckage resources on wreckage floor tiles
	ScrapMetalWeight    float64        // Relative weight for Scrap Metal in wreckage
	ElecWasteWeight     float64        // Relative weight for Electronic Waste in wreckage
}

// DefaultGenConfig represents the default generation settings matching the original game balance.
var DefaultGenConfig = ResourceGenConfig{
	FallbackSpawnChance: 0.05,
	BaseHitsToMine:      3,
	HitsDepthScale:      30,
	Tiers: []ResourceTier{
		{
			MaxDepth:         30,
			SpawnChance:      0.04,
			TitaniumWeight:   70.0,
			CopperWeight:     30.0,
			QuartzWeight:     0.0,
			NickelWeight:     0.0,
			AbyssalOreWeight: 0.0,
		},
		{
			MaxDepth:         60,
			SpawnChance:      0.055,
			TitaniumWeight:   30.0,
			CopperWeight:     35.0,
			QuartzWeight:     20.0,
			NickelWeight:     15.0,
			AbyssalOreWeight: 0.0,
		},
		{
			MaxDepth:         90,
			SpawnChance:      0.07,
			TitaniumWeight:   25.0,
			CopperWeight:     25.0,
			QuartzWeight:     25.0,
			NickelWeight:     15.0,
			AbyssalOreWeight: 10.0,
		},
		{
			MaxDepth:         999999, // Catch-all for super deep zones
			SpawnChance:      0.085,
			TitaniumWeight:   15.0,
			CopperWeight:     15.0,
			QuartzWeight:     30.0,
			NickelWeight:     15.0,
			AbyssalOreWeight: 25.0,
		},
	},
	WreckageSpawnChance: 0.08,
	ScrapMetalWeight:    65.0,
	ElecWasteWeight:     35.0,
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
						var node Resource
						totalW := GenConfig.ScrapMetalWeight + GenConfig.ElecWasteWeight
						var isScrap bool
						if totalW > 0 {
							isScrap = r.Float64()*totalW < GenConfig.ScrapMetalWeight
						} else {
							isScrap = true
						}

						if isScrap {
							node = NewScrapMetalNode(tx, ty)
						} else {
							node = NewElectronicWasteNode(tx, ty)
						}
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
	nodes := []Resource{}
	if grid == nil {
		return nodes
	}
	gridW := len(grid)
	gridH := len(grid[0])

	r := rand.New(rand.NewSource(seed))

	type nodeKind int
	const (
		kindTitanium nodeKind = iota
		kindCopper
		kindQuartz
		kindNickel
		kindAbyssalOre
	)

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
					kind := kindTitanium
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
						totalWeight := activeTier.TitaniumWeight + activeTier.CopperWeight + activeTier.QuartzWeight + activeTier.NickelWeight + activeTier.AbyssalOreWeight
						if totalWeight > 0 {
							roll := r.Float64() * totalWeight
							if roll < activeTier.TitaniumWeight {
								kind = kindTitanium
							} else if roll < activeTier.TitaniumWeight+activeTier.CopperWeight {
								kind = kindCopper
							} else if roll < activeTier.TitaniumWeight+activeTier.CopperWeight+activeTier.QuartzWeight {
								kind = kindQuartz
							} else if roll < activeTier.TitaniumWeight+activeTier.CopperWeight+activeTier.QuartzWeight+activeTier.NickelWeight {
								kind = kindNickel
							} else {
								kind = kindAbyssalOre
							}
						}
					}

					if spawnRoll < spawnChance {
						// Pick one of the adjacent solid wall directions to attach to
						attachDir := possibleDirs[r.Intn(len(possibleDirs))]
						var node Resource
						switch kind {
						case kindTitanium:
							node = NewTitaniumNode(tx, ty)
						case kindCopper:
							node = NewCopperNode(tx, ty)
						case kindQuartz:
							node = NewQuartzNode(tx, ty)
						case kindNickel:
							node = NewNickelNode(tx, ty)
						case kindAbyssalOre:
							node = NewAbyssalOreNode(tx, ty)
						}
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
