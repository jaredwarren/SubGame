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

// InkSquid is a passive cephalopod that swims slowly, but sprays a blinding, slowing
// cloud of black ink and jet-dashes away when approached by the player.
type InkSquid struct {
	BaseEntity
	def          *FaunaDef
	FacingRight  bool
	SwimPhase    float64
	InkCooldown  int
	FleeTimer    int
	WanderTimer  int
	EscapeDir    gvec.Vec2
	MantleColor  color.RGBA
	FinColor     color.RGBA
	TentacleColor color.RGBA
	EyeColor     color.RGBA
}

func (s *InkSquid) stats() *FaunaDef {
	if s.def != nil {
		return s.def
	}
	return InkSquidArchetype
}

var squidPresets = []struct {
	mantle   color.RGBA
	fin      color.RGBA
	tentacle color.RGBA
	eye      color.RGBA
}{
	// Abyssal Violet (Deep purple & neon magenta)
	{
		mantle:   color.RGBA{135, 55, 175, 255},
		fin:      color.RGBA{170, 75, 215, 230},
		tentacle: color.RGBA{115, 40, 150, 240},
		eye:      color.RGBA{255, 225, 90, 255},
	},
	// Bioluminescent Azure (Cyan & deep ocean indigo)
	{
		mantle:   color.RGBA{45, 150, 195, 255},
		fin:      color.RGBA{75, 195, 240, 230},
		tentacle: color.RGBA{30, 120, 165, 240},
		eye:      color.RGBA{220, 255, 255, 255},
	},
	// Radiant Amber (Golden orange & sunset vermilion)
	{
		mantle:   color.RGBA{225, 110, 50, 255},
		fin:      color.RGBA{255, 155, 80, 230},
		tentacle: color.RGBA{190, 85, 35, 240},
		eye:      color.RGBA{255, 245, 180, 255},
	},
	// Ghostly Rose (Translucent coral pink)
	{
		mantle:   color.RGBA{220, 80, 130, 255},
		fin:      color.RGBA{250, 120, 170, 230},
		tentacle: color.RGBA{185, 55, 100, 240},
		eye:      color.RGBA{140, 255, 230, 255},
	},
}

// NewInkSquid creates a new InkSquid at (x, y).
func NewInkSquid(x, y float64, facingRight bool) *InkSquid {
	d := InkSquidArchetype
	preset := squidPresets[rand.Intn(len(squidPresets))]
	return &InkSquid{
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
		EyeColor:      preset.eye,
	}
}

// Update updates the InkSquid AI, movement, player proximity checking, and ink discharge.
func (s *InkSquid) Update(gr Runtime) {
	d := s.stats()
	s.SwimPhase += d.SwimPhaseSpeed

	if s.InkCooldown > 0 {
		s.InkCooldown--
	}

	centerX := s.Pos.X + s.Dimensions.X/2.0
	centerY := s.Pos.Y + s.Dimensions.Y/2.0

	// Player or active vehicle proximity check
	targetPos := gr.PlayerPos()
	targetDims := gr.PlayerDims()
	if gr.HasActiveVehicle() {
		targetPos = gr.ActiveVehiclePos()
		targetDims = gr.ActiveVehicleDims()
	}
	targetCenterX := targetPos.X + targetDims.X/2.0
	targetCenterY := targetPos.Y + targetDims.Y/2.0

	distToPlayer := math.Hypot(targetCenterX-centerX, targetCenterY-centerY)

	// Trigger ink defense when player gets too close and ink is ready
	if distToPlayer < d.ThreatRange && s.InkCooldown <= 0 {
		// Spawn inky defense cloud at current location
		gr.Emit(SpawnInkCloudCmd{
			Pos: gvec.Vec2{X: centerX, Y: centerY},
		})

		// Escape vector pointing away from player
		dx := centerX - targetCenterX
		dy := centerY - targetCenterY
		dist := math.Hypot(dx, dy)
		if dist < 0.001 {
			dx, dy = 1.0, 0.0
			dist = 1.0
		}
		s.EscapeDir = gvec.Vec2{X: dx / dist, Y: dy / dist}

		// Initial burst of high escape speed
		s.Vel = s.EscapeDir.Scale(d.FleeSpeed)
		s.FleeTimer = d.FleeFrames
		s.InkCooldown = d.CooldownFrames
		s.FacingRight = (s.Vel.X >= 0)
	}

	if s.FleeTimer > 0 {
		s.FleeTimer--
		// High-speed jet propulsion
		s.Vel = s.EscapeDir.Scale(d.FleeSpeed * (0.6 + 0.4*float64(s.FleeTimer)/float64(d.FleeFrames)))
		s.FacingRight = (s.EscapeDir.X >= 0)
	} else {
		// Normal passive wandering swimming
		s.WanderTimer--
		if s.WanderTimer <= 0 {
			s.WanderTimer = rand.Intn(120) + 60
			// Random heading with slight forward bias
			angle := rand.Float64() * math.Pi * 2
			s.Vel = gvec.Vec2{
				X: math.Cos(angle) * d.PatrolSpeed,
				Y: math.Sin(angle) * d.PatrolSpeed * 0.5,
			}
			s.FacingRight = (s.Vel.X >= 0)
		}

		// Gentle pulse in swim speed matching mantle contraction
		pulse := 0.7 + 0.3*math.Sin(s.SwimPhase*2.0)
		s.Vel = gvec.Vec2{
			X: s.Vel.X * pulse,
			Y: s.Vel.Y * pulse,
		}
	}

	// Calculate next position and handle cave wall avoidance/bounce
	nextPos := s.Pos.Add(s.Vel)

	if gr.IsSolid(nextPos.X, nextPos.Y, s.Dimensions.X, s.Dimensions.Y) {
		// Bounce off obstacle
		s.Vel = s.Vel.Scale(-0.7)
		s.EscapeDir = s.EscapeDir.Scale(-1.0)
		s.FacingRight = !s.FacingRight
		s.WanderTimer = rand.Intn(90) + 40
	} else {
		s.Pos = nextPos
	}
}

