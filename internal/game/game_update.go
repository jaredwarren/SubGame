package game

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Update advances all game logic by one tick.
func (g *Game) Update() error {
	g.Input.Update()
	g.transitionedThisFrame = false
	g.justExited = false
	g.playerSlowed = false

	if g.nextScene != nil {
		g.TransitionTo(g.nextScene)
		g.nextScene = nil
	}

	g.advanceTimers()
	g.updateEffects()
	g.handleInput()
	g.baseStation.UpdatePower(g.TimeOfDay)

	// Inventory screen consumes all clicks; skip normal game logic while open.
	if g.showInventory {
		g.handleInventoryClicks()
		return nil
	}

	g.checkVehicleDepth()

	vrt := &vehicleRuntimeAdapter{g: g}
	g.updateActiveVehicle(vrt)
	g.checkVehicleEntry()

	if !g.transitionedThisFrame {
		if err := g.currentScene.Update(g); err != nil {
			return err
		}
	}

	g.updateIdleVehicles(vrt)
	g.drainVehicleCommands(vrt)
	g.updateCamera()

	if g.ActiveVehicle == nil {
		g.player.UpdateStats(g.currentState == StateCave, g.Input.IsKeyPressed(ebiten.KeyShift))
	}
	g.player.UpdateAnimation()

	if g.player.CurrentHealth <= 0 {
		g.TransitionTo(g.gameOverState)
	}

	if g.TutorialActive {
		if g.hasSkiffInWorld() {
			g.TutorialActive = false
			g.SetMineWarning("TUTORIAL COMPLETE!", 180, 1)
		}
	}
	return nil
}

