package vehicle

import "github.com/jaredwarren/SubGame/internal/gvec"

// VehicleDef is the shared balance envelope for all craft. Per-vehicle movement
// extras remain on SkiffDef / ScoutSubDef / HeavyMechDef; Controller reads the
// common fields through DefFor.
type VehicleDef struct {
	ID           VehicleID
	Name         string
	Dims         gvec.Vec2
	MaxHealth    float64
	MaxBattery   float64
	CargoSlots   int
	UpgradeSlots int
}

// DefFor returns the shared VehicleDef for an ID.
func DefFor(id VehicleID) *VehicleDef {
	switch id {
	case VehicleSkiff:
		d := SkiffArchetype
		return &VehicleDef{
			ID: VehicleSkiff, Name: "Skiff",
			Dims: d.Dims, MaxHealth: d.MaxHealth, MaxBattery: d.MaxBattery,
			CargoSlots: d.CargoSlots,
		}
	case VehicleScoutSub:
		d := ScoutSubArchetype
		return &VehicleDef{
			ID: VehicleScoutSub, Name: "Scout Sub",
			Dims: d.Dims, MaxHealth: d.MaxHealth, MaxBattery: d.MaxBattery,
			CargoSlots: d.CargoSlots, UpgradeSlots: d.UpgradeSlots,
		}
	case VehicleHeavyMech:
		d := HeavyMechArchetype
		return &VehicleDef{
			ID: VehicleHeavyMech, Name: "Heavy Mech",
			Dims: d.Dims, MaxHealth: d.MaxHealth, MaxBattery: d.MaxBattery,
			CargoSlots: d.CargoSlots, UpgradeSlots: d.UpgradeSlots,
		}
	default:
		return nil
	}
}

// Controller holds shared powered-craft state mutations used by every vehicle Update.
type Controller struct {
	Health, MaxHealth   float64
	Battery, MaxBattery float64
	Vel                 gvec.Vec2
}

// ShouldSkipPilotControl returns true when the craft is not piloted or the
// player is stunned. Callers should zero velocity when this returns true for
// the active craft (except mech gravity fall when inactive).
func ShouldSkipPilotControl(rt Runtime, self Vehicle) (skip, inactive bool) {
	if !rt.IsActiveVehicle(self) {
		return true, true
	}
	if rt.PlayerStunned() {
		return true, false
	}
	return false, false
}

// ApplyDamage reduces health, optionally scaled by damageReduction (1 = full damage).
func (c *Controller) ApplyDamage(amount, damageReduction float64) {
	if damageReduction <= 0 {
		damageReduction = 1
	}
	c.Health -= amount * damageReduction
	if c.Health < 0 {
		c.Health = 0
	}
}

// ApplyRepair increases health up to MaxHealth.
func (c *Controller) ApplyRepair(amount float64) {
	c.Health += amount
	if c.Health > c.MaxHealth {
		c.Health = c.MaxHealth
	}
}

// ApplyRecharge increases battery up to MaxBattery.
func (c *Controller) ApplyRecharge(amount float64) {
	c.Battery += amount
	if c.Battery > c.MaxBattery {
		c.Battery = c.MaxBattery
	}
}

// DrainBattery subtracts amount and clamps at zero. Returns whether power remains.
func (c *Controller) DrainBattery(amount float64) bool {
	c.Battery -= amount
	if c.Battery < 0 {
		c.Battery = 0
	}
	return c.Battery > 0
}

// HasPower reports whether the craft has battery remaining.
func (c *Controller) HasPower() bool {
	return c.Battery > 0
}
