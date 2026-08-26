package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

func TestScoutSub_PickUpAndRedeployPreservesUpgrades(t *testing.T) {
	g := NewGame()
	g.currentState = StateCave
	g.activeTrenchKey = "0_0"

	sub := vehicle.NewScoutSub(100, 100)
	g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], sub)
	g.ActiveVehicle = sub

	// Customize upgrades
	sub.GetUpgrades().Clear()
	sub.GetUpgrades().AddItem(&item.SonarAmplifier{}, 1)
	sub.GetUpgrades().AddItem(&item.DecoyLauncher{}, 1)
	sub.TakeDamage(25.0)
	sub.RechargeBattery(-15.0) // battery = 85

	origHealth := sub.GetHealth()
	origBattery := sub.GetBattery()

	g.player.Inventory.Clear()
	g.PickUpActiveVehicle()

	if g.ActiveVehicle != nil {
		t.Fatal("expected active vehicle to be nil after pickup")
	}

	// Verify kit in player inventory
	var subKit *vehicle.ScoutSubKit
	for _, slot := range g.player.Inventory.Slots {
		if kit, ok := slot.Item.(*vehicle.ScoutSubKit); ok {
			subKit = kit
			break
		}
	}
	if subKit == nil {
		t.Fatal("expected ScoutSubKit in player inventory")
	}
	if subKit.Upgrades == nil || subKit.Upgrades.Count(&item.SonarAmplifier{}) != 1 || subKit.Upgrades.Count(&item.DecoyLauncher{}) != 1 {
		t.Errorf("sub kit upgrades mismatch: %+v", subKit.Upgrades)
	}
	if subKit.Health != origHealth {
		t.Errorf("sub kit health = %.1f, want %.1f", subKit.Health, origHealth)
	}
	if subKit.Battery != origBattery {
		t.Errorf("sub kit battery = %.1f, want %.1f", subKit.Battery, origBattery)
	}

	// Deploy the sub kit again
	g.ActivatePlayerItem(subKit)

	if len(g.CaveVehicles[g.activeTrenchKey]) != 1 {
		t.Fatalf("expected 1 cave vehicle after deploy, got %d", len(g.CaveVehicles[g.activeTrenchKey]))
	}
	deployed := g.CaveVehicles[g.activeTrenchKey][0]
	if deployed.GetUpgrades() == nil {
		t.Fatal("deployed vehicle has nil upgrades")
	}
	if deployed.GetUpgrades().Count(&item.SonarAmplifier{}) != 1 {
		t.Errorf("deployed vehicle missing SonarAmplifier")
	}
	if deployed.GetUpgrades().Count(&item.DecoyLauncher{}) != 1 {
		t.Errorf("deployed vehicle missing DecoyLauncher")
	}
	if deployed.GetHealth() != origHealth {
		t.Errorf("deployed vehicle health = %.1f, want %.1f", deployed.GetHealth(), origHealth)
	}
	if deployed.GetBattery() != origBattery {
		t.Errorf("deployed vehicle battery = %.1f, want %.1f", deployed.GetBattery(), origBattery)
	}
}

