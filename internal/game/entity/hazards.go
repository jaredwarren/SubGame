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

// ShatterBulb is a static oxygen plant that pops when touched, restoring O2.
type ShatterBulb struct {
	BaseEntity
}

func (s *ShatterBulb) Update(gr Runtime) {
	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}
	if rectsOverlap(s.Pos.X, s.Pos.Y, s.Dimensions.X, s.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		s.Pop(gr)
	}
}

// Pop deactivates the bulb, restoring oxygen and emitting a sound wave.
func (s *ShatterBulb) Pop(gr Runtime) {
	if !s.Active {
		return
	}
	s.Active = false
	gr.Emit(RestoreOxygenCmd{Amount: 20})
	gr.Emit(TriggerSoundWaveCmd{
		Pos: gvec.Vec2{X: s.Pos.X + s.Dimensions.X/2.0, Y: s.Pos.Y + s.Dimensions.Y/2.0},
	})
}

func (s *ShatterBulb) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(s.Pos.X - camera.Pos.X)
	sy := float32(s.Pos.Y - camera.Pos.Y)
	sw := float32(s.Dimensions.X)
	sh := float32(s.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	vector.StrokeLine(screen, cx, cy, cx, cy+16, 2.0, color.RGBA{45, 95, 75, 255}, false)
	phase := s.Pos.X + s.Pos.Y
	pulse := float32(math.Cos(timeOfDay*0.02+phase)) * 2.5
	radius := float32(11.0) + pulse
	if radius < 5.0 {
		radius = 5.0
	}
	vector.FillCircle(screen, cx, cy, radius, color.RGBA{0, 220, 240, 60}, false)
	vector.FillCircle(screen, cx, cy, 7, color.RGBA{0, 230, 245, 255}, false)
	vector.StrokeCircle(screen, cx, cy, 7, 0.8, color.RGBA{255, 255, 255, 200}, false)
}

// FalseBulbSnare mimics a ShatterBulb but lunges at and damages the player.
type FalseBulbSnare struct {
	BaseEntity
	State int
}

func (ent *FalseBulbSnare) Update(gr Runtime) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2.0
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0
	dist := math.Hypot(px-ex, py-ey)

	if dist > 360.0 {
		ent.State = 0
		return
	}

	isLit := false
	if gr.FlashlightOn() {
		facingAngle := gr.PlayerFacing()
		if gr.HasActiveVehicle() {
			facingAngle = gr.ActiveVehicleFacing()
		}
		dx := ex - px
		dy := ey - py
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

	soundAlerted := gr.SoundWaveTimer() > 0 && math.Hypot(gr.SoundWaveX()-ex, gr.SoundWaveY()-ey) < 280.0
	if soundAlerted {
		ent.State = 1
	}

	if isLit {
		ent.Vel = gvec.Vec2{}
	} else {
		if dist < 180.0 || ent.State == 1 {
			ent.State = 1
			dx := px - ex
			dy := py - ey
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

	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}
	if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		gr.Emit(DamagePlayerCmd{Amount: 20.0})
		gr.Emit(SetMineWarningCmd{Message: "ATTACKED BY FALSE-BULB SNARE!", Duration: 120, Level: 2})
		ent.Active = false
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

// BrimstoneSiphon is a volcanic vent that fires damaging thermal jets.
type BrimstoneSiphon struct {
	BaseEntity
	Timer     int
	Direction string // "up", "down", "left", "right"
}

func (ent *BrimstoneSiphon) Update(gr Runtime) {
	ent.Timer = (ent.Timer + 1) % 120
	if ent.Timer >= 60 {
		var jx, jy, jw, jh float64
		const jetRange = 160.0

		switch ent.Direction {
		case "up":
			jx, jy, jw, jh = ent.Pos.X, ent.Pos.Y-jetRange, ent.Dimensions.X, jetRange
		case "down":
			jx, jy, jw, jh = ent.Pos.X, ent.Pos.Y+ent.Dimensions.Y, ent.Dimensions.X, jetRange
		case "left":
			jx, jy, jw, jh = ent.Pos.X-jetRange, ent.Pos.Y, jetRange, ent.Dimensions.Y
		default:
			jx, jy, jw, jh = ent.Pos.X+ent.Dimensions.X, ent.Pos.Y, jetRange, ent.Dimensions.Y
		}

		vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
		targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
		if gr.HasActiveVehicle() {
			vPos := gr.ActiveVehiclePos()
			targetX, targetY = vPos.X, vPos.Y
			vDims := gr.ActiveVehicleDims()
			vWidth, vHeight = vDims.X, vDims.Y
		}

		if rectsOverlap(jx, jy, jw, jh, targetX, targetY, vWidth, vHeight) {
			if gr.HasActiveVehicle() {
				gr.Emit(DamageActiveVehicleCmd{Amount: 0.4})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: 0.6})
			}
		}
	}
}

