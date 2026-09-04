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
	def        *ShatterBulbDef
	SwayPhase  float64
	AnchorSide string // "floor", "left", "right"
}

func (s *ShatterBulb) stats() *ShatterBulbDef {
	if s.def != nil {
		return s.def
	}
	return ShatterBulbArchetype
}

// NewShatterBulb creates a ShatterBulb at the given position with floor anchor.
func NewShatterBulb(x, y float64, height ...float64) *ShatterBulb {
	h := float64(44.0)
	if len(height) > 0 && height[0] > 0 {
		h = height[0]
	}
	return NewShatterBulbAnchored(x, y, h, "floor")
}

// NewShatterBulbAnchored creates a ShatterBulb with an explicit anchor ("floor", "left", "right").
func NewShatterBulbAnchored(x, y, height float64, anchor string) *ShatterBulb {
	d := ShatterBulbArchetype
	w := d.FloorWidth
	if w <= 0 {
		w = d.Dims.X
	}
	if anchor == "left" || anchor == "right" {
		if d.WallWidth > 0 {
			w = d.WallWidth
		} else {
			w = 28.0
		}
	} else {
		anchor = "floor"
	}
	if height <= 0 {
		height = 44.0
	}
	return &ShatterBulb{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: w, Y: height},
			Active:     true,
		},
		def:        d,
		SwayPhase:  (x*0.07 + y*0.13),
		AnchorSide: anchor,
	}
}

func (s *ShatterBulb) Update(gr Runtime) {
	s.SwayPhase += 0.03
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
	gr.Emit(RestoreOxygenCmd{Amount: s.stats().RestoreOxygen})
	gr.Emit(TriggerSoundWaveCmd{
		Pos: gvec.Vec2{X: s.Pos.X + s.Dimensions.X/2.0, Y: s.Pos.Y + s.Dimensions.Y/2.0},
	})
}

func (s *ShatterBulb) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(s.Pos.X - camera.Pos.X)
	sy := float32(s.Pos.Y - camera.Pos.Y)
	sw := float32(s.Dimensions.X)
	sh := float32(s.Dimensions.Y)

	// Determine anchor point
	var baseX, baseY float32
	if s.AnchorSide == "left" {
		baseX = sx
		baseY = sy + sh
	} else if s.AnchorSide == "right" {
		baseX = sx + sw
		baseY = sy + sh
	} else { // "floor" or empty fallback
		baseX = sx + sw/2.0
		baseY = sy + sh + 1.0
	}

	// Dynamic swaying phase matching normal kelp
	swayPhase := s.SwayPhase
	if swayPhase == 0 {
		swayPhase = float64(s.Pos.X*0.07 + s.Pos.Y*0.13)
	}

	numSegments := int(sh / 8.0)
	if numSegments < 3 {
		numSegments = 3
	}
	segmentHeight := sh / float32(numSegments)

	lastX := baseX
	lastY := baseY

	for i := 0; i < numSegments; i++ {
		factor := float32(i+1) / float32(numSegments)
		swayOffset := float32(math.Sin(swayPhase+float64(i)*0.4)) * 8.0 * factor
		nextY := baseY - float32(i+1)*segmentHeight

		var nextX float32
		if s.AnchorSide == "left" {
			// Curves to the right (away from left wall), then goes straight up
			curveOffset := sw * float32(math.Sin(float64(factor)*math.Pi/2.0))
			nextX = baseX + curveOffset + swayOffset
		} else if s.AnchorSide == "right" {
			// Curves to the left (away from right wall), then goes straight up
			curveOffset := sw * float32(math.Sin(float64(factor)*math.Pi/2.0))
			nextX = baseX - curveOffset + swayOffset
		} else {
			// Floor anchor (grows straight up)
			nextX = baseX + swayOffset
		}

		// Kelp stalk stroke
		vector.StrokeLine(screen, lastX, lastY, nextX, nextY, 2.5-factor*1.0, color.RGBA{34, 139, 34, 255}, false)

		// Kelp leaf fronds: Clean holdfast skips the anchor segment against the wall
		if s.AnchorSide == "floor" || i > 0 {
			leafSize := 5.0 - factor*2.0
			if leafSize < 2.0 {
				leafSize = 2.0
			}
			vector.FillCircle(screen, nextX-4.0, nextY, leafSize, color.RGBA{46, 150, 60, 220}, false)
			vector.FillCircle(screen, nextX+4.0, nextY, leafSize, color.RGBA{46, 150, 60, 220}, false)
		}

		lastX = nextX
		lastY = nextY
	}

	// Glowing oxygen bubble resting on top of the kelp stalk
	bulbX := lastX
	bulbY := lastY - 4.0

	phase := s.Pos.X + s.Pos.Y
	pulse := float32(math.Cos(timeOfDay*0.02+phase)) * 2.5
	radius := float32(11.0) + pulse
	if radius < 5.0 {
		radius = 5.0
	}

	// Outer atmospheric soft glow
	vector.FillCircle(screen, bulbX, bulbY, radius, color.RGBA{0, 220, 240, 60}, false)
	// Outer cyan aura rim
	vector.StrokeCircle(screen, bulbX, bulbY, radius*0.8, 1.2, color.RGBA{0, 235, 255, 100}, false)
	// Core bubble body
	vector.FillCircle(screen, bulbX, bulbY, 7.5, color.RGBA{0, 230, 245, 255}, false)
	// Inner glossy shine highlight
	vector.FillCircle(screen, bulbX-2.5, bulbY-2.5, 2.2, color.RGBA{255, 255, 255, 220}, false)
	// Crisp bubble outline
	vector.StrokeCircle(screen, bulbX, bulbY, 7.5, 0.8, color.RGBA{255, 255, 255, 210}, false)
}

// PointLight returns the dynamic cyan point light emitted by the oxygen bubble bulb.
func (s *ShatterBulb) PointLight() (pos gvec.Vec2, radius float64, r, g, b float32, intensity float64) {
	if !s.Active {
		return gvec.Vec2{}, 0, 0, 0, 0, 0
	}
	var cx float64
	switch s.AnchorSide {
	case "left":
		cx = s.Pos.X + s.Dimensions.X
	case "right":
		cx = s.Pos.X
	default:
		cx = s.Pos.X + s.Dimensions.X/2.0
	}
	cy := s.Pos.Y + 4.0
	return gvec.Vec2{X: cx, Y: cy}, 38.0, 0.0, 0.90, 0.96, 0.55
}
