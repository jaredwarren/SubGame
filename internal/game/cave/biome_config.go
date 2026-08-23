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

// FloorStyle specifies the procedural texture and decorative style for the cave floor/rock.
type FloorStyle int

const (
	FloorStyleCoralSand   FloorStyle = iota // Sandy grains and ripples (Shallow Reef)
	FloorStyleMoss                          // Mossy patches, lichen spots, green fibrils (Kelp Forest)
	FloorStyleBasalt                        // Basalt rock, cracked volcanic crust, ember flecks (Thermal Barrens)
	FloorStyleAbyssalSilt                   // Dark abyssal silt, fine sediment, bioluminescent crystal flecks (Abyssal Trench)
)

// CaveBiomeSpec defines visual and spawn properties for cave views.
type CaveBiomeSpec struct {
	ID                 string
	Name               string
	FloorStyle         FloorStyle
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
		FloorStyle:         FloorStyleCoralSand,
		CaveRockColor:      color.RGBA{180, 155, 100, 255},
		CaveSandLightColor: color.RGBA{225, 205, 140, 255},
		CaveSandDarkColor:  color.RGBA{145, 120, 75, 255},
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
		FloorStyle:         FloorStyleMoss,
		CaveRockColor:      color.RGBA{65, 88, 55, 255},
		CaveSandLightColor: color.RGBA{115, 185, 80, 255}, // Lush light green moss
		CaveSandDarkColor:  color.RGBA{45, 80, 40, 255},   // Forest green moss clumps
		CaveStrokeColor:    color.RGBA{95, 130, 75, 255},
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
		FloorStyle:         FloorStyleBasalt,
		CaveRockColor:      color.RGBA{38, 32, 32, 255}, // Volcanic basalt stone
		CaveSandLightColor: color.RGBA{85, 68, 65, 255}, // Ash & basalt highlights
		CaveSandDarkColor:  color.RGBA{22, 18, 18, 255}, // Deep fracture charcoal
		CaveStrokeColor:    color.RGBA{58, 48, 48, 255},
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
			{Type: resource.NodeNickel, Weight: 30},
			{Type: resource.NodeCopper, Weight: 30},
		},
	}

	AbyssalBlueBiome = &CaveBiomeSpec{
		ID:                 "abyssal_blue",
		Name:               "Abyssal Trench",
		FloorStyle:         FloorStyleAbyssalSilt,
		CaveRockColor:      color.RGBA{42, 50, 72, 255}, // Deep abyssal slate
		CaveSandLightColor: color.RGBA{85, 110, 155, 255},
		CaveSandDarkColor:  color.RGBA{28, 34, 52, 255},
		CaveStrokeColor:    color.RGBA{65, 78, 108, 255},
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
			{Type: resource.NodeTitanium, Weight: 35},
			{Type: resource.NodeCopper, Weight: 20},
		},
	}

	// DefaultShallowReefBiome is the fallback biome when no overworld biome is available.
	DefaultShallowReefBiome = ShallowReefBiome
)
