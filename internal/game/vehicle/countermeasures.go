package vehicle

import (
	"math"

	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

const (
	countermeasureBatteryCost = 5.0
	countermeasureDepletedMsg = "COUNTERMEASURE DEPLETED / LOW POWER"
)

func emitCountermeasureDepleted(rt Runtime) {
	rt.Emit(SetWarningCmd{
		Message:  countermeasureDepletedMsg,
		Duration: 90,
		Level:    2,
	})
}

// TryLaunchDecoy fires a sonic decoy when cargo and battery allow.
func TryLaunchDecoy(rt Runtime, cargo *item.Inventory, battery *float64, pos, dims gvec.Vec2, facing float64) bool {
	var decoyAmmo item.SonicDecoy
	if !cargo.Has(&decoyAmmo, 1) || *battery < countermeasureBatteryCost {
		emitCountermeasureDepleted(rt)
		return false
	}
	cargo.Remove(&decoyAmmo, 1)
	*battery -= countermeasureBatteryCost

	cosF := math.Cos(facing)
	sinF := math.Sin(facing)
	spawnX := pos.X + dims.X/2.0 + cosF*(dims.X/2.0+10.0)
	spawnY := pos.Y + dims.Y/2.0 + sinF*(dims.Y/2.0+10.0)
	rt.Emit(SpawnDecoyCmd{
		Pos: gvec.Vec2{X: spawnX, Y: spawnY},
		Vel: gvec.Vec2{X: cosF * 6.0, Y: sinF * 6.0},
	})
	return true
}

// TryLaunchDeterrent deploys a chemical deterrent cloud when cargo and battery allow.
func TryLaunchDeterrent(rt Runtime, cargo *item.Inventory, battery *float64, pos, dims gvec.Vec2) bool {
	var chemicalAmmo item.ChemicalDeterrent
	if !cargo.Has(&chemicalAmmo, 1) || *battery < countermeasureBatteryCost {
		emitCountermeasureDepleted(rt)
		return false
	}
	cargo.Remove(&chemicalAmmo, 1)
	*battery -= countermeasureBatteryCost
	rt.Emit(SpawnDeterrentCloudCmd{
		Pos: gvec.Vec2{X: pos.X + dims.X/2.0, Y: pos.Y + dims.Y/2.0},
	})
	return true
}
