package game

import (
	"fmt"
	"math"

	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// EnterCave handles the transition from Overworld to Cave at trench coordinate tx, ty.
func (g *Game) EnterCave(tx, ty int) {
	g.ActiveVehicle = nil // Ensure player is on foot in cave to prevent carrying over overworld vehicles
	g.lastOverworldX = g.player.Pos.X
	g.lastOverworldY = g.player.Pos.Y
	g.activeTrenchX = tx
	g.activeTrenchY = ty

	var activeCave cave.Cave
	outOfBounds := tx < 0 || tx >= g.world.Width || ty < 0 || ty >= g.world.Height

	if outOfBounds {
		g.activeTrenchKey = "void_dive"
		activeCave = cave.NewVoidCave()
		g.caveState.CaveGrid = nil
		g.caveState.IsShallow = false
	} else {
		g.activeTrenchKey = fmt.Sprintf("%d_%d", tx, ty)
		grid := g.world.GetCave(tx, ty)
		g.caveState.CaveGrid = grid

		switch g.world.OverworldMap[tx][ty] {
		case world.TileTrench:
			activeCave = cave.NewOrganicTrenchCave(grid)
			g.caveState.IsShallow = false
		case world.TileWreckage:
			// Find spawn tile
			spawnTx, spawnTy := 50, 50
			for x := 45; x < 55; x++ {
				for y := 45; y < 55; y++ {
					if x >= 0 && x < g.world.Width && y >= 0 && y < g.world.Height {
						if g.world.OverworldMap[x][y] == world.TileWater {
							spawnTx, spawnTy = x, y
							break
						}
					}
				}
			}

			// Gather and sort wreckages
			type wreckage struct {
				wtx, wty int
				metric   float64
			}
			var wreckages []wreckage
			for x := 0; x < g.world.Width; x++ {
				for y := 0; y < g.world.Height; y++ {
					if g.world.OverworldMap[x][y] == world.TileWreckage {
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
			shipIndex := -1
			for idx, w := range wreckages {
				if w.wtx == tx && w.wty == ty {
					shipIndex = idx
					break
				}
			}
			if shipIndex == -1 {
				shipIndex = 0
			}

			activeCave = cave.NewWreckageCorridorCave(grid, shipIndex)
			g.caveState.IsShallow = false
		default:
			activeCave = cave.NewShallowSeabedCave(grid)
			g.caveState.IsShallow = true
		}
	}
	g.caveState.ActiveCave = activeCave

	if _, exists := g.caveNodes[g.activeTrenchKey]; !exists {
		g.caveNodes[g.activeTrenchKey] = activeCave.GenerateResources(int64(tx*97 + ty*41))
	}
	g.caveState.Nodes = g.caveNodes[g.activeTrenchKey]

	if _, exists := g.caveEntities[g.activeTrenchKey]; !exists {
		g.caveEntities[g.activeTrenchKey] = activeCave.GenerateEntities(int64(tx*97 + ty*41))
	}
	g.caveState.Entities = g.caveEntities[g.activeTrenchKey]

	if outOfBounds {
		g.player.Pos.X = float64(30 * config.TileSize)
	} else {
		g.player.Pos.X = float64(len(g.caveState.CaveGrid)/2*config.TileSize) + (config.TileSize-g.player.Width)/2
	}
	g.player.Pos.Y = config.TileSize * 2
	g.player.Vel = gvec.Vec2{}

	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
	g.TransitionTo(g.caveState)
}

// ExitCave handles the transition from Cave to Overworld.
func (g *Game) ExitCave() {
	targetX := g.lastOverworldX
	targetY := g.lastOverworldY - config.TileSize*0.6

	// Fallback to the original entry position (guaranteed to be water/non-solid)
	// if the shifted target position would place the player inside solid land.
	if g.overworldState != nil && g.overworldState.IsSolid(targetX, targetY, g.player.Width, g.player.Height) {
		targetX = g.lastOverworldX
		targetY = g.lastOverworldY
	}

	g.player.Pos.X = targetX
	g.player.Pos.Y = targetY
	g.player.Vel = gvec.Vec2{X: 0, Y: -1.5}

	g.caveNodes[g.activeTrenchKey] = g.caveState.Nodes
	g.caveEntities[g.activeTrenchKey] = g.caveState.Entities

	vehicles := g.CaveVehicles[g.activeTrenchKey]
	if len(vehicles) > 0 {
		g.SetMineWarning(fmt.Sprintf("VEHICLE BEACON ACTIVE AT (%d, %d)", g.activeTrenchX, g.activeTrenchY), 180, 1)
	}

	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
	g.TransitionTo(g.overworldState)
}

