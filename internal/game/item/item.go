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

type O2UpgradeItem interface {
	PlayerUpgradeItem
	GetMaxO2Capacity() float64
}

// Equipment item types
type O2TankHC struct{}

func (o *O2TankHC) GetName() string       { return "High Capacity O2 Tank" }
func (o *O2TankHC) GetMaxStack() int      { return 1 }
func (o *O2TankHC) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (o *O2TankHC) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, o.GetColor(), false)
}
func (o *O2TankHC) IsPlayerUpgrade() bool     { return true }
func (o *O2TankHC) GetMaxO2Capacity() float64 { return 60.0 }

type O2TankUHC struct{}

func (o *O2TankUHC) GetName() string       { return "Ultra High Capacity O2 Tank" }
func (o *O2TankUHC) GetMaxStack() int      { return 1 }
func (o *O2TankUHC) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (o *O2TankUHC) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, o.GetColor(), false)
}
func (o *O2TankUHC) IsPlayerUpgrade() bool     { return true }
func (o *O2TankUHC) GetMaxO2Capacity() float64 { return 140.0 }

type SpeedUpgradeItem interface {
	PlayerUpgradeItem
	GetSpeedUpgrade() map[string]Speed
}

type Speed struct {
	Drag         float64
	Acceleration float64
	TopSpeed     float64
}

type Fins struct{}

func (f *Fins) GetName() string       { return "Propulsion Fins" }
func (f *Fins) GetMaxStack() int      { return 1 }
func (f *Fins) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (f *Fins) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, f.GetColor(), false)
}
func (f *Fins) IsPlayerUpgrade() bool { return true }

func (s *Fins) GetSpeedUpgrade() map[string]Speed {
	return map[string]Speed{
		"overworld": {
			Drag:         0.92,
			Acceleration: 0.12,
			TopSpeed:     2.6,
		},
		"cave": {
			Drag:         0.96,
			Acceleration: 0.30,
			TopSpeed:     6.5,
		},
	}
}

type Scanner struct{}

func (s *Scanner) GetName() string       { return "Scanner Tool" }
func (s *Scanner) GetMaxStack() int      { return 1 }
func (s *Scanner) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (s *Scanner) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	vector.FillCircle(screen, cx, cy, size/2.0, s.GetColor(), false)
}
func (s *Scanner) IsPlayerUpgrade() bool { return true }

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
func (k *ScoutSubKit) IsPlayerUpgrade() bool { return false }

type HeavyMechKit struct{}

func (k *HeavyMechKit) GetName() string       { return "Heavy Mech Kit" }
func (k *HeavyMechKit) GetMaxStack() int      { return 1 }
func (k *HeavyMechKit) GetColor() color.Color { return color.RGBA{218, 98, 16, 255} }
func (k *HeavyMechKit) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Tiny mech torso silhouette
	vector.FillRect(screen, cx-size/3.0, cy-size/3.0, size/1.5, size/1.5, k.GetColor(), false)
	vector.FillRect(screen, cx-size/2.0, cy+size/6.0, size, size/6.0, color.RGBA{60, 70, 80, 255}, false)
}
func (k *HeavyMechKit) IsPlayerUpgrade() bool { return false }

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

type EscapeRocket struct{}

func (e *EscapeRocket) GetName() string       { return "Escape Rocket" }
func (e *EscapeRocket) GetMaxStack() int      { return 1 }
func (e *EscapeRocket) GetColor() color.Color { return color.RGBA{255, 100, 50, 255} }
func (e *EscapeRocket) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	topY := cy - size/2.0
	bottomY := cy + size/2.0
	leftX := cx - size/4.0
	rightX := cx + size/4.0
	midY := cy - size/6.0

	// Nose cone (triangle)
	var path vector.Path
	path.MoveTo(cx, topY)
	path.LineTo(leftX, midY)
	path.LineTo(rightX, midY)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(e.GetColor())
	vector.FillPath(screen, &path, nil, &opts)

	// Body (rectangle)
	vector.FillRect(screen, leftX, midY, size/2.0, bottomY-midY, color.RGBA{220, 220, 220, 255}, false)

	// Thruster flame (orange triangle at bottom)
	var flamePath vector.Path
	flamePath.MoveTo(cx, bottomY+size/4.0)
	flamePath.LineTo(cx-size/6.0, bottomY)
	flamePath.LineTo(cx+size/6.0, bottomY)
	flamePath.Close()
	var flameOpts vector.DrawPathOptions
	flameOpts.ColorScale.ScaleWithColor(color.RGBA{255, 165, 0, 255})
	vector.FillPath(screen, &flamePath, nil, &flameOpts)
}

