package data

import "github.com/jaredwarren/SubGame/internal/game/entity"

// Entity balance tables live in package entity (caves/spawn import entity).
// Re-exported here for a single browseable catalog.

type (
	FaunaDef       = entity.FaunaDef
	BehaviorID     = entity.BehaviorID
	ShockKelpDef   = entity.ShockKelpDef
	ShatterBulbDef = entity.ShatterBulbDef

	// Legacy fauna def aliases — all map to FaunaDef.
	SandViperDef         = entity.SandViperDef
	FalseBulbSnareDef    = entity.FalseBulbSnareDef
	ThermoclineRammerDef = entity.ThermoclineRammerDef
	BrimstoneSiphonDef   = entity.BrimstoneSiphonDef
	ElectroWeaverDef     = entity.ElectroWeaverDef
	VoltaicLurkerDef     = entity.VoltaicLurkerDef
	PassiveFishDef       = entity.PassiveFishDef
	PassiveCrabDef       = entity.PassiveCrabDef
)

var (
	FaunaDefFor = entity.FaunaDefFor

	SandViperArchetype         = entity.SandViperArchetype
	FalseBulbSnareArchetype    = entity.FalseBulbSnareArchetype
	ThermoclineRammerArchetype = entity.ThermoclineRammerArchetype
	BrimstoneSiphonArchetype   = entity.BrimstoneSiphonArchetype
	ShockKelpArchetype         = entity.ShockKelpArchetype
	ElectroWeaverArchetype     = entity.ElectroWeaverArchetype
	VoltaicLurkerArchetype     = entity.VoltaicLurkerArchetype
	PassiveFishArchetype       = entity.PassiveFishArchetype
	PassiveCrabArchetype       = entity.PassiveCrabArchetype
	ShatterBulbArchetype       = entity.ShatterBulbArchetype
)
