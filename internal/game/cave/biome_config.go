package cave

import (
	"image/color"

	"github.com/jaredwarren/SubGame/internal/game/resource"
)

// CaveBiomeSpec defines visual and spawn properties for cave views.
type CaveBiomeSpec struct {
	ID                 string
	Name               string
	CaveRockColor      color.RGBA
	CaveSandLightColor color.RGBA
	CaveSandDarkColor  color.RGBA
	CaveStrokeColor    color.RGBA
	CaveAmbientTint    color.RGBA
	FloraSpawns        []SpawnEntry[FloraID]
	FaunaSpawns        []SpawnEntry[FaunaID]
	MineralSpawns      []SpawnEntry[resource.NodeType]
}

// DefaultShallowReefBiome is the fallback biome when no overworld biome is available.
var DefaultShallowReefBiome = &CaveBiomeSpec{
	ID:                 "shallow_reef",
	Name:               "Shallow Coral Reef",
	CaveRockColor:      color.RGBA{180, 155, 100, 255},
	CaveSandLightColor: color.RGBA{215, 190, 125, 255},
	CaveSandDarkColor:  color.RGBA{150, 130, 80, 255},
	CaveStrokeColor:    color.RGBA{210, 185, 120, 255},
	CaveAmbientTint:    color.RGBA{10, 50, 110, 255},
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
