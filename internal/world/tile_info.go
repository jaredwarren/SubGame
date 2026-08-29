package world

import (
	"math"
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/game/cave"
)

// TileTypeInfo describes the behavior and metadata for a specific TileType.
// Consumers look up this info via GetTileInfo instead of hard-coding switch statements.
type TileTypeInfo struct {
	// IsWater indicates this tile counts as water for BFS distance maps, wave rendering, and color computation.
	IsWater bool

	// IsDiveable indicates the player can press [E] to dive on this tile.
	IsDiveable bool

	// DivePrompt is the text shown when the player hovers over a diveable tile.
	// Empty string falls back to "Press [E] to Dive".
	DivePrompt string

	// EstDiveDepth is the fixed depth text shown on the HUD (e.g. "Est. Dive Depth: Trench (120m)").
	// Empty string means the depth is calculated dynamically from distance-to-land.
	EstDiveDepth string

	// ScatterCount is how many of this feature to scatter in the ocean during world generation (0 = don't scatter).
	ScatterCount int

	// GenerateGrid returns a cave grid for this tile type given a seeded RNG.
	// If nil, the default shallow seabed grid generator is used.
	GenerateGrid func(r *rand.Rand) [][]bool

	// CaveFactory creates a cave.Cave from a grid and world context.
	// If nil, defaults to NewShallowSeabedCave.
	CaveFactory func(grid [][]bool, w *World, tx, ty int) cave.Cave

	// IsShallow indicates whether this cave type is shallow (affects cave state).
	IsShallow bool

	// Subterranean configures the two-tier shallow->deep chasm transition system if non-nil.
	Subterranean *SubterraneanSpec
}

// SubterraneanSpec defines configuration for two-tier cave systems where diving on an
// overworld tile enters an open shallow seabed cave layer with an organic funneled chasm/fissure,
// which seamlessly transitions into a subterranean deep grotto.
type SubterraneanSpec struct {
	DeepKeySuffix       string
	DeepCaveType        cave.CaveType
	DeepMusicTrack      string
	ShallowBiome        *cave.CaveBiomeSpec
	GenerateShallowGrid func(r *rand.Rand, distToLand float64, hasLeftWater, hasRightWater bool) [][]bool
	ShallowFactory      func(grid [][]bool, w *World, tx, ty int) cave.Cave
	GenerateDeepGrid    func(r *rand.Rand) [][]bool
	DeepFactory         func(grid [][]bool, w *World, tx, ty int) cave.Cave
}

// tileRegistry maps each TileType to its info.
var tileRegistry = map[TileType]*TileTypeInfo{
	TileWater: {
		IsWater:    true,
		IsDiveable: true,
		IsShallow:  true,
		CaveFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
			var caveSpec *cave.CaveBiomeSpec
			if w != nil && tx >= 0 && tx < w.Width && ty >= 0 && ty < w.Height {
				spec := GetBiomeInfo(w.BiomeMap[tx][ty])
				if spec != nil {
					caveSpec = spec.CaveSpec
				}
			}
			return cave.NewShallowSeabedCaveWithBiome(grid, caveSpec)
		},
	},
	TileLand: {
		IsWater:    false,
		IsDiveable: false,
	},
	TileTrench: {
		IsWater:      true,
		IsDiveable:   true,
		DivePrompt:   "Press [E] to Dive",
		EstDiveDepth: "Est. Dive Depth: Trench (120m)",
		ScatterCount: 0,
		IsShallow:    true,
		CaveFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
			return cave.NewTrenchShallowCave(grid)
		},
		Subterranean: &SubterraneanSpec{
			DeepKeySuffix:       "_trench",
			DeepCaveType:        cave.CaveOrganicTrench,
			DeepMusicTrack:      "music/cave_deep.mp3",
			ShallowBiome:        cave.AbyssalBlueBiome,
			GenerateShallowGrid: cave.GenerateTrenchShallowGrid,
			ShallowFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
				return cave.NewTrenchShallowCave(grid)
			},
			GenerateDeepGrid: cave.GenerateOrganicTrenchGrid,
			DeepFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
				return cave.NewOrganicTrenchCave(grid)
			},
		},
	},
	TileWreckage: {
		IsWater:      true,
		IsDiveable:   true,
		DivePrompt:   "Press [E] to Salvage Wreckage",
		ScatterCount: 3,
		GenerateGrid: cave.GenerateWreckageGrid,
		CaveFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
			shipIndex := w.ComputeWreckageShipIndex(tx, ty)
			return cave.NewWreckageCorridorCave(grid, shipIndex)
		},
		IsShallow: false,
	},
	TileShockKelpCave: {
		IsWater:      true,
		IsDiveable:   true,
		DivePrompt:   "Press [E] to Dive",
		EstDiveDepth: "Est. Dive Depth: Shock Kelp (75m)",
		ScatterCount: 0,
		IsShallow:    true,
		CaveFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
			return cave.NewShockKelpShallowCave(grid)
		},
		Subterranean: &SubterraneanSpec{
			DeepKeySuffix:       "_shock",
			DeepCaveType:        cave.CaveShockKelp,
			DeepMusicTrack:      "music/cave_kelp.mp3",
			ShallowBiome:        cave.KelpForestBiome,
			GenerateShallowGrid: cave.GenerateShockKelpShallowGrid,
			ShallowFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
				return cave.NewShockKelpShallowCave(grid)
			},
			GenerateDeepGrid: cave.GenerateShockKelpCaveGrid,
			DeepFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
				return cave.NewShockKelpCave(grid)
			},
		},
	},
	TileThermoCave: {
		IsWater:      true,
		IsDiveable:   true,
		DivePrompt:   "Press [E] to Enter Thermo Cave",
		EstDiveDepth: "Est. Dive Depth: Thermo Cave (45m)",
		ScatterCount: 0,
		GenerateGrid: cave.GenerateThermoCaveGrid,
		CaveFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
			return cave.NewThermoCave(grid)
		},
		IsShallow: true,
	},
}

