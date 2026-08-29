package entity

import "github.com/jaredwarren/SubGame/internal/gvec"

// BehaviorID selects the Go Update/Draw implementation for a fauna type.
type BehaviorID int

const (
	BehaviorPassiveFish BehaviorID = iota
	BehaviorPassiveCrab
	BehaviorSandViper
	BehaviorFalseBulbSnare
	BehaviorThermoclineRammer
	BehaviorElectroWeaver
	BehaviorVoltaicLurker
	BehaviorBrimstoneSiphon
	BehaviorInkSquid
	BehaviorLanternfish
	BehaviorGlowSquid
)

// FaunaDef holds balance stats for one fauna type. Unused fields are zero for types
// that do not read them; behaviors stay in Go, keyed by BehaviorID.
type FaunaDef struct {
	ID       FaunaID
	Behavior BehaviorID
	Dims     gvec.Vec2

	Damage          float64
	PlayerDamage    float64
	VehicleDamage   float64
	PlayerDPS       float64
	VehicleDPS      float64
	Knockback       float64
	KnockbackX      float64
	KnockbackY      float64
	PushBack        float64
	WarningDuration int
	DecoyWarningDuration int

	AggroRange              float64
	DecoyRange              float64
	DecoyTargetSize         float64
	LeashRange              float64
	ChaseRange              float64
	TrackRange              float64
	SoundAlertRange         float64
	FlashlightConeHalfAngle float64
	SightRange              float64
	SightHalfWidth          float64
	CatchRange              float64
	FleeRange               float64
	ThreatRange             float64
	LightRange              float64

	PatrolSpeed         float64
	PatrolSpeedX        float64
	PatrolSpeedY        float64
	BobAmplitude        float64
	ChaseSpeed          float64
	ChargeSpeed         float64
	LungeSpeed          float64
	CruiseSpeed         float64
	FleeSpeed           float64
	ApproachSpeed       float64
	OrbitSpeedClose     float64
	IdleSpeed           float64
	RetractSpeed        float64
	WalkSpeed           float64
	SprintVelThreshold  float64
	SwimPhaseSpeed      float64
	SwayPhaseSpeed      float64
	Gravity             float64
	MaxFallSpeed        float64

	WindupFrames       int
	LungeFrames        int
	CooldownFrames     int
	ChargeMaxFrames    int
	StunFrames         int
	StunDuration       int
	FleeFrames         int
	ShellFrames        int
	WalkTurnInterval   int
	PatrolTurnInterval int
	CycleFrames        int
	ActiveStartFrame   int
	StrikeTimerFrames  int
	TimerDecay         int
	MoveStartTimer     int
	ShakeDuration      int

	ChargeMaxDist      float64
	MaxExtension       float64
	TeleportAwayDist   float64
	ApproachDist       float64
	AbyssalDepthTiles  float64
	JetRange           float64
	JetDrawLen         float64
	HeadSize           float64
	CooldownDriftY     float64
	DeterrentSlowScale float64
	ShakeIntensity     float64
}

