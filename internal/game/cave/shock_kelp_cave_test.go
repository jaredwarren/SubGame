package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/entity"
)

func TestShockKelpCaveGeneration(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	grid := GenerateShockKelpCaveGrid(r)

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

func TestShockKelpCaveEntitiesAndResources(t *testing.T) {
	r := rand.New(rand.NewSource(101))
	grid := GenerateShockKelpCaveGrid(r)
	c := NewShockKelpCave(grid)

	// Verify cave type
	if c.GetCaveType() != CaveShockKelp {
		t.Errorf("expected CaveType CaveShockKelp, got %v", c.GetCaveType())
	}

	// Generate entities
	entities := c.GenerateEntities(202)
	if len(entities) == 0 {
		t.Fatal("expected some shock kelp entities to spawn, got 0")
	}

	// Verify all entities are ONLY ShockKelp, VoltaicLurker, or ElectroWeaver
	kelpCount := 0
	lurkerCount := 0
	weaverCount := 0
	for _, ent := range entities {
		switch ent.(type) {
		case *entity.ShockKelp:
			kelpCount++
		case *entity.VoltaicLurker:
			lurkerCount++
		case *entity.ElectroWeaver:
			weaverCount++
		case *entity.Coral:
			// Allowed decorative coral
		default:
			t.Errorf("expected only *entity.ShockKelp, *entity.VoltaicLurker, or *entity.ElectroWeaver to spawn, but got a different entity type: %T", ent)
		}
	}
	if kelpCount == 0 {
		t.Error("expected to spawn some ShockKelp, got 0")
	}
	if lurkerCount == 0 {
		t.Error("expected to spawn some VoltaicLurker, got 0")
	}
	// Note: since weaver count is random (0-2), we don't strictly assert weaverCount > 0,
	// but we could log it or check it. We'll just check if it's within [0, 2].
	if weaverCount < 0 || weaverCount > 2 {
		t.Errorf("expected between 0 and 2 ElectroWeavers, got %d", weaverCount)
	}

	// Generate resources
	resources := c.GenerateResources(303)
	if len(resources) == 0 {
		t.Fatal("expected resource nodes to spawn, got 0")
	}

	// Verify resource kinds: Copper, Quartz, or Titanium
	copperCount := 0
	quartzCount := 0
	titaniumCount := 0
	for _, res := range resources {
		name := res.GetName()
		switch name {
		case "Copper":
			copperCount++
		case "Quartz":
			quartzCount++
		case "Titanium":
			titaniumCount++
		default:
			t.Errorf("unexpected resource spawned in ShockKelpCave: %s", name)
		}
	}

	// Make sure we have a mix of Copper and Quartz (conductive/electrical theme)
	if copperCount == 0 {
		t.Error("expected to spawn some Copper nodes, got 0")
	}
	if quartzCount == 0 {
		t.Error("expected to spawn some Quartz nodes, got 0")
	}
}
