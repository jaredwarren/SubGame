package game

import (
	"fmt"
	"math"

	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/quest"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// compile-time assertion: *Game must satisfy scene.DebugContext
var _ scene.DebugContext = (*Game)(nil)

func (g *Game) GiveItem(name string, qty int) {
	if qty <= 0 {
		return
	}
	it := item.NewItemByName(name)
	if it == nil {
		g.SetMineWarning("Unknown item: "+name, 120, 2)
		return
	}
	leftover := g.player.Inventory.InsertStacks([]item.ItemStack{{Item: it, Quantity: qty}})
	if len(leftover) > 0 {
		leftover = g.player.Hotbar.InsertStacks(leftover)
	}
	g.player.RecalculateUpgrades()
	if len(leftover) > 0 {
		g.SetMineWarning(fmt.Sprintf("Added %dx %s (Inventory full, some lost)", qty-leftover[0].Quantity, name), 120, 1)
	} else {
		g.SetMineWarning(fmt.Sprintf("Added %dx %s", qty, name), 120, 1)
	}
}

func (g *Game) GivePreset(presetName string) {
	switch presetName {
	case "starter":
		g.player.Upgrades.AddItem(&item.Fins{}, 1)
		g.player.Upgrades.AddItem(&item.O2TankHC{}, 1)
		g.player.Hotbar.AddItem(&item.Scanner{}, 1)
		g.player.Inventory.AddItem(&item.CookedFish{}, 5)
		g.player.RecalculateUpgrades()
		g.SetMineWarning("Applied Starter Kit preset", 120, 1)
	case "tools":
		g.player.Upgrades.AddItem(&item.Fins{}, 1)
		g.player.Upgrades.AddItem(&item.O2TankUHC{}, 1)
		g.player.Hotbar.AddItem(&item.Scanner{}, 1)
		g.player.Hotbar.AddItem(&item.Flashlight{}, 1)
		g.player.Hotbar.AddItem(&item.RepairTool{}, 1)
		g.player.RecalculateUpgrades()
		g.SetMineWarning("Applied All Tools preset", 120, 1)
	case "minerals":
		g.player.Inventory.AddItem(&item.Titanium{}, 10)
		g.player.Inventory.AddItem(&item.Copper{}, 10)
		g.player.Inventory.AddItem(&item.Quartz{}, 10)
		g.player.Inventory.AddItem(&item.Nickel{}, 10)
		g.player.Inventory.AddItem(&item.AbyssalOre{}, 10)
		g.player.Inventory.AddItem(&item.ScrapMetal{}, 10)
		g.player.Inventory.AddItem(&item.ElectronicWaste{}, 10)
		g.player.RecalculateUpgrades()
		g.SetMineWarning("Added 10x all minerals", 120, 1)
	case "upgrades":
		g.player.Inventory.AddItem(&item.SonarAmplifier{}, 1)
		g.player.Inventory.AddItem(&item.DecoyLauncher{}, 1)
		g.player.Inventory.AddItem(&item.ChemicalDischarger{}, 1)
		g.player.Inventory.AddItem(&item.ThermalGenerator{}, 1)
		g.player.Inventory.AddItem(&item.PowerCell{}, 2)
		g.SetMineWarning("Added all vehicle upgrades", 120, 1)
	case "rocket":
		g.player.Inventory.AddItem(&item.EscapeRocket{}, 1)
		g.player.Inventory.AddItem(&item.Titanium{}, 20)
		g.player.Inventory.AddItem(&item.Copper{}, 10)
		g.player.Inventory.AddItem(&item.AbyssalOre{}, 10)
		g.player.Inventory.AddItem(&item.Nickel{}, 10)
		g.SetMineWarning("Added Rocket materials preset", 120, 1)
	}
}

func (g *Game) ClearPlayerInventory() {
	g.player.Inventory.Clear()
	g.player.RecalculateUpgrades()
	g.SetMineWarning("Player inventory cleared", 120, 1)
}

func (g *Game) ClearPlayerHotbar() {
	g.player.Hotbar.Clear()
	g.player.RecalculateUpgrades()
	g.SetMineWarning("Player hotbar cleared", 120, 1)
}

func (g *Game) IsGodMode() bool { return g.GodMode }
func (g *Game) ToggleGodMode() {
	g.GodMode = !g.GodMode
	if g.GodMode {
		g.player.CurrentHealth = g.player.MaxHealth
		g.SetMineWarning("God Mode ENABLED", 120, 1)
	} else {
		g.SetMineWarning("God Mode DISABLED", 120, 1)
	}
}

func (g *Game) IsInfiniteO2() bool { return g.InfiniteOxygen }
func (g *Game) ToggleInfiniteO2() {
	g.InfiniteOxygen = !g.InfiniteOxygen
	if g.InfiniteOxygen {
		g.player.CurrentOxygen = g.player.MaxOxygen
		g.SetMineWarning("Infinite Oxygen ENABLED", 120, 1)
	} else {
		g.SetMineWarning("Infinite Oxygen DISABLED", 120, 1)
	}
}

func (g *Game) IsInfiniteStamina() bool { return g.InfiniteStamina }
func (g *Game) ToggleInfiniteStamina() {
	g.InfiniteStamina = !g.InfiniteStamina
	if g.InfiniteStamina {
		g.player.CurrentStamina = g.player.MaxStamina
		g.SetMineWarning("Infinite Stamina ENABLED", 120, 1)
	} else {
		g.SetMineWarning("Infinite Stamina DISABLED", 120, 1)
	}
}

func (g *Game) IsSuperSpeed() bool { return g.SuperSpeed }
func (g *Game) ToggleSuperSpeed() {
	g.SuperSpeed = !g.SuperSpeed
	g.player.SuperSpeed = g.SuperSpeed
	g.player.RecalculateUpgrades()
	if g.SuperSpeed {
		g.SetMineWarning("Super Speed ENABLED (2.5x)", 120, 1)
	} else {
		g.SetMineWarning("Super Speed DISABLED", 120, 1)
	}
}

func (g *Game) IsTimeFrozen() bool { return g.FreezeTimeOfDay }
func (g *Game) ToggleFreezeTime() {
	g.FreezeTimeOfDay = !g.FreezeTimeOfDay
	if g.FreezeTimeOfDay {
		g.SetMineWarning("Time of Day FROZEN", 120, 1)
	} else {
		g.SetMineWarning("Time of Day RESUMED", 120, 1)
	}
}

func (g *Game) IsInfiniteVehicleBattery() bool { return g.InfiniteVehicleBattery }
func (g *Game) ToggleInfiniteVehicleBattery() {
	g.InfiniteVehicleBattery = !g.InfiniteVehicleBattery
	if g.InfiniteVehicleBattery {
		if g.ActiveVehicle != nil {
			g.ActiveVehicle.RechargeBattery(g.ActiveVehicle.GetMaxBattery())
		}
		g.SetMineWarning("Infinite Vehicle Battery ENABLED", 120, 1)
	} else {
		g.SetMineWarning("Infinite Vehicle Battery DISABLED", 120, 1)
	}
}

func (g *Game) IsInfiniteVehicleHull() bool { return g.InfiniteVehicleHull }
func (g *Game) ToggleInfiniteVehicleHull() {
	g.InfiniteVehicleHull = !g.InfiniteVehicleHull
	if g.InfiniteVehicleHull {
		if g.ActiveVehicle != nil {
			g.ActiveVehicle.Repair(g.ActiveVehicle.GetMaxHealth())
		}
		g.SetMineWarning("Infinite Vehicle Hull ENABLED", 120, 1)
	} else {
		g.SetMineWarning("Infinite Vehicle Hull DISABLED", 120, 1)
	}
}

func (g *Game) HealPlayerFull() {
	g.player.CurrentHealth = g.player.MaxHealth
	g.SetMineWarning("Player Health Restored (100 HP)", 120, 1)
}

func (g *Game) RefillO2AndStamina() {
	g.player.CurrentOxygen = g.player.MaxOxygen
	g.player.CurrentStamina = g.player.MaxStamina
	g.SetMineWarning("Oxygen & Stamina Refilled", 120, 1)
}

func (g *Game) KillPlayer() {
	g.player.CurrentHealth = 0
	g.SetMineWarning("Suicide triggered", 120, 2)
}

func (g *Game) TriggerWin() {
	g.TransitionToGameWon()
}

func (g *Game) SpawnVehicle(name string) {
	v := vehicle.NewVehicleByName(name, g.player.Pos.X, g.player.Pos.Y)
	if v == nil {
		g.SetMineWarning("Unknown vehicle: "+name, 120, 2)
		return
	}
	if g.currentState == StateOverworld {
		if v.GetPerspective() != "overworld" {
			g.SetMineWarning(fmt.Sprintf("Cannot spawn %s in overworld! Must spawn in cave.", v.GetName()), 120, 2)
			return
		}
		g.OverworldVehicles = append(g.OverworldVehicles, v)
	} else {
		if v.GetPerspective() != "cave" {
			g.SetMineWarning(fmt.Sprintf("Cannot spawn %s in caves! Must spawn in overworld.", v.GetName()), 120, 2)
			return
		}
		g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], v)
	}
	g.ActiveVehicle = v
	g.SetMineWarning("Spawned and boarded "+v.GetName(), 120, 1)
}

