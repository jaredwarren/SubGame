package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/quest"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Update advances all game logic by one tick.
func (g *Game) Update() error {
	if g.touch != nil {
		g.touch.SetContext(g.touchContext())
		g.touch.SetCanEnterVehicle(g.canEnterVehicleNearby())
		g.touch.SetCanEnterLifePod(g.canEnterLifePodNearby())
		g.touch.SetVehicleCapabilities(g.activeVehicleHasSonar(), g.activeVehicleHasSpecial())
		g.touch.SetHasFlashlightAvailable(g.hasFlashlightAvailable())
		g.touch.SetFlashlightState(g.IsFlashlightOn())
	}
	if ci, ok := g.Input.(*CombinedInput); ok {
		if g.player != nil && g.camera != nil {
			pScreen := gvec.Vec2{
				X: g.player.Pos.X + g.player.Width/2.0 - g.camera.Pos.X,
				Y: g.player.Pos.Y + g.player.Height/2.0 - g.camera.Pos.Y,
			}
			if g.ActiveVehicle != nil {
				vPos := g.ActiveVehicle.GetPos()
				vDims := g.ActiveVehicle.GetDimensions()
				pScreen = gvec.Vec2{
					X: vPos.X + vDims.X/2.0 - g.camera.Pos.X,
					Y: vPos.Y + vDims.Y/2.0 - g.camera.Pos.Y,
				}
			}
			ci.SetAimOrigin(pScreen)
		}
	}
	g.Input.Update()
	audio.Get().Update()
	g.transitionedThisFrame = false
	g.justExited = false
	g.playerSlowed = false

	if g.nextScene != nil {
		g.TransitionTo(g.nextScene)
		g.nextScene = nil
	}

	// Toggle Debug Menu with ` / ~, F1, F3, F12, or Ctrl/Cmd+D
	ctrlPressed := g.Input.IsKeyPressed(ebiten.KeyControl) || g.Input.IsKeyPressed(ebiten.KeyMeta)
	if g.Input.IsKeyJustPressed(ebiten.KeyGraveAccent) || g.Input.IsKeyJustPressed(ebiten.KeyF1) || g.Input.IsKeyJustPressed(ebiten.KeyF3) || g.Input.IsKeyJustPressed(ebiten.KeyF12) || (ctrlPressed && g.Input.IsKeyJustPressed(ebiten.KeyD)) {
		g.showDebugMenu = !g.showDebugMenu
		if g.showDebugMenu {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			g.applyActiveCheats()
			return nil
		}
	}

	if g.showDebugMenu && g.debugMenu != nil {
		if err := g.debugMenu.Update(g); err != nil {
			return err
		}
		g.applyActiveCheats()
		return nil
	}

	if g.currentState == StatePause {
		if !g.transitionedThisFrame {
			if err := g.currentScene.Update(g); err != nil {
				return err
			}
		}
		return nil
	}

	g.advanceTimers()
	g.updateEffects()
	g.updateQuests()
	g.handleWorldTaps()
	g.handleInput()
	if g.currentState == StatePause {
		return nil
	}
	g.baseStation.UpdatePower(g.TimeOfDay)

	// Apply cheat overrides during active gameplay
	g.applyActiveCheats()

	// Inventory screen consumes all clicks; skip normal game logic while open.
	if g.showInventory {
		g.handleInventoryClicks()
		return nil
	}

	g.checkVehicleDepth()

	if g.vehicleRT == nil {
		g.vehicleRT = newVehicleRuntimeAdapter(g)
	}
	vrt := g.vehicleRT
	vrt.cmds = vrt.cmds[:0]
	g.updateActiveVehicle(vrt)
	g.checkVehicleEntry()

	if !g.transitionedThisFrame {
		if err := g.currentScene.Update(g); err != nil {
			return err
		}
	}

	if g.currentState == StateOverworld || g.currentState == StateCave {
		g.updateIdleVehicles(vrt)
		g.drainVehicleCommands(vrt)
		g.updateCamera()
		g.player.UpdateAnimation()

		g.updateLostCargo()
		g.updateAudioAlerts()

		if g.player.CurrentHealth <= 0 && !g.GodMode {
			// Drop cargo once at the moment of death; stay on GameOver without re-dropping.
			audio.Get().PlaySFX("sfx/player_drown.wav")
			g.dropLostCargo()
			g.TransitionTo(g.gameOverState)
		}
	}
	return nil
}

func (g *Game) updateQuests() {
	if g.questManager == nil {
		return
	}
	if g.currentState != StateOverworld && g.currentState != StateCave && g.currentState != StateBaseMenu && g.currentState != StatePause {
		return
	}

	var notifs []quest.ProgressNotification

	// Position-sensitive conditions: cheap polls (no inventory scans).
	switch g.currentState {
	case StateCave:
		notifs = append(notifs, g.questManager.HandleEvent(g, quest.ProgressEvent{
			Kind:  quest.EventDepth,
			Depth: g.MaxDepthReached(),
		})...)
	case StateOverworld:
		notifs = append(notifs, g.questManager.HandleEvent(g, quest.ProgressEvent{
			Kind: quest.EventNearBase,
		})...)
	}

	events := g.pendingQuestEvents
	g.pendingQuestEvents = nil
	for _, ev := range events {
		notifs = append(notifs, g.questManager.HandleEvent(g, ev)...)
	}

	g.applyQuestNotifications(notifs)
}

// EmitQuestEvent queues a progress event for the next updateQuests drain.
func (g *Game) EmitQuestEvent(ev quest.ProgressEvent) {
	if g.questManager == nil {
		return
	}
	g.pendingQuestEvents = append(g.pendingQuestEvents, ev)
}

// NotifyQuestInventoryChanged signals that player inventory gained/lost items.
func (g *Game) NotifyQuestInventoryChanged(id item.ItemID) {
	g.EmitQuestEvent(quest.ProgressEvent{Kind: quest.EventInventory, ItemID: id})
}

// NotifyQuestCrafted signals a successful fabricator craft.
func (g *Game) NotifyQuestCrafted(id item.ItemID) {
	g.EmitQuestEvent(quest.ProgressEvent{Kind: quest.EventCrafted, ItemID: id})
}

