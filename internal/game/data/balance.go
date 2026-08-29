package data

import (
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

// PlayerDef holds default balance stats for a new player.
type PlayerDef struct {
	Width            float64
	Height           float64
	MaxHealth        float64
	MaxOxygen        float64
	MaxStamina       float64
	MaxEnergy        float64
	O2DrainRate      float64
	StaminaDrainRate  float64
	StaminaRegenRate  float64
	MiningStaminaCost float64
	DrownDamageRate   float64
	InventorySlots   int
	UpgradeSlots     int
	HotbarSlots      int
	Buoyancy         float64
	Speed            map[string]item.Speed
}

// PlayerArchetype is the shared default balance table for new players.
var PlayerArchetype = &PlayerDef{
	Width:            20,
	Height:           20,
	MaxHealth:        100,
	MaxOxygen:        100,
	MaxStamina:       100,
	MaxEnergy:        100,
	O2DrainRate:      1.0,
	StaminaDrainRate:  1.5,
	StaminaRegenRate:  1.0,
	MiningStaminaCost: 2.0,
	DrownDamageRate:   30.0,
	InventorySlots:   24,
	UpgradeSlots:     4,
	HotbarSlots:      5,
	Buoyancy:         -0.04,
	Speed: map[string]item.Speed{
		"overworld": {
			Drag:         0.88,
			Acceleration: 0.08,
			TopSpeed:     1.6,
		},
		"cave": {
			Drag:         0.92,
			Acceleration: 0.15,
			TopSpeed:     3.5,
		},
	},
}

// DefaultSpeed is the default movement speed map (alias of PlayerArchetype.Speed).
var DefaultSpeed = PlayerArchetype.Speed

// Resource generation balance (owned by resource; re-exported for catalog browsing).
type (
	ResourceGenConfig = resource.ResourceGenConfig
	ResourceTier      = resource.ResourceTier
)

// DefaultGenConfig aliases resource.DefaultGenConfig.
var DefaultGenConfig = &resource.DefaultGenConfig

// GenConfig aliases the mutable active resource generation config.
var GenConfig = &resource.GenConfig
