package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type ViperState int

const (
	StateViperPatrol   ViperState = 0
	StateViperWindup   ViperState = 1
	StateViperLunge    ViperState = 2
	StateViperCooldown ViperState = 3
)

type SandViper struct {
	BaseEntity
	State       ViperState
	Timer       int // general timer for state duration (e.g. windup, cooldown, lunge)
	SwayPhase   float64
	FacingRight bool
	LungeDir    gvec.Vec2
}

func NewSandViper(x, y float64) *SandViper {
	return &SandViper{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 24, Y: 12},
			Active:     true,
		},
		State:       StateViperPatrol,
		SwayPhase:   rand.Float64() * math.Pi * 2,
		FacingRight: rand.Float64() < 0.5,
	}
}

func (sv *SandViper) Update(gr Runtime) {
	// Increment sway phase for swimming animation
	sv.SwayPhase += 0.08

	// Find target (player or active vehicle)
	targetPos := gr.PlayerPos()
	targetDims := gr.PlayerDims()
	if gr.HasActiveVehicle() {
		targetPos = gr.ActiveVehiclePos()
		targetDims = gr.ActiveVehicleDims()
	}

	// Calculate center positions
	svCenter := sv.Pos.Add(sv.Dimensions.Scale(0.5))
	targetCenter := targetPos.Add(targetDims.Scale(0.5))
	dist := svCenter.Distance(targetCenter)

	switch sv.State {
	case StateViperPatrol:
		// Patrol slowly near the sea floor (moving left and right)
		sv.Timer++
		speed := 0.5
		if sv.FacingRight {
			sv.Vel.X = speed
		} else {
			sv.Vel.X = -speed
		}
		// Gently bob up and down
		sv.Vel.Y = math.Sin(sv.SwayPhase) * 0.2

		nextPos := sv.Pos.Add(sv.Vel)
		if !gr.IsSolid(nextPos.X, nextPos.Y, sv.Dimensions.X, sv.Dimensions.Y) {
			sv.Pos = nextPos
		} else {
			sv.FacingRight = !sv.FacingRight
		}

		// Proximity check for aggro
		if dist < 100.0 {
			sv.State = StateViperWindup
			sv.Timer = 0
			sv.Vel = gvec.Vec2{} // Stop moving during windup
			// Face the player
			sv.FacingRight = targetCenter.X > svCenter.X
		}

	case StateViperWindup:
		sv.Timer++
		// Stop moving
		sv.Vel = gvec.Vec2{}

		// Turn to face the player/vehicle
		sv.FacingRight = targetCenter.X > svCenter.X

		if sv.Timer >= 30 { // 0.5s at 60 FPS
			sv.State = StateViperLunge
			sv.Timer = 0
			// Calculate straight line vector towards target center
			dirX := targetCenter.X - svCenter.X
			dirY := targetCenter.Y - svCenter.Y
			length := math.Hypot(dirX, dirY)
			if length > 0.01 {
				sv.LungeDir = gvec.Vec2{X: dirX / length, Y: dirY / length}
			} else {
				if sv.FacingRight {
					sv.LungeDir = gvec.Vec2{X: 1, Y: 0}
				} else {
					sv.LungeDir = gvec.Vec2{X: -1, Y: 0}
				}
			}
			// Set lunge speed
			sv.Vel = sv.LungeDir.Scale(4.5)
		}

	case StateViperLunge:
		sv.Timer++

		nextPos := sv.Pos.Add(sv.Vel)
		// If we hit a wall, abort the lunge and enter cooldown
		if gr.IsSolid(nextPos.X, nextPos.Y, sv.Dimensions.X, sv.Dimensions.Y) {
			sv.State = StateViperCooldown
			sv.Timer = 0
			sv.Vel = gvec.Vec2{}
			break
		}
		sv.Pos = nextPos

		// Check player/vehicle overlap
		pDims := gr.PlayerDims()
		pPos := gr.PlayerPos()
		if gr.HasActiveVehicle() {
			pPos = gr.ActiveVehiclePos()
			pDims = gr.ActiveVehicleDims()
		}

		if rectsOverlap(sv.Pos.X, sv.Pos.Y, sv.Dimensions.X, sv.Dimensions.Y, pPos.X, pPos.Y, pDims.X, pDims.Y) {
			// Calculate knockback force (in direction of lunge)
			kbForce := 3.5
			forceVec := sv.LungeDir.Scale(kbForce)

			if gr.HasActiveVehicle() {
				gr.Emit(DamageActiveVehicleCmd{Amount: 10.0})
				gr.Emit(KnockbackActiveVehicleCmd{Force: forceVec})
				gr.Emit(SetMineWarningCmd{Message: "VEHICLE NIPPED BY SAND-VIPER!", Duration: 90, Level: 1})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: 10.0})
				gr.Emit(KnockbackPlayerCmd{Force: forceVec})
				gr.Emit(SetMineWarningCmd{Message: "NIPPED BY SAND-VIPER!", Duration: 90, Level: 1})
			}

			// Push back slightly to avoid double hits
			sv.Pos = gvec.Vec2{
				X: sv.Pos.X - sv.LungeDir.X*15.0,
				Y: sv.Pos.Y - sv.LungeDir.Y*15.0,
			}
			sv.State = StateViperCooldown
			sv.Timer = 0
			sv.Vel = gvec.Vec2{}
		} else if sv.Timer >= 20 { // Lunge lasts 20 frames max
			sv.State = StateViperCooldown
			sv.Timer = 0
			sv.Vel = gvec.Vec2{}
		}

	case StateViperCooldown:
		sv.Timer++
		// Drift slowly downwards, simulating settling back down to seabed
		sv.Vel = gvec.Vec2{X: 0, Y: 0.25}
		nextPos := sv.Pos.Add(sv.Vel)
		if !gr.IsSolid(nextPos.X, nextPos.Y, sv.Dimensions.X, sv.Dimensions.Y) {
			sv.Pos = nextPos
		}

		if sv.Timer >= 120 { // 2 seconds cooldown
			sv.State = StateViperPatrol
			sv.Timer = 0
		}
	}
}

