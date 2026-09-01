package entity

import (
	"math"
	"testing"

	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestElectroWeaver_LineOfSight_BlockedBySolid(t *testing.T) {
	weaver := NewElectroWeaver(100, 100)
	weaver.Timer = 50 // existing charge

	// Mock runtime where IsSolid returns true (rock wall in between)
	mr := &mockRuntime{
		playerPos:    gvec.Vec2{X: 300, Y: 100},
		flashlightOn: true,
		isShockKelp:  true,
		solid:        true,
	}

	weaver.Update(mr)

	// Since line of sight is blocked by solid rock, Timer must decay and not increase
	if weaver.Timer >= 50 {
		t.Errorf("Expected timer to decay when LoS is blocked, got %d", weaver.Timer)
	}
	if weaver.State == WeaverStateTracking {
		t.Errorf("Weaver should not track without line of sight")
	}
}

func TestElectroWeaver_LineOfSight_Clear(t *testing.T) {
	weaver := NewElectroWeaver(100, 100)
	weaver.Timer = 50

	// Mock runtime with clear water (solid = false) and flashlight on
	mr := &mockRuntime{
		playerPos:    gvec.Vec2{X: 250, Y: 100},
		flashlightOn: true,
		isShockKelp:  true,
		solid:        false,
	}

	weaver.Update(mr)

	// Clear line of sight + electricity: Timer must increment
	if weaver.Timer != 51 {
		t.Errorf("Expected timer to increment to 51, got %d", weaver.Timer)
	}
	if weaver.State != WeaverStateTracking {
		t.Errorf("Expected weaver to enter WeaverStateTracking, got %v", weaver.State)
	}
}

func TestElectroWeaver_PhysicalLunge_CollisionDamagesPlayer(t *testing.T) {
	weaver := NewElectroWeaver(100, 100)
	weaver.Timer = 299 // 1 tick from 300 (full charge)

	// Player is close enough that the lunge strikes
	mr := &mockRuntime{
		playerPos:    gvec.Vec2{X: 120, Y: 100},
		playerDims:   gvec.Vec2{X: 20, Y: 20},
		flashlightOn: true,
		isShockKelp:  true,
		solid:        false,
	}

	// First tick reaches 300 -> launches lunge and connects
	weaver.Update(mr)

	// Should have dealt 45 damage
	foundDamage := false
	for _, cmd := range mr.commands {
		if dmg, ok := cmd.(DamagePlayerCmd); ok {
			if dmg.Amount == 45 && dmg.Kind == DamageElectric {
				foundDamage = true
			}
		}
	}
	if !foundDamage {
		t.Errorf("Expected DamagePlayerCmd with 45 electric damage, got: %v", mr.commands)
	}

	// Should enter cooldown, timer reset to 0
	if weaver.State != WeaverStateCooldown {
		t.Errorf("Expected weaver to enter WeaverStateCooldown, got %v", weaver.State)
	}
	if weaver.Timer != 0 {
		t.Errorf("Expected weaver timer to reset to 0, got %d", weaver.Timer)
	}

	// Crucially: position should NOT have teleported 350+ pixels away
	distMoved := math.Hypot(weaver.Pos.X-100, weaver.Pos.Y-100)
	if distMoved > 50.0 {
		t.Errorf("Weaver teleported! Distance moved was %f (expected physical proximity)", distMoved)
	}
}

func TestElectroWeaver_PhysicalLunge_MissWall(t *testing.T) {
	weaver := NewElectroWeaver(100, 100)
	weaver.State = WeaverStateLunge
	weaver.LungeDir = gvec.Vec2{X: 1, Y: 0}
	weaver.LungeTimer = 30

	// Next position is a solid wall
	mr := &mockRuntime{
		playerPos:   gvec.Vec2{X: 400, Y: 100},
		isShockKelp: true,
		solid:       true, // Wall collision!
	}

	weaver.Update(mr)

	// Should have reset tracking timer and transitioned to cooldown
	foundTrackingReset := false
	for _, cmd := range mr.commands {
		if track, ok := cmd.(UpdateWeaverTrackingTimerCmd); ok && track.Value == 0 {
			foundTrackingReset = true
		}
	}
	if !foundTrackingReset {
		t.Errorf("Expected UpdateWeaverTrackingTimerCmd with Value 0, got: %v", mr.commands)
	}

	if weaver.State != WeaverStateCooldown {
		t.Errorf("Expected weaver to transition to WeaverStateCooldown after wall collision, got %v", weaver.State)
	}
}

func TestElectroWeaver_AudioTelegraphAndStrikeSFX(t *testing.T) {
	weaver := NewElectroWeaver(100, 100)
	d := ElectroWeaverArchetype

	// 1. Telegraph crackle window (e.g. at StrikeTimerFrames - 120)
	weaver.Timer = d.StrikeTimerFrames - 121 // 1 frame before telegraph interval
	mr := &mockRuntime{
		playerPos:    gvec.Vec2{X: 250, Y: 100},
		flashlightOn: true,
		isShockKelp:  true,
		solid:        false,
	}

	weaver.Update(mr)

	foundChargeSFX := false
	for _, cmd := range mr.commands {
		if sfx, ok := cmd.(PlaySFXCmd); ok && sfx.Path == "sfx/weaver_charge.wav" {
			foundChargeSFX = true
		}
	}
	if !foundChargeSFX {
		t.Errorf("Expected PlaySFXCmd for weaver_charge.wav during telegraph window, got commands: %v", mr.commands)
	}

	// 2. Strike initiation SFX (at StrikeTimerFrames)
	weaver.Timer = d.StrikeTimerFrames - 1
	mr.commands = nil
	weaver.Update(mr)

	foundShockSFX := false
	for _, cmd := range mr.commands {
		if sfx, ok := cmd.(PlaySFXCmd); ok && sfx.Path == "sfx/weaver_shock.wav" {
			foundShockSFX = true
		}
	}
	if !foundShockSFX {
		t.Errorf("Expected PlaySFXCmd for weaver_shock.wav on lunge initiation, got commands: %v", mr.commands)
	}
}
