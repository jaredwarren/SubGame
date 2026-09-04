package vehicle

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type SkiffWakeStyle int

const (
	WakeStyleVLines    SkiffWakeStyle = 0 // Option B: Continuous V-Wake Lines
	WakeStyleArcs      SkiffWakeStyle = 1 // Option A: Directional Wave Arcs
	WakeStyleVSegments SkiffWakeStyle = 2 // Option C: Dynamic V-Line Segments
)

const activeWakeStyle = WakeStyleVSegments

var (
	skiffSheet             *ebiten.Image
	skiffRadialLightImage  *ebiten.Image
	skiffHeadlightImage    *ebiten.Image
	skiffLightTexturesOnce sync.Once
)

type skiffWakePoint struct {
	x, y      float64
	facing    float64
	amplitude float64
	life      float64 // 1.0 -> 0.0
}

// Skiff is the starting surface boat — solar-powered, surface-only.
type Skiff struct {
	Pos           gvec.Vec2
	Vel           gvec.Vec2
	Dimensions    gvec.Vec2
	Facing        float64
	Health        float64
	MaxHealth     float64
	Battery       float64
	MaxBattery    float64
	Cargo         *item.Inventory
	Upgrades      *item.Inventory
	DockedBays    [3]*DockedVehicle // Bay 0 = Scout Sub, Bay 1 = Heavy Mech, Bay 2 = Mini-Lifepod
	wake          []skiffWakePoint
	spawnTimer    int
	lightMult     float64
	sonarCooldown int
	HeadlightsOn  bool
}

// NewSkiff creates a Skiff at the given world position.
func NewSkiff(x, y float64) *Skiff {
	d := SkiffArchetype
	return &Skiff{
		Pos:           gvec.Vec2{X: x, Y: y},
		Dimensions:    d.Dims,
		Health:        d.MaxHealth,
		MaxHealth:     d.MaxHealth,
		Battery:       d.MaxBattery,
		MaxBattery:    d.MaxBattery,
		Cargo:         item.NewInventory(d.CargoSlots),
		Upgrades:      item.NewInventory(d.UpgradeSlots),
		spawnTimer:    0,
		lightMult:     1.0,
		sonarCooldown: 0,
	}
}

func (s *Skiff) GetPos() gvec.Vec2            { return s.Pos }
func (s *Skiff) SetPos(pos gvec.Vec2)         { s.Pos = pos }
func (s *Skiff) GetDimensions() gvec.Vec2     { return s.Dimensions }
func (s *Skiff) GetHealth() float64           { return s.Health }
func (s *Skiff) GetMaxHealth() float64        { return s.MaxHealth }
func (s *Skiff) GetOxygen() float64           { return 100.0 }
func (s *Skiff) GetDepthLimit() float64       { return 0.0 }
func (s *Skiff) GetCargo() *item.Inventory    { return s.Cargo }
func (s *Skiff) GetUpgrades() *item.Inventory { return s.Upgrades }
func (s *Skiff) GetPerspective() string       { return "overworld" }
func (s *Skiff) GetName() string              { return "The Skiff" }
func (s *Skiff) GetID() VehicleID             { return VehicleSkiff }
func (s *Skiff) GetBattery() float64          { return s.Battery }
func (s *Skiff) GetMaxBattery() float64       { return s.MaxBattery }
func (s *Skiff) GetFacing() float64           { return s.Facing }
func (s *Skiff) SetFacing(facing float64)     { s.Facing = facing }
func (s *Skiff) ApplyForce(force gvec.Vec2) {
	s.Vel = s.Vel.Add(force)
}
func (s *Skiff) GetKit() item.Item {
	return &SkiffKit{
		Upgrades: CloneInventory(s.Upgrades),
		Health:   s.Health,
		Battery:  s.Battery,
		HasState: true,
	}
}

// Dock places a vehicle into its dedicated docking bay on the Skiff.
func (s *Skiff) Dock(v Vehicle) (int, bool) {
	if v == nil {
		return -1, false
	}
	bayIdx := s.FindBayForID(v.GetID())
	if bayIdx < 0 || bayIdx >= len(s.DockedBays) {
		// Fallback: search for first empty bay
		for i := 0; i < len(s.DockedBays); i++ {
			if s.DockedBays[i] == nil {
				bayIdx = i
				break
			}
		}
	}
	if bayIdx < 0 || bayIdx >= len(s.DockedBays) {
		return -1, false
	}
	s.DockedBays[bayIdx] = NewDockedVehicleFromVehicle(v)
	return bayIdx, true
}

