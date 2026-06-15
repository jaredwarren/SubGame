package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockEntityRuntime struct {
	pPos       gvec.Vec2
	pDims      gvec.Vec2
	hasVehicle bool
	vPos       gvec.Vec2
	vDims      gvec.Vec2
	solidFunc  func(x, y, w, h float64) bool
	emitted    []entity.GameCommand
}

func (m *mockEntityRuntime) PlayerPos() gvec.Vec2            { return m.pPos }
func (m *mockEntityRuntime) PlayerDims() gvec.Vec2           { return m.pDims }
func (m *mockEntityRuntime) PlayerVel() gvec.Vec2            { return gvec.Vec2{} }
func (m *mockEntityRuntime) PlayerFacing() float64           { return 0.0 }
func (m *mockEntityRuntime) IsPlayerSprinting() bool         { return false }
func (m *mockEntityRuntime) HasActiveVehicle() bool          { return m.hasVehicle }
func (m *mockEntityRuntime) ActiveVehiclePos() gvec.Vec2     { return m.vPos }
func (m *mockEntityRuntime) ActiveVehicleDims() gvec.Vec2    { return m.vDims }
func (m *mockEntityRuntime) ActiveVehicleFacing() float64    { return 0.0 }
func (m *mockEntityRuntime) ActiveVehicleMoving() bool       { return false }
func (m *mockEntityRuntime) FlashlightOn() bool              { return true }
func (m *mockEntityRuntime) SoundWaveTimer() int             { return 0 }
func (m *mockEntityRuntime) SoundWaveX() float64             { return 0 }
func (m *mockEntityRuntime) SoundWaveY() float64             { return 0 }
func (m *mockEntityRuntime) SoundWaveRadius() float64         { return 0 }
func (m *mockEntityRuntime) SonarActive() bool               { return false }
func (m *mockEntityRuntime) TimeOfDay() float64              { return 0.0 }
func (m *mockEntityRuntime) IsShockKelpCave() bool           { return false }
func (m *mockEntityRuntime) IsSolid(x, y, w, h float64) bool {
	if m.solidFunc != nil {
		return m.solidFunc(x, y, w, h)
	}
	return false
}
func (m *mockEntityRuntime) Emit(cmd entity.GameCommand) {
	m.emitted = append(m.emitted, cmd)
}

func TestSandViper_Behavior(t *testing.T) {
	sv := entity.NewSandViper(100, 100)
	rt := &mockEntityRuntime{
		pPos:  gvec.Vec2{X: 500, Y: 500},
		pDims: gvec.Vec2{X: 20, Y: 20},
	}

	// 1. In patrol state, player is far away => should stay in patrol (State 0)
	sv.Update(rt)
	if sv.State != entity.StateViperPatrol {
		t.Errorf("expected StatePatrol (0), got %v", sv.State)
	}

	// 2. Bring player close => should transition to windup (State 1)
	rt.pPos = gvec.Vec2{X: 170, Y: 100}
	sv.Update(rt)
	if sv.State != entity.StateViperWindup {
		t.Errorf("expected StateWindup (1), got %v", sv.State)
	}

	// 3. Keep updating in windup for 30 ticks => should transition to lunge (State 2)
	for i := 0; i < 29; i++ {
		sv.Update(rt)
		if sv.State != entity.StateViperWindup {
			t.Fatalf("expected StateWindup (1) at tick %d, got %v", i, sv.State)
		}
	}
	// The 30th tick in windup will trigger the lunge
	sv.Update(rt)
	if sv.State != entity.StateViperLunge {
		t.Errorf("expected StateLunge (2) after 30 windup ticks, got %v", sv.State)
	}

	// 4. Update in lunge state. Player is close, but not overlapping.
	// SandViper should move toward target center.
	prevPos := sv.Pos
	sv.Update(rt)
	if sv.Pos == prevPos {
		t.Errorf("expected SandViper position to change during lunge, got unchanged %v", sv.Pos)
	}
	if sv.State != entity.StateViperLunge {
		t.Errorf("expected still in StateLunge (2), got %v", sv.State)
	}

	// 5. Overlap player => should emit damage & warning, transition to cooldown (State 3)
	// Put viper exactly at player position
	sv.Pos = rt.pPos
	sv.Update(rt)
	if sv.State != entity.StateViperCooldown {
		t.Errorf("expected StateCooldown (3) after hitting player, got %v", sv.State)
	}

	// Check that player damage was emitted
	hasDamage := false
	for _, cmd := range rt.emitted {
		if d, ok := cmd.(entity.DamagePlayerCmd); ok && d.Amount == 10.0 {
			hasDamage = true
			break
		}
	}
	if !hasDamage {
		t.Errorf("expected DamagePlayerCmd of 10.0 to be emitted, got %v", rt.emitted)
	}

	// 6. Update in cooldown for 120 ticks => should transition back to patrol (State 0)
	for i := 0; i < 119; i++ {
		sv.Update(rt)
		if sv.State != entity.StateViperCooldown {
			t.Fatalf("expected StateCooldown (3) at tick %d, got %v", i, sv.State)
		}
	}
	sv.Update(rt)
	if sv.State != entity.StateViperPatrol {
		t.Errorf("expected StatePatrol (0) after 120 cooldown ticks, got %v", sv.State)
	}
}
