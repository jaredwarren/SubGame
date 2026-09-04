package vehicle

import (
	"github.com/jaredwarren/SubGame/internal/game/item"
)

// DockedVehicle holds full state for a submersible docked on the Skiff.
type DockedVehicle struct {
	ID         VehicleID
	Health     float64
	MaxHealth  float64
	Battery    float64
	MaxBattery float64
	Cargo      *item.Inventory
	Upgrades   *item.Inventory
}

// NewDockedVehicleFromVehicle creates a DockedVehicle snapshot from an active Vehicle instance.
func NewDockedVehicleFromVehicle(v Vehicle) *DockedVehicle {
	if v == nil {
		return nil
	}
	return &DockedVehicle{
		ID:         v.GetID(),
		Health:     v.GetHealth(),
		MaxHealth:  v.GetMaxHealth(),
		Battery:    v.GetBattery(),
		MaxBattery: v.GetMaxBattery(),
		Cargo:      CloneInventory(v.GetCargo()),
		Upgrades:   CloneInventory(v.GetUpgrades()),
	}
}

// NewDefaultDockedVehicle creates a freshly fabricated docked vehicle at full stats.
func NewDefaultDockedVehicle(id VehicleID) *DockedVehicle {
	v, ok := NewVehicleByID(id, 0, 0)
	if !ok || v == nil {
		return nil
	}
	return &DockedVehicle{
		ID:         id,
		Health:     v.GetHealth(),
		MaxHealth:  v.GetMaxHealth(),
		Battery:    v.GetBattery(),
		MaxBattery: v.GetMaxBattery(),
		Cargo:      CloneInventory(v.GetCargo()),
		Upgrades:   CloneInventory(v.GetUpgrades()),
	}
}

// ToVehicle instantiates an active Vehicle from this docked state.
func (d *DockedVehicle) ToVehicle(x, y float64) Vehicle {
	if d == nil {
		return nil
	}
	v, ok := NewVehicleByID(d.ID, x, y)
	if !ok || v == nil {
		return nil
	}
	switch sub := v.(type) {
	case *ScoutSub:
		if d.Health > 0 {
			sub.Health = d.Health
		}
		if d.Battery >= 0 {
			sub.Battery = d.Battery
		}
		if d.Cargo != nil {
			sub.Cargo = CloneInventory(d.Cargo)
		}
		if d.Upgrades != nil {
			sub.Upgrades = CloneInventory(d.Upgrades)
		}
	case *HeavyMech:
		if d.Health > 0 {
			sub.Health = d.Health
		}
		if d.Battery >= 0 {
			sub.Battery = d.Battery
		}
		if d.Cargo != nil {
			sub.Cargo = CloneInventory(d.Cargo)
		}
		if d.Upgrades != nil {
			sub.Upgrades = CloneInventory(d.Upgrades)
		}
	case *MiniLifepod:
		if d.Health > 0 {
			sub.Health = d.Health
		}
		if d.Battery >= 0 {
			sub.Battery = d.Battery
		}
		if d.Cargo != nil {
			sub.Cargo = CloneInventory(d.Cargo)
		}
		if d.Upgrades != nil {
			sub.Upgrades = CloneInventory(d.Upgrades)
		}
		sub.RecalculateProperties()
	default:
		if d.Cargo != nil && v.GetCargo() != nil {
			*v.GetCargo() = *CloneInventory(d.Cargo)
		}
		if d.Upgrades != nil && v.GetUpgrades() != nil {
			*v.GetUpgrades() = *CloneInventory(d.Upgrades)
		}
	}
	return v
}

// GetName returns the human-readable display name for this docked vehicle.
func (d *DockedVehicle) GetName() string {
	if d == nil {
		return ""
	}
	switch d.ID {
	case VehicleScoutSub:
		return "Scout Sub"
	case VehicleHeavyMech:
		return "Heavy Mech"
	case VehicleMiniLifepod:
		return "Mini-Lifepod"
	default:
		return string(d.ID)
	}
}

// Clone creates a deep copy of the docked vehicle state.
func (d *DockedVehicle) Clone() *DockedVehicle {
	if d == nil {
		return nil
	}
	return &DockedVehicle{
		ID:         d.ID,
		Health:     d.Health,
		MaxHealth:  d.MaxHealth,
		Battery:    d.Battery,
		MaxBattery: d.MaxBattery,
		Cargo:      CloneInventory(d.Cargo),
		Upgrades:   CloneInventory(d.Upgrades),
	}
}
