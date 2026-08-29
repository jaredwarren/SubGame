package world

import (
	"math"
	"testing"
)

func TestWreckageDistanceBands(t *testing.T) {
	seeds := []int64{42, 100, 999, 12345, 888888}

	for _, seed := range seeds {
		w := NewWorld(seed)

		spawnX, spawnY := w.FindLifepodSpawn()

		type wreckInfo struct {
			x, y  int
			index int
			dist  float64
		}
		var wrecks []wreckInfo

		for x := 0; x < w.Width; x++ {
			for y := 0; y < w.Height; y++ {
				if w.OverworldMap[x][y] == TileWreckage {
					idx := w.ComputeWreckageShipIndex(x, y)
					dist := math.Hypot(float64(x-spawnX), float64(y-spawnY))
					wrecks = append(wrecks, wreckInfo{
						x:     x,
						y:     y,
						index: idx,
						dist:  dist,
					})
				}
			}
		}

		if len(wrecks) != 3 {
			t.Fatalf("seed %d: expected exactly 3 wrecks, got %d", seed, len(wrecks))
		}

		// Verify distinct indices 0, 1, 2
		seen := make(map[int]wreckInfo)
		for _, wr := range wrecks {
			if _, exists := seen[wr.index]; exists {
				t.Fatalf("seed %d: duplicate shipIndex %d", seed, wr.index)
			}
			seen[wr.index] = wr
		}

		// Verify strictly monotonic distance ordering: dist(0) < dist(1) < dist(2)
		w0 := seen[0]
		w1 := seen[1]
		w2 := seen[2]

		if !(w0.dist < w1.dist && w1.dist < w2.dist) {
			t.Errorf("seed %d: expected strictly increasing distances, got d0=%.1f, d1=%.1f, d2=%.1f",
				seed, w0.dist, w1.dist, w2.dist)
		}

		// Verify expected distance ranges
		if w0.dist < 15 || w0.dist > 55 {
			t.Errorf("seed %d: Ship 0 distance %.1f out of expected band [15, 55]", seed, w0.dist)
		}
		if w1.dist < 50 || w1.dist > 120 {
			t.Errorf("seed %d: Ship 1 distance %.1f out of expected band [50, 120]", seed, w1.dist)
		}
		if w2.dist < 100 {
			t.Errorf("seed %d: Ship 2 distance %.1f too close (< 100)", seed, w2.dist)
		}

		// Verify GetWreckageInfo
		for i := 0; i < 3; i++ {
			info := GetWreckageInfo(i)
			if info.Name == "" || info.EstDepth == "" || info.DivePrompt == "" {
				t.Errorf("missing wreckage metadata for shipIndex %d: %+v", i, info)
			}
		}
	}
}
