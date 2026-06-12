package game

import (
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/resource"
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
func (mockRuntime) BaseStationPos() (gvec.Vec2, gvec.Vec2) { return gvec.Vec2{}, gvec.Vec2{} }
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
func (mockActiveRuntime) BaseStationPos() (gvec.Vec2, gvec.Vec2) { return gvec.Vec2{}, gvec.Vec2{} }
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
func (mockMechRuntime) BaseStationPos() (gvec.Vec2, gvec.Vec2) { return gvec.Vec2{}, gvec.Vec2{} }
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
	// Create a new game session
	g := NewGame()
	recipes := g.GetCraftingRecipes()

	// Initially, verify that UHC O2 Tank recipe is locked
	var uhcRecipe *Recipe
	for idx := range recipes {
		if recipes[idx].NewResult().GetName() == "Ultra High Capacity O2 Tank" {
			uhcRecipe = &recipes[idx]
			break
		}
	}
	if uhcRecipe == nil {
		t.Fatal("expected to find Ultra High Capacity O2 Tank recipe")
	}

	// Force lock it for the test
	uhcRecipe.Unlocked = false

	// Verify it's locked in the fabricator list count
	unlockedCountBefore := 0
	for _, rcp := range recipes {
		if rcp.Unlocked {
			unlockedCountBefore++
		}
	}

	// Simulate receiving UnlockRecipeCmd from a drilled blueprint
	vrt := &vehicleRuntimeAdapter{g: g}
	vrt.Emit(vehicle.UnlockRecipeCmd{RecipeResultName: "Ultra High Capacity O2 Tank"})
	g.drainVehicleCommands(vrt)

	// Verify it is unlocked
	recipesAfter := g.GetCraftingRecipes()
	var uhcRecipeAfter *Recipe
	for idx := range recipesAfter {
		if recipesAfter[idx].NewResult().GetName() == "Ultra High Capacity O2 Tank" {
			uhcRecipeAfter = &recipesAfter[idx]
			break
		}
	}
	if uhcRecipeAfter == nil || !uhcRecipeAfter.Unlocked {
		t.Errorf("expected Ultra High Capacity O2 Tank to be unlocked after UnlockRecipeCmd")
	}

	// Verify that the fabricator unlocked recipe count increased by 1
	unlockedCountAfter := 0
	for _, rcp := range recipesAfter {
		if rcp.Unlocked {
			unlockedCountAfter++
		}
	}
	if unlockedCountAfter != unlockedCountBefore+1 {
		t.Errorf("expected unlocked count to increase by 1, went from %d to %d", unlockedCountBefore, unlockedCountAfter)
	}
}

func TestInventory_Resize_Compaction(t *testing.T) {
	inv := item.NewInventory(5)
	tit := &item.Titanium{}
	qz := &item.Quartz{}

	// Case 1: Compaction with loss
	inv.AddItem(tit, 9) // slot 0: Titanium x 9
	inv.AddItem(qz, 5)  // slot 1: Quartz x 5
	inv.AddItem(tit, 3) // slot 0 -> 10, slot 2 -> 2
	inv.AddItem(qz, 6)  // slot 1 -> 10, slot 3 -> 1
	inv.AddItem(tit, 4) // slot 2 -> 6

	lost := inv.Resize(2)
	if len(lost) != 2 {
		t.Errorf("expected 2 lost stacks, got %d", len(lost))
	}
	
	var titLost, qzLost int
	for _, stack := range lost {
		if _, ok := stack.Item.(*item.Titanium); ok {
			titLost = stack.Quantity
		}
		if _, ok := stack.Item.(*item.Quartz); ok {
			qzLost = stack.Quantity
		}
	}
	if titLost != 6 {
		t.Errorf("expected 6 Titanium lost, got %d", titLost)
	}
	if qzLost != 1 {
		t.Errorf("expected 1 Quartz lost, got %d", qzLost)
	}

	if inv.Count(tit) != 10 {
		t.Errorf("expected 10 Titanium left, got %d", inv.Count(tit))
	}
	if inv.Count(qz) != 10 {
		t.Errorf("expected 10 Quartz left, got %d", inv.Count(qz))
	}

	// Case 2: Compaction without loss
	inv2 := item.NewInventory(5)
	inv2.AddItem(tit, 8)
	inv2.AddItem(qz, 6)

	// Manually split so they occupy extra slots but can be compacted fully
	inv2.Slots[0].Quantity = 3
	inv2.Slots[2] = item.ItemStack{Item: tit, Quantity: 5}
	inv2.Slots[1].Quantity = 2
	inv2.Slots[3] = item.ItemStack{Item: qz, Quantity: 4}

	lost2 := inv2.Resize(2)
	if len(lost2) != 0 {
		t.Errorf("expected 0 lost stacks in compaction case, got %d", len(lost2))
	}
	if inv2.Count(tit) != 8 {
		t.Errorf("expected Titanium count to remain 8, got %d", inv2.Count(tit))
	}
	if inv2.Count(qz) != 6 {
		t.Errorf("expected Quartz count to remain 6, got %d", inv2.Count(qz))
	}
	if inv2.Slots[0].Quantity != 8 || inv2.Slots[1].Quantity != 6 {
		t.Errorf("expected compacted slot quantities to be 8 and 6, got %d and %d", inv2.Slots[0].Quantity, inv2.Slots[1].Quantity)
	}
}

