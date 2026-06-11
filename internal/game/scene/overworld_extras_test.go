package scene

import (
	"math"
	"testing"

	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockGameContext struct {
	ticks          float64
	bubbles        []gvec.Vec2
	shakeDuration  int
	shakeIntensity float64
	player         *player.Player
	activeVehicle  vehicle.Vehicle
}

func (m *mockGameContext) GetTicks() float64 { return m.ticks }
func (m *mockGameContext) SpawnBubble(x, y float64) {
	m.bubbles = append(m.bubbles, gvec.Vec2{X: x, Y: y})
}
func (m *mockGameContext) TriggerScreenShake(dur int, intensity float64) {
	m.shakeDuration = dur
	m.shakeIntensity = intensity
}

func (m *mockGameContext) GetTargetCenter() gvec.Vec2 {
	if m.activeVehicle != nil {
		vPos := m.activeVehicle.GetPos()
		vDims := m.activeVehicle.GetDimensions()
		return gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
	}
	p := m.player
	return gvec.Vec2{X: p.Pos.X + p.Width/2.0, Y: p.Pos.Y + p.Height/2.0}
}

func (m *mockGameContext) GetTargetDimensions() gvec.Vec2 {
	if m.activeVehicle != nil {
		return m.activeVehicle.GetDimensions()
	}
	p := m.player
	return gvec.Vec2{X: p.Width, Y: p.Height}
}

func (m *mockGameContext) IsPiloting() bool {
	return m.activeVehicle != nil
}

func (m *mockGameContext) ApplyTargetForce(force gvec.Vec2) {
	if m.activeVehicle != nil {
		m.activeVehicle.ApplyForce(force)
	} else {
		m.player.Vel = m.player.Vel.Add(force)
	}
}

func (m *mockGameContext) DamageTarget(damage float64) {
	if m.activeVehicle != nil {
		m.activeVehicle.TakeDamage(damage)
	} else {
		m.player.CurrentHealth -= damage
	}
}

type mockCosmeticFishContext struct {
	targetCenter gvec.Vec2
	isSolidFunc  func(x, y float64) bool
}

func (m mockCosmeticFishContext) TargetCenter() gvec.Vec2 {
	return m.targetCenter
}

func (m mockCosmeticFishContext) IsSolid(x, y float64) bool {
	if m.isSolidFunc != nil {
		return m.isSolidFunc(x, y)
	}
	return false
}

func TestCosmeticFishWanderAndFlee(t *testing.T) {
	fish := &oe.CosmeticFish{
		Pos:       gvec.Vec2{X: 100, Y: 100},
		BasePos:   gvec.Vec2{X: 100, Y: 100},
		WobbleVal: 0,
		WobbleSpd: 0.1,
	}

	mockIsSolid := func(x, y float64) bool { return false }

	// 1. Far player position -> wander gently
	playerFar := gvec.Vec2{X: 1000, Y: 1000}
	ctxFar := mockCosmeticFishContext{targetCenter: playerFar, isSolidFunc: mockIsSolid}
	fish.Update(ctxFar)

	// Since player is far, fish should move slightly, but not flee
	if math.Abs(fish.Vel.X) > 1.0 || math.Abs(fish.Vel.Y) > 1.0 {
		t.Errorf("expected slow wandering velocity when player is far, got Vel = %+v", fish.Vel)
	}

	// 2. Near player position -> flee rapidly (update multiple times to accelerate)
	playerNear := gvec.Vec2{X: 90, Y: 100} // Player is 10 pixels to the left of the fish
	ctxNear := mockCosmeticFishContext{targetCenter: playerNear, isSolidFunc: mockIsSolid}
	for i := 0; i < 15; i++ {
		fish.Update(ctxNear)
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
	crate := &oe.FloatingCrate{
		Pos:        gvec.Vec2{X: 100, Y: 100},
		InitialPos: gvec.Vec2{X: 100, Y: 100},
		Collected:  false,
	}

	// Verify initial state
	if crate.Collected {
		t.Error("expected crate to not be collected initially")
	}
}

func TestThermalVentEruptingPushPhysics(t *testing.T) {
	vent := &oe.ThermalVent{
		Pos:        gvec.Vec2{X: 100, Y: 100},
		Radius:     50.0,
		State:      oe.VentErupting,
		StateTimer: 100,
		Intensity:  1.0,
	}

	p := player.NewPlayer(100, 90) // Center is (110, 100), which is 10 pixels to the right of vent
	p.Vel = gvec.Vec2{X: 0, Y: 0}
	ctx := &mockGameContext{
		player: p,
	}

	// In Erupting state, should apply push force and damage
	initialHealth := p.CurrentHealth
	vent.Update(ctx)

	if p.Vel.X <= 0 {
		t.Errorf("expected player to be pushed right (+X) by erupting vent, got Vel.X = %f", p.Vel.X)
	}
	if p.CurrentHealth >= initialHealth {
		t.Errorf("expected player to take damage from erupting vent, got health %f", p.CurrentHealth)
	}
}

func TestThermalVentDormantAndWarning(t *testing.T) {
	vent := &oe.ThermalVent{
		Pos:        gvec.Vec2{X: 100, Y: 100},
		Radius:     50.0,
		State:      oe.VentDormant,
		StateTimer: 100,
	}

	p := player.NewPlayer(100, 90) // Center is (110, 100)
	p.Vel = gvec.Vec2{X: 0, Y: 0}
	ctx := &mockGameContext{
		player: p,
	}

	// 1. Dormant: no push force, no damage
	initialHealth := p.CurrentHealth
	vent.Update(ctx)

	if p.Vel.X != 0 || p.Vel.Y != 0 {
		t.Errorf("expected zero push force in dormant state, got Vel = %+v", p.Vel)
	}
	if p.CurrentHealth != initialHealth {
		t.Errorf("expected no damage in dormant state, got health %f", p.CurrentHealth)
	}

	// 2. Warning: screen rumble, but no push force or damage
	vent.State = oe.VentWarning
	vent.Update(ctx)

	if p.Vel.X != 0 || p.Vel.Y != 0 {
		t.Errorf("expected zero push force in warning state, got Vel = %+v", p.Vel)
	}
	if p.CurrentHealth != initialHealth {
		t.Errorf("expected no damage in warning state, got health %f", p.CurrentHealth)
	}
	if ctx.shakeDuration <= 0 {
		t.Error("expected minor screen rumble in warning state, got shake duration <= 0")
	}
}

func TestCosmeticFishCollision(t *testing.T) {
	fish := &oe.CosmeticFish{
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

	ctx := mockCosmeticFishContext{
		targetCenter: gvec.Vec2{X: 500, Y: 500},
		isSolidFunc:  isSolidMock,
	}
	fish.Update(ctx)

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
