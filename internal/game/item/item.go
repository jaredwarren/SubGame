package item

import (
	"image/color"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Item defines the interface that all inventory-compatible items must implement.
type Item interface {
	GetName() string
	GetMaxStack() int
	// DrawIcon renders this item's icon centered at (cx, cy) with the given size.
	DrawIcon(screen *ebiten.Image, cx, cy, size float32)
	// GetColor returns the primary display color for this item (used in inventory grids).
	GetColor() color.Color
}

// UsableItem is an item that can be actively used by the player from their hand/hotbar.
type UsableItem interface {
	Item
	Use(ctx UsableContext) bool
}

// UsableContext provides localized state queries and side effects for item usage,
// avoiding cyclic imports with the entity or scene packages.
type UsableContext interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
	CursorWorldPos() gvec.Vec2
	SpawnSonicDecoy(pos gvec.Vec2, vel gvec.Vec2)
	SpawnDeterrentCloud(pos gvec.Vec2)
	SetMineWarning(msg string, duration, level int)
}

// PlayerUpgradeItem is an item that acts as a passive upgrade for the player character.
type PlayerUpgradeItem interface {
	Item
	IsPlayerUpgrade() bool
}

// BaseItemProvider allows items (like resource nodes) to define their base item type dynamically.
type BaseItemProvider interface {
	GetBaseItem() Item
}

// Consumable defines items that can be consumed from inventory for health/stamina effects.
type Consumable interface {
	Item
	GetHealthRestore() float64
	GetStaminaRestore() float64
}

// O2UpgradeItem defines upgrades that increase the player's oxygen capacity.
type O2UpgradeItem interface {
	PlayerUpgradeItem
	GetMaxO2Capacity() float64
}

// SpeedUpgradeItem defines upgrades that adjust player movement speeds.
type SpeedUpgradeItem interface {
	PlayerUpgradeItem
	GetSpeedUpgrade() map[string]Speed
}

// Speed holds drag, acceleration, and top speed scalars.
type Speed struct {
	Drag         float64
	Acceleration float64
	TopSpeed     float64
}

// VehicleUpgradeItem is an item that can be installed on vehicles as an upgrade module.
type VehicleUpgradeItem interface {
	Item
	IsVehicleUpgrade() bool
}

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

// BaseModule defines identifiers for base upgrade modules.
type BaseModule int

const (
	ModuleFabricator BaseModule = iota
	ModuleStorage
	ModuleMedical
	ModuleSolar
	ModuleStorageMKII
	ModuleSolarMKII
)

// UpgradeItem defines interfaces for base station upgrade modules.
type UpgradeItem interface {
	Item
	GetModuleType() BaseModule
	GetStorageSlots() int
	GetSolarRecharge() float64
}
