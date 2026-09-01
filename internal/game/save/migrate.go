package save

import (
	"fmt"

	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

// CurrentSaveVersion is written on every new save. Load migrates older versions up.
const CurrentSaveVersion = 3

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
		case 2:
			if err := migrateV2toV3(data); err != nil {
				return err
			}
			data.Version = 3
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

func migrateV2toV3(data *SaveData) error {
	var skiffIdx = -1
	for i, v := range data.Vehicles {
		if v.ID == string(vehicle.VehicleSkiff) || v.Type == "The Skiff" || v.Type == "Skiff" {
			skiffIdx = i
			break
		}
	}

	var legacyKits []struct {
		id       vehicle.VehicleID
		upgrades *item.SavedInventory
		health   float64
		battery  float64
	}

	extractKit := func(inv *item.SavedInventory) {
		if inv == nil {
			return
		}
		for i := range inv.Slots {
			slot := &inv.Slots[i]
			switch slot.ItemID {
			case item.IDSkiffKit:
				if skiffIdx == -1 {
					data.Vehicles = append(data.Vehicles, SavedVehicle{
						ID:         string(vehicle.VehicleSkiff),
						Type:       "The Skiff",
						PosX:       data.Player.PosX,
						PosY:       data.Player.PosY,
						Health:     slot.Health,
						MaxHealth:  150,
						Battery:    slot.Battery,
						MaxBattery: 100,
						Upgrades:   safeDerefInventory(slot.Upgrades),
						Location:   "overworld",
					})
					skiffIdx = len(data.Vehicles) - 1
				}
				slot.ItemID = ""
				slot.ItemName = ""
				slot.Quantity = 0
				slot.Upgrades = nil
			case item.IDScoutSubKit:
				legacyKits = append(legacyKits, struct {
					id       vehicle.VehicleID
					upgrades *item.SavedInventory
					health   float64
					battery  float64
				}{
					id:       vehicle.VehicleScoutSub,
					upgrades: slot.Upgrades,
					health:   slot.Health,
					battery:  slot.Battery,
				})
				slot.ItemID = ""
				slot.ItemName = ""
				slot.Quantity = 0
				slot.Upgrades = nil
			case item.IDHeavyMechKit:
				legacyKits = append(legacyKits, struct {
					id       vehicle.VehicleID
					upgrades *item.SavedInventory
					health   float64
					battery  float64
				}{
					id:       vehicle.VehicleHeavyMech,
					upgrades: slot.Upgrades,
					health:   slot.Health,
					battery:  slot.Battery,
				})
				slot.ItemID = ""
				slot.ItemName = ""
				slot.Quantity = 0
				slot.Upgrades = nil
			}
		}
	}

	extractKit(&data.Player.Inventory)
	extractKit(&data.Player.Hotbar)
	extractKit(&data.BaseStation.Storage)

	if len(legacyKits) > 0 {
		if skiffIdx == -1 {
			data.Vehicles = append(data.Vehicles, SavedVehicle{
				ID:         string(vehicle.VehicleSkiff),
				Type:       "The Skiff",
				PosX:       data.Player.PosX,
				PosY:       data.Player.PosY,
				Health:     150,
				MaxHealth:  150,
				Battery:    100,
				MaxBattery: 100,
				Location:   "overworld",
			})
			skiffIdx = len(data.Vehicles) - 1
		}
		skiff := &data.Vehicles[skiffIdx]
		for _, kit := range legacyKits {
			bayIdx := 0
			if kit.id == vehicle.VehicleHeavyMech {
				bayIdx = 1
			}
			maxHp := 100.0
			if kit.id == vehicle.VehicleHeavyMech {
				maxHp = 200.0
			}
			hp := kit.health
			if hp <= 0 {
				hp = maxHp
			}
			bat := kit.battery
			if bat <= 0 {
				bat = 100.0
			}
			skiff.Docked = append(skiff.Docked, SavedDockedVehicle{
				BayIdx:     bayIdx,
				ID:         string(kit.id),
				Health:     hp,
				MaxHealth:  maxHp,
				Battery:    bat,
				MaxBattery: 100,
				Upgrades:   safeDerefInventory(kit.upgrades),
			})
		}
	}

	return nil
}

func safeDerefInventory(inv *item.SavedInventory) item.SavedInventory {
	if inv == nil {
		return item.SavedInventory{}
	}
	return *inv
}