// NotifyQuestVehicleDeployed signals a vehicle was placed in the world.
func (g *Game) NotifyQuestVehicleDeployed(id vehicle.VehicleID) {
	g.EmitQuestEvent(quest.ProgressEvent{Kind: quest.EventVehicle, VehicleID: id})
}

func (g *Game) applyQuestNotifications(notifs []quest.ProgressNotification) {
	for _, n := range notifs {
		if n.Completed {
			audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
			if g.MineWarning.Timer <= 0 || g.MineWarning.Level <= 1 {
				msg := n.Message
				if !g.isTouchActive() {
					msg += " (Press [J] for PDA)"
				}
				g.SetMineWarning(msg, 200, 1)
			}
		} else {
			audio.Get().PlaySFX("sfx/ui_hover.wav")
			if g.MineWarning.Timer <= 0 || g.MineWarning.Level <= 1 {
				g.SetMineWarning(n.Message, 160, 1)
			}
		}
	}
}

// updateAudioAlerts handles continuous loops and threshold voice alerts.
func (g *Game) updateAudioAlerts() {
	if g.currentState != StateCave && g.currentState != StateOverworld {
		audio.Get().StopLoop("heartbeat")
		audio.Get().StopLoop("breathing")
		return
	}

	// Critical health heartbeat loop (< 25% health)
	if g.player.MaxHealth > 0 && (g.player.CurrentHealth/g.player.MaxHealth) < 0.25 && g.player.CurrentHealth > 0 {
		audio.Get().PlayLoop("heartbeat", "sfx/heartbeat_loop.wav", 0.7)
	} else {
		audio.Get().StopLoop("heartbeat")
	}

	// In cave oxygen monitoring and voice lines
	if g.currentState == StateCave && g.ActiveVehicle == nil {
		if g.player.MaxOxygen > 0 {
			o2Ratio := g.player.CurrentOxygen / g.player.MaxOxygen
			if o2Ratio <= 0.10 {
				if !g.o2CritAlertPlayed {
					g.o2CritAlertPlayed = true
					audio.Get().PlaySFX("sfx/voice_o2_critical.wav")
				}
			} else if o2Ratio <= 0.30 {
				if !g.o2LowAlertPlayed {
					g.o2LowAlertPlayed = true
					audio.Get().PlaySFX("sfx/voice_o2_low.wav")
				}
			}

			if o2Ratio <= 0.30 && g.player.CurrentOxygen > 0 {
				audio.Get().PlayLoop("breathing", "sfx/heavy_breathing_loop.wav", 0.55)
			} else {
				audio.Get().StopLoop("breathing")
			}
		}
	} else {
		// Reset alert flags when surfacing or inside vehicle
		if g.player.CurrentOxygen >= g.player.MaxOxygen*0.8 {
			g.o2LowAlertPlayed = false
			g.o2CritAlertPlayed = false
		}
		audio.Get().StopLoop("breathing")
	}
}

// advanceTimers increments all per-frame counters and timers.
func (g *Game) advanceTimers() {
	g.Ticks += 1.0
	if !g.FreezeTimeOfDay {
		g.TimeOfDay += 1.0
		if g.TimeOfDay >= 14400 {
			g.TimeOfDay = 0.0
		}
	}
	if g.MineWarning.Timer > 0 {
		g.MineWarning.Timer--
	}
	if g.DamageFlash.Timer > 0 {
		g.DamageFlash.Timer--
	}
	if g.damageFeedbackCooldown > 0 {
		g.damageFeedbackCooldown--
	}
	if g.toasts != nil {
		g.toasts.Update()
	}
}

func (g *Game) applyActiveCheats() {
	if g.GodMode {
		g.player.CurrentHealth = g.player.MaxHealth
	}
	if g.InfiniteOxygen {
		g.player.CurrentOxygen = g.player.MaxOxygen
	}
	if g.InfiniteStamina {
		g.player.CurrentStamina = g.player.MaxStamina
	}
	if g.ActiveVehicle != nil {
		if g.InfiniteVehicleBattery {
			g.ActiveVehicle.RechargeBattery(g.ActiveVehicle.GetMaxBattery())
		}
		if g.InfiniteVehicleHull {
			g.ActiveVehicle.Repair(g.ActiveVehicle.GetMaxHealth())
		}
	}
}

// updateEffects ticks the sonar, sound wave, and particle systems.
func (g *Game) updateEffects() {
	g.Sonar.Update()
	if g.SoundWave.Timer > 0 {
		g.SoundWave.Timer--
		g.SoundWave.Radius += 4.5
	}
	g.Particles = particle.UpdateParticles(g.Particles)
}

// touchContext maps the current game state to the virtual button set to show.
func (g *Game) touchContext() TouchContext {
	switch g.currentState {
	case StateOverworld:
		if g.showDebugMenu {
			return TouchContextHidden
		}
		if g.showInventory {
			return TouchContextInventory
		}
		if g.ActiveVehicle != nil {
			return TouchContextDriving
		}
		return TouchContextOnFoot
	case StateCave:
		if g.showDebugMenu {
			return TouchContextHidden
		}
		if g.showInventory {
			return TouchContextInventory
		}
		if g.ActiveVehicle != nil {
			return TouchContextCaveDriving
		}
		return TouchContextCave
	case StateBaseMenu:
		return TouchContextMenu
	default:
		return TouchContextHidden
	}
}

