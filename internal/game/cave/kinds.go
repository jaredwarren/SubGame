package cave

import (
	"fmt"
	"math/rand"
)

// CaveBuildParams carries construction inputs for a CaveBuilder.
type CaveBuildParams struct {
	Rand          *rand.Rand
	Biome         *CaveBiomeSpec
	DistToLand    float64
	HasLeftWater  bool
	HasRightWater bool
}

// CaveBuilder constructs a Cave for a named kind (gameplay + tooling).
type CaveBuilder interface {
	ID() string
	Name() string
	Build(p CaveBuildParams) Cave
}

// CaveKindDef is one entry in the caveKinds table.
type CaveKindDef struct {
	KindID   string
	KindName string
	Factory  func(CaveBuildParams) Cave
}

func (k *CaveKindDef) ID() string   { return k.KindID }
func (k *CaveKindDef) Name() string { return k.KindName }
func (k *CaveKindDef) Build(p CaveBuildParams) Cave {
	return k.Factory(p)
}

// caveKinds is the ordered registry of constructible cave kinds.
var caveKinds = []*CaveKindDef{
	{
		KindID: "shallow", KindName: "Shallow Seabed",
		Factory: func(p CaveBuildParams) Cave {
			grid := GenerateShallowSeabedGrid(p.Rand, p.DistToLand, p.HasLeftWater, p.HasRightWater)
			return NewShallowSeabedCaveWithBiome(grid, p.Biome)
		},
	},
	{
		KindID: "trench_shallow", KindName: "Trench (shallow layer)",
		Factory: func(p CaveBuildParams) Cave {
			grid := GenerateTrenchShallowGrid(p.Rand, p.DistToLand, p.HasLeftWater, p.HasRightWater)
			return NewTrenchShallowCave(grid)
		},
	},
	{
		KindID: "trench", KindName: "Organic Trench (deep)",
		Factory: func(p CaveBuildParams) Cave {
			grid := GenerateOrganicTrenchGrid(p.Rand)
			return NewOrganicTrenchCave(grid)
		},
	},
	{
		KindID: "wreckage", KindName: "Wreckage Corridor",
		Factory: func(p CaveBuildParams) Cave {
			grid := GenerateWreckageGrid(p.Rand)
			return NewWreckageCorridorCave(grid, 0)
		},
	},
	{
		KindID: "void", KindName: "Void",
		Factory: func(p CaveBuildParams) Cave {
			return NewVoidCave()
		},
	},
	{
		KindID: "shock_kelp_shallow", KindName: "Shock Kelp (shallow layer)",
		Factory: func(p CaveBuildParams) Cave {
			grid := GenerateShockKelpShallowGrid(p.Rand, p.DistToLand, p.HasLeftWater, p.HasRightWater)
			return NewShockKelpShallowCave(grid)
		},
	},
	{
		KindID: "shock_kelp", KindName: "Shock Kelp (deep)",
		Factory: func(p CaveBuildParams) Cave {
			grid := GenerateShockKelpCaveGrid(p.Rand)
			return NewShockKelpCave(grid)
		},
	},
	{
		KindID: "thermo", KindName: "Thermo Cave",
		Factory: func(p CaveBuildParams) Cave {
			grid := GenerateThermoCaveGrid(p.Rand)
			return NewThermoCave(grid)
		},
	},
}

// Kind returns the registered CaveBuilder for id, or nil.
func Kind(id string) CaveBuilder {
	for _, k := range caveKinds {
		if k.KindID == id {
			return k
		}
	}
	return nil
}

// AllKinds returns registered cave builders in table order.
func AllKinds() []CaveBuilder {
	out := make([]CaveBuilder, len(caveKinds))
	for i, k := range caveKinds {
		out[i] = k
	}
	return out
}

// BuildKind constructs a cave for the given kind ID.
// biome may be nil for kinds that ignore it (void, deep caves with fixed biomes).
func BuildKind(kindID string, biome *CaveBiomeSpec, seed int64) (Cave, error) {
	b := Kind(kindID)
	if b == nil {
		return nil, fmt.Errorf("unknown cave kind %q", kindID)
	}
	return b.Build(CaveBuildParams{
		Rand:          rand.New(rand.NewSource(seed)),
		Biome:         biome,
		DistToLand:    8.0,
		HasLeftWater:  true,
		HasRightWater: true,
	}), nil
}
