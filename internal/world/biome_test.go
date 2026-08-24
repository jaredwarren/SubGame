package world

import (
	"math"
	"testing"
)

func TestBiomeGeneration(t *testing.T) {
	w := NewWorld(12345)
	if w.BiomeMap == nil {
		t.Fatal("BiomeMap was not initialized")
	}
	if len(w.BiomeMap) != w.Width || len(w.BiomeMap[0]) != w.Height {
		t.Fatalf("BiomeMap dimensions mismatch: got %dx%d, want %dx%d", len(w.BiomeMap), len(w.BiomeMap[0]), w.Width, w.Height)
	}

	biomesSeen := make(map[BiomeID]int)
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			biomesSeen[w.BiomeMap[x][y]]++
		}
	}

	if len(biomesSeen) < 2 {
		t.Errorf("Expected at least 2 distinct biomes in generated world, got %d", len(biomesSeen))
	}
}

func TestSmoothedWaterOffset(t *testing.T) {
	w := NewWorld(54321)
	offset := w.GetSmoothedWaterOffset(50, 50)
	expected := w.computeSmoothedWaterOffset(50, 50)

	if offset != expected {
		t.Errorf("precomputed offset mismatch at (50,50): got %+v, want %+v", offset, expected)
	}

	// Verify offset produces finite numbers within reasonable bounds
	if offset.R < -100 || offset.R > 100 || offset.G < -100 || offset.G > 100 || offset.B < -100 || offset.B > 100 {
		t.Errorf("Smoothed water offset out of bounds: %+v", offset)
	}
}

func TestBiomeTransitionAdjustment(t *testing.T) {
	w := NewWorld(54321)

	origIntensity := BiomeTransitionIntensity
	origRadius := BiomeBlendRadius
	defer func() {
		BiomeTransitionIntensity = origIntensity
		BiomeBlendRadius = origRadius
	}()

	BiomeTransitionIntensity = 1.0
	BiomeBlendRadius = 2
	offset1 := w.computeSmoothedWaterOffset(50, 50)

	BiomeTransitionIntensity = 2.0
	offset2 := w.computeSmoothedWaterOffset(50, 50)

	// Offset at 2.0 should be double that of 1.0 (within float precision)
	const eps = 1e-6
	if offset1.R != 0 && (offset2.R-offset1.R*2.0 > eps || offset2.R-offset1.R*2.0 < -eps) {
		t.Errorf("Expected offset2.R ~= 2 * offset1.R, got offset1=%v, offset2=%v", offset1.R, offset2.R)
	}
}

func TestBiomeSpecialCaveSpawning(t *testing.T) {
	w := NewWorld(424242)

	thermoCount := 0
	shockKelpCount := 0
	trenchCount := 0
	wreckageCount := 0

	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			tt := w.OverworldMap[x][y]
			bID := w.BiomeMap[x][y]

			switch tt {
			case TileThermoCave:
				thermoCount++
				if bID != BiomeThermalBarrens {
					t.Errorf("TileThermoCave spawned in incorrect biome %q at (%d, %d), expected %q", bID, x, y, BiomeThermalBarrens)
				}
			case TileShockKelpCave:
				shockKelpCount++
				if bID != BiomeKelpForest {
					t.Errorf("TileShockKelpCave spawned in incorrect biome %q at (%d, %d), expected %q", bID, x, y, BiomeKelpForest)
				}
			case TileTrench:
				trenchCount++
				if bID != BiomeAbyssalBlue {
					t.Errorf("TileTrench spawned in incorrect biome %q at (%d, %d), expected %q", bID, x, y, BiomeAbyssalBlue)
				}
			case TileWreckage:
				wreckageCount++
			}
		}
	}

	thermoSpec := GetBiomeInfo(BiomeThermalBarrens)
	if thermoCount == 0 || thermoCount > thermoSpec.SpecialCaveMaxCount {
		t.Errorf("Expected 1..%d TileThermoCave, got %d", thermoSpec.SpecialCaveMaxCount, thermoCount)
	}

	kelpSpec := GetBiomeInfo(BiomeKelpForest)
	if shockKelpCount == 0 || shockKelpCount > kelpSpec.SpecialCaveMaxCount {
		t.Errorf("Expected 1..%d TileShockKelpCave, got %d", kelpSpec.SpecialCaveMaxCount, shockKelpCount)
	}

	abyssalSpec := GetBiomeInfo(BiomeAbyssalBlue)
	if trenchCount == 0 || trenchCount > abyssalSpec.SpecialCaveMaxCount {
		t.Errorf("Expected 1..%d TileTrench, got %d", abyssalSpec.SpecialCaveMaxCount, trenchCount)
	}

	if wreckageCount != 3 {
		t.Errorf("Expected 3 TileWreckage, got %d", wreckageCount)
	}
}

func TestThermoCaveIsDiveable(t *testing.T) {
	info := GetTileInfo(TileThermoCave)
	if info == nil {
		t.Fatal("missing TileThermoCave info")
	}
	if !info.IsDiveable {
		t.Error("expected TileThermoCave to be diveable")
	}
	if info.DivePrompt == "" {
		t.Error("expected TileThermoCave to have a dive prompt")
	}
}

func TestFindLifepodSpawn(t *testing.T) {
	seeds := []int64{12345, 98765, 424242, 55555}

	for _, seed := range seeds {
		w := NewWorld(seed)
		tx, ty := w.FindLifepodSpawn()

		if tx < 0 || tx >= w.Width || ty < 0 || ty >= w.Height {
			t.Fatalf("seed %d: invalid spawn coords (%d, %d)", seed, tx, ty)
		}

		if w.OverworldMap[tx][ty] != TileWater {
			t.Errorf("seed %d: expected TileWater at (%d, %d), got %v", seed, tx, ty, w.OverworldMap[tx][ty])
		}

		if w.BiomeMap[tx][ty] != BiomeShallowReef {
			t.Errorf("seed %d: expected BiomeShallowReef at (%d, %d), got %v", seed, tx, ty, w.BiomeMap[tx][ty])
		}

		// Ensure it is reasonably close to center (e.g. within 150 tiles of (250, 250))
		centerX, centerY := float64(w.Width)/2.0, float64(w.Height)/2.0
		dist := hypot(float64(tx)-centerX, float64(ty)-centerY)
		if dist > 150.0 {
			t.Errorf("seed %d: spawn (%d, %d) too far from center: dist=%.1f", seed, tx, ty, dist)
		}
	}
}

func hypot(a, b float64) float64 {
	return math.Sqrt(a*a + b*b)
}