// handleWorldTaps turns unconsumed touch taps on world objects into actions:
// tapping a nearby vehicle boards it (virtual F) and tapping the lifepod opens
// the terminal (virtual E). Consumed taps never become left-clicks.
func (g *Game) handleWorldTaps() {
	if g.touch == nil {
		return
	}
	tap, ok := g.touch.TapCursor()
	if !ok {
		return
	}
	if g.currentState != StateOverworld && g.currentState != StateCave {
		return
	}
	if g.showInventory || g.showDebugMenu || g.ActiveVehicle != nil {
		return
	}

	wx := tap.X + g.camera.Pos.X
	wy := tap.Y + g.camera.Pos.Y
	const pad = 12.0

	for _, v := range g.getVehiclesForCurrentScene() {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		if wx < vPos.X-pad || wx > vPos.X+vDims.X+pad || wy < vPos.Y-pad || wy > vPos.Y+vDims.Y+pad {
			continue
		}
		g.touch.ConsumeTap()
		dist := math.Hypot(vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0,
			vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0)
		if dist < 60.0 {
			g.touch.InjectJustPressed(ebiten.KeyF)
		} else {
			g.SetMineWarning("Move closer to board the "+v.GetName(), 90, 1)
		}
		return
	}

	if g.currentState == StateOverworld && g.baseStation != nil {
		bPos, bSize := g.baseStation.Pos, g.baseStation.Size
		if wx >= bPos.X-pad && wx <= bPos.X+bSize.X+pad && wy >= bPos.Y-pad && wy <= bPos.Y+bSize.Y+pad {
			g.touch.ConsumeTap()
			if g.baseStation.DistanceToPlayer(g.player) < 100.0 {
				g.touch.InjectJustPressed(ebiten.KeyE)
			} else {
				g.SetMineWarning("Move closer to the Life Pod", 90, 1)
			}
		}
	}
}

// handleInput processes all keyboard input that applies regardless of open panels.
func (g *Game) handleInput() {
	if g.Input.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.showInventory {
			g.showInventory = false
			audio.Get().PlaySFX("sfx/inventory_close.wav")
			return
		}
		if g.currentState == StateOverworld || g.currentState == StateCave {
			g.pauseState.PriorState = g.currentState
			g.TransitionTo(g.pauseState)
			return
		}
	}
	if g.Input.IsKeyJustPressed(ebiten.KeyT) || (g.ActiveVehicle != nil && g.Input.IsKeyJustPressed(ebiten.KeyL)) {
		if g.ActiveVehicle != nil {
			if hv, ok := g.ActiveVehicle.(vehicle.HeadlightVehicle); ok && hv.HasHeadlights() {
				hv.ToggleHeadlights()
				audio.Get().PlaySFX("sfx/flashlight_toggle.wav")
			}
		} else if g.Input.IsKeyJustPressed(ebiten.KeyT) {
			g.FlashlightOn = !g.FlashlightOn
			audio.Get().PlaySFX("sfx/flashlight_toggle.wav")
		}
	}
	if (g.Input.IsKeyJustPressed(ebiten.KeyTab) || g.Input.IsKeyJustPressed(ebiten.KeyI)) && (g.currentState == StateOverworld || g.currentState == StateCave) {
		g.showInventory = !g.showInventory
		if g.showInventory {
			audio.Get().PlaySFX("sfx/inventory_open.wav")
		} else {
			audio.Get().PlaySFX("sfx/inventory_close.wav")
		}
	}
	if g.currentState == StateOverworld && g.Input.IsKeyJustPressed(ebiten.KeyE) {
		if g.baseStation != nil && g.baseStation.DistanceToPlayer(g.player) < 100.0 {
			g.activeMiniLifepod = nil
			g.miniLifepodStation = nil
			audio.Get().PlaySFX("sfx/airlock_cycle.wav")
			g.menuOpenedAnywhere = false
			g.baseMenu.ActiveTab = 1
			g.TransitionTo(g.baseMenu)
		} else if pod := g.getNearbyMiniLifepod(); pod != nil {
			g.activeMiniLifepod = pod
			g.syncMiniLifepodBaseStation(pod)
			audio.Get().PlaySFX("sfx/airlock_cycle.wav")
			g.menuOpenedAnywhere = false
			g.baseMenu.ActiveTab = 1
			g.TransitionTo(g.baseMenu)
		}
	}
	if g.Input.IsKeyJustPressed(ebiten.KeyJ) {
		switch g.currentState {
		case StateBaseMenu:
			if g.menuOpenedAnywhere {
				audio.Get().PlaySFX("sfx/ui_cancel.wav")
				g.ClosePDA()
			}
		case StateOverworld, StateCave:
			audio.Get().PlaySFX("sfx/map_open.wav")
			g.TransitionToPDA()
		}
	}
	ctrlPressed := g.Input.IsKeyPressed(ebiten.KeyControl) || g.Input.IsKeyPressed(ebiten.KeyMeta)
	if !ctrlPressed && g.Input.IsKeyJustPressed(ebiten.KeyM) {
		switch g.currentState {
		case StateBaseMenu:
			if g.menuOpenedAnywhere {
				if g.baseMenu != nil && g.baseMenu.ActiveTab == 6 {
					g.ClosePDA()
				} else if g.baseMenu != nil {
					g.baseMenu.ActiveTab = 6
				}
			} else if g.baseMenu != nil {
				g.baseMenu.ActiveTab = 6
			}
		case StateOverworld, StateCave:
			g.TransitionToMap()
		}
	}

	if !ctrlPressed && (g.currentState == StateOverworld || g.currentState == StateCave) {
		if g.touch != nil {
			if slot, ok := g.touch.ConsumeHotbarTouch(); ok && slot >= 0 {
				g.selectHotbarSlot(slot)
			}
		}
		if g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			cur := g.Input.Cursor()
			if slot := HUDHotbarSlotAt(cur.X, cur.Y); slot >= 0 {
				g.selectHotbarSlot(slot)
				if ci, ok := g.Input.(*CombinedInput); ok {
					ci.ConsumeTap()
				}
			} else if g.ActiveVehicle != nil {
				if hv, ok := g.ActiveVehicle.(vehicle.HeadlightVehicle); ok && hv.HasHeadlights() {
					if HUDVehicleLightButtonHit(cur.X, cur.Y) {
						hv.ToggleHeadlights()
						audio.Get().PlaySFX("sfx/flashlight_toggle.wav")
						if ci, ok := g.Input.(*CombinedInput); ok {
							ci.ConsumeTap()
						}
					}
				}
			}
		}
		if g.Input.IsKeyJustPressed(ebiten.Key1) {
			g.selectHotbarSlot(0)
		} else if g.Input.IsKeyJustPressed(ebiten.Key2) {
			g.selectHotbarSlot(1)
		} else if g.Input.IsKeyJustPressed(ebiten.Key3) {
			g.selectHotbarSlot(2)
		} else if g.Input.IsKeyJustPressed(ebiten.Key4) {
			g.selectHotbarSlot(3)
		} else if g.Input.IsKeyJustPressed(ebiten.Key5) {
			g.selectHotbarSlot(4)
		}
	}

	g.handleDebugInput()
}

