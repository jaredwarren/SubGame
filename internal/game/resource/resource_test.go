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
