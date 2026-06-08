package overworld

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/light"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// VentContext defines the interface needed by ThermalVent to interact with the game scene.
type VentContext interface {
	GetTicks() float64
	SpawnBubble(x, y float64)
	TriggerScreenShake(duration int, intensity float64)
	GetPlayer() *player.Player
	GetActiveVehicle() vehicle.Vehicle
}


type VentState int

const (
	VentDormant VentState = iota
	VentWarning
	VentErupting
)

// ThermalVent represents a volcanic/hydrothermal vent dealing damage & pushing things away.
type ThermalVent struct {
	Pos            gvec.Vec2
	Radius         float64
	BubbleCooldown int
	State          VentState
	StateTimer     int
	SeedOffset     int64
	Intensity      float64
}

func NewThermalVent(pos gvec.Vec2, seedOffset int64) *ThermalVent {
	r := rand.New(rand.NewSource(seedOffset))
	return &ThermalVent{
		Pos:            pos,
		Radius:         70.0,
		State:          VentDormant,
		StateTimer:     r.Intn(300) + 120, // stagger initial transitions
		SeedOffset:     seedOffset,
		BubbleCooldown: r.Intn(10) + 5,
	}
}

// Update ticks the thermal vent bubble particle spawn rate and geyser state machine.
func (v *ThermalVent) Update(g VentContext) {
	p := g.GetPlayer()
	isPiloting := false
	var targetCenter gvec.Vec2
	var targetDims gvec.Vec2

	activeVeh := g.GetActiveVehicle()
	if activeVeh != nil {
		isPiloting = true
		vPos := activeVeh.GetPos()
		targetDims = activeVeh.GetDimensions()
		targetCenter = gvec.Vec2{X: vPos.X + targetDims.X/2.0, Y: vPos.Y + targetDims.Y/2.0}
	} else {
		targetCenter = gvec.Vec2{X: p.Pos.X + p.Width/2.0, Y: p.Pos.Y + p.Height/2.0}
		targetDims = gvec.Vec2{X: p.Width, Y: p.Height}
	}

	// Tick down StateTimer and handle transitions
	if v.StateTimer <= 0 {
		r := rand.New(rand.NewSource(int64(g.GetTicks()) + v.SeedOffset))
		switch v.State {
		case VentDormant:
			v.State = VentWarning
			v.StateTimer = r.Intn(31) + 60 // 60-90 ticks (1 - 1.5 seconds)
		case VentWarning:
			v.State = VentErupting
			v.StateTimer = r.Intn(61) + 120 // 120-180 ticks (2 - 3 seconds)
			v.Intensity = 1.2               // INSTANT ERUPTION BURST!
		case VentErupting:
			v.State = VentDormant
			v.StateTimer = r.Intn(221) + 180 // 180-400 ticks (3 - 6.6 seconds)
		}
	}
	v.StateTimer--

	// Smoothly transition intensity based on current state
	switch v.State {
	case VentDormant:
		v.Intensity += (0.0 - v.Intensity) * 0.03 // slowly fade to dormant
	case VentWarning:
		target := 0.4 + 0.1*math.Sin(float64(g.GetTicks())*0.2)
		v.Intensity += (target - v.Intensity) * 0.08 // quick pulse warning
	case VentErupting:
		v.Intensity += (0.5 - v.Intensity) * 0.01 // slowly fade/decay during eruption
	}

	// Spawn rising bubble particles based on state
	v.BubbleCooldown--
	if v.BubbleCooldown <= 0 {
		switch v.State {
		case VentDormant:
			v.BubbleCooldown = rand.Intn(40) + 40 // very low bubbles
		case VentWarning:
			v.BubbleCooldown = rand.Intn(15) + 10 // moderate bubbles
		case VentErupting:
			v.BubbleCooldown = rand.Intn(4) + 2 // constant bubble eruption!
		}

		angle := rand.Float64() * 2.0 * math.Pi
		dist := rand.Float64() * 12.0
		bx := v.Pos.X + math.Cos(angle)*dist
		by := v.Pos.Y + math.Sin(angle)*dist
		g.SpawnBubble(bx, by)
	}

	// Calculate distance to player/vehicle
	dx := targetCenter.X - v.Pos.X
	dy := targetCenter.Y - v.Pos.Y
	dist := math.Hypot(dx, dy)

	// Screen rumble during Warning state
	if v.State == VentWarning && dist < v.Radius {
		g.TriggerScreenShake(1, 0.2)
	}

	// Push forces and damage ONLY apply during Erupting state
	if v.State == VentErupting && dist < v.Radius {
		// Calculate outward push force (stronger closer to center, scaled by intensity)
		ratio := 1.0 - (dist / v.Radius)
		intensityScale := math.Max(0.0, math.Min(1.0, v.Intensity))
		pushStrength := 1.8 * ratio * intensityScale

		var pushX, pushY float64
		if dist > 0.1 {
			pushX = (dx / dist) * pushStrength
			pushY = (dy / dist) * pushStrength
		} else {
			pushX = pushStrength
			pushY = 0
		}

		if isPiloting {
			activeVeh := g.GetActiveVehicle()
			// Apply physical force to the vehicle
			activeVeh.ApplyForce(gvec.Vec2{X: pushX * 0.35, Y: pushY * 0.35})
			// Deal continuous structural damage to the vehicle
			activeVeh.TakeDamage(0.06 * intensityScale)
		} else {
			// Apply force to player velocity
			p := g.GetPlayer()
			p.Vel.X += pushX
			p.Vel.Y += pushY
			// Deal damage to swimming player
			p.CurrentHealth -= 0.15 * intensityScale
		}

		// Trigger visual screen shake if close, scaled by intensity
		if dist < 40.0 {
			g.TriggerScreenShake(1, 1.2*intensityScale)
		}
	}
}