func TestHeavyMech_PickUpAndRedeployPreservesUpgrades(t *testing.T) {
	g := NewGame()
	g.currentState = StateCave
	g.activeTrenchKey = "0_0"

	mech := vehicle.NewHeavyMech(100, 100)
	g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], mech)
	g.ActiveVehicle = mech

	mech.GetUpgrades().Clear()
	mech.GetUpgrades().AddItem(&item.ChemicalDischarger{}, 1)
	mech.TakeDamage(10.0)

	origHealth := mech.GetHealth()

	g.player.Inventory.Clear()
	g.PickUpActiveVehicle()

	var mechKit *vehicle.HeavyMechKit
	for _, slot := range g.player.Inventory.Slots {
		if kit, ok := slot.Item.(*vehicle.HeavyMechKit); ok {
			mechKit = kit
			break
		}
	}
	if mechKit == nil {
		t.Fatal("expected HeavyMechKit in player inventory")
	}
	if mechKit.Upgrades == nil || mechKit.Upgrades.Count(&item.ChemicalDischarger{}) != 1 {
		t.Errorf("mech kit upgrades mismatch: %+v", mechKit.Upgrades)
	}

	// Deploy mech kit
	g.ActivatePlayerItem(mechKit)

	if len(g.CaveVehicles[g.activeTrenchKey]) != 1 {
		t.Fatalf("expected 1 cave vehicle after deploy, got %d", len(g.CaveVehicles[g.activeTrenchKey]))
	}
	deployed := g.CaveVehicles[g.activeTrenchKey][0]
	if deployed.GetUpgrades().Count(&item.ChemicalDischarger{}) != 1 {
		t.Errorf("deployed mech missing ChemicalDischarger")
	}
	if deployed.GetHealth() != origHealth {
		t.Errorf("deployed mech health = %.1f, want %.1f", deployed.GetHealth(), origHealth)
	}
}

func TestVehicle_CargoOverflowCreatesFloatingCrate(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld

	skiff := vehicle.NewSkiff(100, 100)
	g.OverworldVehicles = append(g.OverworldVehicles, skiff)
	g.ActiveVehicle = skiff

	// Put cargo in skiff
	skiff.GetCargo().AddItem(&item.Titanium{}, 10)
	skiff.GetCargo().AddItem(&item.Copper{}, 10)

	// Fill player inventory almost completely (leave exactly 1 slot for skiff kit)
	g.player.Inventory.Clear()
	for i := 0; i < len(g.player.Inventory.Slots)-1; i++ {
		g.player.Inventory.Slots[i] = item.ItemStack{Item: &item.Quartz{}, Quantity: 10}
	}
	g.player.Hotbar.Clear()
	for i := 0; i < len(g.player.Hotbar.Slots); i++ {
		g.player.Hotbar.Slots[i] = item.ItemStack{Item: &item.Quartz{}, Quantity: 10}
	}

	g.PickUpActiveVehicle()

	if g.ActiveVehicle != nil {
		t.Fatal("expected vehicle pickup to succeed")
	}

	// Verify skiff kit is in player inventory
	if !item.HasItem[*vehicle.SkiffKit](g.player.Inventory, 1) {
		t.Fatal("expected SkiffKit in player inventory")
	}

	// Verify excess cargo was dropped in floating crate (LostCargoBeacon)
	if len(g.lostCargo) != 1 {
		t.Fatalf("expected 1 lost cargo beacon for overflow, got %d", len(g.lostCargo))
	}
	beacon := g.lostCargo[0]
	if beacon.Cargo.Count(&item.Titanium{}) != 10 {
		t.Errorf("overflow titanium = %d, want 10", beacon.Cargo.Count(&item.Titanium{}))
	}
	if beacon.Cargo.Count(&item.Copper{}) != 10 {
		t.Errorf("overflow copper = %d, want 10", beacon.Cargo.Count(&item.Copper{}))
	}
}

