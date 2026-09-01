package entity

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

var (
	lanternRadialGlowImage *ebiten.Image
	lanternRadialGlowOnce  sync.Once
)

// getLanternRadialGlowImage generates a procedural smooth Gaussian-like radial light texture
// with continuous falloff to zero at the perimeter for natural bioluminescent light scattering.
func getLanternRadialGlowImage() *ebiten.Image {
	lanternRadialGlowOnce.Do(func() {
		const size = 128
		const center = float64(size) / 2.0
		const maxRadius = center - 2.0

		img := image.NewNRGBA(image.Rect(0, 0, size, size))
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x) + 0.5 - center
				dy := float64(y) + 0.5 - center
				dist := math.Hypot(dx, dy)
				if dist < maxRadius {
					t := dist / maxRadius
					fade := 1.0 - t
					// Smooth cubic-quintic ease with zero edge derivative
					intensity := fade * fade * (3.0 - 2.0*fade) * fade
					if intensity > 1.0 {
						intensity = 1.0
					} else if intensity < 0.0 {
						intensity = 0.0
					}
					alpha := uint8(intensity * 255.0)
					img.SetNRGBA(x, y, color.NRGBA{
						R: 255,
						G: 255,
						B: 255,
						A: alpha,
					})
				}
			}
		}
		lanternRadialGlowImage = ebiten.NewImageFromImage(img)
	})
	return lanternRadialGlowImage
}

