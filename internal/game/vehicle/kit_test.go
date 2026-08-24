package vehicle

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
)

func TestCloneInventory(t *testing.T) {
	if CloneInventory(nil) != nil {
		t.Fatal("nil inventory should clone to nil")
	}
	src := item.NewInventory(2)
	src.AddItem(&item.SonarAmplifier{}, 1)
	dst := CloneInventory(src)
	if dst == src {
		t.Fatal("clone should allocate a new inventory")
	}
	if !item.HasItem[*item.SonarAmplifier](dst, 1) {
		t.Fatal("expected cloned upgrade in inventory")
	}
}

func TestRestoreKitState(t *testing.T) {
	upg := item.NewInventory(1)
	upg.AddItem(&item.DecoyLauncher{}, 1)
	health := 50.0
	battery := 25.0
	var upgrades *item.Inventory
	RestoreKitState(&health, &battery, &upgrades, KitVehicleState{
		Upgrades: upg,
		Health:   80,
		Battery:  90,
		HasState: true,
	})
	if health != 80 || battery != 90 {
		t.Fatalf("health/battery not restored: %v %v", health, battery)
	}
	if !item.HasItem[*item.DecoyLauncher](upgrades, 1) {
		t.Fatal("expected restored upgrades")
	}
}
