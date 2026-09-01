package game

import (
	"fmt"

	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// GetSkiff returns the active or overworld Skiff instance, or nil if none exists.
func (g *Game) GetSkiff() *vehicle.Skiff {
	if s, ok := g.ActiveVehicle.(*vehicle.Skiff); ok {
		return s
	}
	for _, v := range g.OverworldVehicles {
		if s, ok := v.(*vehicle.Skiff); ok {
			return s
		}
	}
	return nil
}

// HasSkiff reports whether a Skiff is active or deployed in the world.
func (g *Game) HasSkiff() bool {
	return g.GetSkiff() != nil
}

// DeploySkiffAtBase spawns a fresh Skiff into clear water next to Life Pod 5.
func (g *Game) DeploySkiffAtBase() *vehicle.Skiff {
	if skiff := g.GetSkiff(); skiff != nil {
		return skiff
	}
	near := gvec.Vec2{
		X: g.baseStation.Pos.X,
		Y: g.baseStation.Pos.Y + 64.0,
	}
	deployPos := g.findNearestClearWaterDeployPos(near, vehicle.SkiffArchetype.Dims)
	skiff := vehicle.NewSkiff(deployPos.X, deployPos.Y)
	g.OverworldVehicles = append(g.OverworldVehicles, skiff)
	g.NotifyQuestVehicleDeployed(vehicle.VehicleSkiff)
	return skiff
}

// HasVehicleInWorldOrDock checks if a vehicle ID exists active, in world, or docked in the Skiff.
func (g *Game) HasVehicleInWorldOrDock(id vehicle.VehicleID) bool {
	if id == vehicle.VehicleSkiff {
		return g.HasSkiff()
	}
	if g.ActiveVehicle != nil && g.ActiveVehicle.GetID() == id {
		return true
	}
	for _, v := range g.OverworldVehicles {
		if v != nil && v.GetID() == id {
			return true
		}
	}
	for _, vList := range g.CaveVehicles {
		for _, v := range vList {
			if v != nil && v.GetID() == id {
				return true
			}
		}
	}
	if skiff := g.GetSkiff(); skiff != nil {
		if skiff.HasDocked(id) {
			return true
		}
	}
	return false
}

// DeploySubFromSkiff launches a docked sub from the Skiff directly into the cave trench below.
func (g *Game) DeploySubFromSkiff(bayIdx int) {
	skiff := g.GetSkiff()
	if skiff == nil {
		g.SetMineWarning("No Skiff available to deploy from!", 120, 2)
		return
	}
	dv := skiff.GetDocked(bayIdx)
	if dv == nil {
		g.SetMineWarning("Docking bay is empty!", 120, 2)
		return
	}

	pX := g.player.Pos.X + g.player.Width/2.0
	pY := g.player.Pos.Y + g.player.Height/2.0
	tx := int(pX) / config.TileSize
	ty := int(pY) / config.TileSize

	v, ok := skiff.Undock(bayIdx, 0, 0)
	if !ok || v == nil {
		return
	}

	g.showInventory = false
	g.EnterCaveWithVehicle(tx, ty, v)
	g.SetMineWarning(fmt.Sprintf("Deployed %s into trench (%d, %d)!", v.GetName(), tx, ty), 150, 1)
}

// DeploySubInCave spawns a docked sub via tether into the cave near the entrance/player.
func (g *Game) DeploySubInCave(bayIdx int) {
	skiff := g.GetSkiff()
	if skiff == nil {
		g.SetMineWarning("No Skiff on surface to tether deploy from!", 120, 2)
		return
	}
	dv := skiff.GetDocked(bayIdx)
	if dv == nil {
		g.SetMineWarning("Docking bay is empty!", 120, 2)
		return
	}

	spawnX := g.player.Pos.X
	spawnY := g.player.Pos.Y

	v, ok := skiff.Undock(bayIdx, spawnX, spawnY)
	if !ok || v == nil {
		return
	}

	g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], v)
	g.ActiveVehicle = v
	audio.Get().PlaySFX("sfx/vehicle_enter.wav")
	g.SetMineWarning(fmt.Sprintf("Deployed %s via Skiff tether!", v.GetName()), 150, 1)
	g.showInventory = false
}

// WinchRecallSub retrieves a deployed sub from any cave trench back into its Skiff docking bay.
func (g *Game) WinchRecallSub(bayIdx int) {
	skiff := g.GetSkiff()
	if skiff == nil {
		g.SetMineWarning("No Skiff available for winch recovery!", 120, 2)
		return
	}

	var targetID vehicle.VehicleID
	switch bayIdx {
	case 0:
		targetID = vehicle.VehicleScoutSub
	case 1:
		targetID = vehicle.VehicleHeavyMech
	default:
		return
	}

	// Search active vehicle
	if g.ActiveVehicle != nil && g.ActiveVehicle.GetID() == targetID {
		v := g.ActiveVehicle
		g.ActiveVehicle = nil
		for key, list := range g.CaveVehicles {
			removeVehicleFromList(&list, v)
			g.CaveVehicles[key] = list
		}
		skiff.Dock(v)
		audio.Get().PlaySFX("sfx/airlock_cycle.wav")
		g.SetMineWarning(fmt.Sprintf("Winched %s back to Skiff dock!", v.GetName()), 150, 1)
		return
	}

	// Search cave vehicles
	for key, list := range g.CaveVehicles {
		for _, v := range list {
			if v != nil && v.GetID() == targetID {
				removeVehicleFromList(&list, v)
				g.CaveVehicles[key] = list
				skiff.Dock(v)
				audio.Get().PlaySFX("sfx/airlock_cycle.wav")
				g.SetMineWarning(fmt.Sprintf("Winched %s back to Skiff dock!", v.GetName()), 150, 1)
				return
			}
		}
	}

	g.SetMineWarning("No deployed "+string(targetID)+" found to winch!", 120, 2)
}
