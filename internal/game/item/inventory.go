package item

import "reflect"

// Inventory manages a collection of item slots.
type Inventory struct {
	Slots []ItemStack
}

// NewInventory creates an empty Inventory of a specific size.
func NewInventory(size int) *Inventory {
	slots := make([]ItemStack, size)
	return &Inventory{
		Slots: slots,
	}
}

// HasItem is a generic query helper.
func HasItem[T any](inv *Inventory, qty int) bool {
	if inv == nil {
		return false
	}
	var zero T
	return inv.HasItem(reflect.TypeOf(zero), qty)
}

// AddItem inserts an item into the inventory, stack-splitting if necessary.
// Returns true if all items were successfully added.
func (inv *Inventory) AddItem(item Item, qty int) bool {
	if item == nil || qty <= 0 {
		return false
	}

	if provider, ok := item.(BaseItemProvider); ok {
		base := provider.GetBaseItem()
		if base == nil {
			return false
		}
		item = base
	}

	t := reflect.TypeOf(item)
	if t == nil {
		return false
	}

	maxStack := item.GetMaxStack()

	// Pre-compute capacity: count available space in existing stacks of type t,
	// plus maxStack space for each empty slot.
	availableSpace := 0
	for i := range inv.Slots {
		if inv.Slots[i].Item == nil {
			availableSpace += maxStack
		} else if reflect.TypeOf(inv.Slots[i].Item) == t {
			availableSpace += maxStack - inv.Slots[i].Quantity
		}
	}
	if availableSpace < qty {
		return false
	}

	// 1. First, attempt to fill existing stacks of this item
	for i := range inv.Slots {
		if inv.Slots[i].Item != nil && reflect.TypeOf(inv.Slots[i].Item) == t && inv.Slots[i].Quantity < maxStack {
			space := maxStack - inv.Slots[i].Quantity
			if qty <= space {
				inv.Slots[i].Quantity += qty
				return true
			}
			inv.Slots[i].Quantity = maxStack
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
		qty -= addQty
	}

	return true
}

// RemoveItem consumes a specific quantity of an item type from the inventory.
// Returns true if the items were successfully removed.
func (inv *Inventory) RemoveItem(t reflect.Type, qty int) bool {
	if inv == nil || t == nil || !inv.HasItem(t, qty) {
		return false
	}

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

	return true
}

// HasItem checks if the inventory contains at least the specified quantity of an item type.
func (inv *Inventory) HasItem(t reflect.Type, qty int) bool {
	if inv == nil || t == nil {
		return false
	}
	return inv.CountOf(t) >= qty
}

// CountOf returns the current quantity for a given item type.
func (inv *Inventory) CountOf(t reflect.Type) int {
	if inv == nil || t == nil {
		return 0
	}
	sum := 0
	for _, slot := range inv.Slots {
		if slot.Item != nil && reflect.TypeOf(slot.Item) == t {
			sum += slot.Quantity
		}
	}
	return sum
}

// Has reports whether this inventory contains at least qty of the item type.
func (inv *Inventory) Has(it Item, qty int) bool {
	if it == nil {
		return false
	}
	if provider, ok := it.(BaseItemProvider); ok {
		if base := provider.GetBaseItem(); base != nil {
			it = base
		}
	}
	return inv.HasItem(reflect.TypeOf(it), qty)
}

// Remove consumes qty items, prioritizing the exact instance it if present in a slot.
func (inv *Inventory) Remove(it Item, qty int) bool {
	if it == nil || inv == nil {
		return false
	}
	if provider, ok := it.(BaseItemProvider); ok {
		if base := provider.GetBaseItem(); base != nil {
			it = base
		}
	}
	for i := len(inv.Slots) - 1; i >= 0; i-- {
		if inv.Slots[i].Item == it {
			if inv.Slots[i].Quantity >= qty {
				inv.Slots[i].Quantity -= qty
				if inv.Slots[i].Quantity == 0 {
					inv.Slots[i].Item = nil
				}
				return true
			}
			qty -= inv.Slots[i].Quantity
			inv.Slots[i].Item = nil
			inv.Slots[i].Quantity = 0
			break
		}
	}
	return inv.RemoveItem(reflect.TypeOf(it), qty)
}

// Count returns how many items of the same type as it are present.
func (inv *Inventory) Count(it Item) int {
	if it == nil {
		return 0
	}
	if provider, ok := it.(BaseItemProvider); ok {
		if base := provider.GetBaseItem(); base != nil {
			it = base
		}
	}
	return inv.CountOf(reflect.TypeOf(it))
}

// Resize changes the size of the inventory slots slice, keeping existing items.
// If shrinking, it attempts to compact non-empty stacks from the truncated region into surviving slots.
// Returns the stacks (ItemStacks) that could not be preserved.
func (inv *Inventory) Resize(newSize int) []ItemStack {
	if inv == nil {
		return nil
	}
	if newSize == len(inv.Slots) {
		return nil
	}

	var lost []ItemStack

	// If shrinking, try to compact items from the truncated slots (newSize to end)
	// into the surviving slots (0 to newSize-1).
	if newSize < len(inv.Slots) {
		for i := newSize; i < len(inv.Slots); i++ {
			slot := inv.Slots[i]
			if slot.Item == nil || slot.Quantity <= 0 {
				continue
			}

			t := reflect.TypeOf(slot.Item)
			maxStack := slot.Item.GetMaxStack()
			qtyToAdd := slot.Quantity

			// Try to top up existing stacks of this item in the surviving range
			for j := 0; j < newSize; j++ {
				if inv.Slots[j].Item != nil && reflect.TypeOf(inv.Slots[j].Item) == t && inv.Slots[j].Quantity < maxStack {
					space := maxStack - inv.Slots[j].Quantity
					if qtyToAdd <= space {
						inv.Slots[j].Quantity += qtyToAdd
						qtyToAdd = 0
						break
					} else {
						inv.Slots[j].Quantity = maxStack
						qtyToAdd -= space
					}
				}
			}

			// If still have items left, try to place them in any empty slots in the surviving range
			if qtyToAdd > 0 {
				for j := 0; j < newSize; j++ {
					if inv.Slots[j].Item == nil {
						add := qtyToAdd
						if add > maxStack {
							add = maxStack
						}
						inv.Slots[j] = ItemStack{Item: slot.Item, Quantity: add}
						qtyToAdd -= add
						if qtyToAdd <= 0 {
							break
						}
					}
				}
			}

			// Any remaining quantity that couldn't be placed in the surviving range is lost
			if qtyToAdd > 0 {
				lost = append(lost, ItemStack{Item: slot.Item, Quantity: qtyToAdd})
			}
		}
	}

	newSlots := make([]ItemStack, newSize)
	copy(newSlots, inv.Slots)
	inv.Slots = newSlots

	return lost
}

// Clone returns a deep copy of the Inventory.
func (inv *Inventory) Clone() *Inventory {
	if inv == nil {
		return nil
	}
	clone := &Inventory{
		Slots: make([]ItemStack, len(inv.Slots)),
	}
	copy(clone.Slots, inv.Slots)
	return clone
}

// WouldResizeLoseItems returns true if resizing to newSize would result in any lost items.
func (inv *Inventory) WouldResizeLoseItems(newSize int) bool {
	if inv == nil || newSize >= len(inv.Slots) {
		return false
	}
	clone := inv.Clone()
	lost := clone.Resize(newSize)
	return len(lost) > 0
}

// Clear empties all slots in the inventory.
func (inv *Inventory) Clear() {
	if inv == nil {
		return
	}
	for i := range inv.Slots {
		inv.Slots[i] = ItemStack{}
	}
}

// IsEmpty returns true if the inventory is nil or contains no items in any slots.
func (inv *Inventory) IsEmpty() bool {
	if inv == nil {
		return true
	}
	for _, slot := range inv.Slots {
		if slot.Item != nil {
			return false
		}
	}
	return true
}

// ExtractAll removes every stack from the inventory and returns them.
func (inv *Inventory) ExtractAll() []ItemStack {
	if inv == nil {
		return nil
	}
	var stacks []ItemStack
	for i := range inv.Slots {
		if inv.Slots[i].Item != nil && inv.Slots[i].Quantity > 0 {
			stacks = append(stacks, inv.Slots[i])
			inv.Slots[i] = ItemStack{}
		}
	}
	return stacks
}

// InsertStacks places items from stacks into available inventory slots,
// merging with existing matching stacks and filling empty slots.
// Returns any leftover item stacks that did not fit.
func (inv *Inventory) InsertStacks(stacks []ItemStack) []ItemStack {
	if inv == nil || len(stacks) == 0 {
		return stacks
	}
	var leftover []ItemStack
	for _, stack := range stacks {
		if stack.Item == nil || stack.Quantity <= 0 {
			continue
		}
		it := stack.Item
		if provider, ok := it.(BaseItemProvider); ok {
			if base := provider.GetBaseItem(); base != nil {
				it = base
			}
		}
		remaining := stack.Quantity
		// Prefer stacking onto existing stacks of the same type.
		t := reflect.TypeOf(it)
		maxStack := it.GetMaxStack()
		for i := range inv.Slots {
			if remaining <= 0 {
				break
			}
			if inv.Slots[i].Item == nil {
				continue
			}
			if reflect.TypeOf(inv.Slots[i].Item) != t {
				continue
			}
			space := maxStack - inv.Slots[i].Quantity
			if space <= 0 {
				continue
			}
			add := remaining
			if add > space {
				add = space
			}
			inv.Slots[i].Quantity += add
			remaining -= add
		}
		for remaining > 0 {
			emptyIdx := -1
			for i := range inv.Slots {
				if inv.Slots[i].Item == nil {
					emptyIdx = i
					break
				}
			}
			if emptyIdx < 0 {
				leftover = append(leftover, ItemStack{Item: stack.Item, Quantity: remaining})
				break
			}
			add := remaining
			if add > maxStack {
				add = maxStack
			}
			inv.Slots[emptyIdx] = ItemStack{Item: stack.Item, Quantity: add}
			remaining -= add
		}
	}
	return leftover
}

// NewInventoryFromStacks builds an inventory large enough for the given stacks (one slot each).
func NewInventoryFromStacks(stacks []ItemStack) *Inventory {
	if len(stacks) == 0 {
		return NewInventory(1)
	}
	inv := NewInventory(len(stacks))
	for i, s := range stacks {
		inv.Slots[i] = s
	}
	return inv
}

// TotalItemCount returns the sum of all stack quantities.
func (inv *Inventory) TotalItemCount() int {
	if inv == nil {
		return 0
	}
	n := 0
	for _, s := range inv.Slots {
		if s.Item != nil {
			n += s.Quantity
		}
	}
	return n
}

// SavedItemStack represents serialized stack data for JSON save/load.
type SavedItemStack struct {
	ItemID   ItemID          `json:"itemId,omitempty"`
	ItemName string          `json:"itemName,omitempty"` // legacy v1; kept for migration/debug
	Quantity int             `json:"quantity"`
	Upgrades *SavedInventory `json:"upgrades,omitempty"`
	Health   float64         `json:"health,omitempty"`
	Battery  float64         `json:"battery,omitempty"`
	HasState bool            `json:"hasState,omitempty"`
}

// SavedInventory represents a serialized snapshot of an Inventory.
type SavedInventory struct {
	Size  int              `json:"size"`
	Slots []SavedItemStack `json:"slots"`
}

// SerializeState converts an Inventory into a JSON-ready SavedInventory snapshot.
func (inv *Inventory) SerializeState() SavedInventory {
	if inv == nil {
		return SavedInventory{}
	}
	saved := SavedInventory{
		Size:  len(inv.Slots),
		Slots: make([]SavedItemStack, len(inv.Slots)),
	}
	for i, slot := range inv.Slots {
		if slot.Item != nil {
			stack := SavedItemStack{
				ItemID:   slot.Item.GetID(),
				ItemName: slot.Item.GetName(),
				Quantity: slot.Quantity,
			}
			if stateful, ok := slot.Item.(StatefulItem); ok {
				upgs, health, battery, hasState := stateful.GetItemState()
				if hasState {
					stack.Upgrades = upgs
					stack.Health = health
					stack.Battery = battery
					stack.HasState = true
				}
			}
			saved.Slots[i] = stack
		}
	}
	return saved
}

// DeserializeInventory restores an Inventory from a SavedInventory snapshot.
func DeserializeInventory(saved SavedInventory) *Inventory {
	size := saved.Size
	if size <= 0 {
		size = len(saved.Slots)
	}
	if size <= 0 {
		size = 1
	}
	inv := NewInventory(size)
	for i, slotData := range saved.Slots {
		if i >= len(inv.Slots) {
			break
		}
		if slotData.Quantity <= 0 {
			continue
		}
		var it Item
		if slotData.ItemID != "" {
			it = NewItemByID(slotData.ItemID)
		}
		if it == nil && slotData.ItemName != "" {
			it = NewItemByName(slotData.ItemName)
		}
		if it == nil {
			continue
		}
		if stateful, ok := it.(StatefulItem); ok && slotData.HasState {
			stateful.SetItemState(slotData.Upgrades, slotData.Health, slotData.Battery, true)
		}
		inv.Slots[i] = ItemStack{
			Item:     it,
			Quantity: slotData.Quantity,
		}
	}
	return inv
}
