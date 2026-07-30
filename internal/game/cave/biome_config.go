package cave

import (
	"image/color"
)

// SpawnEntry represents a weighted entry for flora, fauna, or minerals.
type SpawnEntry struct {
	Type   string
	Weight float64
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
	FloraSpawns        []SpawnEntry
	FaunaSpawns        []SpawnEntry
	MineralSpawns      []SpawnEntry
}

// SelectWeightedEntry picks an item from a slice of SpawnEntry using roll [0, 1).
func SelectWeightedEntry(entries []SpawnEntry, roll float64) string {
	if len(entries) == 0 {
		return ""
	}
	var total float64
	for _, e := range entries {
		total += e.Weight
	}
	if total <= 0 {
		return entries[0].Type
	}
	target := roll * total
	var current float64
	for _, e := range entries {
		current += e.Weight
		if target <= current {
			return e.Type
		}
	}
	return entries[len(entries)-1].Type
}

var DefaultShallowReefBiome = &CaveBiomeSpec{
	ID:                 "shallow_reef",
	Name:               "Shallow Coral Reef",
	CaveRockColor:      color.RGBA{180, 155, 100, 255},
	CaveSandLightColor: color.RGBA{215, 190, 125, 255},
	CaveSandDarkColor:  color.RGBA{150, 130, 80, 255},
	CaveStrokeColor:    color.RGBA{210, 185, 120, 255},
	CaveAmbientTint:    color.RGBA{10, 50, 110, 255},
	FloraSpawns: []SpawnEntry{
		{Type: "coral", Weight: 50},
		{Type: "kelp", Weight: 40},
		{Type: "shock_kelp", Weight: 10},
	},
	FaunaSpawns: []SpawnEntry{
		{Type: "passive_fish", Weight: 60},
		{Type: "passive_crab", Weight: 30},
		{Type: "sand_viper", Weight: 10},
	},
	MineralSpawns: []SpawnEntry{
		{Type: "titanium", Weight: 50},
		{Type: "copper", Weight: 40},
		{Type: "quartz", Weight: 10},
	},
}