// advanceTimers increments all per-frame counters and timers.
func (g *Game) advanceTimers() {
	g.Ticks += 1.0
	g.TimeOfDay += 1.0
	if g.TimeOfDay >= 14400 {
		g.TimeOfDay = 0.0
	}
	if g.MineWarning.Timer > 0 {
		g.MineWarning.Timer--
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

// handleInput processes all keyboard input that applies regardless of open panels.
func (g *Game) handleInput() {
	if g.Input.IsKeyJustPressed(ebiten.KeyT) {
		g.FlashlightOn = !g.FlashlightOn
	}
	if g.Input.IsKeyJustPressed(ebiten.KeyTab) && (g.currentState == StateOverworld || g.currentState == StateCave) {
		g.showInventory = !g.showInventory
	}
	if g.currentState == StateOverworld && g.baseStation.DistanceToPlayer(g.player) < 100.0 && g.Input.IsKeyJustPressed(ebiten.KeyE) {
		g.TransitionTo(g.baseMenu)
	}
	if g.Input.IsKeyJustPressed(ebiten.KeyJ) {
		switch g.currentState {
		case StateBaseMenu:
			g.ClosePDA()
		case StateOverworld, StateCave:
			g.TransitionToPDA()
		}
	}
	
	ctrlPressed := g.Input.IsKeyPressed(ebiten.KeyControl) || g.Input.IsKeyPressed(ebiten.KeyMeta)
	if !ctrlPressed && (g.currentState == StateOverworld || g.currentState == StateCave) {
		if g.Input.IsKeyJustPressed(ebiten.Key1) {
			g.player.ActiveSlot = 0
		} else if g.Input.IsKeyJustPressed(ebiten.Key2) {
			g.player.ActiveSlot = 1
		} else if g.Input.IsKeyJustPressed(ebiten.Key3) {
			g.player.ActiveSlot = 2
		} else if g.Input.IsKeyJustPressed(ebiten.Key4) {
			g.player.ActiveSlot = 3
		} else if g.Input.IsKeyJustPressed(ebiten.Key5) {
			g.player.ActiveSlot = 4
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

	// Spawn vehicles / fill inventory
	if g.currentState != StateOverworld && g.currentState != StateCave {
		return
	}
	ctrlPressed := g.Input.IsKeyPressed(ebiten.KeyControl) || g.Input.IsKeyPressed(ebiten.KeyMeta)
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

func (g *Game) debugSpawnVehicle(v vehicle.Vehicle) {
	if g.currentState == StateOverworld {
		g.OverworldVehicles = append(g.OverworldVehicles, v)
	} else {
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

	if _, isPowerCell := it.(*item.PowerCell); isPowerCell {
		if v.GetBattery() < v.GetMaxBattery() {
			v.RechargeBattery(100.0)
			g.player.Inventory.Remove(it, 1)
			g.player.RecalculateUpgrades()
			return
		}
	}

	if vUpg := v.GetUpgrades(); vUpg != nil {
		if _, ok := it.(item.VehicleUpgradeItem); ok {
			if vUpg.AddItem(item.Clone(it), 1) {
				g.player.Inventory.Remove(it, 1)
				g.player.RecalculateUpgrades()
				return
			}
		}
	}

	if v.GetCargo().AddItem(item.Clone(it), 1) {
		g.player.Inventory.Remove(it, 1)
		g.player.RecalculateUpgrades()
	}
}

// ActivatePlayerItem applies the appropriate action for clicking an item in the player inventory.
func (g *Game) ActivatePlayerItem(it item.Item) {
	if g.player.EquipUpgrade(it) {
		g.player.Inventory.Remove(it, 1)
		g.player.RecalculateUpgrades()
		return
	}
	if consumable, ok := it.(item.Consumable); ok {
		g.player.CurrentHealth = min(g.player.CurrentHealth+consumable.GetHealthRestore(), g.player.MaxHealth)
		g.player.CurrentStamina = min(g.player.CurrentStamina+consumable.GetStaminaRestore(), g.player.MaxStamina)
		g.player.Inventory.Remove(it, 1)
		g.SetMineWarning("Ate "+consumable.GetName()+"!", 90, 1)
		return
	}
	if g.currentState != StateCave && g.currentState != StateOverworld {
		return
	}

	if deployable, ok := it.(vehicle.Deployable); ok {
		veh := deployable.Deploy(g.player.Pos.X, g.player.Pos.Y)
		if g.currentState == StateOverworld {
			g.OverworldVehicles = append(g.OverworldVehicles, veh)
		} else {
			g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], veh)
		}
		if !g.player.Inventory.Remove(it, 1) {
			g.player.Hotbar.Remove(it, 1)
		}
		g.player.RecalculateUpgrades()
		g.showInventory = false
	}
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
	for i, v := range list {
		if v == g.ActiveVehicle {
			g.CaveVehicles[g.activeTrenchKey] = append(list[:i], list[i+1:]...)
			break
		}
	}
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

	if g.Input.IsKeyJustPressed(ebiten.KeyF) {
		g.exitVehicle(vPos, vDims)
	}
}

// exitVehicle ejects the player from the active vehicle, finding a safe position.
func (g *Game) exitVehicle(vPos, vDims gvec.Vec2) {
	safeX, safeY := vPos.X, vPos.Y
	if g.currentState == StateCave {
		switch {
		case !g.caveState.IsSolid(g, vPos.X-32, vPos.Y, g.player.Width, g.player.Height):
			safeX = vPos.X - 32
		case !g.caveState.IsSolid(g, vPos.X+vDims.X+12, vPos.Y, g.player.Width, g.player.Height):
			safeX = vPos.X + vDims.X + 12
		case !g.caveState.IsSolid(g, vPos.X, vPos.Y-32, g.player.Width, g.player.Height):
			safeY = vPos.Y - 32
		}
		g.player.Pos.X = safeX
		g.player.Pos.Y = safeY
	} else {
		g.player.Pos.X = vPos.X - 24
	}
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

// updateIdleVehicles ticks all vehicles that the player is not currently piloting.
func (g *Game) updateIdleVehicles(vrt *vehicleRuntimeAdapter) {
	idle := g.getVehiclesForCurrentScene()
	for _, v := range idle {
		if v != g.ActiveVehicle {
			v.Update(vrt)
		}
	}
}

// drainVehicleCommands applies all fire-and-forget mutations queued by vehicles this tick.
func (g *Game) drainVehicleCommands(rt *vehicleRuntimeAdapter) {
	for _, cmd := range rt.cmds {
		switch c := cmd.(type) {
		case vehicle.ActivateSonarCmd:
			g.Sonar.Activate(c)
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
		case vehicle.SpawnDeterrentCloudCmd:
			cloud := entity.NewDeterrentCloud(c.Pos.X, c.Pos.Y)
			g.caveState.Entities = append(g.caveState.Entities, cloud)
			g.SetCaveEntities(g.GetActiveTrenchKey(), g.caveState.Entities)
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
	if (v.GetCargo() != nil && !v.GetCargo().IsEmpty()) || (v.GetUpgrades() != nil && !v.GetUpgrades().IsEmpty()) {
		g.SetMineWarning("Vehicle cargo and upgrades must be empty to pick up!", 120, 2)
		return
	}
	if g.player.Inventory.AddItem(kit, 1) {
		g.removeVehicle(v)
		g.ActiveVehicle = nil
		g.showInventory = false
		g.SetMineWarning("Picked up "+v.GetName()+"!", 120, 1)
	} else {
		g.SetMineWarning("Inventory full! Cannot pick up vehicle.", 120, 2)
	}
}

func (g *Game) removeVehicle(v vehicle.Vehicle) {
	if g.currentState == StateOverworld {
		for i, ov := range g.OverworldVehicles {
			if ov == v {
				g.OverworldVehicles = append(g.OverworldVehicles[:i], g.OverworldVehicles[i+1:]...)
				break
			}
		}
	} else {
		list := g.CaveVehicles[g.activeTrenchKey]
		for i, cv := range list {
			if cv == v {
				g.CaveVehicles[g.activeTrenchKey] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}
}