func (ent *BrimstoneSiphon) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0

	entityPath.Reset()
	entityPath.MoveTo(cx-16, sy+32)
	entityPath.LineTo(cx+16, sy+32)
	entityPath.LineTo(cx+8, sy+12)
	entityPath.LineTo(cx-8, sy+12)
	entityPath.Close()

	var ventColor color.RGBA
	if ent.Timer >= 60 {
		ventColor = color.RGBA{185, 85, 45, 255}
	} else {
		ventColor = color.RGBA{65, 55, 50, 255}
	}
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(ventColor)
	vector.FillPath(screen, entityPath, nil, &opts)

	if ent.Timer >= 60 {
		jetLen := float32(120.0)
		switch ent.Direction {
		case "up":
			vector.FillRect(screen, cx-8, sy-jetLen+float32(sh)/2, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-3, sy-jetLen+float32(sh)/2, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
		case "down":
			vector.FillRect(screen, cx-8, sy+16, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-3, sy+16, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
		case "left":
			vector.FillRect(screen, cx-jetLen-16, sy-8+float32(sh)/2, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-jetLen-26, sy-3+float32(sh)/2, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
		default:
			vector.FillRect(screen, cx+16, sy-8+float32(sh)/2, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx+16, sy-3+float32(sh)/2, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
		}
	}
}

// ThermoclineRammer is a fast-charging aquatic predator that rams the player.
type ThermoclineRammer struct {
	BaseEntity
	State     int
	Timer     int
	Facing    float64
	StunTimer int
}

func (ent *ThermoclineRammer) Update(gr Runtime) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2.0
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2.0
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
		if !gr.HasActiveVehicle() && gr.IsPlayerSprinting() && (math.Abs(gr.PlayerVel().X) > 1.2 || math.Abs(gr.PlayerVel().Y) > 1.2) {
			isAggroTrigger = true
		}
		if gr.HasActiveVehicle() && gr.ActiveVehicleMoving() {
			isAggroTrigger = true
		}
	}
	if gr.SoundWaveTimer() > 0 && math.Hypot(gr.SoundWaveX()-ex, gr.SoundWaveY()-ey) < 250.0 {
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
			if !gr.IsSolid(ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
				ent.Pos = ent.Pos.Add(ent.Vel)
			} else {
				ent.Facing += math.Pi
			}
		}
	case 1: // charging
		nextX := ent.Pos.X + ent.Vel.X
		nextY := ent.Pos.Y + ent.Vel.Y
		if gr.IsSolid(nextX, nextY, ent.Dimensions.X, ent.Dimensions.Y) {
			ent.State = 2
			ent.StunTimer = 180
			ent.Vel = gvec.Vec2{}
		} else {
			ent.Pos.X = nextX
			ent.Pos.Y = nextY
		}

		vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
		targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
		if gr.HasActiveVehicle() {
			vPos := gr.ActiveVehiclePos()
			targetX, targetY = vPos.X, vPos.Y
			vDims := gr.ActiveVehicleDims()
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

			if gr.HasActiveVehicle() {
				gr.Emit(DamageActiveVehicleCmd{Amount: 30.0})
				gr.Emit(KnockbackActiveVehicleCmd{Force: forceVec})
				gr.Emit(SetMineWarningCmd{Message: "VEHICLE RAMMED BY THERMOCLINE RAMMER!", Duration: 120, Level: 2})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: 25.0})
				gr.Emit(KnockbackPlayerCmd{Force: forceVec})
				gr.Emit(SetMineWarningCmd{Message: "RAMMED BY THERMOCLINE RAMMER!", Duration: 120, Level: 2})
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

// NerveMat is a floor carpet that slows the player on contact.
type NerveMat struct {
	BaseEntity
}

func (ent *NerveMat) Update(gr Runtime) {
	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}
	if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		gr.Emit(SetPlayerSlowedCmd{Slowed: true})
	}
}

func (ent *NerveMat) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)

	vector.FillRect(screen, sx, sy+sh-4, sw, 4, color.RGBA{80, 25, 120, 255}, false)
	for o := float32(4); o < sw; o += 12 {
		vector.StrokeLine(screen, sx+o, sy+sh, sx+o, sy+sh-8, 1.5, color.RGBA{130, 40, 180, 255}, false)
		vector.FillCircle(screen, sx+o, sy+sh-8, 2.0, color.RGBA{180, 60, 220, 255}, false)
	}
}

// ElectroWeaver is a serpentine predator that tracks electrical sources and strikes.
type ElectroWeaver struct {
	BaseEntity
	Timer  int
	Facing float64
}

func (ent *ElectroWeaver) Update(gr Runtime) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2.0
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0
	dist := math.Hypot(px-ex, py-ey)

	inAbyssal := (py / config.TileSize) >= 80
	if !inAbyssal {
		ent.Timer = 0
		return
	}

	isElectricity := gr.FlashlightOn() || gr.SonarActive() || gr.HasActiveVehicle()
	if isElectricity && dist < 500.0 {
		ent.Timer++
		gr.Emit(UpdateWeaverTrackingTimerCmd{Value: float64(ent.Timer)})
		if ent.Timer >= 300 {
			gr.Emit(DamagePlayerCmd{Amount: 45.0})
			gr.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKE! SEVERE DAMAGE!", Duration: 180, Level: 3})
			ent.Pos.X = gr.PlayerPos().X + float64(rand.Intn(120)-60)
			ent.Pos.Y = gr.PlayerPos().Y + float64(rand.Intn(120)-60)
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
			ent.Vel.X = math.Cos(gr.TimeOfDay()/30.0) * 1.2
			ent.Vel.Y = math.Sin(gr.TimeOfDay()/30.0) * 1.2
		}
	} else {
		ent.Vel.X = math.Cos(gr.TimeOfDay()/40.0) * 0.8
		ent.Vel.Y = math.Sin(gr.TimeOfDay()/40.0) * 0.8
	}

	if !gr.IsSolid(ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
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

	for i := 0; i < 5; i++ {
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
