package entity

import (
	"image/color"
	"math"
	"math/rand"

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
	BodyColor   color.RGBA
	TailColor   color.RGBA
	StripeColor color.RGBA
}

var fishPresets = []struct {
	body   color.RGBA
	tail   color.RGBA
	stripe color.RGBA
}{
	// Electric Cyan/Blue (Original)
	{
		body:   color.RGBA{60, 160, 200, 255},
		tail:   color.RGBA{40, 130, 180, 200},
		stripe: color.RGBA{80, 200, 240, 180},
	},
	// Radiant Coral/Orange
	{
		body:   color.RGBA{240, 110, 80, 255},
		tail:   color.RGBA{200, 80, 50, 200},
		stripe: color.RGBA{255, 150, 120, 180},
	},
	// Bioluminescent Violet
	{
		body:   color.RGBA{160, 80, 220, 255},
		tail:   color.RGBA{120, 50, 180, 200},
		stripe: color.RGBA{200, 130, 255, 180},
	},
	// Golden Tangerine
	{
		body:   color.RGBA{240, 170, 50, 255},
		tail:   color.RGBA{190, 120, 20, 200},
		stripe: color.RGBA{255, 210, 90, 180},
	},
	// Emerald Sea-Green
	{
		body:   color.RGBA{40, 180, 110, 255},
		tail:   color.RGBA{20, 140, 80, 200},
		stripe: color.RGBA{90, 220, 160, 180},
	},
	// Sunset Magenta
	{
		body:   color.RGBA{230, 70, 130, 255},
		tail:   color.RGBA{180, 40, 90, 200},
		stripe: color.RGBA{255, 120, 180, 180},
	},
}

func NewPassiveFish(x, y float64, facingRight bool, swimPhase float64) *PassiveFish {
	preset := fishPresets[rand.Intn(len(fishPresets))]
	return &PassiveFish{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 20, Y: 12},
			Active:     true,
		},
		FacingRight: facingRight,
		SwimPhase:   swimPhase,
		BodyColor:   preset.body,
		TailColor:   preset.tail,
		StripeColor: preset.stripe,
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

	vector.FillCircle(screen, cx, cy, 6.0, f.BodyColor, false)

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
	opts.ColorScale.ScaleWithColor(f.TailColor)
	vector.FillPath(screen, entityPath, nil, &opts)

	var eyeX float32
	if f.FacingRight {
		eyeX = cx + 3
	} else {
		eyeX = cx - 3
	}
	vector.FillCircle(screen, eyeX, cy-1.5, 1.5, color.White, false)
	vector.FillCircle(screen, eyeX, cy-1.5, 0.8, color.Black, false)
	vector.StrokeLine(screen, cx-4, cy-3, cx+4, cy-3, 0.8, f.StripeColor, false)
}
