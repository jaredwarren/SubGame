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
	State int
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
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0

	var targetX, targetY float64
	var targetW, targetH float64
	var isDecoy bool

	decoyPos, decoyFound := g.FindClosestDecoy(gvec.Vec2{X: ex, Y: ey}, 280.0)
	if decoyFound {
		targetX = decoyPos.X
		targetY = decoyPos.Y
		targetW, targetH = 16.0, 16.0
		isDecoy = true
	} else {
		if g.HasActiveVehicle() {
			vPos := g.ActiveVehiclePos()
			vDims := g.ActiveVehicleDims()
			targetX = vPos.X + vDims.X/2.0
			targetY = vPos.Y + vDims.Y/2.0
			targetW = vDims.X
			targetH = vDims.Y
		} else {
			targetX = g.PlayerPos().X + g.PlayerDims().X/2.0
			targetY = g.PlayerPos().Y + g.PlayerDims().Y/2.0
			targetW = g.PlayerDims().X
			targetH = g.PlayerDims().Y
		}
	}
	dist := math.Hypot(targetX-ex, targetY-ey)

	if dist > 360.0 {
		ent.State = 0
		return
	}

	// Occlusion check
	if !isDecoy {
		playerTopLeftX, playerTopLeftY := g.PlayerPos().X, g.PlayerPos().Y
		if g.HasActiveVehicle() {
			vPos := g.ActiveVehiclePos()
			playerTopLeftX, playerTopLeftY = vPos.X, vPos.Y
		}
		if g.CheckDeterrentSlowing(playerTopLeftX, playerTopLeftY, targetW, targetH) || g.CheckDeterrentOcclusion(gvec.Vec2{X: ex, Y: ey}, gvec.Vec2{X: targetX, Y: targetY}) {
			ent.State = 0
			ent.Vel = gvec.Vec2{}
			return
		}
	}

	isLit := false
	if !isDecoy && g.FlashlightOn() {
		facingAngle := g.PlayerFacing()
		if g.HasActiveVehicle() {
			facingAngle = g.ActiveVehicleFacing()
		}
		dx := ex - targetX
		dy := ey - targetY
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

	soundAlerted := !isDecoy && g.SoundWaveTimer() > 0 && math.Hypot(g.SoundWaveX()-ex, g.SoundWaveY()-ey) < 280.0
	if soundAlerted || isDecoy {
		ent.State = 1
	}

	if isLit {
		ent.Vel = gvec.Vec2{}
	} else {
		if dist < 180.0 || ent.State == 1 {
			ent.State = 1
			dx := targetX - ex
			dy := targetY - ey
			dDist := math.Hypot(dx, dy)
			if dDist > 0 {
				ent.Vel.X = (dx / dDist) * 3.5
				ent.Vel.Y = (dy / dDist) * 3.5
			}
			ent.Pos = ent.Pos.Add(ent.Vel)
		} else {
			ent.State = 0
		}
	}

	targetTopLeftX := targetX - targetW/2.0
	targetTopLeftY := targetY - targetH/2.0
	if !isDecoy {
		if g.HasActiveVehicle() {
			vPos := g.ActiveVehiclePos()
			targetTopLeftX, targetTopLeftY = vPos.X, vPos.Y
		} else {
			targetTopLeftX, targetTopLeftY = g.PlayerPos().X, g.PlayerPos().Y
		}
	}
	if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetTopLeftX, targetTopLeftY, targetW, targetH) {
		if isDecoy {
			g.Emit(DestroyDecoyCmd{Pos: gvec.Vec2{X: targetX, Y: targetY}})
			ent.Active = false
		} else {
			if g.HasActiveVehicle() {
				g.Emit(DamageActiveVehicleCmd{Amount: 20.0})
				g.Emit(SetMineWarningCmd{Message: "VEHICLE ATTACKED BY FALSE-BULB SNARE!", Duration: 120, Level: 2})
			} else {
				g.Emit(DamagePlayerCmd{Amount: 20.0})
				g.Emit(SetMineWarningCmd{Message: "ATTACKED BY FALSE-BULB SNARE!", Duration: 120, Level: 2})
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
