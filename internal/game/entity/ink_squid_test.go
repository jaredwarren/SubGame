package entity

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockSquidRuntime struct {
	Runtime
	commands   []GameCommand
	playerPos  gvec.Vec2
	playerDims gvec.Vec2
	solidCheck bool
}

func (m *mockSquidRuntime) PlayerPos() gvec.Vec2 {
	return m.playerPos
}

func (m *mockSquidRuntime) PlayerDims() gvec.Vec2 {
	return m.playerDims
}

func (m *mockSquidRuntime) HasActiveVehicle() bool {
	return false
}

func (m *mockSquidRuntime) IsSolid(x, y, w, h float64) bool {
	return m.solidCheck
}

func (m *mockSquidRuntime) Emit(cmd GameCommand) {
	m.commands = append(m.commands, cmd)
}

func TestInkSquid_PassiveWandering(t *testing.T) {
	squid := NewInkSquid(100, 100, true)
	if !squid.IsActive() {
		t.Fatalf("expected squid to be active upon creation")
	}

	rt := &mockSquidRuntime{
		playerPos:  gvec.Vec2{X: 400, Y: 400}, // Far away
		playerDims: gvec.Vec2{X: 20, Y: 28},
	}

	// Update 10 ticks
	for i := 0; i < 10; i++ {
		squid.Update(rt)
	}

	if len(rt.commands) > 0 {
		t.Fatalf("expected no commands emitted when player is far away, got %d", len(rt.commands))
	}
	if squid.FleeTimer > 0 {
		t.Fatalf("expected squid not to flee when unprovoked")
	}
}

func TestInkSquid_ThreatReactionAndInk(t *testing.T) {
	squid := NewInkSquid(100, 100, true)
	rt := &mockSquidRuntime{
		playerPos:  gvec.Vec2{X: 130, Y: 100}, // Within 70px threat range (dist ~30px)
		playerDims: gvec.Vec2{X: 20, Y: 28},
	}

	squid.Update(rt)

	// Verify SpawnInkCloudCmd was emitted
	if len(rt.commands) != 1 {
		t.Fatalf("expected 1 command emitted, got %d", len(rt.commands))
	}
	inkCmd, ok := rt.commands[0].(SpawnInkCloudCmd)
	if !ok {
		t.Fatalf("expected SpawnInkCloudCmd, got %T", rt.commands[0])
	}
	if inkCmd.Pos.X <= 0 || inkCmd.Pos.Y <= 0 {
		t.Fatalf("expected valid ink cloud pos, got %+v", inkCmd.Pos)
	}

	// Verify flee state and escape vector away from player (player is to the right at X=130, squid should escape left)
	if squid.FleeTimer != squid.stats().FleeFrames-1 {
		t.Fatalf("expected FleeTimer to be %d, got %d", squid.stats().FleeFrames-1, squid.FleeTimer)
	}
	if squid.Vel.X >= 0 {
		t.Fatalf("expected squid to jet left away from player, got Vel.X = %f", squid.Vel.X)
	}
	if squid.InkCooldown != squid.stats().CooldownFrames {
		t.Fatalf("expected InkCooldown to be %d, got %d", squid.stats().CooldownFrames, squid.InkCooldown)
	}

	// Next tick: even if player is still right there, squid should NOT emit another ink cloud because of cooldown
	rt.commands = nil
	squid.Update(rt)
	if len(rt.commands) > 0 {
		t.Fatalf("expected no ink emitted while on cooldown, got %d commands", len(rt.commands))
	}
}

func TestInkCloud_LifeAndSlowPlayer(t *testing.T) {
	cloud := NewInkCloud(200, 200)
	if !cloud.IsActive() {
		t.Fatalf("expected cloud to be active")
	}

	// Player touching cloud
	rt := &mockSquidRuntime{
		playerPos:  gvec.Vec2{X: 200, Y: 200},
		playerDims: gvec.Vec2{X: 20, Y: 28},
	}

	cloud.Update(rt)

	// Verify SlowPlayerCmd emitted
	if len(rt.commands) != 1 {
		t.Fatalf("expected 1 command emitted, got %d", len(rt.commands))
	}
	slowCmd, ok := rt.commands[0].(SlowPlayerCmd)
	if !ok {
		t.Fatalf("expected SlowPlayerCmd, got %T", rt.commands[0])
	}
	if slowCmd.Duration != 180 || slowCmd.Factor != 0.5 {
		t.Fatalf("expected duration 180 and factor 0.5, got %+v", slowCmd)
	}

	// Fast-forward lifetime
	cloud.LifeTimer = 1
	cloud.Update(rt)
	if cloud.IsActive() {
		t.Fatalf("expected cloud to become inactive after lifetime expires")
	}
}

func TestInkSquidAndCloud_DrawNoPanic(t *testing.T) {
	squid := NewInkSquid(100, 100, true)
	cloud := NewInkCloud(100, 100)
	screen := ebiten.NewImage(320, 240)
	cam := &camera.Camera{Pos: gvec.Vec2{X: 50, Y: 50}}

	// Draw idle
	squid.Draw(screen, cam, 0.0)
	cloud.Draw(screen, cam, 0.0)

	// Draw fleeing
	squid.FleeTimer = 30
	squid.FacingRight = false
	squid.Draw(screen, cam, 10.0)
	cloud.LifeTimer = 30 // near fade
	cloud.Draw(screen, cam, 10.0)
}
