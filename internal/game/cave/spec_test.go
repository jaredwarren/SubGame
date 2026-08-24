package cave

import "testing"

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