// Draw renders a glowing circular volcanic mouth on the seafloor with states.
func (v *ThermalVent) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, mult float64) {
	sx := float32(v.Pos.X - camX)
	sy := float32(v.Pos.Y - camY)

	drawIntensity := v.Intensity
	if v.State == VentWarning {
		// Add rapid warning flash to drawIntensity
		flash := (int(ticks) % 16) < 8
		if flash {
			drawIntensity += 0.25
		}
	}

	// Clamp drawIntensity to [0, 1] for color/size interpolations
	t := math.Max(0.0, math.Min(1.0, drawIntensity))

	// Interpolate pulse speed and amplitude based on state
	var pulseSpeed, pulseAmp float64
	switch v.State {
	case VentDormant:
		pulseSpeed = 0.02
		pulseAmp = 0.8
	case VentWarning:
		pulseSpeed = 0.12
		pulseAmp = 1.5
	case VentErupting:
		pulseSpeed = 0.25
		pulseAmp = 3.5
	}
	pulse := float32(math.Sin(ticks*pulseSpeed)) * float32(pulseAmp)
	outerRad := float32(16.0) + float32(8.0*t) + pulse

	// Define colors for dormant (t = 0) vs fully erupted (t = 1)
	dormantRed := color.RGBA{70, 15, 5, 120}
	eruptedRed := color.RGBA{220, 55, 10, 220}

	dormantOrange := color.RGBA{90, 25, 5, 150}
	eruptedOrange := color.RGBA{255, 145, 25, 255}

	dormantYellow := color.RGBA{110, 45, 10, 180}
	eruptedYellow := color.RGBA{255, 230, 70, 255}

	dormantBlack := color.RGBA{8, 4, 5, 255}
	eruptedBlack := color.RGBA{18, 8, 10, 255}

	// Interpolate and apply lighting multiplier
	glowRed := light.ApplyLight(lerpColor(dormantRed, eruptedRed, t), mult)
	glowOrange := light.ApplyLight(lerpColor(dormantOrange, eruptedOrange, t), mult)
	glowYellow := light.ApplyLight(lerpColor(dormantYellow, eruptedYellow, t), mult)
	abyssBlack := light.ApplyLight(lerpColor(dormantBlack, eruptedBlack, t), mult)

	vector.FillCircle(screen, sx, sy, outerRad, glowRed, false)
	vector.FillCircle(screen, sx, sy, outerRad-4.0, glowOrange, false)
	vector.FillCircle(screen, sx, sy, outerRad-8.0, glowYellow, false)
	vector.FillCircle(screen, sx, sy, outerRad-12.0, abyssBlack, false)
}

func lerpColor(c1, c2 color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(c1.R) + float64(int(c2.R)-int(c1.R))*t),
		G: uint8(float64(c1.G) + float64(int(c2.G)-int(c1.G))*t),
		B: uint8(float64(c1.B) + float64(int(c2.B)-int(c1.B))*t),
		A: uint8(float64(c1.A) + float64(int(c2.A)-int(c1.A))*t),
	}
}
