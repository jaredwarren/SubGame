package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestSkiffLightGameIntegration(t *testing.T) {
	g := NewGame()
	g.world = world.NewWorld(42)
	g.explorationTracker = exploration.NewTracker(g.world.Width, g.world.Height)
	g.TransitionToOverworld()

	// Mock input
	mockIn := NewMockInput()
	g.Input = mockIn

	// Place player and Skiff at tile (50, 50)
	skiffPos := gvec.Vec2{X: 50 * config.TileSize, Y: 50 * config.TileSize}
	skiff := vehicle.NewSkiff(skiffPos.X, skiffPos.Y)
	skiff.Facing = 0.0 // facing right (+X)
	g.OverworldVehicles = []vehicle.Vehicle{skiff}
	g.ActiveVehicle = skiff
	g.player.Pos = skiffPos

	// Initially without upgrade, headlights unavailable
	if g.hasFlashlightAvailable() {
		t.Fatal("expected flashlight/headlight to not be available without upgrade")
	}
	if g.IsFlashlightOn() {
		t.Fatal("expected headlights off initially")
	}

	// Install SkiffLight module into skiff upgrades
	skiff.Upgrades.AddItem(&item.SkiffLight{}, 1)

	if !g.hasFlashlightAvailable() {
		t.Fatal("expected headlights available after equipping SkiffLight")
	}
	if g.IsFlashlightOn() {
		t.Fatal("expected headlights to start OFF")
	}

	// 1. Toggle ON using KeyL
	mockIn.JustPressedKeys[ebiten.KeyL] = true
	if err := g.Update(); err != nil {
		t.Fatalf("Game.Update failed: %v", err)
	}
	mockIn.JustPressedKeys[ebiten.KeyL] = false

	if !g.IsFlashlightOn() {
		t.Fatal("expected headlights to turn ON after pressing KeyL")
	}

	// 2. Toggle OFF using HUD button click
	minX, minY, maxX, maxY := HUDVehicleLightButtonRect()
	mockIn.CursorPos = gvec.Vec2{X: (minX + maxX) / 2.0, Y: (minY + maxY) / 2.0}
	mockIn.JustPressedMouse[ebiten.MouseButtonLeft] = true
	if err := g.Update(); err != nil {
		t.Fatalf("Game.Update failed: %v", err)
	}
	mockIn.JustPressedMouse[ebiten.MouseButtonLeft] = false

	if g.IsFlashlightOn() {
		t.Fatal("expected headlights to turn OFF after clicking HUD light button")
	}

	// 3. Toggle back ON using KeyT
	mockIn.JustPressedKeys[ebiten.KeyT] = true
	if err := g.Update(); err != nil {
		t.Fatalf("Game.Update failed: %v", err)
	}
	mockIn.JustPressedKeys[ebiten.KeyT] = false

	if !g.IsFlashlightOn() {
		t.Fatal("expected headlights to turn ON after pressing KeyT")
	}

	// 4. Test battery drain & auto-shutoff at night
	g.TimeOfDay = 12000 // Nighttime, no solar recharge
	skiff.Battery = 0.03
	// Tick update: drain 0.02
	if err := g.Update(); err != nil {
		t.Fatalf("Game.Update failed: %v", err)
	}
	if !g.IsFlashlightOn() {
		t.Fatal("expected headlights still on with 0.01 battery remaining")
	}
	// Next tick update: battery drops to 0 -> auto shutoff
	if err := g.Update(); err != nil {
		t.Fatalf("Game.Update failed: %v", err)
	}
	if g.IsFlashlightOn() {
		t.Fatal("expected headlights to automatically shut off when battery is exhausted")
	}
	if skiff.Battery != 0 {
		t.Fatalf("expected battery to be 0, got %f", skiff.Battery)
	}

	// 5. Test forward exploration reveal boost when headlights are on
	// Re-charge battery and turn headlights ON
	skiff.Battery = 50.0
	skiff.ToggleHeadlights()
	if !skiff.IsHeadlightsOn() {
		t.Fatal("expected headlights to toggle on with recharged battery")
	}

	// Skiff is at tile (50, 50) facing 0 (east).
	// Without headlights, exploration radius is RevealRadius (4-5 tiles).
	// Forward reveal projects +3.5 tiles forward with radius RevealRadius+1,
	// reaching tile 50 + 3.5 + 5 = tile 58-59!
	farForwardTileX := 57
	farForwardTileY := 50
	if err := g.Update(); err != nil {
		t.Fatalf("Game.Update failed: %v", err)
	}

	if !g.explorationTracker.IsExplored(farForwardTileX, farForwardTileY) {
		t.Fatalf("expected forward tile (%d, %d) to be revealed by boosted headlights", farForwardTileX, farForwardTileY)
	}
}
