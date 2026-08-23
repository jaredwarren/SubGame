package save

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	savePath := filepath.Join(tempDir, "test_save.json")

	// Construct test save data
	inv := item.NewInventory(10)
	inv.AddItem(&item.Titanium{}, 5)
	inv.AddItem(&item.Copper{}, 3)

	original := &SaveData{
		Version:        1,
		Timestamp:      123456789,
		WorldSeed:      98765,
		TimeOfDay:      1200.0,
		Ticks:          5000.0,
		TutorialActive: true,
		SceneState:     "Overworld",
		LastOverworldX: 320.0,
		LastOverworldY: 640.0,
		Player: SavedPlayer{
			PosX:       320.0,
			PosY:       640.0,
			Health:     100.0,
			MaxHealth:  100.0,
			Oxygen:     45.0,
			MaxOxygen:  45.0,
			Inventory:  inv.SerializeState(),
			ActiveSlot: 1,
		},
		BaseStation: SavedBaseStation{
			PosX:    400.0,
			PosY:    600.0,
			Storage: item.NewInventory(16).SerializeState(),
		},
		Vehicles: []SavedVehicle{
			{
				Type:     "Skiff",
				PosX:     300.0,
				PosY:     600.0,
				Health:   100.0,
				IsActive: true,
				Location: "overworld",
			},
		},
		Story:               []string{"LOG_01", "LOG_02"},
		UnlockedRecipeNames: []string{"Scout Sub Kit"},
		LostCargo: []SavedLostCargo{
			{
				PosX:          640,
				PosY:          320,
				LifetimeTicks: 4000,
				Cargo:         inv.SerializeState(),
			},
		},
	}

	if err := SaveToFile(savePath, original); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	loaded, err := LoadFromFile(savePath)
	if err != nil {
		t.Fatalf("Failed to load save file: %v", err)
	}

	if loaded.WorldSeed != original.WorldSeed {
		t.Errorf("Expected seed %d, got %d", original.WorldSeed, loaded.WorldSeed)
	}
	if loaded.Player.PosX != original.Player.PosX || loaded.Player.PosY != original.Player.PosY {
		t.Errorf("Player pos mismatch")
	}
	if len(loaded.Story) != 2 || loaded.Story[0] != "LOG_01" {
		t.Errorf("Story data mismatch: %v", loaded.Story)
	}

	// Restore inventory
	restoredInv := item.DeserializeInventory(loaded.Player.Inventory)
	if !item.HasItem[*item.Titanium](restoredInv, 5) {
		t.Errorf("Failed to restore 5 Titanium from save")
	}
	if !item.HasItem[*item.Copper](restoredInv, 3) {
		t.Errorf("Failed to restore 3 Copper from save")
	}
	if len(loaded.UnlockedRecipeNames) != 1 || loaded.UnlockedRecipeNames[0] != "Scout Sub Kit" {
		t.Errorf("UnlockedRecipeNames mismatch: %v", loaded.UnlockedRecipeNames)
	}
	if len(loaded.LostCargo) != 1 || loaded.LostCargo[0].LifetimeTicks != 4000 {
		t.Errorf("LostCargo mismatch: %+v", loaded.LostCargo)
	}
}

func TestHasSaveFile(t *testing.T) {
	// Should return false for non-existent file path
	if _, err := os.Stat("non_existent_file_12345.json"); err == nil {
		t.Errorf("Expected file not to exist")
	}
}

func TestGetSlotPath(t *testing.T) {
	if got := GetSlotPath(1); got != "save_1.json" {
		t.Errorf("expected save_1.json, got %s", got)
	}
	if got := GetSlotPath(3); got != "save_3.json" {
		t.Errorf("expected save_3.json, got %s", got)
	}
}

func TestListSlotsAndDelete(t *testing.T) {
	tempDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	slots := ListSlots()
	if len(slots) != NumSlots {
		t.Fatalf("expected %d slots, got %d", NumSlots, len(slots))
	}
	for _, slot := range slots {
		if slot.Occupied {
			t.Errorf("expected slot %d to be empty", slot.Slot)
		}
	}

	data := &SaveData{WorldSeed: 4242, Timestamp: 1700000000}
	if err := SaveToFile(GetSlotPath(2), data); err != nil {
		t.Fatal(err)
	}

	slots = ListSlots()
	if !slots[1].Occupied {
		t.Fatal("expected slot 2 to be occupied")
	}
	if slots[1].WorldSeed != 4242 {
		t.Errorf("expected seed 4242, got %d", slots[1].WorldSeed)
	}

	if err := DeleteSlot(2); err != nil {
		t.Fatal(err)
	}
	slots = ListSlots()
	if slots[1].Occupied {
		t.Error("expected slot 2 to be empty after delete")
	}
}

func TestLegacySaveMigration(t *testing.T) {
	tempDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	if err := SaveToFile(DefaultSaveFileName, &SaveData{WorldSeed: 9999, Timestamp: 1}); err != nil {
		t.Fatal(err)
	}

	slots := ListSlots()
	if !slots[0].Occupied {
		t.Fatal("expected legacy save migrated into slot 1")
	}
	if slots[0].WorldSeed != 9999 {
		t.Errorf("expected migrated seed 9999, got %d", slots[0].WorldSeed)
	}
	if _, err := os.Stat(DefaultSaveFileName); !os.IsNotExist(err) {
		t.Error("expected legacy save.json to be renamed away")
	}
	if _, err := os.Stat(GetSlotPath(1)); err != nil {
		t.Error("expected save_1.json after migration")
	}
}