// handleDebugInput processes development shortcuts that would be stripped from a release build.
func (g *Game) handleDebugInput() {
	// Shader toggles
	if g.Input.IsKeyJustPressed(ebiten.KeyY) {
		g.DebugDisableLightShader = !g.DebugDisableLightShader
		if g.DebugDisableLightShader {
			g.SetMineWarning("Disabled lighting shader mask", 120, 1)
		} else {
			g.SetMineWarning("Enabled lighting shader mask", 120, 1)
		}
	}
	if g.Input.IsKeyJustPressed(ebiten.KeyU) {
		g.DebugDisableWaterShader = !g.DebugDisableWaterShader
		if g.DebugDisableWaterShader {
			g.SetMineWarning("Disabled water displacement shader", 120, 1)
		} else {
			g.SetMineWarning("Enabled water displacement shader", 120, 1)
		}
	}

	if g.Input.IsKeyJustPressed(ebiten.KeyG) {
		g.TransitionTo(g.gameOverState)
	}

	ctrlPressed := g.Input.IsKeyPressed(ebiten.KeyControl) || g.Input.IsKeyPressed(ebiten.KeyMeta)

	// Reveal every map POI (trenches, wrecks, shock kelp, thermo) so icons show on the chart.
	// Works from overworld, cave, or the PDA map itself.
	if ctrlPressed && g.Input.IsKeyJustPressed(ebiten.KeyM) {
		if g.currentState == StateOverworld || g.currentState == StateCave || g.currentState == StateBaseMenu {
			g.debugRevealAllMapPOIs()
		}
		return
	}

	// Spawn vehicles / fill inventory
	if g.currentState != StateOverworld && g.currentState != StateCave {
		return
	}
	switch {
	case ctrlPressed && g.Input.IsKeyJustPressed(ebiten.Key1):
		g.debugSpawnVehicle(vehicle.NewScoutSub(g.player.Pos.X, g.player.Pos.Y))
	case ctrlPressed && g.Input.IsKeyJustPressed(ebiten.Key2):
		g.debugSpawnVehicle(vehicle.NewHeavyMech(g.player.Pos.X, g.player.Pos.Y))
	case ctrlPressed && g.Input.IsKeyJustPressed(ebiten.Key3):
		g.debugSpawnVehicle(vehicle.NewSkiff(g.player.Pos.X, g.player.Pos.Y))
	case ctrlPressed && g.Input.IsKeyJustPressed(ebiten.Key4):
		g.player.Inventory.AddItem(&item.Titanium{}, 10)
		g.player.Inventory.AddItem(&item.Copper{}, 10)
		g.player.Inventory.AddItem(&item.Quartz{}, 10)
		g.player.Inventory.AddItem(&item.Nickel{}, 10)
		g.player.Inventory.AddItem(&item.Tungsten{}, 10)
		g.player.Inventory.AddItem(&item.AbyssalOre{}, 10)
		g.player.RecalculateUpgrades()
	case ctrlPressed && g.Input.IsKeyJustPressed(ebiten.Key5):
		g.player.CurrentHealth = g.player.MaxHealth
		g.player.CurrentOxygen = g.player.MaxOxygen
		g.player.CurrentStamina = g.player.MaxStamina
	case g.Input.IsKeyJustPressed(ebiten.KeyC):
		g.EnterCave(50, 50)
	}
}

// debugRevealAllMapPOIs charts and marks every special overworld tile so typed PDA
// map icons appear for debugging site placement.
func (g *Game) debugRevealAllMapPOIs() {
	if g.world == nil || g.explorationTracker == nil {
		return
	}

	// Thermo vents are stamped onto the overworld map during extras init.
	if g.overworldState != nil {
		g.overworldState.InitializeExtras(g)
	}

	counts := map[world.TileType]int{}
	total := 0
	for tx := 0; tx < g.world.Width; tx++ {
		for ty := 0; ty < g.world.Height; ty++ {
			tt := g.world.OverworldMap[tx][ty]
			switch tt {
			case world.TileTrench, world.TileWreckage, world.TileShockKelpCave, world.TileThermoCave:
				// Small charted blotch so icons aren't floating on pure fog.
				g.explorationTracker.Reveal(tx, ty, 2)
				g.explorationTracker.MarkVisited(tx, ty, tt)
				counts[tt]++
				total++
			}
		}
	}

	if g.baseMenu != nil {
		g.baseMenu.ResetMapCache()
	}

	g.SetMineWarning(fmt.Sprintf(
		"DEBUG MAP: %d POIs (T%d W%d K%d V%d) — open [M]",
		total,
		counts[world.TileTrench],
		counts[world.TileWreckage],
		counts[world.TileShockKelpCave],
		counts[world.TileThermoCave],
	), 240, 1)

	// Jump straight to the chart so icons are visible immediately.
	if g.currentState != StateBaseMenu || g.baseMenu == nil || g.baseMenu.ActiveTab != 6 {
		g.TransitionToMap()
	} else if g.baseMenu != nil {
		g.baseMenu.ActiveTab = 6
	}
}

func (g *Game) debugSpawnVehicle(v vehicle.Vehicle) {
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
}

// handleInventoryClicks routes left-click events to the correct inventory panel handler.
func (g *Game) handleInventoryClicks() {
	if !g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	if g.ActiveVehicle != nil {
		g.hud.HandleVehicleInventoryClicks(g)
	} else {
		g.hud.HandlePlayerInventoryClicks(g)
	}
}

// TransferToVehicle tries to move an item from the player's hand to the active vehicle,
// preferring power-cell recharging → upgrade slot → cargo in that order.
func (g *Game) TransferToVehicle(it item.Item) {
	v := g.ActiveVehicle
	if v == nil {
		return
	}

	removeItem := func() {
		if !g.player.Inventory.Remove(it, 1) && g.player.Hotbar != nil {
			g.player.Hotbar.Remove(it, 1)
		}
		g.player.RecalculateUpgrades()
	}

	if _, isPowerCell := it.(*item.PowerCell); isPowerCell {
		if v.GetBattery() < v.GetMaxBattery() {
			v.RechargeBattery(100.0)
			removeItem()
			return
		}
	}

	if vUpg := v.GetUpgrades(); vUpg != nil {
		if vUpgItem, ok := it.(item.VehicleUpgradeItem); ok && vUpgItem.IsVehicleUpgrade() {
			if vUpg.AddItem(item.Clone(it), 1) {
				removeItem()
				return
			}
		}
	}

	if v.GetCargo() != nil && v.GetCargo().AddItem(item.Clone(it), 1) {
		removeItem()
	}
}

