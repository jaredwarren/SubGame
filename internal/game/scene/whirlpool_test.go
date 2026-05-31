package scene

import (
	"math"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestWhirlpoolBasic(t *testing.T) {
	w := world.NewWorld(12345)
	wp := NewWhirlpool(w.Seed)

	basePos := gvec.Vec2{X: 50.0 * config.TileSize, Y: 50.0 * config.TileSize}
	wp.Relocate(w, basePos)

	// Verify it spawned far from base
	dist := math.Hypot(wp.Pos.X-basePos.X, wp.Pos.Y-basePos.Y)
	if dist < 960.0 {
		t.Errorf("expected whirlpool to spawn far from base station, got distance %f", dist)
	}

	// Verify initially fading in
	if wp.State != WpFadeIn {
		t.Errorf("expected initial state WpFadeIn, got %v", wp.State)
	}
	if wp.Alpha != 0.0 {
		t.Errorf("expected initial Alpha to be 0.0, got %f", wp.Alpha)
	}

	// Simulate one update tick
	wp.Update(w, basePos)
	if wp.Alpha <= 0.0 {
		t.Error("expected Alpha to increase after Update tick")
	}

	// Test pull force when active
	wp.State = WpActive
	wp.Alpha = 1.0

	// Center is at wp.Pos. Let's test a target at distance 100 pixels (wp.Radius is 220)
	target := gvec.Vec2{X: wp.Pos.X + 100, Y: wp.Pos.Y}
	force := wp.PullForce(target)

	// Since target is to the right (+X):
	// - Radial force should be towards the center (-X)
	// - Tangential force (clockwise) should be perpendicular to radial
	if force.X >= 0 {
		t.Errorf("expected radial force to pull towards center (-X), got force.X = %f", force.X)
	}
	if force.Y == 0 {
		t.Errorf("expected non-zero tangential force, got force.Y = 0")
	}

	// Test target outside radius
	farTarget := gvec.Vec2{X: wp.Pos.X + 300, Y: wp.Pos.Y}
	noForce := wp.PullForce(farTarget)
	if noForce.X != 0 || noForce.Y != 0 {
		t.Errorf("expected zero force outside radius, got %+v", noForce)
	}
}
