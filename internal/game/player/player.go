package player

import (
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// InvulnerabilityFrames defines the default duration of hit invulnerability frames (~1.0s at 60 FPS).
const InvulnerabilityFrames = 60

// Player represents the player character, including their physics and stats.
type Player struct {
	// Physics
	Pos             gvec.Vec2
	Vel             gvec.Vec2
	Width           float64
	Height          float64
	Facing          float64 // Angle in radians (for boat/body facing direction)
	FlashlightAngle float64 // Angle in radians for flashlight beam direction

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
	O2DrainRate       float64 // default: 1.0 (O2 units per second)
	StaminaDrainRate  float64 // default: 1.5 (Stamina units per second when sprinting)
	StaminaRegenRate  float64 // default: 1.0 (Stamina units recovered per second)
	MiningStaminaCost float64 // default: 2.0 (Stamina spent per mining tool swing)
	DrownDamageRate   float64 // default: 30.0 (Health units lost per second when drowning)

	// Inventory
	Inventory  *item.Inventory
	Upgrades   *item.Inventory // 4 upgrade/equipment slots
	Hotbar     *item.Inventory // 5 quick-select slots
	ActiveSlot int             // 0 to 4

	// Upgrade Cache (Option A)
	Speed    map[string]item.Speed
	Buoyancy float64

	// Animations
	AnimTick          int
	IsMining          bool
	MiningAnimTimer   int
	LastHealth        float64
	IsDamaged         bool
	DamageAnimTimer   int
	InvulnerableTimer int
	StunTimer         int
	SlowTimer         int
	SlowFactor        float64
	SuperSpeed        bool
}

// NewPlayer initializes a player with default stats and empty inventory.
func NewPlayer(x, y float64) *Player {
	d := PlayerArchetype
	p := &Player{
		Pos:               gvec.Vec2{X: x, Y: y},
		Width:             d.Width,
		Height:            d.Height,
		MaxHealth:         d.MaxHealth,
		CurrentHealth:     d.MaxHealth,
		MaxOxygen:         d.MaxOxygen,
		CurrentOxygen:     d.MaxOxygen,
		MaxStamina:        d.MaxStamina,
		CurrentStamina:    d.MaxStamina,
		MaxEnergy:         d.MaxEnergy,
		CurrentEnergy:     d.MaxEnergy,
		O2DrainRate:       d.O2DrainRate,
		StaminaDrainRate:  d.StaminaDrainRate,
		StaminaRegenRate:  d.StaminaRegenRate,
		MiningStaminaCost: d.MiningStaminaCost,
		DrownDamageRate:   d.DrownDamageRate,
		Inventory:         item.NewInventory(d.InventorySlots),
		Upgrades:          item.NewInventory(d.UpgradeSlots),
		Hotbar:            item.NewInventory(d.HotbarSlots),
		ActiveSlot:        0,
		LastHealth:        d.MaxHealth,
		Speed:             d.Speed,
		Buoyancy:          d.Buoyancy,
	}
	p.RecalculateUpgrades()
	return p
}

// UpdateStats handles core stat loops (depleting/regenerating O2, stamina, etc.)
func (p *Player) UpdateStats(inCave bool, isSprinting bool) {
	if p.StunTimer > 0 {
		p.StunTimer--
		p.Vel = gvec.Vec2{}
	}

	if p.SlowTimer > 0 {
		if !inCave {
			p.SlowTimer = 0
			p.SlowFactor = 1.0
		} else {
			p.SlowTimer--
			if p.SlowTimer == 0 {
				p.SlowFactor = 1.0
			}
		}
	}

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

// TrySpendMiningStamina deducts the per-swing mining cost when the player has stamina left.
func (p *Player) TrySpendMiningStamina() bool {
	if p.CurrentStamina <= 0 {
		return false
	}
	p.CurrentStamina -= p.MiningStaminaCost
	if p.CurrentStamina < 0 {
		p.CurrentStamina = 0
	}
	return true
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

	// Only allow Fins and O2 Tanks for player body gear upgrades
	switch it.(type) {
	case item.O2UpgradeItem, item.SpeedUpgradeItem:
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

	if p.SuperSpeed {
		scaled := make(map[string]item.Speed, len(p.Speed))
		for k, s := range p.Speed {
			scaled[k] = item.Speed{
				Drag:         s.Drag,
				Acceleration: s.Acceleration * 2.5,
				TopSpeed:     s.TopSpeed * 2.5,
			}
		}
		p.Speed = scaled
	}
}

// UpdateAnimation increments frame counts and ticks for player visual animations.
func (p *Player) UpdateAnimation() {
	p.AnimTick++

	// Handle invulnerability timer
	if p.InvulnerableTimer > 0 {
		p.InvulnerableTimer--
	}

	// Handle mining timer
	if p.IsMining {
		p.MiningAnimTimer--
		if p.MiningAnimTimer <= 0 {
			p.IsMining = false
		}
	}

	// Handle passive damage detection (e.g. drowning or direct health drop)
	if p.CurrentHealth < p.LastHealth && p.DamageAnimTimer <= 0 {
		p.IsDamaged = true
		p.DamageAnimTimer = 20 // ~0.3 seconds at 60 FPS for continuous/passive damage
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

// TakeDamage applies combat or hazard damage to the player if not currently invulnerable.
// It activates brief invulnerability frames (~1.0s) to prevent instakills from overlapping hazards.
// Returns true if damage was applied, false if the attack was blocked by invulnerability or amount was non-positive.
func (p *Player) TakeDamage(amount float64) bool {
	if p.InvulnerableTimer > 0 || amount <= 0 {
		return false
	}
	p.CurrentHealth -= amount
	if p.CurrentHealth < 0 {
		p.CurrentHealth = 0
	}
	p.InvulnerableTimer = InvulnerabilityFrames
	p.IsDamaged = true
	p.DamageAnimTimer = InvulnerabilityFrames
	p.LastHealth = p.CurrentHealth
	return true
}

// GetActiveItem returns the item equipped in the active hotbar slot, or nil.
func (p *Player) GetActiveItem() item.Item {
	if p.Hotbar == nil || p.ActiveSlot < 0 || p.ActiveSlot >= len(p.Hotbar.Slots) {
		return nil
	}
	return p.Hotbar.Slots[p.ActiveSlot].Item
}