// ActivatePlayerItem applies the appropriate action for clicking an item in the player inventory.
func (g *Game) ActivatePlayerItem(it item.Item) {
	if it == nil {
		return
	}
	if g.player.EquipUpgrade(it) {
		g.player.Inventory.Remove(it, 1)
		g.player.RecalculateUpgrades()
		return
	}
	if consumable, ok := it.(item.Consumable); ok {
		g.player.CurrentHealth = min(g.player.CurrentHealth+consumable.GetHealthRestore(), g.player.MaxHealth)
		g.player.CurrentStamina = min(g.player.CurrentStamina+consumable.GetStaminaRestore(), g.player.MaxStamina)
		if g.player.Hotbar == nil || !g.player.Hotbar.Remove(it, 1) {
			g.player.Inventory.Remove(it, 1)
		}
		g.player.RecalculateUpgrades()
		audio.Get().PlaySFX("sfx/item_pickup.wav")
		g.SetMineWarning("Ate "+consumable.GetName()+"!", 90, 1)
		return
	}
	if g.currentState != StateCave && g.currentState != StateOverworld {
		return
	}

	if deployable, ok := it.(vehicle.Deployable); ok {
		deployX, deployY := g.player.Pos.X, g.player.Pos.Y
		veh := deployable.Deploy(deployX, deployY)
		if g.currentState == StateOverworld {
			if veh.GetPerspective() != "overworld" {
				g.SetMineWarning(fmt.Sprintf("Cannot deploy %s on the surface! Deploy inside a cave trench.", veh.GetName()), 120, 2)
				audio.Get().PlaySFX("sfx/ui_error.wav")
				return
			}
			// Surface vehicles (e.g. Skiff) must land on clear water, not land/lifepod.
			veh.SetPos(g.findNearestClearWaterDeployPos(g.player.Pos, veh.GetDimensions()))
			g.OverworldVehicles = append(g.OverworldVehicles, veh)
		} else {
			if veh.GetPerspective() != "cave" {
				g.SetMineWarning(fmt.Sprintf("Cannot deploy %s in caves! Deploy on the surface.", veh.GetName()), 120, 2)
				audio.Get().PlaySFX("sfx/ui_error.wav")
				return
			}
			g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], veh)
		}
		if !g.player.Inventory.Remove(it, 1) {
			g.player.Hotbar.Remove(it, 1)
		}
		audio.Get().PlaySFX("sfx/vehicle_exit.wav")
		g.player.RecalculateUpgrades()
		g.showInventory = false
		g.NotifyQuestVehicleDeployed(veh.GetID())
	}
}

// UseRepairTool uses 1 Scrap Metal to repair +25 HP to the nearest vehicle in range.
func (g *Game) UseRepairTool() {
	if g.ActiveVehicle != nil {
		g.SetMineWarning("Exit the vehicle to repair its hull", 90, 1)
		return
	}
	candidates := g.getVehiclesForCurrentScene()
	if len(candidates) == 0 {
		g.SetMineWarning("No vehicle nearby to repair", 90, 1)
		return
	}

	pX := g.player.Pos.X + g.player.Width/2.0
	pY := g.player.Pos.Y + g.player.Height/2.0

	var closestVehicle vehicle.Vehicle
	closestDist := math.MaxFloat64
	for _, v := range candidates {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		vCenterX := vPos.X + vDims.X/2.0
		vCenterY := vPos.Y + vDims.Y/2.0
		dist := math.Hypot(vCenterX-pX, vCenterY-pY)
		if dist < closestDist {
			closestDist = dist
			closestVehicle = v
		}
	}

	if closestVehicle == nil || closestDist > 90.0 {
		g.SetMineWarning("No vehicle nearby to repair", 90, 1)
		return
	}

	if closestVehicle.GetHealth() >= closestVehicle.GetMaxHealth() {
		g.SetMineWarning(closestVehicle.GetName()+" hull is already at maximum integrity!", 90, 1)
		return
	}

	hasScrapInInv := item.HasItem[*item.ScrapMetal](g.player.Inventory, 1)
	hasScrapInHotbar := item.HasItem[*item.ScrapMetal](g.player.Hotbar, 1)
	if !hasScrapInInv && !hasScrapInHotbar {
		g.SetMineWarning("Requires Scrap Metal to repair!", 120, 2)
		return
	}

	scrap := &item.ScrapMetal{}
	if !g.player.Inventory.Remove(scrap, 1) {
		g.player.Hotbar.Remove(scrap, 1)
	}
	g.player.RecalculateUpgrades()

	closestVehicle.Repair(25.0)

	vPos := closestVehicle.GetPos()
	vDims := closestVehicle.GetDimensions()
	sparkX := (pX + vPos.X + vDims.X/2.0) / 2.0
	sparkY := (pY + vPos.Y + vDims.Y/2.0) / 2.0
	sparkColor := color.RGBA{255, 220, 80, 255}
	g.SpawnDebris(sparkX, sparkY, sparkColor)

	g.SetMineWarning(fmt.Sprintf("Repaired %s (+25 HP) [%.0f/%.0f]", closestVehicle.GetName(), closestVehicle.GetHealth(), closestVehicle.GetMaxHealth()), 120, 1)
}

