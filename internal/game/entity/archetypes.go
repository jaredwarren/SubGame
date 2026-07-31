package entity

import "github.com/jaredwarren/SubGame/internal/gvec"

// SandViperDef holds balance stats for SandViper.
type SandViperDef struct {
	Dims             gvec.Vec2
	PatrolSpeed      float64
	BobAmplitude     float64
	AggroRange       float64
	WindupFrames     int
	LungeSpeed       float64
	LungeFrames      int
	Damage           float64
	Knockback        float64
	PushBack         float64
	CooldownFrames   int
	CooldownDriftY   float64
	SwayPhaseSpeed   float64
	WarningDuration  int
}

// FalseBulbSnareDef holds balance stats for FalseBulbSnare.
type FalseBulbSnareDef struct {
	DecoyRange              float64
	LeashRange              float64
	ChaseRange              float64
	ChaseSpeed              float64
	Damage                  float64
	FlashlightConeHalfAngle float64
	SoundAlertRange         float64
	WarningDuration         int
	DecoyTargetSize         float64
}

// ThermoclineRammerDef holds balance stats for ThermoclineRammer.
type ThermoclineRammerDef struct {
	DecoyRange         float64
	AggroRange         float64
	SoundAlertRange    float64
	SprintVelThreshold float64
	ChargeSpeed        float64
	PatrolSpeedX       float64
	PatrolSpeedY       float64
	PatrolTurnInterval int
	ChargeMaxFrames    int
	ChargeMaxDist      float64
	StunFrames         int
	DeterrentSlowScale float64
	PlayerDamage       float64
	VehicleDamage      float64
	Knockback          float64
	PushBack           float64
	WarningDuration    int
	DecoyTargetSize    float64
}

// BrimstoneSiphonDef holds balance stats for BrimstoneSiphon.
type BrimstoneSiphonDef struct {
	CycleFrames      int
	ActiveStartFrame int
	JetRange         float64
	JetDrawLen       float64
	PlayerDPS        float64
	VehicleDPS       float64
}

// ShockKelpDef holds balance stats for ShockKelp.
type ShockKelpDef struct {
	FloorWidth          float64
	WallWidth           float64
	SwayPhaseSpeed      float64
	OrbChance           float64
	ShockCooldownFrames  int
	PlayerDamage        float64
	VehicleDamage       float64
	KnockbackX          float64
	KnockbackY          float64
	WarningDuration     int
}

// ElectroWeaverDef holds balance stats for ElectroWeaver.
type ElectroWeaverDef struct {
	DecoyRange           float64
	TrackRange           float64
	StrikeTimerFrames    int
	PlayerDamage         float64
	TeleportAwayDist     float64
	ApproachDist         float64
	ApproachSpeed        float64
	OrbitSpeedClose      float64
	IdleSpeed            float64
	TimerDecay           int
	MoveStartTimer       int
	AbyssalDepthTiles    float64
	WarningDuration      int
	DecoyWarningDuration int
}

// VoltaicLurkerDef holds balance stats for VoltaicLurker.
type VoltaicLurkerDef struct {
	Dims            gvec.Vec2
	SightRange      float64
	SightHalfWidth  float64
	LungeSpeed      float64
	MaxExtension    float64
	RetractSpeed    float64
	CooldownFrames   int
	Damage          float64
	HeadSize        float64
	StunDuration    int
	ShakeDuration   int
	ShakeIntensity  float64
	WarningDuration int
	SwayPhaseSpeed  float64
}

// PassiveFishDef holds balance stats for PassiveFish.
type PassiveFishDef struct {
	Dims           gvec.Vec2
	CatchRange     float64
	FleeRange      float64
	FleeSpeed      float64
	FleeFrames     int
	CruiseSpeed    float64
	SwimPhaseSpeed float64
}

// PassiveCrabDef holds balance stats for PassiveCrab.
type PassiveCrabDef struct {
	Dims                    gvec.Vec2
	CatchRange              float64
	ThreatRange             float64
	LightRange              float64
	FlashlightConeHalfAngle float64
	ShellFrames             int
	WalkTurnInterval        int
	WalkSpeed               float64
	Gravity                 float64
	MaxFallSpeed            float64
}

// ShatterBulbDef holds shared dims for ShatterBulb spawns.
type ShatterBulbDef struct {
	Dims gvec.Vec2
}

