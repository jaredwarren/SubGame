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
	def    *ElectroWeaverDef
	Timer  int
	Facing float64
}

func (ent *ElectroWeaver) stats() *ElectroWeaverDef {
	if ent.def != nil {
		return ent.def
	}
	return ElectroWeaverArchetype
}

// NewElectroWeaver creates an ElectroWeaver at the given position.
func NewElectroWeaver(x, y float64) *ElectroWeaver {
	d := ElectroWeaverArchetype
	return &ElectroWeaver{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: d.Dims,
			Active:     true,
		},
		def: d,
	}
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
	FindClosestDecoy(pos gvec.Vec2, maxDist float64) (gvec.Vec2, bool)
	CheckDeterrentOcclusion(pos1, pos2 gvec.Vec2) bool
}

func (ent *ElectroWeaver) Update(gr Runtime) {
	ent.update(gr)
}

func (ent *ElectroWeaver) update(g WeaverContext) {
	d := ent.stats()
	px := g.PlayerPos().X + g.PlayerDims().X/2.0
	py := g.PlayerPos().Y + g.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0

	var targetX, targetY float64
	var isDecoy bool

	decoyPos, decoyFound := g.FindClosestDecoy(gvec.Vec2{X: ex, Y: ey}, d.DecoyRange)
	if decoyFound {
		targetX = decoyPos.X
		targetY = decoyPos.Y
		isDecoy = true
	} else {
		targetX = px
		targetY = py
	}
	dist := math.Hypot(targetX-ex, targetY-ey)

	inAbyssal := (py/config.TileSize) >= d.AbyssalDepthTiles || g.IsShockKelpCave()
	if !inAbyssal {
		ent.Timer = 0
		return
	}

	isElectricity := g.FlashlightOn() || g.SonarActive() || g.HasActiveVehicle() || isDecoy

	// Deterrent cloud occlusion + lights off breaking
	if !isDecoy && !g.FlashlightOn() && g.CheckDeterrentOcclusion(gvec.Vec2{X: ex, Y: ey}, gvec.Vec2{X: px, Y: py}) {
		isElectricity = false
		ent.Timer = 0
	}

	if isElectricity && dist < d.TrackRange {
		ent.Timer++
		g.Emit(UpdateWeaverTrackingTimerCmd{Value: float64(ent.Timer)})
		if ent.Timer >= d.StrikeTimerFrames {
			if isDecoy {
				g.Emit(DestroyDecoyCmd{Pos: gvec.Vec2{X: targetX, Y: targetY}})
				g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKES DECOY!", Duration: d.DecoyWarningDuration, Level: 1})

				// Teleport back to cooldown (random angle, away from player)
				angle := rand.Float64() * 2.0 * math.Pi
				ent.Pos.X = px + math.Cos(angle)*d.TeleportAwayDist
				ent.Pos.Y = py + math.Sin(angle)*d.TeleportAwayDist
				ent.Timer = 0
			} else {
				g.Emit(DamagePlayerCmd{Amount: d.PlayerDamage, Kind: DamageElectric})
				g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKE! SEVERE DAMAGE!", Duration: d.WarningDuration, Level: 3})
				ent.Pos.X = g.PlayerPos().X + float64(rand.Intn(120)-60)
				ent.Pos.Y = g.PlayerPos().Y + float64(rand.Intn(120)-60)
				ent.Timer = 0
			}
		}
	} else {
		if ent.Timer > 0 {
			ent.Timer -= d.TimerDecay
			if ent.Timer < 0 {
				ent.Timer = 0
			}
		}
	}

	if ent.Timer > d.MoveStartTimer {
		dx := targetX - ex
		dy := targetY - ey
		dDist := math.Hypot(dx, dy)
		if dDist > d.ApproachDist {
			ent.Vel.X = (dx / dDist) * d.ApproachSpeed
			ent.Vel.Y = (dy / dDist) * d.ApproachSpeed
		} else {
			ent.Vel.X = math.Cos(g.TimeOfDay()/30.0) * d.OrbitSpeedClose
			ent.Vel.Y = math.Sin(g.TimeOfDay()/30.0) * d.OrbitSpeedClose
		}
	} else {
		ent.Vel.X = math.Cos(g.TimeOfDay()/40.0) * d.IdleSpeed
		ent.Vel.Y = math.Sin(g.TimeOfDay()/40.0) * d.IdleSpeed
	}

	if !g.IsSolid(ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
		ent.Pos = ent.Pos.Add(ent.Vel)
	}
}

func (ent *ElectroWeaver) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	d := ent.stats()
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
		sparkRatio := float64(ent.Timer) / float64(d.StrikeTimerFrames)
		for s := 0; s < int(sparkRatio*5); s++ {
			spx := cx + float32(rand.Intn(40)-20)
			spy := cy + float32(rand.Intn(40)-20)
			vector.StrokeLine(screen, cx, cy, spx, spy, 1.0, color.RGBA{160, 220, 255, 255}, false)
		}
	}
}
