package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
)

func TestCaveSpecRegistry(t *testing.T) {
	types := []CaveType{
		CaveOrganicShallow, CaveOrganicTrench, CaveShockKelp,
		CaveThermo, CaveWreckage, CaveVoid,
	}
	for _, ct := range types {
		s := Spec(ct)
		if s == nil {
			t.Fatalf("missing CaveSpec for %v", ct)
		}
		if s.Type != ct {
			t.Fatalf("spec type mismatch: got %v want %v", s.Type, ct)
		}
		if MusicTrack(ct) == "" {
			t.Fatalf("empty music for %v", ct)
		}
	}
}

func TestDeepEntitiesFromSpecNonNil(t *testing.T) {
	// Tiny open room with walls so banded/count spawns have candidates.
	grid := make([][]bool, 12)
	for x := range grid {
		grid[x] = make([]bool, 12)
		for y := range grid[x] {
			grid[x][y] = true
		}
	}
	for x := 3; x <= 8; x++ {
		for y := 3; y <= 8; y++ {
			grid[x][y] = false
		}
	}
	ents := GenerateDeepEntitiesFromSpec(Spec(CaveThermo), grid, 42)
	if ents == nil {
		t.Fatal("expected non-nil entity slice")
	}
}

func TestOrganicTrenchFaunaSpawns(t *testing.T) {
	r := rand.New(rand.NewSource(12345))
	grid := GenerateOrganicTrenchGrid(r)
	cave := NewOrganicTrenchCave(grid)
	ents := cave.GenerateEntities(12345)

	lanternCount := 0
	upperLanterns := 0
	lowerLanterns := 0
	squidCount := 0
	glowSquidCount := 0

	for _, ent := range ents {
		switch e := ent.(type) {
		case *entity.Lanternfish:
			lanternCount++
			ty := int(e.GetPos().Y / float64(config.TileSize))
			if ty < 40 {
				upperLanterns++
			} else {
				lowerLanterns++
			}
		case *entity.InkSquid:
			squidCount++
		case *entity.GlowSquid:
			glowSquidCount++
			ty := int(e.GetPos().Y / float64(config.TileSize))
			if ty < 60 {
				t.Errorf("GlowSquid spawned too shallow at tile Y=%d", ty)
			}
		}
	}

	t.Logf("CaveOrganicTrench entities: Lanternfish=%d (upper=%d, lower=%d), InkSquid=%d, GlowSquid=%d",
		lanternCount, upperLanterns, lowerLanterns, squidCount, glowSquidCount)

	if lanternCount == 0 {
		t.Errorf("expected Lanternfish to spawn in OrganicTrenchCave, got 0")
	}
	if upperLanterns < lowerLanterns {
		t.Errorf("expected more Lanternfish in upper grotto than deep, got upper=%d lower=%d",
			upperLanterns, lowerLanterns)
	}
}

