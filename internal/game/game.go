package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Game implements the ebiten.Game interface and coordinates scenes.
type Game struct {
	currentState          State // Enum tracking for compatibility
	currentScene          Scene
	nextScene             Scene
	transitionedThisFrame bool
	overworldState        *OverworldScene
	caveState             *CaveScene
	baseMenu              *BaseMenuScene
	gameOverState         *GameOverScene
	gameWonState          *GameWonScene
	player                *Player
	hud                   *HUD
	world                 *world.World
	camera                *Camera
	Input                 InputSource

	// Surface return position
	lastOverworldX  float64
	lastOverworldY  float64
	activeTrenchX   int
	activeTrenchY   int
	activeTrenchKey string
	justExited      bool

	// Phase 5: Resource nodes and inventory state
	caveNodes     map[string][]resource.Resource
	showInventory bool

	// Phase 6: Base station and menu
	baseStation *BaseStation

	// Phase 7: Vehicle state tracking
	ActiveVehicle     vehicle.Vehicle
	OverworldVehicles []vehicle.Vehicle
	CaveVehicles      map[string][]vehicle.Vehicle // trenchKey -> list of vehicles spawned in that cave
	TimeOfDay         float64                      // 0 to 14400 ticks (4 min day/night cycle)
	SonarTimer        int
	SonarRadius       float64
	SonarRadiusStep   float64
	SonarSourceX      float64
	SonarSourceY      float64
	MineWarning       string
	MineWarningTimer  int

	// Phase 8: Biomes & Predator AI
	caveEntities        map[string][]CaveEntity
	FlashlightOn        bool
	WeaverTrackingTimer float64
	SoundWaveTimer      int
	SoundWaveRadius     float64
	SoundWaveX          float64
	SoundWaveY          float64
	playerSlowed        bool

	// Phase 9: Particles, screen shake
	Particles      []Particle
	shakeDuration  int
	shakeIntensity float64
}

