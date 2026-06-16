package item

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/assets"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

var (
	iconSprites map[string]*ebiten.Image
	iconsLoaded bool
)

// LoadAssets preloads and chroma-keys all item icon sprites.
func LoadAssets() {
	iconSprites = make(map[string]*ebiten.Image)

	sheet, err := assets.LoadChromaKeyedImage("item_icons")
	if err != nil {
		log.Printf("Error: Failed to load item icons: %v", err)
		return
	}

	bounds := sheet.Bounds()

	cellSize := 316
	startOffset := 76
	if bounds.Dx() != 2048 || bounds.Dy() != 2048 {
		cellSize = bounds.Dx() / 6
		startOffset = 0
	}

	itemCoords := map[string][2]int{
		"Titanium":                   {0, 0},
		"Copper":                     {1, 0},
		"Quartz":                     {2, 0},
		"Abyssal Ore":                {3, 0},
		"Scrap Metal":                {4, 0},
		"Electronic Waste":           {5, 0},
		"High Capacity O2 Tank":      {0, 1},
		"Ultra High Capacity O2 Tank":{1, 1},
		"Propulsion Fins":            {2, 1},
		"Scanner Tool":               {3, 1},
		"Solar Array Module":         {4, 1},
		"Solar Array MKII Module":    {5, 1},
		"Storage Vault Module":       {0, 2},
		"Storage Vault MKII Module":  {1, 2},
		"Scout Sub Kit":              {2, 2},
		"Heavy Mech Kit":             {3, 2},
		"Sonar Amplifier":            {4, 2},
		"Power Cell":                 {5, 2},
		"Thermal Generator":          {0, 4},
		"Escape Rocket":              {1, 4},
		"Raw Fish":                   {2, 4},
		"Cooked Fish":                {3, 4},
		"Raw Crab":                   {4, 4},
		"Cooked Crab":                {5, 4},
		"Sonic Decoy":                {0, 3},
		"Chemical Deterrent":         {1, 3},
		"Decoy Launcher Module":      {2, 3},
		"Chemical Discharger Module": {3, 3},
	}

	for name, coord := range itemCoords {
		col, row := coord[0], coord[1]
		x0 := startOffset + col*cellSize
		y0 := startOffset + row*cellSize
		x1 := x0 + cellSize
		y1 := y0 + cellSize

		if x1 <= bounds.Max.X && y1 <= bounds.Max.Y {
			sub := sheet.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
			iconSprites[name] = sub
		}
	}
	iconsLoaded = true
}

func drawItemIconSprite(screen *ebiten.Image, name string, cx, cy, size float32) bool {
	sprite, ok := iconSprites[name]
	if !ok || sprite == nil {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())

	op.GeoM.Translate(-spriteW/2.0, -spriteH/2.0)
	op.GeoM.Scale(float64(size)/spriteW, float64(size)/spriteH)
	op.GeoM.Translate(float64(cx), float64(cy))

	screen.DrawImage(sprite, op)
	return true
}

// DrawItemIconSprite wraps the internal drawItemIconSprite function so other packages can render item icons.
func DrawItemIconSprite(screen *ebiten.Image, name string, cx, cy, size float32) bool {
	return drawItemIconSprite(screen, name, cx, cy, size)
}

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
	coreColor := color.RGBA{220, 230, 240, 255}
	drawMineralIcon(screen, cx, cy, size, t.GetColor(), coreColor, "Titanium")
}

type Copper struct{}

func (c *Copper) GetName() string       { return "Copper" }
func (c *Copper) GetMaxStack() int      { return 10 }
func (c *Copper) GetColor() color.Color { return color.RGBA{218, 118, 48, 255} }
func (c *Copper) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	coreColor := color.RGBA{240, 160, 80, 255}
	drawMineralIcon(screen, cx, cy, size, c.GetColor(), coreColor, "Copper")
}

type Quartz struct{}

func (q *Quartz) GetName() string       { return "Quartz" }
func (q *Quartz) GetMaxStack() int      { return 10 }
func (q *Quartz) GetColor() color.Color { return color.RGBA{48, 218, 245, 255} }
func (q *Quartz) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	coreColor := color.RGBA{220, 250, 255, 255}
	drawMineralIcon(screen, cx, cy, size, q.GetColor(), coreColor, "Quartz")
}

type AbyssalOre struct{}