// Draw renders the InkSquid with animated pulsating mantle, fluttering fins, and waving tentacles.
func (s *InkSquid) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(s.Pos.X - camera.Pos.X)
	sy := float32(s.Pos.Y - camera.Pos.Y)
	sw := float32(s.Dimensions.X)
	sh := float32(s.Dimensions.Y)

	cx := sx + sw/2.0
	cy := sy + sh/2.0

	pulse := float32(math.Sin(s.SwimPhase*2.0)) * 1.2
	isFleeing := (s.FleeTimer > 0)

	// Orientation orientation factor: 1 for right, -1 for left
	dir := float32(1.0)
	if !s.FacingRight {
		dir = -1.0
	}

	// 1. Undulating Tentacles trailing behind the mantle
	tentacleBaseX := cx - dir*7.0
	numTentacles := 5

	for i := 0; i < numTentacles; i++ {
		tOffset := float32(i-2) * 2.2
		tPhase := s.SwimPhase + float64(i)*0.6

		// In jet-dash mode, tentacles stream straight back in hydrodynamic dart formation
		waveAmp := float32(4.5)
		waveFreq := float32(1.0)
		if isFleeing {
			waveAmp = 1.5
			waveFreq = 2.5
		}

		tentacleTipX := tentacleBaseX - dir*(11.0+float32(math.Sin(tPhase))*2.0)
		tentacleTipY := cy + tOffset + float32(math.Sin(tPhase*float64(waveFreq)))*waveAmp

		// Draw tentacle segment
		vector.StrokeLine(screen, tentacleBaseX, cy+tOffset*0.5, tentacleTipX, tentacleTipY, 1.4, s.TentacleColor, false)
		// Glowing tip suction pad
		vector.FillCircle(screen, tentacleTipX, tentacleTipY, 1.0, s.FinColor, false)
	}

	// 2. Cephalopod Mantle Body (torpedo / cone shaped)
	mantleRadiusX := (sw*0.42 + pulse)
	mantleRadiusY := (sh * 0.38)
	if isFleeing {
		mantleRadiusX += 1.5
		mantleRadiusY -= 1.0
	}

	// Mantle main body
	vector.FillCircle(screen, cx+dir*1.0, cy, mantleRadiusY, s.MantleColor, false)
	vector.FillCircle(screen, cx+dir*5.0, cy, mantleRadiusY*0.85, s.MantleColor, false)
	vector.FillCircle(screen, cx-dir*3.0, cy, mantleRadiusY*0.9, s.MantleColor, false)

	// 3. Lateral Mantle Fins (fluttering wings at the top & bottom of head)
	finWave := float32(math.Cos(s.SwimPhase*3.0)) * 2.0
	finX := cx + dir*6.0
	vector.FillCircle(screen, finX, cy-mantleRadiusY+finWave, 3.2, s.FinColor, false)
	vector.FillCircle(screen, finX, cy+mantleRadiusY-finWave, 3.2, s.FinColor, false)

	// 4. Large Luminous Cephalopod Eye
	eyeX := cx + dir*3.5
	eyeY := cy - 1.5
	vector.FillCircle(screen, eyeX, eyeY, 3.0, s.EyeColor, false)
	// Dark pupil slit
	vector.FillCircle(screen, eyeX+dir*0.5, eyeY, 1.4, color.RGBA{12, 10, 18, 255}, false)
	// Eye shine dot
	vector.FillCircle(screen, eyeX-dir*0.5, eyeY-0.8, 0.8, color.RGBA{255, 255, 255, 240}, false)

	// 5. Siphon nozzle underneath head
	siphonX := cx - dir*1.0
	siphonY := cy + mantleRadiusY*0.75
	vector.FillCircle(screen, siphonX, siphonY, 1.8, s.TentacleColor, false)
}
