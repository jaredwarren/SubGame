package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// DeterrentCloud is a lingering chemical cloud that deters hostile predators.
type DeterrentCloud struct {
	BaseEntity
	LifeTimer int
}

// NewDeterrentCloud creates a DeterrentCloud entity centered at (centerX, centerY).
func NewDeterrentCloud(centerX, centerY float64) *DeterrentCloud {
	return &DeterrentCloud{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: centerX - 96, Y: centerY - 96},
			Dimensions: gvec.Vec2{X: 192, Y: 192},
			Active:     true,
		},
		LifeTimer: 360,
	}
}

// Update updates the DeterrentCloud entity.
func (c *DeterrentCloud) Update(gr Runtime) {
	c.LifeTimer--
	if c.LifeTimer <= 0 {
		c.Active = false
	}
}

// Draw renders the DeterrentCloud with expanding translucent smoke puffs.
func (c *DeterrentCloud) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	sw := float32(c.Dimensions.X)
	sh := float32(c.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	// Calculate expansion and fade
	elapsed := float64(360 - c.LifeTimer)
	var radius float64
	if elapsed < 60 {
		radius = (elapsed / 60.0) * 96.0
	} else {
		radius = 96.0
	}

	var alpha float64 = 0.5
	if c.LifeTimer < 60 {
		alpha = (float64(c.LifeTimer) / 60.0) * 0.5
	}

	scale := radius / 96.0

	// Draw large outer translucent dark purple circle
	outerColor := color.RGBA{48, 20, 72, uint8(alpha * 120)}
	vector.FillCircle(screen, cx, cy, float32(radius), outerColor, false)

	// Draw multiple smaller interior smoke puff circles swaying based on time
	puffs := []struct {
		relX, relY float64
		size       float64
		speed      float64
		phase      float64
	}{
		{relX: -30, relY: -20, size: 40, speed: 0.05, phase: 0},
		{relX: 25, relY: -35, size: 35, speed: 0.04, phase: 1.2},
		{relX: -20, relY: 30, size: 45, speed: 0.06, phase: 2.5},
		{relX: 35, relY: 20, size: 30, speed: 0.03, phase: 3.8},
		{relX: 5, relY: -5, size: 50, speed: 0.05, phase: 4.1},
		{relX: -40, relY: 10, size: 28, speed: 0.07, phase: 5.3},
	}

	for _, p := range puffs {
		swayX := math.Sin(timeOfDay*p.speed+p.phase) * 12.0 * scale
		swayY := math.Cos(timeOfDay*p.speed*1.1+p.phase) * 12.0 * scale

		px := cx + float32(p.relX*scale+swayX)
		py := cy + float32(p.relY*scale+swayY)
		pr := float32(p.size * scale)

		// Draw translucent puff
		puffColor := color.RGBA{32, 12, 52, uint8(alpha * 200)}
		vector.FillCircle(screen, px, py, pr, puffColor, false)

		// Inner glow
		glowColor := color.RGBA{80, 40, 110, uint8(alpha * 100)}
		vector.FillCircle(screen, px-pr*0.1, py-pr*0.1, pr*0.7, glowColor, false)
	}
}
