package item

import "reflect"

// Inventory manages a collection of item slots and caches counts for O(1) queries.
type Inventory struct {
	Slots  []ItemStack
	counts map[reflect.Type]int
}

// NewInventory creates an empty Inventory of a specific size.
func NewInventory(size int) *Inventory {
	slots := make([]ItemStack, size)
	return &Inventory{
		Slots:  slots,
		counts: make(map[reflect.Type]int),
	}
}

// HasItem is a generic query helper that does O(1) map checking (Option B).
func HasItem[T any](inv *Inventory, qty int) bool {
	if inv == nil || inv.counts == nil {
		return false
	}
	var zero T
	return inv.counts[reflect.TypeOf(zero)] >= qty
}

// AddItem inserts an item into the inventory, stack-splitting if necessary.
// Returns true if all items were successfully added.
func (inv *Inventory) AddItem(item Item, qty int) bool {
	if item == nil || qty <= 0 {
		return false
	}
	t := reflect.TypeOf(item)
	maxStack := item.GetMaxStack()

	// 1. First, attempt to fill existing stacks of this item
	for i := range inv.Slots {
		if inv.Slots[i].Item != nil && reflect.TypeOf(inv.Slots[i].Item) == t && inv.Slots[i].Quantity < maxStack {
			space := maxStack - inv.Slots[i].Quantity
			if qty <= space {
				inv.Slots[i].Quantity += qty
				inv.counts[t] += qty
				return true
			}
			inv.Slots[i].Quantity = maxStack
			inv.counts[t] += space
			qty -= space
		}
	}

	// 2. Next, place remaining items in empty slots
	for qty > 0 {
		emptyIdx := -1
		for i := range inv.Slots {
			if inv.Slots[i].Item == nil {
				emptyIdx = i
				break
			}
		}

		// Inventory is full
		if emptyIdx == -1 {
			return false
		}

		addQty := qty
		if addQty > maxStack {
			addQty = maxStack
		}

		inv.Slots[emptyIdx] = ItemStack{Item: item, Quantity: addQty}
		inv.counts[t] += addQty
		qty -= addQty
	}

	return true
}

// RemoveItem consumes a specific quantity of an item type from the inventory.
// Returns true if the items were successfully removed.
func (inv *Inventory) RemoveItem(t reflect.Type, qty int) bool {
	if inv == nil || !inv.HasItem(t, qty) {
		return false
	}
	originalQty := qty

	// Remove from stacks starting from the end
	for i := len(inv.Slots) - 1; i >= 0; i-- {
		if inv.Slots[i].Item != nil && reflect.TypeOf(inv.Slots[i].Item) == t {
			if inv.Slots[i].Quantity >= qty {
				inv.Slots[i].Quantity -= qty
				if inv.Slots[i].Quantity == 0 {
					inv.Slots[i].Item = nil
				}
				break
			} else {
				qty -= inv.Slots[i].Quantity
				inv.Slots[i].Item = nil
				inv.Slots[i].Quantity = 0
			}
		}
	}

	inv.counts[t] -= originalQty
	return true
}

// HasItem checks if the inventory contains at least the specified quantity of an item type.
func (inv *Inventory) HasItem(t reflect.Type, qty int) bool {
	if inv == nil || inv.counts == nil {
		return false
	}
	return inv.counts[t] >= qty
}

// CountOf returns the current quantity for a given item type.
func (inv *Inventory) CountOf(t reflect.Type) int {
	if inv == nil || inv.counts == nil {
		return 0
	}
	return inv.counts[t]
}