func (sv *SandViper) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(sv.Pos.X - camera.Pos.X)
	sy := float32(sv.Pos.Y - camera.Pos.Y)
	sw := float32(sv.Dimensions.X)
	sh := float32(sv.Dimensions.Y)
	cx := sx + sw/2
	cy := sy + sh/2

	// Sand-viper color: sandy-gold/tan #D2B48C (RGB: 210, 180, 140)
	bodyColor := color.RGBA{210, 180, 140, 255}
	darkerBodyColor := color.RGBA{180, 150, 110, 255}
	eyeColor := color.RGBA{255, 235, 60, 255}

	// Calculate heading direction
	var dir float64 = -1.0
	if sv.FacingRight {
		dir = 1.0
	}

	// If in windup, wobble/twitch to show tension
	var windupX, windupY float32 = 0, 0
	if sv.State == StateViperWindup {
		windupX = float32(math.Sin(float64(sv.Timer)*0.8)) * 1.5
		windupY = float32(math.Cos(float64(sv.Timer)*0.8)) * 1.0
	}

	// Draw 5-6 segments wiggling like a snake
	numSegments := 6
	for i := numSegments - 1; i >= 0; i-- {
		t := float64(i) / float64(numSegments-1)
		// segments get smaller toward the tail
		radius := float32(5.0 - t*3.0)
		if i == 0 {
			radius = 5.5 // slightly larger head
		}

		// Calculate segment displacement
		// Tail wiggles with a sine wave based on time and segment index
		var wiggle float64 = 0.0
		if sv.State == StateViperLunge {
			// Fast wiggle during lunge
			wiggle = math.Sin(timeOfDay*0.5+float64(i)*1.2) * 4.0
		} else if sv.State == StateViperCooldown {
			// Barely wiggles in cooldown
			wiggle = math.Sin(timeOfDay*0.05+float64(i)*0.4) * 0.5
		} else {
			// Normal sway during patrol/windup
			wiggle = math.Sin(timeOfDay*0.15+float64(i)*0.8) * 2.5
		}

		segX := cx - float32(dir)*float32(i)*3.5 + windupX
		segY := cy + float32(wiggle) + windupY

		if i == 0 {
			// Head segment
			vector.FillCircle(screen, segX, segY, radius, bodyColor, false)
			// Draw two glowing eyes
			eyeOffsetX := float32(dir) * 1.5
			vector.FillCircle(screen, segX+eyeOffsetX, segY-2.0, 1.2, eyeColor, false)
			vector.FillCircle(screen, segX+eyeOffsetX, segY+2.0, 1.2, eyeColor, false)
		} else {
			// Body segments alternated colored
			clr := bodyColor
			if i%2 == 1 {
				clr = darkerBodyColor
			}
			vector.FillCircle(screen, segX, segY, radius, clr, false)
		}
	}
}