func (g *Game) RepairActiveVehicle() {
	if g.ActiveVehicle == nil {
		g.SetMineWarning("No active vehicle to repair", 120, 2)
		return
	}
	g.ActiveVehicle.Repair(g.ActiveVehicle.GetMaxHealth())
	g.SetMineWarning(g.ActiveVehicle.GetName()+" hull repaired to 100%", 120, 1)
}

func (g *Game) ChargeActiveVehicle() {
	if g.ActiveVehicle == nil {
		g.SetMineWarning("No active vehicle to charge", 120, 2)
		return
	}
	g.ActiveVehicle.RechargeBattery(g.ActiveVehicle.GetMaxBattery())
	g.SetMineWarning(g.ActiveVehicle.GetName()+" battery charged to 100%", 120, 1)
}

func (g *Game) DespawnActiveVehicle() {
	if g.ActiveVehicle == nil {
		g.SetMineWarning("No active vehicle to despawn", 120, 2)
		return
	}
	v := g.ActiveVehicle
	g.removeVehicle(v)
	g.ActiveVehicle = nil
	g.SetMineWarning("Despawned "+v.GetName(), 120, 1)
}

func (g *Game) SetTimeOfDay(tod float64) {
	g.TimeOfDay = tod
	g.SetMineWarning(fmt.Sprintf("Time of Day set to %.0f", tod), 120, 1)
}

