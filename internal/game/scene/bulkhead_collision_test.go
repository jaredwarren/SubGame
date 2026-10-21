package scene

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/gvec"
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

func TestWreckageOuterChamferPassable(t *testing.T) {
	grid := make([][]bool, 20)
	for x := range grid {
		grid[x] = make([]bool, 20)
	}

	// Make tile (5, 5) a solid tile surrounded by water
	grid[5][5] = true

	cs := &CaveScene{
		CaveGrid:   grid,
		ActiveCave: cave.NewWreckageCorridorCave(grid, 0),
	}

	ts := float64(config.TileSize) // 64
	tileX := 5.0 * ts
	tileY := 5.0 * ts

	// The NW corner of tile (5,5) has a 45° chamfer cut of size WreckChamferSize (22px).
	// A small 4x4 bounding box situated at (tileX + 2, tileY + 2) is within the cut-off triangle:
	// relX + relY = 2 + 2 = 4 <= 22, so it should NOT collide.
	if cs.IsSolid(nil, tileX+2, tileY+2, 4, 4) {
		t.Fatalf("expected cut-off corner at (%.1f, %.1f) to be passable under 45-degree chamfer", tileX+2, tileY+2)
	}

	// Deep inside the tile (tileX + 32, tileY + 32), it MUST be solid
	if !cs.IsSolid(nil, tileX+32, tileY+32, 4, 4) {
		t.Fatalf("expected center of solid wreckage tile at (%.1f, %.1f) to be solid", tileX+32, tileY+32)
	}
}

func TestWreckageSlidingDeflection(t *testing.T) {
	grid := make([][]bool, 20)
	for x := range grid {
		grid[x] = make([]bool, 20)
	}
	grid[5][5] = true

	cs := &CaveScene{
		CaveGrid:   grid,
		ActiveCave: cave.NewWreckageCorridorCave(grid, 0),
	}

	ts := float64(config.TileSize)
	tileX := 5.0 * ts
	tileY := 5.0 * ts

	// Position player just outside the NW chamfer of tile (5,5).
	// Tile corner is at (tileX, tileY). The chamfer line connects (tileX+22, tileY) and (tileX, tileY+22).
	// Player has 12x12 dimensions, moving southeast (+X, +Y) directly into the chamfer.
	p := &player.Player{
		Pos:    gvec.Vec2{X: tileX - 4, Y: tileY - 4},
		Vel:    gvec.Vec2{X: 10, Y: 2},
		Width:  12,
		Height: 12,
	}

	initialPos := p.Pos
	cs.checkCollisions(nil, p)

	// In a wreckage corridor with 45° chamfer, swimming into the chamfer at an angle
	// deflects player velocity tangentially along the slope (-X, -Y normal => +X, -Y slide).
	if p.Pos.X <= initialPos.X || p.Pos.Y >= initialPos.Y {
		t.Fatalf("expected player to slide up-right along 45° chamfer, initial: %v, new: %v", initialPos, p.Pos)
	}
}

