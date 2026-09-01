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

// WeaverState represents the behavioral state of an ElectroWeaver.
type WeaverState int

const (
	WeaverStateIdle WeaverState = iota
	WeaverStateTracking
	WeaverStateLunge
	WeaverStateCooldown
)

// ElectroWeaver is a serpentine predator that tracks electrical sources and strikes.
type ElectroWeaver struct {
	BaseEntity
	def           *ElectroWeaverDef
	Timer         int
	Facing        float64
	State         WeaverState
	LungeDir      gvec.Vec2
	LungeTimer    int
	CooldownTimer int
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
		def:   d,
		State: WeaverStateIdle,
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

// hasLineOfSight performs step raycasting along the segment between from and to.
// Returns false if any solid terrain blocks the path.
func hasLineOfSight(g WeaverContext, from, to gvec.Vec2) bool {
	dx := to.X - from.X
	dy := to.Y - from.Y
	dist := math.Hypot(dx, dy)
	if dist < 4.0 {
		return true
	}
	steps := int(dist / 16.0)
	if steps < 2 {
		steps = 2
	}
	for i := 1; i < steps; i++ {
		t := float64(i) / float64(steps)
		x := from.X + dx*t
		y := from.Y + dy*t
		if g.IsSolid(x-3, y-3, 6, 6) {
			return false
		}
	}
	return true
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
		ent.State = WeaverStateIdle
		g.Emit(UpdateWeaverTrackingTimerCmd{Value: 0})
		return
	}

	// 1. Handle Lunge Strike State (high-speed physical charge; zero teleportation)
	if ent.State == WeaverStateLunge {
		ent.LungeTimer--
		lungeSpeed := d.LungeSpeed
		if lungeSpeed <= 0 {
			lungeSpeed = 8.5
		}
		step := ent.LungeDir.Scale(lungeSpeed)
		nextPos := ent.Pos.Add(step)

		// Check collision with target
		collided := false
		if isDecoy {
			if math.Hypot(targetX-(nextPos.X+ent.Dimensions.X/2), targetY-(nextPos.Y+ent.Dimensions.Y/2)) < 36.0 {
				g.Emit(DestroyDecoyCmd{Pos: gvec.Vec2{X: targetX, Y: targetY}})
				g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKES DECOY!", Duration: d.DecoyWarningDuration, Level: 1})
				collided = true
			}
		} else {
			pPos := g.PlayerPos()
			pDims := g.PlayerDims()
			nextCX := nextPos.X + ent.Dimensions.X/2.0
			nextCY := nextPos.Y + ent.Dimensions.Y/2.0
			if (nextCX >= pPos.X-16 && nextCX <= pPos.X+pDims.X+16 &&
				nextCY >= pPos.Y-16 && nextCY <= pPos.Y+pDims.Y+16) ||
				math.Hypot(px-nextCX, py-nextCY) < 32.0 {
				g.Emit(DamagePlayerCmd{Amount: d.PlayerDamage, Kind: DamageElectric})
				g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKE! SEVERE DAMAGE!", Duration: d.WarningDuration, Level: 3})
				collided = true
			}
		}

		if collided {
			ent.State = WeaverStateCooldown
			cooldownFrames := d.CooldownFrames
			if cooldownFrames <= 0 {
				cooldownFrames = 180
			}
			ent.CooldownTimer = cooldownFrames
			ent.Timer = 0
			ent.LungeTimer = 0
			ent.Vel = ent.LungeDir.Scale(-d.IdleSpeed)
			g.Emit(UpdateWeaverTrackingTimerCmd{Value: 0})
			return
		}

		// Wall collision during lunge
		if g.IsSolid(nextPos.X, nextPos.Y, ent.Dimensions.X, ent.Dimensions.Y) {
			ent.State = WeaverStateCooldown
			cooldownFrames := d.CooldownFrames
			if cooldownFrames <= 0 {
				cooldownFrames = 180
			}
			ent.CooldownTimer = cooldownFrames
			ent.Timer = 0
			ent.LungeTimer = 0
			ent.Vel = ent.LungeDir.Scale(-d.IdleSpeed)
			g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER MISSED!", Duration: 90, Level: 1})
			g.Emit(UpdateWeaverTrackingTimerCmd{Value: 0})
			return
		}

		ent.Pos = nextPos
		ent.Facing = math.Atan2(ent.LungeDir.Y, ent.LungeDir.X)

		if ent.LungeTimer <= 0 {
			ent.State = WeaverStateCooldown
			cooldownFrames := d.CooldownFrames
			if cooldownFrames <= 0 {
				cooldownFrames = 180
			}
			ent.CooldownTimer = cooldownFrames
			ent.Timer = 0
			ent.Vel = ent.LungeDir.Scale(-d.IdleSpeed)
			g.Emit(UpdateWeaverTrackingTimerCmd{Value: 0})
		}
		return
	}

	// 2. Handle Cooldown State (smoothly retreats into the shadows)
	if ent.State == WeaverStateCooldown {
		ent.CooldownTimer--
		ent.Timer = 0
		g.Emit(UpdateWeaverTrackingTimerCmd{Value: 0})

		// Back away slowly from target
		retreatVel := ent.LungeDir.Scale(-d.IdleSpeed)
		if !g.IsSolid(ent.Pos.X+retreatVel.X, ent.Pos.Y+retreatVel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
			ent.Pos = ent.Pos.Add(retreatVel)
		}

		if ent.CooldownTimer <= 0 {
			ent.State = WeaverStateIdle
		}
		return
	}

	// 3. Handle Idle & Tracking States (strict line of sight)
	isElectricity := g.FlashlightOn() || g.SonarActive() || g.HasActiveVehicle() || isDecoy

	// Deterrent cloud occlusion + lights off immediately breaks lock
	if !isDecoy && !g.FlashlightOn() && g.CheckDeterrentOcclusion(gvec.Vec2{X: ex, Y: ey}, gvec.Vec2{X: px, Y: py}) {
		ent.Timer = 0
		ent.State = WeaverStateIdle
		g.Emit(UpdateWeaverTrackingTimerCmd{Value: 0})
		return
	}

	// Check line of sight
	hasLoS := hasLineOfSight(g, gvec.Vec2{X: ex, Y: ey}, gvec.Vec2{X: targetX, Y: targetY})
	if !isDecoy && g.CheckDeterrentOcclusion(gvec.Vec2{X: ex, Y: ey}, gvec.Vec2{X: px, Y: py}) {
		hasLoS = false
	}

	if isElectricity && dist < d.TrackRange && hasLoS {
		ent.State = WeaverStateTracking
		ent.Timer++
		g.Emit(UpdateWeaverTrackingTimerCmd{Value: float64(ent.Timer)})

		ent.Facing = math.Atan2(targetY-ey, targetX-ex)

		// Audio crackle telegraph: 1-2 seconds (60-120 frames) before strike
		telegraphWindow := d.StrikeTimerFrames - 120
		if telegraphWindow < 60 {
			telegraphWindow = 60
		}
		if ent.Timer >= telegraphWindow && (ent.Timer-telegraphWindow)%60 == 0 {
			g.Emit(PlaySFXCmd{Path: "sfx/weaver_charge.wav", Volume: 0.6})
		}

		// 100% Charge Reached -> Execute High-Speed Lunge Strike
		if ent.Timer >= d.StrikeTimerFrames {
			ent.State = WeaverStateLunge
			ent.LungeTimer = 55 // up to 55 frames of charge at 8.5 speed = ~460px
			if dist > 0.001 {
				ent.LungeDir = gvec.Vec2{X: (targetX - ex) / dist, Y: (targetY - ey) / dist}
			} else {
				angle := rand.Float64() * math.Pi * 2
				ent.LungeDir = gvec.Vec2{X: math.Cos(angle), Y: math.Sin(angle)}
			}
			ent.Facing = math.Atan2(ent.LungeDir.Y, ent.LungeDir.X)
			ent.Timer = 0
			g.Emit(UpdateWeaverTrackingTimerCmd{Value: 0})
			g.Emit(PlaySFXCmd{Path: "sfx/weaver_shock.wav", Volume: 0.8})

			// If already right on top of target, resolve collision immediately
			if isDecoy && dist < 36.0 {
				g.Emit(DestroyDecoyCmd{Pos: gvec.Vec2{X: targetX, Y: targetY}})
				g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKES DECOY!", Duration: d.DecoyWarningDuration, Level: 1})
				ent.State = WeaverStateCooldown
				ent.CooldownTimer = 180
			} else if !isDecoy && dist < 32.0 {
				g.Emit(DamagePlayerCmd{Amount: d.PlayerDamage, Kind: DamageElectric})
				g.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKE! SEVERE DAMAGE!", Duration: d.WarningDuration, Level: 3})
				ent.State = WeaverStateCooldown
				ent.CooldownTimer = 180
			}
			return
		}
	} else {
		// Line of sight lost or electricity turned off -> drain threat bar
		if ent.Timer > 0 {
			ent.Timer -= d.TimerDecay
			if ent.Timer <= 0 {
				ent.Timer = 0
				ent.State = WeaverStateIdle
			}
			g.Emit(UpdateWeaverTrackingTimerCmd{Value: float64(ent.Timer)})
		}
	}

	// Movement during tracking vs idle
	if ent.State == WeaverStateTracking && ent.Timer > d.MoveStartTimer {
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

	isLunging := (ent.State == WeaverStateLunge)

	// Lunge high-voltage plasma aura
	if isLunging {
		auraRadius := float32(22.0)
		vector.FillCircle(screen, cx, cy, auraRadius, color.RGBA{100, 200, 255, 60}, false)
		vector.FillCircle(screen, cx, cy, auraRadius*0.6, color.RGBA{180, 240, 255, 120}, false)
	}

	for i := range 5 {
		lag := float64(i) * 0.3
		tVal := timeOfDay*0.08 - lag
		offX := math.Cos(tVal) * 6
		offY := math.Sin(tVal) * 4
		segmentX := cx - float32(math.Cos(ent.Facing)*float64(i)*8.0) + float32(offX)
		segmentY := cy - float32(math.Sin(ent.Facing)*float64(i)*8.0) + float32(offY)

		segColor := color.RGBA{140 - uint8(i*18), 45, 205 - uint8(i*12), 255}
		if isLunging {
			segColor = color.RGBA{180 + uint8(i*15), 230, 255, 255}
		}

		segRadius := 6.0 - float32(i)*0.8
		vector.FillCircle(screen, segmentX, segmentY, segRadius, segColor, false)

		if i == 0 {
			eyeColor := color.RGBA{255, 255, 80, 255}
			if isLunging {
				eyeColor = color.RGBA{255, 60, 60, 255}
			}
			vector.FillCircle(screen, segmentX+float32(math.Cos(ent.Facing))*4, segmentY+float32(math.Sin(ent.Facing))*4, 2.0, eyeColor, false)
		}
	}

	// Sparking electric discharge arcs
	if isLunging || ent.Timer > 0 {
		sparkCount := 3
		if !isLunging {
			sparkRatio := float64(ent.Timer) / float64(d.StrikeTimerFrames)
			sparkCount = int(sparkRatio * 6)
		} else {
			sparkCount = 8
		}

		for s := 0; s < sparkCount; s++ {
			spx := cx + float32(rand.Intn(48)-24)
			spy := cy + float32(rand.Intn(48)-24)
			sparkClr := color.RGBA{160, 230, 255, 240}
			if isLunging {
				sparkClr = color.RGBA{255, 255, 200, 255}
			}
			vector.StrokeLine(screen, cx, cy, spx, spy, 1.2, sparkClr, false)
		}
	}
}