func (g *Game) AdvanceTimeOfDay(hours float64) {
	ticks := hours * 600.0 // 1 hour = 600 ticks
	g.TimeOfDay += ticks
	for g.TimeOfDay >= 14400 {
		g.TimeOfDay -= 14400
	}
	g.SetMineWarning(fmt.Sprintf("Advanced time by +%.1fh", hours), 120, 1)
}

func (g *Game) TeleportToLifePod() {
	if g.currentState == StateCave {
		g.ExitCave()
	}
	if g.baseStation != nil {
		g.player.Pos.X = g.baseStation.Pos.X
		g.player.Pos.Y = g.baseStation.Pos.Y - 40
		g.player.Vel = gvec.Vec2{}
		g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
		g.SetMineWarning("Teleported to Life Pod", 120, 1)
	}
}

func (g *Game) TeleportToPOI(poiType world.TileType) {
	if g.world == nil {
		return
	}
	if g.currentState == StateCave {
		g.ExitCave()
	}
	pTX := int(g.player.Pos.X / config.TileSize)
	pTY := int(g.player.Pos.Y / config.TileSize)

	bestTX, bestTY := -1, -1
	bestDist := math.MaxFloat64
	for tx := 0; tx < g.world.Width; tx++ {
		for ty := 0; ty < g.world.Height; ty++ {
			if g.world.OverworldMap[tx][ty] == poiType {
				dx := float64(tx - pTX)
				dy := float64(ty - pTY)
				d := dx*dx + dy*dy
				if d < bestDist {
					bestDist = d
					bestTX, bestTY = tx, ty
				}
			}
		}
	}

	if bestTX >= 0 {
		g.player.Pos.X = float64(bestTX*config.TileSize) + float64(config.TileSize)/2
		g.player.Pos.Y = float64(bestTY*config.TileSize) + float64(config.TileSize)/2
		g.player.Vel = gvec.Vec2{}
		g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
		if g.explorationTracker != nil {
			g.explorationTracker.Reveal(bestTX, bestTY, 3)
		}
		g.SetMineWarning(fmt.Sprintf("Teleported to %v at (%d, %d)", poiType, bestTX, bestTY), 120, 1)
	} else {
		g.SetMineWarning("No POI of that type found in world", 120, 2)
	}
}

func (g *Game) TeleportToVoid() {
	if g.currentState == StateCave {
		g.ExitCave()
	}
	if g.world != nil {
		g.player.Pos.X = float64(g.world.Width*config.TileSize) + 200
		g.player.Pos.Y = float64(g.world.Height*config.TileSize) / 2
		g.player.Vel = gvec.Vec2{}
		g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
		g.SetMineWarning("Teleported to Void Border", 120, 1)
	}
}