// checkVehicleDepth applies crush damage when a cave vehicle exceeds its depth limit,
// and destroys the vehicle if its hull reaches zero.
func (g *Game) checkVehicleDepth() {
	if g.currentState != StateCave || g.ActiveVehicle == nil {
		return
	}
	limit := g.ActiveVehicle.GetDepthLimit()
	if limit <= 0 {
		return
	}
	vPos := g.ActiveVehicle.GetPos()
	vDims := g.ActiveVehicle.GetDimensions()
	depth := (vPos.Y + vDims.Y/2.0) / config.TileSize
	if !g.caveState.IsShallow {
		trenchX, trenchY := g.activeTrenchX, g.activeTrenchY
		if trenchX >= 0 && trenchX < g.world.Width && trenchY >= 0 && trenchY < g.world.Height {
			tt := g.world.OverworldMap[trenchX][trenchY]
			if info := world.GetTileInfo(tt); info != nil && info.Subterranean != nil {
				depth += 34.0
			}
		}
	}
	if depth < 0 {
		depth = 0
	}

	if depth > limit {
		g.ActiveVehicle.TakeDamage(0.08)
		g.SetMineWarning("WARNING: EXCEEDING MAXIMUM HULL DEPTH LIMIT!", 2, 2)
	}
	if g.ActiveVehicle.GetHealth() > 0 {
		return
	}
	// Hull failure
	g.player.CurrentHealth -= 40.0
	g.SetMineWarning("VEHICLE CRUSHED BY DEEP-SEA PRESSURE!", 180, 3)
	list := g.CaveVehicles[g.activeTrenchKey]
	removeVehicleFromList(&list, g.ActiveVehicle)
	g.CaveVehicles[g.activeTrenchKey] = list
	g.ActiveVehicle = nil
}

// updateActiveVehicle ticks the player-piloted vehicle, syncs player position inside it,
// and handles the exit-vehicle keybind.
func (g *Game) updateActiveVehicle(vrt *vehicleRuntimeAdapter) {
	if g.ActiveVehicle == nil {
		return
	}
	g.ActiveVehicle.Update(vrt)

	vPos := g.ActiveVehicle.GetPos()
	vDims := g.ActiveVehicle.GetDimensions()
	g.player.Pos.X = vPos.X + (vDims.X-g.player.Width)/2.0
	g.player.Pos.Y = vPos.Y + (vDims.Y-g.player.Height)/2.0
	g.player.Vel = gvec.Vec2{}

	if g.ActiveVehicle.GetOxygen() > 0 {
		g.player.CurrentOxygen = g.player.MaxOxygen
	}
	// Regenerate stamina and update player stats while resting/piloting a vehicle
	g.player.UpdateStats(g.currentState == StateCave && g.ActiveVehicle.GetOxygen() <= 0, false)

	if g.Input.IsKeyJustPressed(ebiten.KeyF) {
		g.exitVehicle(vPos, vDims)
	}
}

// exitVehicle ejects the player from the active vehicle, finding a safe position.
func (g *Game) exitVehicle(vPos, vDims gvec.Vec2) {
	safeX, safeY := vPos.X, vPos.Y
	if g.currentState == StateCave && g.caveState != nil {
		switch {
		case !g.caveState.IsSolid(g, vPos.X-32, vPos.Y, g.player.Width, g.player.Height):
			safeX = vPos.X - 32
		case !g.caveState.IsSolid(g, vPos.X+vDims.X+12, vPos.Y, g.player.Width, g.player.Height):
			safeX = vPos.X + vDims.X + 12
		case !g.caveState.IsSolid(g, vPos.X, vPos.Y-32, g.player.Width, g.player.Height):
			safeY = vPos.Y - 32
		case !g.caveState.IsSolid(g, vPos.X, vPos.Y+vDims.Y+12, g.player.Width, g.player.Height):
			safeY = vPos.Y + vDims.Y + 12
		}
		g.player.Pos.X = safeX
		g.player.Pos.Y = safeY
	} else if g.currentState == StateOverworld && g.overworldState != nil {
		facing := 0.0
		if g.ActiveVehicle != nil {
			facing = g.ActiveVehicle.GetFacing()
		}
		safePos := g.overworldState.FindSafeExitPosition(vPos, vDims, facing, g.player.Width, g.player.Height, g.baseStation)
		g.player.Pos = safePos
	} else {
		g.player.Pos.X = vPos.X - 24
	}
	g.player.Vel = gvec.Vec2{}
	g.ActiveVehicle = nil
	g.justExited = true
}

// checkVehicleEntry lets the player board a nearby vehicle with [F].
func (g *Game) getVehiclesForCurrentScene() []vehicle.Vehicle {
	switch g.currentState {
	case StateOverworld:
		return g.OverworldVehicles
	case StateCave:
		return g.CaveVehicles[g.activeTrenchKey]
	default:
		return nil
	}
}

// checkVehicleEntry lets the player board a nearby vehicle with [F].
func (g *Game) checkVehicleEntry() {
	if g.ActiveVehicle != nil || g.justExited {
		return
	}
	if !g.Input.IsKeyJustPressed(ebiten.KeyF) {
		return
	}
	candidates := g.getVehiclesForCurrentScene()
	for _, v := range candidates {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		dist := math.Hypot(vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0,
			vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0)
		if dist < 60.0 {
			g.ActiveVehicle = v
			return
		}
	}
}

// canEnterVehicleNearby reports whether an idle vehicle in the current scene is
// within boarding distance (< 60 units) of the player.
func (g *Game) canEnterVehicleNearby() bool {
	if g.ActiveVehicle != nil || g.justExited || g.player == nil {
		return false
	}
	candidates := g.getVehiclesForCurrentScene()
	for _, v := range candidates {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		dist := math.Hypot(vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0,
			vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0)
		if dist < 60.0 {
			return true
		}
	}
	return false
}

// canEnterLifePodNearby reports whether the player is on foot in the overworld and
// within interaction distance of the Life Pod or a deployed Mini-Lifepod.
func (g *Game) canEnterLifePodNearby() bool {
	if g.currentState != StateOverworld || g.ActiveVehicle != nil || g.player == nil {
		return false
	}
	if g.baseStation != nil && g.baseStation.DistanceToPlayer(g.player) < 100.0 {
		return true
	}
	return g.getNearbyMiniLifepod() != nil
}

func (g *Game) getNearbyMiniLifepod() *vehicle.MiniLifepod {
	if g.currentState != StateOverworld || g.player == nil {
		return nil
	}
	px := g.player.Pos.X + g.player.Width/2.0
	py := g.player.Pos.Y + g.player.Height/2.0
	for _, v := range g.OverworldVehicles {
		if pod, ok := v.(*vehicle.MiniLifepod); ok && pod != nil {
			vx := pod.Pos.X + pod.Dimensions.X/2.0
			vy := pod.Pos.Y + pod.Dimensions.Y/2.0
			if math.Hypot(px-vx, py-vy) < 80.0 {
				return pod
			}
		}
	}
	return nil
}

