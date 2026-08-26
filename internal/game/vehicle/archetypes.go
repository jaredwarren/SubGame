package vehicle

import "github.com/jaredwarren/SubGame/internal/gvec"

// SurfaceSonarDef holds balance stats for the Skiff's overworld sonar upgrade.
type SurfaceSonarDef struct {
	BatteryCost        float64 // battery drained per pulse (25.0 = 25%)
	FogRevealRadius    int     // tile radius cleared of fog
	POIDetectionRadius int     // tile radius for detecting unvisited dive sites
	PulseDurationTicks int     // duration of animated sonar ring
	PulseRadiusStep    float64 // speed of expanding ring (px per tick)
	CooldownTicks      int     // cooldown ticks between scans
}

// SkiffDef holds balance stats for the surface Skiff.
type SkiffDef struct {
	Dims              gvec.Vec2
	MaxHealth         float64
	MaxBattery        float64
	CargoSlots        int
	UpgradeSlots      int
	SurfaceSonar      SurfaceSonarDef
	TurnSpeed         float64
	Accel             float64
	MaxSpeed          float64
	NoPowerAccel      float64
	NoPowerMaxSpeed   float64
	ReverseAccelScale float64
	Drag              float64
	BatteryDrain      float64
	WakeSpeedThresh   float64
}

// ScoutSubDef holds balance stats for the Scout Sub.
type ScoutSubDef struct {
	Dims            gvec.Vec2
	MaxHealth       float64
	MaxBattery      float64
	CargoSlots      int
	UpgradeSlots    int
	DepthLimit      float64
	Force           float64
	MaxSpeed        float64
	NoPowerForce    float64
	NoPowerMaxSpeed float64
	Drag            float64
	BatteryDrain    float64
	ThermalRecharge float64
	Waterline       float64
	SonarBatteryCost float64
	SonarDurationTicks int
	SonarRadiusStep float64
}

// HeavyMechDef holds balance stats for the Heavy Mech.
type HeavyMechDef struct {
	Dims              gvec.Vec2
	MaxHealth         float64
	MaxBattery        float64
	CargoSlots        int
	UpgradeSlots      int
	DepthLimit        float64
	DamageReduction   float64 // multiplier applied to incoming damage (0.6 = 40% reduction)
	Gravity           float64
	DragH             float64
	DragV             float64
	WalkForce         float64
	MaxSpeedH         float64
	NoPowerWalkForce  float64
	NoPowerMaxSpeedH  float64
	ThrustForce       float64
	ThrustDrain       float64
	WalkDrain         float64
	Waterline         float64
	SurfaceBuoyancy   float64
}

// SkiffArchetype is the shared balance table for Skiff instances.
var SkiffArchetype = &SkiffDef{
	Dims:              gvec.Vec2{X: 56, Y: 24},
	MaxHealth:         150.0,
	MaxBattery:        100.0,
	CargoSlots:        24,
	UpgradeSlots:      3,
	SurfaceSonar: SurfaceSonarDef{
		BatteryCost:        25.0,
		FogRevealRadius:    18,
		POIDetectionRadius: 35,
		PulseDurationTicks: 120,
		PulseRadiusStep:    6.5,
		CooldownTicks:      60,
	},
	TurnSpeed:         0.04,
	Accel:             0.20,
	MaxSpeed:          6.0,
	NoPowerAccel:      0.04,
	NoPowerMaxSpeed:   1.5,
	ReverseAccelScale: 0.4,
	Drag:              0.94,
	BatteryDrain:      0.02,
	WakeSpeedThresh:   0.4,
}

// ScoutSubArchetype is the shared balance table for ScoutSub instances.
var ScoutSubArchetype = &ScoutSubDef{
	Dims:               gvec.Vec2{X: 48, Y: 32},
	MaxHealth:          100.0,
	MaxBattery:         100.0,
	CargoSlots:         12,
	UpgradeSlots:       2,
	DepthLimit:         60.0,
	Force:              0.20,
	MaxSpeed:           4.5,
	NoPowerForce:       0.04,
	NoPowerMaxSpeed:    1.0,
	Drag:               0.94,
	BatteryDrain:       0.03,
	ThermalRecharge:    0.02,
	Waterline:          -8.0,
	SonarBatteryCost:   10.0,
	SonarDurationTicks: 180,
	SonarRadiusStep:    6.5,
}

// HeavyMechArchetype is the shared balance table for HeavyMech instances.
var HeavyMechArchetype = &HeavyMechDef{
	Dims:             gvec.Vec2{X: 48, Y: 48},
	MaxHealth:        200.0,
	MaxBattery:       100.0,
	CargoSlots:       8,
	UpgradeSlots:     2,
	DepthLimit:       120.0,
	DamageReduction:  0.6,
	Gravity:          0.12,
	DragH:            0.88,
	DragV:            0.95,
	WalkForce:        0.35,
	MaxSpeedH:        2.0,
	NoPowerWalkForce: 0.08,
	NoPowerMaxSpeedH: 0.6,
	ThrustForce:      0.28,
	ThrustDrain:      0.08,
	WalkDrain:        0.01,
	Waterline:        -12.0,
	SurfaceBuoyancy:  0.20,
}
