package item

import "reflect"

// Item defines the interface that all inventory-compatible items must implement.
type Item interface {
	GetName() string
	GetMaxStack() int
}

// Mineral item types
type Titanium struct{}

func (t *Titanium) GetName() string  { return "Titanium" }
func (t *Titanium) GetMaxStack() int { return 10 }

type Copper struct{}

func (c *Copper) GetName() string  { return "Copper" }
func (c *Copper) GetMaxStack() int { return 10 }

type Quartz struct{}

func (q *Quartz) GetName() string  { return "Quartz" }
func (q *Quartz) GetMaxStack() int { return 10 }

type AbyssalOre struct{}

func (a *AbyssalOre) GetName() string  { return "Abyssal Ore" }
func (a *AbyssalOre) GetMaxStack() int { return 10 }

// Equipment item types
type O2TankHC struct{}

func (o *O2TankHC) GetName() string  { return "High Capacity O2 Tank" }
func (o *O2TankHC) GetMaxStack() int { return 1 }

type O2TankUHC struct{}

func (o *O2TankUHC) GetName() string  { return "Ultra High Capacity O2 Tank" }
func (o *O2TankUHC) GetMaxStack() int { return 1 }

type Fins struct{}

func (f *Fins) GetName() string  { return "Propulsion Fins" }
func (f *Fins) GetMaxStack() int { return 1 }

type Scanner struct{}

func (s *Scanner) GetName() string  { return "Scanner Tool" }
func (s *Scanner) GetMaxStack() int { return 1 }

// Deployable vehicle kit item types.
type ScoutSubKit struct{}

func (k *ScoutSubKit) GetName() string  { return "Scout Sub Kit" }
func (k *ScoutSubKit) GetMaxStack() int { return 1 }

type HeavyMechKit struct{}

func (k *HeavyMechKit) GetName() string  { return "Heavy Mech Kit" }
func (k *HeavyMechKit) GetMaxStack() int { return 1 }

// NewItemFromType instantiates a new concrete Item struct using reflect.New.
func NewItemFromType(t reflect.Type) Item {
	return reflect.New(t.Elem()).Interface().(Item)
}

// Clone returns a new instance of the same concrete item type.
func Clone(it Item) Item {
	if it == nil {
		return nil
	}
	return NewItemFromType(reflect.TypeOf(it))
}

// ItemStack represents a quantity of a specific item type.
type ItemStack struct {
	Item     Item
	Quantity int
}
