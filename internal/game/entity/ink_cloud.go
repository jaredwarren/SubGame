package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// InkCloud is a dense expanding cloud of black cephalopod ink that obscures vision and slows the player.
type InkCloud struct {
	BaseEntity
	LifeTimer int
}

const (
	inkCloudMaxLife = 300   // 5 seconds at 60 FPS
	inkCloudMaxRad  = 175.0 // Much larger cloud radius around player
	inkSlowDuration = 180   // 3 seconds of slow
	inkSlowFactor   = 0.5   // 50% movement speed reduction
)

// NewInkCloud creates an InkCloud entity centered at (centerX, centerY).
func NewInkCloud(centerX, centerY float64) *InkCloud {
	return &InkCloud{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: centerX - inkCloudMaxRad, Y: centerY - inkCloudMaxRad},
			Dimensions: gvec.Vec2{X: inkCloudMaxRad * 2, Y: inkCloudMaxRad * 2},
			Active:     true,
		},
		LifeTimer: inkCloudMaxLife,
	}
}

// Update updates the InkCloud entity and checks for player overlap to apply slow.
func (c *InkCloud) Update(gr Runtime) {
	c.LifeTimer--
	if c.LifeTimer <= 0 {
		c.Active = false
		return
	}

	// Check if player or active vehicle is touching the in-water ink cloud
	pPos := gr.PlayerPos()
	pDims := gr.PlayerDims()
	targetX, targetY := pPos.X, pPos.Y
	targetW, targetH := pDims.X, pDims.Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		targetW, targetH = vDims.X, vDims.Y
	}

	// Current effective cloud radius based on expansion
	elapsed := float64(inkCloudMaxLife - c.LifeTimer)
	var radius float64 = inkCloudMaxRad
	if elapsed < 45 {
		radius = 65.0 + (elapsed/45.0)*(inkCloudMaxRad-65.0)
	}

	cx := c.Pos.X + c.Dimensions.X/2.0
	cy := c.Pos.Y + c.Dimensions.Y/2.0

	// Player center distance to ink cloud center
	playerCenterX := targetX + targetW/2.0
	playerCenterY := targetY + targetH/2.0
	dist := math.Hypot(playerCenterX-cx, playerCenterY-cy)

	if dist <= radius+math.Max(targetW, targetH)/2.0 {
		gr.Emit(SlowPlayerCmd{
			Duration: inkSlowDuration,
			Factor:   inkSlowFactor,
		})
	}
}

// Draw renders the billowing translucent cephalopod ink cloud.
func (c *InkCloud) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	sw := float32(c.Dimensions.X)
	sh := float32(c.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	elapsed := float64(inkCloudMaxLife - c.LifeTimer)
	var radius float64 = inkCloudMaxRad
	if elapsed < 45 {
		radius = 65.0 + (elapsed/45.0)*(inkCloudMaxRad-65.0)
	}

	var alpha float64 = 0.42
	if c.LifeTimer < 60 {
		alpha = (float64(c.LifeTimer) / 60.0) * 0.42
	}

	scale := radius / inkCloudMaxRad

	// Outer soft diffuse dispersion haze
	outerColor := color.RGBA{22, 25, 45, uint8(alpha * 55)}
	vector.FillCircle(screen, cx, cy, float32(radius), outerColor, false)

	// Rolling interior organic ink billows (enlarged for grander soft clouds)
	puffs := []struct {
		relX, relY float64
		size       float64
		speed      float64
		phase      float64
	}{
		{relX: 0, relY: 0, size: 115, speed: 0.04, phase: 0.0},
		{relX: -45, relY: -30, size: 95, speed: 0.05, phase: 1.1},
		{relX: 42, relY: -35, size: 88, speed: 0.04, phase: 2.2},
		{relX: -36, relY: 40, size: 92, speed: 0.06, phase: 3.3},
		{relX: 48, relY: 34, size: 84, speed: 0.05, phase: 4.4},
		{relX: 12, relY: -55, size: 76, speed: 0.03, phase: 5.1},
		{relX: -58, relY: 12, size: 80, speed: 0.05, phase: 1.8},
		{relX: 55, relY: -12, size: 85, speed: 0.04, phase: 3.7},
	}

	for _, p := range puffs {
		swayX := math.Sin(timeOfDay*p.speed+p.phase) * 14.0 * scale
		swayY := math.Cos(timeOfDay*p.speed*1.2+p.phase) * 14.0 * scale

		px := cx + float32(p.relX*scale+swayX)
		py := cy + float32(p.relY*scale+swayY)
		pr := float32(p.size * scale)

		// Translucent organic ink puff
		puffColor := color.RGBA{16, 16, 32, uint8(alpha * 95)}
		vector.FillCircle(screen, px, py, pr, puffColor, false)

		// Deep velvet translucent core
		coreColor := color.RGBA{26, 24, 46, uint8(alpha * 65)}
		vector.FillCircle(screen, px-pr*0.1, py-pr*0.1, pr*0.7, coreColor, false)
	}
}
