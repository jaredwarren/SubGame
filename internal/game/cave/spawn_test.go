package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

func TestSelectWeightedEntry(t *testing.T) {
	entries := []SpawnEntry[FaunaID]{
		{Type: FaunaPassiveFish, Weight: 100},
		{Type: FaunaSandViper, Weight: 0},
	}
	got := SelectWeightedEntry(entries, 0.5)
	if got != FaunaPassiveFish {
		t.Fatalf("got %v, want FaunaPassiveFish", got)
	}

	if SelectWeightedEntry[FaunaID](nil, 0.5) != 0 {
		t.Fatal("empty entries should return zero value")
	}
}

func TestSpawnFaunaAllIDs(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for id := FaunaID(0); id < FaunaCount; id++ {
		ent := SpawnFauna(id, 5, 5, r)
		if ent == nil {
			t.Fatalf("SpawnFauna(%d) returned nil", id)
		}
		if !ent.IsActive() {
			t.Fatalf("SpawnFauna(%d) inactive", id)
		}
	}
	if SpawnFauna(FaunaCount, 0, 0, r) != nil {
		t.Fatal("unknown fauna id should return nil")
	}
}

func TestSpawnFloraAllIDs(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for id := FloraID(0); id < FloraCount; id++ {
		ent := SpawnFlora(id, 5, 5, 40, r)
		if ent == nil {
			t.Fatalf("SpawnFlora(%d) returned nil", id)
		}
	}
	// FloraCoral historically maps to kelp
	ent := SpawnFlora(FloraCoral, 3, 3, 32, r)
	if _, ok := ent.(*entity.Kelp); !ok {
		t.Fatalf("FloraCoral should spawn Kelp, got %T", ent)
	}
}

func TestDefaultBiomeMineralTypes(t *testing.T) {
	for _, s := range DefaultShallowReefBiome.MineralSpawns {
		switch s.Type {
		case resource.NodeTitanium, resource.NodeCopper, resource.NodeQuartz,
			resource.NodeNickel, resource.NodeAbyssalOre:
			// ok
		default:
			t.Fatalf("unexpected mineral type %v", s.Type)
		}
	}
}
