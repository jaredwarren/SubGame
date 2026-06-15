package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// ElectroWeaver is a serpentine predator that tracks electrical sources and strikes.
type ElectroWeaver struct {
	BaseEntity
	Timer  int
	Facing float64
}

// WeaverContext defines the context interface needed by ElectroWeaver.
type WeaverContext interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
	FlashlightOn() bool
	SonarActive() bool
	HasActiveVehicle() bool
	TimeOfDay() float64
	IsSolid(x, y, w, h float64) bool
	Emit(cmd GameCommand)
	IsShockKelpCave() bool
}

func (ent *ElectroWeaver) Update(gr Runtime) {
	ent.update(gr)
}

func (ent *ElectroWeaver) update(g WeaverContext) {
	px := g.PlayerPos().X + g.PlayerDims().X/2.0
	py := g.PlayerPos().Y + g.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0
	dist := math.Hypot(px-ex, py-ey)

	inAbyssal := (py/config.TileSize) >= 80 || g.IsShockKelpCave()
	if !inAbyssal {
		ent.Timer = 0
		return
	}

	isElectricity := g.FlashlightOn() || g.SonarActive() || g.HasActiveVehicle()
	if isElectricity && dist < 500.0 {
		ent.Timer++
		g.Emit(UpdateWeaverTrackingTimerCmd{Value: float64(ent.Timer)})
		if ent.Timer >= 300 {
			g.Emit(DamagePlayerCmd{Amount: 45.0})
			g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKE! SEVERE DAMAGE!", Duration: 180, Level: 3})
			ent.Pos.X = g.PlayerPos().X + float64(rand.Intn(120)-60)
			ent.Pos.Y = g.PlayerPos().Y + float64(rand.Intn(120)-60)
			ent.Timer = 0
		}
	} else {
		if ent.Timer > 0 {
			ent.Timer -= 2
			if ent.Timer < 0 {
				ent.Timer = 0
			}
		}
	}

	if ent.Timer > 60 {
		dx := px - ex
		dy := py - ey
		dDist := math.Hypot(dx, dy)
		if dDist > 100 {
			ent.Vel.X = (dx / dDist) * 1.5
			ent.Vel.Y = (dy / dDist) * 1.5
		} else {
			ent.Vel.X = math.Cos(g.TimeOfDay()/30.0) * 1.2
			ent.Vel.Y = math.Sin(g.TimeOfDay()/30.0) * 1.2
		}
	} else {
		ent.Vel.X = math.Cos(g.TimeOfDay()/40.0) * 0.8
		ent.Vel.Y = math.Sin(g.TimeOfDay()/40.0) * 0.8
	}

	if !g.IsSolid(ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
		ent.Pos = ent.Pos.Add(ent.Vel)
	}
}

func (ent *ElectroWeaver) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	for i := range 5 {
		lag := float64(i) * 0.3
		tVal := timeOfDay*0.08 - lag
		offX := math.Cos(tVal) * 6
		offY := math.Sin(tVal) * 4
		segmentX := cx - float32(math.Cos(ent.Facing)*float64(i)*8.0) + float32(offX)
		segmentY := cy - float32(math.Sin(ent.Facing)*float64(i)*8.0) + float32(offY)
		segColor := color.RGBA{140 - uint8(i*18), 45, 205 - uint8(i*12), 255}
		vector.FillCircle(screen, segmentX, segmentY, 6.0-float32(i)*0.8, segColor, false)
		if i == 0 {
			vector.FillCircle(screen, segmentX+float32(math.Cos(ent.Facing))*4, segmentY+float32(math.Sin(ent.Facing))*4, 2.0, color.RGBA{255, 255, 80, 255}, false)
		}
	}

	if ent.Timer > 0 {
		sparkRatio := float64(ent.Timer) / 300.0
		for s := 0; s < int(sparkRatio*5); s++ {
			spx := cx + float32(rand.Intn(40)-20)
			spy := cy + float32(rand.Intn(40)-20)
			vector.StrokeLine(screen, cx, cy, spx, spy, 1.0, color.RGBA{160, 220, 255, 255}, false)
		}
	}
}