var faunaRegistry = map[FaunaID]*FaunaDef{
	FaunaPassiveFish: {
		ID: FaunaPassiveFish, Behavior: BehaviorPassiveFish,
		Dims: gvec.Vec2{X: 20, Y: 12},
		CatchRange: 80, FleeRange: 120, FleeSpeed: 3.5, FleeFrames: 60,
		CruiseSpeed: 0.6, SwimPhaseSpeed: 0.04,
	},
	FaunaPassiveCrab: {
		ID: FaunaPassiveCrab, Behavior: BehaviorPassiveCrab,
		Dims: gvec.Vec2{X: 16, Y: 10},
		CatchRange: 64, ThreatRange: 100, LightRange: 300,
		FlashlightConeHalfAngle: 0.42, ShellFrames: 90,
		WalkTurnInterval: 180, WalkSpeed: 0.35, Gravity: 0.3, MaxFallSpeed: 4.0,
	},
	FaunaSandViper: {
		ID: FaunaSandViper, Behavior: BehaviorSandViper,
		Dims: gvec.Vec2{X: 24, Y: 12},
		PatrolSpeed: 0.5, BobAmplitude: 0.2, AggroRange: 100,
		WindupFrames: 30, LungeSpeed: 4.5, LungeFrames: 20, Damage: 10,
		Knockback: 3.5, PushBack: 15, CooldownFrames: 120, CooldownDriftY: 0.25,
		SwayPhaseSpeed: 0.08, WarningDuration: 90,
	},
	FaunaFalseBulbSnare: {
		ID: FaunaFalseBulbSnare, Behavior: BehaviorFalseBulbSnare,
		Dims: gvec.Vec2{X: 24, Y: 32},
		DecoyRange: 280, LeashRange: 360, ChaseRange: 180, ChaseSpeed: 3.5,
		Damage: 20, FlashlightConeHalfAngle: 0.42, SoundAlertRange: 280,
		WarningDuration: 120, DecoyTargetSize: 16,
	},
	FaunaThermoclineRammer: {
		ID: FaunaThermoclineRammer, Behavior: BehaviorThermoclineRammer,
		Dims: gvec.Vec2{X: 36, Y: 24},
		DecoyRange: 350, AggroRange: 250, SoundAlertRange: 250,
		SprintVelThreshold: 1.2, ChargeSpeed: 6.2, PatrolSpeedX: 0.8, PatrolSpeedY: 0.4,
		PatrolTurnInterval: 120, ChargeMaxFrames: 90, ChargeMaxDist: 350, StunFrames: 180,
		DeterrentSlowScale: 0.5, PlayerDamage: 25, VehicleDamage: 30,
		Knockback: 6.5, PushBack: 40, WarningDuration: 120, DecoyTargetSize: 16,
	},
	FaunaBrimstoneSiphon: {
		ID: FaunaBrimstoneSiphon, Behavior: BehaviorBrimstoneSiphon,
		Dims: gvec.Vec2{X: 32, Y: 32},
		CycleFrames: 120, ActiveStartFrame: 60, JetRange: 160, JetDrawLen: 120,
		PlayerDPS: 0.6, VehicleDPS: 0.4,
	},
	FaunaElectroWeaver: {
		ID: FaunaElectroWeaver, Behavior: BehaviorElectroWeaver,
		Dims: gvec.Vec2{X: 40, Y: 20},
		DecoyRange: 500, TrackRange: 500, StrikeTimerFrames: 300, PlayerDamage: 45,
		LungeSpeed: 8.5, CooldownFrames: 180, ApproachDist: 100, ApproachSpeed: 1.5,
		OrbitSpeedClose: 1.2, IdleSpeed: 0.8, TimerDecay: 2, MoveStartTimer: 60,
		AbyssalDepthTiles: 80, WarningDuration: 180, DecoyWarningDuration: 120,
	},
	FaunaVoltaicLurker: {
		ID: FaunaVoltaicLurker, Behavior: BehaviorVoltaicLurker,
		Dims: gvec.Vec2{X: 64, Y: 64},
		SightRange: 130, SightHalfWidth: 12, LungeSpeed: 6, MaxExtension: 80,
		RetractSpeed: 3, CooldownFrames: 480, Damage: 15, HeadSize: 16,
		StunDuration: 90, ShakeDuration: 20, ShakeIntensity: 4,
		WarningDuration: 90, SwayPhaseSpeed: 0.05,
	},
	FaunaInkSquid: {
		ID: FaunaInkSquid, Behavior: BehaviorInkSquid,
		Dims: gvec.Vec2{X: 22, Y: 16},
		PatrolSpeed: 0.6, FleeSpeed: 3.8, FleeFrames: 60,
		ThreatRange: 70.0, CooldownFrames: 480, SwimPhaseSpeed: 0.04,
	},
	FaunaLanternfish: {
		ID: FaunaLanternfish, Behavior: BehaviorLanternfish,
		Dims: gvec.Vec2{X: 18, Y: 12},
		CatchRange: 80, FleeRange: 100, FleeSpeed: 2.4, FleeFrames: 50,
		CruiseSpeed: 0.7, SwimPhaseSpeed: 0.05,
	},
	FaunaGlowSquid: {
		ID: FaunaGlowSquid, Behavior: BehaviorGlowSquid,
		Dims: gvec.Vec2{X: 24, Y: 18},
		PatrolSpeed: 0.65, FleeSpeed: 4.2, FleeFrames: 70,
		ThreatRange: 75.0, CooldownFrames: 420, SwimPhaseSpeed: 0.045,
	},
}

// Legacy per-type aliases — all fauna balance rows are FaunaDef entries.
type (
	SandViperDef         = FaunaDef
	FalseBulbSnareDef    = FaunaDef
	ThermoclineRammerDef = FaunaDef
	BrimstoneSiphonDef   = FaunaDef
	ElectroWeaverDef     = FaunaDef
	VoltaicLurkerDef     = FaunaDef
	PassiveFishDef       = FaunaDef
	PassiveCrabDef       = FaunaDef
	InkSquidDef          = FaunaDef
	LanternfishDef       = FaunaDef
	GlowSquidDef         = FaunaDef
)

// Legacy archetype pointers alias faunaRegistry rows.
var (
	SandViperArchetype         = faunaRegistry[FaunaSandViper]
	FalseBulbSnareArchetype    = faunaRegistry[FaunaFalseBulbSnare]
	ThermoclineRammerArchetype = faunaRegistry[FaunaThermoclineRammer]
	BrimstoneSiphonArchetype   = faunaRegistry[FaunaBrimstoneSiphon]
	ElectroWeaverArchetype     = faunaRegistry[FaunaElectroWeaver]
	VoltaicLurkerArchetype     = faunaRegistry[FaunaVoltaicLurker]
	PassiveFishArchetype       = faunaRegistry[FaunaPassiveFish]
	PassiveCrabArchetype       = faunaRegistry[FaunaPassiveCrab]
	InkSquidArchetype          = faunaRegistry[FaunaInkSquid]
	LanternfishArchetype       = faunaRegistry[FaunaLanternfish]
	GlowSquidArchetype         = faunaRegistry[FaunaGlowSquid]
)

// FaunaDefFor returns the balance row for id, or nil.
func FaunaDefFor(id FaunaID) *FaunaDef {
	return faunaRegistry[id]
}
