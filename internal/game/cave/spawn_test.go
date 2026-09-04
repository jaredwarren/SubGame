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
	// 7x7 solid ring with open center at (3,3) so wall-anchored fauna can attach.
	grid := make([][]bool, 7)
	for x := range grid {
		grid[x] = make([]bool, 7)
		for y := range grid[x] {
			grid[x][y] = true
		}
	}
	grid[3][3] = false

	for id := FaunaID(0); id < FaunaCount; id++ {
		ent := SpawnFauna(id, 3, 3, grid, r)
		if ent == nil {
			t.Fatalf("SpawnFauna(%d) returned nil", id)
		}
		if !ent.IsActive() {
			t.Fatalf("SpawnFauna(%d) inactive", id)
		}
	}
	if SpawnFauna(FaunaCount, 0, 0, grid, r) != nil {
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

func TestSpawnFloraAnchored_ShatterBulb(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	leftEnt := SpawnFloraAnchored(FloraShatterBulb, 4, 5, 44, "left", r)
	leftBulb, ok := leftEnt.(*entity.ShatterBulb)
	if !ok {
		t.Fatalf("expected *entity.ShatterBulb, got %T", leftEnt)
	}
	if leftBulb.AnchorSide != "left" {
		t.Fatalf("expected left anchor, got %s", leftBulb.AnchorSide)
	}
	if leftBulb.Pos.X != 4*64 {
		t.Fatalf("expected left bulb Pos.X to be %d, got %f", 4*64, leftBulb.Pos.X)
	}

	rightEnt := SpawnFloraAnchored(FloraShatterBulb, 4, 5, 44, "right", r)
	rightBulb, ok := rightEnt.(*entity.ShatterBulb)
	if !ok {
		t.Fatalf("expected *entity.ShatterBulb, got %T", rightEnt)
	}
	if rightBulb.AnchorSide != "right" {
		t.Fatalf("expected right anchor, got %s", rightBulb.AnchorSide)
	}
	expectedRightX := float64(4*64 + 64) - entity.ShatterBulbArchetype.WallWidth
	if rightBulb.Pos.X != expectedRightX {
		t.Fatalf("expected right bulb Pos.X to be %f, got %f", expectedRightX, rightBulb.Pos.X)
	}

	floorEnt := SpawnFloraAnchored(FloraShatterBulb, 4, 5, 44, "floor", r)
	floorBulb, ok := floorEnt.(*entity.ShatterBulb)
	if !ok {
		t.Fatalf("expected *entity.ShatterBulb, got %T", floorEnt)
	}
	if floorBulb.AnchorSide != "floor" {
		t.Fatalf("expected floor anchor, got %s", floorBulb.AnchorSide)
	}
}

func TestDefaultBiomeMineralTypes(t *testing.T) {
	for _, s := range DefaultShallowReefBiome.MineralSpawns {
		switch s.Type {
		case resource.NodeTitanium, resource.NodeCopper, resource.NodeQuartz,
			resource.NodeNickel, resource.NodeTungsten, resource.NodeAbyssalOre:
			// ok
		default:
			t.Fatalf("unexpected mineral type %v", s.Type)
		}
	}
}

func TestBiomeFloorStyles(t *testing.T) {
	tests := []struct {
		name      string
		biome     *CaveBiomeSpec
		wantStyle FloorStyle
	}{
		{"ShallowReef", ShallowReefBiome, FloorStyleCoralSand},
		{"KelpForest", KelpForestBiome, FloorStyleMoss},
		{"ThermalBarrens", ThermalBarrensBiome, FloorStyleBasalt},
		{"AbyssalBlue", AbyssalBlueBiome, FloorStyleAbyssalSilt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.biome.FloorStyle != tt.wantStyle {
				t.Errorf("%s: got FloorStyle %v, want %v", tt.name, tt.biome.FloorStyle, tt.wantStyle)
			}
		})
	}
}

func TestShallowSeabedCaveFloorRendering(t *testing.T) {
	grid := [][]bool{
		{true, false, true},
		{true, true, true},
	}

	biomes := []*CaveBiomeSpec{
		ShallowReefBiome,
		KelpForestBiome,
		ThermalBarrensBiome,
		AbyssalBlueBiome,
		nil, // tests default fallback
	}

	for _, b := range biomes {
		cave := NewShallowSeabedCaveWithBiome(grid, b)
		if len(cave.tileImages) != 8 {
			t.Fatalf("expected 8 tile images, got %d", len(cave.tileImages))
		}
		for i, img := range cave.tileImages {
			if img == nil {
				t.Fatalf("tileImage[%d] is nil for biome %+v", i, b)
			}
		}
	}
}

func TestShallowReefTitaniumSpawns(t *testing.T) {
	// 1. Verify Titanium has high weight in ShallowReefBiome
	var titaniumWeight float64
	for _, m := range ShallowReefBiome.MineralSpawns {
		if m.Type == resource.NodeTitanium {
			titaniumWeight = m.Weight
		}
	}
	if titaniumWeight < 70 {
		t.Fatalf("expected titanium weight in ShallowReefBiome to be >= 70, got %f", titaniumWeight)
	}

	// 2. Verify at least 1 titanium node is guaranteed in shallow seabed cave across multiple seeds
	r := rand.New(rand.NewSource(42))
	for seed := int64(1); seed <= 50; seed++ {
		grid := GenerateShallowSeabedGrid(r, 5.0, true, true)
		c := NewShallowSeabedCaveWithBiome(grid, ShallowReefBiome)
		resources := c.GenerateResources(seed)
		titaniumCount := 0
		for _, res := range resources {
			if rn, ok := res.(*resource.ResourceNode); ok && rn.Type == resource.NodeTitanium {
				titaniumCount++
			}
		}
		if titaniumCount < 1 {
			t.Fatalf("expected at least 1 titanium node for seed %d, got %d", seed, titaniumCount)
		}
	}
}

func TestShallowCaveInkSquidSpawns(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	for seed := int64(1); seed <= 30; seed++ {
		grid := GenerateShallowSeabedGrid(r, 5.0, true, true)
		c := NewShallowSeabedCaveWithBiome(grid, ShallowReefBiome)
		entities := c.GenerateEntities(seed)
		squidCount := 0
		for _, ent := range entities {
			if _, ok := ent.(*entity.InkSquid); ok {
				squidCount++
			}
		}
		if squidCount < 1 {
			t.Fatalf("expected at least 1 ink squid for seed %d, got %d", seed, squidCount)
		}
	}
}


