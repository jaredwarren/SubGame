package scene

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestSyncMapImageIncremental(t *testing.T) {
	w := world.NewWorld(12345)
	tracker := exploration.NewTracker(w.Width, w.Height)
	menu := NewBaseMenuScene()

	// Initial reveal at spawn
	tracker.Reveal(50, 50, 5)

	// First sync builds the full map
	menu.syncMapImage(w, tracker)
	if menu.mapImage == nil || len(menu.mapPixels) != w.Width*w.Height*4 {
		t.Fatal("map image or pixels not initialized after first sync")
	}

	spawnIdx := 50*w.Width + 50
	spawnOff := spawnIdx * 4
	if menu.mapPixels[spawnOff+0] == exploration.FogColor[0] &&
		menu.mapPixels[spawnOff+1] == exploration.FogColor[1] &&
		menu.mapPixels[spawnOff+2] == exploration.FogColor[2] {
		t.Fatal("spawn tile should not be fog color after initial sync")
	}

	farIdx := 150*w.Width + 150
	farOff := farIdx * 4
	if menu.mapPixels[farOff+0] != exploration.FogColor[0] ||
		menu.mapPixels[farOff+1] != exploration.FogColor[1] ||
		menu.mapPixels[farOff+2] != exploration.FogColor[2] {
		t.Fatal("unexplored tile at (150,150) should be fog color")
	}

	// Player drives skiff to (150, 150)
	tracker.Reveal(150, 150, 5)

	// Second sync incrementally updates map without reloading save
	menu.syncMapImage(w, tracker)

	if menu.mapPixels[farOff+0] == exploration.FogColor[0] &&
		menu.mapPixels[farOff+1] == exploration.FogColor[1] &&
		menu.mapPixels[farOff+2] == exploration.FogColor[2] {
		t.Fatal("tile at (150,150) should be cleared of fog after second syncMapImage call")
	}
}

func TestSyncMapImageOverflowRebuild(t *testing.T) {
	w := world.NewWorld(12345)
	tracker := exploration.NewTracker(w.Width, w.Height)
	menu := NewBaseMenuScene()

	menu.syncMapImage(w, tracker)

	// Reveal massive area to overflow the tracker backlog (>8192 tiles)
	tracker.Reveal(250, 250, 60)
	if !tracker.Overflowed() {
		t.Fatal("expected tracker to overflow with large reveal")
	}

	// syncMapImage should detect overflow and rebuild entire map
	menu.syncMapImage(w, tracker)

	if tracker.Overflowed() {
		t.Fatal("tracker overflow flag should be cleared after syncMapImage")
	}

	centerIdx := 250*w.Width + 250
	centerOff := centerIdx * 4
	if menu.mapPixels[centerOff+0] == exploration.FogColor[0] &&
		menu.mapPixels[centerOff+1] == exploration.FogColor[1] &&
		menu.mapPixels[centerOff+2] == exploration.FogColor[2] {
		t.Fatal("tile at (250,250) should be charted after overflow rebuild")
	}
}
