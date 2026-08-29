package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestExitSkiffNearLandAvoidsSoftLock(t *testing.T) {
	g := NewGame()
	g.world = world.NewWorld(999)
	g.TransitionToOverworld()

	mockIn := NewMockInput()
	g.Input = mockIn

	// Place land directly to the left (West) of tile (50, 50)
	g.world.OverworldMap[50][50] = world.TileWater
	g.world.OverworldMap[49][50] = world.TileLand
	g.world.OverworldMap[49][49] = world.TileLand
	g.world.OverworldMap[49][51] = world.TileLand

	// Deploy skiff at (50, 50)
	skiffPos := gvec.Vec2{X: 50 * config.TileSize, Y: 50 * config.TileSize}
	skiff := vehicle.NewSkiff(skiffPos.X, skiffPos.Y)
	g.OverworldVehicles = []vehicle.Vehicle{skiff}
	g.ActiveVehicle = skiff
	g.player.Pos = skiffPos

	// Player is piloting the skiff
	if g.ActiveVehicle == nil {
		t.Fatal("expected player to be piloting skiff")
	}

	// Press 'F' to exit the skiff
	mockIn.JustPressedKeys[ebiten.KeyF] = true
	if err := g.Update(); err != nil {
		t.Fatalf("Game.Update failed: %v", err)
	}
	mockIn.JustPressedKeys[ebiten.KeyF] = false

	if g.ActiveVehicle != nil {
		t.Fatal("expected player to have exited vehicle")
	}

	// Player must NOT be inside solid land!
	if g.overworldState.IsSolid(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height) {
		t.Fatalf("player exited inside solid land at (%f, %f)!", g.player.Pos.X, g.player.Pos.Y)
	}

	// Verify player can move freely (not soft-locked)
	initialPos := g.player.Pos
	mockIn.PressedKeys[ebiten.KeyD] = true // Move right (+X)
	for i := 0; i < 5; i++ {
		if err := g.Update(); err != nil {
			t.Fatalf("Game.Update failed: %v", err)
		}
	}
	mockIn.PressedKeys[ebiten.KeyD] = false

	if g.player.Pos.X <= initialPos.X {
		t.Fatalf("player is soft-locked: X pos did not change after pressing D (initial=%f, current=%f)",
			initialPos.X, g.player.Pos.X)
	}
}
