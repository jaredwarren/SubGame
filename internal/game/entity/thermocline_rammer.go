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
	def          *ThermoclineRammerDef
	State        int
	Timer        int
	Facing       float64
	StunTimer    int
	ChargeOrigin gvec.Vec2
}

func (ent *ThermoclineRammer) stats() *ThermoclineRammerDef {
	if ent.def != nil {
		return ent.def
	}
	return ThermoclineRammerArchetype
}

// NewThermoclineRammer creates a ThermoclineRammer at the given position.
func NewThermoclineRammer(x, y float64) *ThermoclineRammer {
	d := ThermoclineRammerArchetype
	return &ThermoclineRammer{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 36, Y: 24},
			Active:     true,
		},
		def: d,
	}
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
	FindClosestDecoy(pos gvec.Vec2, maxDist float64) (gvec.Vec2, bool)
	CheckDeterrentSlowing(x, y, w, h float64) bool
}

func (ent *ThermoclineRammer) Update(gr Runtime) {
	ent.update(gr)
}

func (ent *ThermoclineRammer) update(g RammerContext) {
	d := ent.stats()
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0

	var targetX, targetY float64
	var targetW, targetH float64
	var isDecoy bool

	decoyPos, decoyFound := g.FindClosestDecoy(gvec.Vec2{X: ex, Y: ey}, d.DecoyRange)
	if decoyFound {
		targetX = decoyPos.X
		targetY = decoyPos.Y
		targetW, targetH = d.DecoyTargetSize, d.DecoyTargetSize
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

	if ent.State == 2 {
		ent.StunTimer--
		if ent.StunTimer <= 0 {
			ent.State = 0
		}
		return
	}

	isAggroTrigger := false
	if decoyFound {
		isAggroTrigger = true
	} else if dist < d.AggroRange {
		if !g.HasActiveVehicle() && g.IsPlayerSprinting() && (math.Abs(g.PlayerVel().X) > d.SprintVelThreshold || math.Abs(g.PlayerVel().Y) > d.SprintVelThreshold) {
			isAggroTrigger = true
		}
		if g.HasActiveVehicle() && g.ActiveVehicleMoving() {
			isAggroTrigger = true
		}
	}
	if !decoyFound && g.SoundWaveTimer() > 0 && math.Hypot(g.SoundWaveX()-ex, g.SoundWaveY()-ey) < d.SoundAlertRange {
		isAggroTrigger = true
	}

	switch ent.State {
	case 0: // patrol
		if isAggroTrigger {
			ent.State = 1
			ent.Timer = 0
			ent.ChargeOrigin = ent.Pos
			dx := targetX - ex
			dy := targetY - ey
			if math.Abs(dx) > math.Abs(dy) {
				ent.Vel.Y = 0
				if dx > 0 {
					ent.Vel.X, ent.Facing = d.ChargeSpeed, 0.0
				} else {
					ent.Vel.X, ent.Facing = -d.ChargeSpeed, math.Pi
				}
			} else {
				ent.Vel.X = 0
				if dy > 0 {
					ent.Vel.Y, ent.Facing = d.ChargeSpeed, math.Pi/2.0
				} else {
					ent.Vel.Y, ent.Facing = -d.ChargeSpeed, -math.Pi/2.0
				}
			}
		} else {
			ent.Timer++
			if ent.Timer%d.PatrolTurnInterval == 0 {
				ent.Facing += math.Pi
			}
			ent.Vel.X = math.Cos(ent.Facing) * d.PatrolSpeedX
			ent.Vel.Y = math.Sin(ent.Facing) * d.PatrolSpeedY
			if !g.IsSolid(ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
				ent.Pos = ent.Pos.Add(ent.Vel)
			} else {
				ent.Facing += math.Pi
			}
		}
	case 1: // charging
		ent.Timer++
		displacement := math.Hypot(ent.Pos.X-ent.ChargeOrigin.X, ent.Pos.Y-ent.ChargeOrigin.Y)
		if ent.Timer >= d.ChargeMaxFrames || displacement > d.ChargeMaxDist {
			ent.State = 2
			ent.StunTimer = d.StunFrames
			ent.Vel = gvec.Vec2{}
			break
		}

		currentVel := ent.Vel
		if g.CheckDeterrentSlowing(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y) {
			currentVel = currentVel.Scale(d.DeterrentSlowScale)
		}

		nextX := ent.Pos.X + currentVel.X
		nextY := ent.Pos.Y + currentVel.Y
		if g.IsSolid(nextX, nextY, ent.Dimensions.X, ent.Dimensions.Y) {
			ent.State = 2
			ent.StunTimer = d.StunFrames
			ent.Vel = gvec.Vec2{}
		} else {
			ent.Pos.X = nextX
			ent.Pos.Y = nextY
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
				ent.Vel = gvec.Vec2{}
				ent.State = 2
				ent.StunTimer = d.StunFrames
			} else {
				dirX, dirY := 0.0, 0.0
				speed := math.Hypot(ent.Vel.X, ent.Vel.Y)
				if speed > 0.1 {
					dirX = ent.Vel.X / speed
					dirY = ent.Vel.Y / speed
				} else {
					dx := targetX - ex
					dy := targetY - ey
					dist := math.Hypot(dx, dy)
					if dist > 0.1 {
						dirX = dx / dist
						dirY = dy / dist
					} else {
						dirX = 1.0
					}
				}

				forceVec := gvec.Vec2{X: dirX * d.Knockback, Y: dirY * d.Knockback}

				if g.HasActiveVehicle() {
					g.Emit(DamageActiveVehicleCmd{Amount: d.VehicleDamage})
					g.Emit(KnockbackActiveVehicleCmd{Force: forceVec})
					g.Emit(SetMineWarningCmd{Message: "VEHICLE RAMMED BY THERMOCLINE RAMMER!", Duration: d.WarningDuration, Level: 2})
				} else {
					g.Emit(DamagePlayerCmd{Amount: d.PlayerDamage})
					g.Emit(KnockbackPlayerCmd{Force: forceVec})
					g.Emit(SetMineWarningCmd{Message: "RAMMED BY THERMOCLINE RAMMER!", Duration: d.WarningDuration, Level: 2})
				}

				// Push rammer back in opposite direction to prevent continuous overlap
				ent.Pos.X -= dirX * d.PushBack
				ent.Pos.Y -= dirY * d.PushBack
				ent.Vel = gvec.Vec2{}
				ent.State = 2
				ent.StunTimer = d.StunFrames
			}
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
