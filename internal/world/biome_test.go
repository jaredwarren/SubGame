package world

import (
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

	// Verify offset produces finite numbers within reasonable bounds
	if offset.R < -50 || offset.R > 50 || offset.G < -50 || offset.G > 50 || offset.B < -50 || offset.B > 50 {
		t.Errorf("Smoothed water offset out of bounds: %+v", offset)
	}
}
