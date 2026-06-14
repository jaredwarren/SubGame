package base

import (
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/assets"
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
	if timeOfDay < 10800 {
		rate = b.SolarRechargeRate
	}
	b.Power += rate
	if b.Power > b.MaxPower {
		b.Power = b.MaxPower
	}
}

var (
	lifePodSprite       *ebiten.Image
	lifePodSpriteLoaded bool
)

// LoadAssets preloads and chroma-keys the base Life Pod sprite.
func LoadAssets() {
	sprite, err := assets.LoadChromaKeyedImage("lifepod_surface", assets.WithTrim())
	if err != nil {
		log.Printf("Error: Failed to load lifepod surface: %v", err)
		return
	}
	lifePodSprite = sprite
	lifePodSpriteLoaded = true
}

// Draw renders the base station in the overworld viewport.
func (b *BaseStation) Draw(screen *ebiten.Image, camera *camera.Camera, lightMult float64) {
	sx := float32(b.Pos.X - camera.Pos.X)
	sy := float32(b.Pos.Y - camera.Pos.Y)

	if lifePodSprite != nil {
		wImg, hImg := lifePodSprite.Bounds().Dx(), lifePodSprite.Bounds().Dy()
		if wImg > 0 && hImg > 0 {
			op := &ebiten.DrawImageOptions{}

			// We want the draw width of the lifepod to be 72.0 pixels, scaling the height proportionally
			const targetWidth = 72.0
			scale := targetWidth / float64(wImg)
			op.GeoM.Scale(scale, scale)

			// Center the scaled image relative to the base station's collision/anchor box
			drawW := targetWidth
			drawH := float64(hImg) * scale
			dx := float64(sx) + b.Size.X/2.0 - drawW/2.0
			dy := float64(sy) + b.Size.Y/2.0 - drawH/2.0

			op.GeoM.Translate(dx, dy)

			// Apply lighting scale
			mult := float32(lightMult)
			op.ColorScale.Scale(mult, mult, mult, 1.0)

			screen.DrawImage(lifePodSprite, op)
			return
		}
	}

	// Fallback to original vector drawing code if image is not loaded/available
	podColor := applyLight(color.RGBA{240, 240, 245, 255}, lightMult) // Clean white pod
	vector.FillRect(screen, sx, sy, float32(b.Size.X), float32(b.Size.Y), podColor, false)
	vector.StrokeRect(screen, sx, sy, float32(b.Size.X), float32(b.Size.Y), 2.0, applyLight(color.RGBA{220, 100, 30, 255}, lightMult), false) // Orange highlights

	// Inner window/hatch details
	vector.FillCircle(screen, sx+24, sy+24, 10, applyLight(color.RGBA{40, 80, 120, 255}, lightMult), false)
	vector.StrokeCircle(screen, sx+24, sy+24, 10, 1.0, applyLight(color.RGBA{180, 200, 220, 255}, lightMult), false)

	// Draw antenna blinking red light
	vector.FillCircle(screen, sx+24, sy+4, 3, applyLight(color.RGBA{235, 45, 45, 255}, lightMult), false)
}

func applyLight(c color.RGBA, mult float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * mult),
		G: uint8(float64(c.G) * mult),
		B: uint8(float64(c.B) * mult),
		A: c.A,
	}
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

// WouldUninstallOverflow returns true if removing the upgrade at the given slot index
// would cause the base station's storage inventory to lose items due to shrinking capacity.
func (b *BaseStation) WouldUninstallOverflow(upgradeIdx int) bool {
	if b.Upgrades == nil || upgradeIdx < 0 || upgradeIdx >= len(b.Upgrades.Slots) {
		return false
	}
	upgItem := b.Upgrades.Slots[upgradeIdx].Item
	if upgItem == nil {
		return false
	}
	upg, ok := upgItem.(item.UpgradeItem)
	if !ok {
		return false
	}

	// If it doesn't grant storage, it's always safe
	if upg.GetStorageSlots() == 0 {
		return false
	}

	// Calculate what the storage capacity would be WITHOUT this upgrade
	storageSlots := 24
	for i, slot := range b.Upgrades.Slots {
		if i == upgradeIdx || slot.Item == nil {
			continue
		}
		if otherUpg, ok := slot.Item.(item.UpgradeItem); ok {
			if slots := otherUpg.GetStorageSlots(); slots > storageSlots {
				storageSlots = slots
			}
		}
	}

	if b.Storage != nil {
		return b.Storage.WouldResizeLoseItems(storageSlots)
	}
	return false
}
