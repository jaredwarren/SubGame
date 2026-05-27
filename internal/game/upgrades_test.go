package game

import (
	"reflect"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
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
	b := NewBaseStation(0, 0)

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
	b := NewBaseStation(0, 0)

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
