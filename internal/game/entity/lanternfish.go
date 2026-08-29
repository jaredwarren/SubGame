package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Lanternfish is an abyssal bioluminescent fish with a glowing dorsal lure and photophores.
type Lanternfish struct {
	BaseEntity
	def         *FaunaDef
	FacingRight bool
	SwimPhase   float64
	FleeTimer   int
	BodyColor   color.RGBA
	LureColor   color.RGBA
	GlowColor   color.RGBA
	WanderTimer int
}

var lanternPresets = []struct {
	body color.RGBA
	lure color.RGBA
	glow color.RGBA
}{
	// Electric Cyan
	{
		body: color.RGBA{18, 28, 55, 240},
		lure: color.RGBA{80, 240, 255, 255},
		glow: color.RGBA{30, 180, 240, 120},
	},
	// Radiant Aqua-Green
	{
		body: color.RGBA{15, 38, 45, 240},
		lure: color.RGBA{60, 255, 190, 255},
		glow: color.RGBA{20, 210, 150, 120},
	},
	// Abyssal Violet
	{
		body: color.RGBA{32, 18, 52, 240},
		lure: color.RGBA{210, 110, 255, 255},
		glow: color.RGBA{160, 60, 230, 120},
	},
	// Deep Amber-Gold
	{
		body: color.RGBA{35, 28, 18, 240},
		lure: color.RGBA{255, 200, 70, 255},
		glow: color.RGBA{230, 160, 30, 120},
	},
}

// NewLanternfish creates a new bioluminescent lanternfish entity.
func NewLanternfish(x, y float64, facingRight bool, swimPhase float64) *Lanternfish {
	d := FaunaDefFor(FaunaLanternfish)
	preset := lanternPresets[rand.Intn(len(lanternPresets))]
	return &Lanternfish{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: d.Dims,
			Active:     true,
		},
		def:         d,
		FacingRight: facingRight,
		SwimPhase:   swimPhase,
		BodyColor:   preset.body,
		LureColor:   preset.lure,
		GlowColor:   preset.glow,
		WanderTimer: rand.Intn(120),
	}
}

func (f *Lanternfish) stats() *FaunaDef {
	if f.def != nil {
		return f.def
	}
	return FaunaDefFor(FaunaLanternfish)
}

func (f *Lanternfish) GetHarvestedItem() item.Item {
	return &item.RawFish{}
}

func (f *Lanternfish) CanCatch(playerPos gvec.Vec2) bool {
	d := f.stats()
	cx := f.Pos.X + f.Dimensions.X/2
	cy := f.Pos.Y + f.Dimensions.Y/2
	dx := playerPos.X - cx
	dy := playerPos.Y - cy
	return dx*dx+dy*dy < d.CatchRange*d.CatchRange
}

func (f *Lanternfish) CatchPrompt(playerPos gvec.Vec2) string {
	if f.CanCatch(playerPos) {
		return "Press [E] to Catch Lanternfish"
	}
	return ""
}

func (f *Lanternfish) Update(gr Runtime) {
	d := f.stats()
	pPos := gr.PlayerPos()
	cx := f.Pos.X + f.Dimensions.X/2
	cy := f.Pos.Y + f.Dimensions.Y/2
	dx := pPos.X - cx
	dy := pPos.Y - cy
	distSq := dx*dx + dy*dy

	f.SwimPhase += d.SwimPhaseSpeed

	// Flee if player is too close
	if distSq < d.FleeRange*d.FleeRange {
		f.FleeTimer = 45
		f.FacingRight = dx < 0
	}

	if f.FleeTimer > 0 {
		f.FleeTimer--
		f.SwimPhase += 0.08
		speed := d.FleeSpeed
		if f.FacingRight {
			f.Vel.X = speed
		} else {
			f.Vel.X = -speed
		}
		f.Vel.Y = math.Sin(f.SwimPhase*2.0) * 0.8
	} else {
		// Gentle meandering wander
		f.WanderTimer--
		if f.WanderTimer <= 0 {
			f.WanderTimer = 90 + rand.Intn(90)
			if rand.Float64() < 0.35 {
				f.FacingRight = !f.FacingRight
			}
		}
		speed := d.CruiseSpeed
		if f.FacingRight {
			f.Vel.X = speed
		} else {
			f.Vel.X = -speed
		}
		f.Vel.Y = math.Sin(f.SwimPhase) * 0.45
	}

	nextX := f.Pos.X + f.Vel.X
	nextY := f.Pos.Y + f.Vel.Y
	if !gr.IsSolid(nextX, nextY, f.Dimensions.X, f.Dimensions.Y) {
		f.Pos.X = nextX
		f.Pos.Y = nextY
	} else {
		f.FacingRight = !f.FacingRight
		f.FleeTimer = 0
	}
}

