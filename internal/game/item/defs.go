package item

import "image/color"

// ItemDef is the data-driven description of an inventory item. Runtime item
// structs still exist (reflection factories), but identity and browseable
// catalog lookups should go through ItemID → Def rather than display names.
//
// Full Stack-based runtime (no per-item structs) is deferred; this catalog is
// the bridge so new content can key on IDs immediately.
type ItemDef struct {
	ID       ItemID
	Name     string
	MaxStack int
	Color    color.Color

	// Capability fields — zero / nil means the item lacks that capability.
	MaxO2Capacity  float64
	SpeedUpgrade   map[string]Speed
	ModuleType     BaseModule
	StorageSlots   int
	SolarRecharge  float64
	HealthRestore  float64
	StaminaRestore float64
	Usable         bool
	IsConsumable   bool
	IsO2Upgrade    bool
	IsSpeedUpgrade bool
	IsBaseModule   bool
}

// Catalog is the ItemID → ItemDef table built from the item registry.
var Catalog = map[ItemID]*ItemDef{}

func init() {
	rebuildCatalog()
}

func rebuildCatalog() {
	if len(idToType) == 0 {
		initItemRegistryLookup()
	}
	for id, t := range idToType {
		meta := getMeta(t)
		def := &ItemDef{
			ID:             id,
			Name:           meta.Name,
			MaxStack:       meta.MaxStack,
			Color:          meta.Color,
			MaxO2Capacity:  meta.MaxO2Capacity,
			SpeedUpgrade:   meta.SpeedUpgrade,
			ModuleType:     meta.ModuleType,
			StorageSlots:   meta.StorageSlots,
			SolarRecharge:  meta.SolarRecharge,
			HealthRestore:  meta.HealthRestore,
			StaminaRestore: meta.StaminaRestore,
			Usable:          meta.Use != nil,
			IsConsumable:   meta.HealthRestore > 0 || meta.StaminaRestore > 0,
			IsO2Upgrade:    meta.MaxO2Capacity > 0,
			IsSpeedUpgrade: len(meta.SpeedUpgrade) > 0,
			IsBaseModule:   meta.StorageSlots > 0 || meta.SolarRecharge > 0,
		}
		Catalog[id] = def
	}
}

// Def returns the catalog entry for id, or nil if unknown.
func Def(id ItemID) *ItemDef {
	if len(Catalog) == 0 {
		rebuildCatalog()
	}
	return Catalog[id]
}
