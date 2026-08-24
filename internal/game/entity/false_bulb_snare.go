package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// FalseBulbSnare mimics a ShatterBulb but lunges at and damages the player.
type FalseBulbSnare struct {
	BaseEntity
	def   *FalseBulbSnareDef
	State int
}

func (ent *FalseBulbSnare) stats() *FalseBulbSnareDef {
	if ent.def != nil {
		return ent.def
	}
	return FalseBulbSnareArchetype
}

// NewFalseBulbSnare creates a FalseBulbSnare at the given position.
func NewFalseBulbSnare(x, y float64) *FalseBulbSnare {
	d := FalseBulbSnareArchetype
	return &FalseBulbSnare{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: d.Dims,
			Active:     true,
		},
		def:   d,
		State: 0,
	}
}

// SnareContext defines the context interface needed by FalseBulbSnare.
type SnareContext interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
	FlashlightOn() bool
	PlayerFacing() float64
	HasActiveVehicle() bool
	ActiveVehicleFacing() float64
	ActiveVehiclePos() gvec.Vec2
	ActiveVehicleDims() gvec.Vec2
	SoundWaveTimer() int
	SoundWaveX() float64
	SoundWaveY() float64
	Emit(cmd GameCommand)
	FindClosestDecoy(pos gvec.Vec2, maxDist float64) (gvec.Vec2, bool)
	CheckDeterrentOcclusion(pos1, pos2 gvec.Vec2) bool
	CheckDeterrentSlowing(x, y, w, h float64) bool
}

func (ent *FalseBulbSnare) Update(gr Runtime) {
	ent.update(gr)
}

func (ent *FalseBulbSnare) update(g SnareContext) {
	d := ent.stats()
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0

	tgt := AcquireTarget(g, gvec.Vec2{X: ex, Y: ey}, d.DecoyRange, d.DecoyTargetSize)
	targetX, targetY := tgt.CenterX, tgt.CenterY
	targetW, targetH := tgt.Width, tgt.Height
	isDecoy := tgt.IsDecoy
	dist := math.Hypot(targetX-ex, targetY-ey)

	if dist > d.LeashRange {
		ent.State = 0
		return
	}

	if !isDecoy {
		if g.CheckDeterrentSlowing(tgt.TopLeftX, tgt.TopLeftY, targetW, targetH) || g.CheckDeterrentOcclusion(gvec.Vec2{X: ex, Y: ey}, gvec.Vec2{X: targetX, Y: targetY}) {
			ent.State = 0
			ent.Vel = gvec.Vec2{}
			return
		}
	}

	isLit := !isDecoy && InFlashlightCone(g, gvec.Vec2{X: ex, Y: ey}, gvec.Vec2{X: targetX, Y: targetY}, d.FlashlightConeHalfAngle)

	soundAlerted := !isDecoy && g.SoundWaveTimer() > 0 && math.Hypot(g.SoundWaveX()-ex, g.SoundWaveY()-ey) < d.SoundAlertRange
	if soundAlerted || isDecoy {
		ent.State = 1
	}

	if isLit {
		ent.Vel = gvec.Vec2{}
	} else if dist < d.ChaseRange || ent.State == 1 {
		ent.State = 1
		dx := targetX - ex
		dy := targetY - ey
		dDist := math.Hypot(dx, dy)
		if dDist > 0 {
			ent.Vel.X = (dx / dDist) * d.ChaseSpeed
			ent.Vel.Y = (dy / dDist) * d.ChaseSpeed
		}
		ent.Pos = ent.Pos.Add(ent.Vel)
	} else {
		ent.State = 0
	}

	if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, tgt.TopLeftX, tgt.TopLeftY, targetW, targetH) {
		if isDecoy {
			g.Emit(DestroyDecoyCmd{Pos: gvec.Vec2{X: targetX, Y: targetY}})
			ent.Active = false
		} else {
			if g.HasActiveVehicle() {
				g.Emit(DamageActiveVehicleCmd{Amount: d.Damage})
				g.Emit(SetMineWarningCmd{Message: "VEHICLE ATTACKED BY FALSE-BULB SNARE!", Duration: d.WarningDuration, Level: 2})
			} else {
				g.Emit(DamagePlayerCmd{Amount: d.Damage})
				g.Emit(SetMineWarningCmd{Message: "ATTACKED BY FALSE-BULB SNARE!", Duration: d.WarningDuration, Level: 2})
			}
			ent.Active = false
		}
	}
}

func (ent *FalseBulbSnare) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	vector.StrokeLine(screen, cx, sy, cx, cy, 2.0, color.RGBA{45, 95, 75, 255}, false)

	if ent.State == 1 {
		vector.FillCircle(screen, cx, cy, 12, color.RGBA{230, 75, 45, 80}, false)
		vector.FillCircle(screen, cx, cy, 7, color.RGBA{245, 95, 25, 255}, false)
		vector.StrokeLine(screen, cx, cy-4, cx, cy+4, 1.5, color.RGBA{0, 0, 0, 255}, false)
	} else {
		phase := ent.Pos.X + ent.Pos.Y
		pulse := float32(math.Cos(timeOfDay*0.02+phase)) * 2.5
		radius := float32(11.0) + pulse
		if radius < 5.0 {
			radius = 5.0
		}
		vector.FillCircle(screen, cx, cy, radius, color.RGBA{0, 220, 240, 60}, false)
		vector.FillCircle(screen, cx, cy, 7, color.RGBA{0, 220, 240, 255}, false)
		vector.StrokeCircle(screen, cx, cy, 7, 0.8, color.RGBA{255, 255, 255, 180}, false)
	}
}
