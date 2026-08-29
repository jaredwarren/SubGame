package vehicle

import (
	"image"
	"image/color"
	_ "image/png"
	"math"

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
	d := ScoutSubArchetype
	upg := item.NewInventory(d.UpgradeSlots)
	upg.AddItem(&item.SonarAmplifier{}, 1)
	return &ScoutSub{
		Pos:        gvec.Vec2{X: x, Y: y},
		Dimensions: d.Dims,
		Health:     d.MaxHealth,
		MaxHealth:  d.MaxHealth,
		Battery:    d.MaxBattery,
		MaxBattery: d.MaxBattery,
		Cargo:      item.NewInventory(d.CargoSlots),
		Upgrades:   upg,
		Sonar: SonarSettings{
			BatteryCost: d.SonarBatteryCost,
			Pulse:       SonarPulse{DurationTicks: d.SonarDurationTicks, RadiusStep: d.SonarRadiusStep},
		},
	}
}

func (sub *ScoutSub) GetPos() gvec.Vec2            { return sub.Pos }
func (sub *ScoutSub) SetPos(pos gvec.Vec2)         { sub.Pos = pos }
func (sub *ScoutSub) GetDimensions() gvec.Vec2     { return sub.Dimensions }
func (sub *ScoutSub) GetHealth() float64           { return sub.Health }
func (sub *ScoutSub) GetMaxHealth() float64        { return sub.MaxHealth }
func (sub *ScoutSub) GetOxygen() float64           { return 100.0 }
func (sub *ScoutSub) GetDepthLimit() float64 {
	if sub.Upgrades != nil && item.HasItem[*item.ScoutSubDepthMK1](sub.Upgrades, 1) {
		return 120.0
	}
	return ScoutSubArchetype.DepthLimit
}
func (sub *ScoutSub) GetCargo() *item.Inventory    { return sub.Cargo }
func (sub *ScoutSub) GetUpgrades() *item.Inventory { return sub.Upgrades }
func (sub *ScoutSub) GetPerspective() string       { return "cave" }
func (sub *ScoutSub) GetName() string              { return "Scout Sub" }
func (sub *ScoutSub) GetID() VehicleID             { return VehicleScoutSub }
func (sub *ScoutSub) GetBattery() float64          { return sub.Battery }
func (sub *ScoutSub) GetMaxBattery() float64       { return sub.MaxBattery }
func (sub *ScoutSub) GetFacing() float64           { return sub.Facing }
func (sub *ScoutSub) SetFacing(facing float64)     { sub.Facing = facing }
func (sub *ScoutSub) ApplyForce(force gvec.Vec2) {
	sub.Vel = sub.Vel.Add(force)
}
func (sub *ScoutSub) GetKit() item.Item {
	return &ScoutSubKit{
		Upgrades: CloneInventory(sub.Upgrades),
		Health:   sub.Health,
		Battery:  sub.Battery,
		HasState: true,
	}
}

func (sub *ScoutSub) TakeDamage(amount float64) {
	SyncDamage(&sub.Health, &sub.MaxHealth, amount, 1)
}

func (sub *ScoutSub) Repair(amount float64) {
	SyncRepair(&sub.Health, &sub.MaxHealth, amount)
}

func (sub *ScoutSub) RechargeBattery(amount float64) {
	SyncRecharge(&sub.Battery, &sub.MaxBattery, amount)
}

func (sub *ScoutSub) hasUpgrade() bool {
	return item.HasItem[*item.SonarAmplifier](sub.Upgrades, 1)
}

