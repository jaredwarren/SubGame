package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/save"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

func TestSkiff_DualDockingBays(t *testing.T) {
	skiff := vehicle.NewSkiff(100, 100)

	scout := vehicle.NewScoutSub(0, 0)
	scout.GetUpgrades().AddItem(&item.ScoutSubDepthMK1{}, 1)
	scout.GetCargo().AddItem(&item.Titanium{}, 5)
	scout.TakeDamage(20.0)
	scout.RechargeBattery(-30.0) // 70 battery

	mech := vehicle.NewHeavyMech(0, 0)
	mech.GetUpgrades().AddItem(&item.ChemicalDischarger{}, 1)
	mech.GetCargo().AddItem(&item.Copper{}, 3)
	mech.TakeDamage(15.0)

	// Dock both
	bay0, ok0 := skiff.Dock(scout)
	if !ok0 || bay0 != 0 {
		t.Fatalf("expected scout to dock in bay 0, got bay %d (ok=%v)", bay0, ok0)
	}

	bay1, ok1 := skiff.Dock(mech)
	if !ok1 || bay1 != 1 {
		t.Fatalf("expected mech to dock in bay 1, got bay %d (ok=%v)", bay1, ok1)
	}

	if !skiff.HasDocked(vehicle.VehicleScoutSub) {
		t.Error("expected Skiff to report HasDocked ScoutSub")
	}
	if !skiff.HasDocked(vehicle.VehicleHeavyMech) {
		t.Error("expected Skiff to report HasDocked HeavyMech")
	}

	// Undock scout and verify preserved state
	undockedScout, ok := skiff.Undock(0, 200, 200)
	if !ok || undockedScout == nil {
		t.Fatal("failed to undock scout sub")
	}
	if undockedScout.GetHealth() != scout.GetHealth() {
		t.Errorf("scout health = %.1f, want %.1f", undockedScout.GetHealth(), scout.GetHealth())
	}
	if undockedScout.GetBattery() != scout.GetBattery() {
		t.Errorf("scout battery = %.1f, want %.1f", undockedScout.GetBattery(), scout.GetBattery())
	}
	if undockedScout.GetUpgrades().Count(&item.ScoutSubDepthMK1{}) != 1 {
		t.Error("undocked scout missing ScoutSubDepthMK1 upgrade")
	}
	if undockedScout.GetCargo().Count(&item.Titanium{}) != 5 {
		t.Errorf("undocked scout titanium = %d, want 5", undockedScout.GetCargo().Count(&item.Titanium{}))
	}
	if skiff.GetDocked(0) != nil {
		t.Error("expected bay 0 to be nil after undock")
	}

	// Undock mech and verify preserved state
	undockedMech, ok := skiff.Undock(1, 300, 300)
	if !ok || undockedMech == nil {
		t.Fatal("failed to undock heavy mech")
	}
	if undockedMech.GetHealth() != mech.GetHealth() {
		t.Errorf("mech health = %.1f, want %.1f", undockedMech.GetHealth(), mech.GetHealth())
	}
	if undockedMech.GetUpgrades().Count(&item.ChemicalDischarger{}) != 1 {
		t.Error("undocked mech missing ChemicalDischarger upgrade")
	}
	if undockedMech.GetCargo().Count(&item.Copper{}) != 3 {
		t.Errorf("undocked mech copper = %d, want 3", undockedMech.GetCargo().Count(&item.Copper{}))
	}
}

func TestGame_DeploySkiffAtBase(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil
	g.ActiveVehicle = nil

	if g.HasSkiff() {
		t.Fatal("expected HasSkiff to be false initially")
	}

	skiff := g.DeploySkiffAtBase()
	if skiff == nil {
		t.Fatal("expected DeploySkiffAtBase to return non-nil skiff")
	}
	if !g.HasSkiff() {
		t.Fatal("expected HasSkiff to be true after deployment")
	}
	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected 1 overworld vehicle, got %d", len(g.OverworldVehicles))
	}

	// Calling deploy again returns the existing skiff
	second := g.DeploySkiffAtBase()
	if second != skiff || len(g.OverworldVehicles) != 1 {
		t.Error("duplicate DeploySkiffAtBase should return existing skiff without creating a second one")
	}
}