type VheicleUpgradeItem interface {
	Item
	IsVehicleUpgrade() bool
}

type SonarAmplifier struct{}

func (s *SonarAmplifier) GetName() string       { return "Sonar Amplifier" }
func (s *SonarAmplifier) GetMaxStack() int      { return 1 }
func (s *SonarAmplifier) GetColor() color.Color { return color.RGBA{0, 240, 255, 255} }
func (s *SonarAmplifier) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// A nice cyan/white concentric rings icon representing sonar amplification
	vector.StrokeCircle(screen, cx, cy, size/2.0, 2.0, s.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/3.5, 1.5, color.RGBA{255, 255, 255, 200}, false)
	vector.FillCircle(screen, cx, cy, 3, s.GetColor(), false)
}
func (s *SonarAmplifier) IsVehicleUpgrade() bool { return true }

type PowerCell struct{}

func (p *PowerCell) GetName() string       { return "Power Cell" }
func (p *PowerCell) GetMaxStack() int      { return 5 }
func (p *PowerCell) GetColor() color.Color { return color.RGBA{220, 180, 40, 255} }
func (p *PowerCell) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Draw a yellow cylinder battery cell with a light grey top tip
	vector.FillRect(screen, cx-size/4.0, cy-size/3.0, size/2.0, size*0.7, p.GetColor(), false)
	vector.FillRect(screen, cx-size/8.0, cy-size/2.0, size/4.0, size/6.0, color.RGBA{180, 190, 200, 255}, false)
}

type ThermalGenerator struct{}

func (t *ThermalGenerator) GetName() string       { return "Thermal Generator" }
func (t *ThermalGenerator) GetMaxStack() int      { return 1 }
func (t *ThermalGenerator) GetColor() color.Color { return color.RGBA{235, 100, 50, 255} }
func (t *ThermalGenerator) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Draw a diamond container with an inner orange flame/core
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.5, t.GetColor(), false)
	vector.FillCircle(screen, cx, cy, size/4.0, color.RGBA{255, 120, 0, 255}, false)
}
func (t *ThermalGenerator) IsVehicleUpgrade() bool { return true }

type ScrapMetal struct{}

func (s *ScrapMetal) GetName() string       { return "Scrap Metal" }
func (s *ScrapMetal) GetMaxStack() int      { return 10 }
func (s *ScrapMetal) GetColor() color.Color { return color.RGBA{140, 110, 95, 255} }
func (s *ScrapMetal) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Draw an angled metallic sheet
	var path vector.Path
	path.MoveTo(cx-size/3.0, cy-size/3.0)
	path.LineTo(cx+size/3.0, cy-size/4.0)
	path.LineTo(cx+size/4.0, cy+size/3.0)
	path.LineTo(cx-size/3.0, cy+size/4.0)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(s.GetColor())
	vector.FillPath(screen, &path, nil, &opts)
	vector.StrokeLine(screen, cx-size/3.0, cy, cx+size/3.0, cy-size/10.0, 1.5, color.RGBA{180, 150, 130, 255}, false)
}

type ElectronicWaste struct{}

