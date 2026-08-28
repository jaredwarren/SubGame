package vehicle

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockSkiffRuntime struct {
	timeOfDay float64
	input     InputSource
}

func (m *mockSkiffRuntime) TimeOfDay() float64                                    { return m.timeOfDay }
func (m *mockSkiffRuntime) IsActiveVehicle(v Vehicle) bool                         { return true }
func (m *mockSkiffRuntime) Input() InputSource                                     { return m.input }
func (m *mockSkiffRuntime) PlayerScreenCenter() gvec.Vec2                          { return gvec.Vec2{} }
func (m *mockSkiffRuntime) PlayerSlowed() bool                                     { return false }
func (m *mockSkiffRuntime) PlayerStunned() bool                                    { return false }
func (m *mockSkiffRuntime) IsOverworldSolidAt(tx, ty int) bool                     { return false }
func (m *mockSkiffRuntime) IsCaveSolidAt(tx, ty int) bool                          { return false }
func (m *mockSkiffRuntime) CanUseSonar() bool                                      { return true }
func (m *mockSkiffRuntime) BaseStationPos() (gvec.Vec2, gvec.Vec2)                  { return gvec.Vec2{}, gvec.Vec2{} }
func (m *mockSkiffRuntime) Emit(cmd GameCommand)                                   {}

type mockSkiffInput struct {
	justPressed map[ebiten.Key]bool
	pressed     map[ebiten.Key]bool
}

func (mi *mockSkiffInput) Cursor() gvec.Vec2                  { return gvec.Vec2{} }
func (mi *mockSkiffInput) IsKeyPressed(k ebiten.Key) bool     { return mi.pressed[k] }
func (mi *mockSkiffInput) IsKeyJustPressed(k ebiten.Key) bool { return mi.justPressed[k] }

func TestSkiffHeadlightsUpgradeAndToggle(t *testing.T) {
	skiff := NewSkiff(100, 100)

	// Initially, without SkiffLight installed:
	if skiff.HasHeadlights() {
		t.Fatal("expected skiff not to have headlights initially")
	}
	if skiff.IsHeadlightsOn() {
		t.Fatal("expected headlights not to be on initially")
	}
	if skiff.ToggleHeadlights() {
		t.Fatal("toggle should fail without SkiffLight upgrade installed")
	}

	// Install SkiffLight module:
	ok := skiff.Upgrades.AddItem(&item.SkiffLight{}, 1)
	if !ok {
		t.Fatal("failed to install SkiffLight into upgrades")
	}

	if !skiff.HasHeadlights() {
		t.Fatal("expected skiff to report HasHeadlights() = true after upgrade installed")
	}
	if skiff.IsHeadlightsOn() {
		t.Fatal("headlights should remain off until toggled")
	}

	// Toggle on:
	if !skiff.ToggleHeadlights() {
		t.Fatal("expected headlights to toggle on")
	}
	if !skiff.IsHeadlightsOn() {
		t.Fatal("expected IsHeadlightsOn() to be true")
	}

	// Toggle off:
	if skiff.ToggleHeadlights() {
		t.Fatal("expected headlights to toggle off")
	}
	if skiff.IsHeadlightsOn() {
		t.Fatal("expected IsHeadlightsOn() to be false")
	}
}

func TestSkiffHeadlightsBatteryDrainAndAutoShutoff(t *testing.T) {
	skiff := NewSkiff(100, 100)
	skiff.Upgrades.AddItem(&item.SkiffLight{}, 1)

	// Set battery to small amount and nighttime (so no solar recharge)
	skiff.Battery = 0.05
	skiff.ToggleHeadlights()

	if !skiff.IsHeadlightsOn() {
		t.Fatal("expected headlights on")
	}

	runtime := &mockSkiffRuntime{
		timeOfDay: 12000, // Nighttime (no solar charging)
		input: &mockSkiffInput{
			justPressed: make(map[ebiten.Key]bool),
			pressed:     make(map[ebiten.Key]bool),
		},
	}

	// Update tick 1: drain should reduce battery by 0.02
	skiff.Update(runtime)
	if skiff.Battery > 0.035 || skiff.Battery < 0.025 {
		t.Fatalf("expected battery ~0.03 after 1 tick of drain, got %f", skiff.Battery)
	}
	if !skiff.IsHeadlightsOn() {
		t.Fatal("headlights should still be on with battery remaining")
	}

	// Update tick 2: drain again (down to ~0.01)
	skiff.Update(runtime)
	if !skiff.IsHeadlightsOn() {
		t.Fatal("headlights should still be on with battery remaining")
	}

	// Update tick 3: battery should deplete to 0 and headlights auto-shutoff
	skiff.Update(runtime)
	if skiff.Battery > 0 {
		// one more tick if needed
		skiff.Update(runtime)
	}

	if skiff.Battery != 0 {
		t.Fatalf("expected battery to clamp at 0, got %f", skiff.Battery)
	}
	if skiff.IsHeadlightsOn() {
		t.Fatal("expected headlights to automatically turn off when battery is depleted")
	}

	// Trying to toggle on with 0 battery should fail
	if skiff.ToggleHeadlights() {
		t.Fatal("should not be able to toggle headlights on with 0 battery")
	}
	if skiff.IsHeadlightsOn() {
		t.Fatal("headlights should remain off with 0 battery")
	}
}

func TestSkiffHeadlightsDirectToggle(t *testing.T) {
	skiff := NewSkiff(100, 100)
	skiff.Upgrades.AddItem(&item.SkiffLight{}, 1)

	if !skiff.ToggleHeadlights() {
		t.Fatal("expected ToggleHeadlights() to turn headlights on")
	}
	if !skiff.IsHeadlightsOn() {
		t.Fatal("expected IsHeadlightsOn() to be true")
	}

	if skiff.ToggleHeadlights() {
		t.Fatal("expected ToggleHeadlights() to turn headlights off")
	}
	if skiff.IsHeadlightsOn() {
		t.Fatal("expected IsHeadlightsOn() to be false")
	}
}
