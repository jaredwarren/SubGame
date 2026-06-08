package overworld

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/light"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// CosmeticFish represents an individual fish swimming near shorelines.
type CosmeticFish struct {
	Pos       gvec.Vec2
	Vel       gvec.Vec2
	Angle     float64
	BasePos   gvec.Vec2
	WobbleVal float64
	WobbleSpd float64
}

// CosmeticFishContext defines the context interface needed by CosmeticFish.
type CosmeticFishContext interface {
	TargetCenter() gvec.Vec2
	IsSolid(x, y float64) bool
}

// Update updates the fish position, flocking behavior, and fleeing from the player.
func (f *CosmeticFish) Update(g CosmeticFishContext) {
	targetCenter := g.TargetCenter()
	dx := f.Pos.X - targetCenter.X
	dy := f.Pos.Y - targetCenter.Y
	dist := math.Hypot(dx, dy)

	f.WobbleVal += f.WobbleSpd

	if dist < 120.0 {
		// Flee rapidly from the player
		angle := math.Atan2(dy, dx)
		targetVelX := math.Cos(angle) * 3.5
		targetVelY := math.Sin(angle) * 3.5
		f.Vel.X = f.Vel.X*0.88 + targetVelX*0.12
		f.Vel.Y = f.Vel.Y*0.88 + targetVelY*0.12
	} else {
		// Wander slowly around its base position
		bdx := f.BasePos.X - f.Pos.X
		bdy := f.BasePos.Y - f.Pos.Y
		bdist := math.Hypot(bdx, bdy)

		var wanderX, wanderY float64
		if bdist > 96.0 {
			// Steer back towards home/base position
			bangle := math.Atan2(bdy, bdx)
			wanderX = math.Cos(bangle) * 0.8
			wanderY = math.Sin(bangle) * 0.8
		} else {
			// Drift/swim gently using a sine wobble
			wanderX = math.Cos(f.WobbleVal) * 0.5
			wanderY = math.Sin(f.WobbleVal*0.5) * 0.5
		}

		f.Vel.X = f.Vel.X*0.94 + wanderX*0.06
		f.Vel.Y = f.Vel.Y*0.94 + wanderY*0.06
	}

	// Move and handle solid/land check separately for X and Y
	newX := f.Pos.X + f.Vel.X
	if g.IsSolid(newX, f.Pos.Y) {
		f.Vel.X = -f.Vel.X * 0.5
	} else {
		f.Pos.X = newX
	}

	newY := f.Pos.Y + f.Vel.Y
	if g.IsSolid(f.Pos.X, newY) {
		f.Vel.Y = -f.Vel.Y * 0.5
	} else {
		f.Pos.Y = newY
	}

	if f.Vel.Length() > 0.05 {
		f.Angle = math.Atan2(f.Vel.Y, f.Vel.X)
	}
}

// Draw renders a tiny procedurally animated fish.
func (f *CosmeticFish) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, mult float64) {
	sx := float32(f.Pos.X - camX)
	sy := float32(f.Pos.Y - camY)

	// Length and width of fish body
	bodyLen := 7.0
	bodyW := 3.5

	cos := math.Cos(f.Angle)
	sin := math.Sin(f.Angle)

	// Animated tail wiggle based on speed and time
	wiggleSpd := 0.22
	if f.Vel.Length() > 1.5 {
		wiggleSpd = 0.45
	}
	wiggle := math.Sin(ticks*wiggleSpd+f.WobbleVal) * 0.5
	tailCos := math.Cos(f.Angle + math.Pi + wiggle)
	tailSin := math.Sin(f.Angle + math.Pi + wiggle)

	// Coordinates for the triangular/diamond fish body
	tipX := sx + float32(bodyLen*0.5*cos)
	tipY := sy + float32(bodyLen*0.5*sin)

	blX := sx + float32(-bodyLen*0.5*cos+bodyW*0.5*-sin)
	blY := sy + float32(-bodyLen*0.5*sin+bodyW*0.5*cos)

	brX := sx + float32(-bodyLen*0.5*cos-bodyW*0.5*-sin)
	brY := sy + float32(-bodyLen*0.5*sin-bodyW*0.5*cos)

	fishColor := color.RGBA{110, 190, 220, 180}
	fishColor = light.ApplyLight(fishColor, mult)

	var path vector.Path
	path.MoveTo(tipX, tipY)
	path.LineTo(blX, blY)
	path.LineTo(brX, brY)
	path.Close()

	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(fishColor)
	vector.FillPath(screen, &path, nil, &opts)

	// Triangular wiggling tail fin
	tailBaseX := sx + float32(-bodyLen*0.5*cos)
	tailBaseY := sy + float32(-bodyLen*0.5*sin)
	tailTipL_X := tailBaseX + float32(4.5*tailCos+2.2*-tailSin)
	tailTipL_Y := tailBaseY + float32(4.5*tailSin+2.2*tailCos)
	tailTipR_X := tailBaseX + float32(4.5*tailCos-2.2*-tailSin)
	tailTipR_Y := tailBaseY + float32(4.5*tailSin-2.2*tailCos)

	var tailPath vector.Path
	tailPath.MoveTo(tailBaseX, tailBaseY)
	tailPath.LineTo(tailTipL_X, tailTipL_Y)
	tailPath.LineTo(tailTipR_X, tailTipR_Y)
	tailPath.Close()
	vector.FillPath(screen, &tailPath, nil, &opts)
}
