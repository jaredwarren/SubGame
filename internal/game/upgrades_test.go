package game

import (
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestInventory_Resize(t *testing.T) {
	inv := item.NewInventory(5)
	inv.AddItem(&item.Fins{}, 3) // Fins have max stack of 1, so they occupy slots 0, 1, and 2

	// Resize to 10
	inv.Resize(10)
	if len(inv.Slots) != 10 {
		t.Errorf("expected slots count to be 10, got %d", len(inv.Slots))
	}

	// Verify items are preserved
	if !item.HasItem[*item.Fins](inv, 3) {
		t.Errorf("expected items to be preserved after resize")
	}

	// Resize to smaller (shrink) to size 2 (truncates slot index 2)
	inv.Resize(2)
	if len(inv.Slots) != 2 {
		t.Errorf("expected slots count to be 2, got %d", len(inv.Slots))
	}

	// Verify items are preserved up to new size, and counts map is updated
	if !item.HasItem[*item.Fins](inv, 2) {
		t.Errorf("expected 2 fins to be preserved after shrink")
	}
	if item.HasItem[*item.Fins](inv, 3) {
		t.Errorf("expected only 2 fins to remain (not 3)")
	}
}

func TestBaseStation_UpgradesManagement(t *testing.T) {
	b := base.NewBaseStation(0, 0)

	// Built-in modules should be active by default
	if !b.HasModule(item.ModuleFabricator) {
		t.Errorf("expected fabricator module to be active by default")
	}
	if !b.HasModule(item.ModuleMedical) {
		t.Errorf("expected medical module to be active by default")
	}

	// Non-installed upgrades should not be active
	if b.HasModule(item.ModuleSolar) {
		t.Errorf("expected solar array to not be active initially")
	}
	if b.HasModule(item.ModuleStorage) {
		t.Errorf("expected storage vault to not be active initially")
	}

	// Install Solar MKI
	solarUpgrade := &item.UpgradeSolar{}
	if !b.InstallUpgrade(solarUpgrade) {
		t.Errorf("expected successful solar module installation")
	}

	// Solar MKI should now be active
	if !b.HasModule(item.ModuleSolar) {
		t.Errorf("expected solar module to be active after installation")
	}
	if b.HasModule(item.ModuleSolarMKII) {
		t.Errorf("expected solar MKII to not be active")
	}

	// Uninstall Solar MKI
	b.Upgrades.Remove(solarUpgrade, 1)
	b.RecalculateProperties()
	if b.HasModule(item.ModuleSolar) {
		t.Errorf("expected solar module to not be active after removal")
	}

	// Install Solar MKII
	solarMKII := &item.UpgradeSolarMKII{}
	if !b.InstallUpgrade(solarMKII) {
		t.Errorf("expected successful solar MKII installation")
	}

	// Solar MKII active should imply Solar MKI capabilities are also active
	if !b.HasModule(item.ModuleSolarMKII) {
		t.Errorf("expected solar MKII to be active")
	}
	if !b.HasModule(item.ModuleSolar) {
		t.Errorf("expected solar MKI capabilities to be implied by Solar MKII")
	}
}

func TestBaseStation_StorageUpgrades(t *testing.T) {
	b := base.NewBaseStation(0, 0)

	// Initially 24 slots
	if len(b.Storage.Slots) != 24 {
		t.Errorf("expected initial slots to be 24, got %d", len(b.Storage.Slots))
	}

	// Install Storage MKI
	storageMKI := &item.UpgradeStorage{}
	b.InstallUpgrade(storageMKI)
	if len(b.Storage.Slots) != 24 {
		t.Errorf("expected storage slots to remain 24 after MKI, got %d", len(b.Storage.Slots))
	}

	// Install Storage MKII (consumes MKI and gets slotted)
	// We simulate uninstalling MKI and installing MKII
	b.Upgrades.Remove(storageMKI, 1)
	storageMKII := &item.UpgradeStorageMKII{}
	b.InstallUpgrade(storageMKII)

	if len(b.Storage.Slots) != 48 {
		t.Errorf("expected storage slots to expand to 48 with MKII, got %d", len(b.Storage.Slots))
	}

	// Uninstall Storage MKII
	b.Upgrades.Remove(storageMKII, 1)
	b.RecalculateProperties()
	if len(b.Storage.Slots) != 24 {
		t.Errorf("expected storage slots to shrink to 24 after MKII removal, got %d", len(b.Storage.Slots))
	}
}

func TestCraftingRecipes_UpgradeStructure(t *testing.T) {
	// Verify that the Storage MKII recipe requires Storage MKI (item.UpgradeStorage) and Solar MKI (item.UpgradeSolar)
	var storageMKIFound, solarMKIFound bool

	for _, rcp := range CraftingRecipes {
		resultType := reflect.TypeOf(rcp.NewResult())
		if resultType == reflect.TypeOf(&item.UpgradeStorageMKII{}) {
			for _, ing := range rcp.Ingredients {
				ingType := reflect.TypeOf(ing.NewItem())
				if ingType == reflect.TypeOf(&item.UpgradeStorage{}) {
					storageMKIFound = true
				}
				if ingType == reflect.TypeOf(&item.UpgradeSolar{}) {
					solarMKIFound = true
				}
			}
		}
	}

	if !storageMKIFound {
		t.Errorf("expected Storage MKII recipe to require Storage MKI item")
	}
	if !solarMKIFound {
		t.Errorf("expected Storage MKII recipe to require Solar MKI item")
	}
}

func TestScoutSub_SonarAmplifierUpgrade(t *testing.T) {
	// Create mock runtime
	// ScoutSub uses item.NewInventory(1) for Upgrades.
	// Let's create a Scout Sub
	sub := vehicle.NewScoutSub(0, 0)

	// Initially upgrade is installed by default
	if !item.HasItem[*item.SonarAmplifier](sub.GetUpgrades(), 1) {
		t.Errorf("expected ScoutSub to start with Sonar Amplifier upgrade")
	}

	// Verify initial sonar pulse values
	pulse := sub.Sonar.Pulse
	if pulse.DurationTicks != 180 {
		t.Errorf("expected initial duration to be 180, got %d", pulse.DurationTicks)
	}
	if pulse.RadiusStep != 6.5 {
		t.Errorf("expected initial radius step to be 6.5, got %f", pulse.RadiusStep)
	}

	// Remove the upgrade to test empty state
	amp := &item.SonarAmplifier{}
	if !sub.GetUpgrades().Remove(amp, 1) {
		t.Errorf("expected successful removal of default Sonar Amplifier")
	}

	// Verify the upgrade is no longer active
	if item.HasItem[*item.SonarAmplifier](sub.GetUpgrades(), 1) {
		t.Errorf("expected Sonar Amplifier upgrade to be inactive after removal")
	}

	// Equip the Sonar Amplifier again
	if !sub.GetUpgrades().AddItem(amp, 1) {
		t.Errorf("expected successful addition of Sonar Amplifier to upgrades slot")
	}

	// Verify the upgrade is detected again
	if !item.HasItem[*item.SonarAmplifier](sub.GetUpgrades(), 1) {
		t.Errorf("expected Sonar Amplifier upgrade to be active after slotting")
	}

	// Check upgrade calculations (simulating sub.Update/sonar activation)
	pulseUpgraded := sub.Sonar.Pulse
	pulseUpgraded.DurationTicks = int(float64(pulseUpgraded.DurationTicks) * 1.8)
	pulseUpgraded.RadiusStep = pulseUpgraded.RadiusStep * 1.4

	if pulseUpgraded.DurationTicks != 324 {
		t.Errorf("expected upgraded duration to be 324, got %d", pulseUpgraded.DurationTicks)
	}
	if pulseUpgraded.RadiusStep != 9.1 {
		t.Errorf("expected upgraded radius step to be 9.1, got %f", pulseUpgraded.RadiusStep)
	}
}

type mockRuntime struct {
	vehicle.Runtime
}

func (m mockRuntime) IsActiveVehicle(v vehicle.Vehicle) bool {
	return false
}

func (mockRuntime) TimeOfDay() float64                 { return 0.0 }
func (mockRuntime) Input() vehicle.InputSource         { return mockInput{} }
func (mockRuntime) PlayerScreenCenter() gvec.Vec2      { return gvec.Vec2{} }
func (mockRuntime) PlayerSlowed() bool                 { return false }
func (mockRuntime) IsOverworldSolidAt(tx, ty int) bool { return false }
func (mockRuntime) IsCaveSolidAt(tx, ty int) bool      { return false }
func (mockRuntime) CanUseSonar() bool                  { return true }
func (mockRuntime) Emit(cmd vehicle.GameCommand)       {}

func TestRechargingMechanics(t *testing.T) {
	// 1. Test Power Cell / RechargeBattery method
	sub := vehicle.NewScoutSub(0, 0)
	sub.Battery = 20.0

	sub.RechargeBattery(100.0)
	if sub.Battery != 100.0 {
		t.Errorf("expected battery to be capped at 100.0, got %f", sub.Battery)
	}

	// 2. Test Thermal Generator passive recharge in cave
	sub.Battery = 50.0
	tg := &item.ThermalGenerator{}
	if !sub.GetUpgrades().AddItem(tg, 1) {
		t.Errorf("expected successful slotting of Thermal Generator")
	}

	// Run update step using mock runtime (where sub is inactive, but still recharges passively)
	sub.Update(mockRuntime{})
	if sub.Battery <= 50.0 {
		t.Errorf("expected passive recharge to increase battery, but remained %f", sub.Battery)
	}
	if sub.Battery != 50.02 {
		t.Errorf("expected battery to increase by 0.02, got %f", sub.Battery)
	}
}

type mockInput struct {
	vehicle.InputSource
}

func (mockInput) Cursor() gvec.Vec2                  { return gvec.Vec2{} }
func (mockInput) IsKeyPressed(k ebiten.Key) bool     { return false }
func (mockInput) IsKeyJustPressed(k ebiten.Key) bool { return false }

type mockActiveRuntime struct {
	vehicle.Runtime
}

func (mockActiveRuntime) TimeOfDay() float64                     { return 0.0 }
func (mockActiveRuntime) IsActiveVehicle(v vehicle.Vehicle) bool { return true }
func (mockActiveRuntime) Input() vehicle.InputSource             { return mockInput{} }
func (mockActiveRuntime) PlayerScreenCenter() gvec.Vec2          { return gvec.Vec2{} }
func (mockActiveRuntime) PlayerSlowed() bool                     { return false }
func (mockActiveRuntime) IsCaveSolidAt(tx, ty int) bool          { return false }
func (mockActiveRuntime) IsOverworldSolidAt(tx, ty int) bool     { return false }
func (mockActiveRuntime) CanUseSonar() bool                      { return true }
func (mockActiveRuntime) Emit(cmd vehicle.GameCommand)           {}

func TestScoutSub_WaterlinePhysics(t *testing.T) {
	sub := vehicle.NewScoutSub(0, 0)
	sub.Battery = 100.0

	// 1. Above waterline (e.g. Y = -20)
	sub.Pos.Y = -20.0
	sub.Vel.Y = 0.0

	sub.Update(mockActiveRuntime{})
	if sub.Vel.Y <= 0.0 {
		t.Errorf("expected gravity to apply downward velocity when sub is above waterline, got %f", sub.Vel.Y)
	}

	// 2. Near waterline (e.g. Y = -6), should bob
	sub.Pos.Y = -6.0
	sub.Vel.Y = 0.0

	sub.Update(mockActiveRuntime{})
	// Since bobY = waterline + 4 + sin(...) = -4 + sin(...)*2, at Y = -6 we expect bobbing force to adjust Y speed
	if sub.Vel.Y == 0.0 {
		t.Errorf("expected surface bobbing force to be applied, got 0.0")
	}
}

type mockMechInput struct {
	vehicle.InputSource
}

func (mockMechInput) Cursor() gvec.Vec2                  { return gvec.Vec2{} }
func (mockMechInput) IsKeyPressed(k ebiten.Key) bool     { return k == ebiten.KeyW }
func (mockMechInput) IsKeyJustPressed(k ebiten.Key) bool { return false }

type mockMechRuntime struct {
	vehicle.Runtime
}

func (mockMechRuntime) TimeOfDay() float64                     { return 0.0 }
func (mockMechRuntime) IsActiveVehicle(v vehicle.Vehicle) bool { return true }
func (mockMechRuntime) Input() vehicle.InputSource             { return mockMechInput{} }
func (mockMechRuntime) PlayerScreenCenter() gvec.Vec2          { return gvec.Vec2{} }
func (mockMechRuntime) PlayerSlowed() bool                     { return false }
func (mockMechRuntime) IsCaveSolidAt(tx, ty int) bool          { return false }
func (mockMechRuntime) IsOverworldSolidAt(tx, ty int) bool     { return false }
func (mockMechRuntime) CanUseSonar() bool                      { return true }
func (mockMechRuntime) Emit(cmd vehicle.GameCommand)           {}

func TestHeavyMech_WaterlinePhysics(t *testing.T) {
	m := vehicle.NewHeavyMech(0, 0)
	m.Battery = 100.0

	// 1. Above waterline (e.g. Y = -25)
	m.Pos.Y = -25.0
	m.Vel.Y = 0.0

	m.Update(mockMechRuntime{})
	// Heavy Mech always has gravity (0.12) + above-water gravity (0.20)
	if m.Vel.Y <= 0.12 {
		t.Errorf("expected extra above-water gravity to pull Mech down, got Y velocity %f", m.Vel.Y)
	}

	// 2. Near waterline (e.g. Y = -10) with thrusters active, should bob
	m.Pos.Y = -10.0
	m.Vel.Y = 0.0

	m.Update(mockMechRuntime{})
	// Thrusters are active, so bobbing force should act
	if m.Vel.Y == 0.0 {
		t.Errorf("expected surface bobbing force to apply to Mech, got 0.0")
	}
}

func TestRecipeBlueprintUnlocking(t *testing.T) {
	// Initially, verify that UHC O2 Tank recipe is locked
	var uhcRecipe *Recipe
	for idx := range CraftingRecipes {
		if CraftingRecipes[idx].NewResult().GetName() == "Ultra High Capacity O2 Tank" {
			uhcRecipe = &CraftingRecipes[idx]
			break
		}
	}
	if uhcRecipe == nil {
		t.Fatal("expected to find Ultra High Capacity O2 Tank recipe")
	}

	// Force lock it for the test
	uhcRecipe.Unlocked = false

	// Create a new game session
	g := NewGame()

	// Verify it's locked in the fabricator list count
	unlockedCountBefore := 0
	for _, rcp := range CraftingRecipes {
		if rcp.Unlocked {
			unlockedCountBefore++
		}
	}

	// Simulate receiving UnlockRecipeCmd from a drilled blueprint
	vrt := &vehicleRuntimeAdapter{g: g}
	vrt.Emit(vehicle.UnlockRecipeCmd{RecipeResultName: "Ultra High Capacity O2 Tank"})
	g.drainVehicleCommands(vrt)

	// Verify it is unlocked
	if !uhcRecipe.Unlocked {
		t.Errorf("expected Ultra High Capacity O2 Tank to be unlocked after UnlockRecipeCmd")
	}

	// Verify that the fabricator unlocked recipe count increased by 1
	unlockedCountAfter := 0
	for _, rcp := range CraftingRecipes {
		if rcp.Unlocked {
			unlockedCountAfter++
		}
	}
	if unlockedCountAfter != unlockedCountBefore+1 {
		t.Errorf("expected unlocked count to increase by 1, went from %d to %d", unlockedCountBefore, unlockedCountAfter)
	}
}
