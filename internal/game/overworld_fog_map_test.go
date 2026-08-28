package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestOverworldSkiffFogMapUpdateWithoutReload(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.world = world.NewWorld(12345)
	g.explorationTracker = exploration.NewTracker(g.world.Width, g.world.Height)

	// Deploy and board skiff at (50, 50)
	skiffPos := gvec.Vec2{X: 50 * config.TileSize, Y: 50 * config.TileSize}
	skiff := vehicle.NewSkiff(skiffPos.X, skiffPos.Y)
	g.OverworldVehicles = []vehicle.Vehicle{skiff}
	g.ActiveVehicle = skiff
	g.player.Pos = skiffPos

	// Initial map opening (e.g. at spawn)
	g.TransitionToMap()
	g.ClosePDA()

	// Verify (120, 120) is unexplored initially
	if g.explorationTracker.IsExplored(120, 120) {
		t.Fatal("expected tile (120,120) to be unexplored initially")
	}

	// Move skiff to (120, 120) and tick Game.Update multiple times (simulating driving)
	newPos := gvec.Vec2{X: 120 * config.TileSize, Y: 120 * config.TileSize}
	skiff.SetPos(newPos)
	g.player.Pos = newPos

	for i := 0; i < 5; i++ {
		if err := g.Update(); err != nil {
			t.Fatalf("Game.Update failed: %v", err)
		}
	}

	// Verify exploration tracker now has (120, 120) explored
	if !g.explorationTracker.IsExplored(120, 120) {
		t.Fatal("expected tile (120,120) to be explored after driving skiff")
	}

	// Prior to the fix, Game.Update() drained and discarded newly revealed tiles every tick,
	// leaving the backlog empty when the map screen opened.
	// Verify that the dirty backlog contains the newly charted tiles.
	dirty := g.explorationTracker.Drain()
	if len(dirty) == 0 {
		t.Fatal("expected dirty tiles in explorationTracker backlog after driving skiff and ticking Game.Update")
	}

	targetIdx := 120*g.world.Width + 120
	found := false
	for _, idx := range dirty {
		if idx == targetIdx {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected target tile index %d to be in dirty backlog", targetIdx)
	}
}

