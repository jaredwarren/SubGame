package vehicle

import (
	"image/color"
	_ "image/png"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// MiniLifepod is a craftable, portable field habitat deployable from the Skiff onto surface water.
type MiniLifepod struct {
	Pos               gvec.Vec2
	Dimensions        gvec.Vec2
	Facing            float64
	Health            float64
	MaxHealth         float64
	Battery           float64
	MaxBattery        float64
	Cargo             *item.Inventory
	Upgrades          *item.Inventory // 2 slots for active upgrades (Solar, etc.)
	SolarRechargeRate float64
	ActiveModules     map[item.BaseModule]bool
	Config            MiniLifepodDef
	AnimTimer         float64
}

// NewMiniLifepod creates a MiniLifepod at the given surface world position.
func NewMiniLifepod(x, y float64) *MiniLifepod {
	d := *MiniLifepodArchetype
	upgSlots := d.UpgradeSlots
	if upgSlots < 1 {
		upgSlots = 2
	}
	cargoSlots := d.CargoSlots

	pod := &MiniLifepod{
		Pos:               gvec.Vec2{X: x, Y: y},
		Dimensions:        d.Dims,
		Health:            d.MaxHealth,
		MaxHealth:         d.MaxHealth,
		Battery:           d.MaxBattery,
		MaxBattery:        d.MaxBattery,
		Cargo:             item.NewInventory(cargoSlots),
		Upgrades:          item.NewInventory(upgSlots),
		ActiveModules:     make(map[item.BaseModule]bool),
		Config:            d,
		SolarRechargeRate: d.SolarRechargeRate,
	}
	pod.RecalculateProperties()
	return pod
}

func (pod *MiniLifepod) GetPos() gvec.Vec2            { return pod.Pos }
func (pod *MiniLifepod) SetPos(pos gvec.Vec2)         { pod.Pos = pos }
func (pod *MiniLifepod) GetDimensions() gvec.Vec2     { return pod.Dimensions }
func (pod *MiniLifepod) GetHealth() float64           { return pod.Health }
func (pod *MiniLifepod) GetMaxHealth() float64        { return pod.MaxHealth }
func (pod *MiniLifepod) GetOxygen() float64           { return 100.0 }
func (pod *MiniLifepod) GetDepthLimit() float64       { return 0.0 } // Surface only
func (pod *MiniLifepod) GetCargo() *item.Inventory    { return pod.Cargo }
func (pod *MiniLifepod) GetUpgrades() *item.Inventory { return pod.Upgrades }
func (pod *MiniLifepod) GetPerspective() string       { return "overworld" }
func (pod *MiniLifepod) GetName() string              { return "Mini-Lifepod" }
func (pod *MiniLifepod) GetID() VehicleID             { return VehicleMiniLifepod }
func (pod *MiniLifepod) GetBattery() float64          { return pod.Battery }
func (pod *MiniLifepod) GetMaxBattery() float64       { return pod.MaxBattery }
func (pod *MiniLifepod) GetFacing() float64           { return pod.Facing }
func (pod *MiniLifepod) SetFacing(facing float64)     { pod.Facing = facing }
func (pod *MiniLifepod) ApplyForce(force gvec.Vec2)   {}

func (pod *MiniLifepod) GetKit() item.Item {
	return &MiniLifepodKit{
		Upgrades: CloneInventory(pod.Upgrades),
		Health:   pod.Health,
		Battery:  pod.Battery,
		HasState: true,
	}
}

func (pod *MiniLifepod) TakeDamage(amount float64) {
	SyncDamage(&pod.Health, &pod.MaxHealth, amount, 1)
}

func (pod *MiniLifepod) Repair(amount float64) {
	SyncRepair(&pod.Health, &pod.MaxHealth, amount)
}

func (pod *MiniLifepod) RechargeBattery(amount float64) {
	SyncRecharge(&pod.Battery, &pod.MaxBattery, amount)
}

// RecalculateProperties updates dynamic modules and solar recharge rate from installed upgrades.
func (pod *MiniLifepod) RecalculateProperties() {
	pod.SolarRechargeRate = pod.Config.SolarRechargeRate
	if pod.SolarRechargeRate <= 0 {
		pod.SolarRechargeRate = 0.01
	}

	pod.ActiveModules = map[item.BaseModule]bool{
		item.ModuleFabricator:  pod.Config.HasFabricator,
		item.ModuleMedical:     pod.Config.HasMedicalBay,
		item.ModuleSolar:       false,
		item.ModuleSolarMKII:   false,
		item.ModuleStorage:     pod.Config.HasStorage,
		item.ModuleStorageMKII: false,
	}

	if pod.Upgrades != nil {
		for _, slot := range pod.Upgrades.Slots {
			if slot.Item == nil {
				continue
			}
			if upg, ok := slot.Item.(item.UpgradeItem); ok {
				modType := upg.GetModuleType()
				pod.ActiveModules[modType] = true

				if recharge := upg.GetSolarRecharge(); recharge > pod.SolarRechargeRate {
					pod.SolarRechargeRate = recharge
				}
			}
		}
	}

	if pod.ActiveModules[item.ModuleSolarMKII] {
		pod.ActiveModules[item.ModuleSolar] = true
	}
	if pod.ActiveModules[item.ModuleStorageMKII] {
		pod.ActiveModules[item.ModuleStorage] = true
	}
}

// HasModule reports whether a specific base module is active on the MiniLifepod.
func (pod *MiniLifepod) HasModule(mod item.BaseModule) bool {
	if pod.ActiveModules == nil {
		return false
	}
	return pod.ActiveModules[mod]
}

// Update simulates passive solar recharging and buoyancy animation.
func (pod *MiniLifepod) Update(runtime Runtime) {
	pod.AnimTimer++
	rate := 0.005
	if runtime.TimeOfDay() < 10800 {
		rate = pod.SolarRechargeRate
	}
	pod.Battery += rate
	if pod.Battery > pod.MaxBattery {
		pod.Battery = pod.MaxBattery
	}
}

// Draw renders the Mini-Lifepod floating in the overworld with buoyancy and antenna beacon.
func (pod *MiniLifepod) Draw(screen *ebiten.Image, camX, camY float64) {
	sx := float32(pod.Pos.X - camX)
	sy := float32(pod.Pos.Y - camY)

	// Buoyant gentle bobbing
	bob := float32(math.Sin(pod.AnimTimer*0.05) * 1.5)
	sy += bob

	w := float32(pod.Dimensions.X)
	h := float32(pod.Dimensions.Y)
	cx := sx + w/2.0
	cy := sy + h/2.0

	// Water ripple / shadow ring
	vector.FillCircle(screen, cx, cy+4, w*0.48, color.RGBA{12, 35, 55, 120}, false)

	// Outer hull (clean high-tech pod: white capsule with dark blue & orange trim)
	vector.FillCircle(screen, cx, cy, w*0.42, color.RGBA{235, 240, 248, 255}, false)
	vector.StrokeCircle(screen, cx, cy, w*0.42, 2.0, color.RGBA{225, 110, 35, 255}, false)

	// Solar panel wings (if Solar active or base wings)
	hasSolar := pod.HasModule(item.ModuleSolar) || pod.HasModule(item.ModuleSolarMKII)
	panelColor := color.RGBA{28, 55, 90, 255}
	if hasSolar {
		panelColor = color.RGBA{35, 110, 180, 255}
	}
	vector.FillRect(screen, cx-w*0.52, cy-3, 5, 6, panelColor, false)
	vector.FillRect(screen, cx+w*0.52-5, cy-3, 5, 6, panelColor, false)

	// Central observation dome / airlock port
	vector.FillCircle(screen, cx, cy, w*0.22, color.RGBA{25, 75, 115, 255}, false)
	vector.StrokeCircle(screen, cx, cy, w*0.22, 1.2, color.RGBA{70, 180, 230, 255}, false)
	vector.FillCircle(screen, cx-2, cy-2, w*0.08, color.RGBA{170, 235, 255, 220}, false) // Window glint

	// Antenna stem & blinking beacon light
	vector.StrokeLine(screen, cx, cy-w*0.42, cx, cy-w*0.42-6, 1.5, color.RGBA{160, 175, 195, 255}, false)
	beaconAlpha := uint8(140 + int(math.Sin(pod.AnimTimer*0.1)*115))
	beaconColor := color.RGBA{245, 60, 60, beaconAlpha}
	vector.FillCircle(screen, cx, cy-w*0.42-7, 2.5, beaconColor, false)
}

// MiniLifepodKit represents the deployable kit for the portable Mini-Lifepod.
type MiniLifepodKit struct {
	Upgrades *item.Inventory
	Health   float64
	Battery  float64
	HasState bool
}

func (k *MiniLifepodKit) GetID() item.ItemID    { return item.IDMiniLifepodKit }
func (k *MiniLifepodKit) GetName() string       { return "Mini-Lifepod Kit" }
func (k *MiniLifepodKit) GetMaxStack() int      { return 1 }
func (k *MiniLifepodKit) GetColor() color.Color { return color.RGBA{230, 120, 40, 255} }

func (k *MiniLifepodKit) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if item.DrawItemIconSprite(screen, k.GetName(), cx, cy, size) {
		return
	}
	// Vector fallback for Mini-Lifepod Kit (white/orange habitat pod)
	r := size * 0.38
	vector.FillCircle(screen, cx, cy, r, color.RGBA{235, 240, 248, 255}, false)
	vector.StrokeCircle(screen, cx, cy, r, 1.8, color.RGBA{225, 110, 35, 255}, false)
	vector.FillCircle(screen, cx, cy, r*0.5, color.RGBA{30, 80, 125, 255}, false)
	vector.FillCircle(screen, cx, cy-r-2, 2.0, color.RGBA{240, 50, 50, 255}, false)
}

func (k *MiniLifepodKit) IsPlayerUpgrade() bool { return false }

func (k *MiniLifepodKit) Clone() item.Item {
	return &MiniLifepodKit{
		Upgrades: CloneInventory(k.Upgrades),
		Health:   k.Health,
		Battery:  k.Battery,
		HasState: k.HasState,
	}
}

func (k *MiniLifepodKit) GetItemState() (*item.SavedInventory, float64, float64, bool) {
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

func (k *MiniLifepodKit) SetItemState(upgrades *item.SavedInventory, health float64, battery float64, hasState bool) {
	k.HasState = hasState
	k.Health = health
	k.Battery = battery
	if upgrades != nil {
		k.Upgrades = item.DeserializeInventory(*upgrades)
	} else {
		k.Upgrades = nil
	}
}

func (k *MiniLifepodKit) Deploy(x, y float64) Vehicle {
	pod := NewMiniLifepod(x, y)
	RestoreKitState(&pod.Health, &pod.Battery, &pod.Upgrades, KitVehicleState{
		Upgrades: k.Upgrades,
		Health:   k.Health,
		Battery:  k.Battery,
		HasState: k.HasState,
	})
	pod.RecalculateProperties()
	return pod
}
