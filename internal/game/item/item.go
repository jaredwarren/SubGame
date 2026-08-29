package item

import (
	"image/color"
	"reflect"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Item defines the interface that all inventory-compatible items must implement.
type Item interface {
	GetID() ItemID
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
var nameToType = make(map[string]reflect.Type)
var idToType = make(map[ItemID]reflect.Type)

// externalItemFactories lets other packages (e.g. vehicle kits) register
// deserialization factories without importing those packages from item.
var externalItemFactories = make(map[ItemID]func() Item)

func initItemRegistryLookup() {
	for t, meta := range itemRegistry {
		structType := t
		if t.Kind() == reflect.Pointer {
			structType = t.Elem()
		}
		if meta != nil && meta.Name != "" {
			nameToType[meta.Name] = structType
			if meta.ID == "" {
				if id, ok := ItemIDFromName(meta.Name); ok {
					meta.ID = id
				}
			}
			if meta.ID != "" {
				idToType[meta.ID] = structType
			}
		}
		nameToType[structType.Name()] = structType
		if meta != nil && meta.ID == "" {
			if id, ok := ItemIDFromName(structType.Name()); ok {
				meta.ID = id
				idToType[meta.ID] = structType
			}
		}
	}
}

// RegisterItemByName registers a factory for an item display/type name used by save/load.
// Safe to call from package init; later registrations overwrite earlier ones for the same name.
func RegisterItemByName(name string, factory func() Item) {
	if name == "" || factory == nil {
		return
	}
	id, ok := ItemIDFromName(name)
	if !ok {
		id = ItemID(strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	}
	externalItemFactories[id] = factory
	displayNameToID[name] = id
}

func NewItemFromType(t reflect.Type) Item {
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return reflect.New(t).Interface().(Item)
}

// NewItemByID instantiates an item from its stable ItemID.
func NewItemByID(id ItemID) Item {
	if id == "" {
		return nil
	}
	if len(idToType) == 0 {
		initItemRegistryLookup()
	}
	if t, ok := idToType[id]; ok {
		return NewItemFromType(t)
	}
	if factory, ok := externalItemFactories[id]; ok {
		return factory()
	}
	return nil
}

// NewItemByName instantiates a new concrete Item using the item's display name or type name.
func NewItemByName(name string) Item {
	if id, ok := ItemIDFromName(name); ok {
		if it := NewItemByID(id); it != nil {
			return it
		}
	}
	if len(nameToType) == 0 {
		initItemRegistryLookup()
	}
	if t, ok := nameToType[name]; ok {
		return NewItemFromType(t)
	}
	return nil
}

// Cloner defines an item that can create a deep copy of itself (including internal state).
type Cloner interface {
	Clone() Item
}

// StatefulItem is an item that saves and restores custom state (e.g., vehicle upgrades/health).
type StatefulItem interface {
	Item
	GetItemState() (upgrades *SavedInventory, health float64, battery float64, hasState bool)
	SetItemState(upgrades *SavedInventory, health float64, battery float64, hasState bool)
}

// Clone returns a new instance of the item, preserving internal state if Cloner is implemented.
func Clone(it Item) Item {
	if it == nil {
		return nil
	}
	if cloner, ok := it.(Cloner); ok {
		return cloner.Clone()
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
	ID             ItemID
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

// itemRegistry is a compile-time static lookup of item metadata keyed by concrete type.
var itemRegistry = map[reflect.Type]*ItemMetadata{
	reflect.TypeFor[Titanium]():   mineralMetadata(MaterialTitanium),
	reflect.TypeFor[Copper]():     mineralMetadata(MaterialCopper),
	reflect.TypeFor[Quartz]():     mineralMetadata(MaterialQuartz),
	reflect.TypeFor[AbyssalOre](): mineralMetadata(MaterialAbyssalOre),
	reflect.TypeFor[Nickel]():     mineralMetadata(MaterialNickel),
	reflect.TypeFor[Tungsten]():   mineralMetadata(MaterialTungsten),
	reflect.TypeFor[ScrapMetal](): {
		Name:     MaterialScrapMetal.Name,
		MaxStack: MaterialScrapMetal.MaxStack,
		Color:    MaterialScrapMetal.Color,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			m := MaterialScrapMetal
			if drawItemIconSprite(screen, m.Name, cx, cy, size) {
				return
			}
			var path vector.Path
			path.MoveTo(cx-size/3.0, cy-size/3.0)
			path.LineTo(cx+size/3.0, cy-size/4.0)
			path.LineTo(cx+size/4.0, cy+size/3.0)
			path.LineTo(cx-size/3.0, cy+size/4.0)
			path.Close()
			var opts vector.DrawPathOptions
			opts.ColorScale.ScaleWithColor(m.Color)
			vector.FillPath(screen, &path, nil, &opts)
			vector.StrokeLine(screen, cx-size/3.0, cy, cx+size/3.0, cy-size/10.0, 1.5, color.RGBA{180, 150, 130, 255}, false)
		},
	},
	reflect.TypeFor[ElectronicWaste](): {
		Name:     MaterialElectronicWaste.Name,
		MaxStack: MaterialElectronicWaste.MaxStack,
		Color:    MaterialElectronicWaste.Color,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			m := MaterialElectronicWaste
			if drawItemIconSprite(screen, m.Name, cx, cy, size) {
				return
			}
			vector.FillRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, m.Color, false)
			vector.StrokeRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, 1.0, color.RGBA{120, 200, 140, 255}, false)
			vector.FillRect(screen, cx-size/6.0, cy-size/6.0, size/3.0, size/3.0, color.RGBA{40, 40, 40, 255}, false)
			vector.FillRect(screen, cx-size/3.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
			vector.FillRect(screen, cx, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
			vector.FillRect(screen, cx+size/4.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
		},
	},
	reflect.TypeFor[RawFish](): {
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
	},
	reflect.TypeFor[CookedFish](): {
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
	},
	reflect.TypeFor[RawCrab](): {
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
	},
	reflect.TypeFor[CookedCrab](): {
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
	},
	reflect.TypeFor[O2TankHC](): {
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
	},
	reflect.TypeFor[O2TankUHC](): {
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
	},
	reflect.TypeFor[Fins](): {
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
	},
	reflect.TypeFor[Scanner](): {
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
	},
	reflect.TypeFor[Flashlight](): {
		Name:     "Flashlight",
		MaxStack: 1,
		Color:    color.RGBA{240, 220, 90, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Flashlight"
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			// Main handle / casing
			vector.FillRect(screen, cx-size*0.35, cy-size*0.12, size*0.45, size*0.24, color.RGBA{45, 60, 75, 255}, false)
			vector.StrokeRect(screen, cx-size*0.35, cy-size*0.12, size*0.45, size*0.24, 1.0, color.RGBA{80, 110, 135, 255}, false)
			// Grip ribs
			vector.FillRect(screen, cx-size*0.25, cy-size*0.14, size*0.06, size*0.28, color.RGBA{30, 40, 50, 255}, false)
			vector.FillRect(screen, cx-size*0.12, cy-size*0.14, size*0.06, size*0.28, color.RGBA{30, 40, 50, 255}, false)
			vector.FillRect(screen, cx+size*0.01, cy-size*0.14, size*0.06, size*0.28, color.RGBA{30, 40, 50, 255}, false)
			// Head / bezel
			vector.FillRect(screen, cx+size*0.10, cy-size*0.22, size*0.18, size*0.44, color.RGBA{70, 90, 110, 255}, false)
			vector.StrokeRect(screen, cx+size*0.10, cy-size*0.22, size*0.18, size*0.44, 1.0, color.RGBA{240, 210, 60, 255}, false)
			// Lens
			vector.FillRect(screen, cx+size*0.28, cy-size*0.18, size*0.06, size*0.36, color.RGBA{220, 245, 255, 255}, false)
			// Power switch
			vector.FillRect(screen, cx-size*0.15, cy-size*0.20, size*0.12, size*0.08, color.RGBA{240, 90, 60, 255}, false)
			// Emission glow
			vector.FillCircle(screen, cx+size*0.32, cy, size*0.15, color.RGBA{255, 240, 100, 140}, false)
		},
	},
	reflect.TypeFor[RepairTool](): {
		Name:     "Repair Tool",
		MaxStack: 1,
		Color:    color.RGBA{235, 140, 40, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Repair Tool"
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			// Welder pistol grip / body
			vector.FillRect(screen, cx-size*0.30, cy-size*0.15, size*0.38, size*0.26, color.RGBA{50, 62, 75, 255}, false)
			vector.StrokeRect(screen, cx-size*0.30, cy-size*0.15, size*0.38, size*0.26, 1.0, color.RGBA{85, 105, 130, 255}, false)
			// Grip handle angled down
			vector.FillRect(screen, cx-size*0.28, cy+size*0.10, size*0.16, size*0.28, color.RGBA{35, 45, 55, 255}, false)
			vector.StrokeRect(screen, cx-size*0.28, cy+size*0.10, size*0.16, size*0.28, 1.0, color.RGBA{65, 80, 100, 255}, false)
			// Heat-treated nozzle barrel
			vector.FillRect(screen, cx+size*0.08, cy-size*0.08, size*0.24, size*0.16, color.RGBA{220, 120, 35, 255}, false)
			vector.StrokeRect(screen, cx+size*0.08, cy-size*0.08, size*0.24, size*0.16, 1.0, color.RGBA{255, 180, 70, 255}, false)
			// Plasma electrode tip
			vector.FillRect(screen, cx+size*0.32, cy-size*0.04, size*0.08, size*0.08, color.RGBA{140, 240, 255, 255}, false)
			// Spark / arc glow
			vector.FillCircle(screen, cx+size*0.38, cy, size*0.12, color.RGBA{255, 230, 90, 180}, false)
			vector.FillCircle(screen, cx+size*0.38, cy, size*0.06, color.RGBA{255, 255, 255, 240}, false)
		},
	},
	reflect.TypeFor[UpgradeSolar](): {
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
	},
	reflect.TypeFor[UpgradeSolarMKII](): {
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
	},
	reflect.TypeFor[UpgradeStorage](): {
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
	},
	reflect.TypeFor[UpgradeStorageMKII](): {
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
	},
	reflect.TypeFor[DecoyLauncher](): {
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
	},
	reflect.TypeFor[ChemicalDischarger](): {
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
	},
	reflect.TypeFor[SonarAmplifier](): {
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
	},
	reflect.TypeFor[SurfaceSonar](): {
		Name:     "Surface Sonar Module",
		MaxStack: 1,
		Color:    color.RGBA{40, 210, 245, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Surface Sonar Module"
			clr := color.RGBA{40, 210, 245, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			// Base module casing
			vector.FillRect(screen, cx-size/2.2, cy-size/2.2, size*0.9, size*0.9, color.RGBA{25, 45, 65, 255}, false)
			vector.StrokeRect(screen, cx-size/2.2, cy-size/2.2, size*0.9, size*0.9, 1.5, clr, false)
			// Concentric radar/sonar arcs
			vector.StrokeCircle(screen, cx, cy+size/6.0, size/2.8, 1.5, clr, false)
			vector.StrokeCircle(screen, cx, cy+size/6.0, size/4.5, 1.2, color.RGBA{140, 240, 255, 220}, false)
			vector.FillCircle(screen, cx, cy+size/6.0, 2.5, color.RGBA{255, 255, 255, 255}, false)
			// Surface water line
			vector.StrokeLine(screen, cx-size/2.6, cy-size/8.0, cx+size/2.6, cy-size/8.0, 1.0, color.RGBA{100, 200, 255, 180}, false)
		},
	},
	reflect.TypeFor[SkiffLight](): {
		Name:     "Skiff Light Module",
		MaxStack: 1,
		Color:    color.RGBA{255, 225, 75, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Skiff Light Module"
			clr := color.RGBA{255, 225, 75, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			// Base module casing (dark navy blue)
			vector.FillRect(screen, cx-size/2.2, cy-size/2.2, size*0.9, size*0.9, color.RGBA{25, 45, 65, 255}, false)
			vector.StrokeRect(screen, cx-size/2.2, cy-size/2.2, size*0.9, size*0.9, 1.5, color.RGBA{70, 100, 140, 255}, false)
			// Reflector dish
			vector.FillCircle(screen, cx-size/8.0, cy, size/3.2, color.RGBA{38, 52, 74, 255}, false)
			vector.StrokeCircle(screen, cx-size/8.0, cy, size/3.2, 1.2, color.RGBA{180, 195, 215, 255}, false)
			// Bulb and glow
			vector.FillCircle(screen, cx-size/8.0, cy, size/5.5, clr, false)
			vector.FillCircle(screen, cx-size/8.0, cy, size/11.0, color.RGBA{255, 255, 255, 255}, false)
			// Beaming light rays forward
			vector.StrokeLine(screen, cx+size/8.0, cy-size/5.5, cx+size/2.1, cy-size/3.0, 1.4, clr, false)
			vector.StrokeLine(screen, cx+size/6.0, cy, cx+size/1.9, cy, 1.8, clr, false)
			vector.StrokeLine(screen, cx+size/8.0, cy+size/5.5, cx+size/2.1, cy+size/3.0, 1.4, clr, false)
		},
	},
	reflect.TypeFor[PowerCell](): {
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
	},
	reflect.TypeFor[ThermalGenerator](): {
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
	},
	reflect.TypeFor[EscapeRocket](): {
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
	},
	reflect.TypeFor[ScoutSubDepthMK1](): {
		Name:     "Scout Sub Depth Module MK1",
		MaxStack: 1,
		Color:    color.RGBA{0, 200, 240, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			name := "Scout Sub Depth Module MK1"
			clr := color.RGBA{0, 200, 240, 255}
			if drawItemIconSprite(screen, name, cx, cy, size) {
				return
			}
			// Pressure module housing
			vector.FillRect(screen, cx-size/2.2, cy-size/2.2, size*0.9, size*0.9, color.RGBA{30, 45, 60, 255}, false)
			vector.StrokeRect(screen, cx-size/2.2, cy-size/2.2, size*0.9, size*0.9, 1.5, clr, false)
			// Depth downward chevron / arrow
			vector.StrokeLine(screen, cx-size/4.0, cy-size/6.0, cx, cy+size/6.0, 2.0, clr, false)
			vector.StrokeLine(screen, cx+size/4.0, cy-size/6.0, cx, cy+size/6.0, 2.0, clr, false)
			// Depth gauge line
			vector.StrokeLine(screen, cx-size/3.0, cy+size/3.5, cx+size/3.0, cy+size/3.5, 1.5, color.RGBA{255, 255, 255, 220}, false)
		},
	},
	reflect.TypeFor[SonicDecoy](): {
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
	},
	reflect.TypeFor[ChemicalDeterrent](): {
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
	},
}

// -----------------------------------------------------------------
// Promoted Generic Nodes for Type-Safe interface matching
// -----------------------------------------------------------------

type BaseItem[T any] struct{}

func (b BaseItem[T]) GetID() ItemID {
	return getMeta(reflect.TypeFor[T]()).ID
}

func (b BaseItem[T]) GetName() string {
	return getMeta(reflect.TypeFor[T]()).Name
}

func (b BaseItem[T]) GetMaxStack() int {
	return getMeta(reflect.TypeFor[T]()).MaxStack
}

func (b BaseItem[T]) GetColor() color.Color {
	return getMeta(reflect.TypeFor[T]()).Color
}

func (b BaseItem[T]) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	getMeta(reflect.TypeFor[T]()).DrawIcon(screen, cx, cy, size)
}

type ConsumableNode[T any] struct {
	BaseItem[T]
}

func (c ConsumableNode[T]) GetHealthRestore() float64 {
	return getMeta(reflect.TypeFor[T]()).HealthRestore
}

func (c ConsumableNode[T]) GetStaminaRestore() float64 {
	return getMeta(reflect.TypeFor[T]()).StaminaRestore
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
	return getMeta(reflect.TypeFor[T]()).MaxO2Capacity
}

type SpeedUpgradeNode[T any] struct {
	PlayerUpgradeNode[T]
}

func (s SpeedUpgradeNode[T]) GetSpeedUpgrade() map[string]Speed {
	return getMeta(reflect.TypeFor[T]()).SpeedUpgrade
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
	return getMeta(reflect.TypeFor[T]()).ModuleType
}

func (u BaseUpgradeNode[T]) GetStorageSlots() int {
	return getMeta(reflect.TypeFor[T]()).StorageSlots
}

func (u BaseUpgradeNode[T]) GetSolarRecharge() float64 {
	return getMeta(reflect.TypeFor[T]()).SolarRecharge
}

type UsableNode[T any] struct {
	BaseItem[T]
}

func (u UsableNode[T]) Use(ctx UsableContext) bool {
	return getMeta(reflect.TypeFor[T]()).Use(ctx)
}
