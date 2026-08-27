package vehicle

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

const playerSlowScale = 0.5

// ScaleForPower returns force and max speed, using no-power values when battery is empty.
func ScaleForPower(force, maxSpeed, noPowerForce, noPowerMaxSpeed float64, hasPower bool) (float64, float64) {
	if !hasPower {
		return noPowerForce, noPowerMaxSpeed
	}
	return force, maxSpeed
}

// ScaleForSlow halves force and max speed when the player is slowed.
func ScaleForSlow(force, maxSpeed float64, slowed bool) (float64, float64) {
	if slowed {
		return force * playerSlowScale, maxSpeed * playerSlowScale
	}
	return force, maxSpeed
}

// ApplyDragClamp scales velocity by drag and clamps to maxSpeed.
func ApplyDragClamp(vel *gvec.Vec2, drag, maxSpeed float64) {
	*vel = vel.Scale(drag)
	speed := vel.Length()
	if speed > maxSpeed {
		*vel = vel.Scale(maxSpeed / speed)
	}
}

// ApplyLateralDrag decays velocity perpendicular to facing so the craft tracks
// its heading instead of sliding sideways.
func ApplyLateralDrag(vel *gvec.Vec2, facing, keep float64) {
	if keep >= 1 {
		return
	}
	if keep < 0 {
		keep = 0
	}
	fx := math.Cos(facing)
	fy := math.Sin(facing)
	fwd := vel.X*fx + vel.Y*fy
	latX := vel.X - fwd*fx
	latY := vel.Y - fwd*fy
	vel.X = fwd*fx + latX*keep
	vel.Y = fwd*fy + latY*keep
}

// TurnScaleForSpeed returns a 0–1 multiplier that is idleScale at rest and 1 at maxSpeed.
func TurnScaleForSpeed(speed, maxSpeed, idleScale float64) float64 {
	if maxSpeed <= 0 {
		return 1
	}
	frac := speed / maxSpeed
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	return idleScale + (1-idleScale)*frac
}

const analogStickDeadzone = 0.18

type analogStickInput interface {
	StickAxes() (gvec.Vec2, bool)
}

// AnalogAimAxes returns stick direction as a desired facing angle and magnitude as
// throttle. ok is true whenever a stick touch is held (even in the deadzone) so
// callers can ignore digital WASD for that frame.
func AnalogAimAxes(input InputSource) (desiredFacing, throttle float64, ok bool) {
	a, ok := input.(analogStickInput)
	if !ok {
		return 0, 0, false
	}
	vec, held := a.StickAxes()
	if !held {
		return 0, 0, false
	}
	mag := math.Hypot(vec.X, vec.Y)
	if mag <= analogStickDeadzone {
		return 0, 0, true
	}
	t := (mag - analogStickDeadzone) / (1 - analogStickDeadzone)
	if t > 1 {
		t = 1
	}
	return math.Atan2(vec.Y, vec.X), t, true
}

// SteerToward rotates facing toward desired by at most maxStep radians (shortest path).
func SteerToward(facing *float64, desired, maxStep float64) {
	delta := desired - *facing
	for delta > math.Pi {
		delta -= 2 * math.Pi
	}
	for delta < -math.Pi {
		delta += 2 * math.Pi
	}
	if math.Abs(delta) <= maxStep {
		*facing = desired
		return
	}
	if delta > 0 {
		*facing += maxStep
	} else {
		*facing -= maxStep
	}
}

// DrainBatteryOnMove subtracts drain when moving with power; returns whether power remains.
func DrainBatteryOnMove(battery *float64, moving, hasPower bool, drain float64) bool {
	if moving && hasPower {
		*battery -= drain
		if *battery < 0 {
			*battery = 0
		}
	}
	return *battery > 0
}

// ClampBattery keeps battery within [0, maxBattery].
func ClampBattery(battery, maxBattery *float64) {
	if *battery > *maxBattery {
		*battery = *maxBattery
	}
	if *battery < 0 {
		*battery = 0
	}
}

// CursorFacing returns the angle from the player screen center toward the cursor.
func CursorFacing(rt Runtime) float64 {
	input := rt.Input()
	cursor := input.Cursor()
	center := rt.PlayerScreenCenter()
	return math.Atan2(cursor.Y-center.Y, cursor.X-center.X)
}

// ApplyWASDThrust applies cave-style WASD thrust. Up is blocked above waterline.
func ApplyWASDThrust(input InputSource, posY, waterline float64, vel *gvec.Vec2, force float64) (moving bool) {
	if input.IsKeyPressed(ebiten.KeyW) || input.IsKeyPressed(ebiten.KeyArrowUp) {
		if posY > waterline {
			vel.Y -= force
			moving = true
		}
	}
	if input.IsKeyPressed(ebiten.KeyS) || input.IsKeyPressed(ebiten.KeyArrowDown) {
		vel.Y += force
		moving = true
	}
	if input.IsKeyPressed(ebiten.KeyA) || input.IsKeyPressed(ebiten.KeyArrowLeft) {
		vel.X -= force
		moving = true
	}
	if input.IsKeyPressed(ebiten.KeyD) || input.IsKeyPressed(ebiten.KeyArrowRight) {
		vel.X += force
		moving = true
	}
	return moving
}

// WaterlineOpts tunes surface buoyancy and idle bob for cave craft.
type WaterlineOpts struct {
	Waterline       float64
	SurfacePush     float64
	BobSpring       float64
	BobWhenThruster bool // when true, bob only while thrustersActive; scout uses moving instead
}

// ApplyWaterline adjusts vertical velocity for surface buoyancy and bob.
func ApplyWaterline(posY float64, velY, timeOfDay float64, moving, thrustersActive bool, opts WaterlineOpts) float64 {
	if posY < opts.Waterline {
		return velY + opts.SurfacePush
	}
	bobActive := !moving
	if opts.BobWhenThruster {
		bobActive = thrustersActive
	}
	if bobActive && posY < opts.Waterline+16.0 {
		bobY := opts.Waterline + 4.0 + math.Sin(timeOfDay*0.05)*2.0
		return velY + (bobY-posY)*opts.BobSpring
	}
	return velY
}

// MaybeEmitPropBubble spawns a prop bubble behind the craft with the given chance.
func MaybeEmitPropBubble(rt Runtime, pos, dims gvec.Vec2, facing float64, chance float64) {
	if rand.Float64() >= chance {
		return
	}
	propX := pos.X
	if math.Cos(facing) < 0 {
		propX = pos.X + dims.X
	}
	rt.Emit(SpawnBubbleCmd{Pos: gvec.Vec2{X: propX, Y: pos.Y + dims.Y/2.0}})
}

// MoveAxisSeparated runs axis-separated collision resolution.
func MoveAxisSeparated(pos, vel *gvec.Vec2, dims gvec.Vec2, isSolid func(gvec.Vec2) bool, onHitX, onHitY func()) {
	gvec.MoveAxisSeparated(pos, vel, dims, isSolid, onHitX, onHitY)
}

// CaveSpeedImpact applies damage and screen shake when impact speed exceeds minSpeed.
func CaveSpeedImpact(takeDamage func(float64), rt Runtime, absSpeed, minSpeed, damageScale, shakeScale float64) {
	if absSpeed <= minSpeed {
		return
	}
	takeDamage(absSpeed * damageScale)
	rt.Emit(TriggerShakeCmd{Duration: 15, Intensity: absSpeed * shakeScale})
}