func (a *AbyssalOre) GetName() string       { return "Abyssal Ore" }
func (a *AbyssalOre) GetMaxStack() int      { return 10 }
func (a *AbyssalOre) GetColor() color.Color { return color.RGBA{148, 48, 218, 255} }
func (a *AbyssalOre) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	coreColor := color.RGBA{230, 180, 255, 255}
	drawMineralIcon(screen, cx, cy, size, a.GetColor(), coreColor, "Abyssal Ore")
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
	if drawItemIconSprite(screen, o.GetName(), cx, cy, size) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, o.GetColor(), false)
}
func (o *O2TankHC) IsPlayerUpgrade() bool     { return true }
func (o *O2TankHC) GetMaxO2Capacity() float64 { return 60.0 }

type O2TankUHC struct{}

func (o *O2TankUHC) GetName() string       { return "Ultra High Capacity O2 Tank" }
func (o *O2TankUHC) GetMaxStack() int      { return 1 }
func (o *O2TankUHC) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (o *O2TankUHC) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, o.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, f.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, s.GetName(), cx, cy, size) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, s.GetColor(), false)
}
func (s *Scanner) IsPlayerUpgrade() bool { return true }

type SonicDecoy struct{}

func (d *SonicDecoy) GetName() string       { return "Sonic Decoy" }
func (d *SonicDecoy) GetMaxStack() int      { return 5 }
func (d *SonicDecoy) GetColor() color.Color { return color.RGBA{180, 210, 50, 255} }
func (d *SonicDecoy) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, d.GetName(), cx, cy, size) {
		return
	}
	// Draw neon yellow-green cylinder/concentric rings fallback
	vector.FillRect(screen, cx-size/4.0, cy-size/3.0, size/2.0, size*0.7, d.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.5, color.RGBA{255, 255, 255, 180}, false)
	vector.FillCircle(screen, cx, cy, 3, color.White, false)
}
func (d *SonicDecoy) Use(ctx UsableContext) bool {
	playerCenter := gvec.Vec2{
		X: ctx.PlayerPos().X + ctx.PlayerDims().X/2.0,
		Y: ctx.PlayerPos().Y + ctx.PlayerDims().Y/2.0,
	}
	cursor := ctx.CursorWorldPos()
	dir := gvec.Vec2{X: cursor.X - playerCenter.X, Y: cursor.Y - playerCenter.Y}
	dist := dir.Length()
	if dist > 0 {
		dir = dir.Scale(1.0 / dist)
	} else {
		dir = gvec.Vec2{X: 1, Y: 0}
	}
	launchVel := dir.Scale(6.0)

	ctx.SpawnSonicDecoy(playerCenter, launchVel)
	ctx.SetMineWarning("Sonic Decoy Launched!", 90, 1)
	return true
}

type ChemicalDeterrent struct{}

func (c *ChemicalDeterrent) GetName() string       { return "Chemical Deterrent" }
func (c *ChemicalDeterrent) GetMaxStack() int      { return 5 }
func (c *ChemicalDeterrent) GetColor() color.Color { return color.RGBA{40, 25, 60, 255} }
func (c *ChemicalDeterrent) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, c.GetName(), cx, cy, size) {
		return
	}
	// Draw dark purple capsule with warning stripes
	vector.FillCircle(screen, cx, cy, size/3.0, c.GetColor(), false)
	vector.FillRect(screen, cx-size/6.0, cy-size/2.0, size/3.0, size, c.GetColor(), false)
	// Orange hazard stripe
	vector.FillRect(screen, cx-size/6.0, cy-size/8.0, size/3.0, size/4.0, color.RGBA{240, 110, 40, 255}, false)
}
func (c *ChemicalDeterrent) Use(ctx UsableContext) bool {
	cursor := ctx.CursorWorldPos()
	ctx.SpawnDeterrentCloud(cursor)
	ctx.SetMineWarning("Chemical Deterrent Released!", 90, 1)
	return true
}

type DecoyLauncher struct{}

func (l *DecoyLauncher) GetName() string       { return "Decoy Launcher Module" }
func (l *DecoyLauncher) GetMaxStack() int      { return 1 }
func (l *DecoyLauncher) GetColor() color.Color { return color.RGBA{110, 120, 130, 255} }
func (l *DecoyLauncher) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, l.GetName(), cx, cy, size) {
		return
	}
	// Draw tube launcher fallback
	vector.FillRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, l.GetColor(), false)
	vector.StrokeRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, 1.5, color.RGBA{220, 220, 220, 255}, false)
	// Green activation diode
	vector.FillCircle(screen, cx, cy-size/4.0, 3, color.RGBA{50, 240, 100, 255}, false)
}
func (l *DecoyLauncher) IsVehicleUpgrade() bool { return true }

type ChemicalDischarger struct{}

