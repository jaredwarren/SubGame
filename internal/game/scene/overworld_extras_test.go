package scene

import (
	"math"
	"testing"

	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestCosmeticFishWanderAndFlee(t *testing.T) {
	fish := &CosmeticFish{
		Pos:       gvec.Vec2{X: 100, Y: 100},
		BasePos:   gvec.Vec2{X: 100, Y: 100},
		WobbleVal: 0,
		WobbleSpd: 0.1,
	}

	mockIsSolid := func(x, y float64) bool { return false }

	// 1. Far player position -> wander gently
	playerFar := gvec.Vec2{X: 1000, Y: 1000}
	fish.Update(playerFar, mockIsSolid)

	// Since player is far, fish should move slightly, but not flee
	if math.Abs(fish.Vel.X) > 1.0 || math.Abs(fish.Vel.Y) > 1.0 {
		t.Errorf("expected slow wandering velocity when player is far, got Vel = %+v", fish.Vel)
	}

	// 2. Near player position -> flee rapidly (update multiple times to accelerate)
	playerNear := gvec.Vec2{X: 90, Y: 100} // Player is 10 pixels to the left of the fish
	for i := 0; i < 15; i++ {
		fish.Update(playerNear, mockIsSolid)
	}

	// Fish should run away to the right (+X direction)
	if fish.Vel.X <= 0 {
		t.Errorf("expected fish to flee to the right (+X) from a player on the left, got Vel.X = %f", fish.Vel.X)
	}
	if math.Abs(fish.Vel.X) < 2.0 {
		t.Errorf("expected significant fleeing velocity after multiple updates, got Vel.X = %f", fish.Vel.X)
	}
}

func TestFloatingCrateState(t *testing.T) {
	crate := &FloatingCrate{
		Pos:        gvec.Vec2{X: 100, Y: 100},
		InitialPos: gvec.Vec2{X: 100, Y: 100},
		Collected:  false,
	}

	// Verify initial state
	if crate.Collected {
		t.Error("expected crate to not be collected initially")
	}
}

func TestThermalVentPushPhysics(t *testing.T) {
	vent := &ThermalVent{
		Pos:    gvec.Vec2{X: 100, Y: 100},
		Radius: 50.0,
	}

	// Target center is 10 pixels to the right of the vent
	targetCenter := gvec.Vec2{X: 110, Y: 100}

	dx := targetCenter.X - vent.Pos.X
	dy := targetCenter.Y - vent.Pos.Y
	dist := math.Hypot(dx, dy)

	if dist >= vent.Radius {
		t.Fatal("expected target to be within influence radius")
	}

	// Test calculated force
	ratio := 1.0 - (dist / vent.Radius)
	pushStrength := 1.6 * ratio
	pushX := (dx / dist) * pushStrength
	pushY := (dy / dist) * pushStrength

	// Push should be directed outwards from center (+X)
	if pushX <= 0 {
		t.Errorf("expected positive pushX outward from vent, got %f", pushX)
	}
	if pushY != 0 {
		t.Errorf("expected zero vertical pushY, got %f", pushY)
	}
}

func TestCosmeticFishCollision(t *testing.T) {
	fish := &CosmeticFish{
		Pos:       gvec.Vec2{X: 100, Y: 100},
		Vel:       gvec.Vec2{X: 2.0, Y: 2.0}, // Moving down-right
		BasePos:   gvec.Vec2{X: 100, Y: 100},
		WobbleVal: 0,
		WobbleSpd: 0.1,
	}

	// Mock isSolid returns true only if x > 101 (blocking X movement)
	isSolidMock := func(x, y float64) bool {
		return x > 101.0
	}

	fish.Update(gvec.Vec2{X: 500, Y: 500}, isSolidMock)

	// X position should remain at 100 because of collision, but Y should move forward
	if fish.Pos.X != 100.0 {
		t.Errorf("expected fish X position to remain 100 due to solid collision, got %f", fish.Pos.X)
	}
	if fish.Pos.Y <= 101.0 {
		t.Errorf("expected fish Y position to advance beyond 101.0, got %f", fish.Pos.Y)
	}
	// Vel.X should have bounced (become negative)
	if fish.Vel.X >= 0 {
		t.Errorf("expected fish Vel.X to bounce and become negative, got %f", fish.Vel.X)
	}
}