// Undock removes a vehicle from the given docking bay and spawns it as an active Vehicle.
func (s *Skiff) Undock(bayIdx int, x, y float64) (Vehicle, bool) {
	if bayIdx < 0 || bayIdx >= len(s.DockedBays) || s.DockedBays[bayIdx] == nil {
		return nil, false
	}
	v := s.DockedBays[bayIdx].ToVehicle(x, y)
	s.DockedBays[bayIdx] = nil
	return v, true
}

// GetDocked returns the docked vehicle at the given bay index.
func (s *Skiff) GetDocked(bayIdx int) *DockedVehicle {
	if bayIdx < 0 || bayIdx >= len(s.DockedBays) {
		return nil
	}
	return s.DockedBays[bayIdx]
}

// SetDocked directly sets the docked vehicle at the given bay index.
func (s *Skiff) SetDocked(bayIdx int, dv *DockedVehicle) {
	if bayIdx >= 0 && bayIdx < len(s.DockedBays) {
		s.DockedBays[bayIdx] = dv
	}
}

// HasDocked reports whether the Skiff currently holds the specified vehicle ID.
func (s *Skiff) HasDocked(id VehicleID) bool {
	for _, bay := range s.DockedBays {
		if bay != nil && bay.ID == id {
			return true
		}
	}
	return false
}

// GetDockedByID returns the docked vehicle record and bay index for a vehicle ID.
func (s *Skiff) GetDockedByID(id VehicleID) (*DockedVehicle, int) {
	for i, bay := range s.DockedBays {
		if bay != nil && bay.ID == id {
			return bay, i
		}
	}
	return nil, -1
}

// FindBayForID maps a known VehicleID to its dedicated bay index (0 for Scout Sub, 1 for Heavy Mech, 2 for Mini-Lifepod).
func (s *Skiff) FindBayForID(id VehicleID) int {
	switch id {
	case VehicleScoutSub:
		return 0
	case VehicleHeavyMech:
		return 1
	case VehicleMiniLifepod:
		return 2
	default:
		return -1
	}
}

// SkiffKit represents the deployable kit for the Skiff.
type SkiffKit struct {
	Upgrades *item.Inventory
	Health   float64
	Battery  float64
	HasState bool
}

func (k *SkiffKit) GetID() item.ItemID    { return item.IDSkiffKit }
func (k *SkiffKit) GetName() string       { return "Skiff Kit" }
func (k *SkiffKit) GetMaxStack() int      { return 1 }
func (k *SkiffKit) GetColor() color.Color { return color.RGBA{235, 100, 30, 255} }
func (k *SkiffKit) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if item.DrawItemIconSprite(screen, k.GetName(), cx, cy, size) {
		return
	}
	// Vector fallback for Skiff Kit (small orange boat silhouette)
	vector.FillRect(screen, cx-size/2.0, cy-size/8.0, size, size/4.0, k.GetColor(), false)
	vector.FillCircle(screen, cx, cy-size/8.0, size/4.0, color.RGBA{40, 80, 110, 255}, false)
}
func (k *SkiffKit) IsPlayerUpgrade() bool { return false }

func (k *SkiffKit) Clone() item.Item {
	return &SkiffKit{
		Upgrades: CloneInventory(k.Upgrades),
		Health:   k.Health,
		Battery:  k.Battery,
		HasState: k.HasState,
	}
}