func (d *ChemicalDischarger) GetName() string       { return "Chemical Discharger Module" }
func (d *ChemicalDischarger) GetMaxStack() int      { return 1 }
func (d *ChemicalDischarger) GetColor() color.Color { return color.RGBA{130, 80, 180, 255} }
func (d *ChemicalDischarger) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, d.GetName(), cx, cy, size) {
		return
	}
	// Draw double nozzle/violet canister fallback
	vector.FillRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, d.GetColor(), false)
	// Nozzles
	vector.FillRect(screen, cx-size/4.0, cy-size/1.8, size/6.0, size/4.0, color.RGBA{80, 80, 90, 255}, false)
	vector.FillRect(screen, cx+size/12.0, cy-size/1.8, size/6.0, size/4.0, color.RGBA{80, 80, 90, 255}, false)
}
func (d *ChemicalDischarger) IsVehicleUpgrade() bool { return true }



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
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, e.GetName(), cx, cy, size) {
		return
	}
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

type VehicleUpgradeItem interface {
	Item
	IsVehicleUpgrade() bool
}

type SonarAmplifier struct{}

func (s *SonarAmplifier) GetName() string       { return "Sonar Amplifier" }
func (s *SonarAmplifier) GetMaxStack() int      { return 1 }
func (s *SonarAmplifier) GetColor() color.Color { return color.RGBA{0, 240, 255, 255} }
func (s *SonarAmplifier) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, s.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, p.GetName(), cx, cy, size) {
		return
	}
	// Draw a yellow cylinder battery cell with a light grey top tip
	vector.FillRect(screen, cx-size/4.0, cy-size/3.0, size/2.0, size*0.7, p.GetColor(), false)
	vector.FillRect(screen, cx-size/8.0, cy-size/2.0, size/4.0, size/6.0, color.RGBA{180, 190, 200, 255}, false)
}

type ThermalGenerator struct{}