// GetTileInfo returns the TileTypeInfo for a tile type, or nil if unregistered.
func GetTileInfo(t TileType) *TileTypeInfo {
	return tileRegistry[t]
}

// AllTileInfos returns the full registry map for iteration.
func AllTileInfos() map[TileType]*TileTypeInfo {
	return tileRegistry
}

// WreckageInfo defines the descriptive presentation metadata for a tiered shipwreck.
type WreckageInfo struct {
	Name       string
	EstDepth   string
	DivePrompt string
}

// GetWreckageInfo returns the tier-specific metadata for the given shipIndex (0, 1, or 2).
func GetWreckageInfo(shipIndex int) WreckageInfo {
	switch shipIndex {
	case 1:
		return WreckageInfo{
			Name:       "Submersible Transport Wreckage",
			EstDepth:   "Est. Dive Depth: 60m (Zero Oxygen Below 40m)",
			DivePrompt: "Press [E] to Salvage Transport Wreck (60m)",
		}
	case 2:
		return WreckageInfo{
			Name:       "AetherCorp Flagship Wreckage",
			EstDepth:   "Est. Dive Depth: 100m+ (Deep Vault Sealed)",
			DivePrompt: "Press [E] to Salvage Flagship Wreck (100m+)",
		}
	default:
		return WreckageInfo{
			Name:       "Research Tender Wreckage",
			EstDepth:   "Est. Dive Depth: 30m (Scout Sub Blueprints < 40m)",
			DivePrompt: "Press [E] to Salvage Research Tender (30m)",
		}
	}
}

// ComputeWreckageShipIndex returns the sorted index of the wreckage at (tx, ty)
// among all wreckages, ordered by Euclidean distance from the Lifepod spawn point.
func (w *World) ComputeWreckageShipIndex(tx, ty int) int {
	// Find spawn reference tile (Shallow Reef water nearest center)
	spawnTx, spawnTy := w.FindLifepodSpawn()

	type wreckage struct {
		wtx, wty int
		dist     float64
	}
	var wreckages []wreckage
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			if w.OverworldMap[x][y] == TileWreckage {
				dx := float64(x - spawnTx)
				dy := float64(y - spawnTy)
				dist := math.Hypot(dx, dy)
				wreckages = append(wreckages, wreckage{
					wtx:  x,
					wty:  y,
					dist: dist,
				})
			}
		}
	}

	// Sort by distance ascending
	for i := 0; i < len(wreckages); i++ {
		for j := i + 1; j < len(wreckages); j++ {
			if wreckages[i].dist > wreckages[j].dist {
				wreckages[i], wreckages[j] = wreckages[j], wreckages[i]
			}
		}
	}

	// Find our index
	for idx, wr := range wreckages {
		if wr.wtx == tx && wr.wty == ty {
			return idx
		}
	}
	return 0
}