func (e *ElectronicWaste) GetName() string       { return "Electronic Waste" }
func (e *ElectronicWaste) GetMaxStack() int      { return 10 }
func (e *ElectronicWaste) GetColor() color.Color { return color.RGBA{70, 130, 90, 255} }
func (e *ElectronicWaste) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Draw a green circuit board chip
	vector.FillRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, e.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, 1.0, color.RGBA{120, 200, 140, 255}, false)
	// Draw microchip core
	vector.FillRect(screen, cx-size/6.0, cy-size/6.0, size/3.0, size/3.0, color.RGBA{40, 40, 40, 255}, false)
	// Tiny copper pin details
	vector.FillRect(screen, cx-size/3.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
	vector.FillRect(screen, cx, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
	vector.FillRect(screen, cx+size/4.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
}

type RawFish struct{}

func (f *RawFish) GetName() string       { return "Raw Fish" }
func (f *RawFish) GetMaxStack() int      { return 5 }
func (f *RawFish) GetColor() color.Color { return color.RGBA{70, 140, 180, 255} }
func (f *RawFish) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Draw fish body (oval and tail)
	vector.FillCircle(screen, cx, cy, size/3.5, f.GetColor(), false)
	// Tail
	var path vector.Path
	path.MoveTo(cx-size/3.5, cy)
	path.LineTo(cx-size/1.8, cy-size/4.0)
	path.LineTo(cx-size/1.8, cy+size/4.0)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(f.GetColor())
	vector.FillPath(screen, &path, nil, &opts)
	// Eye
	vector.FillCircle(screen, cx+size/6.0, cy-size/10.0, 2.0, color.White, false)
}
func (f *RawFish) GetHealthRestore() float64  { return 0.0 }
func (f *RawFish) GetStaminaRestore() float64 { return 5.0 }

type CookedFish struct{}

func (f *CookedFish) GetName() string       { return "Cooked Fish" }
func (f *CookedFish) GetMaxStack() int      { return 5 }
func (f *CookedFish) GetColor() color.Color { return color.RGBA{170, 110, 60, 255} }
func (f *CookedFish) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Golden brown cooked fish
	vector.FillCircle(screen, cx, cy, size/3.5, f.GetColor(), false)
	var path vector.Path
	path.MoveTo(cx-size/3.5, cy)
	path.LineTo(cx-size/1.8, cy-size/4.0)
	path.LineTo(cx-size/1.8, cy+size/4.0)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(f.GetColor())
	vector.FillPath(screen, &path, nil, &opts)
	// Grill lines
	vector.StrokeLine(screen, cx, cy-size/6.0, cx-size/6.0, cy+size/6.0, 1.5, color.RGBA{100, 60, 30, 255}, false)
	vector.StrokeLine(screen, cx+size/8.0, cy-size/6.0, cx-size/12.0, cy+size/6.0, 1.5, color.RGBA{100, 60, 30, 255}, false)
}
func (f *CookedFish) GetHealthRestore() float64  { return 25.0 }
func (f *CookedFish) GetStaminaRestore() float64 { return 15.0 }

type RawCrab struct{}

func (c *RawCrab) GetName() string       { return "Raw Crab" }
func (c *RawCrab) GetMaxStack() int      { return 5 }
func (c *RawCrab) GetColor() color.Color { return color.RGBA{180, 50, 50, 255} }
func (c *RawCrab) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Crab body circle
	vector.FillCircle(screen, cx, cy, size/4.0, c.GetColor(), false)
	// Claws
	vector.FillRect(screen, cx-size/2.5, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillRect(screen, cx+size/2.5-size/5.0, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	// Little eyes
	vector.FillCircle(screen, cx-size/10.0, cy-size/4.0, 1.5, color.White, false)
	vector.FillCircle(screen, cx+size/10.0, cy-size/4.0, 1.5, color.White, false)
}
func (c *RawCrab) GetHealthRestore() float64  { return 0.0 }
func (c *RawCrab) GetStaminaRestore() float64 { return 8.0 }

type CookedCrab struct{}

func (c *CookedCrab) GetName() string       { return "Cooked Crab" }
func (c *CookedCrab) GetMaxStack() int      { return 5 }
func (c *CookedCrab) GetColor() color.Color { return color.RGBA{240, 90, 50, 255} }
func (c *CookedCrab) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Orange-red cooked crab
	vector.FillCircle(screen, cx, cy, size/4.0, c.GetColor(), false)
	vector.FillRect(screen, cx-size/2.5, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillRect(screen, cx+size/2.5-size/5.0, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillCircle(screen, cx-size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
	vector.FillCircle(screen, cx+size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
}
func (c *CookedCrab) GetHealthRestore() float64  { return 20.0 }
func (c *CookedCrab) GetStaminaRestore() float64 { return 20.0 }
