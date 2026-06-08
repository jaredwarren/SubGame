package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// ShatterBulb is a static oxygen plant that pops when touched, restoring O2.
type ShatterBulb struct {
	BaseEntity
}

func (s *ShatterBulb) Update(gr Runtime) {
	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}
	if rectsOverlap(s.Pos.X, s.Pos.Y, s.Dimensions.X, s.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		s.Pop(gr)
	}
}

// Pop deactivates the bulb, restoring oxygen and emitting a sound wave.
func (s *ShatterBulb) Pop(gr Runtime) {
	if !s.Active {
		return
	}
	s.Active = false
	gr.Emit(RestoreOxygenCmd{Amount: 20})
	gr.Emit(TriggerSoundWaveCmd{
		Pos: gvec.Vec2{X: s.Pos.X + s.Dimensions.X/2.0, Y: s.Pos.Y + s.Dimensions.Y/2.0},
	})
}

func (s *ShatterBulb) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(s.Pos.X - camera.Pos.X)
	sy := float32(s.Pos.Y - camera.Pos.Y)
	sw := float32(s.Dimensions.X)
	sh := float32(s.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	vector.StrokeLine(screen, cx, cy, cx, cy+16, 2.0, color.RGBA{45, 95, 75, 255}, false)
	phase := s.Pos.X + s.Pos.Y
	pulse := float32(math.Cos(timeOfDay*0.02+phase)) * 2.5
	radius := float32(11.0) + pulse
	if radius < 5.0 {
		radius = 5.0
	}
	vector.FillCircle(screen, cx, cy, radius, color.RGBA{0, 220, 240, 60}, false)
	vector.FillCircle(screen, cx, cy, 7, color.RGBA{0, 230, 245, 255}, false)
	vector.StrokeCircle(screen, cx, cy, 7, 0.8, color.RGBA{255, 255, 255, 200}, false)
}
