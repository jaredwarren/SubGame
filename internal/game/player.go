package game

import (
	"math"
)

// Player represents the player character, including their physics and stats.
type Player struct {
	// Physics
	X, Y     float64
	Vx, Vy   float64
	Width    float64
	Height   float64
	Facing   float64 // Angle in radians (for flashlight/boat direction)

	// Stats
	MaxHealth      float64
	CurrentHealth  float64
	MaxOxygen      float64
	CurrentOxygen  float64
	MaxStamina     float64
	CurrentStamina float64
	MaxEnergy      float64
	CurrentEnergy  float64

	// Inventory
	Inventory *Inventory
}

// NewPlayer initializes a player with default stats and empty inventory.
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:              x,
		Y:              y,
		Width:          20,
		Height:         20,
		MaxHealth:      100,
		CurrentHealth:  100,
		MaxOxygen:      100, // 100 seconds of O2 initially
		CurrentOxygen:  100,
		MaxStamina:     100,
		CurrentStamina: 100,
		MaxEnergy:      100,
		CurrentEnergy:  100,
		Inventory:      NewInventory(24),
	}
}

// UpdateStats handles core stat loops (depleting/regenerating O2, stamina, etc.)
func (p *Player) UpdateStats(inCave bool, isSprinting bool) {
	// Oxygen management
	if inCave {
		p.CurrentOxygen -= 1.0 / 60.0 // Drain 1 O2 per second (at 60 FPS)
		if p.CurrentOxygen < 0 {
			p.CurrentOxygen = 0
			p.CurrentHealth -= 0.5 // Drowning damage
		}
	} else {
		// Instantly refill or quickly refill oxygen on surface
		p.CurrentOxygen = p.MaxOxygen
	}

	// Stamina management
	if isSprinting && (math.Abs(p.Vx) > 0.1 || math.Abs(p.Vy) > 0.1) {
		p.CurrentStamina -= 1.5 / 60.0 // Sprinting drains stamina
		if p.CurrentStamina < 0 {
			p.CurrentStamina = 0
		}
	} else {
		p.CurrentStamina += 1.0 / 60.0 // Regenerate stamina when not sprinting
		if p.CurrentStamina > p.MaxStamina {
			p.CurrentStamina = p.MaxStamina
		}
	}

	// Health clamp
	if p.CurrentHealth < 0 {
		p.CurrentHealth = 0
	}
	if p.CurrentHealth > p.MaxHealth {
		p.CurrentHealth = p.MaxHealth
	}
}

// pCenterX returns the screen X position where the player is drawn (centered).
func pCenterX(p *Player) float64 {
	return ScreenWidth / 2
}

// pCenterY returns the screen Y position where the player is drawn (centered).
func pCenterY(p *Player) float64 {
	return ScreenHeight / 2
}

// ItemType represents different types of collectible items and equipment.
type ItemType int

const (
	ItemNone ItemType = iota
	ItemTitanium
	ItemCopper
	ItemQuartz
	ItemAbyssalOre
	ItemO2TankHC
	ItemO2TankUHC
	ItemFins
	ItemScanner
)

// String returns the user-facing name of the item type.
func (t ItemType) String() string {
	switch t {
	case ItemTitanium:
		return "Titanium"
	case ItemCopper:
		return "Copper"
	case ItemQuartz:
		return "Quartz"
	case ItemAbyssalOre:
		return "Abyssal Ore"
	case ItemO2TankHC:
		return "High Capacity O2 Tank"
	case ItemO2TankUHC:
		return "Ultra High Capacity O2 Tank"
	case ItemFins:
		return "Propulsion Fins"
	case ItemScanner:
		return "Scanner Tool"
	default:
		return "Empty"
	}
}

// ItemStack represents a quantity of a specific item type.
type ItemStack struct {
	Type     ItemType
	Quantity int
}

// Inventory manages a collection of item slots.
type Inventory struct {
	Slots []ItemStack
}

// NewInventory creates an empty Inventory of a specific size.
func NewInventory(size int) *Inventory {
	slots := make([]ItemStack, size)
	for i := range slots {
		slots[i] = ItemStack{Type: ItemNone, Quantity: 0}
	}
	return &Inventory{Slots: slots}
}

// AddItem inserts an item type into the inventory, stack-splitting if necessary.
// Returns true if all items were successfully added.
func (inv *Inventory) AddItem(itemType ItemType, qty int) bool {
	if itemType == ItemNone || qty <= 0 {
		return false
	}

	maxStack := 10
	// Equipment items are non-stackable (max stack of 1)
	if itemType == ItemO2TankHC || itemType == ItemO2TankUHC || itemType == ItemFins || itemType == ItemScanner {
		maxStack = 1
	}

	// 1. First, attempt to fill existing stacks of this item
	for i := range inv.Slots {
		if inv.Slots[i].Type == itemType && inv.Slots[i].Quantity < maxStack {
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
			if inv.Slots[i].Type == ItemNone {
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

		inv.Slots[emptyIdx] = ItemStack{Type: itemType, Quantity: addQty}
		qty -= addQty
	}

	return true
}

// RemoveItem consumes a specific quantity of an item type from the inventory.
// Returns true if the items were successfully removed.
func (inv *Inventory) RemoveItem(itemType ItemType, qty int) bool {
	if !inv.HasItem(itemType, qty) {
		return false
	}

	// Remove from stacks starting from the end
	for i := len(inv.Slots) - 1; i >= 0; i-- {
		if inv.Slots[i].Type == itemType {
			if inv.Slots[i].Quantity >= qty {
				inv.Slots[i].Quantity -= qty
				if inv.Slots[i].Quantity == 0 {
					inv.Slots[i].Type = ItemNone
				}
				break
			} else {
				qty -= inv.Slots[i].Quantity
				inv.Slots[i].Type = ItemNone
				inv.Slots[i].Quantity = 0
			}
		}
	}

	return true
}

// HasItem checks if the inventory contains at least the specified quantity of an item type.
func (inv *Inventory) HasItem(itemType ItemType, qty int) bool {
	count := 0
	for _, slot := range inv.Slots {
		if slot.Type == itemType {
			count += slot.Quantity
		}
	}
	return count >= qty
}