// drawLanternRadialGlow renders a smooth, subtle light aura centered at (cx, cy) with bilinear GPU filtering.
func drawLanternRadialGlow(screen *ebiten.Image, cx, cy, radius float32, clr color.RGBA, alphaScale float32) {
	glowImg := getLanternRadialGlowImage()
	if glowImg == nil || radius <= 0 || alphaScale <= 0 {
		return
	}
	b := glowImg.Bounds()
	w := float32(b.Dx())
	h := float32(b.Dy())

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	scale := float64(radius*2.0) / float64(w)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(cx), float64(cy))

	op.Filter = ebiten.FilterLinear
	r := float32(clr.R) / 255.0
	g := float32(clr.G) / 255.0
	bClr := float32(clr.B) / 255.0
	a := (float32(clr.A) / 255.0) * alphaScale
	op.ColorScale.Scale(r, g, bClr, a)

	screen.DrawImage(glowImg, op)
}

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
		body: color.RGBA{16, 26, 52, 240},
		lure: color.RGBA{70, 240, 255, 255},
		glow: color.RGBA{30, 190, 255, 140},
	},
	// Radiant Aqua-Green
	{
		body: color.RGBA{14, 36, 42, 240},
		lure: color.RGBA{50, 255, 185, 255},
		glow: color.RGBA{25, 225, 155, 140},
	},
	// Abyssal Violet
	{
		body: color.RGBA{30, 16, 48, 240},
		lure: color.RGBA{215, 115, 255, 255},
		glow: color.RGBA{170, 65, 245, 140},
	},
	// Deep Amber-Gold
	{
		body: color.RGBA{34, 26, 16, 240},
		lure: color.RGBA{255, 205, 70, 255},
		glow: color.RGBA{240, 165, 30, 140},
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
	lurePulse := float32(math.Sin(f.SwimPhase*3.2+0.8))*0.5 + 0.5

	var dir float32 = 1.0
	var tailX, headX float32
	if f.FacingRight {
		tailX = cx - 7.5
		headX = cx + 5.5
		dir = 1.0
	} else {
		tailX = cx + 7.5
		headX = cx - 5.5
		dir = -1.0
	}

	// 1. Subtle, ethereal body ambient glow (smooth procedural radial light, no flat rings)
	bodyGlowRad := float32(22.0 + 4.0*pulse)
	drawLanternRadialGlow(screen, cx, cy, bodyGlowRad, f.GlowColor, 0.28+0.12*pulse)
	drawLanternRadialGlow(screen, cx+dir*1.5, cy, bodyGlowRad*0.55, f.GlowColor, 0.45+0.15*pulse)

	// 2. Esca & dorsal lantern position calculation
	lureBaseX := cx + 1.0*dir
	lureBaseY := cy - 4.5
	lureTipX := headX + 3.2*dir
	lureTipY := cy - 5.5 + float32(math.Sin(float64(f.SwimPhase)*1.5))*1.2

	// 3. Glowing dorsal lure bulb flare (procedural soft bloom flare)
	lureBloomRad := float32(14.0 + 3.5*lurePulse)
	drawLanternRadialGlow(screen, lureTipX, lureTipY, lureBloomRad, f.LureColor, 0.35+0.18*lurePulse)
	drawLanternRadialGlow(screen, lureTipX, lureTipY, lureBloomRad*0.5, f.LureColor, 0.70+0.25*lurePulse)

	// 4. Ventral photophore soft micro-glows along belly
	photoClr := color.RGBA{f.LureColor.R, f.LureColor.G, f.LureColor.B, uint8(190 + 65*pulse)}
	for i := float32(-2.5); i <= 2.5; i += 2.5 {
		px := cx + i*dir
		py := cy + 3.8
		drawLanternRadialGlow(screen, px, py, 4.5+1.0*pulse, f.LureColor, 0.40+0.20*pulse)
	}

	// 5. Translucent deep fish body
	vector.FillCircle(screen, cx, cy, 5.5, f.BodyColor, true)
	vector.FillCircle(screen, cx+dir*2.0, cy, 4.5, f.BodyColor, true)
	vector.FillCircle(screen, cx-dir*2.5, cy, 3.8, f.BodyColor, true)

	// 6. Tail fin with sinusoidal undulation
	wiggle := float32(math.Sin(float64(f.SwimPhase))) * 2.5
	tailClr := color.RGBA{
		uint8(min(255, int(f.BodyColor.R)+25)),
		uint8(min(255, int(f.BodyColor.G)+35)),
		uint8(min(255, int(f.BodyColor.B)+45)),
		190,
	}
	vector.StrokeLine(screen, tailX, cy, tailX-4*dir+wiggle, cy-4.5, 1.6, tailClr, true)
	vector.StrokeLine(screen, tailX, cy, tailX-4*dir+wiggle, cy+4.5, 1.6, tailClr, true)
	vector.StrokeLine(screen, tailX, cy, tailX-5*dir+wiggle*1.2, cy, 1.2, tailClr, true)

	// 7. Pectoral fin flutter
	pecWiggle := float32(math.Cos(float64(f.SwimPhase)*2.0)) * 1.5
	pecClr := color.RGBA{tailClr.R, tailClr.G, tailClr.B, 150}
	vector.StrokeLine(screen, cx-dir*0.5, cy+1.0, cx-dir*3.0, cy+3.5+pecWiggle, 1.2, pecClr, true)

	// 8. Ventral photophore light organ cores
	for i := float32(-2.5); i <= 2.5; i += 2.5 {
		px := cx + i*dir
		py := cy + 3.8
		vector.FillCircle(screen, px, py, 0.9, photoClr, true)
		vector.FillCircle(screen, px, py, 0.4, color.White, true)
	}

	// 9. Dorsal antenna stem (thin curved esca stalk)
	antennaClr := color.RGBA{110, 150, 195, 220}
	midAntennaX := (lureBaseX + lureTipX) / 2.0
	midAntennaY := lureBaseY - 3.0
	vector.StrokeLine(screen, lureBaseX, lureBaseY, midAntennaX, midAntennaY, 1.0, antennaClr, true)
	vector.StrokeLine(screen, midAntennaX, midAntennaY, lureTipX, lureTipY, 1.0, antennaClr, true)

	// 10. Esca lure core bulb & intense radiant point
	vector.FillCircle(screen, lureTipX, lureTipY, 2.0, f.LureColor, true)
	vector.FillCircle(screen, lureTipX, lureTipY, 1.0, color.RGBA{255, 255, 255, 240}, true)

	// 11. Large dark-adapted abyssal eye with bioluminescent glint
	eyeX := headX - 1.2*dir
	vector.FillCircle(screen, eyeX, cy-1.0, 1.9, color.RGBA{200, 230, 250, 255}, true)
	vector.FillCircle(screen, eyeX, cy-1.0, 1.1, color.RGBA{10, 12, 22, 255}, true)
	vector.FillCircle(screen, eyeX-dir*0.4, cy-1.4, 0.5, color.White, true)
}

// PointLight returns the dynamic point light emitted by the lanternfish's lure for the cave lighting shader.
func (f *Lanternfish) PointLight() (pos gvec.Vec2, radius float64, r, g, b float32, intensity float64) {
	if !f.Active {
		return gvec.Vec2{}, 0, 0, 0, 0, 0
	}
	dir := 1.0
	headX := f.Pos.X + f.Dimensions.X/2.0 + 5.5
	if !f.FacingRight {
		dir = -1.0
		headX = f.Pos.X + f.Dimensions.X/2.0 - 5.5
	}
	lureTipX := headX + 3.2*dir
	lureTipY := f.Pos.Y + f.Dimensions.Y/2.0 - 5.5 + math.Sin(f.SwimPhase*1.5)*1.2
	pulse := math.Sin(f.SwimPhase*2.5)*0.5 + 0.5

	radius = 42.0 + 8.0*pulse
	r = float32(f.LureColor.R) / 255.0
	g = float32(f.LureColor.G) / 255.0
	b = float32(f.LureColor.B) / 255.0
	intensity = 0.60
	return gvec.Vec2{X: lureTipX, Y: lureTipY}, radius, r, g, b, intensity
}

