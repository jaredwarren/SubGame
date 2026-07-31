package data

import "github.com/jaredwarren/SubGame/internal/game/vehicle"

// Vehicle balance tables live in package vehicle (so recipes here can import vehicle
// without an import cycle). Re-exported for a single browseable catalog.

type (
	SkiffDef     = vehicle.SkiffDef
	ScoutSubDef  = vehicle.ScoutSubDef
	HeavyMechDef = vehicle.HeavyMechDef
)

var (
	SkiffArchetype     = vehicle.SkiffArchetype
	ScoutSubArchetype  = vehicle.ScoutSubArchetype
	HeavyMechArchetype = vehicle.HeavyMechArchetype
)
