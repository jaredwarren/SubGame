package entity

import "github.com/jaredwarren/SubGame/internal/gvec"

// ShockKelpDef holds balance stats for ShockKelp flora.
type ShockKelpDef struct {
	FloorWidth          float64
	WallWidth           float64
	SwayPhaseSpeed      float64
	OrbChance           float64
	ShockCooldownFrames int
	PlayerDamage        float64
	VehicleDamage       float64
	KnockbackX          float64
	KnockbackY          float64
	WarningDuration     int
}

// ShatterBulbDef holds shared dims and pop effect for ShatterBulb flora.
type ShatterBulbDef struct {
	Dims           gvec.Vec2
	RestoreOxygen  float64
}

var ShockKelpArchetype = &ShockKelpDef{
	FloorWidth:          16,
	WallWidth:           28,
	SwayPhaseSpeed:      0.035,
	OrbChance:           0.5,
	ShockCooldownFrames: 80,
	PlayerDamage:        8,
	VehicleDamage:       12,
	KnockbackX:          4.5,
	KnockbackY:          -2.5,
	WarningDuration:     90,
}

var ShatterBulbArchetype = &ShatterBulbDef{
	Dims:          gvec.Vec2{X: 24, Y: 24},
	RestoreOxygen: 20,
}