func (g *Game) DirectDiveCave(caveType string) {
	if g.currentState == StateCave {
		g.ExitCave()
	}
	var targetTile world.TileType
	switch caveType {
	case "shallow":
		targetTile = world.TileTrench
	case "trench":
		targetTile = world.TileTrench
	case "kelp":
		targetTile = world.TileShockKelpCave
	case "thermo":
		targetTile = world.TileThermoCave
	default:
		targetTile = world.TileTrench
	}

	bestTX, bestTY := 50, 50
	if g.world != nil {
		found := false
		for tx := 0; tx < g.world.Width; tx++ {
			for ty := 0; ty < g.world.Height; ty++ {
				if g.world.OverworldMap[tx][ty] == targetTile {
					bestTX, bestTY = tx, ty
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	g.EnterCave(bestTX, bestTY)
	g.SetMineWarning(fmt.Sprintf("Dove into %s cave at (%d, %d)", caveType, bestTX, bestTY), 120, 1)
}

func (g *Game) SurfaceToOverworld() {
	if g.currentState == StateCave {
		g.ExitCave()
		g.SetMineWarning("Surfaced to Overworld", 120, 1)
	} else {
		g.SetMineWarning("Already in Overworld", 120, 2)
	}
}

func (g *Game) RevealFullMap() {
	if g.world != nil && g.explorationTracker != nil {
		for tx := 0; tx < g.world.Width; tx++ {
			for ty := 0; ty < g.world.Height; ty++ {
				g.explorationTracker.Reveal(tx, ty, 1)
				tt := g.world.OverworldMap[tx][ty]
				switch tt {
				case world.TileTrench, world.TileWreckage, world.TileShockKelpCave, world.TileThermoCave:
					g.explorationTracker.MarkVisited(tx, ty, tt)
				}
			}
		}
		if g.baseMenu != nil {
			g.baseMenu.ResetMapCache()
		}
		g.SetMineWarning("Full Overworld Map Revealed", 150, 1)
	}
}

func (g *Game) ResetFogOfWar() {
	if g.world != nil {
		g.explorationTracker = exploration.NewTracker(g.world.Width, g.world.Height)
		pTX := int(g.player.Pos.X / config.TileSize)
		pTY := int(g.player.Pos.Y / config.TileSize)
		g.explorationTracker.Reveal(pTX, pTY, exploration.RevealRadius)
		if g.baseMenu != nil {
			g.baseMenu.ResetMapCache()
		}
		g.SetMineWarning("Fog of War Reset", 150, 1)
	}
}

func (g *Game) UnlockAllRecipes() {
	for i := range g.craftingRecipes {
		g.craftingRecipes[i].Unlocked = true
	}
	g.SetMineWarning("All Crafting Blueprints Unlocked!", 150, 1)
	audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
}

func (g *Game) UnlockAllLore() {
	if g.storyManager != nil {
		for _, entry := range g.storyManager.GetEntries() {
			entry.Unlocked = true
		}
		g.SetMineWarning("All PDA Lore Logs Unlocked!", 150, 1)
		audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
	}
}

func (g *Game) CompleteCurrentTask() {
	if g.questManager != nil {
		for _, cat := range g.questManager.Categories {
			for _, q := range cat.Quests {
				if !q.Completed {
					for _, t := range q.Tasks {
						if !t.Completed {
							t.CurrentCount = t.RequiredCount
							t.Completed = true
							if q.IsAllTasksCompleted() {
								q.Completed = true
							}
							g.SetMineWarning("Completed task: "+t.Description, 150, 1)
							audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
							return
						}
					}
				}
			}
		}
	}
	g.SetMineWarning("No incomplete task in active quest", 120, 2)
}

func (g *Game) CompleteCurrentQuest() {
	if g.questManager != nil {
		for _, cat := range g.questManager.Categories {
			for _, q := range cat.Quests {
				if !q.Completed {
					for _, t := range q.Tasks {
						t.CurrentCount = t.RequiredCount
						t.Completed = true
					}
					q.Completed = true
					g.SetMineWarning("Completed Quest: "+q.Title, 150, 1)
					audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
					return
				}
			}
		}
	}
	g.SetMineWarning("No incomplete quest found", 120, 2)
}

func (g *Game) CompleteAllQuests() {
	if g.questManager != nil {
		for _, cat := range g.questManager.Categories {
			for _, q := range cat.Quests {
				for _, t := range q.Tasks {
					t.CurrentCount = t.RequiredCount
					t.Completed = true
				}
				q.Completed = true
			}
		}
		g.SetMineWarning("ALL Quests Completed!", 180, 1)
		audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
	}
}

func (g *Game) ResetAllQuests() {
	g.questManager = quest.NewQuestManager()
	g.SetMineWarning("Quests Reset to Beginning", 150, 1)
}

func (g *Game) CloseDebugMenu() {
	g.showDebugMenu = false
}