func (g *Game) syncMiniLifepodBaseStation(pod *vehicle.MiniLifepod) {
	if pod == nil {
		g.miniLifepodStation = nil
		return
	}
	pod.RecalculateProperties()
	st := &base.BaseStation{
		Pos:               pod.Pos,
		Size:              pod.Dimensions,
		Power:             pod.Battery,
		MaxPower:          pod.MaxBattery,
		Storage:           pod.Cargo,
		Upgrades:          pod.Upgrades,
		SolarRechargeRate: pod.SolarRechargeRate,
		ActiveModules:     pod.ActiveModules,
	}
	g.miniLifepodStation = st
}

// activeVehicleHasSonar reports whether the vehicle currently being piloted has a
// functional sonar upgrade installed.
func (g *Game) activeVehicleHasSonar() bool {
	if g.ActiveVehicle == nil {
		return false
	}
	upg := g.ActiveVehicle.GetUpgrades()
	if upg == nil {
		return false
	}
	switch g.ActiveVehicle.GetID() {
	case vehicle.VehicleSkiff:
		return item.HasItem[*item.SurfaceSonar](upg, 1)
	case vehicle.VehicleScoutSub:
		return item.HasItem[*item.SonarAmplifier](upg, 1)
	default:
		return false
	}
}

// activeVehicleHasSpecial reports whether the vehicle currently being piloted has a
// special action upgrade installed (decoy launcher or chemical discharger).
func (g *Game) activeVehicleHasSpecial() bool {
	if g.ActiveVehicle == nil {
		return false
	}
	upg := g.ActiveVehicle.GetUpgrades()
	if upg == nil {
		return false
	}
	switch g.ActiveVehicle.GetID() {
	case vehicle.VehicleScoutSub:
		return item.HasItem[*item.DecoyLauncher](upg, 1) || item.HasItem[*item.ChemicalDischarger](upg, 1)
	case vehicle.VehicleHeavyMech:
		return item.HasItem[*item.DecoyLauncher](upg, 1) || item.HasItem[*item.ChemicalDischarger](upg, 1)
	default:
		return false
	}
}

// hasFlashlightAvailable reports whether the player is currently piloting a vehicle
// (which has headlights) or is on foot actively holding a Flashlight tool.
func (g *Game) hasFlashlightAvailable() bool {
	if g.ActiveVehicle != nil {
		if hv, ok := g.ActiveVehicle.(vehicle.HeadlightVehicle); ok {
			return hv.HasHeadlights()
		}
		return true
	}
	if g.player == nil {
		return false
	}
	_, ok := g.player.GetActiveItem().(*item.Flashlight)
	return ok
}

// selectHotbarSlot selects a hotbar slot index. If the slot is already selected and
// contains a consumable, it consumes the item.
func (g *Game) selectHotbarSlot(slot int) {
	if g.player == nil || g.player.Hotbar == nil || slot < 0 || slot >= len(g.player.Hotbar.Slots) {
		return
	}
	prevSlot := g.player.ActiveSlot
	if prevSlot == slot {
		if consumable, ok := g.player.Hotbar.Slots[slot].Item.(item.Consumable); ok {
			g.ActivatePlayerItem(consumable)
			return
		}
	}
	g.player.ActiveSlot = slot
	audio.Get().PlaySFX("sfx/hotbar_switch.wav")
	if _, ok := g.player.GetActiveItem().(*item.Flashlight); ok {
		g.FlashlightOn = true
	}
}

// updateIdleVehicles ticks all vehicles that the player is not currently piloting.
func (g *Game) updateIdleVehicles(vrt *vehicleRuntimeAdapter) {
	idle := g.getVehiclesForCurrentScene()
	for _, v := range idle {
		if v != g.ActiveVehicle {
			v.Update(vrt)
		}
	}
	// In cave state, the surface Skiff is not in the cave scene list, but continues
	// solar charging and charging its docked vehicles in the background.
	if g.currentState == StateCave {
		if skiff := g.GetSkiff(); skiff != nil && skiff != g.ActiveVehicle {
			skiff.Update(vrt)
		}
	}
}

