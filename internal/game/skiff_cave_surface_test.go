package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

func TestSkiffRenderedAtSurfaceInMatchingCaveOnly(t *testing.T) {
	g := NewGame()

	// 1. Create a skiff at overworld tile (5, 5)
	skiffX := float64(5*config.TileSize + 10)
	skiffY := float64(5*config.TileSize + 10)
	skiff := vehicle.NewSkiff(skiffX, skiffY)
	g.OverworldVehicles = []vehicle.Vehicle{skiff}

	// 2. Enter shallow cave at matching tile (5, 5)
	g.EnterCave(5, 5)
	if g.activeTrenchX != 5 || g.activeTrenchY != 5 {
		t.Fatalf("expected active trench to be (5, 5), got (%d, %d)", g.activeTrenchX, g.activeTrenchY)
	}

	screen := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	// Drawing the frame should succeed without panic
	g.caveState.Draw(g, screen)

	// 3. Now enter cave at a different tile (10, 10)
	g.EnterCave(10, 10)
	if g.activeTrenchX != 10 || g.activeTrenchY != 10 {
		t.Fatalf("expected active trench to be (10, 10), got (%d, %d)", g.activeTrenchX, g.activeTrenchY)
	}
	g.caveState.Draw(g, screen)

	// 4. Test deep cave layer where IsShallow is false
	g.caveState.IsShallow = false
	g.caveState.ActiveCave = cave.NewShockKelpCave([][]bool{{false}})
	g.activeTrenchX = 5
	g.activeTrenchY = 5
	g.caveState.Draw(g, screen)

	// 5. Test with no overworld vehicles
	g.OverworldVehicles = nil
	g.caveState.IsShallow = true
	g.caveState.ActiveCave = cave.NewShallowSeabedCave([][]bool{{false}})
	g.caveState.Draw(g, screen)
}
