package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// BaseModule represents a functional slot upgrade in the base schematic.
type BaseModule int

const (
	ModuleFabricator BaseModule = iota
	ModuleStorage
	ModuleMedical
	ModuleSolar
)

func (m BaseModule) String() string {
	switch m {
	case ModuleFabricator:
		return "Fabricator Module"
	case ModuleStorage:
		return "Storage Vault"
	case ModuleMedical:
		return "Medical Bay"
	case ModuleSolar:
		return "Solar Array"
	default:
		return "Module"
	}
}

// BaseStation represents a player base / Life Pod anchor terminal.
type BaseStation struct {
	Pos      gvec.Vec2
	Size     gvec.Vec2
	Power    float64
	MaxPower float64
	Storage  *item.Inventory
	Modules  map[BaseModule]bool
}

// NewBaseStation instantiates a BaseStation (e.g. starting Life Pod).
func NewBaseStation(x, y float64) *BaseStation {
	b := &BaseStation{
		Pos:      gvec.Vec2{X: x, Y: y},
		Size:     gvec.Vec2{X: 48, Y: 48},
		Power:    75.0,
		MaxPower: 100.0,
		Storage:  item.NewInventory(24), // 24 storage slots in base vault
		Modules:  make(map[BaseModule]bool),
	}
	// Starting base includes Fabricator and Medical Bay, solar is upgradeable
	b.Modules[ModuleFabricator] = true
	b.Modules[ModuleMedical] = true
	b.Modules[ModuleStorage] = false
	b.Modules[ModuleSolar] = false

	return b
}

// UpdatePower simulates base solar recharging and module draining.
func (b *BaseStation) UpdatePower() {
	// Recharging: if solar array is installed, recharge power
	if b.Modules[ModuleSolar] {
		b.Power += 0.08 // solar power trickle
	} else {
		b.Power += 0.01 // very slow emergency backup trickle
	}

	if b.Power > b.MaxPower {
		b.Power = b.MaxPower
	}
}

// Draw renders the base station in the overworld viewport.
func (b *BaseStation) Draw(screen *ebiten.Image, camera *Camera) {
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
func (b *BaseStation) DistanceToPlayer(p *Player) float64 {
	bx := b.Pos.X + b.Size.X/2.0
	by := b.Pos.Y + b.Size.Y/2.0
	px := p.Pos.X + p.Width/2.0
	py := p.Pos.Y + p.Height/2.0
	return math.Hypot(bx-px, by-py)
}
