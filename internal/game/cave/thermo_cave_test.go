package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/entity"
)

func TestThermoCaveGeneration(t *testing.T) {
	r := rand.New(rand.NewSource(12345))
	grid := GenerateThermoCaveGrid(r)

	// Verify dimensions (60x60)
	if len(grid) != 60 {
		t.Errorf("expected width to be 60, got %d", len(grid))
	}
	if len(grid[0]) != 60 {
		t.Errorf("expected height to be 60, got %d", len(grid[0]))
	}

	// Verify entrance is open
	for y := 0; y < 5; y++ {
		for x := 27; x <= 33; x++ {
			if grid[x][y] {
				t.Errorf("expected entrance at (%d, %d) to be empty water, but got solid wall", x, y)
			}
		}
	}

	// Verify bounds are solid walls (except entrance at top)
	for x := 0; x < 60; x++ {
		if x != 30 && !grid[x][0] && (x < 27 || x > 33) {
			t.Errorf("expected outer top border at (%d, 0) to be solid, got empty", x)
		}
		if !grid[x][59] {
			t.Errorf("expected bottom border at (%d, 59) to be solid, got empty", x)
		}
	}
	for y := 0; y < 60; y++ {
		if !grid[0][y] {
			t.Errorf("expected left border at (0, %d) to be solid, got empty", y)
		}
		if !grid[59][y] {
			t.Errorf("expected right border at (59, %d) to be solid, got empty", y)
		}
	}
}

func TestThermoCaveEntitiesAndResources(t *testing.T) {
	r := rand.New(rand.NewSource(54321))
	grid := GenerateThermoCaveGrid(r)
	c := NewThermoCave(grid)

	// Verify cave type
	if c.GetCaveType() != CaveThermo {
		t.Errorf("expected CaveType CaveThermo, got %v", c.GetCaveType())
	}

	// Generate entities
	entities := c.GenerateEntities(999)

	// Count entities
	rammerCount := 0
	siphonCount := 0
	for _, ent := range entities {
		switch ent.(type) {
		case *entity.ThermoclineRammer:
			rammerCount++
		case *entity.BrimstoneSiphon:
			siphonCount++
		case *entity.Coral:
			// Allowed decorative coral
		default:
			t.Errorf("unexpected entity type spawned in ThermoCave: %T", ent)
		}
	}

	// Verify counts: 1-2 rammers, 4-6 siphons
	if rammerCount < 1 || rammerCount > 2 {
		t.Errorf("expected 1 or 2 ThermoclineRammers, got %d", rammerCount)
	}
	if siphonCount < 4 || siphonCount > 6 {
		t.Errorf("expected between 4 and 6 BrimstoneSiphons, got %d", siphonCount)
	}

	// Generate resources
	resources := c.GenerateResources(111)
	if len(resources) == 0 {
		t.Fatal("expected resource nodes to spawn, got 0")
	}
}
