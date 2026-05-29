package base

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// BaseStation represents a player base / Life Pod anchor terminal.
type BaseStation struct {
	Pos               gvec.Vec2
	Size              gvec.Vec2
	Power             float64
	MaxPower          float64
	Storage           *item.Inventory
	Upgrades          *item.Inventory // 4 slots for active upgrades/modules
	SolarRechargeRate float64
	ActiveModules     map[item.BaseModule]bool
}

// NewBaseStation instantiates a BaseStation (e.g. starting Life Pod).
func NewBaseStation(x, y float64) *BaseStation {
	b := &BaseStation{
		Pos:           gvec.Vec2{X: x, Y: y},
		Size:          gvec.Vec2{X: 48, Y: 48},
		Power:         75.0,
		MaxPower:      100.0,
		Storage:       item.NewInventory(24), // 24 storage slots in base vault
		Upgrades:      item.NewInventory(4),  // 4 upgrade slots
		ActiveModules: make(map[item.BaseModule]bool),
	}
	b.RecalculateProperties()
	return b
}

// UpdatePower simulates base solar recharging and module draining.
func (b *BaseStation) UpdatePower(timeOfDay float64) {
	rate := 0.01 // baseline trickle
	if timeOfDay < 7200 {
		rate = b.SolarRechargeRate
	}
	b.Power += rate
	if b.Power > b.MaxPower {
		b.Power = b.MaxPower
	}
}

// Draw renders the base station in the overworld viewport.
func (b *BaseStation) Draw(screen *ebiten.Image, camera *camera.Camera) {
	sx := float32(b.Pos.X - camera.Pos.X)
	sy := float32(b.Pos.Y - camera.Pos.Y)

	// Draw Life Pod hull (rounded hexagonal/pod shape)
	podColor := color.RGBA{240, 240, 245, 255} // Clean white pod
	vector.FillRect(screen, sx, sy, float32(b.Size.X), float32(b.Size.Y), podColor, false)
	vector.StrokeRect(screen, sx, sy, float32(b.Size.X), float32(b.Size.Y), 2.0, color.RGBA{220, 100, 30, 255}, false) // Orange highlights

	// Inner window/hatch details
	vector.FillCircle(screen, sx+24, sy+24, 10, color.RGBA{40, 80, 120, 255}, false)
	vector.StrokeCircle(screen, sx+24, sy+24, 10, 1.0, color.RGBA{180, 200, 220, 255}, false)

	// Draw antenna blinking red light
	vector.FillCircle(screen, sx+24, sy+4, 3, color.RGBA{235, 45, 45, 255}, false)
}

// DistanceToPlayer returns the distance from base center to player center.
func (b *BaseStation) DistanceToPlayer(p *player.Player) float64 {
	bx := b.Pos.X + b.Size.X/2.0
	by := b.Pos.Y + b.Size.Y/2.0
	px := p.Pos.X + p.Width/2.0
	py := p.Pos.Y + p.Height/2.0
	return math.Hypot(bx-px, by-py)
}

// RecalculateProperties updates dynamic stats and active modules from installed upgrades.
func (b *BaseStation) RecalculateProperties() {
	b.SolarRechargeRate = 0.01 // baseline trickle
	storageSlots := 24         // baseline vault size

	// Reset active modules map (Fabricator and Medical Bay are default built-ins)
	b.ActiveModules = map[item.BaseModule]bool{
		item.ModuleFabricator:  true,
		item.ModuleMedical:     true,
		item.ModuleSolar:       false,
		item.ModuleSolarMKII:   false,
		item.ModuleStorage:     false,
		item.ModuleStorageMKII: false,
	}

	if b.Upgrades != nil {
		for _, slot := range b.Upgrades.Slots {
			if slot.Item == nil {
				continue
			}
			if upg, ok := slot.Item.(item.UpgradeItem); ok {
				modType := upg.GetModuleType()
				b.ActiveModules[modType] = true

				if recharge := upg.GetSolarRecharge(); recharge > b.SolarRechargeRate {
					b.SolarRechargeRate = recharge
				}
				if slots := upg.GetStorageSlots(); slots > storageSlots {
					storageSlots = slots
				}
			}
		}
	}

	// Upgrade level implications
	if b.ActiveModules[item.ModuleSolarMKII] {
		b.ActiveModules[item.ModuleSolar] = true
	}
	if b.ActiveModules[item.ModuleStorageMKII] {
		b.ActiveModules[item.ModuleStorage] = true
	}

	if b.Storage != nil {
		b.Storage.Resize(storageSlots)
	}
}

// HasModule checks if a module is active, by checking default built-ins or scanning upgrades inventory.
func (b *BaseStation) HasModule(mod item.BaseModule) bool {
	if b.ActiveModules == nil {
		return false
	}
	return b.ActiveModules[mod]
}

// InstallUpgrade attempts to slot an upgrade item into the base upgrades inventory.
func (b *BaseStation) InstallUpgrade(it item.Item) bool {
	if it == nil || b.Upgrades == nil {
		return false
	}
	if _, ok := it.(item.UpgradeItem); ok {
		if b.Upgrades.AddItem(it, 1) {
			b.RecalculateProperties()
			return true
		}
	}
	return false
}
