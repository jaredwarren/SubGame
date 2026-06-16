package entity

import (
	"math"
	"testing"

	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockRuntime struct {
	Runtime
	commands      []GameCommand
	playerPos     gvec.Vec2
	playerDims    gvec.Vec2
	playerVel     gvec.Vec2
	playerFacing  float64
	sprinting     bool
	hasVehicle    bool
	vehiclePos    gvec.Vec2
	vehicleDims   gvec.Vec2
	vehicleFacing float64
	vehicleMoving bool
	flashlightOn  bool
	sonarActive   bool
	decoyPos      gvec.Vec2
	decoyFound    bool
	occluded      bool
	slowing       bool
	solid         bool
	isShockKelp   bool
}

func (m *mockRuntime) PlayerPos() gvec.Vec2 { return m.playerPos }
func (m *mockRuntime) PlayerDims() gvec.Vec2 {
	if m.playerDims.X == 0 && m.playerDims.Y == 0 {
		return gvec.Vec2{X: 20, Y: 20}
	}
	return m.playerDims
}
func (m *mockRuntime) PlayerVel() gvec.Vec2      { return m.playerVel }
func (m *mockRuntime) PlayerFacing() float64     { return m.playerFacing }
func (m *mockRuntime) IsPlayerSprinting() bool   { return m.sprinting }
func (m *mockRuntime) HasActiveVehicle() bool    { return m.hasVehicle }
func (m *mockRuntime) ActiveVehiclePos() gvec.Vec2 { return m.vehiclePos }
func (m *mockRuntime) ActiveVehicleDims() gvec.Vec2 { return m.vehicleDims }
func (m *mockRuntime) ActiveVehicleFacing() float64 { return m.vehicleFacing }
func (m *mockRuntime) ActiveVehicleMoving() bool { return m.vehicleMoving }
func (m *mockRuntime) FlashlightOn() bool { return m.flashlightOn }
func (m *mockRuntime) SonarActive() bool { return m.sonarActive }
func (m *mockRuntime) SoundWaveTimer() int { return 0 }
func (m *mockRuntime) SoundWaveX() float64 { return 0 }
func (m *mockRuntime) SoundWaveY() float64 { return 0 }
func (m *mockRuntime) SoundWaveRadius() float64 { return 0 }
func (m *mockRuntime) TimeOfDay() float64 { return 0 }
func (m *mockRuntime) IsShockKelpCave() bool { return m.isShockKelp }
func (m *mockRuntime) IsSolid(x, y, w, h float64) bool { return m.solid }
func (m *mockRuntime) FindClosestDecoy(pos gvec.Vec2, maxDist float64) (gvec.Vec2, bool) {
	return m.decoyPos, m.decoyFound
}
func (m *mockRuntime) CheckDeterrentOcclusion(pos1, pos2 gvec.Vec2) bool {
	return m.occluded
}
func (m *mockRuntime) CheckDeterrentSlowing(x, y, w, h float64) bool {
	return m.slowing
}
func (m *mockRuntime) Emit(cmd GameCommand) {
	m.commands = append(m.commands, cmd)
}

func TestSonicDecoy_Update(t *testing.T) {
	decoy := NewSonicDecoy(100, 100, gvec.Vec2{X: 10, Y: 0})
	mr := &mockRuntime{}

	// Update once
	decoy.Update(mr)

	if !decoy.IsActive() {
		t.Error("Decoy should be active initially")
	}
	if decoy.LifeTimer != 359 {
		t.Errorf("Expected LifeTimer to be 359, got %d", decoy.LifeTimer)
	}

	// Drag calculation: 10 * 0.95 = 9.5. New pos X should be 100 + 9.5 = 109.5
	expectedX := 109.5
	if math.Abs(decoy.Pos.X-expectedX) > 0.001 {
		t.Errorf("Expected X pos %f, got %f", expectedX, decoy.Pos.X)
	}

	// Update 59 more times (total 60 times). LifeTimer becomes 300, triggering sound wave.
	for i := 0; i < 59; i++ {
		decoy.Update(mr)
	}

	if decoy.LifeTimer != 300 {
		t.Errorf("Expected LifeTimer to be 300, got %d", decoy.LifeTimer)
	}

	if len(mr.commands) != 1 {
		t.Fatalf("Expected exactly 1 command emitted, got %d", len(mr.commands))
	}

	cmd, ok := mr.commands[0].(TriggerSoundWaveCmd)
	if !ok {
		t.Fatalf("Expected TriggerSoundWaveCmd, got %T", mr.commands[0])
	}

	expectedCenterX := decoy.Pos.X + decoy.Dimensions.X/2.0
	expectedCenterY := decoy.Pos.Y + decoy.Dimensions.Y/2.0
	if math.Abs(cmd.Pos.X-expectedCenterX) > 0.001 || math.Abs(cmd.Pos.Y-expectedCenterY) > 0.001 {
		t.Errorf("Sound wave center position incorrect: expected %+v, got %+v", gvec.Vec2{X: expectedCenterX, Y: expectedCenterY}, cmd.Pos)
	}
}

func TestDeterrentCloud_Update(t *testing.T) {
	cloud := NewDeterrentCloud(150, 150)
	mr := &mockRuntime{}

	cloud.Update(mr)

	if !cloud.IsActive() {
		t.Error("Cloud should be active initially")
	}
	if cloud.LifeTimer != 359 {
		t.Errorf("Expected LifeTimer to be 359, got %d", cloud.LifeTimer)
	}

	// Advance to end of life (359 updates remaining)
	for i := 0; i < 359; i++ {
		cloud.Update(mr)
	}

	if cloud.IsActive() {
		t.Error("Cloud should be deactivated after 360 updates")
	}
}

func TestFalseBulbSnare_DecoyInteraction(t *testing.T) {
	snare := &FalseBulbSnare{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 16, Y: 16},
			Active:     true,
		},
	}
	mr := &mockRuntime{
		playerPos:  gvec.Vec2{X: 500, Y: 500},
		decoyPos:   gvec.Vec2{X: 110, Y: 110},
		decoyFound: true,
	}

	// First update: should alert/aggro on the decoy (State becomes 1)
	snare.Update(mr)
	if snare.State != 1 {
		t.Errorf("Expected snare State to be 1 (aggro), got %d", snare.State)
	}

	// Move snare to overlap the decoy center
	snare.Pos = gvec.Vec2{X: 102, Y: 102} // close enough to decoy center (110, 110)
	snare.Update(mr)

	// Since they overlap, it should emit DestroyDecoyCmd and deactivate
	if snare.Active {
		t.Error("Expected snare to be inactive after hitting decoy")
	}

	foundDestroyCmd := false
	for _, cmd := range mr.commands {
		if _, ok := cmd.(DestroyDecoyCmd); ok {
			foundDestroyCmd = true
			break
		}
	}
	if !foundDestroyCmd {
		t.Errorf("Expected DestroyDecoyCmd, got %v", mr.commands)
	}
}