func TestVehicleKit_SaveLoadPersistence(t *testing.T) {
	inv := item.NewInventory(5)

	subKit := &vehicle.ScoutSubKit{
		Upgrades: item.NewInventory(4),
		Health:   77.0,
		Battery:  65.0,
		HasState: true,
	}
	subKit.Upgrades.AddItem(&item.SonarAmplifier{}, 1)
	subKit.Upgrades.AddItem(&item.DecoyLauncher{}, 1)

	inv.Slots[0] = item.ItemStack{Item: subKit, Quantity: 1}

	// Serialize
	saved := inv.SerializeState()

	// Deserialize
	restored := item.DeserializeInventory(saved)

	if restored.Slots[0].Item == nil {
		t.Fatal("restored item is nil")
	}
	kit, ok := restored.Slots[0].Item.(*vehicle.ScoutSubKit)
	if !ok {
		t.Fatalf("restored item is %T, want *vehicle.ScoutSubKit", restored.Slots[0].Item)
	}
	if !kit.HasState {
		t.Error("kit.HasState is false")
	}
	if kit.Health != 77.0 {
		t.Errorf("kit.Health = %.1f, want 77.0", kit.Health)
	}
	if kit.Battery != 65.0 {
		t.Errorf("kit.Battery = %.1f, want 65.0", kit.Battery)
	}
	if kit.Upgrades == nil {
		t.Fatal("kit.Upgrades is nil")
	}
	if kit.Upgrades.Count(&item.SonarAmplifier{}) != 1 {
		t.Errorf("missing SonarAmplifier in restored upgrades")
	}
	if kit.Upgrades.Count(&item.DecoyLauncher{}) != 1 {
		t.Errorf("missing DecoyLauncher in restored upgrades")
	}

	// Deploy restored kit
	deployed := kit.Deploy(200, 300)
	if deployed.GetHealth() != 77.0 {
		t.Errorf("deployed health = %.1f, want 77.0", deployed.GetHealth())
	}
	if deployed.GetBattery() != 65.0 {
		t.Errorf("deployed battery = %.1f, want 65.0", deployed.GetBattery())
	}
	if deployed.GetUpgrades().Count(&item.SonarAmplifier{}) != 1 {
		t.Errorf("deployed missing SonarAmplifier")
	}
	if deployed.GetUpgrades().Count(&item.DecoyLauncher{}) != 1 {
		t.Errorf("deployed missing DecoyLauncher")
	}
}

func TestSkiff_PickUpAndRedeployPreservesUpgrades(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld

	skiff := vehicle.NewSkiff(100, 100)
	g.OverworldVehicles = append(g.OverworldVehicles, skiff)
	g.ActiveVehicle = skiff

	if skiff.GetUpgrades() == nil {
		t.Fatal("skiff Upgrades should not be nil")
	}
	if len(skiff.GetUpgrades().Slots) != 3 {
		t.Fatalf("expected 3 upgrade slots on Skiff, got %d", len(skiff.GetUpgrades().Slots))
	}

	skiff.GetUpgrades().AddItem(&item.SurfaceSonar{}, 1)
	skiff.TakeDamage(30.0)
	skiff.RechargeBattery(-20.0) // 80 battery

	origHealth := skiff.GetHealth()
	origBattery := skiff.GetBattery()

	g.player.Inventory.Clear()
	g.PickUpActiveVehicle()

	var skiffKit *vehicle.SkiffKit
	for _, slot := range g.player.Inventory.Slots {
		if kit, ok := slot.Item.(*vehicle.SkiffKit); ok {
			skiffKit = kit
			break
		}
	}
	if skiffKit == nil {
		t.Fatal("expected SkiffKit in player inventory")
	}
	if skiffKit.Upgrades == nil || skiffKit.Upgrades.Count(&item.SurfaceSonar{}) != 1 {
		t.Errorf("skiff kit upgrades mismatch: %+v", skiffKit.Upgrades)
	}

	// Redeploy the skiff
	g.ActivatePlayerItem(skiffKit)
	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected 1 overworld vehicle after redeploy, got %d", len(g.OverworldVehicles))
	}
	deployed := g.OverworldVehicles[0]
	if deployed.GetUpgrades() == nil || deployed.GetUpgrades().Count(&item.SurfaceSonar{}) != 1 {
		t.Errorf("deployed skiff missing SurfaceSonar upgrade")
	}
	if deployed.GetHealth() != origHealth {
		t.Errorf("deployed skiff health = %.1f, want %.1f", deployed.GetHealth(), origHealth)
	}
	if deployed.GetBattery() != origBattery {
		t.Errorf("deployed skiff battery = %.1f, want %.1f", deployed.GetBattery(), origBattery)
	}
}
