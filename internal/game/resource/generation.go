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
		for ty := 1; ty < gridH-2; ty++ {
			if !grid[tx][ty] && !isWreckageCorridorOrShaft(tx, ty) { // open tile in a room
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

	// Spawn Blueprints
	t1UtilityRecipes := []string{
		"Ultra High Capacity O2 Tank",
		"Solar Array MKII Module",
		"Storage Vault MKII Module",
		"Sonar Amplifier",
		"Thermal Generator",
		"Surface Sonar Module",
	}

	var selected []string
	if shipIndex == 0 {
		// Guaranteed Scout Sub Kit in Ship 0 (Bridge / upper decks)
		selected = append(selected, "Scout Sub Kit")
		shuffled := sliceutil.Shuffle(t1UtilityRecipes, r)
		numExtra := 2 + r.Intn(2) // 2 or 3 extra T1 blueprints (total 3-4)
		if numExtra > len(shuffled) {
			numExtra = len(shuffled)
		}
		selected = append(selected, shuffled[:numExtra]...)
	} else if shipIndex == 1 {
		// Guaranteed Heavy Mech Kit (lower deck) and Scout Sub Depth Module MK1 (upper deck)
		selected = append(selected, "Scout Sub Depth Module MK1", "Heavy Mech Kit")
		allT1 := append([]string{"Thermal Generator", "Solar Array MKII Module", "Surface Sonar Module"}, t1UtilityRecipes...)
		uniqueT1 := []string{}
		seenT1 := map[string]bool{
			"Heavy Mech Kit":             true,
			"Scout Sub Depth Module MK1": true,
		}
		for _, name := range allT1 {
			if !seenT1[name] {
				seenT1[name] = true
				uniqueT1 = append(uniqueT1, name)
			}
		}
		shuffled := sliceutil.Shuffle(uniqueT1, r)
		numExtra := 1 + r.Intn(2) // 1 or 2 extra blueprints (total 3-4)
		if numExtra > len(shuffled) {
			numExtra = len(shuffled)
		}
		selected = append(selected, shuffled[:numExtra]...)
	} else if shipIndex == 2 {
		// Guaranteed Escape Rocket in Ship 2 placed strictly inside the Deep Vault room,
		// with its entrance doorway completely sealed by Reinforced Blast Bulkheads.
		rooms := FindWreckageRooms(grid)
		var bestRoom *WreckageRoom
		maxDepth := -1
		for i := range rooms {
			rm := &rooms[i]
			if (len(rm.FloorTiles) > 0 || len(rm.Tiles) > 0) && rm.MaxY > maxDepth {
				maxDepth = rm.MaxY
				bestRoom = rm
			}
		}

		if bestRoom != nil {
			var bpTile [2]int
			if len(bestRoom.FloorTiles) > 0 {
				bpTile = bestRoom.FloorTiles[r.Intn(len(bestRoom.FloorTiles))]
			} else {
				bpTile = bestRoom.Tiles[r.Intn(len(bestRoom.Tiles))]
			}
			removeNodeAt := func(tx, ty int) {
				for idx, n := range nodes {
					ntx, nty := n.GetTilePos()
					if ntx == tx && nty == ty {
						nodes = append(nodes[:idx], nodes[idx+1:]...)
						return
					}
				}
			}
			removeNodeAt(bpTile[0], bpTile[1])
			nodes = append(nodes, NewBlueprintNode(bpTile[0], bpTile[1], "Escape Rocket"))

			if len(bestRoom.DoorwayTiles) > 0 {
				for _, dTile := range bestRoom.DoorwayTiles {
					removeNodeAt(dTile[0], dTile[1])
					nodes = append(nodes, NewBulkheadNode(dTile[0], dTile[1]))
				}
			} else {
				for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					nx, ny := bpTile[0]+d[0], bpTile[1]+d[1]
					if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH && !grid[nx][ny] {
						removeNodeAt(nx, ny)
						nodes = append(nodes, NewBulkheadNode(nx, ny))
						break
					}
				}
			}
		}
	}

	for _, recipeName := range selected {
		isTier2 := (recipeName == "Heavy Mech Kit" || recipeName == "Escape Rocket")
		var floorList *[][2]int
		if isTier2 {
			floorList = &lowerFloors
		} else {
			floorList = &upperFloors
		}

		if len(*floorList) > 0 {
			var chosenTile [2]int
			found := false

			// Special positioning:
			// For Scout Sub Kit in Ship 0: prefer topmost floor tile (Bridge)
			// For Heavy Mech Kit in Ship 1: prefer deepest floor tile (Engineering Bay)
			// For Escape Rocket in Ship 2: prefer deepest floor tile (Deep Vault)
			if recipeName == "Scout Sub Kit" {
				// Spawn Scout Sub Kit close to depth 40m in those rooms, but strictly above 40m (ty < 40)
				maxShallowY := -1
				for _, tile := range *floorList {
					if tile[1] < 40 && tile[1] > maxShallowY {
						maxShallowY = tile[1]
					}
				}

				var closeCandidates []int
				var anyShallowCandidates []int
				for idx, tile := range *floorList {
					if !isOccupied(tile[0], tile[1]) && tile[1] < 40 {
						anyShallowCandidates = append(anyShallowCandidates, idx)
						// Rooms close to 40m (within 5 tiles of deepest shallow floor)
						if maxShallowY >= 0 && tile[1] >= maxShallowY-5 {
							closeCandidates = append(closeCandidates, idx)
						}
					}
				}

				candidates := closeCandidates
				if len(candidates) == 0 {
					candidates = anyShallowCandidates
				}

				if len(candidates) > 0 {
					chosenIdx := candidates[r.Intn(len(candidates))]
					chosenTile = (*floorList)[chosenIdx]
					found = true
					*floorList = append((*floorList)[:chosenIdx], (*floorList)[chosenIdx+1:]...)
				}
			} else if recipeName == "Scout Sub Depth Module MK1" {
				var shallowCandidates []int
				for idx, tile := range *floorList {
					if !isOccupied(tile[0], tile[1]) && tile[1] < 40 {
						shallowCandidates = append(shallowCandidates, idx)
					}
				}
				if len(shallowCandidates) > 0 {
					chosenIdx := shallowCandidates[r.Intn(len(shallowCandidates))]
					chosenTile = (*floorList)[chosenIdx]
					found = true
					*floorList = append((*floorList)[:chosenIdx], (*floorList)[chosenIdx+1:]...)
				}
			} else if isTier2 {
				bestIdx := -1
				maxY := -1
				for idx, tile := range *floorList {
					if !isOccupied(tile[0], tile[1]) && tile[1] > maxY {
						maxY = tile[1]
						bestIdx = idx
					}
				}
				if bestIdx >= 0 {
					chosenTile = (*floorList)[bestIdx]
					found = true
					*floorList = append((*floorList)[:bestIdx], (*floorList)[bestIdx+1:]...)
				}
			}

			// Fallback: random available tile in the floorList
			if !found {
				shuffledIndices := r.Perm(len(*floorList))
				for _, idx := range shuffledIndices {
					tile := (*floorList)[idx]
					if !isOccupied(tile[0], tile[1]) {
						chosenTile = tile
						found = true
						*floorList = append((*floorList)[:idx], (*floorList)[idx+1:]...)
						break
					}
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

// WreckageRoom represents a room carved into the ship's hull, including its interior and doorway neck.
type WreckageRoom struct {
	Tiles        [][2]int
	FloorTiles   [][2]int
	DoorwayTiles [][2]int
	MinX, MaxX   int
	MinY, MaxY   int
}

// isWreckageCorridorOrShaft returns true if (tx, ty) is part of the central shaft or deck corridors.
func isWreckageCorridorOrShaft(tx, ty int) bool {
	// Central elevator shaft
	if tx >= 27 && tx <= 32 {
		return true
	}
	// Horizontal corridors (decks)
	if (ty >= 2 && ty <= 3) ||
		(ty >= 24 && ty <= 27) ||
		(ty >= 52 && ty <= 55) ||
		(ty >= 80 && ty <= 83) ||
		(ty >= 108 && ty <= 111) {
		return true
	}
	return false
}

// FindWreckageRooms finds all rooms in a wreckage grid and identifies their floor and doorway tiles.
func FindWreckageRooms(grid [][]bool) []WreckageRoom {
	gridW := len(grid)
	if gridW == 0 {
		return nil
	}
	gridH := len(grid[0])

	visited := make([][]bool, gridW)
	for x := range visited {
		visited[x] = make([]bool, gridH)
	}

	var rooms []WreckageRoom

	for x := 1; x < gridW-1; x++ {
		for y := 1; y < gridH-1; y++ {
			if grid[x][y] || isWreckageCorridorOrShaft(x, y) || visited[x][y] {
				continue
			}

			// BFS connected component for this room
			queue := [][2]int{{x, y}}
			visited[x][y] = true
			var comp [][2]int

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]
				comp = append(comp, curr)

				dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
				for _, d := range dirs {
					nx, ny := curr[0] + d[0], curr[1] + d[1]
					if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH && !grid[nx][ny] && !isWreckageCorridorOrShaft(nx, ny) && !visited[nx][ny] {
						visited[nx][ny] = true
						queue = append(queue, [2]int{nx, ny})
					}
				}
			}

			if len(comp) < 4 {
				continue
			}

			// In this room component, identify doorway tiles (tiles adjacent to an open corridor)
			var doorwayTiles [][2]int
			isDoorway := make(map[[2]int]bool)
			minX, maxX := 999, -1
			minY, maxY := 999, -1

			for _, pt := range comp {
				px, py := pt[0], pt[1]
				if px < minX {
					minX = px
				}
				if px > maxX {
					maxX = px
				}
				if py < minY {
					minY = py
				}
				if py > maxY {
					maxY = py
				}

				dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
				for _, d := range dirs {
					nx, ny := px + d[0], py + d[1]
					if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH && !grid[nx][ny] && isWreckageCorridorOrShaft(nx, ny) {
						doorwayTiles = append(doorwayTiles, pt)
						isDoorway[pt] = true
						break
					}
				}
			}

			// Room floor tiles: non-doorway tiles that have a solid tile directly beneath them
			var floorTiles [][2]int
			for _, pt := range comp {
				if isDoorway[pt] {
					continue
				}
				px, py := pt[0], pt[1]
				if py+1 < gridH && grid[px][py+1] {
					floorTiles = append(floorTiles, pt)
				}
			}

			rooms = append(rooms, WreckageRoom{
				Tiles:        comp,
				FloorTiles:   floorTiles,
				DoorwayTiles: doorwayTiles,
				MinX:         minX,
				MaxX:         maxX,
				MinY:         minY,
				MaxY:         maxY,
			})
		}
	}

	return rooms
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
