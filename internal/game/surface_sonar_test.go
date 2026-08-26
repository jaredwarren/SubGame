package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/data"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// mockVehicleRuntime implements vehicle.Runtime for testing vehicle behavior in isolation.
type mockVehicleRuntime struct {
	timeOfDay    float64
	active       bool
	input        *mockInputSource
	slowed       bool
	stunned      bool
	canSonar     bool
	cmds         []vehicle.GameCommand
}

type mockInputSource struct {
	justPressed map[ebiten.Key]bool
	pressed     map[ebiten.Key]bool
}

func (m *mockInputSource) Cursor() gvec.Vec2               { return gvec.Vec2{} }
func (m *mockInputSource) IsKeyJustPressed(k ebiten.Key) bool { return m.justPressed[k] }
func (m *mockInputSource) IsKeyPressed(k ebiten.Key) bool     { return m.pressed[k] }

func (m *mockVehicleRuntime) TimeOfDay() float64                  { return m.timeOfDay }
func (m *mockVehicleRuntime) IsActiveVehicle(v vehicle.Vehicle) bool { return m.active }
func (m *mockVehicleRuntime) Input() vehicle.InputSource          { return m.input }
func (m *mockVehicleRuntime) PlayerScreenCenter() gvec.Vec2       { return gvec.Vec2{} }
func (m *mockVehicleRuntime) PlayerSlowed() bool                  { return m.slowed }
func (m *mockVehicleRuntime) PlayerStunned() bool                 { return m.stunned }
func (m *mockVehicleRuntime) IsOverworldSolidAt(tx, ty int) bool  { return false }
func (m *mockVehicleRuntime) IsCaveSolidAt(tx, ty int) bool       { return false }
func (m *mockVehicleRuntime) CanUseSonar() bool                   { return m.canSonar }
func (m *mockVehicleRuntime) BaseStationPos() (gvec.Vec2, gvec.Vec2) {
	return gvec.Vec2{}, gvec.Vec2{}
}
func (m *mockVehicleRuntime) Emit(cmd vehicle.GameCommand) {
	m.cmds = append(m.cmds, cmd)
}

func TestSurfaceSonar_ItemDefinitionAndRecipe(t *testing.T) {
	// 1. Verify item ID lookup
	sonarItem := &item.SurfaceSonar{}
	if sonarItem.GetName() != "Surface Sonar Module" {
		t.Errorf("expected name 'Surface Sonar Module', got '%s'", sonarItem.GetName())
	}
	id, ok := item.ItemIDFromName("Surface Sonar Module")
	if !ok || id != item.IDSurfaceSonar {
		t.Errorf("ItemIDFromName failed, got %v, ok=%v", id, ok)
	}

	// 2. Verify recipe existence and ingredients
	var found *data.Recipe
	for _, r := range data.CraftingRecipes {
		if r.NewResult().GetName() == "Surface Sonar Module" {
			found = &r
			break
		}
	}
	if found == nil {
		t.Fatal("Surface Sonar Module recipe not found in CraftingRecipes")
	}
	if found.Tier != 1 {
		t.Errorf("expected tier 1 recipe, got %d", found.Tier)
	}
	if found.Unlocked {
		t.Error("expected recipe to be locked initially")
	}
	// Ingredients: 6 Titanium, 4 Copper, 3 Quartz
	ingCount := map[string]int{}
	for _, ing := range found.Ingredients {
		ingCount[ing.NewItem().GetName()] = ing.Quantity
	}
	if ingCount["Titanium"] != 6 || ingCount["Copper"] != 4 || ingCount["Quartz"] != 3 {
		t.Errorf("unexpected recipe ingredients: %+v", ingCount)
	}
}

func TestSkiff_SurfaceSonarActivation(t *testing.T) {
	skiff := vehicle.NewSkiff(100, 100)
	rt := &mockVehicleRuntime{
		active: true,
		input: &mockInputSource{
			justPressed: map[ebiten.Key]bool{ebiten.KeyQ: true},
			pressed:     map[ebiten.Key]bool{},
		},
	}

	// 1. Without Surface Sonar installed, pressing [Q] does not activate sonar
	skiff.Update(rt)
	if len(rt.cmds) != 0 {
		t.Fatalf("expected 0 commands without module, got %d", len(rt.cmds))
	}
	if skiff.GetBattery() != skiff.GetMaxBattery() {
		t.Errorf("battery should not be drained without module")
	}

	// 2. Install Surface Sonar
	skiff.GetUpgrades().AddItem(&item.SurfaceSonar{}, 1)
	origBattery := skiff.GetBattery()

	skiff.Update(rt)
	if len(rt.cmds) != 1 {
		t.Fatalf("expected 1 command with module, got %d", len(rt.cmds))
	}
	cmd, ok := rt.cmds[0].(vehicle.ActivateSurfaceSonarCmd)
	if !ok {
		t.Fatalf("expected ActivateSurfaceSonarCmd, got %T", rt.cmds[0])
	}
	if cmd.FogRevealRadius != 18 {
		t.Errorf("expected FogRevealRadius=18, got %d", cmd.FogRevealRadius)
	}
	if cmd.POIDetectionRadius != 35 {
		t.Errorf("expected POIDetectionRadius=35, got %d", cmd.POIDetectionRadius)
	}
	expectedDrain := vehicle.SkiffArchetype.SurfaceSonar.BatteryCost // 25.0
	if skiff.GetBattery() > origBattery-expectedDrain+0.1 {
		t.Errorf("battery not drained properly: got %.1f, want %.1f", skiff.GetBattery(), origBattery-expectedDrain)
	}

	// 3. Immediately pressing [Q] again should not trigger due to cooldown
	rt.cmds = nil
	skiff.Update(rt)
	if len(rt.cmds) != 0 {
		t.Errorf("expected cooldown to prevent back-to-back activation")
	}
}

