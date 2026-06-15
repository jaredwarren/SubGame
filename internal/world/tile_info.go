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
}

// tileRegistry maps each TileType to its info.
var tileRegistry = map[TileType]*TileTypeInfo{
	TileWater: {
		IsWater:    true,
		IsDiveable: true,
		IsShallow:  true,
	},
	TileLand: {
		IsWater:    false,
		IsDiveable: false,
	},
	TileTrench: {
		IsWater:      true,
		IsDiveable:   true,
		ScatterCount: 6,
		EstDiveDepth: "Est. Dive Depth: Trench (120m)",
		GenerateGrid: cave.GenerateOrganicTrenchGrid,
		CaveFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
			return cave.NewOrganicTrenchCave(grid)
		},
		IsShallow: false,
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
		DivePrompt:   "Press [E] to Enter Shock Kelp Cave",
		EstDiveDepth: "Est. Dive Depth: Shock Kelp Cave (60m)",
		ScatterCount: 4,
		GenerateGrid: cave.GenerateShockKelpCaveGrid,
		CaveFactory: func(grid [][]bool, w *World, tx, ty int) cave.Cave {
			return cave.NewShockKelpCave(grid)
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

// ComputeWreckageShipIndex returns the sorted index of the wreckage at (tx, ty)
// among all wreckages, ordered by distance from a reference spawn point.
func (w *World) ComputeWreckageShipIndex(tx, ty int) int {
	// Find spawn reference tile (water near center)
	spawnTx, spawnTy := 50, 50
	for x := 45; x < 55; x++ {
		for y := 45; y < 55; y++ {
			if x >= 0 && x < w.Width && y >= 0 && y < w.Height {
				if w.OverworldMap[x][y] == TileWater {
					spawnTx, spawnTy = x, y
					break
				}
			}
		}
	}

	type wreckage struct {
		wtx, wty int
		metric   float64
	}
	var wreckages []wreckage
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			if w.OverworldMap[x][y] == TileWreckage {
				dx := float64(x - spawnTx)
				dy := float64(y - spawnTy)
				dist := math.Hypot(dx, dy)
				wreckages = append(wreckages, wreckage{
					wtx:    x,
					wty:    y,
					metric: dist + float64(y),
				})
			}
		}
	}

	// Sort by metric ascending
	for i := 0; i < len(wreckages); i++ {
		for j := i + 1; j < len(wreckages); j++ {
			if wreckages[i].metric > wreckages[j].metric {
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
