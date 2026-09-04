package vehicle

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
)

func TestMiniLifepod_CreationAndDefaults(t *testing.T) {
	pod := NewMiniLifepod(100, 200)
	if pod == nil {
		t.Fatal("expected NewMiniLifepod to return non-nil")
	}
	if pod.GetID() != VehicleMiniLifepod {
		t.Errorf("got ID %s, want %s", pod.GetID(), VehicleMiniLifepod)
	}
	if pod.GetName() != "Mini-Lifepod" {
		t.Errorf("got name %s, want Mini-Lifepod", pod.GetName())
	}
	if pod.GetPerspective() != "overworld" {
		t.Errorf("got perspective %s, want overworld", pod.GetPerspective())
	}
	if pod.GetHealth() != MiniLifepodArchetype.MaxHealth {
		t.Errorf("got health %.1f, want %.1f", pod.GetHealth(), MiniLifepodArchetype.MaxHealth)
	}
	if pod.GetBattery() != MiniLifepodArchetype.MaxBattery {
		t.Errorf("got battery %.1f, want %.1f", pod.GetBattery(), MiniLifepodArchetype.MaxBattery)
	}
	if len(pod.GetUpgrades().Slots) != MiniLifepodArchetype.UpgradeSlots {
		t.Errorf("got upgrade slots %d, want %d", len(pod.GetUpgrades().Slots), MiniLifepodArchetype.UpgradeSlots)
	}
	if !pod.HasModule(item.ModuleFabricator) {
		t.Error("expected Fabricator module to be active by default")
	}
	if !pod.HasModule(item.ModuleMedical) {
		t.Error("expected Medical module to be active by default")
	}
	if pod.HasModule(item.ModuleSolar) {
		t.Error("expected Solar module to be inactive initially")
	}
}

func TestMiniLifepod_SolarModuleInstallation(t *testing.T) {
	pod := NewMiniLifepod(0, 0)
	pod.Upgrades.AddItem(&item.UpgradeSolar{}, 1)
	pod.RecalculateProperties()

	if !pod.HasModule(item.ModuleSolar) {
		t.Error("expected Solar module to become active after installing UpgradeSolar")
	}
	if pod.SolarRechargeRate <= 0.01 {
		t.Errorf("expected solar recharge rate > 0.01, got %.4f", pod.SolarRechargeRate)
	}
}

func TestMiniLifepod_KitAndDeploy(t *testing.T) {
	pod := NewMiniLifepod(50, 60)
	pod.Health = 85.0
	pod.Battery = 35.0
	pod.Upgrades.AddItem(&item.UpgradeSolar{}, 1)
	pod.RecalculateProperties()

	kitItem := pod.GetKit()
	kit, ok := kitItem.(*MiniLifepodKit)
	if !ok {
		t.Fatalf("expected *MiniLifepodKit, got %T", kitItem)
	}
	if kit.GetID() != item.IDMiniLifepodKit {
		t.Errorf("got kit ID %s, want %s", kit.GetID(), item.IDMiniLifepodKit)
	}

	deployedVehicle := kit.Deploy(200, 300)
	deployedPod, ok := deployedVehicle.(*MiniLifepod)
	if !ok {
		t.Fatalf("expected *MiniLifepod, got %T", deployedVehicle)
	}
	if deployedPod.Pos.X != 200 || deployedPod.Pos.Y != 300 {
		t.Errorf("deployed at (%.1f, %.1f), want (200, 300)", deployedPod.Pos.X, deployedPod.Pos.Y)
	}
	if deployedPod.Health != 85.0 {
		t.Errorf("got health %.1f, want 85.0", deployedPod.Health)
	}
	if deployedPod.Battery != 35.0 {
		t.Errorf("got battery %.1f, want 35.0", deployedPod.Battery)
	}
	if !deployedPod.HasModule(item.ModuleSolar) {
		t.Error("expected deployed pod to retain installed Solar module")
	}
}

func TestSkiff_DockAndUndockMiniLifepod(t *testing.T) {
	skiff := NewSkiff(0, 0)
	pod := NewMiniLifepod(10, 10)
	pod.Battery = 40.0

	bayIdx, ok := skiff.Dock(pod)
	if !ok || bayIdx != 2 {
		t.Fatalf("expected MiniLifepod to dock at bay 2, got bay %d (ok=%v)", bayIdx, ok)
	}
	if !skiff.HasDocked(VehicleMiniLifepod) {
		t.Error("expected Skiff to report HasDocked(VehicleMiniLifepod) = true")
	}

	docked := skiff.GetDocked(2)
	if docked == nil || docked.ID != VehicleMiniLifepod {
		t.Fatalf("expected docked vehicle at bay 2 to be VehicleMiniLifepod, got %+v", docked)
	}
	if docked.GetName() != "Mini-Lifepod" {
		t.Errorf("docked vehicle GetName() = %s, want Mini-Lifepod", docked.GetName())
	}

	undocked, ok := skiff.Undock(2, 50, 75)
	if !ok || undocked == nil {
		t.Fatal("expected successful undock from bay 2")
	}
	undockedPod, ok := undocked.(*MiniLifepod)
	if !ok {
		t.Fatalf("expected *MiniLifepod, got %T", undocked)
	}
	if undockedPod.Pos.X != 50 || undockedPod.Pos.Y != 75 {
		t.Errorf("undocked pos = %+v, want (50, 75)", undockedPod.Pos)
	}
	if skiff.GetDocked(2) != nil {
		t.Error("expected bay 2 to be empty after undock")
	}
}

func TestSkiff_ChargesDockedMiniLifepod(t *testing.T) {
	skiff := NewSkiff(0, 0)
	skiff.Battery = 100.0

	pod := NewMiniLifepod(0, 0)
	pod.Battery = 20.0
	skiff.Dock(pod)

	if !skiff.IsBayCharging(2) {
		t.Error("expected bay 2 to report IsBayCharging = true")
	}

	skiff.UpdateDockedCharging(1.0)
	docked := skiff.GetDocked(2)
	if docked.Battery <= 20.0 {
		t.Errorf("expected docked pod battery > 20.0 after charging, got %.2f", docked.Battery)
	}
}
