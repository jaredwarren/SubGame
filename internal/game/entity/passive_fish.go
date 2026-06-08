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
