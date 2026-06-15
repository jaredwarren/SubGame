package vehicle

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

var (
	scoutSubSheet *ebiten.Image
)

// SonarSettings configures sonar behaviour for a vehicle.
type SonarSettings struct {
	BatteryCost float64
	Pulse       SonarPulse
}

// ScoutSub is a cave-capable mini-submarine with sonar and mid-range depth.
type ScoutSub struct {
	Pos        gvec.Vec2
	Vel        gvec.Vec2
	Dimensions gvec.Vec2
	Facing     float64
	Health     float64
	MaxHealth  float64
	Battery    float64
	MaxBattery float64
	Cargo      *item.Inventory
	Upgrades   *item.Inventory
	Sonar      SonarSettings
}

// NewScoutSub creates a ScoutSub at the given world position.
func NewScoutSub(x, y float64) *ScoutSub {
	upg := item.NewInventory(2)
	upg.AddItem(&item.SonarAmplifier{}, 1)
	return &ScoutSub{
		Pos:        gvec.Vec2{X: x, Y: y},
		Dimensions: gvec.Vec2{X: 48, Y: 32},
		Health:     100.0,
		MaxHealth:  100.0,
		Battery:    100.0,
		MaxBattery: 100.0,
		Cargo:      item.NewInventory(12),
		Upgrades:   upg,
		Sonar: SonarSettings{
			BatteryCost: 10.0,
			Pulse:       SonarPulse{DurationTicks: 180, RadiusStep: 6.5},
		},
	}
}

func (sub *ScoutSub) GetPos() gvec.Vec2            { return sub.Pos }
func (sub *ScoutSub) SetPos(pos gvec.Vec2)         { sub.Pos = pos }
func (sub *ScoutSub) GetDimensions() gvec.Vec2     { return sub.Dimensions }
func (sub *ScoutSub) GetHealth() float64           { return sub.Health }
func (sub *ScoutSub) GetMaxHealth() float64        { return sub.MaxHealth }
func (sub *ScoutSub) GetOxygen() float64           { return 100.0 }
func (sub *ScoutSub) GetDepthLimit() float64       { return 60.0 }
func (sub *ScoutSub) GetCargo() *item.Inventory    { return sub.Cargo }
func (sub *ScoutSub) GetUpgrades() *item.Inventory { return sub.Upgrades }
func (sub *ScoutSub) GetPerspective() string       { return "cave" }
func (sub *ScoutSub) GetName() string              { return "Scout Sub" }
func (sub *ScoutSub) GetBattery() float64          { return sub.Battery }
func (sub *ScoutSub) GetMaxBattery() float64       { return sub.MaxBattery }
func (sub *ScoutSub) GetFacing() float64           { return sub.Facing }
func (sub *ScoutSub) ApplyForce(force gvec.Vec2) {
	sub.Vel = sub.Vel.Add(force)
}
func (sub *ScoutSub) GetKit() item.Item { return &ScoutSubKit{} }


func (sub *ScoutSub) TakeDamage(amount float64) {
	sub.Health -= amount
	if sub.Health < 0 {
		sub.Health = 0
	}
}

func (sub *ScoutSub) RechargeBattery(amount float64) {
	sub.Battery += amount
	if sub.Battery > sub.MaxBattery {
		sub.Battery = sub.MaxBattery
	}
}

func (sub *ScoutSub) hasUpgrade() bool {
	return item.HasItem[*item.SonarAmplifier](sub.Upgrades, 1)
}