func TestFalseBulbSnare_DeterrentCloudOcclusion(t *testing.T) {
	snare := &FalseBulbSnare{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 16, Y: 16},
			Active:     true,
		},
		State: 1, // Already aggro
	}
	mr := &mockRuntime{
		playerPos: gvec.Vec2{X: 120, Y: 120},
		occluded:  true, // Occluded by cloud
	}

	snare.Update(mr)

	// Cloud should break aggro
	if snare.State != 0 {
		t.Errorf("Expected cloud to break aggro (State 0), got %d", snare.State)
	}
}

func TestThermoclineRammer_DecoyInteraction(t *testing.T) {
	rammer := &ThermoclineRammer{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 36, Y: 24},
			Active:     true,
		},
		State: 0, // patrol
	}
	mr := &mockRuntime{
		playerPos:  gvec.Vec2{X: 500, Y: 500},
		decoyPos:   gvec.Vec2{X: 120, Y: 110},
		decoyFound: true,
	}

	// First update should alert on decoy, state becomes 1 (charging)
	rammer.Update(mr)
	if rammer.State != 1 {
		t.Errorf("Expected rammer to start charging decoy (State 1), got %d", rammer.State)
	}

	// Hitting decoy: set position to overlap decoy center (120, 110)
	rammer.Pos = gvec.Vec2{X: 110, Y: 105} // overlap with 120, 110
	rammer.Update(mr)

	// Hitting decoy should emit DestroyDecoyCmd and enter State 2 (stun)
	if rammer.State != 2 {
		t.Errorf("Expected rammer to be stunned (State 2) after hitting decoy, got %d", rammer.State)
	}

	foundDestroyCmd := false
	for _, cmd := range mr.commands {
		if _, ok := cmd.(DestroyDecoyCmd); ok {
			foundDestroyCmd = true
			break
		}
	}
	if !foundDestroyCmd {
		t.Errorf("Expected DestroyDecoyCmd, got %v", mr.commands)
	}
}

