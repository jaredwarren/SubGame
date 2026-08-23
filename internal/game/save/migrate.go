package save

import (
	"fmt"

	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

// CurrentSaveVersion is written on every new save. Load migrates older versions up.
const CurrentSaveVersion = 2

// MigrateSaveData upgrades data in-place to CurrentSaveVersion.
func MigrateSaveData(data *SaveData) error {
	if data == nil {
		return fmt.Errorf("nil save data")
	}
	if data.Version == 0 {
		data.Version = 1
	}
	for data.Version < CurrentSaveVersion {
		switch data.Version {
		case 1:
			if err := migrateV1toV2(data); err != nil {
				return err
			}
			data.Version = 2
		default:
			return fmt.Errorf("unsupported save version %d", data.Version)
		}
	}
	return nil
}

func migrateV1toV2(data *SaveData) error {
	migrateInventorySlots(&data.Player.Inventory)
	migrateInventorySlots(&data.Player.Upgrades)
	migrateInventorySlots(&data.Player.Hotbar)
	migrateInventorySlots(&data.BaseStation.Storage)
	migrateInventorySlots(&data.BaseStation.Upgrades)

	for i := range data.Vehicles {
		v := &data.Vehicles[i]
		migrateInventorySlots(&v.Cargo)
		migrateInventorySlots(&v.Upgrades)
		if v.ID == "" && v.Type != "" {
			id, ok := vehicle.VehicleIDFromName(v.Type)
			if !ok {
				return fmt.Errorf("unknown vehicle type in save: %q", v.Type)
			}
			v.ID = string(id)
		}
	}
	for i := range data.LostCargo {
		migrateInventorySlots(&data.LostCargo[i].Cargo)
	}
	return nil
}

func migrateInventorySlots(inv *item.SavedInventory) {
	if inv == nil {
		return
	}
	for i := range inv.Slots {
		slot := &inv.Slots[i]
		if slot.ItemID == "" && slot.ItemName != "" {
			if id, ok := item.ItemIDFromName(slot.ItemName); ok {
				slot.ItemID = id
			}
		}
		if slot.Upgrades != nil {
			migrateInventorySlots(slot.Upgrades)
		}
	}
}
