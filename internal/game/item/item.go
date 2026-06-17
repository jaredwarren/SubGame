package item

import (
	"image/color"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Item defines the interface that all inventory-compatible items must implement.
type Item interface {
	GetName() string
	GetMaxStack() int
	DrawIcon(screen *ebiten.Image, cx, cy, size float32)
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

// -----------------------------------------------------------------
// Static Registry Map and Metadata structure
// -----------------------------------------------------------------

type ItemMetadata struct {
	Name           string
	MaxStack       int
	Color          color.Color
	DrawIcon       func(screen *ebiten.Image, cx, cy, size float32)
	MaxO2Capacity  float64
	SpeedUpgrade   map[string]Speed
	ModuleType     BaseModule
	StorageSlots   int
	SolarRecharge  float64
	HealthRestore  float64
	StaminaRestore float64
	Use            func(ctx UsableContext) bool
}

var itemRegistry = make(map[reflect.Type]*ItemMetadata)

func register[T any](meta *ItemMetadata) {
	var zero T
	t := reflect.TypeOf(zero)
	itemRegistry[t] = meta
}

func getMeta(t reflect.Type) *ItemMetadata {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	meta, ok := itemRegistry[t]
	if !ok {
		panic("unregistered item type: " + t.String())
	}
	return meta
}

func init() {
	register[Titanium](&ItemMetadata{
		Name:     "Titanium",
		MaxStack: 10,
		Color:    color.RGBA{168, 178, 188, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{220, 230, 240, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{168, 178, 188, 255}, coreColor, "Titanium")
		},
	})
	register[Copper](&ItemMetadata{
		Name:     "Copper",
		MaxStack: 10,
		Color:    color.RGBA{218, 118, 48, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{240, 160, 80, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{218, 118, 48, 255}, coreColor, "Copper")
		},
	})
	register[Quartz](&ItemMetadata{
		Name:     "Quartz",
		MaxStack: 10,
		Color:    color.RGBA{48, 218, 245, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{220, 250, 255, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{48, 218, 245, 255}, coreColor, "Quartz")
		},
	})
	register[AbyssalOre](&ItemMetadata{
		Name:     "Abyssal Ore",
		MaxStack: 10,
		Color:    color.RGBA{148, 48, 218, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{230, 180, 255, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{148, 48, 218, 255}, coreColor, "Abyssal Ore")
		},
	})
	register[Nickel](&ItemMetadata{
		Name:     "Nickel",
		MaxStack: 10,
		Color:    color.RGBA{162, 175, 148, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{222, 235, 208, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{162, 175, 148, 255}, coreColor, "Nickel")
		},
	})
	register[ScrapMetal](&ItemMetadata{
		Name:     "Scrap Metal",
		MaxStack: 10,
		Color:    color.RGBA{140, 110, 95, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			if drawItemIconSprite(screen, "Scrap Metal", cx, cy, size) {
				return
			}
			var path vector.Path
			path.MoveTo(cx-size/3.0, cy-size/3.0)
			path.LineTo(cx+size/3.0, cy-size/4.0)
			path.LineTo(cx+size/4.0, cy+size/3.0)
			path.LineTo(cx-size/3.0, cy+size/4.0)
			path.Close()
			var opts vector.DrawPathOptions
			opts.ColorScale.ScaleWithColor(color.RGBA{140, 110, 95, 255})
			vector.FillPath(screen, &path, nil, &opts)
			vector.StrokeLine(screen, cx-size/3.0, cy, cx+size/3.0, cy-size/10.0, 1.5, color.RGBA{180, 150, 130, 255}, false)
		},
	})
	register[ElectronicWaste](&ItemMetadata{
		Name:     "Electronic Waste",
		MaxStack: 10,
		Color:    color.RGBA{70, 130, 90, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			if drawItemIconSprite(screen, "Electronic Waste", cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, color.RGBA{70, 130, 90, 255}, false)
			vector.StrokeRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, 1.0, color.RGBA{120, 200, 140, 255}, false)
			vector.FillRect(screen, cx-size/6.0, cy-size/6.0, size/3.0, size/3.0, color.RGBA{40, 40, 40, 255}, false)
			vector.FillRect(screen, cx-size/3.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
			vector.FillRect(screen, cx, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
			vector.FillRect(screen, cx+size/4.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
		},
	})
	register[RawFish](&ItemMetadata{
		Name:           "Raw Fish",
		MaxStack:       5,
		Color:          color.RGBA{70, 140, 180, 255},
		HealthRestore:  0.0,
		StaminaRestore: 5.0,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Raw Fish"
			clr := color.RGBA{70, 140, 180, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/3.5, clr, false)
			var path vector.Path
			path.MoveTo(cx-size/3.5, cy)
			path.LineTo(cx-size/1.8, cy-size/4.0)
			path.LineTo(cx-size/1.8, cy+size/4.0)
			path.Close()
			var opts vector.DrawPathOptions
			opts.ColorScale.ScaleWithColor(clr)
			vector.FillPath(screen, &path, nil, &opts)
			vector.FillCircle(screen, cx+size/6.0, cy-size/10.0, 2.0, color.White, false)
		},
	})
	register[CookedFish](&ItemMetadata{
		Name:           "Cooked Fish",
		MaxStack:       5,
		Color:          color.RGBA{170, 110, 60, 255},
		HealthRestore:  25.0,
		StaminaRestore: 15.0,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Cooked Fish"
			clr := color.RGBA{170, 110, 60, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/3.5, clr, false)
			var path vector.Path
			path.MoveTo(cx-size/3.5, cy)
			path.LineTo(cx-size/1.8, cy-size/4.0)
			path.LineTo(cx-size/1.8, cy+size/4.0)
			path.Close()
			var opts vector.DrawPathOptions
			opts.ColorScale.ScaleWithColor(clr)
			vector.FillPath(screen, &path, nil, &opts)
			vector.StrokeLine(screen, cx, cy-size/6.0, cx-size/6.0, cy+size/6.0, 1.5, color.RGBA{100, 60, 30, 255}, false)
			vector.StrokeLine(screen, cx+size/8.0, cy-size/6.0, cx-size/12.0, cy+size/6.0, 1.5, color.RGBA{100, 60, 30, 255}, false)
		},
	})
	register[RawCrab](&ItemMetadata{
		Name:           "Raw Crab",
		MaxStack:       5,
		Color:          color.RGBA{180, 50, 50, 255},
		HealthRestore:  0.0,
		StaminaRestore: 8.0,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Raw Crab"
			clr := color.RGBA{180, 50, 50, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/4.0, clr, false)
			vector.FillRect(screen, cx-size/2.5, cy-size/4.0, size/5.0, size/5.0, clr, false)
			vector.FillRect(screen, cx+size/2.5-size/5.0, cy-size/4.0, size/5.0, size/5.0, clr, false)
			vector.FillCircle(screen, cx-size/10.0, cy-size/4.0, 1.5, color.White, false)
			vector.FillCircle(screen, cx+size/10.0, cy-size/4.0, 1.5, color.White, false)
		},
	})
	register[CookedCrab](&ItemMetadata{
		Name:           "Cooked Crab",
		MaxStack:       5,
		Color:          color.RGBA{240, 90, 50, 255},
		HealthRestore:  20.0,
		StaminaRestore: 20.0,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Cooked Crab"
			clr := color.RGBA{240, 90, 50, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/4.0, clr, false)
			vector.FillRect(screen, cx-size/2.5, cy-size/4.0, size/5.0, size/5.0, clr, false)
			vector.FillRect(screen, cx+size/2.5-size/5.0, cy-size/4.0, size/5.0, size/5.0, clr, false)
			vector.FillCircle(screen, cx-size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
			vector.FillCircle(screen, cx+size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
		},
	})
	register[O2TankHC](&ItemMetadata{
		Name:          "High Capacity O2 Tank",
		MaxStack:      1,
		Color:         color.RGBA{98, 198, 148, 255},
		MaxO2Capacity: 60.0,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "High Capacity O2 Tank"
			clr := color.RGBA{98, 198, 148, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/2.0, clr, false)
		},
	})
	register[O2TankUHC](&ItemMetadata{
		Name:          "Ultra High Capacity O2 Tank",
		MaxStack:      1,
		Color:         color.RGBA{98, 198, 148, 255},
		MaxO2Capacity: 140.0,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Ultra High Capacity O2 Tank"
			clr := color.RGBA{98, 198, 148, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/2.0, clr, false)
		},
	})
	register[Fins](&ItemMetadata{
		Name:     "Propulsion Fins",
		MaxStack: 1,
		Color:    color.RGBA{98, 198, 148, 255},
		SpeedUpgrade: map[string]Speed{
			"overworld": {Drag: 0.92, Acceleration: 0.12, TopSpeed: 2.6},
			"cave":      {Drag: 0.96, Acceleration: 0.30, TopSpeed: 6.5},
		},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Propulsion Fins"
			clr := color.RGBA{98, 198, 148, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/2.0, clr, false)
		},
	})
	register[Scanner](&ItemMetadata{
		Name:     "Scanner Tool",
		MaxStack: 1,
		Color:    color.RGBA{98, 198, 148, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Scanner Tool"
			clr := color.RGBA{98, 198, 148, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/2.0, clr, false)
		},
	})
	register[UpgradeSolar](&ItemMetadata{
		Name:          "Solar Array Module",
		MaxStack:      1,
		Color:         color.RGBA{220, 200, 30, 255},
		ModuleType:    ModuleSolar,
		SolarRecharge: 0.08,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Solar Array Module"
			clr := color.RGBA{220, 200, 30, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, clr, false)
			vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, color.RGBA{255, 255, 255, 128}, false)
		},
	})
	register[UpgradeSolarMKII](&ItemMetadata{
		Name:          "Solar Array MKII Module",
		MaxStack:      1,
		Color:         color.RGBA{240, 220, 50, 255},
		ModuleType:    ModuleSolarMKII,
		SolarRecharge: 0.20,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Solar Array MKII Module"
			clr := color.RGBA{240, 220, 50, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, clr, false)
			vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 2.0, color.RGBA{255, 255, 255, 200}, false)
		},
	})
	register[UpgradeStorage](&ItemMetadata{
		Name:         "Storage Vault Module",
		MaxStack:     1,
		Color:        color.RGBA{130, 150, 180, 255},
		ModuleType:   ModuleStorage,
		StorageSlots: 24,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Storage Vault Module"
			clr := color.RGBA{130, 150, 180, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, clr, false)
			vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, color.RGBA{255, 255, 255, 128}, false)
		},
	})
	register[UpgradeStorageMKII](&ItemMetadata{
		Name:         "Storage Vault MKII Module",
		MaxStack:     1,
		Color:        color.RGBA{150, 180, 220, 255},
		ModuleType:   ModuleStorageMKII,
		StorageSlots: 48,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Storage Vault MKII Module"
			clr := color.RGBA{150, 180, 220, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, clr, false)
			vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 2.0, color.RGBA{255, 255, 255, 200}, false)
		},
	})
	register[DecoyLauncher](&ItemMetadata{
		Name:     "Decoy Launcher Module",
		MaxStack: 1,
		Color:    color.RGBA{110, 120, 130, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Decoy Launcher Module"
			clr := color.RGBA{110, 120, 130, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, clr, false)
			vector.StrokeRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, 1.5, color.RGBA{220, 220, 220, 255}, false)
			vector.FillCircle(screen, cx, cy-size/4.0, 3, color.RGBA{50, 240, 100, 255}, false)
		},
	})
	register[ChemicalDischarger](&ItemMetadata{
		Name:     "Chemical Discharger Module",
		MaxStack: 1,
		Color:    color.RGBA{130, 80, 180, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Chemical Discharger Module"
			clr := color.RGBA{130, 80, 180, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, clr, false)
			vector.FillRect(screen, cx-size/4.0, cy-size/1.8, size/6.0, size/4.0, color.RGBA{80, 80, 90, 255}, false)
			vector.FillRect(screen, cx+size/12.0, cy-size/1.8, size/6.0, size/4.0, color.RGBA{80, 80, 90, 255}, false)
		},
	})
	register[SonarAmplifier](&ItemMetadata{
		Name:     "Sonar Amplifier",
		MaxStack: 1,
		Color:    color.RGBA{0, 240, 255, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Sonar Amplifier"
			clr := color.RGBA{0, 240, 255, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.StrokeCircle(screen, cx, cy, size/2.0, 2.0, clr, false)
			vector.StrokeCircle(screen, cx, cy, size/3.5, 1.5, color.RGBA{255, 255, 255, 200}, false)
			vector.FillCircle(screen, cx, cy, 3, clr, false)
		},
	})
	register[PowerCell](&ItemMetadata{
		Name:     "Power Cell",
		MaxStack: 5,
		Color:    color.RGBA{220, 180, 40, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Power Cell"
			clr := color.RGBA{220, 180, 40, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/4.0, cy-size/3.0, size/2.0, size*0.7, clr, false)
			vector.FillRect(screen, cx-size/8.0, cy-size/2.0, size/4.0, size/6.0, color.RGBA{180, 190, 200, 255}, false)
		},
	})
	register[ThermalGenerator](&ItemMetadata{
		Name:     "Thermal Generator",
		MaxStack: 1,
		Color:    color.RGBA{235, 100, 50, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Thermal Generator"
			clr := color.RGBA{235, 100, 50, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.5, clr, false)
			vector.FillCircle(screen, cx, cy, size/4.0, color.RGBA{255, 120, 0, 255}, false)
		},
	})
	register[EscapeRocket](&ItemMetadata{
		Name:     "Escape Rocket",
		MaxStack: 1,
		Color:    color.RGBA{255, 100, 50, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Escape Rocket"
			clr := color.RGBA{255, 100, 50, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			topY := cy - size/2.0
			bottomY := cy + size/2.0
			leftX := cx - size/4.0
			rightX := cx + size/4.0
			midY := cy - size/6.0

			var path vector.Path
			path.MoveTo(cx, topY)
			path.LineTo(leftX, midY)
			path.LineTo(rightX, midY)
			path.Close()
			var opts vector.DrawPathOptions
			opts.ColorScale.ScaleWithColor(clr)
			vector.FillPath(screen, &path, nil, &opts)

			vector.FillRect(screen, leftX, midY, size/2.0, bottomY-midY, color.RGBA{220, 220, 220, 255}, false)

			var flamePath vector.Path
			flamePath.MoveTo(cx, bottomY+size/4.0)
			flamePath.LineTo(cx-size/6.0, bottomY)
			flamePath.LineTo(cx+size/6.0, bottomY)
			flamePath.Close()
			var flameOpts vector.DrawPathOptions
			flameOpts.ColorScale.ScaleWithColor(color.RGBA{255, 165, 0, 255})
			vector.FillPath(screen, &flamePath, nil, &flameOpts)
		},
	})
	register[SonicDecoy](&ItemMetadata{
		Name:     "Sonic Decoy",
		MaxStack: 5,
		Color:    color.RGBA{180, 210, 50, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Sonic Decoy"
			clr := color.RGBA{180, 210, 50, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/4.0, cy-size/3.0, size/2.0, size*0.7, clr, false)
			vector.StrokeCircle(screen, cx, cy, size/2.0, 1.5, color.RGBA{255, 255, 255, 180}, false)
			vector.FillCircle(screen, cx, cy, 3, color.White, false)
		},
		Use: func(ctx UsableContext) bool {
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
		},
	})
	register[ChemicalDeterrent](&ItemMetadata{
		Name:     "Chemical Deterrent",
		MaxStack: 5,
		Color:    color.RGBA{40, 25, 60, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Chemical Deterrent"
			clr := color.RGBA{40, 25, 60, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			vector.FillCircle(screen, cx, cy, size/3.0, clr, false)
			vector.FillRect(screen, cx-size/6.0, cy-size/2.0, size/3.0, size, clr, false)
			vector.FillRect(screen, cx-size/6.0, cy-size/8.0, size/3.0, size/4.0, color.RGBA{240, 110, 40, 255}, false)
		},
		Use: func(ctx UsableContext) bool {
			cursor := ctx.CursorWorldPos()
			ctx.SpawnDeterrentCloud(cursor)
			ctx.SetMineWarning("Chemical Deterrent Released!", 90, 1)
			return true
		},
	})
}

// -----------------------------------------------------------------
// Promoted Generic Nodes for Type-Safe interface matching
// -----------------------------------------------------------------

type BaseItem[T any] struct{}

func (b BaseItem[T]) GetName() string {
	var zero T
	return getMeta(reflect.TypeOf(zero)).Name
}

func (b BaseItem[T]) GetMaxStack() int {
	var zero T
	return getMeta(reflect.TypeOf(zero)).MaxStack
}

func (b BaseItem[T]) GetColor() color.Color {
	var zero T
	return getMeta(reflect.TypeOf(zero)).Color
}

func (b BaseItem[T]) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	var zero T
	getMeta(reflect.TypeOf(zero)).DrawIcon(screen, cx, cy, size)
}

type ConsumableNode[T any] struct {
	BaseItem[T]
}

func (c ConsumableNode[T]) GetHealthRestore() float64 {
	var zero T
	return getMeta(reflect.TypeOf(zero)).HealthRestore
}

func (c ConsumableNode[T]) GetStaminaRestore() float64 {
	var zero T
	return getMeta(reflect.TypeOf(zero)).StaminaRestore
}

type PlayerUpgradeNode[T any] struct {
	BaseItem[T]
}

func (p PlayerUpgradeNode[T]) IsPlayerUpgrade() bool {
	return true
}

type O2UpgradeNode[T any] struct {
	PlayerUpgradeNode[T]
}

func (o O2UpgradeNode[T]) GetMaxO2Capacity() float64 {
	var zero T
	return getMeta(reflect.TypeOf(zero)).MaxO2Capacity
}

type SpeedUpgradeNode[T any] struct {
	PlayerUpgradeNode[T]
}

func (s SpeedUpgradeNode[T]) GetSpeedUpgrade() map[string]Speed {
	var zero T
	return getMeta(reflect.TypeOf(zero)).SpeedUpgrade
}

type VehicleUpgradeNode[T any] struct {
	BaseItem[T]
}

func (v VehicleUpgradeNode[T]) IsVehicleUpgrade() bool {
	return true
}

type BaseUpgradeNode[T any] struct {
	BaseItem[T]
}

func (u BaseUpgradeNode[T]) GetModuleType() BaseModule {
	var zero T
	return getMeta(reflect.TypeOf(zero)).ModuleType
}

func (u BaseUpgradeNode[T]) GetStorageSlots() int {
	var zero T
	return getMeta(reflect.TypeOf(zero)).StorageSlots
}

func (u BaseUpgradeNode[T]) GetSolarRecharge() float64 {
	var zero T
	return getMeta(reflect.TypeOf(zero)).SolarRecharge
}

type UsableNode[T any] struct {
	BaseItem[T]
}

func (u UsableNode[T]) Use(ctx UsableContext) bool {
	var zero T
	return getMeta(reflect.TypeOf(zero)).Use(ctx)
}
