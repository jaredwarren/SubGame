package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/entity"
)

func TestFindFloorRuns(t *testing.T) {
	// Create a 20x10 grid
	grid := make([][]bool, 20)
	for x := range grid {
		grid[x] = make([]bool, 10)
	}
	// Solid floor at Y=5 for X from 4 to 12
	for x := 4; x <= 12; x++ {
		grid[x][5] = true
	}

	runs := FindFloorRuns(grid)
	if len(runs) == 0 {
		t.Fatal("expected at least 1 floor run, got 0")
	}

	// At Y=4, tiles from 4 to 12 should form a single run of length 9
	found := false
	for _, r := range runs {
		if len(r) == 9 && r[0].Y == 4 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected floor run of length 9 at Y=4, runs found: %+v", runs)
	}
}

func TestSpawnFloraGroves(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	grid := make([][]bool, 40)
	for x := range grid {
		grid[x] = make([]bool, 20)
	}
	// Floor 1: X=3 to 15 at Y=8 (solid at Y=9)
	for x := 3; x <= 15; x++ {
		grid[x][9] = true
	}
	// Floor 2: X=22 to 35 at Y=14 (solid at Y=15)
	for x := 22; x <= 35; x++ {
		grid[x][15] = true
	}

	runs := FindFloorRuns(grid)
	cfg := FloraClusterConfig{
		NumClusters: 2,
		MinPerGrove: 3,
		MaxPerGrove: 5,
		MinSpacing:  8.0,
		MinHeight:   30.0,
		MaxHeight:   50.0,
	}

	ents := SpawnFloraGroves(grid, runs, cfg, FloraKelp, r)
	if len(ents) < 6 {
		t.Errorf("expected at least 6 kelp entities across 2 groves, got %d", len(ents))
	}
	for _, e := range ents {
		if _, ok := e.(*entity.Kelp); !ok {
			t.Errorf("expected entity to be Kelp, got %T", e)
		}
	}
}

func TestSpawnFishShoals(t *testing.T) {
	r := rand.New(rand.NewSource(101))
	// 50x50 open cave with outer walls
	grid := make([][]bool, 50)
	for x := range grid {
		grid[x] = make([]bool, 50)
		grid[x][0] = true
		grid[x][49] = true
	}
	for y := 0; y < 50; y++ {
		grid[0][y] = true
		grid[49][y] = true
	}

	cfg := ShoalConfig{
		NumShoals:    2,
		MinPerShoal:  4,
		MaxPerShoal:  5,
		MinSpacing:   15.0,
		RadiusPixels: 32.0,
	}

	ents := SpawnFishShoals(grid, cfg, FaunaPassiveFish, r)
	if len(ents) < 8 {
		t.Fatalf("expected at least 8 fish in 2 shoals, got %d", len(ents))
	}

	// Verify all entities are PassiveFish
	for _, e := range ents {
		if _, ok := e.(*entity.PassiveFish); !ok {
			t.Errorf("expected PassiveFish, got %T", e)
		}
	}
}

func TestSpawnCoralColonies(t *testing.T) {
	r := rand.New(rand.NewSource(777))
	grid := make([][]bool, 30)
	for x := range grid {
		grid[x] = make([]bool, 20)
		for y := range grid[x] {
			grid[x][y] = true
		}
	}
	// Carve a central chamber from 5,5 to 24,14
	for x := 5; x <= 24; x++ {
		for y := 5; y <= 14; y++ {
			grid[x][y] = false
		}
	}

	cfg := CoralColonyConfig{
		NumColonies:  3,
		MinPerColony: 2,
		MaxPerColony: 4,
		MinSpacing:   6.0,
	}

	ents := SpawnCoralColonies(grid, cfg, entity.CoralBiomeShallow, entity.CoralShallowVariantCount, r)
	if len(ents) < 6 {
		t.Errorf("expected at least 6 corals in colonies, got %d", len(ents))
	}
	for _, e := range ents {
		if _, ok := e.(*entity.Coral); !ok {
			t.Errorf("expected Coral, got %T", e)
		}
	}
}

func TestSpawnFloraGroves_KelpAndBubbleRatio(t *testing.T) {
	r := rand.New(rand.NewSource(999))
	grid := make([][]bool, 60)
	for x := range grid {
		grid[x] = make([]bool, 20)
		for y := range grid[x] {
			grid[x][y] = false
		}
		// Floor at Y=10
		grid[x][11] = true
	}

	runs := FindFloorRuns(grid)
	cfg := FloraClusterConfig{
		NumClusters:    4,
		MinPerGrove:    8,
		MaxPerGrove:    10,
		MinSpacing:     8.0,
		MinHeight:      30.0,
		MaxHeight:      50.0,
		BubbleO2Chance: 0.15,
	}

	ents := SpawnFloraGroves(grid, runs, cfg, FloraKelp, r)
	kelpCount := 0
	bulbCount := 0
	for _, e := range ents {
		switch e.(type) {
		case *entity.Kelp:
			kelpCount++
		case *entity.ShatterBulb:
			bulbCount++
		}
	}

	if kelpCount == 0 {
		t.Fatalf("expected majority kelp, got 0")
	}
	if kelpCount <= bulbCount {
		t.Fatalf("expected significantly more kelp than bulbs, got kelp=%d, bulbs=%d", kelpCount, bulbCount)
	}
	t.Logf("Groves spawned: kelp=%d, bubble bulbs=%d (ratio: %.1f%% bulbs)", kelpCount, bulbCount, float64(bulbCount)/float64(kelpCount+bulbCount)*100)
}

