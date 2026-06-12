package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestThermoclineRammer_ChargeLimits(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.caveState)

	rammer := &entity.ThermoclineRammer{
		BaseEntity: entity.BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 36, Y: 24},
			Active:     true,
		},
		State:        1,
		Timer:        0,
		ChargeOrigin: gvec.Vec2{X: 100, Y: 100},
	}
	rammer.Vel = gvec.Vec2{X: 6.2, Y: 0}
	g.caveState.Entities = []entity.CaveEntity{rammer}

	// 1. Verify charge time limit (90 ticks)
	// We run update 90 times.
	for i := 0; i < 90; i++ {
		if rammer.State == 2 {
			break
		}
		err := g.Update()
		if err != nil {
			t.Fatal(err)
		}
	}

	if rammer.State != 2 {
		t.Errorf("expected ThermoclineRammer to transition to stun state (2) after 90 ticks, got state %d", rammer.State)
	}
	if rammer.StunTimer != 180 {
		t.Errorf("expected StunTimer to be 180, got %d", rammer.StunTimer)
	}

	// 2. Verify charge displacement limit (350.0 pixels)
	rammer.State = 1
	rammer.Timer = 0
	rammer.ChargeOrigin = gvec.Vec2{X: 100, Y: 100}
	rammer.Pos = gvec.Vec2{X: 460, Y: 100} // displacement of 360.0 (> 350.0)
	rammer.Vel = gvec.Vec2{X: 6.2, Y: 0}

	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	if rammer.State != 2 {
		t.Errorf("expected ThermoclineRammer to abort charge due to displacement (> 350.0) and transition to state 2, got %d", rammer.State)
	}
}

func TestCaveScene_EntityBoundaryClamping(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.caveState)

	// Set up a custom small grid: 10 x 10 tiles.
	// Grid is width 10, height 10.
	gridW := 10
	gridH := 10
	g.caveState.CaveGrid = make([][]bool, gridW)
	for i := 0; i < gridW; i++ {
		g.caveState.CaveGrid[i] = make([]bool, gridH)
	}

	rammer := &entity.ThermoclineRammer{
		BaseEntity: entity.BaseEntity{
			Pos:        gvec.Vec2{X: -100, Y: 1000},
			Dimensions: gvec.Vec2{X: 36, Y: 24},
			Active:     true,
		},
	}
	g.caveState.Entities = []entity.CaveEntity{rammer}

	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// maxX := 10 * 64 = 640. dims.X = 36 => maxPosX := 640 - 36 = 604
	// maxY := 10 * 64 = 640. dims.Y = 24 => maxPosY := 640 - 24 = 616
	expectedX := 0.0
	expectedY := float64(gridH*config.TileSize - 24)

	if rammer.Pos.X != expectedX {
		t.Errorf("expected clamped X to be %f, got %f", expectedX, rammer.Pos.X)
	}
	if rammer.Pos.Y != expectedY {
		t.Errorf("expected clamped Y to be %f, got %f", expectedY, rammer.Pos.Y)
	}

	// Test top boundary clamping: Y < -32.0 gets clamped to -32.0
	rammer.Pos = gvec.Vec2{X: 100, Y: -100}
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if rammer.Pos.Y != -32.0 {
		t.Errorf("expected clamped Y to be -32.0, got %f", rammer.Pos.Y)
	}
}
