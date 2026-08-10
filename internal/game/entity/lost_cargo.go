package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/light"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// DayCycleTicks is one full day/night cycle (matches Game.TimeOfDay wrap).
const DayCycleTicks = 14400

// CargoBeaconLifetimeDays is how long a lost-cargo beacon lasts before despawning.
const CargoBeaconLifetimeDays = 3

// CargoBeaconLifetimeTicks is the lifetime of a lost-cargo marker in simulation ticks.
const CargoBeaconLifetimeTicks = CargoBeaconLifetimeDays * DayCycleTicks

// LostCargoBeacon holds inventory the player dropped on death.
// Always placed in overworld space so the return trip is on the surface chart.
type LostCargoBeacon struct {
	Pos           gvec.Vec2
	Cargo         *item.Inventory
	LifetimeTicks int
}

// NewLostCargoBeacon creates a surface beacon at pos holding the given cargo stacks.
func NewLostCargoBeacon(pos gvec.Vec2, stacks []item.ItemStack) *LostCargoBeacon {
	return &LostCargoBeacon{
		Pos:           pos,
		Cargo:         item.NewInventoryFromStacks(stacks),
		LifetimeTicks: CargoBeaconLifetimeTicks,
	}
}

// Active reports whether the beacon still has recoverable cargo.
func (b *LostCargoBeacon) Active() bool {
	return b != nil && b.Cargo != nil && !b.Cargo.IsEmpty() && b.LifetimeTicks > 0
}

// TickLifetime decreases remaining lifetime. Returns true when the beacon expires.
func (b *LostCargoBeacon) TickLifetime() bool {
	if b == nil {
		return true
	}
	b.LifetimeTicks--
	return b.LifetimeTicks <= 0 || b.Cargo == nil || b.Cargo.IsEmpty()
}

// TryRecover transfers cargo into the player inventory.
func (b *LostCargoBeacon) TryRecover(playerInv *item.Inventory) (recovered int, fullyEmpty bool) {
	if b == nil || b.Cargo == nil || b.Cargo.IsEmpty() {
		return 0, true
	}
	before := b.Cargo.TotalItemCount()
	leftover := playerInv.InsertStacks(b.Cargo.ExtractAll())
	b.Cargo = item.NewInventoryFromStacks(leftover)
	after := b.Cargo.TotalItemCount()
	recovered = before - after
	return recovered, b.Cargo.IsEmpty()
}

// Draw renders a large wooden cargo crate with a distress beacon on top.
func (b *LostCargoBeacon) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, lightMult float64) {
	if !b.Active() {
		return
	}
	sx := float32(b.Pos.X - camX)
	sy := float32(b.Pos.Y - camY)

	// Bob gently on the surface (same idea as floating crates).
	bob := float32(math.Sin(ticks*0.05) * 2.5)
	sy += bob

	const size = 22.0
	const half = size / 2.0

	// Wooden crate body (larger / warmer than wreckage crates so it reads as "yours").
	fillClr := light.ApplyLight(color.RGBA{160, 95, 40, 255}, lightMult)
	strokeClr := light.ApplyLight(color.RGBA{230, 150, 70, 255}, lightMult)
	bandClr := light.ApplyLight(color.RGBA{90, 50, 25, 255}, lightMult)

	vector.FillRect(screen, sx-half, sy-half, size, size, fillClr, false)
	vector.StrokeRect(screen, sx-half, sy-half, size, size, 2.0, strokeClr, false)
	// Plank cross
	vector.StrokeLine(screen, sx-half+3, sy-half+3, sx+half-3, sy+half-3, 1.4, strokeClr, false)
	vector.StrokeLine(screen, sx+half-3, sy-half+3, sx-half+3, sy+half-3, 1.4, strokeClr, false)
	// Metal bands
	vector.FillRect(screen, sx-half, sy-3, size, 2.5, bandClr, false)
	vector.FillRect(screen, sx-3, sy-half, 2.5, size, bandClr, false)

	// Distress buoy above crate
	pulse := float32(0.55 + 0.45*math.Sin(ticks*0.14))
	buoyY := sy - half - 10
	vector.StrokeLine(screen, sx, sy-half, sx, buoyY+4, 1.5, color.RGBA{255, 180, 60, 220}, false)
	vector.FillCircle(screen, sx, buoyY, 5+2*pulse, color.RGBA{255, 70, 20, uint8(200 + 55*pulse)}, false)
	vector.StrokeCircle(screen, sx, buoyY, 6+2*pulse, 1.2, color.RGBA{255, 230, 100, 255}, false)
	// Expanding ping rings
	ring := 8 + float32(math.Mod(ticks*0.2, 18.0))
	alpha := uint8(max(0, 160-int(ring*6)))
	vector.StrokeCircle(screen, sx, buoyY, ring, 1.0, color.RGBA{255, 140, 40, alpha}, false)

	// Label always drawn large enough to spot while swimming nearby.
	ebitenutil.DebugPrintAt(screen, "LOST CARGO", int(sx)-38, int(sy-half)-28)
}

// DrawMapIcon draws a chart marker for the lost-cargo site (always pierces fog).
func DrawLostCargoMapIcon(screen *ebiten.Image, px, py float32, ticks float64) {
	pulse := float32(3.0 + (math.Sin(ticks*0.15)+1.0)*1.5)
	vector.FillCircle(screen, px, py, pulse, color.RGBA{255, 90, 20, 255}, false)
	vector.StrokeCircle(screen, px, py, pulse+2, 1.2, color.RGBA{255, 210, 80, 230}, false)
	// Small package diamond
	vector.StrokeLine(screen, px-3, py, px, py-4, 1.4, color.RGBA{255, 240, 180, 255}, false)
	vector.StrokeLine(screen, px, py-4, px+3, py, 1.4, color.RGBA{255, 240, 180, 255}, false)
	vector.StrokeLine(screen, px+3, py, px, py+4, 1.4, color.RGBA{255, 240, 180, 255}, false)
	vector.StrokeLine(screen, px, py+4, px-3, py, 1.4, color.RGBA{255, 240, 180, 255}, false)
}