func (sub *ScoutSub) Update(runtime Runtime) {
	if item.HasItem[*item.ThermalGenerator](sub.Upgrades, 1) {
		sub.Battery += 0.02
		if sub.Battery > sub.MaxBattery {
			sub.Battery = sub.MaxBattery
		}
	}

	if !runtime.IsActiveVehicle(sub) {
		sub.Vel = gvec.Vec2{}
		return
	}

	if runtime.PlayerStunned() {
		sub.Vel = gvec.Vec2{}
		return
	}

	input := runtime.Input()
	cursor := input.Cursor()
	center := runtime.PlayerScreenCenter()
	sub.Facing = math.Atan2(cursor.Y-center.Y, cursor.X-center.X)

	var force = 0.20
	var maxSpeed = 4.5
	const drag = 0.94

	hasPower := sub.Battery > 0
	if !hasPower {
		force = 0.04
		maxSpeed = 1.0
	}
	if runtime.PlayerSlowed() {
		force *= 0.5
		maxSpeed *= 0.5
	}

	const waterline = -8.0
	moving := false
	if input.IsKeyPressed(ebiten.KeyW) || input.IsKeyPressed(ebiten.KeyArrowUp) {
		if sub.Pos.Y > waterline {
			sub.Vel.Y -= force
			moving = true
		}
	}
	if input.IsKeyPressed(ebiten.KeyS) || input.IsKeyPressed(ebiten.KeyArrowDown) {
		sub.Vel.Y += force
		moving = true
	}
	if input.IsKeyPressed(ebiten.KeyA) || input.IsKeyPressed(ebiten.KeyArrowLeft) {
		sub.Vel.X -= force
		moving = true
	}
	if input.IsKeyPressed(ebiten.KeyD) || input.IsKeyPressed(ebiten.KeyArrowRight) {
		sub.Vel.X += force
		moving = true
	}

	if moving && hasPower {
		sub.Battery -= 0.03
		if sub.Battery < 0 {
			sub.Battery = 0
		}
		if rand.Float64() < 0.35 {
			propX := sub.Pos.X
			if math.Cos(sub.Facing) < 0 {
				propX = sub.Pos.X + sub.Dimensions.X
			}
			runtime.Emit(SpawnBubbleCmd{Pos: gvec.Vec2{X: propX, Y: sub.Pos.Y + sub.Dimensions.Y/2.0}})
		}
	}

	sub.Vel = sub.Vel.Scale(drag)
	speed := sub.Vel.Length()
	if speed > maxSpeed {
		sub.Vel = sub.Vel.Scale(maxSpeed / speed)
	}

	if sub.Pos.Y < waterline {
		sub.Vel.Y += 0.15
	} else if !moving && sub.Pos.Y < waterline+16.0 {
		bobY := waterline + 4.0 + math.Sin(float64(runtime.TimeOfDay())*0.05)*2.0
		sub.Vel.Y += (bobY - sub.Pos.Y) * 0.03
	}

	sub.checkCollisions(runtime)

	if hasPower && input.IsKeyJustPressed(ebiten.KeyQ) {
		if runtime.CanUseSonar() && sub.Battery >= sub.Sonar.BatteryCost {
			sub.Battery -= sub.Sonar.BatteryCost
			if sub.Battery < 0 {
				sub.Battery = 0
			}
			pulse := sub.Sonar.Pulse
			isUpgraded := sub.hasUpgrade()
			if isUpgraded {
				pulse.DurationTicks = int(float64(pulse.DurationTicks) * 1.8)
				pulse.RadiusStep = pulse.RadiusStep * 1.4
			}
			runtime.Emit(ActivateSonarCmd{
				Source: gvec.Vec2{X: sub.Pos.X + sub.Dimensions.X/2.0, Y: sub.Pos.Y + sub.Dimensions.Y/2.0},
				Pulse:  pulse,
				Bright: isUpgraded,
			})
		}
	}
}

func (sub *ScoutSub) checkCollisions(runtime Runtime) {
	gvec.MoveAxisSeparated(&sub.Pos, &sub.Vel, sub.Dimensions, func(pos gvec.Vec2) bool {
		return sub.isSolid(runtime, pos)
	}, func() {
		speed := math.Abs(sub.Vel.X)
		if speed > 2.0 {
			sub.TakeDamage(speed * 4.0)
			runtime.Emit(TriggerShakeCmd{Duration: 15, Intensity: speed * 2.0})
		}
	}, func() {
		speed := math.Abs(sub.Vel.Y)
		if speed > 2.0 {
			sub.TakeDamage(speed * 4.0)
			runtime.Emit(TriggerShakeCmd{Duration: 15, Intensity: speed * 2.0})
		}
	})
}

