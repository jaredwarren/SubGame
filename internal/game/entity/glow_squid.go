package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// GlowSquid is a deep-sea bioluminescent cephalopod that illuminates dark abyssal trenches
// and jet-dashes away while spraying a radiant bioluminescent ink cloud when approached.
type GlowSquid struct {
	BaseEntity
	def           *FaunaDef
	FacingRight   bool
	SwimPhase     float64
	InkCooldown   int
	FleeTimer     int
	WanderTimer   int
	EscapeDir     gvec.Vec2
	MantleColor   color.RGBA
	FinColor      color.RGBA
	TentacleColor color.RGBA
	GlowColor     color.RGBA
	EyeColor      color.RGBA
}

var glowSquidPresets = []struct {
	mantle   color.RGBA
	fin      color.RGBA
	tentacle color.RGBA
	glow     color.RGBA
	eye      color.RGBA
}{
	// Electric Cyan / Azure
	{
		mantle:   color.RGBA{25, 210, 245, 255},
		fin:      color.RGBA{80, 240, 255, 230},
		tentacle: color.RGBA{15, 175, 220, 240},
		glow:     color.RGBA{20, 180, 255, 130},
		eye:      color.RGBA{255, 245, 140, 255},
	},
	// Radiant Emerald / Seafoam
	{
		mantle:   color.RGBA{30, 240, 170, 255},
		fin:      color.RGBA{90, 255, 210, 230},
		tentacle: color.RGBA{20, 195, 135, 240},
		glow:     color.RGBA{25, 220, 150, 130},
		eye:      color.RGBA{220, 255, 255, 255},
	},
	// Neon Amethyst / Ultraviolet
	{
		mantle:   color.RGBA{190, 75, 255, 255},
		fin:      color.RGBA{225, 130, 255, 230},
		tentacle: color.RGBA{150, 50, 215, 240},
		glow:     color.RGBA{170, 50, 240, 130},
		eye:      color.RGBA{255, 220, 100, 255},
	},
}

// NewGlowSquid creates a new GlowSquid entity at (x, y).
func NewGlowSquid(x, y float64, facingRight bool) *GlowSquid {
	d := FaunaDefFor(FaunaGlowSquid)
	preset := glowSquidPresets[rand.Intn(len(glowSquidPresets))]
	return &GlowSquid{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: d.Dims,
			Active:     true,
		},
		def:           d,
		FacingRight:   facingRight,
		SwimPhase:     rand.Float64() * math.Pi * 2,
		WanderTimer:   rand.Intn(90) + 45,
		MantleColor:   preset.mantle,
		FinColor:      preset.fin,
		TentacleColor: preset.tentacle,
		GlowColor:     preset.glow,
		EyeColor:      preset.eye,
	}
}

func (s *GlowSquid) stats() *FaunaDef {
	if s.def != nil {
		return s.def
	}
	return FaunaDefFor(FaunaGlowSquid)
}

func (s *GlowSquid) Update(gr Runtime) {
	d := s.stats()
	s.SwimPhase += d.SwimPhaseSpeed

	if s.InkCooldown > 0 {
		s.InkCooldown--
	}

	pPos := gr.PlayerPos()
	pDims := gr.PlayerDims()
	targetX := pPos.X + pDims.X/2.0
	targetY := pPos.Y + pDims.Y/2.0
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		vDims := gr.ActiveVehicleDims()
		targetX = vPos.X + vDims.X/2.0
		targetY = vPos.Y + vDims.Y/2.0
	}

	cx := s.Pos.X + s.Dimensions.X/2.0
	cy := s.Pos.Y + s.Dimensions.Y/2.0
	dx := cx - targetX
	dy := cy - targetY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < d.ThreatRange && s.InkCooldown <= 0 {
		gr.Emit(SpawnInkCloudCmd{Pos: gvec.Vec2{X: cx, Y: cy}})
		s.InkCooldown = d.CooldownFrames
		s.FleeTimer = d.FleeFrames

		if dist > 0.001 {
			s.EscapeDir = gvec.Vec2{X: dx / dist, Y: dy / dist}
		} else {
			angle := rand.Float64() * math.Pi * 2
			s.EscapeDir = gvec.Vec2{X: math.Cos(angle), Y: math.Sin(angle)}
		}
		s.FacingRight = s.EscapeDir.X >= 0
		s.Vel = s.EscapeDir.Scale(d.FleeSpeed)
	}

	if s.FleeTimer > 0 {
		s.FleeTimer--
		fleeProg := float64(s.FleeTimer) / float64(d.FleeFrames)
		speed := d.PatrolSpeed + (d.FleeSpeed-d.PatrolSpeed)*fleeProg
		s.Vel = s.EscapeDir.Scale(speed)
	} else {
		s.WanderTimer--
		if s.WanderTimer <= 0 {
			s.WanderTimer = rand.Intn(90) + 40
			angle := rand.Float64() * math.Pi * 2
			s.Vel = gvec.Vec2{
				X: math.Cos(angle) * d.PatrolSpeed,
				Y: math.Sin(angle) * (d.PatrolSpeed * 0.6),
			}
			s.FacingRight = s.Vel.X >= 0
		}
	}

	nextPos := s.Pos.Add(s.Vel)
	if gr.IsSolid(nextPos.X, nextPos.Y, s.Dimensions.X, s.Dimensions.Y) {
		s.Vel = s.Vel.Scale(-0.7)
		s.EscapeDir = s.EscapeDir.Scale(-1.0)
		s.FacingRight = !s.FacingRight
		s.WanderTimer = rand.Intn(90) + 40
	} else {
		s.Pos = nextPos
	}
}

