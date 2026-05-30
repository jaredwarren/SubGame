package vehicle

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Skiff is the starting surface boat — solar-powered, surface-only.
type Skiff struct {
	Pos        gvec.Vec2
	Vel        gvec.Vec2
	Dimensions gvec.Vec2
	Facing     float64
	Health     float64
	MaxHealth  float64
	Battery    float64
	MaxBattery float64
	Cargo      *item.Inventory
}

// NewSkiff creates a Skiff at the given world position.
func NewSkiff(x, y float64) *Skiff {
	return &Skiff{
		Pos:        gvec.Vec2{X: x, Y: y},
		Dimensions: gvec.Vec2{X: 56, Y: 24},
		Health:     150.0,
		MaxHealth:  150.0,
		Battery:    100.0,
		MaxBattery: 100.0,
		Cargo:      item.NewInventory(24),
	}
}

func (s *Skiff) GetPos() gvec.Vec2            { return s.Pos }
func (s *Skiff) SetPos(pos gvec.Vec2)         { s.Pos = pos }
func (s *Skiff) GetDimensions() gvec.Vec2     { return s.Dimensions }
func (s *Skiff) GetHealth() float64           { return s.Health }
func (s *Skiff) GetMaxHealth() float64        { return s.MaxHealth }
func (s *Skiff) GetOxygen() float64           { return 100.0 }
func (s *Skiff) GetDepthLimit() float64       { return 0.0 }
func (s *Skiff) GetCargo() *item.Inventory    { return s.Cargo }
func (s *Skiff) GetUpgrades() *item.Inventory { return nil }
func (s *Skiff) GetPerspective() string       { return "overworld" }
func (s *Skiff) GetName() string              { return "The Skiff" }
func (s *Skiff) GetBattery() float64          { return s.Battery }
func (s *Skiff) GetMaxBattery() float64       { return s.MaxBattery }
func (s *Skiff) GetFacing() float64           { return s.Facing }

func (s *Skiff) TakeDamage(amount float64) {
	s.Health -= amount
	if s.Health < 0 {
		s.Health = 0
	}
}

func (s *Skiff) RechargeBattery(amount float64) {
	s.Battery += amount
	if s.Battery > s.MaxBattery {
		s.Battery = s.MaxBattery
	}
}

func (s *Skiff) Update(runtime Runtime) {
	isDaytime := runtime.TimeOfDay() < 7200
	if isDaytime {
		s.Battery += 0.05
		if s.Battery > s.MaxBattery {
			s.Battery = s.MaxBattery
		}
	}

	if !runtime.IsActiveVehicle(s) {
		s.Vel = gvec.Vec2{}
		return
	}

	const turnSpeed = 0.04
	input := runtime.Input()
	if input.IsKeyPressed(ebiten.KeyA) || input.IsKeyPressed(ebiten.KeyArrowLeft) {
		s.Facing -= turnSpeed
	}
	if input.IsKeyPressed(ebiten.KeyD) || input.IsKeyPressed(ebiten.KeyArrowRight) {
		s.Facing += turnSpeed
	}

	var accel = 0.15
	var maxSpeed = 5.0

	hasPower := s.Battery > 0
	if !hasPower {
		accel = 0.03
		maxSpeed = 1.2
	}

	moving := false
	if input.IsKeyPressed(ebiten.KeyW) || input.IsKeyPressed(ebiten.KeyArrowUp) {
		s.Vel.X += math.Cos(s.Facing) * accel
		s.Vel.Y += math.Sin(s.Facing) * accel
		moving = true
	} else if input.IsKeyPressed(ebiten.KeyS) || input.IsKeyPressed(ebiten.KeyArrowDown) {
		s.Vel.X -= math.Cos(s.Facing) * (accel * 0.4)
		s.Vel.Y -= math.Sin(s.Facing) * (accel * 0.4)
		moving = true
	}

	if moving && hasPower {
		s.Battery -= 0.02
		if s.Battery < 0 {
			s.Battery = 0
		}
	}

	const drag = 0.94
	s.Vel = s.Vel.Scale(drag)

	speed := s.Vel.Length()
	if speed > maxSpeed {
		s.Vel = s.Vel.Scale(maxSpeed / speed)
	}

	s.checkCollisions(runtime)
}

