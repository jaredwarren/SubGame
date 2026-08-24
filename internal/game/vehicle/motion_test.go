package vehicle

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestScaleForPower(t *testing.T) {
	force, speed := ScaleForPower(10, 5, 1, 0.5, true)
	if force != 10 || speed != 5 {
		t.Fatalf("powered: got force=%v speed=%v", force, speed)
	}
	force, speed = ScaleForPower(10, 5, 1, 0.5, false)
	if force != 1 || speed != 0.5 {
		t.Fatalf("unpowered: got force=%v speed=%v", force, speed)
	}
}

func TestApplyDragClamp(t *testing.T) {
	vel := gvec.Vec2{X: 10, Y: 0}
	ApplyDragClamp(&vel, 0.5, 3)
	if vel.X != 3 {
		t.Fatalf("expected clamp to 3, got %v", vel.X)
	}
}

func TestDrainBatteryOnMove(t *testing.T) {
	battery := 1.0
	if !DrainBatteryOnMove(&battery, true, true, 0.5) {
		t.Fatal("expected power remaining")
	}
	if battery != 0.5 {
		t.Fatalf("expected battery 0.5, got %v", battery)
	}
	if DrainBatteryOnMove(&battery, true, true, 1.0) {
		t.Fatal("expected no power after drain")
	}
	if battery != 0 {
		t.Fatalf("expected battery clamped to 0, got %v", battery)
	}
}

func TestApplyWaterline(t *testing.T) {
	opts := WaterlineOpts{Waterline: -8, SurfacePush: 0.15, BobSpring: 0.03}
	got := ApplyWaterline(-10, 0, 0, false, false, opts)
	if got != 0.15 {
		t.Fatalf("surface push: got %v want 0.15", got)
	}
	got = ApplyWaterline(20, 0, 0, true, false, opts)
	if got != 0 {
		t.Fatalf("below bob threshold with moving: got %v want 0", got)
	}
}
