package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Game implements the ebiten.Game interface and coordinates scenes.
type Game struct {
	currentState   State
	player         *Player
	overworldState *OverworldState
	caveState      *CaveState
	hud            *HUD
	world          *world.World
	camera         *Camera

	// Surface return position
	lastOverworldX float64
	lastOverworldY float64
	activeTrenchX  int
	activeTrenchY  int
	justExited     bool

	// Phase 5: Resource nodes and inventory state
	caveNodes     map[string][]ResourceNode
	showInventory bool

	// Phase 6: Base station and menu
	baseStation *BaseStation
	baseMenu    *BaseMenu

	// Phase 7: Vehicle state tracking
	ActiveVehicle     Vehicle
	OverworldVehicles []Vehicle
	CaveVehicles      map[string][]Vehicle // trenchKey -> list of vehicles spawned in that cave
	TimeOfDay         float64              // 0 to 14400 ticks (4 min day/night cycle)
	SonarTimer        int
	SonarRadius       float64
	SonarSourceX      float64
	SonarSourceY      float64
	MineWarning       string
	MineWarningTimer  int

	// Phase 8: Biomes & Predator AI
	caveEntities        map[string][]*CaveEntity
	FlashlightOn        bool
	WeaverTrackingTimer float64
	SoundWaveTimer      int
	SoundWaveRadius     float64
	SoundWaveX          float64
	SoundWaveY          float64
	playerSlowed        bool
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
	skiff := NewSkiff(spawnX, spawnY)
	overworldVehicles := []Vehicle{skiff}

	return &Game{
		currentState:      StateOverworld,
		player:            p,
		overworldState:    NewOverworldState(p, w),
		caveState:         NewCaveState(p),
		hud:               NewHUD(),
		world:             w,
		camera:            cam,
		caveNodes:         make(map[string][]ResourceNode),
		showInventory:     false,
		baseStation:       baseStation,
		baseMenu:          NewBaseMenu(),
		ActiveVehicle:     skiff,
		OverworldVehicles: overworldVehicles,
		CaveVehicles:      make(map[string][]Vehicle),
		TimeOfDay:         0.0,
		caveEntities:      make(map[string][]*CaveEntity),
		FlashlightOn:      true,
	}
}