func TestGame_OverworldDeploySubIntoCave(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil
	g.ActiveVehicle = nil

	skiff := g.DeploySkiffAtBase()
	skiff.SetDocked(0, vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub))
	g.ActiveVehicle = skiff

	g.DeploySubFromSkiff(0)

	if g.currentState != StateCave {
		t.Fatalf("expected state to be StateCave, got %v", g.currentState)
	}
	if g.ActiveVehicle == nil || g.ActiveVehicle.GetID() != vehicle.VehicleScoutSub {
		t.Fatalf("expected active vehicle to be ScoutSub in cave, got %v", g.ActiveVehicle)
	}
	if skiff.GetDocked(0) != nil {
		t.Error("expected Skiff Bay 0 to be empty after deploy")
	}
	if len(g.CaveVehicles[g.activeTrenchKey]) != 1 {
		t.Fatalf("expected 1 cave vehicle in active trench, got %d", len(g.CaveVehicles[g.activeTrenchKey]))
	}
}

func TestGame_InCaveTetherDeploy(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil

	skiff := g.DeploySkiffAtBase()
	skiff.SetDocked(0, vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub))
	skiff.SetDocked(1, vehicle.NewDefaultDockedVehicle(vehicle.VehicleHeavyMech))

	// Player dives into cave on foot
	g.EnterCave(50, 50)
	if g.ActiveVehicle != nil {
		t.Fatal("expected player to be on foot upon diving")
	}

	// Deploy Scout Sub via tether
	g.DeploySubInCave(0)
	if g.ActiveVehicle == nil || g.ActiveVehicle.GetID() != vehicle.VehicleScoutSub {
		t.Fatalf("expected active vehicle to be ScoutSub after tether deploy, got %v", g.ActiveVehicle)
	}
	if skiff.GetDocked(0) != nil {
		t.Error("expected Skiff Bay 0 to be empty after tether deploy")
	}

	// Exit Scout Sub on foot
	g.exitVehicle(g.ActiveVehicle.GetPos(), g.ActiveVehicle.GetDimensions())
	if g.ActiveVehicle != nil {
		t.Fatal("expected player to be on foot after exit")
	}

	// Deploy Heavy Mech via tether in the same cave
	g.DeploySubInCave(1)
	if g.ActiveVehicle == nil || g.ActiveVehicle.GetID() != vehicle.VehicleHeavyMech {
		t.Fatalf("expected active vehicle to be HeavyMech after tether deploy, got %v", g.ActiveVehicle)
	}
	if len(g.CaveVehicles[g.activeTrenchKey]) != 2 {
		t.Fatalf("expected 2 cave vehicles deployed in trench, got %d", len(g.CaveVehicles[g.activeTrenchKey]))
	}
}

func TestGame_SurfacingAutoDocksToSkiff(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil

	skiff := g.DeploySkiffAtBase()
	skiff.SetDocked(0, vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub))
	g.ActiveVehicle = skiff

	// Deploy sub into cave
	g.DeploySubFromSkiff(0)
	sub := g.ActiveVehicle

	// Put loot in sub cargo and install upgrade
	sub.GetCargo().AddItem(&item.Titanium{}, 7)
	sub.GetUpgrades().AddItem(&item.ScoutSubDepthMK1{}, 1)
	sub.TakeDamage(10.0)

	// Surface from cave while piloting sub
	g.ExitCave()

	if g.currentState != StateOverworld {
		t.Fatalf("expected state to be StateOverworld after ExitCave, got %v", g.currentState)
	}
	if g.ActiveVehicle != nil {
		t.Errorf("expected player on foot / out of sub upon surfacing, got active %v", g.ActiveVehicle)
	}

	// Verify sub auto-docked back to Skiff Bay 0
	docked := skiff.GetDocked(0)
	if docked == nil {
		t.Fatal("expected ScoutSub to be docked in Skiff Bay 0 upon surfacing")
	}
	if docked.ID != vehicle.VehicleScoutSub {
		t.Errorf("docked ID = %s, want %s", docked.ID, vehicle.VehicleScoutSub)
	}
	if docked.Cargo == nil || docked.Cargo.Count(&item.Titanium{}) != 7 {
		t.Errorf("docked cargo titanium count mismatch: %+v", docked.Cargo)
	}
	if docked.Upgrades == nil || docked.Upgrades.Count(&item.ScoutSubDepthMK1{}) != 1 {
		t.Errorf("docked upgrades missing ScoutSubDepthMK1: %+v", docked.Upgrades)
	}
	if len(g.CaveVehicles[g.activeTrenchKey]) != 0 {
		t.Errorf("expected 0 cave vehicles remaining in trench after auto-dock, got %d", len(g.CaveVehicles[g.activeTrenchKey]))
	}
}