var SandViperArchetype = &SandViperDef{
	Dims:            gvec.Vec2{X: 24, Y: 12},
	PatrolSpeed:     0.5,
	BobAmplitude:    0.2,
	AggroRange:      100,
	WindupFrames:    30,
	LungeSpeed:      4.5,
	LungeFrames:     20,
	Damage:          10,
	Knockback:       3.5,
	PushBack:        15,
	CooldownFrames:  120,
	CooldownDriftY:  0.25,
	SwayPhaseSpeed:  0.08,
	WarningDuration: 90,
}

var FalseBulbSnareArchetype = &FalseBulbSnareDef{
	DecoyRange:              280,
	LeashRange:              360,
	ChaseRange:              180,
	ChaseSpeed:              3.5,
	Damage:                  20,
	FlashlightConeHalfAngle: 0.42,
	SoundAlertRange:         280,
	WarningDuration:         120,
	DecoyTargetSize:         16,
}

var ThermoclineRammerArchetype = &ThermoclineRammerDef{
	DecoyRange:         350,
	AggroRange:         250,
	SoundAlertRange:    250,
	SprintVelThreshold: 1.2,
	ChargeSpeed:        6.2,
	PatrolSpeedX:       0.8,
	PatrolSpeedY:       0.4,
	PatrolTurnInterval: 120,
	ChargeMaxFrames:    90,
	ChargeMaxDist:      350,
	StunFrames:         180,
	DeterrentSlowScale: 0.5,
	PlayerDamage:       25,
	VehicleDamage:      30,
	Knockback:          6.5,
	PushBack:           40,
	WarningDuration:    120,
	DecoyTargetSize:    16,
}

var BrimstoneSiphonArchetype = &BrimstoneSiphonDef{
	CycleFrames:      120,
	ActiveStartFrame: 60,
	JetRange:         160,
	JetDrawLen:       120,
	PlayerDPS:        0.6,
	VehicleDPS:       0.4,
}

var ShockKelpArchetype = &ShockKelpDef{
	FloorWidth:         16,
	WallWidth:          28,
	SwayPhaseSpeed:     0.035,
	OrbChance:          0.5,
	ShockCooldownFrames: 80,
	PlayerDamage:       8,
	VehicleDamage:      12,
	KnockbackX:         4.5,
	KnockbackY:         -2.5,
	WarningDuration:    90,
}

var ElectroWeaverArchetype = &ElectroWeaverDef{
	DecoyRange:           500,
	TrackRange:           500,
	StrikeTimerFrames:    300,
	PlayerDamage:         45,
	TeleportAwayDist:     350,
	ApproachDist:         100,
	ApproachSpeed:        1.5,
	OrbitSpeedClose:      1.2,
	IdleSpeed:            0.8,
	TimerDecay:           2,
	MoveStartTimer:       60,
	AbyssalDepthTiles:    80,
	WarningDuration:      180,
	DecoyWarningDuration: 120,
}

var VoltaicLurkerArchetype = &VoltaicLurkerDef{
	Dims:            gvec.Vec2{X: 64, Y: 64},
	SightRange:      130,
	SightHalfWidth:  12,
	LungeSpeed:      6,
	MaxExtension:    80,
	RetractSpeed:    3,
	CooldownFrames:  480,
	Damage:          15,
	HeadSize:        16,
	StunDuration:    90,
	ShakeDuration:   20,
	ShakeIntensity:  4,
	WarningDuration: 90,
	SwayPhaseSpeed:  0.05,
}

var PassiveFishArchetype = &PassiveFishDef{
	Dims:           gvec.Vec2{X: 20, Y: 12},
	CatchRange:     80,
	FleeRange:      120,
	FleeSpeed:      3.5,
	FleeFrames:     60,
	CruiseSpeed:    0.6,
	SwimPhaseSpeed: 0.04,
}

var PassiveCrabArchetype = &PassiveCrabDef{
	Dims:                    gvec.Vec2{X: 16, Y: 10},
	CatchRange:              64,
	ThreatRange:             100,
	LightRange:              300,
	FlashlightConeHalfAngle: 0.42,
	ShellFrames:             90,
	WalkTurnInterval:        180,
	WalkSpeed:               0.35,
	Gravity:                 0.3,
	MaxFallSpeed:            4.0,
}

var ShatterBulbArchetype = &ShatterBulbDef{
	Dims: gvec.Vec2{X: 24, Y: 24},
}
