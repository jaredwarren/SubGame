package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
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

// vehicleRuntimeAdapter satisfies vehicle.Runtime. Query methods are synchronous
// and read directly from *Game. Mutations are submitted via Emit and buffered in
// cmds; game.Update() drains the queue after all vehicles have ticked.
type vehicleRuntimeAdapter struct {
	g    *Game
	cmds []gv.GameCommand
}

// Emit queues a fire-and-forget command to be processed after all vehicles tick.
func (a *vehicleRuntimeAdapter) Emit(cmd gv.GameCommand) {
	a.cmds = append(a.cmds, cmd)
}

func (a *vehicleRuntimeAdapter) TimeOfDay() float64 {
	return a.g.TimeOfDay
}

func (a *vehicleRuntimeAdapter) IsActiveVehicle(v gv.Vehicle) bool {
	return a.g.ActiveVehicle == v
}

func (a *vehicleRuntimeAdapter) Input() gv.InputSource {
	return vehicleInputAdapter{input: a.g.Input}
}

func (a *vehicleRuntimeAdapter) PlayerScreenCenter() gvec.Vec2 {
	return gvec.Vec2{X: config.ScreenWidth / 2.0, Y: config.ScreenHeight / 2.0}
}

func (a *vehicleRuntimeAdapter) PlayerSlowed() bool {
	return a.g.playerSlowed
}

func (a *vehicleRuntimeAdapter) PlayerStunned() bool {
	return a.g.player.StunTimer > 0
}

func (a *vehicleRuntimeAdapter) IsOverworldSolidAt(tx, ty int) bool {
	if tx < 0 || tx >= a.g.world.Width || ty < 0 || ty >= a.g.world.Height {
		return true
	}
	return a.g.world.OverworldMap[tx][ty] == world.TileLand
}

func (a *vehicleRuntimeAdapter) IsCaveSolidAt(tx, ty int) bool {
	grid := a.g.caveState.CaveGrid
	if len(grid) == 0 || len(grid[0]) == 0 {
		return false
	}
	if tx < 0 || tx >= len(grid) {
		if a.g.caveState.IsShallow {
			currentTx, currentTy := a.g.activeTrenchX, a.g.activeTrenchY
			var neighborTx int
			if tx < 0 {
				neighborTx = currentTx - 1
			} else {
				neighborTx = currentTx + 1
			}
			if neighborTx >= 0 && neighborTx < a.g.world.Width && a.g.world.OverworldMap[neighborTx][currentTy] == world.TileWater {
				return false
			}
		}
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

func (a *vehicleRuntimeAdapter) CanUseSonar() bool {
	return a.g.Sonar.Timer <= 0
}

func (a *vehicleRuntimeAdapter) BaseStationPos() (gvec.Vec2, gvec.Vec2) {
	if a.g.baseStation == nil {
		return gvec.Vec2{}, gvec.Vec2{}
	}
	return a.g.baseStation.Pos, a.g.baseStation.Size
}