func (sub *ScoutSub) Update(runtime Runtime) {
	d := ScoutSubArchetype
	if item.HasItem[*item.ThermalGenerator](sub.Upgrades, 1) {
		sub.Battery += d.ThermalRecharge
		ClampBattery(&sub.Battery, &sub.MaxBattery)
	}

	if skip, _ := ShouldSkipPilotControl(runtime, sub); skip {
		sub.Vel = gvec.Vec2{}
		return
	}

	input := runtime.Input()
	sub.Facing = CursorFacing(runtime)

	hasPower := sub.Battery > 0
	force, maxSpeed := ScaleForPower(d.Force, d.MaxSpeed, d.NoPowerForce, d.NoPowerMaxSpeed, hasPower)
	force, maxSpeed = ScaleForSlow(force, maxSpeed, runtime.PlayerSlowed())

	moving := ApplyWASDThrust(input, sub.Pos.Y, d.Waterline, &sub.Vel, force)
	if moving && hasPower {
		hasPower = DrainBatteryOnMove(&sub.Battery, moving, hasPower, d.BatteryDrain)
		MaybeEmitPropBubble(runtime, sub.Pos, sub.Dimensions, sub.Facing, 0.35)
	}

	ApplyDragClamp(&sub.Vel, d.Drag, maxSpeed)
	sub.Vel.Y = ApplyWaterline(sub.Pos.Y, sub.Vel.Y, runtime.TimeOfDay(), moving, false, WaterlineOpts{
		Waterline:   d.Waterline,
		SurfacePush: 0.15,
		BobSpring:   0.03,
	})

	sub.checkCollisions(runtime)
	sub.trySonar(runtime, input, hasPower)
	sub.tryCountermeasures(runtime, input, hasPower)
}

func (sub *ScoutSub) trySonar(runtime Runtime, input InputSource, hasPower bool) {
	if !hasPower || !input.IsKeyJustPressed(ebiten.KeyQ) {
		return
	}
	if !runtime.CanUseSonar() || sub.Battery < sub.Sonar.BatteryCost {
		return
	}
	sub.Battery -= sub.Sonar.BatteryCost
	ClampBattery(&sub.Battery, &sub.MaxBattery)

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

func (sub *ScoutSub) tryCountermeasures(runtime Runtime, input InputSource, hasPower bool) {
	if !hasPower || !input.IsKeyJustPressed(ebiten.KeySpace) {
		return
	}
	if item.HasItem[*item.DecoyLauncher](sub.Upgrades, 1) {
		TryLaunchDecoy(runtime, sub.Cargo, &sub.Battery, sub.Pos, sub.Dimensions, sub.Facing)
		return
	}
	if item.HasItem[*item.ChemicalDischarger](sub.Upgrades, 1) {
		TryLaunchDeterrent(runtime, sub.Cargo, &sub.Battery, sub.Pos, sub.Dimensions)
	}
}

func (sub *ScoutSub) checkCollisions(runtime Runtime) {
	onImpact := func(absSpeed float64) {
		CaveSpeedImpact(sub.TakeDamage, runtime, absSpeed, 2.0, 4.0, 2.0)
	}
	MoveAxisSeparated(&sub.Pos, &sub.Vel, sub.Dimensions, func(pos gvec.Vec2) bool {
		return sub.isSolid(runtime, pos)
	}, func() { onImpact(math.Abs(sub.Vel.X)) }, func() { onImpact(math.Abs(sub.Vel.Y)) })
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
type ScoutSubKit struct {
	Upgrades *item.Inventory
	Health   float64
	Battery  float64
	HasState bool
}

func (k *ScoutSubKit) GetID() item.ItemID    { return item.IDScoutSubKit }
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

func (k *ScoutSubKit) Clone() item.Item {
	return &ScoutSubKit{
		Upgrades: CloneInventory(k.Upgrades),
		Health:   k.Health,
		Battery:  k.Battery,
		HasState: k.HasState,
	}
}

func (k *ScoutSubKit) GetItemState() (*item.SavedInventory, float64, float64, bool) {
	if !k.HasState {
		return nil, 0, 0, false
	}
	var savedUpg *item.SavedInventory
	if k.Upgrades != nil {
		s := k.Upgrades.SerializeState()
		savedUpg = &s
	}
	return savedUpg, k.Health, k.Battery, true
}

func (k *ScoutSubKit) SetItemState(upgrades *item.SavedInventory, health float64, battery float64, hasState bool) {
	k.HasState = hasState
	k.Health = health
	k.Battery = battery
	if upgrades != nil {
		k.Upgrades = item.DeserializeInventory(*upgrades)
	} else {
		k.Upgrades = nil
	}
}

func (k *ScoutSubKit) Deploy(x, y float64) Vehicle {
	sub := NewScoutSub(x, y)
	RestoreKitState(&sub.Health, &sub.Battery, &sub.Upgrades, KitVehicleState{
		Upgrades: k.Upgrades,
		Health:   k.Health,
		Battery:  k.Battery,
		HasState: k.HasState,
	})
	return sub
}

