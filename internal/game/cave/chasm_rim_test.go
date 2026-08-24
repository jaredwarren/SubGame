package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/entity"
)

func TestChasmRimSpecFor(t *testing.T) {
	trench := ChasmRimSpecFor(CaveOrganicTrench)
	if trench == nil || trench.Target != CaveOrganicTrench {
		t.Fatal("expected organic trench rim spec")
	}
	if trench.RockColor.R != 22 {
		t.Fatalf("unexpected trench rock color: %+v", trench.RockColor)
	}

	kelp := ChasmRimSpecFor(CaveShockKelp)
	if kelp == nil || kelp.Target != CaveShockKelp {
		t.Fatal("expected shock kelp rim spec")
	}
	if kelp.VeinStyle != ChasmVeinDual {
		t.Fatal("expected dual vein style for shock kelp chasm")
	}

	// Unknown targets fall back to shock kelp palette.
	fallback := ChasmRimSpecFor(CaveVoid)
	if fallback.Target != CaveShockKelp {
		t.Fatalf("expected shock kelp fallback, got %v", fallback.Target)
	}
}

func TestSpawnChasmRimEntities_ShockKelp(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	grid := GenerateShockKelpShallowGrid(r, 10.0, true, true)
	cave := NewShockKelpShallowCave(grid)

	spec := ChasmRimSpecFor(cave.GetChasmTarget())
	entities := spawnChasmRimEntities(spec, grid, cave.ChasmX, cave.ChasmWidth, rand.New(rand.NewSource(99)))

	hasKelp := false
	for _, ent := range entities {
		if _, ok := ent.(*entity.ShockKelp); ok {
			hasKelp = true
			break
		}
	}
	if !hasKelp {
		t.Fatal("expected shock kelp rim spawns from ChasmRimSpec")
	}
}

func TestSpawnChasmRimEntities_Trench(t *testing.T) {
	r := rand.New(rand.NewSource(77))
	grid := GenerateTrenchShallowGrid(r, 10.0, true, true)
	cave := NewTrenchShallowCave(grid)

	spec := ChasmRimSpecFor(cave.GetChasmTarget())
	entities := spawnChasmRimEntities(spec, grid, cave.ChasmX, cave.ChasmWidth, rand.New(rand.NewSource(77)))

	hasBulb := false
	hasSnare := false
	for _, ent := range entities {
		switch ent.(type) {
		case *entity.ShatterBulb:
			hasBulb = true
		case *entity.FalseBulbSnare:
			hasSnare = true
		}
	}
	if !hasBulb {
		t.Error("expected shatter bulb rim spawns for trench chasm")
	}
	if !hasSnare {
		t.Error("expected false bulb snare rim spawns for trench chasm")
	}
}
