package resource

import (
	"testing"
)

func TestDefaultGenConfig(t *testing.T) {
	// Verify default config structure
	if GenConfig.BaseHitsToMine != 3 {
		t.Errorf("expected BaseHitsToMine to be 3, got %d", GenConfig.BaseHitsToMine)
	}
	if GenConfig.HitsDepthScale != 30 {
		t.Errorf("expected HitsDepthScale to be 30, got %d", GenConfig.HitsDepthScale)
	}
	if len(GenConfig.Tiers) != 4 {
		t.Errorf("expected 4 tiers, got %d", len(GenConfig.Tiers))
	}
}

func TestGenerateResourceNodesWithConfig(t *testing.T) {
	// Create a dummy grid: 10x10 all solid rock
	grid := make([][]bool, 10)
	for x := range grid {
		grid[x] = make([]bool, 10)
		for y := range grid[x] {
			grid[x][y] = true
		}
	}
	// Make (5, 5) empty water so adjacent blocks are "exposed"
	grid[5][5] = false

	// Backup current config and restore after test
	originalConfig := GenConfig
	defer func() {
		GenConfig = originalConfig
	}()

	// Apply a test configuration that spawns resources 100% of the time,
	// only Titanium, and with 5 base hits.
	GenConfig = ResourceGenConfig{
		FallbackSpawnChance: 1.0,
		BaseHitsToMine:      5,
		HitsDepthScale:      10,
		Tiers: []ResourceTier{
			{
				MaxDepth:         100,
				SpawnChance:      1.0,
				TitaniumWeight:   1.0,
				CopperWeight:     0.0,
				QuartzWeight:     0.0,
				NickelWeight:     0.0,
				AbyssalOreWeight: 0.0,
			},
		},
	}

	nodes := GenerateResourceNodes(grid, 42)

	// Since we set spawn chance to 100%, and grid has rock adjacent to empty water at (5,5),
	// we expect some resource nodes to generate.
	if len(nodes) == 0 {
		t.Fatal("expected some nodes to be generated under 100% spawn chance, got 0")
	}

	for _, node := range nodes {
		// All nodes must be Titanium
		if node.GetName() != "Titanium" {
			t.Errorf("expected generated node to be Titanium, got %s", node.GetName())
		}
		// BaseHitsToMine is 5, scale is 10, depth is around 4-6, so hits to mine should be >= 5
		if node.GetHitsToMine() < 5 {
			t.Errorf("expected node hits to mine to be >= 5, got %d", node.GetHitsToMine())
		}
	}
}

func TestBlueprintNodeSpawning(t *testing.T) {
	// Create a dummy grid representing a wreckage cave layout (60x120)
	grid := make([][]bool, 60)
	for x := range grid {
		grid[x] = make([]bool, 120)
		for y := range grid[x] {
			grid[x][y] = true // solid bulkhead
		}
	}

	// Carve two rooms: one upper deck, one lower deck
	// Upper room floor at Y = 22, carved open space at Y = 21
	for tx := 5; tx <= 15; tx++ {
		grid[tx][21] = false
	}
	// Lower room floor at Y = 78, carved open space at Y = 77
	for tx := 5; tx <= 15; tx++ {
		grid[tx][77] = false
	}

	// Test Ship 0 (Tier 1 blueprints only, upper decks only)
	nodesShip0 := GenerateWreckageResources(grid, 42, 0)
	var bpNodesShip0 []*BlueprintNode
	for _, n := range nodesShip0 {
		if bp, ok := n.(*BlueprintNode); ok {
			bpNodesShip0 = append(bpNodesShip0, bp)
		}
	}

	if len(bpNodesShip0) == 0 {
		t.Fatal("expected blueprints to spawn in Ship 0, got 0")
	}

	t2Recipes := map[string]bool{
		"Heavy Mech Kit": true,
		"Escape Rocket":  true,
	}

	for _, bp := range bpNodesShip0 {
		// Verify only Tier 1 blueprints spawn
		if t2Recipes[bp.RecipeResultName] {
			t.Errorf("Ship 0 spawned Tier 2 recipe: %s", bp.RecipeResultName)
		}
		// Verify spawned on upper deck (Y <= 51)
		_, ty := bp.GetTilePos()
		if ty > 51 {
			t.Errorf("Ship 0 blueprint spawned on lower deck: Y = %d", ty)
		}
	}

	// Test Ship 2 (Tier 2 blueprints only, lower decks only)
	nodesShip2 := GenerateWreckageResources(grid, 42, 2)
	var bpNodesShip2 []*BlueprintNode
	for _, n := range nodesShip2 {
		if bp, ok := n.(*BlueprintNode); ok {
			bpNodesShip2 = append(bpNodesShip2, bp)
		}
	}

	if len(bpNodesShip2) == 0 {
		t.Fatal("expected blueprints to spawn in Ship 2, got 0")
	}

	for _, bp := range bpNodesShip2 {
		// Verify only Tier 2 blueprints spawn
		if !t2Recipes[bp.RecipeResultName] {
			t.Errorf("Ship 2 spawned Tier 1 recipe: %s", bp.RecipeResultName)
		}
		// Verify spawned on lower deck (Y > 51)
		_, ty := bp.GetTilePos()
		if ty <= 51 {
			t.Errorf("Ship 2 blueprint spawned on upper deck: Y = %d", ty)
		}
	}

	// Check duplicate prevention in the same ship
	seen := make(map[string]bool)
	for _, bp := range bpNodesShip0 {
		if seen[bp.RecipeResultName] {
			t.Errorf("Ship 0 spawned duplicate blueprint for: %s", bp.RecipeResultName)
		}
		seen[bp.RecipeResultName] = true
	}
}
