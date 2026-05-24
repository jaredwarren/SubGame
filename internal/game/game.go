package game

import (
	"fmt"
	"image/color"

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

	// Phase 5: Resource nodes and inventory state
	caveNodes     map[string][]ResourceNode
	showInventory bool

	// Phase 6: Base station and menu
	baseStation *BaseStation
	baseMenu    *BaseMenu
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

	// Spawn Base Station (Life Pod 5) slightly offset from the starting boat coordinates
	baseStation := NewBaseStation(spawnX+96.0, spawnY-64.0)

	return &Game{
		currentState:   StateOverworld,
		player:         p,
		overworldState: NewOverworldState(p, w),
		caveState:      NewCaveState(p),
		hud:            NewHUD(),
		world:          w,
		camera:         cam,
		caveNodes:      make(map[string][]ResourceNode),
		showInventory:  false,
		baseStation:    baseStation,
		baseMenu:       NewBaseMenu(),
	}
}

// Update updates the game logical state.
func (g *Game) Update() error {
	// Toggle inventory overlay
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) && (g.currentState == StateOverworld || g.currentState == StateCave) {
		g.showInventory = !g.showInventory
	}

	// Debug overrides to switch states manually (will force default coordinates if used)
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		g.currentState = StateOverworld
		g.showInventory = false
		g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		// Load a default central cave for testing
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

		g.showInventory = false
		g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
		g.currentState = StateCave
	} else if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.currentState = StateBaseMenu
	} else if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.currentState = StateGameOver
	}

	// Update Base Station Solar power grid loops
	if g.baseStation != nil {
		g.baseStation.UpdatePower()
	}

	// If inventory panel is open, freeze standard gameplay logic updates
	if g.showInventory {
		return nil
	}

	// Update the active state scene
	switch g.currentState {
	case StateOverworld:
		// Interactivity check for opening Base Terminal
		if g.baseStation.DistanceToPlayer(g.player) < 80.0 && inpututil.IsKeyJustPressed(ebiten.KeyE) {
			g.currentState = StateBaseMenu
			g.showInventory = false
			return nil
		}

		nextState, transited := g.overworldState.Update()
		if transited && nextState == StateCave {
			// Find player's current overworld tile coordinates
			tx := int(g.player.X+g.player.Width/2) / TileSize
			ty := int(g.player.Y+g.player.Height/2) / TileSize

			// Store return coordinates
			g.lastOverworldX = g.player.X
			g.lastOverworldY = g.player.Y
			g.activeTrenchX = tx
			g.activeTrenchY = ty

			// Generate and load cave grid
			g.caveState.CaveGrid = g.world.GetCave(tx, ty)

			// Generate resource nodes if this trench cave is entered for the first time
			key := fmt.Sprintf("%d_%d", tx, ty)
			if _, exists := g.caveNodes[key]; !exists {
				g.caveNodes[key] = GenerateResourceNodes(g.caveState.CaveGrid, int64(tx*97+ty*41))
			}
			g.caveState.Nodes = g.caveNodes[key]

			// Spawn player at top center of cave
			g.player.X = float64(len(g.caveState.CaveGrid)/2*TileSize) + (TileSize-g.player.Width)/2
			g.player.Y = TileSize * 2
			g.player.Vx = 0
			g.player.Vy = 0

			g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
			g.currentState = StateCave
		}
	case StateCave:
		nextState, transited := g.caveState.Update()
		if transited && nextState == StateOverworld {
			// Restore player position in Overworld next to the trench
			g.player.X = g.lastOverworldX
			// Place boat slightly above/away from trench center to prevent instant loop
			g.player.Y = g.lastOverworldY - TileSize*0.6
			g.player.Vx = 0
			g.player.Vy = -1.5 // Give a small upward nudge out of the trench

			// Save back any modifications to the cave's resource nodes
			key := fmt.Sprintf("%d_%d", g.activeTrenchX, g.activeTrenchY)
			g.caveNodes[key] = g.caveState.Nodes

			g.camera.CenterOn(g.player.X, g.player.Y, g.player.Width, g.player.Height)
			g.currentState = StateOverworld
		}
	case StateBaseMenu:
		// Return back to overworld upon pressing E, O or Escape
		if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeyO) {
			g.currentState = StateOverworld
		} else {
			g.baseMenu.Update(g.player, g.baseStation)
		}
	case StateGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			// Restart game by creating a new session
			*g = *NewGame()
		}
	}

	// Smoothly track camera if in overworld or cave states
	if g.currentState == StateOverworld || g.currentState == StateCave {
		g.camera.Track(g.player.X, g.player.Y, g.player.Width, g.player.Height, 0.08)
	}

	// Health check triggers game over
	if g.player.CurrentHealth <= 0 {
		g.currentState = StateGameOver
	}

	return nil
}

// Draw renders the game graphics.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen with a color representing the state
	switch g.currentState {
	case StateOverworld:
		g.overworldState.Draw(screen, g.camera)
		
		// Render Base Station Life Pod
		g.baseStation.Draw(screen, g.camera)
		if g.baseStation.DistanceToPlayer(g.player) < 80.0 {
			sx := float32(g.baseStation.X-g.camera.X) + float32(g.baseStation.Width)/2.0 - 90
			sy := float32(g.baseStation.Y-g.camera.Y) - 30
			vector.DrawFilledRect(screen, sx, sy, 180, 24, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, "Press [E] to Open Terminal", int(sx)+12, int(sy)+4)
		}

		g.hud.Draw(screen, g.player)
	case StateCave:
		g.caveState.Draw(screen, g.camera)
		g.hud.Draw(screen, g.player)
	case StateBaseMenu:
		g.baseMenu.Draw(screen, g.player, g.baseStation)
	case StateGameOver:
		screen.Fill(color.RGBA{50, 10, 10, 255})
		ebitenutil.DebugPrint(screen, "GAME OVER\n\nYour hull cracked or you ran out of oxygen.\n\nPress ENTER to respawn.")
	}

	// Draw universal debugging layout instructions at top-left
	// legend := "Controls:\n- WASD / Arrows to Move/Steer\n- Shift to Sprint/Boost\n- Mouse to Aim Flashlight (Cave State)\n- Key [O] Overworld | Key [C] Cave | Key [M] Base Menu"
	vector.DrawFilledRect(screen, 10, 10, 380, 85, color.RGBA{0, 0, 0, 150}, false)
	vector.StrokeRect(screen, 10, 10, 380, 85, 1.0, color.RGBA{70, 70, 70, 255}, false)
	// ebitenutil.DebugPrintAt(screen, legend, 18, 18)

	// Phase 5: Render Inventory Overlay on top of gameplay and HUD
	if g.showInventory {
		g.hud.DrawInventory(screen, g.player.Inventory)
	}
}

// Layout determines the virtual game screen resolution.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
