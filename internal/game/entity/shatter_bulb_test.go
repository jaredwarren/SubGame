package entity

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockBulbRuntime struct {
	Runtime
	commands   []GameCommand
	playerPos  gvec.Vec2
	playerDims gvec.Vec2
}

func (m *mockBulbRuntime) PlayerPos() gvec.Vec2 {
	return m.playerPos
}

func (m *mockBulbRuntime) PlayerDims() gvec.Vec2 {
	return m.playerDims
}

func (m *mockBulbRuntime) HasActiveVehicle() bool {
	return false
}

func (m *mockBulbRuntime) Emit(cmd GameCommand) {
	m.commands = append(m.commands, cmd)
}

func TestShatterBulb_CreationAndSway(t *testing.T) {
	bulb := NewShatterBulb(100, 200)
	if !bulb.IsActive() {
		t.Fatalf("expected bulb to be active on creation")
	}
	initialSway := bulb.SwayPhase

	rt := &mockBulbRuntime{
		playerPos:  gvec.Vec2{X: 500, Y: 500},
		playerDims: gvec.Vec2{X: 16, Y: 16},
	}
	bulb.Update(rt)

	if bulb.SwayPhase <= initialSway {
		t.Fatalf("expected bulb SwayPhase to advance, was %f, got %f", initialSway, bulb.SwayPhase)
	}
}

func TestShatterBulb_DrawNoPanic(t *testing.T) {
	bulb := NewShatterBulb(100, 200)
	screen := ebiten.NewImage(320, 240)
	cam := &camera.Camera{}
	cam.Pos = gvec.Vec2{X: 50, Y: 150}

	// Test drawing at different times of day / sway phases
	bulb.Draw(screen, cam, 0.0)
	bulb.Draw(screen, cam, 50.0)
	bulb.Draw(screen, cam, 100.0)
}

func TestShatterBulb_PopWhenPlayerCollides(t *testing.T) {
	bulb := NewShatterBulb(100, 200)
	rt := &mockBulbRuntime{
		playerPos:  gvec.Vec2{X: 100, Y: 200},
		playerDims: gvec.Vec2{X: 24, Y: 24},
	}

	bulb.Update(rt)

	if bulb.IsActive() {
		t.Fatalf("expected bulb to become inactive upon colliding with player")
	}

	foundRestore := false
	for _, cmd := range rt.commands {
		if o2Cmd, ok := cmd.(RestoreOxygenCmd); ok {
			if o2Cmd.Amount > 0 {
				foundRestore = true
			}
		}
	}
	if !foundRestore {
		t.Fatalf("expected RestoreOxygenCmd to be emitted upon pop")
	}
}
