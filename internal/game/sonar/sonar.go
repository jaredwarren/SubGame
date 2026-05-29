package sonar

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

type Sonar struct {
	State       SonarState
	Timer       int
	MaxDuration int
	Radius      float64
	RadiusStep  float64
	SourceX     float64
	SourceY     float64
	Bright      bool
}

func NewSonar() *Sonar {
	return &Sonar{
		State: SonarState{
			RadiusStep: 6.5,
		},
	}
}

func (s *Sonar) Activate(c vehicle.ActivateSonarCmd) {
	s.Timer = c.Pulse.DurationTicks
	s.MaxDuration = c.Pulse.DurationTicks
	s.Radius = 0
	s.RadiusStep = c.Pulse.RadiusStep
	s.SourceX = c.Source.X
	s.SourceY = c.Source.Y
	s.Bright = c.Bright
}

func (s *Sonar) Update() {
	if s.Timer > 0 {
		s.Timer--
		s.Radius += s.RadiusStep
	}
}

func (s *Sonar) Draw(screen *ebiten.Image, camera *camera.Camera) {
	// Sonar ring overlay
	if s.Timer > 0 {
		maxDur := float32(s.MaxDuration)
		if maxDur <= 0 {
			maxDur = 180.0
		}
		alpha := float32(s.Timer) / maxDur
		scx := float32(s.SourceX - camera.Pos.X)
		scy := float32(s.SourceY - camera.Pos.Y)
		r := float32(s.Radius)

		glowColor := color.RGBA{30, 200, 240, uint8(255 * alpha * 0.5)}
		ringColor := color.RGBA{120, 240, 255, uint8(255 * alpha)}
		coreColor := color.RGBA{255, 255, 255, uint8(255 * alpha)}

		if s.Bright {
			// Upgraded: double ring ripple effect, thick bright cyan glow
			glowColor = color.RGBA{0, 245, 255, uint8(255 * alpha * 0.8)}
			ringColor = color.RGBA{100, 235, 255, uint8(255 * alpha)}

			// Wide soft glow bloom
			vector.StrokeCircle(screen, scx, scy, r, 12.0, glowColor, false)
			// Outer thick main ring
			vector.StrokeCircle(screen, scx, scy, r, 3.0, ringColor, false)
			// Hot-white core line
			vector.StrokeCircle(screen, scx, scy, r, 1.5, coreColor, false)

			// Inner ripple trailing slightly behind
			if r > 15.0 {
				innerAlpha := alpha * 0.7
				innerRingColor := color.RGBA{0, 200, 255, uint8(255 * innerAlpha)}
				vector.StrokeCircle(screen, scx, scy, r-12.0, 1.5, innerRingColor, false)
			}
		} else {
			// Tight soft glow bloom
			vector.StrokeCircle(screen, scx, scy, r, 5.0, glowColor, false)
			// Bright thin main ring
			vector.StrokeCircle(screen, scx, scy, r, 2.0, ringColor, false)
			// Hot-white core line
			vector.StrokeCircle(screen, scx, scy, r, 1.0, coreColor, false)
		}
	}
}

// SonarState tracks the currently active sonar ping in the cave.
// All fields are zero-value when inactive (Timer == 0 means no active ping).
type SonarState struct {
	Timer       int
	MaxDuration int
	Radius      float64
	RadiusStep  float64
	SourceX     float64
	SourceY     float64
	Bright      bool
}
