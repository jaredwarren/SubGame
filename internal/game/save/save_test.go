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
