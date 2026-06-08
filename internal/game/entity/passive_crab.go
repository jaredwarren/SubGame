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

// CrabContext defines the context interface needed by PassiveCrab.
type CrabContext interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
	FlashlightOn() bool
	PlayerFacing() float64
	IsSolid(x, y, w, h float64) bool
}

func (c *PassiveCrab) Update(gr Runtime) {
	c.update(gr)
}

func (c *PassiveCrab) update(g CrabContext) {
	px := g.PlayerPos().X + g.PlayerDims().X/2
	py := g.PlayerPos().Y + g.PlayerDims().Y/2
	cx := c.Pos.X + c.Dimensions.X/2
	cy := c.Pos.Y + c.Dimensions.Y/2
	dist := math.Hypot(px-cx, py-cy)

	isLit := false
	if g.FlashlightOn() && dist < 300 {
		facingAngle := g.PlayerFacing()
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
	if !g.IsSolid(nextX, c.Pos.Y, c.Dimensions.X, c.Dimensions.Y) {
		c.Pos.X = nextX
	} else {
		c.FacingRight = !c.FacingRight
		c.Vel.X = 0
	}

	nextY := c.Pos.Y + c.Vel.Y
	if !g.IsSolid(c.Pos.X, nextY, c.Dimensions.X, c.Dimensions.Y) {
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
