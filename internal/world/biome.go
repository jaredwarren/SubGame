package world

import (
	"github.com/jaredwarren/SubGame/internal/game/cave"
)

// Adjustable parameters for overworld biome transitions.
var (
	// BiomeTransitionIntensity scales the water color offset across biomes to control how
	// pronounced the biome color transitions are in the overworld.
	// Increase this value (e.g. 2.0 - 2.5) to make biome differences more obvious;
	// decrease it (e.g. 1.0) for more subtle transitions.
	BiomeTransitionIntensity = 1.75

	// BiomeBlendRadius controls the smoothing radius (in tiles) between adjacent biomes.
	// A radius of 2 produces a 5x5 tile smooth blend. Set to 1 for sharper transitions,
	// or 3+ for wider, softer transitions.
	BiomeBlendRadius = 2
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

	// Special cave entrance tile spawned in this biome (e.g. TileThermoCave, TileShockKelpCave, TileTrench)
	SpecialCaveTile TileType

	// Percentage/probability of spawning a special cave on an eligible deep-water ocean tile in this biome
	SpecialCaveSpawnChance float64

	// Maximum number of special caves allowed to spawn in this biome
	SpecialCaveMaxCount int

	// Minimum distance (in tiles) between special caves of this type
	SpecialCaveMinDist float64
}

var biomeRegistry = map[BiomeID]*BiomeSpec{
	BiomeShallowReef: {
		ID:                     BiomeShallowReef,
		Name:                   "Shallow Coral Reef",
		WaterColorOffset:       ColorOffset{R: 0, G: 14, B: 18}, // Subtle Cyan/Teal boost
		CaveSpec:               cave.ShallowReefBiome,
		SpecialCaveTile:        TileWater,
		SpecialCaveSpawnChance: 0,
		SpecialCaveMaxCount:    0,
		SpecialCaveMinDist:     0,
	},
	BiomeKelpForest: {
		ID:                     BiomeKelpForest,
		Name:                   "Kelp Forest",
		WaterColorOffset:       ColorOffset{R: -14, G: 20, B: -12}, // Subtle Emerald Green tint
		CaveSpec:               cave.KelpForestBiome,
		SpecialCaveTile:        TileShockKelpCave,
		SpecialCaveSpawnChance: 0.00025, // 0.025% chance per eligible ocean tile
		SpecialCaveMaxCount:    5,
		SpecialCaveMinDist:     8.0,
	},
	BiomeThermalBarrens: {
		ID:                     BiomeThermalBarrens,
		Name:                   "Thermal Barrens",
		WaterColorOffset:       ColorOffset{R: 24, G: -10, B: -18}, // Subtle Amber/Reddish tint
		CaveSpec:               cave.ThermalBarrensBiome,
		SpecialCaveTile:        TileThermoCave,
		SpecialCaveSpawnChance: 0.00030, // 0.030% chance per eligible ocean tile
		SpecialCaveMaxCount:    6,
		SpecialCaveMinDist:     8.0,
	},
	BiomeAbyssalBlue: {
		ID:                     BiomeAbyssalBlue,
		Name:                   "Abyssal Trench",
		WaterColorOffset:       ColorOffset{R: 10, G: -14, B: 26}, // Deep Indigo/Violet tint
		CaveSpec:               cave.AbyssalBlueBiome,
		SpecialCaveTile:        TileTrench,
		SpecialCaveSpawnChance: 0.00035, // 0.035% chance per eligible ocean tile
		SpecialCaveMaxCount:    8,
		SpecialCaveMinDist:     8.0,
	},
}

// GetBiomeInfo returns the BiomeSpec for a biome ID, falling back to BiomeShallowReef if not found.
func GetBiomeInfo(id BiomeID) *BiomeSpec {
	if spec, ok := biomeRegistry[id]; ok {
		return spec
	}
	return biomeRegistry[BiomeShallowReef]
}