// Update updates the game logical state.
func (g *Game) Update() error {
	g.justExited = false
	g.playerSlowed = false

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
		g.SonarRadius += 6.5 // Expanding wave speed
	}

	// Update popped Shatter-bulb sound wave circle
	if g.SoundWaveTimer > 0 {
		g.SoundWaveTimer--
		g.SoundWaveRadius += 4.5
	}

	// Toggle flashlight keybind (T)
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.FlashlightOn = !g.FlashlightOn
	}

	// Toggle inventory overlay
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) && (g.currentState == StateOverworld || g.currentState == StateCave) {
		g.showInventory = !g.showInventory
	}

	// Debug overrides to switch states manually (forces ejecting vehicle to prevent coordinate glitches)
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		g.ActiveVehicle = nil
		g.currentState = StateOverworld
		g.showInventory = false
		g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.ActiveVehicle = nil
		// Load default central cave for testing
		g.activeTrenchX = 50
		g.activeTrenchY = 50
		g.caveState.CaveGrid = g.world.GetCave(50, 50)
		g.player.X = float64(len(g.caveState.CaveGrid) / 2 * TileSize)
		g.player.Y = TileSize * 2

		// Load nodes for debug cave
		key := "50_50"
		if _, exists := g.caveNodes[key]; !exists {
			g.caveNodes[key] = GenerateResourceNodes(g.caveState.CaveGrid, 50*97+50*41)
		}
		g.caveState.Nodes = g.caveNodes[key]

		if _, exists := g.caveEntities[key]; !exists {
			g.caveEntities[key] = GenerateCaveEntities(g.caveState.CaveGrid, 50*97+50*41)
		}
		g.caveState.Entities = g.caveEntities[key]

		g.showInventory = false
		g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
		g.currentState = StateCave
	} else if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.currentState = StateBaseMenu
	} else if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.currentState = StateGameOver
	}

	// Debug shortcuts for testing vehicles and resources
	if g.currentState == StateOverworld || g.currentState == StateCave {
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			// Spawn Scout Sub at player location and pilot it
			sub := NewScoutSub(g.player.X, g.player.Y)
			if g.currentState == StateOverworld {
				g.OverworldVehicles = append(g.OverworldVehicles, sub)
			} else {
				key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)
				g.CaveVehicles[key] = append(g.CaveVehicles[key], sub)
			}
			g.ActiveVehicle = sub
		} else if inpututil.IsKeyJustPressed(ebiten.Key2) {
			// Spawn Heavy Mech at player location and pilot it
			mech := NewHeavyMech(g.player.X, g.player.Y)
			if g.currentState == StateOverworld {
				g.OverworldVehicles = append(g.OverworldVehicles, mech)
			} else {
				key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)
				g.CaveVehicles[key] = append(g.CaveVehicles[key], mech)
			}
			g.ActiveVehicle = mech
		} else if inpututil.IsKeyJustPressed(ebiten.Key3) {
			// Spawn Skiff at player location and pilot it
			skiff := NewSkiff(g.player.X, g.player.Y)
			if g.currentState == StateOverworld {
				g.OverworldVehicles = append(g.OverworldVehicles, skiff)
			} else {
				key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)
				g.CaveVehicles[key] = append(g.CaveVehicles[key], skiff)
			}
			g.ActiveVehicle = skiff
		} else if inpututil.IsKeyJustPressed(ebiten.Key4) {
			// Add 10x Titanium, Copper, Quartz, Abyssal Ore to player inventory
			g.player.Inventory.AddItem(ItemTitanium, 10)
			g.player.Inventory.AddItem(ItemCopper, 10)
			g.player.Inventory.AddItem(ItemQuartz, 10)
			g.player.Inventory.AddItem(ItemAbyssalOre, 10)
		} else if inpututil.IsKeyJustPressed(ebiten.Key5) {
			// Heal player and restore stats
			g.player.CurrentHealth = g.player.MaxHealth
			g.player.CurrentOxygen = g.player.MaxOxygen
			g.player.CurrentStamina = g.player.MaxStamina
		}
	}

	// Update Base Station Solar power loops
	if g.baseStation != nil {
		g.baseStation.UpdatePower()
	}

	// ---------------------------------------------------------
	// Inventory Click Interactions (Transfers & Deployments)
	// ---------------------------------------------------------
	if g.showInventory {
		if g.ActiveVehicle != nil {
			// Click checks to transfer cargo between player inventory and vehicle
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				mx, my := ebiten.CursorPosition()
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
								item := &g.player.Inventory.Slots[idx]
								if item.Type != ItemNone {
									if g.ActiveVehicle.GetCargo().AddItem(item.Type, 1) {
										g.player.Inventory.RemoveItem(item.Type, 1)
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
								item := &vInv.Slots[idx]
								if item.Type != ItemNone {
									if g.player.Inventory.AddItem(item.Type, 1) {
										vInv.RemoveItem(item.Type, 1)
									}
								}
							}
						}
					}
				}
			}
		} else {
			// Single Player Inventory click deployments (Only inside Caves on foot)
			if g.currentState == StateCave && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				mx, my := ebiten.CursorPosition()
				panelX := float64(ScreenWidth-600) / 2.0
				panelY := float64(ScreenHeight-340) / 2.0
				cols := 8
				slotSz := 56.0
				gap := 10.0
				startX := panelX + (600.0-float64(cols*(56+10)-10))/2.0
				startY := panelY + 60.0

				for r := 0; r < 3; r++ {
					for c := 0; c < 8; c++ {
						idx := r*8 + c
						sx := int(startX) + c*int(slotSz+gap)
						sy := int(startY) + r*int(slotSz+gap)

						if mx >= sx && mx < sx+int(slotSz) && my >= sy && my < sy+int(slotSz) {
							if idx < len(g.player.Inventory.Slots) {
								item := &g.player.Inventory.Slots[idx]
								key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)

								if item.Type == ItemScoutSub {
									sub := NewScoutSub(g.player.X, g.player.Y)
									g.CaveVehicles[key] = append(g.CaveVehicles[key], sub)
									g.player.Inventory.RemoveItem(ItemScoutSub, 1)
									g.showInventory = false
								} else if item.Type == ItemHeavyMech {
									mech := NewHeavyMech(g.player.X, g.player.Y)
									g.CaveVehicles[key] = append(g.CaveVehicles[key], mech)
									g.player.Inventory.RemoveItem(ItemHeavyMech, 1)
									g.showInventory = false
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
		_, vy := g.ActiveVehicle.GetPos()
		_, vh := g.ActiveVehicle.GetDimensions()
		depth := (vy + vh/2.0) / TileSize

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
			key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)
			list := g.CaveVehicles[key]
			for idx, v := range list {
				if v == g.ActiveVehicle {
					g.CaveVehicles[key] = append(list[:idx], list[idx+1:]...)
					break
				}
			}
			g.ActiveVehicle = nil
		}
	}

	// ---------------------------------------------------------
	// Piloting / Vehicle Movement Loops
	// ---------------------------------------------------------
	if g.ActiveVehicle != nil {
		g.ActiveVehicle.Update(g)

		// Sync player location inside active vehicle
		vx, vy := g.ActiveVehicle.GetPos()
		vw, vh := g.ActiveVehicle.GetDimensions()
		g.player.X = vx + (vw-g.player.Width)/2.0
		g.player.Y = vy + (vh-g.player.Height)/2.0
		g.player.Vx = 0
		g.player.Vy = 0

		// Vehicle cockpit replenishes/locks pilot oxygen
		if g.ActiveVehicle.GetOxygen() > 0.0 {
			g.player.CurrentOxygen = g.player.MaxOxygen
		}

		// Exit vehicle keybind (F)
		if inpututil.IsKeyJustPressed(ebiten.KeyF) {
			safeX, safeY := vx, vy
			if g.currentState == StateCave {
				// Eject safely next to vehicle check
				if !g.caveState.isSolid(vx-32, vy, g.player.Width, g.player.Height) {
					safeX = vx - 32
				} else if !g.caveState.isSolid(vx+vw+12, vy, g.player.Width, g.player.Height) {
					safeX = vx + vw + 12
				} else if !g.caveState.isSolid(vx, vy-32, g.player.Width, g.player.Height) {
					safeY = vy - 32
				}
				g.player.X = safeX
				g.player.Y = safeY
			} else {
				g.player.X = vx - 24
			}
			g.ActiveVehicle = nil
			g.justExited = true
		}
	}

	// Update active scene
	switch g.currentState {
	case StateOverworld:
		// Base solar charge updates for overworld vehicles
		for _, v := range g.OverworldVehicles {
			if v != g.ActiveVehicle {
				v.Update(g)
			}
		}

		// Base terminal access check (only if on foot, or in Skiff next to it)
		if g.baseStation.DistanceToPlayer(g.player) < 80.0 && inpututil.IsKeyJustPressed(ebiten.KeyE) {
			g.currentState = StateBaseMenu
			g.showInventory = false
			return nil
		}

		// Proximity checks for entering overworld vehicles
		if g.ActiveVehicle == nil && !g.justExited {
			for _, v := range g.OverworldVehicles {
				vx, vy := v.GetPos()
				vw, vh := v.GetDimensions()
				dist := math.Hypot(vx+vw/2.0-g.player.X-g.player.Width/2.0, vy+vh/2.0-g.player.Y-g.player.Height/2.0)
				if dist < 60.0 {
					if inpututil.IsKeyJustPressed(ebiten.KeyF) {
						g.ActiveVehicle = v
						break
					}
				}
			}
		}

		// Swim on foot update
		if g.ActiveVehicle == nil {
			nextState, transited := g.overworldState.Update()
			if transited && nextState == StateCave {
				tx := int(g.player.X+g.player.Width/2) / TileSize
				ty := int(g.player.Y+g.player.Height/2) / TileSize

				g.lastOverworldX = g.player.X
				g.lastOverworldY = g.player.Y
				g.activeTrenchX = tx
				g.activeTrenchY = ty

				g.caveState.CaveGrid = g.world.GetCave(tx, ty)

				key := fmt.Sprintf("%d_%d", tx, ty)
				if _, exists := g.caveNodes[key]; !exists {
					g.caveNodes[key] = GenerateResourceNodes(g.caveState.CaveGrid, int64(tx*97+ty*41))
				}
				g.caveState.Nodes = g.caveNodes[key]

				if _, exists := g.caveEntities[key]; !exists {
					g.caveEntities[key] = GenerateCaveEntities(g.caveState.CaveGrid, int64(tx*97+ty*41))
				}
				g.caveState.Entities = g.caveEntities[key]

				g.player.X = float64(len(g.caveState.CaveGrid)/2*TileSize) + (TileSize-g.player.Width)/2
				g.player.Y = TileSize * 2
				g.player.Vx = 0
				g.player.Vy = 0

				g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
				g.currentState = StateCave
			}
		}

	case StateCave:
		key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)

		// Update other cave vehicles left in this cavern
		for _, v := range g.CaveVehicles[key] {
			if v != g.ActiveVehicle {
				v.Update(g)
			}
		}

		// Reset electrical tracking timer; active Electro-Weavers will update it
		g.WeaverTrackingTimer = 0.0

		// Update cave entities
		for _, ent := range g.caveState.Entities {
			ent.Update(g, g.caveState)
		}

		// Clean up deactivated entities
		var activeEnts []*CaveEntity
		for _, ent := range g.caveState.Entities {
			if ent.Active {
				activeEnts = append(activeEnts, ent)
			}
		}
		g.caveState.Entities = activeEnts
		g.caveEntities[key] = activeEnts

		// Heavy Mech drilling handler
		if mech, ok := g.ActiveVehicle.(*HeavyMech); ok && !mech.IsDrilling {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				mx, my := ebiten.CursorPosition()
				camX := mech.X - ScreenWidth/2 + mech.Width/2
				camY := mech.Y - ScreenHeight/2 + mech.Height/2
				worldX := camX + float64(mx)
				worldY := camY + float64(my)

				mtx := int(worldX) / TileSize
				mty := int(worldY) / TileSize

				for i := range g.caveState.Nodes {
					node := &g.caveState.Nodes[i]
					if node.Tx == mtx && node.Ty == mty && node.HitsToMine > 0 {
						px := mech.X + mech.Width/2
						py := mech.Y + mech.Height/2
						nx := float64(node.Tx*TileSize + TileSize/2)
						ny := float64(node.Ty*TileSize + TileSize/2)
						dist := math.Hypot(px-nx, py-ny)

						if dist <= 120.0 {
							mech.DrillStrike(node)
							break
						}
					}
				}
			}
		}

		// Proximity checks for entering cave vehicles
		if g.ActiveVehicle == nil && !g.justExited {
			for _, v := range g.CaveVehicles[key] {
				vx, vy := v.GetPos()
				vw, vh := v.GetDimensions()
				dist := math.Hypot(vx+vw/2.0-g.player.X-g.player.Width/2.0, vy+vh/2.0-g.player.Y-g.player.Height/2.0)
				if dist < 60.0 {
					if inpututil.IsKeyJustPressed(ebiten.KeyF) {
						g.ActiveVehicle = v
						break
					}
				}
			}
		}

		// On foot swimming updates
		if g.ActiveVehicle == nil {
			nextState, transited := g.caveState.Update(g)
			if transited && nextState == StateOverworld {
				g.player.X = g.lastOverworldX
				g.player.Y = g.lastOverworldY - TileSize*0.6
				g.player.Vx = 0
				g.player.Vy = -1.5

				g.caveNodes[key] = g.caveState.Nodes
				g.caveEntities[key] = g.caveState.Entities
				g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
				g.currentState = StateOverworld
			}
		}

	case StateBaseMenu:
		if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeyO) {
			g.currentState = StateOverworld
		} else {
			g.baseMenu.Update(g.player, g.baseStation)
		}

	case StateGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			*g = *NewGame()
		}
	}

	// Smooth camera tracking
	if g.currentState == StateOverworld || g.currentState == StateCave {
		g.camera.Track(g.player.X, g.player.Y, g.player.Width, g.player.Height, 0.08)

		// Apply camera shake/jitter if tracked by Electro-Weaver
		if g.currentState == StateCave && g.WeaverTrackingTimer > 0 {
			shakeMag := (g.WeaverTrackingTimer / 300.0) * 8.0
			g.camera.X += rand.Float64()*shakeMag - (shakeMag / 2.0)
			g.camera.Y += rand.Float64()*shakeMag - (shakeMag / 2.0)
		}
	}

	// Oxygen and Health updates when on foot (prevent drains inside vehicles)
	if g.ActiveVehicle == nil {
		inCave := g.currentState == StateCave
		g.player.UpdateStats(inCave, ebiten.IsKeyPressed(ebiten.KeyShift))
	}

	if g.player.CurrentHealth <= 0 {
		g.currentState = StateGameOver
	}

	return nil
}

