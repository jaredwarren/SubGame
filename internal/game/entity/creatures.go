package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// PassiveFish is a catchable swimming creature that flees from the player.
type PassiveFish struct {
	BaseEntity
	FacingRight bool
	SwimPhase   float64
	FleeTimer   int
}

func NewPassiveFish(x, y float64, facingRight bool, swimPhase float64) *PassiveFish {
	return &PassiveFish{
		BaseEntity: BaseEntity{
			Type:       EntPassiveFish,
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 20, Y: 12},
			Active:     true,
		},
		FacingRight: facingRight,
		SwimPhase:   swimPhase,
	}
}

func (f *PassiveFish) GetHarvestedItem() item.Item { return &item.RawFish{} }

func (f *PassiveFish) CanCatch(playerPos gvec.Vec2) bool {
	cx := f.Pos.X + f.Dimensions.X/2
	cy := f.Pos.Y + f.Dimensions.Y/2
	return math.Hypot(playerPos.X-cx, playerPos.Y-cy) <= 80.0
}

func (f *PassiveFish) Update(gr Runtime) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2
	fx := f.Pos.X + f.Dimensions.X/2
	fy := f.Pos.Y + f.Dimensions.Y/2
	dist := math.Hypot(px-fx, py-fy)

	f.SwimPhase += 0.04

	if f.FleeTimer > 0 {
		f.FleeTimer--
		speed := 3.5
		if f.FacingRight {
			f.Vel.X = speed
		} else {
			f.Vel.X = -speed
		}
		f.Vel.Y = math.Sin(f.SwimPhase*2) * 1.0
	} else if dist < 120 {
		f.FleeTimer = 60
		f.FacingRight = px < fx
	} else {
		speed := 0.6
		if f.FacingRight {
			f.Vel.X = speed
		} else {
			f.Vel.X = -speed
		}
		f.Vel.Y = math.Sin(f.SwimPhase) * 0.4
	}

	nextX := f.Pos.X + f.Vel.X
	nextY := f.Pos.Y + f.Vel.Y
	if !gr.IsSolid(nextX, nextY, f.Dimensions.X, f.Dimensions.Y) {
		f.Pos.X = nextX
		f.Pos.Y = nextY
	} else {
		f.FacingRight = !f.FacingRight
		f.FleeTimer = 0
	}
}

func (f *PassiveFish) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(f.Pos.X - camera.Pos.X)
	sy := float32(f.Pos.Y - camera.Pos.Y)
	sw := float32(f.Dimensions.X)
	sh := float32(f.Dimensions.Y)
	cx := sx + sw/2
	cy := sy + sh/2

	vector.FillCircle(screen, cx, cy, 6.0, color.RGBA{60, 160, 200, 255}, false)

	var tailX float32
	if f.FacingRight {
		tailX = cx - 8
	} else {
		tailX = cx + 8
	}
	wiggle := float32(math.Sin(timeOfDay*0.12+float64(f.SwimPhase))) * 3
	entityPath.Reset()
	entityPath.MoveTo(tailX, cy)
	entityPath.LineTo(tailX-4+wiggle, cy-5)
	entityPath.LineTo(tailX-4+wiggle, cy+5)
	entityPath.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(color.RGBA{40, 130, 180, 200})
	vector.FillPath(screen, entityPath, nil, &opts)

	var eyeX float32
	if f.FacingRight {
		eyeX = cx + 3
	} else {
		eyeX = cx - 3
	}
	vector.FillCircle(screen, eyeX, cy-1.5, 1.5, color.White, false)
	vector.FillCircle(screen, eyeX, cy-1.5, 0.8, color.Black, false)
	vector.StrokeLine(screen, cx-4, cy-3, cx+4, cy-3, 0.8, color.RGBA{80, 200, 240, 180}, false)
}

// PassiveCrab is a catchable floor creature that retreats into its shell when threatened.
type PassiveCrab struct {
	BaseEntity
	FacingRight bool
	InShell     bool
	ShellTimer  int
	WalkTimer   int
}

func (c *PassiveCrab) GetHarvestedItem() item.Item { return &item.RawCrab{} }

func (c *PassiveCrab) CanCatch(playerPos gvec.Vec2) bool {
	cx := c.Pos.X + c.Dimensions.X/2
	cy := c.Pos.Y + c.Dimensions.Y/2
	return math.Hypot(playerPos.X-cx, playerPos.Y-cy) <= 64.0
}