func TestBaseMenu_UninstallStorage_OverflowRefusal(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.baseMenu)
	g.baseMenu.ActiveTab = 0
	g.Input = NewMockInput()

	// Install Storage MKII
	storageMKII := &item.UpgradeStorageMKII{}
	g.baseStation.InstallUpgrade(storageMKII)

	if len(g.baseStation.Storage.Slots) != 48 {
		t.Fatalf("expected storage slots to expand to 48, got %d", len(g.baseStation.Storage.Slots))
	}

	// Add 30 items to storage so it exceeds 24 slots capacity (Fins have max stack of 1)
	fins := &item.Fins{}
	for i := 0; i < 30; i++ {
		g.baseStation.Storage.AddItem(fins, 1)
	}

	// Find the upgrade slot index
	upgradeIdx := -1
	for i, slot := range g.baseStation.Upgrades.Slots {
		if slot.Item == storageMKII {
			upgradeIdx = i
			break
		}
	}
	if upgradeIdx == -1 {
		t.Fatal("expected storage MKII to be found in base upgrades")
	}

	// Simulate clicking to uninstall Storage MKII
	mockInput := g.Input.(*MockInput)
	const (
		panelW = 800
		panelH = 500
	)
	panelX := float64(config.ScreenWidth-panelW) / 2.0
	panelY := float64(config.ScreenHeight-panelH) / 2.0
	mockInput.CursorPos = gvec.Vec2{
		X: panelX + 445 + float64(upgradeIdx)*46 + 20,
		Y: panelY + 140 + 20,
	}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	// Update game
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Verify uninstallation is blocked
	if g.baseStation.Upgrades.Slots[upgradeIdx].Item != storageMKII {
		t.Error("expected Storage MKII to remain installed")
	}
	if len(g.baseStation.Storage.Slots) != 48 {
		t.Errorf("expected storage slots to remain 48, got %d", len(g.baseStation.Storage.Slots))
	}

	// Verify SetMineWarning message was displayed
	msg, _ := g.GetMineWarning()
	expectedMsg := "Vault has too many items to uninstall storage upgrade!"
	if msg != expectedMsg {
		t.Errorf("expected mine warning %q, got %q", expectedMsg, msg)
	}
}

func TestHeavyMechDrill_CargoFull(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.caveState)

	// Create a mech and make it active
	mech := vehicle.NewHeavyMech(100, 100)
	g.ActiveVehicle = mech
	g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], mech)

	// Fill mech's cargo hold completely
	cargo := mech.GetCargo()
	tit := &item.Titanium{}
	for i := 0; i < len(cargo.Slots); i++ {
		cargo.AddItem(tit, 10)
	}

	// Verify that adding another item to cargo fails
	if cargo.AddItem(&item.Quartz{}, 1) {
		t.Fatal("expected cargo hold to be full")
	}

	// Create a resource node at (5, 5) tile coordinates
	node := resource.NewCopperNode(5, 5)
	node.SetHitsToMine(1)
	g.caveState.Nodes = []resource.Resource{node}

	// Trigger drill strike on the node
	mech.DrillStrike(node)

	// Run update loop for 16 frames to finish the drill animation (timer starts at 15)
	for i := 0; i < 16; i++ {
		err := g.Update()
		if err != nil {
			t.Fatal(err)
		}
	}

	// Assert copper node remains in the cave nodes
	if len(g.caveState.Nodes) != 1 || g.caveState.Nodes[0] != node {
		t.Errorf("expected copper node to remain in the cave nodes, got %d nodes", len(g.caveState.Nodes))
	}
	// Assert hits were refunded to 1
	if node.GetHitsToMine() != 1 {
		t.Errorf("expected node hits to be refunded to 1, got %d", node.GetHitsToMine())
	}
	// Assert warning message was displayed
	msg, _ := g.GetMineWarning()
	expectedMsg := "Cargo hold full! Cannot mine resource."
	if msg != expectedMsg {
		t.Errorf("expected warning %q, got %q", expectedMsg, msg)
	}
}

func TestFalseBulbSnare_VehicleCollision(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.caveState)

	// Create a mech and make it active (piloted)
	mech := vehicle.NewHeavyMech(100, 100)
	g.ActiveVehicle = mech
	g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], mech)

	// Initial player health is max
	initialPlayerHp := g.player.CurrentHealth
	initialMechHp := mech.GetHealth()

	// Spawn a FalseBulbSnare overlapping the mech
	snare := &entity.FalseBulbSnare{
		BaseEntity: entity.BaseEntity{
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 16, Y: 16},
			Active:     true,
		},
	}
	g.caveState.Entities = []entity.CaveEntity{snare}

	// Update game once to process the update and collision
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Assert:
	// 1. Snare is no longer active
	if snare.IsActive() {
		t.Error("expected FalseBulbSnare to be inactive after collision")
	}
	// 2. Mech took hull damage (health decreased)
	if mech.GetHealth() >= initialMechHp {
		t.Errorf("expected active vehicle to take damage, went from %f to %f", initialMechHp, mech.GetHealth())
	}
	// 3. Player took ZERO personal damage
	if g.player.CurrentHealth != initialPlayerHp {
		t.Errorf("expected player to take zero damage, went from %f to %f", initialPlayerHp, g.player.CurrentHealth)
	}
	// 4. Correct warning message was displayed
	msg, _ := g.GetMineWarning()
	expectedMsg := "VEHICLE ATTACKED BY FALSE-BULB SNARE!"
	if msg != expectedMsg {
		t.Errorf("expected warning %q, got %q", expectedMsg, msg)
	}
}

func TestInventory_AddItemBlueprintNode(t *testing.T) {
	inv := item.NewInventory(5)
	node := resource.NewBlueprintNode(0, 0, "Ultra High Capacity O2 Tank")

	// Verify that adding a BlueprintNode (whose GetBaseItem returns nil) does not panic and returns false.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AddItem panicked with BlueprintNode: %v", r)
		}
	}()

	res := inv.AddItem(node, 1)
	if res {
		t.Error("expected AddItem with BlueprintNode to return false, got true")
	}
}
