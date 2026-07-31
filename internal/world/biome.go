package world

import (
	"github.com/jaredwarren/SubGame/internal/game/cave"
)

// BiomeID identifies a specific biome.
type BiomeID string

const (
	BiomeShallowReef    BiomeID = "shallow_reef"
	BiomeKelpForest     BiomeID = "kelp_forest"
	BiomeThermalBarrens BiomeID = "thermal_barrens"
	BiomeAbyssalBlue    BiomeID = "abyssal_blue"
)

// ColorOffset represents signed RGB offset values for water tinting.
type ColorOffset struct {
	R, G, B float64
}

// BiomeSpec defines data-driven visual and spawning configuration for a biome.
type BiomeSpec struct {
	ID   BiomeID
	Name string

	// Water color delta added to the base ocean depth color gradient (subtle tinting)
	WaterColorOffset ColorOffset

	// Cave configuration (shared pointer into cave package biome vars)
	CaveSpec *cave.CaveBiomeSpec
}

var biomeRegistry = map[BiomeID]*BiomeSpec{
	BiomeShallowReef: {
		ID:               BiomeShallowReef,
		Name:             "Shallow Coral Reef",
		WaterColorOffset: ColorOffset{R: 0, G: 12, B: 15}, // Subtle Cyan/Teal boost
		CaveSpec:         cave.ShallowReefBiome,
	},
	BiomeKelpForest: {
		ID:               BiomeKelpForest,
		Name:             "Kelp Forest",
		WaterColorOffset: ColorOffset{R: -12, G: 18, B: -10}, // Subtle Emerald Green tint
		CaveSpec:         cave.KelpForestBiome,
	},
	BiomeThermalBarrens: {
		ID:               BiomeThermalBarrens,
		Name:             "Thermal Barrens",
		WaterColorOffset: ColorOffset{R: 22, G: -8, B: -15}, // Subtle Amber/Reddish tint
		CaveSpec:         cave.ThermalBarrensBiome,
	},
	BiomeAbyssalBlue: {
		ID:               BiomeAbyssalBlue,
		Name:             "Abyssal Trench",
		WaterColorOffset: ColorOffset{R: 8, G: -12, B: 24}, // Deep Indigo/Violet tint
		CaveSpec:         cave.AbyssalBlueBiome,
	},
}

// GetBiomeInfo returns the BiomeSpec for a biome ID, falling back to BiomeShallowReef if not found.
func GetBiomeInfo(id BiomeID) *BiomeSpec {
	if spec, ok := biomeRegistry[id]; ok {
		return spec
	}
	return biomeRegistry[BiomeShallowReef]
}