func (c *PassiveCrab) Update(gr Runtime) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2
	cx := c.Pos.X + c.Dimensions.X/2
	cy := c.Pos.Y + c.Dimensions.Y/2
	dist := math.Hypot(px-cx, py-cy)

	isLit := false
	if gr.FlashlightOn() && dist < 300 {
		facingAngle := gr.PlayerFacing()
		dx := cx - px
		dy := cy - py
		angleToEnt := math.Atan2(dy, dx)
		diff := angleToEnt - facingAngle
		for diff > math.Pi {
			diff -= 2 * math.Pi
		}
		for diff < -math.Pi {
			diff += 2 * math.Pi
		}
		if math.Abs(diff) < 0.42 {
			isLit = true
		}
	}

	if dist < 100 || isLit {
		c.InShell = true
		c.ShellTimer = 90
		c.Vel.X = 0
	} else if c.ShellTimer > 0 {
		c.ShellTimer--
		if c.ShellTimer <= 0 {
			c.InShell = false
		}
	}

	if !c.InShell {
		c.WalkTimer++
		if c.WalkTimer%180 == 0 {
			c.FacingRight = !c.FacingRight
		}
		speed := 0.35
		if c.FacingRight {
			c.Vel.X = speed
		} else {
			c.Vel.X = -speed
		}
	}

	c.Vel.Y += 0.3
	if c.Vel.Y > 4.0 {
		c.Vel.Y = 4.0
	}

	nextX := c.Pos.X + c.Vel.X
	if !gr.IsSolid(nextX, c.Pos.Y, c.Dimensions.X, c.Dimensions.Y) {
		c.Pos.X = nextX
	} else {
		c.FacingRight = !c.FacingRight
		c.Vel.X = 0
	}

	nextY := c.Pos.Y + c.Vel.Y
	if !gr.IsSolid(c.Pos.X, nextY, c.Dimensions.X, c.Dimensions.Y) {
		c.Pos.Y = nextY
	} else {
		c.Vel.Y = 0
	}
}

func (c *PassiveCrab) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	sw := float32(c.Dimensions.X)
	sh := float32(c.Dimensions.Y)
	ccx := sx + sw/2
	ccy := sy + sh/2

	shellColor := color.RGBA{180, 60, 50, 255}
	legColor := color.RGBA{160, 45, 40, 255}

	if c.InShell {
		vector.FillCircle(screen, ccx, ccy, 6.0, shellColor, false)
		vector.StrokeCircle(screen, ccx, ccy, 6.0, 1.0, color.RGBA{140, 40, 35, 255}, false)
		vector.StrokeLine(screen, ccx-3, ccy-2, ccx+3, ccy-2, 0.8, color.RGBA{130, 35, 30, 255}, false)
		vector.StrokeLine(screen, ccx-2, ccy+1, ccx+2, ccy+1, 0.8, color.RGBA{130, 35, 30, 255}, false)
		return
	}

	vector.FillCircle(screen, ccx, ccy, 5.0, shellColor, false)

	legWiggle := float32(math.Sin(timeOfDay*0.15+float64(c.WalkTimer)*0.1)) * 2
	vector.StrokeLine(screen, ccx-4, ccy+2, ccx-7, ccy+5+legWiggle, 1.2, legColor, false)
	vector.StrokeLine(screen, ccx-3, ccy+3, ccx-6, ccy+6-legWiggle, 1.2, legColor, false)
	vector.StrokeLine(screen, ccx+4, ccy+2, ccx+7, ccy+5-legWiggle, 1.2, legColor, false)
	vector.StrokeLine(screen, ccx+3, ccy+3, ccx+6, ccy+6+legWiggle, 1.2, legColor, false)

	clawColor := color.RGBA{200, 70, 55, 255}
	if c.FacingRight {
		vector.FillCircle(screen, ccx+7, ccy-2, 2.5, clawColor, false)
		vector.FillCircle(screen, ccx-6, ccy-1, 2.0, clawColor, false)
	} else {
		vector.FillCircle(screen, ccx-7, ccy-2, 2.5, clawColor, false)
		vector.FillCircle(screen, ccx+6, ccy-1, 2.0, clawColor, false)
	}

	vector.StrokeLine(screen, ccx-2, ccy-4, ccx-2, ccy-7, 0.8, legColor, false)
	vector.FillCircle(screen, ccx-2, ccy-7, 1.0, color.White, false)
	vector.StrokeLine(screen, ccx+2, ccy-4, ccx+2, ccy-7, 0.8, legColor, false)
	vector.FillCircle(screen, ccx+2, ccy-7, 1.0, color.White, false)
}

// Kelp is a decorative swaying sea plant.
type Kelp struct {
	BaseEntity
	SwayPhase float64
}

func (k *Kelp) Update(gr Runtime) {
	k.SwayPhase += 0.03
}

func (k *Kelp) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(k.Pos.X - camera.Pos.X)
	sy := float32(k.Pos.Y - camera.Pos.Y)
	sw := float32(k.Dimensions.X)
	sh := float32(k.Dimensions.Y)
	cx := sx + sw/2.0
	bottomY := sy + sh

	numSegments := int(sh / 8.0)
	if numSegments < 3 {
		numSegments = 3
	}
	segmentHeight := sh / float32(numSegments)

	lastX := cx
	lastY := bottomY

	for i := 0; i < numSegments; i++ {
		factor := float64(i+1) / float64(numSegments)
		swayOffset := float32(math.Sin(k.SwayPhase+float64(i)*0.4)) * 8.0 * float32(factor)
		nextX := cx + swayOffset
		nextY := bottomY - float32(i+1)*segmentHeight

		vector.StrokeLine(screen, lastX, lastY, nextX, nextY, 2.5-float32(factor)*1.0, color.RGBA{34, 139, 34, 255}, false)

		leafSize := 5.0 - float32(factor)*2.0
		if leafSize < 2.0 {
			leafSize = 2.0
		}
		vector.FillCircle(screen, nextX-4.0, nextY, leafSize, color.RGBA{46, 150, 60, 220}, false)
		vector.FillCircle(screen, nextX+4.0, nextY, leafSize, color.RGBA{46, 150, 60, 220}, false)

		lastX = nextX
		lastY = nextY
	}
}