func TestGame_WinchRecallSub(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil

	skiff := g.DeploySkiffAtBase()
	skiff.SetDocked(0, vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub))

	// Deploy sub into cave
	g.DeploySubFromSkiff(0)
	sub := g.ActiveVehicle
	sub.GetCargo().AddItem(&item.Nickel{}, 4)

	// Player exits vehicle and swims up to surface on foot
	g.exitVehicle(sub.GetPos(), sub.GetDimensions())
	g.ExitCave()

	if skiff.GetDocked(0) != nil {
		t.Fatal("sub should not be in dock since it was left in the cave")
	}
	if len(g.CaveVehicles[g.activeTrenchKey]) != 1 {
		t.Fatalf("expected 1 sub in cave vehicles, got %d", len(g.CaveVehicles[g.activeTrenchKey]))
	}

	// Winch recall sub from Skiff
	g.WinchRecallSub(0)

	docked := skiff.GetDocked(0)
	if docked == nil {
		t.Fatal("expected sub to be docked in Skiff after WinchRecallSub")
	}
	if docked.Cargo.Count(&item.Nickel{}) != 4 {
		t.Errorf("winched sub cargo nickel = %d, want 4", docked.Cargo.Count(&item.Nickel{}))
	}
	if len(g.CaveVehicles[g.activeTrenchKey]) != 0 {
		t.Errorf("expected 0 cave vehicles in trench after winch recall, got %d", len(g.CaveVehicles[g.activeTrenchKey]))
	}
}

func TestSave_V2ToV3Migration_LegacyKits(t *testing.T) {
	v2Data := &save.SaveData{
		Version: 2,
		Player: save.SavedPlayer{
			Inventory: item.SavedInventory{
				Size: 24,
				Slots: []item.SavedItemStack{
					{
						ItemID:   item.IDSkiffKit,
						ItemName: "Skiff Kit",
						Quantity: 1,
						Health:   140.0,
						Battery:  90.0,
					},
					{
						ItemID:   item.IDScoutSubKit,
						ItemName: "Scout Sub Kit",
						Quantity: 1,
						Health:   85.0,
						Battery:  75.0,
						Upgrades: &item.SavedInventory{
							Size: 2,
							Slots: []item.SavedItemStack{
								{ItemID: item.IDSonarAmplifier, Quantity: 1},
							},
						},
					},
				},
			},
			Hotbar: item.SavedInventory{
				Size: 5,
				Slots: []item.SavedItemStack{
					{
						ItemID:   item.IDHeavyMechKit,
						ItemName: "Heavy Mech Kit",
						Quantity: 1,
						Health:   180.0,
						Battery:  95.0,
					},
				},
			},
		},
	}

	err := save.MigrateSaveData(v2Data)
	if err != nil {
		t.Fatalf("MigrateSaveData failed: %v", err)
	}

	if v2Data.Version != 3 {
		t.Errorf("migrated save version = %d, want 3", v2Data.Version)
	}

	// Verify kits removed from player inventory & hotbar
	for _, s := range v2Data.Player.Inventory.Slots {
		if s.ItemID == item.IDSkiffKit || s.ItemID == item.IDScoutSubKit {
			t.Errorf("expected kit removed from inventory, found %+v", s)
		}
	}
	for _, s := range v2Data.Player.Hotbar.Slots {
		if s.ItemID == item.IDHeavyMechKit {
			t.Errorf("expected mech kit removed from hotbar, found %+v", s)
		}
	}

	// Verify Skiff created in Vehicles with docked subs
	if len(v2Data.Vehicles) != 1 {
		t.Fatalf("expected 1 vehicle (Skiff) created in save, got %d", len(v2Data.Vehicles))
	}
	skiffSave := v2Data.Vehicles[0]
	if skiffSave.ID != string(vehicle.VehicleSkiff) {
		t.Errorf("vehicle ID = %s, want %s", skiffSave.ID, vehicle.VehicleSkiff)
	}
	if len(skiffSave.Docked) != 2 {
		t.Fatalf("expected 2 docked vehicles in Skiff, got %d", len(skiffSave.Docked))
	}

	// Verify Scout Sub in Bay 0
	scoutDocked := skiffSave.Docked[0]
	if scoutDocked.BayIdx != 0 || scoutDocked.ID != string(vehicle.VehicleScoutSub) {
		t.Errorf("scout docked mismatch: %+v", scoutDocked)
	}
	if scoutDocked.Health != 85.0 || scoutDocked.Battery != 75.0 {
		t.Errorf("scout health/battery = %.1f/%.1f, want 85.0/75.0", scoutDocked.Health, scoutDocked.Battery)
	}
	if scoutDocked.Upgrades.Slots[0].ItemID != item.IDSonarAmplifier {
		t.Errorf("scout upgrades missing SonarAmplifier: %+v", scoutDocked.Upgrades)
	}

	// Verify Heavy Mech in Bay 1
	mechDocked := skiffSave.Docked[1]
	if mechDocked.BayIdx != 1 || mechDocked.ID != string(vehicle.VehicleHeavyMech) {
		t.Errorf("mech docked mismatch: %+v", mechDocked)
	}
	if mechDocked.Health != 180.0 || mechDocked.Battery != 95.0 {
		t.Errorf("mech health/battery = %.1f/%.1f, want 180.0/95.0", mechDocked.Health, mechDocked.Battery)
	}
}