func (sub *ScoutSub) isSolid(runtime Runtime, pos gvec.Vec2) bool {
	return solidAt(runtime.IsCaveSolidAt, pos, sub.Dimensions)
}

func (sub *ScoutSub) Draw(screen *ebiten.Image, camX, camY float64) {
	sx := float32(sub.Pos.X - camX)
	sy := float32(sub.Pos.Y - camY)
	w := float32(sub.Dimensions.X)
	h := float32(sub.Dimensions.Y)

	isFacingRight := math.Cos(sub.Facing) >= 0

	if scoutSubSheet != nil {
		rect := image.Rect(481, 141, 2752, 1472)
		sprite := scoutSubSheet.SubImage(rect).(*ebiten.Image)

		op := &ebiten.DrawImageOptions{}

		// Center the cropped sprite on the origin (0, 0)
		op.GeoM.Translate(-1135.5, -665.5)

		// The original illustration faces left. If we face right, flip it.
		facingSign := 1.0
		if isFacingRight {
			facingSign = -1.0
		}

		// Scale so the cropped sprite has a draw width of 64.0 pixels
		const frameScale = 64.0 / 2271.0
		op.GeoM.Scale(facingSign*frameScale, frameScale)

		// Translate to screen coordinates, centered on the sub's collision box center
		op.GeoM.Translate(float64(sx)+float64(w)/2.0, float64(sy)+float64(h)/2.0)

		screen.DrawImage(sprite, op)
		return
	}

	// Fallback to original vector drawing code
	subBgClr := color.RGBA{15, 160, 185, 255}
	domeClr := color.RGBA{80, 205, 255, 180}
	outlineClr := color.RGBA{240, 240, 250, 255}

	vector.FillRect(screen, sx+4, sy+4, w-8, h-8, subBgClr, false)
	vector.StrokeRect(screen, sx+4, sy+4, w-8, h-8, 1.5, outlineClr, false)

	if isFacingRight {
		vector.FillRect(screen, sx+w-12, sy+6, 8, h-12, domeClr, false)
		vector.StrokeRect(screen, sx+w-12, sy+6, 8, h-12, 1.0, color.RGBA{255, 255, 255, 255}, false)
		vector.FillRect(screen, sx, sy+h/2.0-8, 4, 16, color.RGBA{220, 100, 30, 255}, false)
	} else {
		vector.FillRect(screen, sx, sy+6, 8, h-12, domeClr, false)
		vector.StrokeRect(screen, sx, sy+6, 8, h-12, 1.0, color.RGBA{255, 255, 255, 255}, false)
		vector.FillRect(screen, sx+w-4, sy+h/2.0-8, 4, 16, color.RGBA{220, 100, 30, 255}, false)
	}

	vector.FillCircle(screen, sx+w/2.0, sy+h/2.0, 5, color.RGBA{20, 30, 50, 255}, false)
}

// ScoutSubKit represents the deployable kit for the Scout Submarine.
type ScoutSubKit struct{}

func (k *ScoutSubKit) GetName() string       { return "Scout Sub Kit" }
func (k *ScoutSubKit) GetMaxStack() int      { return 1 }
func (k *ScoutSubKit) GetColor() color.Color { return color.RGBA{15, 160, 185, 255} }
func (k *ScoutSubKit) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if item.DrawItemIconSprite(screen, k.GetName(), cx, cy, size) {
		return
	}
	// Small sub capsule silhouette
	vector.FillRect(screen, cx-size/2.0, cy-size/4.0, size, size/2.0, k.GetColor(), false)
	vector.FillCircle(screen, cx+size/4.0, cy, size/4.0, color.RGBA{80, 205, 255, 255}, false)
}
func (k *ScoutSubKit) IsPlayerUpgrade() bool { return false }
func (k *ScoutSubKit) Deploy(x, y float64) Vehicle {
	return NewScoutSub(x, y)
}

