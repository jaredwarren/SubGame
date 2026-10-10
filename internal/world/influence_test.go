package world

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/pcghash"
)

func TestBuildInfluenceMapDeterministic(t *testing.T) {
	prev := pcghash.UseSpinGlass
	pcghash.UseSpinGlass = true
	defer func() { pcghash.UseSpinGlass = prev }()

	mk := func(seed int64) *World {
		w := &World{Width: 48, Height: 48, Seed: seed}
		w.OverworldMap = make([][]TileType, w.Width)
		for x := 0; x < w.Width; x++ {
			w.OverworldMap[x] = make([]TileType, w.Height)
			for y := 0; y < w.Height; y++ {
				// Small island in the corner, water elsewhere.
				if x < 8 && y < 8 {
					w.OverworldMap[x][y] = TileLand
				} else {
					w.OverworldMap[x][y] = TileWater
				}
			}
		}
		w.buildLandDistMap()
		w.buildWaterDistMap()
		w.BuildInfluenceMap()
		return w
	}

	a := mk(42)
	b := mk(42)
	if a.Influence == nil || b.Influence == nil {
		t.Fatal("expected Influence when UseSpinGlass is on")
	}
	if a.Influence.Width != b.Influence.Width || a.Influence.Height != b.Influence.Height {
		t.Fatal("size mismatch")
	}
	for x := 0; x < a.Influence.Width; x++ {
		for y := 0; y < a.Influence.Height; y++ {
			if a.Influence.Labels[x][y] != b.Influence.Labels[x][y] {
				t.Fatalf("label mismatch at %d,%d", x, y)
			}
			if a.Influence.Danger[x][y] != b.Influence.Danger[x][y] ||
				a.Influence.Vital[x][y] != b.Influence.Vital[x][y] {
				t.Fatalf("channel mismatch at %d,%d", x, y)
			}
		}
	}

	_, dangerNear, vitalNear, ok := a.InfluenceAt(9, 9)
	if !ok {
		t.Fatal("InfluenceAt failed")
	}
	_, dangerFar, vitalFar, _ := a.InfluenceAt(40, 40)
	if dangerFar+0.05 < dangerNear {
		t.Fatalf("expected far ocean danger >= near-shore; near=%v far=%v", dangerNear, dangerFar)
	}
	if vitalNear+0.05 < vitalFar {
		t.Fatalf("expected near-shore vitality >= far; near=%v far=%v", vitalNear, vitalFar)
	}
}

func TestBuildInfluenceMapNoopWhenFlagOff(t *testing.T) {
	prev := pcghash.UseSpinGlass
	pcghash.UseSpinGlass = false
	defer func() { pcghash.UseSpinGlass = prev }()

	w := &World{Width: 16, Height: 16, Seed: 1}
	w.OverworldMap = make([][]TileType, 16)
	for x := 0; x < 16; x++ {
		w.OverworldMap[x] = make([]TileType, 16)
	}
	w.buildLandDistMap()
	w.buildWaterDistMap()
	w.BuildInfluenceMap()
	if w.Influence != nil {
		t.Fatal("expected nil Influence when flag off")
	}
}
