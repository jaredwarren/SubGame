package game

import (
	"math"

	"github.com/jaredwarren/SubGame/internal/game/item"
)

// Player represents the player character, including their physics and stats.
type Player struct {
	// Physics
	Pos    Vec2
	Vel    Vec2
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

	// Upgrade Cache (Option A)
	HasFins bool
}

// NewPlayer initializes a player with default stats and empty inventory.
func NewPlayer(x, y float64) *Player {
	p := &Player{
		Pos:              Vec2{X: x, Y: y},
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
	}
	p.RecalculateUpgrades()
	return p
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
	if isSprinting && (math.Abs(p.Vel.X) > 0.1 || math.Abs(p.Vel.Y) > 0.1) {
		p.CurrentStamina -= p.StaminaDrainRate / 60.0 // Sprinting drains stamina
		if p.CurrentStamina < 0 {
			p.CurrentStamina = 0
		}
	} else {
		p.CurrentStamina += p.StaminaRegenRate / 60.0 // Regenerate stamina when not sprinting
		if p.CurrentStamina > p.MaxStamina {
			p.CurrentStamina = p.MaxStamina
		}
	}

	// Health clamp
	if p.CurrentHealth < 0 {
		p.CurrentHealth = 0
	}
	if p.CurrentHealth > p.MaxHealth {
		p.CurrentHealth = p.MaxHealth
	}
}

// pCenterX returns the screen X position where the player is drawn (centered).
func pCenterX(p *Player) float64 {
	return ScreenWidth / 2
}

// pCenterY returns the screen Y position where the player is drawn (centered).
func pCenterY(p *Player) float64 {
	return ScreenHeight / 2
}

// RecalculateUpgrades scans the inventory and updates cached upgrade flags and capacity stats.
func (p *Player) RecalculateUpgrades() {
	p.HasFins = item.HasItem[*item.Fins](p.Inventory, 1)

	if item.HasItem[*item.O2TankUHC](p.Inventory, 1) {
		p.MaxOxygen = 240.0
	} else if item.HasItem[*item.O2TankHC](p.Inventory, 1) {
		p.MaxOxygen = 160.0
	} else {
		p.MaxOxygen = 100.0
	}
}
