package cave

import (
	"image/color"

	"github.com/jaredwarren/SubGame/internal/game/resource"
)

// SpawnRules controls entity spawn densities for cave generation.
type SpawnRules struct {
	ShatterBulbChance   float64
	OpenWaterFishChance float64
	FaunaChance         float64
	FloraChance         float64
	CoralChance         float64
}

// DefaultSpawnRules matches historical shallow seabed spawn densities.
var DefaultSpawnRules = SpawnRules{
	ShatterBulbChance:   0.08,
	OpenWaterFishChance: 0.012,
	FaunaChance:         0.03,
	FloraChance:         0.28,
	CoralChance:         0.10,
}

// CaveBiomeSpec defines visual and spawn properties for cave views.
type CaveBiomeSpec struct {
	ID                 string
	Name               string
	CaveRockColor      color.RGBA
	CaveSandLightColor color.RGBA
	CaveSandDarkColor  color.RGBA
	CaveStrokeColor    color.RGBA
	CaveAmbientTint    color.RGBA
	Rules              SpawnRules
	FloraSpawns        []SpawnEntry[FloraID]
	FaunaSpawns        []SpawnEntry[FaunaID]
	MineralSpawns      []SpawnEntry[resource.NodeType]
}

// SpawnRulesOrDefault returns biome spawn rules, or DefaultSpawnRules if the spec is nil.
func (s *CaveBiomeSpec) SpawnRulesOrDefault() SpawnRules {
	if s == nil {
		return DefaultSpawnRules
	}
	return s.Rules
}

var (
	ShallowReefBiome = &CaveBiomeSpec{
		ID:                 "shallow_reef",
		Name:               "Shallow Coral Reef",
		CaveRockColor:      color.RGBA{180, 155, 100, 255},
		CaveSandLightColor: color.RGBA{215, 190, 125, 255},
		CaveSandDarkColor:  color.RGBA{150, 130, 80, 255},
		CaveStrokeColor:    color.RGBA{210, 185, 120, 255},
		CaveAmbientTint:    color.RGBA{10, 50, 110, 255},
		Rules:              DefaultSpawnRules,
		FloraSpawns: []SpawnEntry[FloraID]{
			{Type: FloraCoral, Weight: 50},
			{Type: FloraKelp, Weight: 40},
			{Type: FloraShockKelp, Weight: 10},
		},
		FaunaSpawns: []SpawnEntry[FaunaID]{
			{Type: FaunaPassiveFish, Weight: 60},
			{Type: FaunaPassiveCrab, Weight: 30},
			{Type: FaunaSandViper, Weight: 10},
		},
		MineralSpawns: []SpawnEntry[resource.NodeType]{
			{Type: resource.NodeTitanium, Weight: 50},
			{Type: resource.NodeCopper, Weight: 40},
			{Type: resource.NodeQuartz, Weight: 10},
		},
	}

	KelpForestBiome = &CaveBiomeSpec{
		ID:                 "kelp_forest",
		Name:               "Kelp Forest",
		CaveRockColor:      color.RGBA{120, 145, 95, 255},
		CaveSandLightColor: color.RGBA{160, 185, 120, 255},
		CaveSandDarkColor:  color.RGBA{90, 115, 70, 255},
		CaveStrokeColor:    color.RGBA{140, 165, 110, 255},
		CaveAmbientTint:    color.RGBA{10, 70, 60, 255},
		Rules:              DefaultSpawnRules,
		FloraSpawns: []SpawnEntry[FloraID]{
			{Type: FloraKelp, Weight: 60},
			{Type: FloraShockKelp, Weight: 30},
			{Type: FloraCoral, Weight: 10},
		},
		FaunaSpawns: []SpawnEntry[FaunaID]{
			{Type: FaunaPassiveFish, Weight: 45},
			{Type: FaunaPassiveCrab, Weight: 45},
			{Type: FaunaSandViper, Weight: 10},
		},
		MineralSpawns: []SpawnEntry[resource.NodeType]{
			{Type: resource.NodeQuartz, Weight: 50},
			{Type: resource.NodeCopper, Weight: 35},
			{Type: resource.NodeTitanium, Weight: 15},
		},
	}

	ThermalBarrensBiome = &CaveBiomeSpec{
		ID:                 "thermal_barrens",
		Name:               "Thermal Barrens",
		CaveRockColor:      color.RGBA{110, 85, 75, 255},
		CaveSandLightColor: color.RGBA{150, 115, 95, 255},
		CaveSandDarkColor:  color.RGBA{80, 60, 50, 255},
		CaveStrokeColor:    color.RGBA{130, 95, 80, 255},
		CaveAmbientTint:    color.RGBA{80, 30, 30, 255},
		Rules:              DefaultSpawnRules,
		FloraSpawns: []SpawnEntry[FloraID]{
			{Type: FloraShatterBulb, Weight: 50},
			{Type: FloraShockKelp, Weight: 35},
			{Type: FloraCoral, Weight: 15},
		},
		FaunaSpawns: []SpawnEntry[FaunaID]{
			{Type: FaunaSandViper, Weight: 50},
			{Type: FaunaPassiveCrab, Weight: 30},
			{Type: FaunaPassiveFish, Weight: 20},
		},
		MineralSpawns: []SpawnEntry[resource.NodeType]{
			{Type: resource.NodeQuartz, Weight: 40},
			{Type: resource.NodeAbyssalOre, Weight: 30},
			{Type: resource.NodeCopper, Weight: 30},
		},
	}

	AbyssalBlueBiome = &CaveBiomeSpec{
		ID:                 "abyssal_blue",
		Name:               "Abyssal Trench",
		CaveRockColor:      color.RGBA{80, 90, 115, 255},
		CaveSandLightColor: color.RGBA{110, 125, 155, 255},
		CaveSandDarkColor:  color.RGBA{55, 65, 85, 255},
		CaveStrokeColor:    color.RGBA{95, 110, 135, 255},
		CaveAmbientTint:    color.RGBA{15, 20, 60, 255},
		Rules:              DefaultSpawnRules,
		FloraSpawns: []SpawnEntry[FloraID]{
			{Type: FloraShatterBulb, Weight: 60},
			{Type: FloraShockKelp, Weight: 30},
			{Type: FloraCoral, Weight: 10},
		},
		FaunaSpawns: []SpawnEntry[FaunaID]{
			{Type: FaunaSandViper, Weight: 55},
			{Type: FaunaPassiveFish, Weight: 30},
			{Type: FaunaPassiveCrab, Weight: 15},
		},
		MineralSpawns: []SpawnEntry[resource.NodeType]{
			{Type: resource.NodeNickel, Weight: 45},
			{Type: resource.NodeAbyssalOre, Weight: 35},
			{Type: resource.NodeTitanium, Weight: 20},
		},
	}

	// DefaultShallowReefBiome is the fallback biome when no overworld biome is available.
	DefaultShallowReefBiome = ShallowReefBiome
)