func (f *Lanternfish) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(f.Pos.X - camera.Pos.X)
	sy := float32(f.Pos.Y - camera.Pos.Y)
	sw := float32(f.Dimensions.X)
	sh := float32(f.Dimensions.Y)
	cx := sx + sw/2
	cy := sy + sh/2

	pulse := float32(math.Sin(f.SwimPhase*2.5))*0.5 + 0.5

	// 1. Soft ethereal bioluminescent glow (transparent multi-ring radial falloff)
	baseR, baseG, baseB := f.GlowColor.R, f.GlowColor.G, f.GlowColor.B
	maxRadius := float32(18.0 + 4.0*pulse)
	// Outer faint aura (soft transparent perimeter)
	vector.FillCircle(screen, cx, cy, maxRadius, color.RGBA{baseR, baseG, baseB, uint8(14 + 8*pulse)}, false)
	// Mid ambient radiance
	vector.FillCircle(screen, cx, cy, maxRadius*0.62, color.RGBA{baseR, baseG, baseB, uint8(24 + 12*pulse)}, false)
	// Inner body bloom
	vector.FillCircle(screen, cx, cy, maxRadius*0.35, color.RGBA{baseR, baseG, baseB, uint8(38 + 16*pulse)}, false)

	// 2. Translucent deep fish body
	vector.FillCircle(screen, cx, cy, 5.5, f.BodyColor, false)

	// 3. Tail fin with sinusoidal undulation
	var tailX, headX float32
	var dir float32 = 1.0
	if f.FacingRight {
		tailX = cx - 7.5
		headX = cx + 5.5
		dir = 1.0
	} else {
		tailX = cx + 7.5
		headX = cx - 5.5
		dir = -1.0
	}
	wiggle := float32(math.Sin(float64(f.SwimPhase))) * 2.5
	tailClr := color.RGBA{f.BodyColor.R + 20, f.BodyColor.G + 30, f.BodyColor.B + 40, 180}
	vector.StrokeLine(screen, tailX, cy, tailX-4*dir+wiggle, cy-4, 1.5, tailClr, false)
	vector.StrokeLine(screen, tailX, cy, tailX-4*dir+wiggle, cy+4, 1.5, tailClr, false)

	// 4. Large dark-adapted eye
	eyeX := headX - 1.5*dir
	vector.FillCircle(screen, eyeX, cy-1.0, 1.8, color.RGBA{180, 220, 240, 255}, false)
	vector.FillCircle(screen, eyeX, cy-1.0, 1.0, color.RGBA{10, 10, 20, 255}, false)

	// 5. Ventral photophore dots along belly (pulsing bioluminescence)
	photoClr := color.RGBA{f.LureColor.R, f.LureColor.G, f.LureColor.B, uint8(180 + 75*pulse)}
	for i := float32(-2); i <= 2; i += 2 {
		vector.FillCircle(screen, cx+i*dir, cy+3.5, 0.9, photoClr, false)
	}

	// 6. Dorsal lantern antenna (esca) with glowing bulb tip
	lureBaseX := cx + 1.0*dir
	lureBaseY := cy - 4.5
	lureTipX := headX + 3.0*dir
	lureTipY := cy - 5.0 + float32(math.Sin(float64(f.SwimPhase)*1.5))*1.2

	antennaClr := color.RGBA{100, 140, 180, 200}
	vector.StrokeLine(screen, lureBaseX, lureBaseY, lureTipX, lureTipY, 1.0, antennaClr, false)

	// Glowing lure bulb (soft transparent light flare)
	lureR, lureG, lureB := f.LureColor.R, f.LureColor.G, f.LureColor.B
	vector.FillCircle(screen, lureTipX, lureTipY, 6.0+2.0*pulse, color.RGBA{lureR, lureG, lureB, uint8(20 + 15*pulse)}, false)
	vector.FillCircle(screen, lureTipX, lureTipY, 3.5+1.0*pulse, color.RGBA{lureR, lureG, lureB, uint8(60 + 30*pulse)}, false)
	vector.FillCircle(screen, lureTipX, lureTipY, 1.8, f.LureColor, false)
	vector.FillCircle(screen, lureTipX, lureTipY, 0.8, color.White, false)
}
