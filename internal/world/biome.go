package world

import (
	"image/color"

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

	// Cave configuration
	CaveSpec cave.CaveBiomeSpec
}

var biomeRegistry = map[BiomeID]*BiomeSpec{
	BiomeShallowReef: {
		ID:               BiomeShallowReef,
		Name:             "Shallow Coral Reef",
		WaterColorOffset: ColorOffset{R: 0, G: 12, B: 15}, // Subtle Cyan/Teal boost
		CaveSpec: cave.CaveBiomeSpec{
			ID:                 "shallow_reef",
			Name:               "Shallow Coral Reef",
			CaveRockColor:      color.RGBA{180, 155, 100, 255},
			CaveSandLightColor: color.RGBA{215, 190, 125, 255},
			CaveSandDarkColor:  color.RGBA{150, 130, 80, 255},
			CaveStrokeColor:    color.RGBA{210, 185, 120, 255},
			CaveAmbientTint:    color.RGBA{10, 50, 110, 255},
			FloraSpawns: []cave.SpawnEntry{
				{Type: "coral", Weight: 50},
				{Type: "kelp", Weight: 40},
				{Type: "shock_kelp", Weight: 10},
			},
			FaunaSpawns: []cave.SpawnEntry{
				{Type: "passive_fish", Weight: 60},
				{Type: "passive_crab", Weight: 30},
				{Type: "sand_viper", Weight: 10},
			},
			MineralSpawns: []cave.SpawnEntry{
				{Type: "titanium", Weight: 50},
				{Type: "copper", Weight: 40},
				{Type: "quartz", Weight: 10},
			},
		},
	},
	BiomeKelpForest: {
		ID:               BiomeKelpForest,
		Name:             "Kelp Forest",
		WaterColorOffset: ColorOffset{R: -12, G: 18, B: -10}, // Subtle Emerald Green tint
		CaveSpec: cave.CaveBiomeSpec{
			ID:                 "kelp_forest",
			Name:               "Kelp Forest",
			CaveRockColor:      color.RGBA{120, 145, 95, 255},
			CaveSandLightColor: color.RGBA{160, 185, 120, 255},
			CaveSandDarkColor:  color.RGBA{90, 115, 70, 255},
			CaveStrokeColor:    color.RGBA{140, 165, 110, 255},
			CaveAmbientTint:    color.RGBA{10, 70, 60, 255},
			FloraSpawns: []cave.SpawnEntry{
				{Type: "kelp", Weight: 60},
				{Type: "shock_kelp", Weight: 30},
				{Type: "coral", Weight: 10},
			},
			FaunaSpawns: []cave.SpawnEntry{
				{Type: "passive_fish", Weight: 45},
				{Type: "passive_crab", Weight: 45},
				{Type: "sand_viper", Weight: 10},
			},
			MineralSpawns: []cave.SpawnEntry{
				{Type: "quartz", Weight: 50},
				{Type: "copper", Weight: 35},
				{Type: "titanium", Weight: 15},
			},
		},
	},
	BiomeThermalBarrens: {
		ID:               BiomeThermalBarrens,
		Name:             "Thermal Barrens",
		WaterColorOffset: ColorOffset{R: 22, G: -8, B: -15}, // Subtle Amber/Reddish tint
		CaveSpec: cave.CaveBiomeSpec{
			ID:                 "thermal_barrens",
			Name:               "Thermal Barrens",
			CaveRockColor:      color.RGBA{110, 85, 75, 255},
			CaveSandLightColor: color.RGBA{150, 115, 95, 255},
			CaveSandDarkColor:  color.RGBA{80, 60, 50, 255},
			CaveStrokeColor:    color.RGBA{130, 95, 80, 255},
			CaveAmbientTint:    color.RGBA{80, 30, 30, 255},
			FloraSpawns: []cave.SpawnEntry{
				{Type: "shatter_bulb", Weight: 50},
				{Type: "shock_kelp", Weight: 35},
				{Type: "coral", Weight: 15},
			},
			FaunaSpawns: []cave.SpawnEntry{
				{Type: "sand_viper", Weight: 50},
				{Type: "passive_crab", Weight: 30},
				{Type: "passive_fish", Weight: 20},
			},
			MineralSpawns: []cave.SpawnEntry{
				{Type: "quartz", Weight: 40},
				{Type: "abyssal_ore", Weight: 30},
				{Type: "copper", Weight: 30},
			},
		},
	},
	BiomeAbyssalBlue: {
		ID:               BiomeAbyssalBlue,
		Name:             "Abyssal Trench",
		WaterColorOffset: ColorOffset{R: 8, G: -12, B: 24}, // Deep Indigo/Violet tint
		CaveSpec: cave.CaveBiomeSpec{
			ID:                 "abyssal_blue",
			Name:               "Abyssal Trench",
			CaveRockColor:      color.RGBA{80, 90, 115, 255},
			CaveSandLightColor: color.RGBA{110, 125, 155, 255},
			CaveSandDarkColor:  color.RGBA{55, 65, 85, 255},
			CaveStrokeColor:    color.RGBA{95, 110, 135, 255},
			CaveAmbientTint:    color.RGBA{15, 20, 60, 255},
			FloraSpawns: []cave.SpawnEntry{
				{Type: "shatter_bulb", Weight: 60},
				{Type: "shock_kelp", Weight: 30},
				{Type: "coral", Weight: 10},
			},
			FaunaSpawns: []cave.SpawnEntry{
				{Type: "sand_viper", Weight: 55},
				{Type: "passive_fish", Weight: 30},
				{Type: "passive_crab", Weight: 15},
			},
			MineralSpawns: []cave.SpawnEntry{
				{Type: "nickel", Weight: 45},
				{Type: "abyssal_ore", Weight: 35},
				{Type: "titanium", Weight: 20},
			},
		},
	},
}

// GetBiomeInfo returns the BiomeSpec for a biome ID, falling back to BiomeShallowReef if not found.
func GetBiomeInfo(id BiomeID) *BiomeSpec {
	if spec, ok := biomeRegistry[id]; ok {
		return spec
	}
	return biomeRegistry[BiomeShallowReef]
}