func (t *ThermalGenerator) GetName() string       { return "Thermal Generator" }
func (t *ThermalGenerator) GetMaxStack() int      { return 1 }
func (t *ThermalGenerator) GetColor() color.Color { return color.RGBA{235, 100, 50, 255} }
func (t *ThermalGenerator) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, t.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, s.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, e.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, f.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, f.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, c.GetName(), cx, cy, size) {
		return
	}
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
	if drawItemIconSprite(screen, c.GetName(), cx, cy, size) {
		return
	}
	// Orange-red cooked crab
	vector.FillCircle(screen, cx, cy, size/4.0, c.GetColor(), false)
	vector.FillRect(screen, cx-size/2.5, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillRect(screen, cx+size/2.5-size/5.0, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillCircle(screen, cx-size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
	vector.FillCircle(screen, cx+size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
}
func (c *CookedCrab) GetHealthRestore() float64  { return 20.0 }
func (c *CookedCrab) GetStaminaRestore() float64 { return 20.0 }

var gPath vector.Path

func darkenColor(c color.Color, factor float32) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(float32(r>>8) * factor),
		G: uint8(float32(g>>8) * factor),
		B: uint8(float32(b>>8) * factor),
		A: uint8(a >> 8),
	}
}

func blendColor(c1, c2 color.Color, t float32) color.RGBA {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return color.RGBA{
		R: uint8((1.0-t)*float32(r1>>8) + t*float32(r2>>8)),
		G: uint8((1.0-t)*float32(g1>>8) + t*float32(g2>>8)),
		B: uint8((1.0-t)*float32(b1>>8) + t*float32(b2>>8)),
		A: uint8((1.0-t)*float32(a1>>8) + t*float32(a2>>8)),
	}
}

func rotateVec(v [2]float32, angle float32) [2]float32 {
	cosA := float32(math.Cos(float64(angle)))
	sinA := float32(math.Sin(float64(angle)))
	return [2]float32{
		v[0]*cosA - v[1]*sinA,
		v[0]*sinA + v[1]*cosA,
	}
}

func localToScreen(cx, cy float32, lx, ly float32, dirVec, perpVec [2]float32) (float32, float32) {
	return cx + lx*perpVec[0] + ly*dirVec[0], cy + lx*perpVec[1] + ly*dirVec[1]
}

func drawShard(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, length, width float32, shadowColor, highlightColor color.Color) {
	// Left Face
	gPath.Reset()
	lx0, ly0 := localToScreen(cx, cy, 0, 0, dirVec, perpVec)
	lx1, ly1 := localToScreen(cx, cy, -width/2, 0, dirVec, perpVec)
	lx2, ly2 := localToScreen(cx, cy, -width/2, length*0.4, dirVec, perpVec)
	lx3, ly3 := localToScreen(cx, cy, 0, length, dirVec, perpVec)
	lx4, ly4 := localToScreen(cx, cy, 0, length*0.45, dirVec, perpVec)

	gPath.MoveTo(lx0, ly0)
	gPath.LineTo(lx1, ly1)
	gPath.LineTo(lx2, ly2)
	gPath.LineTo(lx3, ly3)
	gPath.LineTo(lx4, ly4)
	gPath.Close()
	var shadowOpts vector.DrawPathOptions
	shadowOpts.ColorScale.ScaleWithColor(shadowColor)
	vector.FillPath(screen, &gPath, nil, &shadowOpts)

	// Right Face
	gPath.Reset()
	rx0, ry0 := localToScreen(cx, cy, 0, 0, dirVec, perpVec)
	rx1, ry1 := localToScreen(cx, cy, 0, length*0.45, dirVec, perpVec)
	rx2, ry2 := localToScreen(cx, cy, 0, length, dirVec, perpVec)
	rx3, ry3 := localToScreen(cx, cy, width/2, length*0.4, dirVec, perpVec)
	rx4, ry4 := localToScreen(cx, cy, width/2, 0, dirVec, perpVec)

	gPath.MoveTo(rx0, ry0)
	gPath.LineTo(rx1, ry1)
	gPath.LineTo(rx2, ry2)
	gPath.LineTo(rx3, ry3)
	gPath.LineTo(rx4, ry4)
	gPath.Close()
	var highlightOpts vector.DrawPathOptions
	highlightOpts.ColorScale.ScaleWithColor(highlightColor)
	vector.FillPath(screen, &gPath, nil, &highlightOpts)
}

func drawCrystalCluster(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color, isSpiky bool) {
	baseLength := float32(28.0) * scale
	baseWidth := float32(11.0) * scale
	if isSpiky {
		baseLength = float32(34.0) * scale
		baseWidth = float32(7.0) * scale
	}

	leftDir := rotateVec(dirVec, -0.42)
	leftPerp := rotateVec(perpVec, -0.42)
	drawShard(screen, cx, cy, leftDir, leftPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	rightDir := rotateVec(dirVec, 0.42)
	rightPerp := rotateVec(perpVec, 0.42)
	drawShard(screen, cx, cy, rightDir, rightPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	drawShard(screen, cx, cy, dirVec, perpVec, baseLength, baseWidth, shadowColor, highlightColor)
}

func drawNodule(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, mineralColor, coreColor color.Color) {
	R := float32(12.0) * scale
	if R < 2.0 {
		R = 2.0
	}

	type bump struct {
		lx, ly float32
		r      float32
	}

	bumps := []bump{
		{-R * 0.45, R * 0.3, R * 0.75},
		{R * 0.45, R * 0.3, R * 0.75},
		{0, R * 0.5, R},
		{0, R * 0.85, R * 0.6},
	}

	for _, b := range bumps {
		bx, by := localToScreen(cx, cy, b.lx, b.ly, dirVec, perpVec)
		vector.FillCircle(screen, bx, by, b.r, darkenColor(mineralColor, 0.9), false)
		vector.StrokeCircle(screen, bx, by, b.r, 1.0, darkenColor(mineralColor, 0.5), false)

		hx, hy := localToScreen(cx, cy, b.lx-b.r*0.25, b.ly+b.r*0.25, dirVec, perpVec)
		hr := b.r * 0.28
		if hr < 1.0 {
			hr = 1.0
		}
		vector.FillCircle(screen, hx, hy, hr, blendColor(coreColor, color.White, 0.6), false)
	}
}

func drawQuartzNeedles(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color) {
	baseLength := float32(36.0) * scale
	baseWidth := float32(4.5) * scale

	angles := []float32{-0.5, -0.18, 0.2, 0.55}
	lengths := []float32{0.7, 1.0, 0.85, 0.65}

	for i, angle := range angles {
		d := rotateVec(dirVec, angle)
		p := rotateVec(perpVec, angle)
		drawShard(screen, cx, cy, d, p, baseLength*lengths[i], baseWidth, shadowColor, highlightColor)
	}
}

func drawMineralIcon(screen *ebiten.Image, cx, cy, size float32, mineralColor, coreColor color.Color, mineralName string) {
	scale := size / 40.0
	if scale < 0.2 {
		scale = 0.2
	}

	dirVec := [2]float32{0, -1}
	perpVec := [2]float32{1, 0}

	shadowColor := darkenColor(mineralColor, 0.82)
	highlightColor := blendColor(mineralColor, coreColor, 0.65)

	switch mineralName {
	case "Copper":
		drawNodule(screen, cx, cy, dirVec, perpVec, scale, mineralColor, coreColor)
	case "Quartz":
		drawQuartzNeedles(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor)
	case "Abyssal Ore":
		drawSpikyCrystal := true
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, drawSpikyCrystal)
	default:
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, false)
	}
}
