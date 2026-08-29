package scene

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

func TestBulkheadCollision(t *testing.T) {
	grid := make([][]bool, 20)
	for x := range grid {
		grid[x] = make([]bool, 20)
	}

	cs := &CaveScene{
		CaveGrid: grid,
	}

	// At start with empty grid, (5, 5) tile is passable
	tx, ty := 5, 5
	worldX := float64(tx * config.TileSize)
	worldY := float64(ty * config.TileSize)

	if cs.IsSolid(nil, worldX+2, worldY+2, 12, 12) {
		t.Fatal("expected empty tile (5, 5) to not be solid")
	}

	// Add a Reinforced Blast Bulkhead node at (5, 5)
	bulkhead := resource.NewBulkheadNode(tx, ty)
	cs.Nodes = []resource.Resource{bulkhead}

	if !bulkhead.RequiresMech() {
		t.Fatal("expected bulkhead to require mech")
	}

	// Now (5, 5) must be solid
	if !cs.IsSolid(nil, worldX+2, worldY+2, 12, 12) {
		t.Fatal("expected tile with active bulkhead to be solid")
	}

	// When hits reach 0, bulkhead is no longer solid
	bulkhead.SetHitsToMine(0)
	if cs.IsSolid(nil, worldX+2, worldY+2, 12, 12) {
		t.Fatal("expected tile with destroyed bulkhead (0 hits) to not be solid")
	}
}
