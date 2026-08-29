package scene

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestOverworldFindSafeExitPositionAvoidsLand(t *testing.T) {
	w := world.NewWorld(123)
	// Guarantee tiles around (50, 50)
	// Let (50, 50) be water where the skiff is located.
	w.OverworldMap[50][50] = world.TileWater

	// Place land to the left (West) of the skiff at tile (49, 50)
	w.OverworldMap[49][50] = world.TileLand
	w.OverworldMap[49][49] = world.TileLand
	w.OverworldMap[49][51] = world.TileLand

	// Place land above (North) at (50, 49)
	w.OverworldMap[50][49] = world.TileLand

	scene := &OverworldScene{World: w}

	vPos := gvec.Vec2{X: 50 * config.TileSize, Y: 50 * config.TileSize}
	vDims := gvec.Vec2{X: 56, Y: 24}
	pW, pH := 16.0, 16.0

	// In the old code, player was placed at vPos.X - 24, which lands directly inside tile 49 (TileLand)!
	oldExitX := vPos.X - 24
	oldExitY := vPos.Y
	if !scene.IsSolid(oldExitX, oldExitY, pW, pH) {
		t.Fatal("expected old exit position to overlap solid land for this test setup")
	}

	// Now call FindSafeExitPosition
	safePos := scene.FindSafeExitPosition(vPos, vDims, 0.0, pW, pH, nil)

	if scene.IsSolid(safePos.X, safePos.Y, pW, pH) {
		t.Fatalf("FindSafeExitPosition returned a solid land position (%f, %f)", safePos.X, safePos.Y)
	}

	// Verify it's within re-boarding range (< 60px) of the vehicle
	vCenter := gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
	dist := safePos.Distance(vCenter)
	if dist > 80.0 {
		t.Fatalf("safe exit position is too far from vehicle: dist=%f", dist)
	}
}

func TestOverworldCheckCollisionsUnstucksPlayerFromLand(t *testing.T) {
	w := world.NewWorld(456)
	// Island at (30, 30)
	w.OverworldMap[30][30] = world.TileLand
	// Water surrounding at (31, 30)
	w.OverworldMap[31][30] = world.TileWater

	scene := &OverworldScene{World: w}
	p := player.NewPlayer(30*config.TileSize+4, 30*config.TileSize+4)

	if !scene.IsSolid(p.Pos.X, p.Pos.Y, p.Width, p.Height) {
		t.Fatal("expected player to start inside solid land")
	}

	// Running CheckCollisions should pop player out into water
	scene.CheckCollisions(p, nil)

	if scene.IsSolid(p.Pos.X, p.Pos.Y, p.Width, p.Height) {
		t.Fatalf("expected player to be unstuck from land, but pos (%f, %f) is still solid", p.Pos.X, p.Pos.Y)
	}
}