func TestSkiff_DockedVehicleCharging_SingleSub(t *testing.T) {
	skiff := vehicle.NewSkiff(100, 100)
	skiff.Battery = 100.0

	dv := vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub)
	dv.Battery = 50.0
	skiff.SetDocked(0, dv)

	if !skiff.IsBayCharging(0) {
		t.Error("expected IsBayCharging(0) to be true")
	}

	// 10 seconds of charging at 1.0/sec
	skiff.UpdateDockedCharging(10.0)

	if dv.Battery != 60.0 {
		t.Errorf("docked sub battery = %.1f, want 60.0", dv.Battery)
	}
	if skiff.Battery != 90.0 {
		t.Errorf("skiff battery = %.1f, want 90.0", skiff.Battery)
	}
}

func TestSkiff_DockedVehicleCharging_SafetyReserve(t *testing.T) {
	skiff := vehicle.NewSkiff(100, 100)
	skiff.Battery = 25.0 // 5.0 units above 20.0 safety threshold

	dv := vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub)
	dv.Battery = 10.0
	skiff.SetDocked(0, dv)

	// Attempt 10 seconds of charging (wants 10 units, but only 5 available before safety limit)
	skiff.UpdateDockedCharging(10.0)

	if skiff.Battery != 20.0 {
		t.Errorf("skiff battery = %.1f, want safety reserve 20.0", skiff.Battery)
	}
	if dv.Battery != 15.0 {
		t.Errorf("docked sub battery = %.1f, want 15.0", dv.Battery)
	}
	if skiff.IsBayCharging(0) {
		t.Error("expected IsBayCharging to be false once safety reserve is reached")
	}
}

func TestSkiff_DockedVehicleCharging_DualSubSplit(t *testing.T) {
	skiff := vehicle.NewSkiff(100, 100)
	skiff.Battery = 100.0

	sub0 := vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub)
	sub0.Battery = 50.0
	skiff.SetDocked(0, sub0)

	sub1 := vehicle.NewDefaultDockedVehicle(vehicle.VehicleHeavyMech)
	sub1.Battery = 50.0
	skiff.SetDocked(1, sub1)

	// 10 seconds of charging (10.0 total power transferred, 5.0 to each sub)
	skiff.UpdateDockedCharging(10.0)

	if sub0.Battery != 55.0 {
		t.Errorf("sub0 battery = %.1f, want 55.0", sub0.Battery)
	}
	if sub1.Battery != 55.0 {
		t.Errorf("sub1 battery = %.1f, want 55.0", sub1.Battery)
	}
	if skiff.Battery != 90.0 {
		t.Errorf("skiff battery = %.1f, want 90.0", skiff.Battery)
	}
}

func TestSkiff_DockedVehicleCharging_DuringCaveDive(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil

	skiff := g.DeploySkiffAtBase()
	skiff.Battery = 90.0

	// Dock a scout sub with 40% battery
	sub := vehicle.NewDefaultDockedVehicle(vehicle.VehicleScoutSub)
	sub.Battery = 40.0
	skiff.SetDocked(0, sub)

	// Enter cave on foot
	g.EnterCave(50, 50)
	if g.currentState != StateCave {
		t.Fatalf("expected StateCave, got %v", g.currentState)
	}

	// Tick game loop for 120 frames (2 seconds)
	for i := 0; i < 120; i++ {
		g.updateIdleVehicles(g.vehicleRT)
	}

	// Sub should have recharged in background while player was in cave
	if sub.Battery <= 40.0 {
		t.Errorf("expected sub battery to increase during cave dive, got %.1f", sub.Battery)
	}
}