// NewGame creates and returns a new Game instance.
func NewGame() *Game {
	w := world.NewWorld(12345)

	// Search for a starting water tile around the center map coordinates
	spawnX := 50.0 * TileSize
	spawnY := 50.0 * TileSize
	found := false
	for x := 45; x < 55; x++ {
		for y := 45; y < 55; y++ {
			if w.OverworldMap[x][y] == world.TileWater {
				spawnX = float64(x*TileSize) + (TileSize-20.0)/2.0
				spawnY = float64(y*TileSize) + (TileSize-20.0)/2.0
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	p := NewPlayer(spawnX, spawnY)
	cam := NewCamera(spawnX, spawnY)
	cam.CenterOn(spawnX, spawnY, p.Width, p.Height)

	// Spawn Base Station (Life Pod 5) slightly offset from the starting coordinates
	baseStation := NewBaseStation(spawnX+96.0, spawnY-64.0)

	// Phase 7: Create starting surface skiff and place player inside it
	skiff := vehicle.NewSkiff(spawnX, spawnY)
	overworldVehicles := []vehicle.Vehicle{skiff}

	g := &Game{
		currentState:      StateOverworld,
		player:            p,
		hud:               NewHUD(),
		world:             w,
		camera:            cam,
		Input:             NewEbitenInput(),
		caveNodes:         make(map[string][]resource.Resource),
		showInventory:     false,
		baseStation:       baseStation,
		ActiveVehicle:     skiff,
		OverworldVehicles: overworldVehicles,
		CaveVehicles:      make(map[string][]vehicle.Vehicle),
		TimeOfDay:         0.0,
		SonarRadiusStep:   6.5,
		caveEntities:      make(map[string][]CaveEntity),
		FlashlightOn:      true,
	}

	// Initialize scenes
	g.overworldState = NewOverworldScene(w)
	g.caveState = NewCaveScene()
	g.baseMenu = NewBaseMenuScene()
	g.gameOverState = NewGameOverScene()
	g.gameWonState = NewGameWonScene()

	// Set initial scene
	g.TransitionTo(g.overworldState)

	return g
}

// TransitionTo switches the active scene cleanly, executing teardown and initialization hooks.
func (g *Game) TransitionTo(next Scene) {
	if g.currentScene != nil {
		g.currentScene.OnExit(g)
	}
	g.currentScene = next
	if next != nil {
		next.OnEnter(g)
	}
	g.transitionedThisFrame = true
}

// EnterCave handles the transition from Overworld to Cave at trench coordinate tx, ty.
func (g *Game) EnterCave(tx, ty int) {
	g.lastOverworldX = g.player.Pos.X
	g.lastOverworldY = g.player.Pos.Y
	g.activeTrenchX = tx
	g.activeTrenchY = ty
	g.activeTrenchKey = fmt.Sprintf("%d_%d", tx, ty)

	g.caveState.CaveGrid = g.world.GetCave(tx, ty)
	g.caveState.IsShallow = g.world.OverworldMap[tx][ty] != world.TileTrench

	if _, exists := g.caveNodes[g.activeTrenchKey]; !exists {
		g.caveNodes[g.activeTrenchKey] = resource.GenerateResourceNodes(g.caveState.CaveGrid, int64(tx*97+ty*41))
	}
	g.caveState.Nodes = g.caveNodes[g.activeTrenchKey]

	if _, exists := g.caveEntities[g.activeTrenchKey]; !exists {
		g.caveEntities[g.activeTrenchKey] = GenerateCaveEntities(g.caveState.CaveGrid, int64(tx*97+ty*41), g.caveState.IsShallow)
	}
	g.caveState.Entities = g.caveEntities[g.activeTrenchKey]

	g.player.Pos.X = float64(len(g.caveState.CaveGrid)/2*TileSize) + (TileSize-g.player.Width)/2
	g.player.Pos.Y = TileSize * 2
	g.player.Vel = gvec.Vec2{}

	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
	g.TransitionTo(g.caveState)
}

// ExitCave handles the transition from Cave to Overworld.
func (g *Game) ExitCave() {
	g.player.Pos.X = g.lastOverworldX
	g.player.Pos.Y = g.lastOverworldY - TileSize*0.6
	g.player.Vel = gvec.Vec2{X: 0, Y: -1.5}

	g.caveNodes[g.activeTrenchKey] = g.caveState.Nodes
	g.caveEntities[g.activeTrenchKey] = g.caveState.Entities

	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
	g.TransitionTo(g.overworldState)
}

// Respawn resets the game completely on death.
func (g *Game) Respawn() {
	// Reposition player to Life Pod coordinates
	g.player.Pos = gvec.Vec2{X: g.baseStation.Pos.X - 96.0, Y: g.baseStation.Pos.Y + 64.0}
	g.player.Vel = gvec.Vec2{}

	// Refill stats
	g.player.CurrentHealth = g.player.MaxHealth
	g.player.CurrentOxygen = g.player.MaxOxygen
	g.player.CurrentStamina = g.player.MaxStamina

	// Eject from active vehicle
	g.ActiveVehicle = nil

	// Clear player's main inventory slots while keeping equipped upgrades safe
	g.player.Inventory.Clear()

	// Clear any active screen shake
	g.shakeDuration = 0
	g.shakeIntensity = 0.0

	// Close inventory screen overlay
	g.showInventory = false

	// Position camera centered on player at the Life Pod
	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)

	// Transition back to overworld scene
	g.TransitionTo(g.overworldState)
}

// TriggerScreenShake registers a screen shake request.
func (g *Game) TriggerScreenShake(duration int, intensity float64) {
	if intensity > g.shakeIntensity || g.shakeDuration <= 0 {
		g.shakeIntensity = intensity
	}
	if duration > g.shakeDuration {
		g.shakeDuration = duration
	}
}

// Update updates the game logical state.
func (g *Game) Update() error {
	g.transitionedThisFrame = false
	// Update user input polling cache
	g.Input.Update()

	g.justExited = false
	g.playerSlowed = false

	// Check for queued scene changes
	if g.nextScene != nil {
		g.TransitionTo(g.nextScene)
		g.nextScene = nil
	}

	// Increment day/night cycle timeOfDay (reset after 4 minutes)
	g.TimeOfDay += 1.0
	if g.TimeOfDay >= 14400 {
		g.TimeOfDay = 0.0
	}

	// Update warning display timer
	if g.MineWarningTimer > 0 {
		g.MineWarningTimer--
	}

	// Expand Sonar Ping wavefront inside caves
	if g.SonarTimer > 0 {
		g.SonarTimer--
		g.SonarRadius += g.SonarRadiusStep
	}

	// Update popped Shatter-bulb sound wave circle
	if g.SoundWaveTimer > 0 {
		g.SoundWaveTimer--
		g.SoundWaveRadius += 4.5
	}

	// Update active particles (bubbles, mining debris)
	g.UpdateParticles()

	// Toggle flashlight keybind (T)
	if g.Input.IsKeyJustPressed(ebiten.KeyT) {
		g.FlashlightOn = !g.FlashlightOn
	}

	// Toggle inventory overlay
	if g.Input.IsKeyJustPressed(ebiten.KeyTab) && (g.currentState == StateOverworld || g.currentState == StateCave) {
		g.showInventory = !g.showInventory
	}

	// Debug overrides to switch states manually (forces ejecting vehicle to prevent coordinate glitches)
	if g.Input.IsKeyJustPressed(ebiten.KeyO) {
		g.ActiveVehicle = nil
		g.showInventory = false
		g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
		g.TransitionTo(g.overworldState)
	} else if g.Input.IsKeyJustPressed(ebiten.KeyC) {
		g.ActiveVehicle = nil
		g.activeTrenchX = 50
		g.activeTrenchY = 50
		g.activeTrenchKey = "50_50"
		g.caveState.CaveGrid = g.world.GetCave(50, 50)
		g.player.Pos.X = float64(len(g.caveState.CaveGrid) / 2 * TileSize)
		g.player.Pos.Y = TileSize * 2

		if _, exists := g.caveNodes[g.activeTrenchKey]; !exists {
			g.caveNodes[g.activeTrenchKey] = resource.GenerateResourceNodes(g.caveState.CaveGrid, 50*97+50*41)
		}
		g.caveState.Nodes = g.caveNodes[g.activeTrenchKey]

		isShallow := g.world.OverworldMap[50][50] != world.TileTrench
		g.caveState.IsShallow = isShallow

		if _, exists := g.caveEntities[g.activeTrenchKey]; !exists {
			g.caveEntities[g.activeTrenchKey] = GenerateCaveEntities(g.caveState.CaveGrid, 50*97+50*41, isShallow)
		}
		g.caveState.Entities = g.caveEntities[g.activeTrenchKey]

		g.showInventory = false
		g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
		g.TransitionTo(g.caveState)
	} else if g.currentState == StateOverworld && g.baseStation.DistanceToPlayer(g.player) < 100.0 && g.Input.IsKeyJustPressed(ebiten.KeyE) {
		g.TransitionTo(g.baseMenu)
	} else if g.Input.IsKeyJustPressed(ebiten.KeyG) {
		g.TransitionTo(g.gameOverState)
	}

	// Debug shortcuts for testing vehicles and resources
	if g.currentState == StateOverworld || g.currentState == StateCave {
		if g.Input.IsKeyJustPressed(ebiten.Key1) {
			sub := vehicle.NewScoutSub(g.player.Pos.X, g.player.Pos.Y)
			if g.currentState == StateOverworld {
				g.OverworldVehicles = append(g.OverworldVehicles, sub)
			} else {
				g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], sub)
			}
			g.ActiveVehicle = sub
		} else if g.Input.IsKeyJustPressed(ebiten.Key2) {
			mech := vehicle.NewHeavyMech(g.player.Pos.X, g.player.Pos.Y)
			if g.currentState == StateOverworld {
				g.OverworldVehicles = append(g.OverworldVehicles, mech)
			} else {
				g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], mech)
			}
			g.ActiveVehicle = mech
		} else if g.Input.IsKeyJustPressed(ebiten.Key3) {
			skiff := vehicle.NewSkiff(g.player.Pos.X, g.player.Pos.Y)
			if g.currentState == StateOverworld {
				g.OverworldVehicles = append(g.OverworldVehicles, skiff)
			} else {
				g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], skiff)
			}
			g.ActiveVehicle = skiff
		} else if g.Input.IsKeyJustPressed(ebiten.Key4) {
			g.player.Inventory.AddItem(&item.Titanium{}, 10)
			g.player.Inventory.AddItem(&item.Copper{}, 10)
			g.player.Inventory.AddItem(&item.Quartz{}, 10)
			g.player.Inventory.AddItem(&item.AbyssalOre{}, 10)
			g.player.RecalculateUpgrades()
		} else if g.Input.IsKeyJustPressed(ebiten.Key5) {
			g.player.CurrentHealth = g.player.MaxHealth
			g.player.CurrentOxygen = g.player.MaxOxygen
			g.player.CurrentStamina = g.player.MaxStamina
		}
	}

	// Update Base Station Solar power loops
	if g.baseStation != nil {
		g.baseStation.UpdatePower(g.TimeOfDay)
	}

	// ---------------------------------------------------------
	// Inventory Click Interactions (Transfers & Deployments)
	// ---------------------------------------------------------
	if g.showInventory {
		if g.ActiveVehicle != nil {
			if g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				cursor := g.Input.Cursor()
				mx, my := int(cursor.X), int(cursor.Y)
				panelX := float64(ScreenWidth-960) / 2.0
				panelY := float64(ScreenHeight-360) / 2.0

				pStartX := panelX + 30
				pStartY := panelY + 60
				slotSz := 48.0
				gap := 8.0

				// 1. Move item from Player Inventory to Vehicle Cargo
				for r := 0; r < 3; r++ {
					for c := 0; c < 8; c++ {
						idx := r*8 + c
						sx := int(pStartX) + c*int(slotSz+gap)
						sy := int(pStartY) + r*int(slotSz+gap)

						if mx >= sx && mx < sx+int(slotSz) && my >= sy && my < sy+int(slotSz) {
							if idx < len(g.player.Inventory.Slots) {
								slot := &g.player.Inventory.Slots[idx]
								if slot.Item != nil {
									if g.ActiveVehicle.GetCargo().AddItem(item.Clone(slot.Item), 1) {
										g.player.Inventory.Remove(slot.Item, 1)
										g.player.RecalculateUpgrades()
									}
								}
							}
						}
					}
				}

				// 2. Move item from Vehicle Cargo to Player Inventory
				vInv := g.ActiveVehicle.GetCargo()
				numSlots := len(vInv.Slots)
				var vCols, vRows int
				if numSlots == 24 {
					vCols, vRows = 8, 3
				} else if numSlots == 12 {
					vCols, vRows = 6, 2
				} else {
					vCols, vRows = 4, 2
				}

				vStartX := panelX + 510
				vStartY := panelY + 60
				for r := 0; r < vRows; r++ {
					for c := 0; c < vCols; c++ {
						idx := r*vCols + c
						sx := int(vStartX) + c*int(slotSz+gap)
						sy := int(vStartY) + r*int(slotSz+gap)

						if mx >= sx && mx < sx+int(slotSz) && my >= sy && my < sy+int(slotSz) {
							if idx < numSlots {
								slot := &vInv.Slots[idx]
								if slot.Item != nil {
									if g.player.Inventory.AddItem(item.Clone(slot.Item), 1) {
										vInv.Remove(slot.Item, 1)
										g.player.RecalculateUpgrades()
									}
								}
							}
						}
					}
				}
			}
		} else {
			// Single Player Inventory click interactions (equipping upgrades & deploying vehicle kits)
			if g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				cursor := g.Input.Cursor()
				mx, my := int(cursor.X), int(cursor.Y)
				panelX := float64(ScreenWidth-600) / 2.0
				panelY := float64(ScreenHeight-420) / 2.0
				cols := 8
				slotSz := 56.0
				gap := 10.0
				startX := panelX + (600.0-float64(cols*(56+10)-10))/2.0
				startY := panelY + 60.0

				// 1. Check click on main inventory grid
				for r := 0; r < 3; r++ {
					for c := 0; c < 8; c++ {
						idx := r*8 + c
						sx := int(startX) + c*int(slotSz+gap)
						sy := int(startY) + r*int(slotSz+gap)

						if mx >= sx && mx < sx+int(slotSz) && my >= sy && my < sy+int(slotSz) {
							if idx < len(g.player.Inventory.Slots) {
								slot := &g.player.Inventory.Slots[idx]
								if slot.Item != nil {
									// Try equipping as upgrade first
									if g.player.EquipUpgrade(slot.Item) {
										g.player.Inventory.Remove(slot.Item, 1)
										g.player.RecalculateUpgrades()
									} else if g.currentState == StateCave {
										// Deploy vehicle kits only inside caves
										switch slot.Item.(type) {
										case *item.ScoutSubKit:
											sub := vehicle.NewScoutSub(g.player.Pos.X, g.player.Pos.Y)
											g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], sub)
											g.player.Inventory.Remove(slot.Item, 1)
											g.player.RecalculateUpgrades()
											g.showInventory = false
										case *item.HeavyMechKit:
											mech := vehicle.NewHeavyMech(g.player.Pos.X, g.player.Pos.Y)
											g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], mech)
											g.player.Inventory.Remove(slot.Item, 1)
											g.player.RecalculateUpgrades()
											g.showInventory = false
										}
									}
								}
							}
						}
					}
				}

				// 2. Check click on player equipped gear slots
				gearStartX := panelX + (600.0-(4.0*slotSz+3.0*gap))/2.0
				gearSlotsY := startY + 3.0*(slotSz+gap) + 5.0 + 22.0
				for c := 0; c < 4; c++ {
					sx := int(gearStartX) + c*int(slotSz+gap)
					sy := int(gearSlotsY)

					if mx >= sx && mx < sx+int(slotSz) && my >= sy && my < sy+int(slotSz) {
						if g.player.Upgrades != nil && c < len(g.player.Upgrades.Slots) {
							slot := &g.player.Upgrades.Slots[c]
							if slot.Item != nil {
								// Uninstall back to main inventory
								if g.player.Inventory.AddItem(item.Clone(slot.Item), 1) {
									g.player.Upgrades.Remove(slot.Item, 1)
									g.player.RecalculateUpgrades()
								}
							}
						}
					}
				}
			}
		}
		return nil
	}

	// ---------------------------------------------------------
	// Depth Limit Crushing Damage Checks
	// ---------------------------------------------------------
	if g.currentState == StateCave && g.ActiveVehicle != nil {
		limit := g.ActiveVehicle.GetDepthLimit()
		vPos := g.ActiveVehicle.GetPos()
		vDims := g.ActiveVehicle.GetDimensions()
		depth := (vPos.Y + vDims.Y/2.0) / TileSize

		if limit > 0.0 && depth > limit {
			g.ActiveVehicle.TakeDamage(0.08) // Apply crush damage over time
			g.MineWarning = "WARNING: EXCEEDING MAXIMUM HULL DEPTH LIMIT!"
			g.MineWarningTimer = 2
		}

		// Hull Integrity failure destroys the vehicle
		if g.ActiveVehicle.GetHealth() <= 0.0 {
			// Destroy active vehicle and damage pilot
			g.player.CurrentHealth -= 40.0
			g.MineWarning = "VEHICLE CRUSHED BY DEEP-SEA PRESSURE!"
			g.MineWarningTimer = 180

			// Remove vehicle from list
			list := g.CaveVehicles[g.activeTrenchKey]
			for idx, v := range list {
				if v == g.ActiveVehicle {
					g.CaveVehicles[g.activeTrenchKey] = append(list[:idx], list[idx+1:]...)
					break
				}
			}
			g.ActiveVehicle = nil
		}
	}

	// ---------------------------------------------------------
	// Piloting / Vehicle Movement Loops
	// ---------------------------------------------------------
	vehicleRuntime := vehicleRuntimeAdapter{g: g}
	if g.ActiveVehicle != nil {
		g.ActiveVehicle.Update(vehicleRuntime)

		// Sync player location inside active vehicle
		vPos := g.ActiveVehicle.GetPos()
		vDims := g.ActiveVehicle.GetDimensions()
		g.player.Pos.X = vPos.X + (vDims.X-g.player.Width)/2.0
		g.player.Pos.Y = vPos.Y + (vDims.Y-g.player.Height)/2.0
		g.player.Vel = gvec.Vec2{}

		// Vehicle cockpit replenishes/locks pilot oxygen
		if g.ActiveVehicle.GetOxygen() > 0.0 {
			g.player.CurrentOxygen = g.player.MaxOxygen
		}

		// Exit vehicle keybind (F)
		if g.Input.IsKeyJustPressed(ebiten.KeyF) {
			safeX, safeY := vPos.X, vPos.Y
			if g.currentState == StateCave {
				if !g.caveState.isSolid(vPos.X-32, vPos.Y, g.player.Width, g.player.Height) {
					safeX = vPos.X - 32
				} else if !g.caveState.isSolid(vPos.X+vDims.X+12, vPos.Y, g.player.Width, g.player.Height) {
					safeX = vPos.X + vDims.X + 12
				} else if !g.caveState.isSolid(vPos.X, vPos.Y-32, g.player.Width, g.player.Height) {
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
	}

	// Proximity checks for entering vehicles
	if g.ActiveVehicle == nil && !g.justExited {
		if g.currentState == StateOverworld {
			for _, v := range g.OverworldVehicles {
				vPos := v.GetPos()
				vDims := v.GetDimensions()
				dist := math.Hypot(vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0, vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0)
				if dist < 60.0 {
					if g.Input.IsKeyJustPressed(ebiten.KeyF) {
						g.ActiveVehicle = v
						break
					}
				}
			}
		} else if g.currentState == StateCave {
			for _, v := range g.CaveVehicles[g.activeTrenchKey] {
				vPos := v.GetPos()
				vDims := v.GetDimensions()
				dist := math.Hypot(vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0, vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0)
				if dist < 60.0 {
					if g.Input.IsKeyJustPressed(ebiten.KeyF) {
						g.ActiveVehicle = v
						break
					}
				}
			}
		}
	}

	// Delegate active scene logic
	if !g.transitionedThisFrame {
		if err := g.currentScene.Update(g); err != nil {
			return err
		}
	}

	// Solar power trickles for overworld vehicles if on foot
	if g.currentState == StateOverworld {
		for _, v := range g.OverworldVehicles {
			if v != g.ActiveVehicle {
				v.Update(vehicleRuntime)
			}
		}
	} else if g.currentState == StateCave {
		// Update other cave vehicles left in this cavern
		for _, v := range g.CaveVehicles[g.activeTrenchKey] {
			if v != g.ActiveVehicle {
				v.Update(vehicleRuntime)
			}
		}
	}

	// Smooth camera tracking
	if g.currentState == StateOverworld || g.currentState == StateCave {
		g.camera.Track(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height, 0.08)

		// Apply camera shake/jitter if tracked by Electro-Weaver
		if g.currentState == StateCave && g.WeaverTrackingTimer > 0 {
			shakeMag := (g.WeaverTrackingTimer / 300.0) * 8.0
			g.camera.Pos.X += rand.Float64()*shakeMag - (shakeMag / 2.0)
			g.camera.Pos.Y += rand.Float64()*shakeMag - (shakeMag / 2.0)
		}

		// Apply dynamic vehicle collision/landing screen shake
		if g.shakeDuration > 0 {
			g.camera.Pos.X += rand.Float64()*g.shakeIntensity - (g.shakeIntensity / 2.0)
			g.camera.Pos.Y += rand.Float64()*g.shakeIntensity - (g.shakeIntensity / 2.0)
			g.shakeDuration--
		}
	}

	// Oxygen and Health updates when on foot (prevent drains inside vehicles)
	if g.ActiveVehicle == nil {
		inCave := g.currentState == StateCave
		g.player.UpdateStats(inCave, g.Input.IsKeyPressed(ebiten.KeyShift))
	}

	if g.player.CurrentHealth <= 0 {
		g.TransitionTo(g.gameOverState)
	}

	return nil
}

// Draw renders the game graphics.
func (g *Game) Draw(screen *ebiten.Image) {
	// 1. Draw the active scene
	g.currentScene.Draw(g, screen)

	// Draw active particles if in overworld or cave
	if g.currentState == StateOverworld || g.currentState == StateCave {
		g.DrawParticles(screen)
	}

	// 2. Render general scene independent overlays (overworld vehicles, Life Pod in overworld)
	if g.currentState == StateOverworld {
		// Render overworld vehicles (Skiff)
		for _, v := range g.OverworldVehicles {
			v.Draw(screen, g.camera.Pos.X, g.camera.Pos.Y)
		}

		// Render Base Life Pod
		g.baseStation.Draw(screen, g.camera)
		if g.baseStation.DistanceToPlayer(g.player) < 100.0 {
			sx := float32(g.baseStation.Pos.X-g.camera.Pos.X) + float32(g.baseStation.Size.X)/2.0 - 90
			sy := float32(g.baseStation.Pos.Y-g.camera.Pos.Y) - 30
			vector.FillRect(screen, sx, sy, 180, 24, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, "Press [E] to Open Terminal", int(sx)+12, int(sy)+4)
		}

		// Entry prompt for overworld vehicles
		if g.ActiveVehicle == nil {
			for _, v := range g.OverworldVehicles {
				vPos := v.GetPos()
				vDims := v.GetDimensions()
				dist := math.Hypot(vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0, vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0)
				if dist < 60.0 {
					sx := float32(vPos.X-g.camera.Pos.X) + float32(vDims.X)/2.0 - 75
					sy := float32(vPos.Y-g.camera.Pos.Y) - 25
					vector.FillRect(screen, sx, sy, 150, 20, color.RGBA{0, 0, 0, 180}, false)
					ebitenutil.DebugPrintAt(screen, "Press [F] to Pilot", int(sx)+22, int(sy)+2)
				}
			}
		}
	} else if g.currentState == StateCave {
		// Render cave vehicles
		for _, v := range g.CaveVehicles[g.activeTrenchKey] {
			v.Draw(screen, g.camera.Pos.X, g.camera.Pos.Y)
		}

		// Entry prompt for cave vehicles
		if g.ActiveVehicle == nil {
			for _, v := range g.CaveVehicles[g.activeTrenchKey] {
				vPos := v.GetPos()
				vDims := v.GetDimensions()
				dist := math.Hypot(vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0, vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0)
				if dist < 60.0 {
					sx := float32(vPos.X-g.camera.Pos.X) + float32(vDims.X)/2.0 - 75
					sy := float32(vPos.Y-g.camera.Pos.Y) - 25
					vector.FillRect(screen, sx, sy, 150, 20, color.RGBA{0, 0, 0, 180}, false)
					ebitenutil.DebugPrintAt(screen, "Press [F] to Pilot", int(sx)+22, int(sy)+2)
				}
			}
		}

		// Sonar ring overlay
		if g.SonarTimer > 0 {
			alpha := float32(g.SonarTimer) / 180.0
			clr := color.RGBA{45, 175, 215, uint8(255 * alpha * 0.45)}
			scx := float32(g.SonarSourceX - g.camera.Pos.X)
			scy := float32(g.SonarSourceY - g.camera.Pos.Y)
			vector.StrokeCircle(screen, scx, scy, float32(g.SonarRadius), 2.5, clr, false)
			vector.StrokeCircle(screen, scx, scy, float32(g.SonarRadius), 1.0, color.RGBA{220, 250, 255, uint8(255 * alpha)}, false)
		}
	}

	// 3. Draw HUD (which includes status bars, telemetry, warnings, and inventory)
	if g.currentState == StateOverworld || g.currentState == StateCave {
		g.hud.Draw(screen, g)
	}

	// Draw Warning messages if active
	if g.MineWarningTimer > 0 {
		wx := float32(ScreenWidth)/2.0 - 160
		wy := float32(ScreenHeight) / 4.0
		vector.FillRect(screen, wx, wy, 320, 30, color.RGBA{24, 6, 8, 220}, false)
		vector.StrokeRect(screen, wx, wy, 320, 30, 1.2, color.RGBA{235, 45, 45, 255}, false)
		ebitenutil.DebugPrintAt(screen, g.MineWarning, int(wx)+12, int(wy)+7)
	}
}

// Layout determines the virtual game screen resolution.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