func (s *Skiff) checkCollisions(runtime Runtime) {
	newX := s.Pos.X + s.Vel.X
	if s.isSolid(runtime, gvec.Vec2{X: newX, Y: s.Pos.Y}) {
		s.Vel.X = 0
	} else {
		s.Pos.X = newX
	}
	newY := s.Pos.Y + s.Vel.Y
	if s.isSolid(runtime, gvec.Vec2{X: s.Pos.X, Y: newY}) {
		s.Vel.Y = 0
	} else {
		s.Pos.Y = newY
	}
}

func (s *Skiff) isSolid(runtime Runtime, pos gvec.Vec2) bool {
	x1 := int(math.Floor(pos.X)) / TileSize
	x2 := int(math.Floor(pos.X+s.Dimensions.X)) / TileSize
	y1 := int(math.Floor(pos.Y)) / TileSize
	y2 := int(math.Floor(pos.Y+s.Dimensions.Y)) / TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if runtime.IsOverworldSolidAt(tx, ty) {
				return true
			}
		}
	}
	return false
}

func (s *Skiff) Draw(screen *ebiten.Image, camX, camY float64) {
	cosF := math.Cos(s.Facing)
	sinF := math.Sin(s.Facing)

	rotatePoint := func(px, py float64) (float32, float32) {
		rx := px*cosF - py*sinF
		ry := px*sinF + py*cosF
		return float32(s.Pos.X + s.Dimensions.X/2.0 + rx - camX), float32(s.Pos.Y + s.Dimensions.Y/2.0 + ry - camY)
	}

	x1, y1 := rotatePoint(28, 0)
	x2, y2 := rotatePoint(14, 12)
	x3, y3 := rotatePoint(-28, 12)
	x4, y4 := rotatePoint(-28, -12)
	x5, y5 := rotatePoint(14, -12)

	hullColor := color.RGBA{220, 230, 240, 255}
	stripeColor := color.RGBA{235, 100, 30, 255}

	drawFilledTriangle(screen, x1, y1, x2, y2, x3, y3, hullColor)
	drawFilledTriangle(screen, x1, y1, x3, y3, x4, y4, hullColor)
	drawFilledTriangle(screen, x1, y1, x4, y4, x5, y5, hullColor)

	vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x2, y2, x3, y3, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x3, y3, x4, y4, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x4, y4, x5, y5, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x5, y5, x1, y1, 1.5, stripeColor, false)

	cx1, cy1 := rotatePoint(10, 0)
	cx2, cy2 := rotatePoint(-4, 7)
	cx3, cy3 := rotatePoint(-16, 7)
	cx4, cy4 := rotatePoint(-16, -7)
	cx5, cy5 := rotatePoint(-4, -7)

	cabinColor := color.RGBA{40, 80, 110, 255}
	drawFilledTriangle(screen, cx1, cy1, cx2, cy2, cx3, cy3, cabinColor)
	drawFilledTriangle(screen, cx1, cy1, cx3, cy3, cx4, cy4, cabinColor)
	drawFilledTriangle(screen, cx1, cy1, cx4, cy4, cx5, cy5, cabinColor)

	sp1x, sp1y := rotatePoint(-10, -5)
	sp2x, sp2y := rotatePoint(-24, -5)
	sp3x, sp3y := rotatePoint(-24, 5)
	sp4x, sp4y := rotatePoint(-10, 5)

	solarColor := color.RGBA{15, 120, 215, 255}
	drawFilledTriangle(screen, sp1x, sp1y, sp2x, sp2y, sp3x, sp3y, solarColor)
	drawFilledTriangle(screen, sp1x, sp1y, sp3x, sp3y, sp4x, sp4y, solarColor)
	vector.StrokeLine(screen, sp1x, sp1y, sp2x, sp2y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp2x, sp2y, sp3x, sp3y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp3x, sp3y, sp4x, sp4y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp4x, sp4y, sp1x, sp1y, 0.8, color.RGBA{220, 240, 255, 180}, false)
}