func TestGame_DrainSurfaceSonarCommand(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld

	// Set up world and exploration tracker
	g.world = world.NewWorld(12345)
	g.explorationTracker = exploration.NewTracker(g.world.Width, g.world.Height)

	// Place some POIs: one kelp cave at (50, 50) and one thermo cave at (60, 50)
	g.world.OverworldMap[50][50] = world.TileShockKelpCave
	g.world.OverworldMap[60][50] = world.TileThermoCave

	// Origin at tile (40, 50)
	originX := float64(40 * config.TileSize)
	originY := float64(50 * config.TileSize)

	// Distance to (50, 50) is 10 tiles (within 18-tile fog reveal and 35-tile POI detection)
	// Distance to (60, 50) is 20 tiles (outside 18-tile fog reveal, but within 35-tile POI detection!)

	rt := newVehicleRuntimeAdapter(g)
	rt.cmds = []vehicle.GameCommand{
		vehicle.ActivateSurfaceSonarCmd{
			Source: gvec.Vec2{X: originX, Y: originY},
			Pulse: vehicle.SonarPulse{
				DurationTicks: 120,
				RadiusStep:    6.5,
			},
			FogRevealRadius:    18,
			POIDetectionRadius: 35,
		},
	}

	g.drainVehicleCommands(rt)

	// 1. Fog cleared at (50, 50) because distance is 10 <= 18
	if !g.explorationTracker.IsExplored(50, 50) {
		t.Errorf("expected tile (50, 50) to be explored by fog reveal radius")
	}

	// 2. Fog NOT cleared at (60, 50) because distance is 20 > 18
	if g.explorationTracker.IsExplored(60, 50) {
		t.Errorf("expected tile (60, 50) to NOT be explored by fog reveal radius")
	}

	// 3. Both POIs discovered by extended 35-tile detection radius
	if !g.explorationTracker.IsPOIDiscovered(50, 50) {
		t.Errorf("expected kelp cave at (50, 50) to be marked discovered")
	}
	if !g.explorationTracker.IsPOIDiscovered(60, 50) {
		t.Errorf("expected thermo cave at (60, 50) to be marked discovered through fog")
	}

	// 4. Notification warning was shown
	if g.MineWarning.Message == "" {
		t.Error("expected warning banner message to be displayed")
	}
}

func TestExplorationTracker_POIDiscoverySerialization(t *testing.T) {
	tracker := exploration.NewTracker(100, 100)
	tracker.MarkPOIDiscovered(25, 30)
	tracker.MarkPOIDiscovered(80, 45)

	if !tracker.IsPOIDiscovered(25, 30) || !tracker.IsPOIDiscovered(80, 45) {
		t.Fatal("expected POIs to be discovered")
	}
	if tracker.IsPOIDiscovered(10, 10) {
		t.Error("unmarked POI should not be discovered")
	}

	saved := tracker.SerializeState()
	restored := exploration.NewTracker(100, 100)
	restored.DeserializeState(saved)

	if !restored.IsPOIDiscovered(25, 30) {
		t.Errorf("restored tracker missing discovered POI (25, 30)")
	}
	if !restored.IsPOIDiscovered(80, 45) {
		t.Errorf("restored tracker missing discovered POI (80, 45)")
	}
	if restored.IsPOIDiscovered(10, 10) {
		t.Error("unmarked POI should not be discovered in restored tracker")
	}
}

func TestVehicleUpgrade_ClickTransfersToVehicle(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld

	skiff := vehicle.NewSkiff(100, 100)
	g.OverworldVehicles = append(g.OverworldVehicles, skiff)
	g.ActiveVehicle = skiff

	// Put a Surface Sonar module into player inventory
	g.player.Inventory.Clear()
	g.player.Hotbar.Clear()
	sonar := &item.SurfaceSonar{}
	g.player.Inventory.AddItem(sonar, 1)

	// Call TransferToVehicle
	g.TransferToVehicle(sonar)

	// Should now be in skiff's Upgrades, not in player inventory or hotbar
	if g.player.Inventory.Count(sonar) != 0 {
		t.Errorf("expected sonar removed from player inventory, got %d", g.player.Inventory.Count(sonar))
	}
	if g.player.Hotbar.Count(sonar) != 0 {
		t.Errorf("expected sonar not in player hotbar, got %d", g.player.Hotbar.Count(sonar))
	}
	if skiff.GetUpgrades().Count(sonar) != 1 {
		t.Errorf("expected sonar in skiff Upgrades, got %d", skiff.GetUpgrades().Count(sonar))
	}

	// Now put another upgrade in hotbar and test transferring from hotbar to vehicle
	amp := &item.SonarAmplifier{}
	g.player.Hotbar.AddItem(amp, 1)
	g.TransferToVehicle(amp)

	if g.player.Hotbar.Count(amp) != 0 {
		t.Errorf("expected amplifier removed from player hotbar, got %d", g.player.Hotbar.Count(amp))
	}
	if skiff.GetUpgrades().Count(amp) != 1 {
		t.Errorf("expected amplifier in skiff Upgrades, got %d", skiff.GetUpgrades().Count(amp))
	}
}
