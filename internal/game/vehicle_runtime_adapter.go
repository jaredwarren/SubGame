package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	gv "github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

type vehicleInputAdapter struct {
	input InputSource
}

func (a vehicleInputAdapter) Cursor() gvec.Vec2 {
	cursor := a.input.Cursor()
	return gvec.Vec2{X: cursor.X, Y: cursor.Y}
}

func (a vehicleInputAdapter) IsKeyJustPressed(k ebiten.Key) bool {
	return a.input.IsKeyJustPressed(k)
}

func (a vehicleInputAdapter) IsKeyPressed(k ebiten.Key) bool {
	return a.input.IsKeyPressed(k)
}

type vehicleRuntimeAdapter struct {
	g *Game
}

func (a vehicleRuntimeAdapter) TimeOfDay() float64 {
	return a.g.TimeOfDay
}

func (a vehicleRuntimeAdapter) IsActiveVehicle(v gv.Vehicle) bool {
	return a.g.ActiveVehicle == v
}

func (a vehicleRuntimeAdapter) Input() gv.InputSource {
	return vehicleInputAdapter{input: a.g.Input}
}

func (a vehicleRuntimeAdapter) PlayerScreenCenter() gvec.Vec2 {
	return gvec.Vec2{X: ScreenWidth / 2.0, Y: ScreenHeight / 2.0}
}

func (a vehicleRuntimeAdapter) PlayerSlowed() bool {
	return a.g.playerSlowed
}

func (a vehicleRuntimeAdapter) IsOverworldSolidAt(tx, ty int) bool {
	if tx < 0 || tx >= a.g.world.Width || ty < 0 || ty >= a.g.world.Height {
		return true
	}
	return a.g.world.OverworldMap[tx][ty] == world.TileLand
}

func (a vehicleRuntimeAdapter) IsCaveSolidAt(tx, ty int) bool {
	grid := a.g.caveState.CaveGrid
	if len(grid) == 0 || len(grid[0]) == 0 {
		return false
	}
	if tx < 0 || tx >= len(grid) {
		return true
	}
	if ty < 0 {
		return false
	}
	if ty >= len(grid[0]) {
		return true
	}
	return grid[tx][ty]
}

func (a vehicleRuntimeAdapter) CanUseSonar() bool {
	return a.g.SonarTimer <= 0
}

func (a vehicleRuntimeAdapter) ActivateSonar(source gvec.Vec2, pulse gv.SonarPulse) {
	a.g.SonarTimer = pulse.DurationTicks
	a.g.SonarRadius = 0
	a.g.SonarRadiusStep = pulse.RadiusStep
	a.g.SonarSourceX = source.X
	a.g.SonarSourceY = source.Y
}

func (a vehicleRuntimeAdapter) RemoveCaveNodeAt(tx, ty int) {
	for idx, node := range a.g.caveState.Nodes {
		nodeTx, nodeTy := node.GetTilePos()
		if nodeTx == tx && nodeTy == ty {
			a.g.caveState.Nodes = append(a.g.caveState.Nodes[:idx], a.g.caveState.Nodes[idx+1:]...)
			return
		}
	}
}
