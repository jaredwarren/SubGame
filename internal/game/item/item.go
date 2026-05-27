package item

import (
	"image/color"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

// Mineral item types
type Titanium struct{}

func (t *Titanium) GetName() string       { return "Titanium" }
func (t *Titanium) GetMaxStack() int      { return 10 }
func (t *Titanium) GetColor() color.Color { return color.RGBA{168, 178, 188, 255} }
func (t *Titanium) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, t.GetColor(), false)
}

type Copper struct{}

func (c *Copper) GetName() string       { return "Copper" }
func (c *Copper) GetMaxStack() int      { return 10 }
func (c *Copper) GetColor() color.Color { return color.RGBA{218, 118, 48, 255} }
func (c *Copper) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, c.GetColor(), false)
}

type Quartz struct{}

func (q *Quartz) GetName() string       { return "Quartz" }
func (q *Quartz) GetMaxStack() int      { return 10 }
func (q *Quartz) GetColor() color.Color { return color.RGBA{48, 218, 245, 255} }
func (q *Quartz) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, q.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
}

type AbyssalOre struct{}

func (a *AbyssalOre) GetName() string       { return "Abyssal Ore" }
func (a *AbyssalOre) GetMaxStack() int      { return 10 }
func (a *AbyssalOre) GetColor() color.Color { return color.RGBA{148, 48, 218, 255} }
func (a *AbyssalOre) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, a.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
}

// Equipment item types
type O2TankHC struct{}

func (o *O2TankHC) GetName() string       { return "High Capacity O2 Tank" }
func (o *O2TankHC) GetMaxStack() int      { return 1 }
func (o *O2TankHC) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (o *O2TankHC) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, o.GetColor(), false)
}

type O2TankUHC struct{}

func (o *O2TankUHC) GetName() string       { return "Ultra High Capacity O2 Tank" }
func (o *O2TankUHC) GetMaxStack() int      { return 1 }
func (o *O2TankUHC) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (o *O2TankUHC) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, o.GetColor(), false)
}

type Fins struct{}

func (f *Fins) GetName() string       { return "Propulsion Fins" }
func (f *Fins) GetMaxStack() int      { return 1 }
func (f *Fins) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (f *Fins) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, f.GetColor(), false)
}

type Scanner struct{}

func (s *Scanner) GetName() string       { return "Scanner Tool" }
func (s *Scanner) GetMaxStack() int      { return 1 }
func (s *Scanner) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (s *Scanner) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, s.GetColor(), false)
}

// Deployable vehicle kit item types.
type ScoutSubKit struct{}

func (k *ScoutSubKit) GetName() string       { return "Scout Sub Kit" }
func (k *ScoutSubKit) GetMaxStack() int      { return 1 }
func (k *ScoutSubKit) GetColor() color.Color { return color.RGBA{15, 160, 185, 255} }
func (k *ScoutSubKit) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Small sub capsule silhouette
	vector.FillRect(screen, cx-size/2.0, cy-size/4.0, size, size/2.0, k.GetColor(), false)
	vector.FillCircle(screen, cx+size/4.0, cy, size/4.0, color.RGBA{80, 205, 255, 255}, false)
}

type HeavyMechKit struct{}

func (k *HeavyMechKit) GetName() string       { return "Heavy Mech Kit" }
func (k *HeavyMechKit) GetMaxStack() int      { return 1 }
func (k *HeavyMechKit) GetColor() color.Color { return color.RGBA{218, 98, 16, 255} }
func (k *HeavyMechKit) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Tiny mech torso silhouette
	vector.FillRect(screen, cx-size/3.0, cy-size/3.0, size/1.5, size/1.5, k.GetColor(), false)
	vector.FillRect(screen, cx-size/2.0, cy+size/6.0, size, size/6.0, color.RGBA{60, 70, 80, 255}, false)
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

type BaseModule int

const (
	ModuleFabricator BaseModule = iota
	ModuleStorage
	ModuleMedical
	ModuleSolar
	ModuleStorageMKII
	ModuleSolarMKII
)

func (m BaseModule) String() string {
	switch m {
	case ModuleFabricator:
		return "Fabricator Module"
	case ModuleStorage:
		return "Storage Vault"
	case ModuleStorageMKII:
		return "Storage Vault MKII"
	case ModuleMedical:
		return "Medical Bay"
	case ModuleSolar:
		return "Solar Array"
	case ModuleSolarMKII:
		return "Solar Array MKII"
	default:
		return "Module"
	}
}

type UpgradeItem interface {
	Item
	GetModuleType() BaseModule
	GetStorageSlots() int
	GetSolarRecharge() float64
}

// Base module upgrade items
type UpgradeSolar struct{}

func (u *UpgradeSolar) GetName() string       { return "Solar Array Module" }
func (u *UpgradeSolar) GetMaxStack() int      { return 1 }
func (u *UpgradeSolar) GetColor() color.Color { return color.RGBA{220, 200, 30, 255} }
func (u *UpgradeSolar) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, color.RGBA{255, 255, 255, 128}, false)
}
func (u *UpgradeSolar) GetModuleType() BaseModule { return ModuleSolar }
func (u *UpgradeSolar) GetStorageSlots() int      { return 0 }
func (u *UpgradeSolar) GetSolarRecharge() float64 { return 0.08 }

type UpgradeSolarMKII struct{}

func (u *UpgradeSolarMKII) GetName() string       { return "Solar Array MKII Module" }
func (u *UpgradeSolarMKII) GetMaxStack() int      { return 1 }
func (u *UpgradeSolarMKII) GetColor() color.Color { return color.RGBA{240, 220, 50, 255} }
func (u *UpgradeSolarMKII) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 2.0, color.RGBA{255, 255, 255, 200}, false)
}
func (u *UpgradeSolarMKII) GetModuleType() BaseModule { return ModuleSolarMKII }
func (u *UpgradeSolarMKII) GetStorageSlots() int      { return 0 }
func (u *UpgradeSolarMKII) GetSolarRecharge() float64 { return 0.20 }

type UpgradeStorage struct{}

func (u *UpgradeStorage) GetName() string       { return "Storage Vault Module" }
func (u *UpgradeStorage) GetMaxStack() int      { return 1 }
func (u *UpgradeStorage) GetColor() color.Color { return color.RGBA{130, 150, 180, 255} }
func (u *UpgradeStorage) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, color.RGBA{255, 255, 255, 128}, false)
}
func (u *UpgradeStorage) GetModuleType() BaseModule { return ModuleStorage }
func (u *UpgradeStorage) GetStorageSlots() int      { return 24 }
func (u *UpgradeStorage) GetSolarRecharge() float64 { return 0.0 }

type UpgradeStorageMKII struct{}

func (u *UpgradeStorageMKII) GetName() string       { return "Storage Vault MKII Module" }
func (u *UpgradeStorageMKII) GetMaxStack() int      { return 1 }
func (u *UpgradeStorageMKII) GetColor() color.Color { return color.RGBA{150, 180, 220, 255} }
func (u *UpgradeStorageMKII) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 2.0, color.RGBA{255, 255, 255, 200}, false)
}
func (u *UpgradeStorageMKII) GetModuleType() BaseModule { return ModuleStorageMKII }
func (u *UpgradeStorageMKII) GetStorageSlots() int      { return 48 }
func (u *UpgradeStorageMKII) GetSolarRecharge() float64 { return 0.0 }

