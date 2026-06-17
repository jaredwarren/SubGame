package vehicle

import (
	"image"
	"image/color"
	_ "image/png"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type SkiffWakeStyle int

const (
	WakeStyleVLines    SkiffWakeStyle = 0 // Option B: Continuous V-Wake Lines
	WakeStyleArcs      SkiffWakeStyle = 1 // Option A: Directional Wave Arcs
	WakeStyleVSegments SkiffWakeStyle = 2 // Option C: Dynamic V-Line Segments
)

const activeWakeStyle = WakeStyleVSegments

var (
	skiffSheet *ebiten.Image
)

type skiffWakePoint struct {
	x, y      float64
	facing    float64
	amplitude float64
	life      float64 // 1.0 -> 0.0
}

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
	wake       []skiffWakePoint
	spawnTimer int
	lightMult  float64
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
		spawnTimer: 0,
		lightMult:  1.0,
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
func (s *Skiff) ApplyForce(force gvec.Vec2) {
	s.Vel = s.Vel.Add(force)
}
func (s *Skiff) GetKit() item.Item { return &SkiffKit{} }

// SkiffKit represents the deployable kit for the Skiff.
type SkiffKit struct{}

func (k *SkiffKit) GetName() string       { return "Skiff Kit" }
func (k *SkiffKit) GetMaxStack() int      { return 1 }
func (k *SkiffKit) GetColor() color.Color { return color.RGBA{235, 100, 30, 255} }
func (k *SkiffKit) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if item.DrawItemIconSprite(screen, k.GetName(), cx, cy, size) {
		return
	}
	// Vector fallback for Skiff Kit (small orange boat silhouette)
	vector.FillRect(screen, cx-size/2.0, cy-size/8.0, size, size/4.0, k.GetColor(), false)
	vector.FillCircle(screen, cx, cy-size/8.0, size/4.0, color.RGBA{40, 80, 110, 255}, false)
}
func (k *SkiffKit) IsPlayerUpgrade() bool { return false }
func (k *SkiffKit) Deploy(x, y float64) Vehicle {
	return NewSkiff(x, y)
}


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
	// Update wake points
	var activeWake []skiffWakePoint
	decayRate := 0.025
	if activeWakeStyle == WakeStyleArcs || activeWakeStyle == WakeStyleVSegments {
		decayRate = 0.015
	}

	for i := range s.wake {
		s.wake[i].life -= decayRate
		if s.wake[i].life > 0 {
			activeWake = append(activeWake, s.wake[i])
		}
	}
	s.wake = activeWake

	// Update light multiplier
	s.lightMult = s.getLightMultiplier(runtime.TimeOfDay())

	isDaytime := runtime.TimeOfDay() < 10800
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

	if runtime.PlayerStunned() {
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

	var accel = 0.20
	var maxSpeed = 6.0

	hasPower := s.Battery > 0
	if !hasPower {
		accel = 0.04
		maxSpeed = 1.5
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

	// Spawn new wake point when moving
	if moving && speed > 0.4 {
		spawnInterval := 1
		if activeWakeStyle == WakeStyleArcs || activeWakeStyle == WakeStyleVSegments {
			spawnInterval = 4
		}
		s.spawnTimer++
		if s.spawnTimer >= spawnInterval {
			s.spawnTimer = 0

			cosF := math.Cos(s.Facing)
			sinF := math.Sin(s.Facing)
			cx := s.Pos.X + s.Dimensions.X/2.0
			cy := s.Pos.Y + s.Dimensions.Y/2.0

			// Center the spawn point exactly at the bow (front tip) of the skiff
			frontX := cx + cosF*28.0
			frontY := cy + sinF*28.0

			s.wake = append(s.wake, skiffWakePoint{
				x:         frontX,
				y:         frontY,
				facing:    s.Facing,
				amplitude: speed * 25.0, // speed determines intensity
				life:      1.0,
			})
		}
	}
}

func (s *Skiff) checkCollisions(runtime Runtime) {
	bPos, bSize := runtime.BaseStationPos()
	hasBase := bSize.X > 0 && bSize.Y > 0

	isSolid := func(pos gvec.Vec2) bool {
		if s.isSolid(runtime, pos) {
			return true
		}
		if hasBase {
			return pos.X < bPos.X+bSize.X && pos.X+s.Dimensions.X > bPos.X &&
				pos.Y < bPos.Y+bSize.Y && pos.Y+s.Dimensions.Y > bPos.Y
		}
		return false
	}

	gvec.MoveAxisSeparated(&s.Pos, &s.Vel, s.Dimensions, isSolid, nil, nil)
}

func (s *Skiff) isSolid(runtime Runtime, pos gvec.Vec2) bool {
	return solidAt(runtime.IsOverworldSolidAt, pos, s.Dimensions)
}

func (s *Skiff) Draw(screen *ebiten.Image, camX, camY float64) {
	if activeWakeStyle == WakeStyleVLines {
		// Draw continuous V-wake lines (Option B)
		if len(s.wake) >= 2 {
			const spreadAngle = 0.42
			for i := 0; i < len(s.wake)-1; i++ {
				p1 := s.wake[i]
				p2 := s.wake[i+1]

				// Calculate left/right coordinates for p1
				age1 := 1.0 - p1.life
				dist1 := age1 * 32.0

				lAngle1 := p1.facing + math.Pi - spreadAngle
				lX1 := p1.x + math.Cos(lAngle1)*dist1 - camX
				lY1 := p1.y + math.Sin(lAngle1)*dist1 - camY

				rAngle1 := p1.facing + math.Pi + spreadAngle
				rX1 := p1.x + math.Cos(rAngle1)*dist1 - camX
				rY1 := p1.y + math.Sin(rAngle1)*dist1 - camY

				// Calculate left/right coordinates for p2
				age2 := 1.0 - p2.life
				dist2 := age2 * 32.0

				lAngle2 := p2.facing + math.Pi - spreadAngle
				lX2 := p2.x + math.Cos(lAngle2)*dist2 - camX
				lY2 := p2.y + math.Sin(lAngle2)*dist2 - camY

				rAngle2 := p2.facing + math.Pi + spreadAngle
				rX2 := p2.x + math.Cos(rAngle2)*dist2 - camX
				rY2 := p2.y + math.Sin(rAngle2)*dist2 - camY

				avgLife := (p1.life + p2.life) * 0.5

				// Outer light blue-cyan halo
				clrOuter := s.applyLight(color.RGBA{160, 220, 255, uint8(avgLife * 80.0)})
				vector.StrokeLine(screen, float32(lX1), float32(lY1), float32(lX2), float32(lY2), 2.5, clrOuter, true)
				vector.StrokeLine(screen, float32(rX1), float32(rY1), float32(rX2), float32(rY2), 2.5, clrOuter, true)

				// Inner white core
				clrInner := s.applyLight(color.RGBA{255, 255, 255, uint8(avgLife * 160.0)})
				vector.StrokeLine(screen, float32(lX1), float32(lY1), float32(lX2), float32(lY2), 1.0, clrInner, true)
				vector.StrokeLine(screen, float32(rX1), float32(rY1), float32(rX2), float32(rY2), 1.0, clrInner, true)
			}
		}
	} else if activeWakeStyle == WakeStyleArcs {
		// Draw expanding backward-facing wave arcs (Option A)
		for _, p := range s.wake {
			alpha := p.life * p.amplitude
			if alpha > 255 {
				alpha = 255
			}
			if alpha < 0 {
				alpha = 0
			}

			// Radius expands based on age
			radius := 2.0 + (1.0-p.life)*80.0

			// Center angle is reverse of historical facing
			centerAngle := p.facing + math.Pi
			const halfSweep = 1.0 // ~114 degrees sweep arc

			// Outer light-blue-cyan halo
			clrOuter := s.applyLight(color.RGBA{160, 220, 255, uint8(alpha * 0.5)})
			drawArc(screen, p.x-camX, p.y-camY, radius, centerAngle, halfSweep, 2.5, clrOuter)

			// Inner white core
			clrInner := s.applyLight(color.RGBA{255, 255, 255, uint8(alpha)})
			drawArc(screen, p.x-camX, p.y-camY, radius, centerAngle, halfSweep, 1.0, clrInner)
		}
	} else if activeWakeStyle == WakeStyleVSegments {
		// Draw expanding backward-facing V-line segments (Option C: Combination)
		const spreadAngle = 0.65
		for _, p := range s.wake {
			// Opacity decays quadratically for a distinct, smooth fade-out
			maxAlpha := 100.0 + p.amplitude
			if maxAlpha > 230.0 {
				maxAlpha = 230.0
			}
			alpha := p.life * p.life * maxAlpha

			age := 1.0 - p.life
			dist := age * 80.0 // expands further outwards (up to 80 pixels)

			distStart := dist * 0.3
			distEnd := dist

			// Thickness also decays over time to simulate wave dissipation
			thickOuter := float32(2.5 * p.life)
			thickInner := float32(1.0 * p.life)
			if thickOuter < 0.1 {
				thickOuter = 0.1
			}
			if thickInner < 0.1 {
				thickInner = 0.1
			}

			// Left segment
			leftAngle := p.facing + math.Pi - spreadAngle
			lXStart := p.x + math.Cos(leftAngle)*distStart - camX
			lYStart := p.y + math.Sin(leftAngle)*distStart - camY
			lXEnd := p.x + math.Cos(leftAngle)*distEnd - camX
			lYEnd := p.y + math.Sin(leftAngle)*distEnd - camY

			// Right segment
			rightAngle := p.facing + math.Pi + spreadAngle
			rXStart := p.x + math.Cos(rightAngle)*distStart - camX
			rYStart := p.y + math.Sin(rightAngle)*distStart - camY
			rXEnd := p.x + math.Cos(rightAngle)*distEnd - camX
			rYEnd := p.y + math.Sin(rightAngle)*distEnd - camY

			// Outer light-blue-cyan halo
			clrOuter := s.applyLight(color.RGBA{160, 220, 255, uint8(alpha * 0.5)})
			vector.StrokeLine(screen, float32(lXStart), float32(lYStart), float32(lXEnd), float32(lYEnd), thickOuter, clrOuter, true)
			vector.StrokeLine(screen, float32(rXStart), float32(rYStart), float32(rXEnd), float32(rYEnd), thickOuter, clrOuter, true)

			// Inner white core
			clrInner := s.applyLight(color.RGBA{255, 255, 255, uint8(alpha)})
			vector.StrokeLine(screen, float32(lXStart), float32(lYStart), float32(lXEnd), float32(lYEnd), thickInner, clrInner, true)
			vector.StrokeLine(screen, float32(rXStart), float32(rYStart), float32(rXEnd), float32(rYEnd), thickInner, clrInner, true)
		}
	}

	if skiffSheet != nil {
		rect := image.Rect(348, 82, 676, 948)
		sprite := skiffSheet.SubImage(rect).(*ebiten.Image)

		op := &ebiten.DrawImageOptions{}

		// Center the cropped sprite on the origin (0, 0)
		op.GeoM.Translate(-164.0, -433.0)

		// Scale the cropped sprite to fit the target dimensions
		scaleX := s.Dimensions.Y / 328.0
		scaleY := s.Dimensions.X / 866.0
		op.GeoM.Scale(scaleX, scaleY)

		// Rotate so it aligns with s.Facing
		op.GeoM.Rotate(s.Facing + math.Pi/2.0)

		// Translate to screen coordinates, centered on the boat's collision box center
		cx := s.Pos.X + s.Dimensions.X/2.0 - camX
		cy := s.Pos.Y + s.Dimensions.Y/2.0 - camY
		op.GeoM.Translate(cx, cy)

		// Apply lighting scale
		mult := float32(s.lightMult)
		op.ColorScale.Scale(mult, mult, mult, 1.0)

		screen.DrawImage(sprite, op)
		return
	}

	// Fallback to original vector drawing code
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

func (s *Skiff) getLightMultiplier(timeOfDay float64) float64 {
	if timeOfDay >= 0 && timeOfDay < 1200 {
		return 0.2 + (timeOfDay/1200.0)*0.8
	}
	if timeOfDay >= 1200 && timeOfDay < 9600 {
		return 1.0
	}
	if timeOfDay >= 9600 && timeOfDay < 10800 {
		return 1.0 - ((timeOfDay-9600.0)/1200.0)*0.8
	}
	return 0.2
}

func (s *Skiff) applyLight(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * s.lightMult),
		G: uint8(float64(c.G) * s.lightMult),
		B: uint8(float64(c.B) * s.lightMult),
		A: c.A,
	}
}

func drawArc(screen *ebiten.Image, cx, cy, radius float64, centerAngle, halfSweep float64, thickness float32, clr color.Color) {
	const segments = 8
	step := (halfSweep * 2.0) / float64(segments)
	startAngle := centerAngle - halfSweep
	var lastX, lastY float32
	for i := 0; i <= segments; i++ {
		angle := startAngle + float64(i)*step
		px := float32(cx + math.Cos(angle)*radius)
		py := float32(cy + math.Sin(angle)*radius)
		if i > 0 {
			vector.StrokeLine(screen, lastX, lastY, px, py, thickness, clr, true)
		}
		lastX, lastY = px, py
	}
}
