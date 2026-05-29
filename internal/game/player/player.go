package player

import (
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Player represents the player character, including their physics and stats.
type Player struct {
	// Physics
	Pos    gvec.Vec2
	Vel    gvec.Vec2
	Width  float64
	Height float64
	Facing float64 // Angle in radians (for flashlight/boat direction)

	// Stats
	MaxHealth      float64
	CurrentHealth  float64
	MaxOxygen      float64
	CurrentOxygen  float64
	MaxStamina     float64
	CurrentStamina float64
	MaxEnergy      float64
	CurrentEnergy  float64

	// Customizable Stat Rates (expressed per second)
	O2DrainRate      float64 // default: 1.0 (O2 units per second)
	StaminaDrainRate float64 // default: 1.5 (Stamina units per second when sprinting)
	StaminaRegenRate float64 // default: 1.0 (Stamina units recovered per second)
	DrownDamageRate  float64 // default: 30.0 (Health units lost per second when drowning)

	// Inventory
	Inventory *item.Inventory
	Upgrades  *item.Inventory // 4 upgrade/equipment slots

	// Upgrade Cache (Option A)
	Speed    map[string]item.Speed
	Buoyancy float64

	// Animations
	AnimTick        int
	IsMining        bool
	MiningAnimTimer int
	LastHealth      float64
	IsDamaged       bool
	DamageAnimTimer int
}

// NewPlayer initializes a player with default stats and empty inventory.
func NewPlayer(x, y float64) *Player {
	p := &Player{
		Pos:              gvec.Vec2{X: x, Y: y},
		Width:            20,
		Height:           20,
		MaxHealth:        100,
		CurrentHealth:    100,
		MaxOxygen:        100, // 100 seconds of O2 initially
		CurrentOxygen:    100,
		MaxStamina:       100,
		CurrentStamina:   100,
		MaxEnergy:        100,
		CurrentEnergy:    100,
		O2DrainRate:      1.0,
		StaminaDrainRate: 1.5,
		StaminaRegenRate: 1.0,
		DrownDamageRate:  30.0,
		Inventory:        item.NewInventory(24),
		Upgrades:         item.NewInventory(4),
		LastHealth:       100.0,
		Speed:            DefaultSpeed,
		Buoyancy:         -0.04,
	}
	p.RecalculateUpgrades()
	return p
}

var DefaultSpeed = map[string]item.Speed{
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
}

// UpdateStats handles core stat loops (depleting/regenerating O2, stamina, etc.)
func (p *Player) UpdateStats(inCave bool, isSprinting bool) {
	// Oxygen management
	if inCave {
		p.CurrentOxygen -= p.O2DrainRate / 60.0 // Drain O2 per second (at 60 FPS)
		if p.CurrentOxygen < 0 {
			p.CurrentOxygen = 0
			p.CurrentHealth -= p.DrownDamageRate / 60.0 // Drowning damage per second
		}
	} else {
		// Instantly refill or quickly refill oxygen on surface
		p.CurrentOxygen = p.MaxOxygen
	}

	// Stamina management
	if isSprinting {
		p.CurrentStamina -= p.StaminaDrainRate / 60.0
		if p.CurrentStamina < 0 {
			p.CurrentStamina = 0
		}
	} else {
		p.CurrentStamina += p.StaminaRegenRate / 60.0
		if p.CurrentStamina > p.MaxStamina {
			p.CurrentStamina = p.MaxStamina
		}
	}
}

// ClampStats restricts status metrics to their bounds.
func (p *Player) ClampStats() {
	if p.CurrentOxygen < 0 {
		p.CurrentOxygen = 0
	}
	if p.CurrentOxygen > p.MaxOxygen {
		p.CurrentOxygen = p.MaxOxygen
	}
	if p.CurrentStamina < 0 {
		p.CurrentStamina = 0
	}
	if p.CurrentStamina > p.MaxStamina {
		p.CurrentStamina = p.MaxStamina
	}
	if p.CurrentHealth < 0 {
		p.CurrentHealth = 0
	}
	if p.CurrentHealth > p.MaxHealth {
		p.CurrentHealth = p.MaxHealth
	}
}

// CenterX returns the screen X position where the player is drawn (centered).
func (p *Player) CenterX() float64 {
	return config.ScreenWidth / 2
}

// CenterY returns the screen Y position where the player is drawn (centered).
func (p *Player) CenterY() float64 {
	return config.ScreenHeight / 2
}

// EquipUpgrade attempts to slot an item into the player's upgrades slots.
func (p *Player) EquipUpgrade(it any) bool {
	if it == nil || p.Upgrades == nil {
		return false
	}

	// Only allow Fins and O2 Tanks for player upgrades
	_, ok := it.(item.PlayerUpgradeItem)
	if ok {
		if p.Upgrades.AddItem(it.(item.Item), 1) {
			p.RecalculateUpgrades()
			return true
		}
	}
	return false
}

// RecalculateUpgrades scans the upgrades and updates cached upgrade flags and capacity stats.
func (p *Player) RecalculateUpgrades() {
	p.MaxOxygen = 100.0
	p.Speed = DefaultSpeed

	for _, v := range p.Upgrades.Slots {
		if _, ok := v.Item.(item.O2UpgradeItem); ok {
			p.MaxOxygen += v.Item.(item.O2UpgradeItem).GetMaxO2Capacity()
		}

		if _, ok := v.Item.(item.SpeedUpgradeItem); ok {
			p.Speed = v.Item.(item.SpeedUpgradeItem).GetSpeedUpgrade()
		}
	}

}

// UpdateAnimation increments frame counts and ticks for player visual animations.
func (p *Player) UpdateAnimation() {
	p.AnimTick++

	// Handle mining timer
	if p.IsMining {
		p.MiningAnimTimer--
		if p.MiningAnimTimer <= 0 {
			p.IsMining = false
		}
	}

	// Handle damage detection (health drops)
	if p.CurrentHealth < p.LastHealth {
		p.IsDamaged = true
		p.DamageAnimTimer = 20 // ~0.3 seconds at 60 FPS
	}
	p.LastHealth = p.CurrentHealth

	// Handle damage timer
	if p.IsDamaged {
		p.DamageAnimTimer--
		if p.DamageAnimTimer <= 0 {
			p.IsDamaged = false
		}
	}
}