// Draw renders the game graphics.
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.currentState {
	case StateOverworld:
		// Render tiles
		g.overworldState.Draw(screen, g.camera, g.ActiveVehicle != nil)

		// Render overworld vehicles (Skiff)
		for _, v := range g.OverworldVehicles {
			v.Draw(screen, g.camera)
		}

		// Render Base Life Pod
		g.baseStation.Draw(screen, g.camera)
		if g.baseStation.DistanceToPlayer(g.player) < 80.0 {
			sx := float32(g.baseStation.X-g.camera.X) + float32(g.baseStation.Width)/2.0 - 90
			sy := float32(g.baseStation.Y-g.camera.Y) - 30
			vector.DrawFilledRect(screen, sx, sy, 180, 24, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, "Press [E] to Open Terminal", int(sx)+12, int(sy)+4)
		}

		// Entry prompt for overworld vehicles
		if g.ActiveVehicle == nil {
			for _, v := range g.OverworldVehicles {
				vx, vy := v.GetPos()
				vw, vh := v.GetDimensions()
				dist := math.Hypot(vx+vw/2.0-g.player.X-g.player.Width/2.0, vy+vh/2.0-g.player.Y-g.player.Height/2.0)
				if dist < 60.0 {
					sx := float32(vx-g.camera.X) + float32(vw)/2.0 - 75
					sy := float32(vy-g.camera.Y) - 25
					vector.DrawFilledRect(screen, sx, sy, 150, 20, color.RGBA{0, 0, 0, 180}, false)
					ebitenutil.DebugPrintAt(screen, "Press [F] to Pilot", int(sx)+22, int(sy)+2)
				}
			}
		}

		g.hud.Draw(screen, g)

	case StateCave:
		// Render cave tiles and flashlight mask
		g.caveState.Draw(screen, g.camera, g)

		// Render cave vehicles
		key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)
		for _, v := range g.CaveVehicles[key] {
			v.Draw(screen, g.camera)
		}

		// Entry prompt for cave vehicles
		if g.ActiveVehicle == nil {
			for _, v := range g.CaveVehicles[key] {
				vx, vy := v.GetPos()
				vw, vh := v.GetDimensions()
				dist := math.Hypot(vx+vw/2.0-g.player.X-g.player.Width/2.0, vy+vh/2.0-g.player.Y-g.player.Height/2.0)
				if dist < 60.0 {
					sx := float32(vx-g.camera.X) + float32(vw)/2.0 - 75
					sy := float32(vy-g.camera.Y) - 25
					vector.DrawFilledRect(screen, sx, sy, 150, 20, color.RGBA{0, 0, 0, 180}, false)
					ebitenutil.DebugPrintAt(screen, "Press [F] to Pilot", int(sx)+22, int(sy)+2)
				}
			}
		}

		// Sonar ring overlay
		if g.SonarTimer > 0 {
			alpha := float32(g.SonarTimer) / 180.0
			clr := color.RGBA{45, 175, 215, uint8(255 * alpha * 0.45)}
			scx := float32(g.SonarSourceX - g.camera.X)
			scy := float32(g.SonarSourceY - g.camera.Y)
			vector.StrokeCircle(screen, scx, scy, float32(g.SonarRadius), 2.5, clr, false)
			vector.StrokeCircle(screen, scx, scy, float32(g.SonarRadius), 1.0, color.RGBA{220, 250, 255, uint8(255 * alpha)}, false)
		}

		g.hud.Draw(screen, g)

	case StateBaseMenu:
		g.baseMenu.Draw(screen, g.player, g.baseStation)

	case StateGameOver:
		screen.Fill(color.RGBA{50, 10, 10, 255})
		ebitenutil.DebugPrint(screen, "GAME OVER\n\nYour hull cracked or you ran out of oxygen.\n\nPress ENTER to respawn.")
	}

	// Draw Warning messages if active
	if g.MineWarningTimer > 0 {
		wx := float32(ScreenWidth)/2.0 - 160
		wy := float32(ScreenHeight) / 4.0
		vector.DrawFilledRect(screen, wx, wy, 320, 30, color.RGBA{24, 6, 8, 220}, false)
		vector.StrokeRect(screen, wx, wy, 320, 30, 1.2, color.RGBA{235, 45, 45, 255}, false)
		ebitenutil.DebugPrintAt(screen, g.MineWarning, int(wx)+12, int(wy)+7)
	}

	// Render inventory overlay
	if g.showInventory {
		if g.ActiveVehicle != nil {
			g.hud.DrawVehicleInventory(screen, g.player.Inventory, g.ActiveVehicle.GetCargo(), g.ActiveVehicle.GetName())
		} else {
			g.hud.DrawInventory(screen, g.player.Inventory)
		}
	}
}

// Layout determines the virtual game screen resolution.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
