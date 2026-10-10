package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
)

func TestTunnelOccurrenceAdjustment(t *testing.T) {
	t.Cleanup(ResetCaveDepthConfig)

	// Default should be 0.80
	ResetCaveDepthConfig()
	if got := GetTunnelOccurrenceChance(); got != 0.80 {
		t.Fatalf("expected default tunnel chance 0.80, got %f", got)
	}

	// Adjust to 0.0 (Off)
	SetTunnelOccurrenceChance(0.0)
	if got := GetTunnelOccurrenceChance(); got != 0.0 {
		t.Fatalf("expected 0.0, got %f", got)
	}

	// Adjust to 0.5 (Medium)
	SetTunnelOccurrenceChance(0.5)
	if got := GetTunnelOccurrenceChance(); got != 0.5 {
		t.Fatalf("expected 0.5, got %f", got)
	}

	// Clamping tests
	SetTunnelOccurrenceChance(-0.25)
	if got := GetTunnelOccurrenceChance(); got != 0.0 {
		t.Fatalf("expected clamped 0.0 for negative, got %f", got)
	}

	SetTunnelOccurrenceChance(1.75)
	if got := GetTunnelOccurrenceChance(); got != 1.0 {
		t.Fatalf("expected clamped 1.0 for overflow, got %f", got)
	}

	// Reset restores default
	ResetCaveDepthConfig()
	if got := GetTunnelOccurrenceChance(); got != 0.80 {
		t.Fatalf("expected 0.80 after reset, got %f", got)
	}
}

func TestSubGroundTunnels_OccurrenceOffVsOn(t *testing.T) {
	t.Cleanup(ResetCaveDepthConfig)

	const seed = 42

	// 1. With TunnelOccurrenceChance = 0.0, no tunnels should be carved into the rock
	SetTunnelOccurrenceChance(0.0)
	rOff := rand.New(rand.NewSource(seed))
	gridOff := GenerateShallowSeabedGrid(rOff, 10.0, true, true)

	// Check deep rock (below y=45) — with no tunnels, it must be 100% solid
	h := len(gridOff[0])
	w := len(gridOff)
	deepOpenOff := 0
	for x := 10; x < w-10; x++ {
		for y := 45; y < h-5; y++ {
			if !gridOff[x][y] {
				deepOpenOff++
			}
		}
	}
	if deepOpenOff != 0 {
		t.Fatalf("expected 0 deep open tiles when tunnel occurrence is 0.0, got %d", deepOpenOff)
	}

	// 2. With TunnelOccurrenceChance = 1.0, sub-ground tunnels MUST be carved
	SetTunnelOccurrenceChance(1.0)
	rOn := rand.New(rand.NewSource(seed))
	gridOn := GenerateShallowSeabedGrid(rOn, 10.0, true, true)

	deepOpenOn := 0
	for x := 10; x < w-10; x++ {
		for y := 30; y < h-5; y++ {
			if !gridOn[x][y] {
				deepOpenOn++
			}
		}
	}
	if deepOpenOn == 0 {
		t.Fatalf("expected sub-ground tunnels carved when tunnel occurrence is 1.0, got %d", deepOpenOn)
	}

	// 3. Verify depth grids on the cave: always background inside sub-ground tunnels
	cave := NewShallowSeabedCave(gridOn)
	bgGrid := cave.GetBgGrid()
	fgGrid := cave.GetFgGrid()
	if bgGrid == nil {
		t.Fatal("expected non-nil BgGrid for ShallowSeabedCave")
	}
	if fgGrid != nil {
		t.Fatal("expected nil FgGrid while foreground tiles are removed")
	}

	// Check that open water deep below the floor surface is ALWAYS backed by BgGrid
	for x := 10; x < w-10; x++ {
		for y := 35; y < h-5; y++ {
			if !gridOn[x][y] {
				if !bgGrid[x][y] {
					t.Fatalf("expected open subterranean tunnel tile (%d, %d) to have BgGrid=true", x, y)
				}
			}
		}
	}
}

func TestDepthGridCollisionIndependence(t *testing.T) {
	t.Cleanup(ResetCaveDepthConfig)

	SetTunnelOccurrenceChance(1.0)
	r := rand.New(rand.NewSource(999))
	grid := GenerateShallowSeabedGrid(r, 12.0, true, true)
	cave := NewShallowSeabedCave(grid)

	bgGrid := cave.GetBgGrid()
	fgGrid := cave.GetFgGrid()

	w := len(grid)
	h := len(grid[0])

	// For every tile where BgGrid or FgGrid is present:
	// It MUST be passable open water (grid[x][y] == false), NEVER obstructing player collision!
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if bgGrid[x][y] {
				if grid[x][y] {
					t.Fatalf("BgGrid tile at (%d, %d) must NOT be solid midground rock", x, y)
				}
			}
			if fgGrid != nil && fgGrid[x][y] {
				if grid[x][y] {
					t.Fatalf("FgGrid tile at (%d, %d) must NOT be solid midground rock", x, y)
				}
			}
		}
	}
}

func TestResourceAndHazardSafetyOnDepthTiles(t *testing.T) {
	t.Cleanup(ResetCaveDepthConfig)
	SetTunnelOccurrenceChance(1.0)

	seeds := []int64{101, 202, 303, 404, 505}
	for _, seed := range seeds {
		r := rand.New(rand.NewSource(seed))
		grid := GenerateShallowSeabedGrid(r, 8.0, true, true)
		c := NewShallowSeabedCave(grid)

		// Check resources
		resources := c.GenerateResources(seed)
		for _, res := range resources {
			tx, ty := res.GetTilePos()
			// Resource node coordinates must be solid or directly adjacent to solid rock
			hasSolidAdjacent := false
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					nx, ny := tx+dx, ty+dy
					if nx >= 0 && nx < len(grid) && ny >= 0 && ny < len(grid[0]) {
						if grid[nx][ny] {
							hasSolidAdjacent = true
							break
						}
					}
				}
				if hasSolidAdjacent {
					break
				}
			}
			if !hasSolidAdjacent {
				t.Fatalf("resource node %s at (%d, %d) is floating in open space without solid midground rock", res.GetName(), tx, ty)
			}
		}

		// Check entities (especially hazard flora)
		entities := c.GenerateEntities(seed)
		ts := config.TileSize
		for _, ent := range entities {
			switch e := ent.(type) {
			case *entity.ShockKelp:
				tx := int(e.Pos.X) / ts
				ty := int(e.Pos.Y) / ts
				// Must attach to solid rock
				if tx < 0 || tx >= len(grid) || ty < 0 || ty >= len(grid[0]) || (!grid[tx][ty] && (ty+1 >= len(grid[0]) || !grid[tx][ty+1])) {
					t.Fatalf("ShockKelp at (%d, %d) does not have solid rock attachment", tx, ty)
				}
			}
		}
	}
}
