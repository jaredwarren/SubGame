package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// ThermoclineRammer is a fast-charging aquatic predator that rams the player.
type ThermoclineRammer struct {
	BaseEntity
	State     int
	Timer     int
	Facing    float64
	StunTimer int
}

// RammerContext defines the context interface needed by ThermoclineRammer.
type RammerContext interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
	PlayerVel() gvec.Vec2
	IsPlayerSprinting() bool
	HasActiveVehicle() bool
	ActiveVehicleMoving() bool
	ActiveVehiclePos() gvec.Vec2
	ActiveVehicleDims() gvec.Vec2
	SoundWaveTimer() int
	SoundWaveX() float64
	SoundWaveY() float64
	IsSolid(x, y, w, h float64) bool
	Emit(cmd GameCommand)
}

func (ent *ThermoclineRammer) Update(gr Runtime) {
	ent.update(gr)
}

func (ent *ThermoclineRammer) update(g RammerContext) {
	px := g.PlayerPos().X + g.PlayerDims().X/2.0
	py := g.PlayerPos().Y + g.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0
	dist := math.Hypot(px-ex, py-ey)

	if ent.State == 2 {
		ent.StunTimer--
		if ent.StunTimer <= 0 {
			ent.State = 0
		}
		return
	}

	isAggroTrigger := false
	if dist < 250.0 {
		if !g.HasActiveVehicle() && g.IsPlayerSprinting() && (math.Abs(g.PlayerVel().X) > 1.2 || math.Abs(g.PlayerVel().Y) > 1.2) {
			isAggroTrigger = true
		}
		if g.HasActiveVehicle() && g.ActiveVehicleMoving() {
			isAggroTrigger = true
		}
	}
	if g.SoundWaveTimer() > 0 && math.Hypot(g.SoundWaveX()-ex, g.SoundWaveY()-ey) < 250.0 {
		isAggroTrigger = true
	}

	switch ent.State {
	case 0: // patrol
		if isAggroTrigger {
			ent.State = 1
			dx := px - ex
			dy := py - ey
			if math.Abs(dx) > math.Abs(dy) {
				ent.Vel.Y = 0
				if dx > 0 {
					ent.Vel.X, ent.Facing = 6.2, 0.0
				} else {
					ent.Vel.X, ent.Facing = -6.2, math.Pi
				}
			} else {
				ent.Vel.X = 0
				if dy > 0 {
					ent.Vel.Y, ent.Facing = 6.2, math.Pi/2.0
				} else {
					ent.Vel.Y, ent.Facing = -6.2, -math.Pi/2.0
				}
			}
		} else {
			ent.Timer++
			if ent.Timer%120 == 0 {
				ent.Facing += math.Pi
			}
			ent.Vel.X = math.Cos(ent.Facing) * 0.8
			ent.Vel.Y = math.Sin(ent.Facing) * 0.4
			if !g.IsSolid(ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
				ent.Pos = ent.Pos.Add(ent.Vel)
			} else {
				ent.Facing += math.Pi
			}
		}
	case 1: // charging
		nextX := ent.Pos.X + ent.Vel.X
		nextY := ent.Pos.Y + ent.Vel.Y
		if g.IsSolid(nextX, nextY, ent.Dimensions.X, ent.Dimensions.Y) {
			ent.State = 2
			ent.StunTimer = 180
			ent.Vel = gvec.Vec2{}
		} else {
			ent.Pos.X = nextX
			ent.Pos.Y = nextY
		}

		vWidth, vHeight := g.PlayerDims().X, g.PlayerDims().Y
		targetX, targetY := g.PlayerPos().X, g.PlayerPos().Y
		if g.HasActiveVehicle() {
			vPos := g.ActiveVehiclePos()
			targetX, targetY = vPos.X, vPos.Y
			vDims := g.ActiveVehicleDims()
			vWidth, vHeight = vDims.X, vDims.Y
		}
		if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
			dirX, dirY := 0.0, 0.0
			speed := math.Hypot(ent.Vel.X, ent.Vel.Y)
			if speed > 0.1 {
				dirX = ent.Vel.X / speed
				dirY = ent.Vel.Y / speed
			} else {
				dx := (targetX + vWidth/2.0) - ex
				dy := (targetY + vHeight/2.0) - ey
				dist := math.Hypot(dx, dy)
				if dist > 0.1 {
					dirX = dx / dist
					dirY = dy / dist
				} else {
					dirX = 1.0
				}
			}

			kbForce := 6.5
			forceVec := gvec.Vec2{X: dirX * kbForce, Y: dirY * kbForce}

			if g.HasActiveVehicle() {
				g.Emit(DamageActiveVehicleCmd{Amount: 30.0})
				g.Emit(KnockbackActiveVehicleCmd{Force: forceVec})
				g.Emit(SetMineWarningCmd{Message: "VEHICLE RAMMED BY THERMOCLINE RAMMER!", Duration: 120, Level: 2})
			} else {
				g.Emit(DamagePlayerCmd{Amount: 25.0})
				g.Emit(KnockbackPlayerCmd{Force: forceVec})
				g.Emit(SetMineWarningCmd{Message: "RAMMED BY THERMOCLINE RAMMER!", Duration: 120, Level: 2})
			}

			// Push rammer back in opposite direction to prevent continuous overlap
			pushBackDistance := 40.0
			ent.Pos.X -= dirX * pushBackDistance
			ent.Pos.Y -= dirY * pushBackDistance
			ent.Vel = gvec.Vec2{}
			ent.State = 2
			ent.StunTimer = 180
		}
	}
}

func (ent *ThermoclineRammer) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	vector.FillCircle(screen, cx, cy, 8.0, color.RGBA{195, 95, 45, 255}, false)

	cosF := float32(math.Cos(ent.Facing))
	sinF := float32(math.Sin(ent.Facing))
	entityPath.Reset()
	hx := cx + cosF*12
	hy := cy + sinF*12
	entityPath.MoveTo(hx, hy)
	entityPath.LineTo(cx-sinF*6, cy+cosF*6)
	entityPath.LineTo(cx+sinF*6, cy-cosF*6)
	entityPath.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(color.RGBA{120, 130, 140, 255})
	vector.FillPath(screen, entityPath, nil, &opts)

	tx := cx - cosF*10
	ty := cy - sinF*10
	vector.StrokeLine(screen, tx, ty, tx-sinF*8, ty+cosF*8, 2.0, color.RGBA{195, 95, 45, 255}, false)
	vector.StrokeLine(screen, tx, ty, tx+sinF*8, ty-cosF*8, 2.0, color.RGBA{195, 95, 45, 255}, false)

	if ent.State == 2 {
		starAng := float64(ent.StunTimer) * 0.15
		sx1 := cx + float32(math.Cos(starAng))*14
		sy1 := cy - 14 + float32(math.Sin(starAng))*5
		sx2 := cx + float32(math.Cos(starAng+math.Pi))*14
		sy2 := cy - 14 + float32(math.Sin(starAng+math.Pi))*5
		vector.FillCircle(screen, sx1, sy1, 2.5, color.RGBA{255, 230, 40, 255}, false)
		vector.FillCircle(screen, sx2, sy2, 2.5, color.RGBA{255, 230, 40, 255}, false)
	}
}
