package vehicle

// VehicleID is a stable identity for vehicles in saves and quests.
type VehicleID string

const (
	VehicleSkiff     VehicleID = "skiff"
	VehicleScoutSub  VehicleID = "scout_sub"
	VehicleHeavyMech VehicleID = "heavy_mech"
)

// nameToVehicleID maps historical GetName() / save Type strings to VehicleID.
var nameToVehicleID = map[string]VehicleID{
	"skiff":              VehicleSkiff,
	"Skiff":              VehicleSkiff,
	"Skiff Boat":         VehicleSkiff,
	"Surface Skiff":      VehicleSkiff,
	"The Skiff":          VehicleSkiff,
	"scout_sub":          VehicleScoutSub,
	"ScoutSub":           VehicleScoutSub,
	"Scout Sub":          VehicleScoutSub,
	"Scout Submarine":    VehicleScoutSub,
	"heavy_mech":         VehicleHeavyMech,
	"HeavyMech":          VehicleHeavyMech,
	"Heavy Mech":         VehicleHeavyMech,
	"Heavy Mech Walker":  VehicleHeavyMech,
}

// VehicleIDFromName resolves a display/alias name to a VehicleID.
func VehicleIDFromName(name string) (VehicleID, bool) {
	id, ok := nameToVehicleID[name]
	return id, ok
}