// drainVehicleCommands applies all fire-and-forget mutations queued by vehicles this tick.
func (g *Game) drainVehicleCommands(rt *vehicleRuntimeAdapter) {
	for _, cmd := range rt.cmds {
		switch c := cmd.(type) {
		case vehicle.ActivateSonarCmd:
			g.Sonar.Activate(c)
			audio.Get().PlaySFX("sfx/sub_sonar_ping.wav")
		case vehicle.ActivateSurfaceSonarCmd:
			g.Sonar.Activate(vehicle.ActivateSonarCmd{
				Source: c.Source,
				Pulse:  c.Pulse,
				Bright: true,
			})
			audio.Get().PlaySFX("sfx/sub_sonar_ping.wav")

			tx := int(c.Source.X) / config.TileSize
			ty := int(c.Source.Y) / config.TileSize
			if g.explorationTracker != nil {
				g.explorationTracker.Reveal(tx, ty, c.FogRevealRadius)
			}

			detectedCount := 0
			if g.world != nil && g.explorationTracker != nil {
				r2 := c.POIDetectionRadius * c.POIDetectionRadius
				minX := max(0, tx-c.POIDetectionRadius)
				maxX := min(g.world.Width-1, tx+c.POIDetectionRadius)
				minY := max(0, ty-c.POIDetectionRadius)
				maxY := min(g.world.Height-1, ty+c.POIDetectionRadius)

				for py := minY; py <= maxY; py++ {
					dy := py - ty
					for px := minX; px <= maxX; px++ {
						dx := px - tx
						if dx*dx+dy*dy > r2 {
							continue
						}
						tt := g.world.OverworldMap[px][py]
						switch tt {
						case world.TileShockKelpCave, world.TileThermoCave, world.TileTrench, world.TileWreckage:
							if !g.explorationTracker.IsVisited(px, py) {
								g.explorationTracker.MarkPOIDiscovered(px, py)
								detectedCount++
							}
						}
					}
				}
			}

			if detectedCount > 0 {
				if detectedCount == 1 {
					g.SetMineWarning("Sector Scanned: 1 site located", 180, 1)
				} else {
					g.SetMineWarning(fmt.Sprintf("Sector Scanned: %d sites located", detectedCount), 180, 1)
				}
			} else {
				g.SetMineWarning("Sector Scanned: No sites in range", 180, 1)
			}
		case vehicle.RemoveCaveNodeCmd:
			nodes := g.caveState.Nodes
			for i, node := range nodes {
				tx, ty := node.GetTilePos()
				if tx == c.TX && ty == c.TY {
					g.caveState.Nodes = append(nodes[:i], nodes[i+1:]...)
					break
				}
			}
		case vehicle.UnlockRecipeCmd:
			recipes := g.GetCraftingRecipes()
			for idx := range recipes {
				if recipes[idx].NewResult().GetName() == c.RecipeResultName {
					recipes[idx].Unlocked = true
					g.SetMineWarning("Unlocked: "+c.RecipeResultName+"!", 120, 1)
					audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
					break
				}
			}
		case vehicle.SpawnBubbleCmd:
			g.Particles = append(g.Particles, particle.NewBubbleParticle(c.Pos.X, c.Pos.Y))
		case vehicle.SpawnDebrisCmd:
			g.Particles = append(g.Particles, particle.NewDebrisParticles(c.Pos.X, c.Pos.Y, c.Color)...)
		case vehicle.TriggerShakeCmd:
			g.TriggerScreenShake(c.Duration, c.Intensity)
		case vehicle.SetWarningCmd:
			g.SetMineWarning(c.Message, c.Duration, c.Level)
		case vehicle.SpawnDecoyCmd:
			decoy := entity.NewSonicDecoy(c.Pos.X, c.Pos.Y, c.Vel)
			g.caveState.Entities = append(g.caveState.Entities, decoy)
			g.SetCaveEntities(g.GetActiveTrenchKey(), g.caveState.Entities)
			audio.Get().PlaySFX("sfx/decoy_launch.wav")
		case vehicle.SpawnDeterrentCloudCmd:
			cloud := entity.NewDeterrentCloud(c.Pos.X, c.Pos.Y)
			g.caveState.Entities = append(g.caveState.Entities, cloud)
			g.SetCaveEntities(g.GetActiveTrenchKey(), g.caveState.Entities)
			audio.Get().PlaySFX("sfx/deterrent_disperse.wav")
		case vehicle.AddItemToastCmd:
			g.AddItemToast(c.Item, c.Quantity)
		}
	}
	rt.cmds = rt.cmds[:0]
}

// updateCamera smoothly tracks the player and applies screen shake effects.
func (g *Game) updateCamera() {
	if g.currentState != StateOverworld && g.currentState != StateCave {
		return
	}
	g.camera.Track(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height, 0.08)

	if g.currentState == StateCave && g.caveState.CaveGrid != nil {
		caveW := len(g.caveState.CaveGrid)
		maxCamX := float64(caveW*config.TileSize - config.ScreenWidth)
		if g.camera.Pos.X < 0 {
			g.camera.Pos.X = 0
		}
		if g.camera.Pos.X > maxCamX {
			g.camera.Pos.X = maxCamX
		}
	}

	if g.currentState == StateCave && g.WeaverTrackingTimer > 0 {
		shakeMag := (g.WeaverTrackingTimer / 300.0) * 8.0
		g.camera.Pos.X += rand.Float64()*shakeMag - shakeMag/2.0
		g.camera.Pos.Y += rand.Float64()*shakeMag - shakeMag/2.0
	}
	if g.Shake.Duration > 0 {
		g.camera.Pos.X += rand.Float64()*g.Shake.Intensity - g.Shake.Intensity/2.0
		g.camera.Pos.Y += rand.Float64()*g.Shake.Intensity - g.Shake.Intensity/2.0
		g.Shake.Duration--
	}
}

// --- small helpers ---



func (g *Game) PickUpActiveVehicle() {
	v := g.ActiveVehicle
	if v == nil {
		return
	}
	kit := v.GetKit()
	if kit == nil {
		return
	}

	if !g.player.Inventory.AddItem(kit, 1) {
		g.SetMineWarning("Inventory full! Cannot pick up vehicle.", 120, 2)
		return
	}

	var cargoStacks []item.ItemStack
	if vCargo := v.GetCargo(); vCargo != nil {
		cargoStacks = vCargo.ExtractAll()
	}

	hadOverflow := false
	if len(cargoStacks) > 0 {
		leftover := g.player.Inventory.InsertStacks(cargoStacks)
		if len(leftover) > 0 {
			leftover = g.player.Hotbar.InsertStacks(leftover)
		}
		if len(leftover) > 0 {
			hadOverflow = true
			var dropPos gvec.Vec2
			if g.currentState == StateCave {
				dropPos = gvec.Vec2{
					X: float64(g.activeTrenchX * config.TileSize),
					Y: float64(g.activeTrenchY * config.TileSize),
				}
				if dropPos.X == 0 && dropPos.Y == 0 {
					dropPos = gvec.Vec2{X: g.lastOverworldX, Y: g.lastOverworldY}
				}
			} else {
				dropPos = g.player.Pos
			}
			beacon := entity.NewLostCargoBeacon(dropPos, leftover)
			g.lostCargo = append(g.lostCargo, beacon)
		}
	}

	g.player.RecalculateUpgrades()
	g.removeVehicle(v)
	g.ActiveVehicle = nil
	g.showInventory = false
	g.AddItemToast(kit, 1)

	if hadOverflow {
		g.SetMineWarning("Picked up "+v.GetName()+"! Excess cargo dropped in cargo crate.", 150, 1)
	}
}

func (g *Game) removeVehicle(v vehicle.Vehicle) {
	if g.currentState == StateOverworld {
		removeVehicleFromList(&g.OverworldVehicles, v)
		return
	}
	list := g.CaveVehicles[g.activeTrenchKey]
	removeVehicleFromList(&list, v)
	g.CaveVehicles[g.activeTrenchKey] = list
}