func (s *GlowSquid) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(s.Pos.X - camera.Pos.X)
	sy := float32(s.Pos.Y - camera.Pos.Y)
	sw := float32(s.Dimensions.X)
	sh := float32(s.Dimensions.Y)

	cx := sx + sw/2.0
	cy := sy + sh/2.0

	pulse := float32(math.Sin(s.SwimPhase*2.2)) * 1.5
	pulseNorm := float32(math.Sin(s.SwimPhase*2.0))*0.5 + 0.5
	isFleeing := (s.FleeTimer > 0)

	dir := float32(1.0)
	if !s.FacingRight {
		dir = -1.0
	}

	// 1. Radiant Atmospheric Bioluminescent Halo (transparent soft radial falloff)
	baseR, baseG, baseB := s.GlowColor.R, s.GlowColor.G, s.GlowColor.B
	maxRadius := float32(24.0 + 6.0*pulseNorm)
	// Outer faint haze (soft perimeter)
	vector.FillCircle(screen, cx, cy, maxRadius, color.RGBA{baseR, baseG, baseB, uint8(12 + 8*pulseNorm)}, false)
	// Mid ambient radiance
	vector.FillCircle(screen, cx, cy, maxRadius*0.62, color.RGBA{baseR, baseG, baseB, uint8(22 + 12*pulseNorm)}, false)
	// Inner body bloom
	vector.FillCircle(screen, cx, cy, maxRadius*0.35, color.RGBA{baseR, baseG, baseB, uint8(36 + 18*pulseNorm)}, false)

	// 2. Undulating Bioluminescent Tentacles
	tentacleBaseX := cx - dir*8.0
	numTentacles := 5

	for i := 0; i < numTentacles; i++ {
		tOffset := float32(i-2) * 2.4
		tPhase := s.SwimPhase + float64(i)*0.6

		waveAmp := float32(5.0)
		waveFreq := float32(1.0)
		if isFleeing {
			waveAmp = 1.6
			waveFreq = 2.8
		}

		tentacleTipX := tentacleBaseX - dir*(13.0+float32(math.Sin(tPhase))*2.5)
		tentacleTipY := cy + tOffset + float32(math.Sin(tPhase*float64(waveFreq)))*waveAmp

		vector.StrokeLine(screen, tentacleBaseX, cy+tOffset*0.5, tentacleTipX, tentacleTipY, 1.6, s.TentacleColor, false)
		// Glowing luminous suction pad
		vector.FillCircle(screen, tentacleTipX, tentacleTipY, 1.4, s.FinColor, false)
		vector.FillCircle(screen, tentacleTipX, tentacleTipY, 0.6, color.White, false)
	}

	// 3. Glowing Cephalopod Mantle Body
	mantleRadiusX := (sw*0.44 + pulse)
	mantleRadiusY := (sh * 0.40)
	if isFleeing {
		mantleRadiusX += 1.8
		mantleRadiusY -= 1.0
	}

	vector.FillCircle(screen, cx+dir*1.0, cy, mantleRadiusY, s.MantleColor, false)
	vector.FillCircle(screen, cx+dir*6.0, cy, mantleRadiusY*0.85, s.MantleColor, false)
	vector.FillCircle(screen, cx-dir*3.0, cy, mantleRadiusY*0.9, s.MantleColor, false)

	// Bioluminescent photophore spots along mantle
	spotClr := color.RGBA{255, 255, 255, uint8(180 + 75*pulseNorm)}
	for ox := float32(-2); ox <= 4; ox += 3 {
		vector.FillCircle(screen, cx+dir*ox, cy-mantleRadiusY*0.4, 0.9, spotClr, false)
		vector.FillCircle(screen, cx+dir*ox, cy+mantleRadiusY*0.4, 0.9, spotClr, false)
	}

	// 4. Lateral Fluttering Mantle Fins
	finWave := float32(math.Cos(s.SwimPhase*3.2)) * 2.2
	finX := cx + dir*7.0
	vector.FillCircle(screen, finX, cy-mantleRadiusY+finWave, 3.6, s.FinColor, false)
	vector.FillCircle(screen, finX, cy+mantleRadiusY-finWave, 3.6, s.FinColor, false)

	// 5. Large Luminous Golden/Azure Cephalopod Eye
	eyeX := cx + dir*4.0
	eyeY := cy - 1.6
	vector.FillCircle(screen, eyeX, eyeY, 3.4, s.EyeColor, false)
	vector.FillCircle(screen, eyeX+dir*0.5, eyeY, 1.5, color.RGBA{10, 10, 24, 255}, false)
	vector.FillCircle(screen, eyeX-dir*0.5, eyeY-0.8, 0.9, color.White, false)

	// 6. Siphon nozzle
	siphonX := cx - dir*1.0
	siphonY := cy + mantleRadiusY*0.75
	vector.FillCircle(screen, siphonX, siphonY, 2.0, s.TentacleColor, false)
}