func (k *SkiffKit) GetItemState() (*item.SavedInventory, float64, float64, bool) {
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

func (k *SkiffKit) SetItemState(upgrades *item.SavedInventory, health float64, battery float64, hasState bool) {
	k.HasState = hasState
	k.Health = health
	k.Battery = battery
	if upgrades != nil {
		k.Upgrades = item.DeserializeInventory(*upgrades)
	} else {
		k.Upgrades = nil
	}
}

func (k *SkiffKit) Deploy(x, y float64) Vehicle {
	skiff := NewSkiff(x, y)
	RestoreKitState(&skiff.Health, &skiff.Battery, &skiff.Upgrades, KitVehicleState{
		Upgrades: k.Upgrades,
		Health:   k.Health,
		Battery:  k.Battery,
		HasState: k.HasState,
	})
	return skiff
}


func (s *Skiff) TakeDamage(amount float64) {
	SyncDamage(&s.Health, &s.MaxHealth, amount, 1)
}

func (s *Skiff) Repair(amount float64) {
	SyncRepair(&s.Health, &s.MaxHealth, amount)
}

func (s *Skiff) RechargeBattery(amount float64) {
	SyncRecharge(&s.Battery, &s.MaxBattery, amount)
}

const (
	// SkiffDockedChargeRatePerSecond is the maximum total battery (in units) transferred per second.
	SkiffDockedChargeRatePerSecond = 1.0
	// SkiffMinSafetyBatteryReserve is the minimum battery level the Skiff must retain (will not drain below this).
	SkiffMinSafetyBatteryReserve = 20.0
)

// UpdateDockedCharging transfers battery from the Skiff to docked submersibles.
func (s *Skiff) UpdateDockedCharging(dt float64) {
	if s.Battery <= SkiffMinSafetyBatteryReserve {
		return
	}

	var needingCharge []*DockedVehicle
	for _, dv := range s.DockedBays {
		if dv != nil && dv.MaxBattery > 0 && dv.Battery < dv.MaxBattery {
			needingCharge = append(needingCharge, dv)
		}
	}

	if len(needingCharge) == 0 {
		return
	}

	// Maximum battery that can be drained from the Skiff this step
	maxDrain := SkiffDockedChargeRatePerSecond * dt
	availablePower := s.Battery - SkiffMinSafetyBatteryReserve
	if maxDrain > availablePower {
		maxDrain = availablePower
	}
	if maxDrain <= 0 {
		return
	}

	// Distribute power equally across needing vehicles
	slicePerVehicle := maxDrain / float64(len(needingCharge))
	totalTransferred := 0.0

	for _, dv := range needingCharge {
		needed := dv.MaxBattery - dv.Battery
		transfer := min(slicePerVehicle, needed)
		dv.Battery += transfer
		totalTransferred += transfer
	}

	s.Battery -= totalTransferred
	if s.Battery < SkiffMinSafetyBatteryReserve {
		s.Battery = SkiffMinSafetyBatteryReserve
	}
}

// IsBayCharging reports whether the vehicle in the given docking bay is actively receiving charge from the Skiff.
func (s *Skiff) IsBayCharging(bayIdx int) bool {
	if bayIdx < 0 || bayIdx >= len(s.DockedBays) {
		return false
	}
	dv := s.DockedBays[bayIdx]
	return dv != nil && s.Battery > SkiffMinSafetyBatteryReserve && dv.MaxBattery > 0 && dv.Battery < dv.MaxBattery
}

func (s *Skiff) Update(runtime Runtime) {
	s.tickWake()
	s.lightMult = s.getLightMultiplier(runtime.TimeOfDay())

	if runtime.TimeOfDay() < 10800 {
		s.Battery += 0.05
		ClampBattery(&s.Battery, &s.MaxBattery)
	}

	s.UpdateDockedCharging(1.0 / 60.0)

	if skip, _ := ShouldSkipPilotControl(runtime, s); skip {
		s.Vel = gvec.Vec2{}
		return
	}

	d := SkiffArchetype
	input := runtime.Input()

	throttle := 0.0
	if desired, stickThrottle, analog := AnalogAimAxes(input); analog {
		// Point the stick = go that way. Steer toward stick heading, thrust by push amount.
		if stickThrottle > 0.01 {
			SteerToward(&s.Facing, desired, d.TurnSpeed*2.2)
			throttle = stickThrottle
		}
	} else {
		steer := 0.0
		if input.IsKeyPressed(ebiten.KeyA) || input.IsKeyPressed(ebiten.KeyArrowLeft) {
			steer -= 1
		}
		if input.IsKeyPressed(ebiten.KeyD) || input.IsKeyPressed(ebiten.KeyArrowRight) {
			steer += 1
		}
		if input.IsKeyPressed(ebiten.KeyW) || input.IsKeyPressed(ebiten.KeyArrowUp) {
			throttle += 1
		} else if input.IsKeyPressed(ebiten.KeyS) || input.IsKeyPressed(ebiten.KeyArrowDown) {
			throttle -= 1
		}
		turnScale := TurnScaleForSpeed(s.Vel.Length(), d.MaxSpeed, d.TurnIdleScale)
		s.Facing += d.TurnSpeed * steer * turnScale
	}

	hasPower := s.Battery > 0
	accel, maxSpeed := ScaleForPower(d.Accel, d.MaxSpeed, d.NoPowerAccel, d.NoPowerMaxSpeed, hasPower)

	moving := false
	fx := math.Cos(s.Facing)
	fy := math.Sin(s.Facing)
	if throttle > 0.01 {
		s.Vel.X += fx * accel * throttle
		s.Vel.Y += fy * accel * throttle
		moving = true
	} else if throttle < -0.01 {
		reverse := accel * d.ReverseAccelScale
		s.Vel.X += fx * reverse * throttle
		s.Vel.Y += fy * reverse * throttle
		moving = true
	}

	if moving && hasPower {
		hasPower = DrainBatteryOnMove(&s.Battery, moving, hasPower, d.BatteryDrain)
	}

	// Stronger grip while under power so the hull follows the bow instead of skating.
	keep := d.LateralKeep
	if moving && throttle > 0 {
		keep *= 0.55
	}
	ApplyLateralDrag(&s.Vel, s.Facing, keep)
	ApplyDragClamp(&s.Vel, d.Drag, maxSpeed)
	s.checkCollisions(runtime)
	s.maybeSpawnWake(moving, d)
	s.trySonar(runtime, input, hasPower)
	s.updateHeadlights()
}

func (s *Skiff) HasHeadlights() bool {
	return item.HasItem[*item.SkiffLight](s.Upgrades, 1)
}

func (s *Skiff) IsHeadlightsOn() bool {
	return s.HeadlightsOn && s.HasHeadlights() && s.Battery > 0
}

func (s *Skiff) ToggleHeadlights() bool {
	if !s.HasHeadlights() || s.Battery <= 0 {
		s.HeadlightsOn = false
		return false
	}
	s.HeadlightsOn = !s.HeadlightsOn
	return s.HeadlightsOn
}

func (s *Skiff) updateHeadlights() {
	if !s.HasHeadlights() {
		s.HeadlightsOn = false
		return
	}
	if s.HeadlightsOn {
		if s.Battery <= 0 {
			s.Battery = 0
			s.HeadlightsOn = false
			return
		}
		drain := SkiffArchetype.SkiffLight.BatteryDrain
		if drain <= 0 {
			drain = 0.02
		}
		s.Battery -= drain
		if s.Battery <= 0 {
			s.Battery = 0
			s.HeadlightsOn = false
		}
	}
}

func (s *Skiff) hasSurfaceSonar() bool {
	return item.HasItem[*item.SurfaceSonar](s.Upgrades, 1)
}

func (s *Skiff) trySonar(runtime Runtime, input InputSource, hasPower bool) {
	if s.sonarCooldown > 0 {
		s.sonarCooldown--
	}
	if !s.hasSurfaceSonar() {
		return
	}
	if !hasPower || !input.IsKeyJustPressed(ebiten.KeyQ) {
		return
	}
	d := SkiffArchetype.SurfaceSonar
	if s.sonarCooldown > 0 || s.Battery < d.BatteryCost {
		return
	}
	s.Battery -= d.BatteryCost
	ClampBattery(&s.Battery, &s.MaxBattery)
	s.sonarCooldown = d.CooldownTicks

	cx := s.Pos.X + s.Dimensions.X/2.0
	cy := s.Pos.Y + s.Dimensions.Y/2.0
	runtime.Emit(ActivateSurfaceSonarCmd{
		Source: gvec.Vec2{X: cx, Y: cy},
		Pulse: SonarPulse{
			DurationTicks: d.PulseDurationTicks,
			RadiusStep:    d.PulseRadiusStep,
		},
		FogRevealRadius:    d.FogRevealRadius,
		POIDetectionRadius: d.POIDetectionRadius,
	})
}

func (s *Skiff) tickWake() {
	var activeWake []skiffWakePoint
	decayRate := 0.025
	if activeWakeStyle == WakeStyleArcs || activeWakeStyle == WakeStyleVSegments {
		decayRate = 0.015
	}
	for i := range s.wake {
		s.wake[i].life -= decayRate
		if s.wake[i].life > 0 {
			activeWake = append(activeWake, s.wake[i])
		}
	}
	s.wake = activeWake
}

func (s *Skiff) maybeSpawnWake(moving bool, d *SkiffDef) {
	speed := s.Vel.Length()
	if moving && speed > d.WakeSpeedThresh {
		spawnInterval := 1
		if activeWakeStyle == WakeStyleArcs || activeWakeStyle == WakeStyleVSegments {
			spawnInterval = 4
		}
		s.spawnTimer++
		if s.spawnTimer >= spawnInterval {
			s.spawnTimer = 0

			cosF := math.Cos(s.Facing)
			sinF := math.Sin(s.Facing)
			cx := s.Pos.X + s.Dimensions.X/2.0
			cy := s.Pos.Y + s.Dimensions.Y/2.0

			// Center the spawn point exactly at the bow (front tip) of the skiff
			frontX := cx + cosF*28.0
			frontY := cy + sinF*28.0

			s.wake = append(s.wake, skiffWakePoint{
				x:         frontX,
				y:         frontY,
				facing:    s.Facing,
				amplitude: speed * 25.0, // speed determines intensity
				life:      1.0,
			})
		}
	}
}

func (s *Skiff) checkCollisions(runtime Runtime) {
	bPos, bSize := runtime.BaseStationPos()
	hasBase := bSize.X > 0 && bSize.Y > 0

	isSolid := func(pos gvec.Vec2) bool {
		if s.isSolid(runtime, pos) {
			return true
		}
		if hasBase {
			return pos.X < bPos.X+bSize.X && pos.X+s.Dimensions.X > bPos.X &&
				pos.Y < bPos.Y+bSize.Y && pos.Y+s.Dimensions.Y > bPos.Y
		}
		return false
	}

	MoveAxisSeparated(&s.Pos, &s.Vel, s.Dimensions, isSolid, nil, nil)
}

func (s *Skiff) isSolid(runtime Runtime, pos gvec.Vec2) bool {
	return solidAt(runtime.IsOverworldSolidAt, pos, s.Dimensions)
}

func (s *Skiff) Draw(screen *ebiten.Image, camX, camY float64) {
	if activeWakeStyle == WakeStyleVLines {
		// Draw continuous V-wake lines (Option B)
		if len(s.wake) >= 2 {
			const spreadAngle = 0.42
			for i := 0; i < len(s.wake)-1; i++ {
				p1 := s.wake[i]
				p2 := s.wake[i+1]

				// Calculate left/right coordinates for p1
				age1 := 1.0 - p1.life
				dist1 := age1 * 32.0

				lAngle1 := p1.facing + math.Pi - spreadAngle
				lX1 := p1.x + math.Cos(lAngle1)*dist1 - camX
				lY1 := p1.y + math.Sin(lAngle1)*dist1 - camY

				rAngle1 := p1.facing + math.Pi + spreadAngle
				rX1 := p1.x + math.Cos(rAngle1)*dist1 - camX
				rY1 := p1.y + math.Sin(rAngle1)*dist1 - camY

				// Calculate left/right coordinates for p2
				age2 := 1.0 - p2.life
				dist2 := age2 * 32.0

				lAngle2 := p2.facing + math.Pi - spreadAngle
				lX2 := p2.x + math.Cos(lAngle2)*dist2 - camX
				lY2 := p2.y + math.Sin(lAngle2)*dist2 - camY

				rAngle2 := p2.facing + math.Pi + spreadAngle
				rX2 := p2.x + math.Cos(rAngle2)*dist2 - camX
				rY2 := p2.y + math.Sin(rAngle2)*dist2 - camY

				avgLife := (p1.life + p2.life) * 0.5

				// Outer light blue-cyan halo
				clrOuter := s.applyLight(color.RGBA{160, 220, 255, uint8(avgLife * 80.0)})
				vector.StrokeLine(screen, float32(lX1), float32(lY1), float32(lX2), float32(lY2), 2.5, clrOuter, true)
				vector.StrokeLine(screen, float32(rX1), float32(rY1), float32(rX2), float32(rY2), 2.5, clrOuter, true)

				// Inner white core
				clrInner := s.applyLight(color.RGBA{255, 255, 255, uint8(avgLife * 160.0)})
				vector.StrokeLine(screen, float32(lX1), float32(lY1), float32(lX2), float32(lY2), 1.0, clrInner, true)
				vector.StrokeLine(screen, float32(rX1), float32(rY1), float32(rX2), float32(rY2), 1.0, clrInner, true)
			}
		}
	} else if activeWakeStyle == WakeStyleArcs {
		// Draw expanding backward-facing wave arcs (Option A)
		for _, p := range s.wake {
			alpha := p.life * p.amplitude
			if alpha > 255 {
				alpha = 255
			}
			if alpha < 0 {
				alpha = 0
			}

			// Radius expands based on age
			radius := 2.0 + (1.0-p.life)*80.0

			// Center angle is reverse of historical facing
			centerAngle := p.facing + math.Pi
			const halfSweep = 1.0 // ~114 degrees sweep arc

			// Outer light-blue-cyan halo
			clrOuter := s.applyLight(color.RGBA{160, 220, 255, uint8(alpha * 0.5)})
			drawArc(screen, p.x-camX, p.y-camY, radius, centerAngle, halfSweep, 2.5, clrOuter)

			// Inner white core
			clrInner := s.applyLight(color.RGBA{255, 255, 255, uint8(alpha)})
			drawArc(screen, p.x-camX, p.y-camY, radius, centerAngle, halfSweep, 1.0, clrInner)
		}
	} else if activeWakeStyle == WakeStyleVSegments {
		// Draw expanding backward-facing V-line segments (Option C: Combination)
		const spreadAngle = 0.65
		for _, p := range s.wake {
			// Opacity decays quadratically for a distinct, smooth fade-out
			maxAlpha := 100.0 + p.amplitude
			if maxAlpha > 230.0 {
				maxAlpha = 230.0
			}
			alpha := p.life * p.life * maxAlpha

			age := 1.0 - p.life
			dist := age * 80.0 // expands further outwards (up to 80 pixels)

			distStart := dist * 0.3
			distEnd := dist

			// Thickness also decays over time to simulate wave dissipation
			thickOuter := float32(2.5 * p.life)
			thickInner := float32(1.0 * p.life)
			if thickOuter < 0.1 {
				thickOuter = 0.1
			}
			if thickInner < 0.1 {
				thickInner = 0.1
			}

			// Left segment
			leftAngle := p.facing + math.Pi - spreadAngle
			lXStart := p.x + math.Cos(leftAngle)*distStart - camX
			lYStart := p.y + math.Sin(leftAngle)*distStart - camY
			lXEnd := p.x + math.Cos(leftAngle)*distEnd - camX
			lYEnd := p.y + math.Sin(leftAngle)*distEnd - camY

			// Right segment
			rightAngle := p.facing + math.Pi + spreadAngle
			rXStart := p.x + math.Cos(rightAngle)*distStart - camX
			rYStart := p.y + math.Sin(rightAngle)*distStart - camY
			rXEnd := p.x + math.Cos(rightAngle)*distEnd - camX
			rYEnd := p.y + math.Sin(rightAngle)*distEnd - camY

			// Outer light-blue-cyan halo
			clrOuter := s.applyLight(color.RGBA{160, 220, 255, uint8(alpha * 0.5)})
			vector.StrokeLine(screen, float32(lXStart), float32(lYStart), float32(lXEnd), float32(lYEnd), thickOuter, clrOuter, true)
			vector.StrokeLine(screen, float32(rXStart), float32(rYStart), float32(rXEnd), float32(rYEnd), thickOuter, clrOuter, true)

			// Inner white core
			clrInner := s.applyLight(color.RGBA{255, 255, 255, uint8(alpha)})
			vector.StrokeLine(screen, float32(lXStart), float32(lYStart), float32(lXEnd), float32(lYEnd), thickInner, clrInner, true)
			vector.StrokeLine(screen, float32(rXStart), float32(rYStart), float32(rXEnd), float32(rYEnd), thickInner, clrInner, true)
		}
	}

	// Draw headlights and radial ambient glow on water surface before drawing boat hull
	s.drawHeadlights(screen, camX, camY)

	if skiffSheet != nil {
		rect := image.Rect(348, 82, 676, 948)
		sprite := skiffSheet.SubImage(rect).(*ebiten.Image)

		op := &ebiten.DrawImageOptions{}

		// Center the cropped sprite on the origin (0, 0)
		op.GeoM.Translate(-164.0, -433.0)

		// Scale the cropped sprite to fit the target dimensions
		scaleX := s.Dimensions.Y / 328.0
		scaleY := s.Dimensions.X / 866.0
		op.GeoM.Scale(scaleX, scaleY)

		// Rotate so it aligns with s.Facing
		op.GeoM.Rotate(s.Facing + math.Pi/2.0)

		// Translate to screen coordinates, centered on the boat's collision box center
		cx := s.Pos.X + s.Dimensions.X/2.0 - camX
		cy := s.Pos.Y + s.Dimensions.Y/2.0 - camY
		op.GeoM.Translate(cx, cy)

		// Apply lighting scale
		mult := float32(s.lightMult)
		op.ColorScale.Scale(mult, mult, mult, 1.0)

		screen.DrawImage(sprite, op)
		s.drawHeadlightLenses(screen, camX, camY)
		return
	}

	// Fallback to original vector drawing code
	cosF := math.Cos(s.Facing)
	sinF := math.Sin(s.Facing)

	rotatePoint := func(px, py float64) (float32, float32) {
		rx := px*cosF - py*sinF
		ry := px*sinF + py*cosF
		return float32(s.Pos.X + s.Dimensions.X/2.0 + rx - camX), float32(s.Pos.Y + s.Dimensions.Y/2.0 + ry - camY)
	}

	x1, y1 := rotatePoint(28, 0)
	x2, y2 := rotatePoint(14, 12)
	x3, y3 := rotatePoint(-28, 12)
	x4, y4 := rotatePoint(-28, -12)
	x5, y5 := rotatePoint(14, -12)

	hullColor := color.RGBA{220, 230, 240, 255}
	stripeColor := color.RGBA{235, 100, 30, 255}

	drawFilledTriangle(screen, x1, y1, x2, y2, x3, y3, hullColor)
	drawFilledTriangle(screen, x1, y1, x3, y3, x4, y4, hullColor)
	drawFilledTriangle(screen, x1, y1, x4, y4, x5, y5, hullColor)

	vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x2, y2, x3, y3, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x3, y3, x4, y4, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x4, y4, x5, y5, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x5, y5, x1, y1, 1.5, stripeColor, false)

	cx1, cy1 := rotatePoint(10, 0)
	cx2, cy2 := rotatePoint(-4, 7)
	cx3, cy3 := rotatePoint(-16, 7)
	cx4, cy4 := rotatePoint(-16, -7)
	cx5, cy5 := rotatePoint(-4, -7)

	cabinColor := color.RGBA{40, 80, 110, 255}
	drawFilledTriangle(screen, cx1, cy1, cx2, cy2, cx3, cy3, cabinColor)
	drawFilledTriangle(screen, cx1, cy1, cx3, cy3, cx4, cy4, cabinColor)
	drawFilledTriangle(screen, cx1, cy1, cx4, cy4, cx5, cy5, cabinColor)

	sp1x, sp1y := rotatePoint(-10, -5)
	sp2x, sp2y := rotatePoint(-24, -5)
	sp3x, sp3y := rotatePoint(-24, 5)
	sp4x, sp4y := rotatePoint(-10, 5)

	solarColor := color.RGBA{15, 120, 215, 255}
	drawFilledTriangle(screen, sp1x, sp1y, sp2x, sp2y, sp3x, sp3y, solarColor)
	drawFilledTriangle(screen, sp1x, sp1y, sp3x, sp3y, sp4x, sp4y, solarColor)
	vector.StrokeLine(screen, sp1x, sp1y, sp2x, sp2y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp2x, sp2y, sp3x, sp3y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp3x, sp3y, sp4x, sp4y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp4x, sp4y, sp1x, sp1y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	s.drawHeadlightLenses(screen, camX, camY)
}

func ensureSkiffLightTextures() {
	skiffLightTexturesOnce.Do(func() {
		// 1. Generate 360-degree soft radial ambient light texture (128x128)
		const radSize = 128
		const radCenter = 64.0
		const radR = 62.0

		radImg := image.NewNRGBA(image.Rect(0, 0, radSize, radSize))
		for y := 0; y < radSize; y++ {
			for x := 0; x < radSize; x++ {
				dx := float64(x) - radCenter
				dy := float64(y) - radCenter
				dist := math.Hypot(dx, dy)
				if dist < radR {
					falloff := 1.0 - (dist / radR)
					// Smooth cubic decay
					intensity := falloff * falloff * (3.0 - 2.0*falloff)
					alpha := uint8(intensity * 48.0)
					radImg.SetNRGBA(x, y, color.NRGBA{
						R: 255,
						G: 232,
						B: 160,
						A: alpha,
					})
				}
			}
		}
		skiffRadialLightImage = ebiten.NewImageFromImage(radImg)

		// 2. Generate smooth forward headlight cone texture (320x400)
		// Origin at (0, 200) pointing right (+X)
		const coneW = 320
		const coneH = 400
		const originY = 200.0
		const maxDist = 290.0
		const halfAngle = 0.70 // ~40 deg half-angle (~80 deg total cone spread)

		coneImg := image.NewNRGBA(image.Rect(0, 0, coneW, coneH))
		for y := 0; y < coneH; y++ {
			for x := 0; x < coneW; x++ {
				dx := float64(x)
				dy := float64(y) - originY
				dist := math.Hypot(dx, dy)
				if dx > 0 && dist > 1.0 && dist < maxDist {
					angle := math.Atan2(dy, dx)
					absAngle := math.Abs(angle)
					if absAngle < halfAngle {
						// Distance falloff (quadratic)
						distNorm := dist / maxDist
						distFade := 1.0 - distNorm
						distFade = distFade * distFade

						// Angular edge feathering (smoothstep)
						angNorm := absAngle / halfAngle
						angFade := 1.0 - angNorm
						angFade = angFade * angFade * (3.0 - 2.0*angFade)

						// Soft dual beam emphasis for the two headlights
						beam1 := math.Exp(-math.Pow((angle-0.12)*7.0, 2.0))
						beam2 := math.Exp(-math.Pow((angle+0.12)*7.0, 2.0))
						coreGlow := (beam1 + beam2) * 0.35

						totalIntensity := distFade * (angFade*0.65 + coreGlow)
						if totalIntensity > 1.0 {
							totalIntensity = 1.0
						}

						alpha := uint8(totalIntensity * 60.0)
						coneImg.SetNRGBA(x, y, color.NRGBA{
							R: 255,
							G: 240,
							B: 185,
							A: alpha,
						})
					}
				}
			}
		}
		skiffHeadlightImage = ebiten.NewImageFromImage(coneImg)
	})
}

func (s *Skiff) drawHeadlights(screen *ebiten.Image, camX, camY float64) {
	if !s.IsHeadlightsOn() {
		return
	}

	ensureSkiffLightTextures()

	cx := s.Pos.X + s.Dimensions.X/2.0 - camX
	cy := s.Pos.Y + s.Dimensions.Y/2.0 - camY

	cosF := math.Cos(s.Facing)
	sinF := math.Sin(s.Facing)

	d := SkiffArchetype.SkiffLight
	radRadius := float32(d.RadialRadius)
	if radRadius <= 0 {
		radRadius = 65.0
	}

	// 1. Draw smooth forward headlight cone
	if skiffHeadlightImage != nil {
		bowX := cx + cosF*26.0
		bowY := cy + sinF*26.0

		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		op.GeoM.Translate(0, -200.0)
		op.GeoM.Rotate(s.Facing)
		op.GeoM.Translate(bowX, bowY)
		op.Blend = ebiten.BlendLighter
		screen.DrawImage(skiffHeadlightImage, op)
	}

	// 2. Draw smooth 360-degree radial ambient light around hull
	if skiffRadialLightImage != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		scale := float64(radRadius*2.0) / 128.0
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx-float64(radRadius), cy-float64(radRadius))
		op.Blend = ebiten.BlendLighter
		screen.DrawImage(skiffRadialLightImage, op)
	}
}

