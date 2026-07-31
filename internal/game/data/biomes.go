package data

import "github.com/jaredwarren/SubGame/internal/game/cave"

// Cave biome specs and spawn IDs live in package cave (they reference resource.NodeType).
// Re-exported here for catalog browsing.

type (
	CaveBiomeSpec = cave.CaveBiomeSpec
	SpawnRules    = cave.SpawnRules
	FloraID       = cave.FloraID
	FaunaID       = cave.FaunaID
)

const (
	FloraKelp        = cave.FloraKelp
	FloraShockKelp   = cave.FloraShockKelp
	FloraShatterBulb = cave.FloraShatterBulb
	FloraCoral       = cave.FloraCoral

	FaunaPassiveFish = cave.FaunaPassiveFish
	FaunaPassiveCrab = cave.FaunaPassiveCrab
	FaunaSandViper   = cave.FaunaSandViper
)

var (
	DefaultSpawnRules       = cave.DefaultSpawnRules
	ShallowReefBiome        = cave.ShallowReefBiome
	KelpForestBiome         = cave.KelpForestBiome
	ThermalBarrensBiome     = cave.ThermalBarrensBiome
	AbyssalBlueBiome        = cave.AbyssalBlueBiome
	DefaultShallowReefBiome = cave.DefaultShallowReefBiome
)
