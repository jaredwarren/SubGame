package vehicle

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestApplyLateralDrag_KillsSidewaysVelocity(t *testing.T) {
	vel := gvec.Vec2{X: 4, Y: 4} // facing +X, so Y is fully lateral
	ApplyLateralDrag(&vel, 0, 0)
	if math.Abs(vel.Y) > 1e-9 {
		t.Errorf("expected lateral velocity killed, got %+v", vel)
	}
	if math.Abs(vel.X-4) > 1e-9 {
		t.Errorf("expected forward velocity kept, got %+v", vel)
	}
}

func TestTurnScaleForSpeed(t *testing.T) {
	if got := TurnScaleForSpeed(0, 6, 0.45); math.Abs(got-0.45) > 1e-9 {
		t.Errorf("idle scale = %v, want 0.45", got)
	}
	if got := TurnScaleForSpeed(6, 6, 0.45); math.Abs(got-1) > 1e-9 {
		t.Errorf("full speed scale = %v, want 1", got)
	}
}

type mockStickInput struct {
	vec     gvec.Vec2
	held    bool
	pressed map[ebiten.Key]bool
}

func (m mockStickInput) Cursor() gvec.Vec2                  { return gvec.Vec2{} }
func (m mockStickInput) IsKeyJustPressed(k ebiten.Key) bool { return false }
func (m mockStickInput) IsKeyPressed(k ebiten.Key) bool     { return m.pressed[k] }
func (m mockStickInput) StickAxes() (gvec.Vec2, bool)       { return m.vec, m.held }

func TestAnalogAimAxes_PointsInStickDirection(t *testing.T) {
	facing, throttle, ok := AnalogAimAxes(mockStickInput{vec: gvec.Vec2{X: 1, Y: 0}, held: true})
	if !ok {
		t.Fatal("expected analog stick to be held")
	}
	if math.Abs(facing) > 1e-9 {
		t.Errorf("expected facing 0 (right), got %v", facing)
	}
	if throttle <= 0.8 {
		t.Errorf("expected strong throttle, got %v", throttle)
	}
}

func TestSteerToward_ShortestPath(t *testing.T) {
	facing := 0.0
	SteerToward(&facing, math.Pi*0.5, 0.1)
	if facing <= 0 {
		t.Errorf("expected positive turn toward +Y, got %v", facing)
	}
}

type steerRuntime struct {
	input InputSource
}

func (s steerRuntime) TimeOfDay() float64                     { return 1000 }
func (s steerRuntime) IsActiveVehicle(v Vehicle) bool         { return true }
func (s steerRuntime) Input() InputSource                     { return s.input }
func (s steerRuntime) PlayerScreenCenter() gvec.Vec2          { return gvec.Vec2{} }
func (s steerRuntime) PlayerSlowed() bool                     { return false }
func (s steerRuntime) PlayerStunned() bool                    { return false }
func (s steerRuntime) IsOverworldSolidAt(tx, ty int) bool     { return false }
func (s steerRuntime) IsCaveSolidAt(tx, ty int) bool          { return false }
func (s steerRuntime) CanUseSonar() bool                      { return false }
func (s steerRuntime) BaseStationPos() (gvec.Vec2, gvec.Vec2) { return gvec.Vec2{}, gvec.Vec2{} }
func (s steerRuntime) Emit(cmd GameCommand)                   {}

func TestSkiff_TracksHeadingInsteadOfSliding(t *testing.T) {
	skiff := NewSkiff(0, 0)
	skiff.Facing = 0
	skiff.Vel = gvec.Vec2{X: 0, Y: 4}

	rt := steerRuntime{input: mockStickInput{pressed: map[ebiten.Key]bool{}}}
	for i := 0; i < 20; i++ {
		skiff.Update(rt)
	}
	if math.Abs(skiff.Vel.Y) >= 1.5 {
		t.Errorf("expected lateral slide to decay, still Vel.Y=%v", skiff.Vel.Y)
	}
}

func TestSkiff_StickAimsAndThrusts(t *testing.T) {
	skiff := NewSkiff(0, 0)
	skiff.Facing = math.Pi // facing left
	rt := steerRuntime{input: mockStickInput{vec: gvec.Vec2{X: 1, Y: 0}, held: true}}
	for i := 0; i < 40; i++ {
		skiff.Update(rt)
	}
	// Should have turned toward +X and built forward speed.
	if math.Abs(skiff.Facing) > 0.5 && math.Abs(skiff.Facing-2*math.Pi) > 0.5 {
		t.Errorf("expected skiff to face right toward stick, facing=%v", skiff.Facing)
	}
	if skiff.Vel.X <= 0.5 {
		t.Errorf("expected rightward thrust after aiming at stick, vel=%+v", skiff.Vel)
	}
}