func (s *Skiff) drawHeadlightLenses(screen *ebiten.Image, camX, camY float64) {
	if !s.IsHeadlightsOn() {
		return
	}

	cx := s.Pos.X + s.Dimensions.X/2.0 - camX
	cy := s.Pos.Y + s.Dimensions.Y/2.0 - camY
	cosF := math.Cos(s.Facing)
	sinF := math.Sin(s.Facing)
	bowX := cx + cosF*26.0
	bowY := cy + sinF*26.0
	perpX := -sinF
	perpY := cosF

	for _, side := range []float64{-5.0, 5.0} {
		lx := float32(bowX + perpX*side)
		ly := float32(bowY + perpY*side)
		// Small soft glow halo
		vector.FillCircle(screen, lx, ly, 3.0, color.NRGBA{R: 255, G: 235, B: 140, A: 100}, true)
		// Small crisp bulb point
		vector.FillCircle(screen, lx, ly, 1.2, color.NRGBA{R: 255, G: 255, B: 240, A: 220}, true)
	}
}

func (s *Skiff) getLightMultiplier(timeOfDay float64) float64 {
	if timeOfDay >= 0 && timeOfDay < 1200 {
		return 0.2 + (timeOfDay/1200.0)*0.8
	}
	if timeOfDay >= 1200 && timeOfDay < 9600 {
		return 1.0
	}
	if timeOfDay >= 9600 && timeOfDay < 10800 {
		return 1.0 - ((timeOfDay-9600.0)/1200.0)*0.8
	}
	return 0.2
}

func (s *Skiff) applyLight(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * s.lightMult),
		G: uint8(float64(c.G) * s.lightMult),
		B: uint8(float64(c.B) * s.lightMult),
		A: c.A,
	}
}


