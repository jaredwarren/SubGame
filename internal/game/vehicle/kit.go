package vehicle

import "github.com/jaredwarren/SubGame/internal/game/item"

// CloneInventory deep-copies an inventory, or returns nil when src is nil.
func CloneInventory(src *item.Inventory) *item.Inventory {
	if src == nil {
		return nil
	}
	dst := item.NewInventory(len(src.Slots))
	for i, slot := range src.Slots {
		if slot.Item != nil {
			dst.Slots[i] = item.ItemStack{
				Item:     item.Clone(slot.Item),
				Quantity: slot.Quantity,
			}
		}
	}
	return dst
}

// KitVehicleState holds deployable kit payload shared by scout sub and heavy mech kits.
type KitVehicleState struct {
	Upgrades *item.Inventory
	Health   float64
	Battery  float64
	HasState bool
}

// RestoreKitState applies saved kit state onto a freshly spawned vehicle.
func RestoreKitState(health, battery *float64, upgrades **item.Inventory, state KitVehicleState) {
	if !state.HasState {
		return
	}
	if state.Upgrades != nil && upgrades != nil {
		*upgrades = CloneInventory(state.Upgrades)
	}
	if state.Health > 0 {
		*health = state.Health
	}
	if state.Battery >= 0 {
		*battery = state.Battery
	}
}