func TestThermoclineRammer_DeterrentCloudSlowing(t *testing.T) {
	rammer := &ThermoclineRammer{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 36, Y: 24},
			Active:     true,
		},
		State: 1, // charging
	}
	rammer.Vel = gvec.Vec2{X: 6.2, Y: 0}
	mr := &mockRuntime{
		playerPos: gvec.Vec2{X: 200, Y: 100},
		slowing:   true, // cloud slowing active
	}

	prevPos := rammer.Pos
	rammer.Update(mr)

	// Speed should be scaled by 50%: 6.2 * 0.5 = 3.1.
	movedX := rammer.Pos.X - prevPos.X
	if math.Abs(movedX-3.1) > 0.001 {
		t.Errorf("Expected charge movement to be 3.1 (scaled by 50%%), got %f", movedX)
	}
}

func TestElectroWeaver_DecoyInteraction(t *testing.T) {
	weaver := &ElectroWeaver{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 16, Y: 16},
			Active:     true,
		},
		Timer: 299, // 1 tick away from striking
	}
	mr := &mockRuntime{
		playerPos:   gvec.Vec2{X: 500, Y: 500},
		decoyPos:    gvec.Vec2{X: 120, Y: 120},
		decoyFound:  true,
		isShockKelp: true, // to satisfy inAbyssal requirement (y/64 >= 80 or shock kelp cave)
	}

	weaver.Update(mr)

	// Since Timer reaches 300, it strikes the decoy, emits DestroyDecoyCmd, and teleports/resets timer
	if weaver.Timer != 0 {
		t.Errorf("Expected weaver timer to reset to 0, got %d", weaver.Timer)
	}

	foundDestroyCmd := false
	for _, cmd := range mr.commands {
		if _, ok := cmd.(DestroyDecoyCmd); ok {
			foundDestroyCmd = true
			break
		}
	}
	if !foundDestroyCmd {
		t.Errorf("Expected DestroyDecoyCmd to be emitted, got %v", mr.commands)
	}
}

func TestElectroWeaver_DeterrentCloudBreakLock(t *testing.T) {
	weaver := &ElectroWeaver{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 16, Y: 16},
			Active:     true,
		},
		Timer: 150, // tracking lock mid-way
	}
	mr := &mockRuntime{
		playerPos:    gvec.Vec2{X: 120, Y: 120},
		flashlightOn: false, // flashlight off
		occluded:     true,  // blocked by cloud
		isShockKelp:  true,
	}

	weaver.Update(mr)

	// Lock should immediately break (Timer resets to 0)
	if weaver.Timer != 0 {
		t.Errorf("Expected tracking lock to break (Timer 0), got %d", weaver.Timer)
	}
}
