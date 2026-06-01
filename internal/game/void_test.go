package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestOverworldVoidBoundaries(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.ActiveVehicle = nil // Player on foot
	g.player.Pos = gvec.Vec2{X: -100, Y: -100} // Place player out of bounds (top-left)

	// Verify that the position is solid check returns false (i.e. we can sail past boundary)
	if g.overworldState.IsSolid(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height) {
		t.Error("expected out of bounds coordinates to NOT be solid")
	}

	// Update movement: player has velocity
	g.player.Vel = gvec.Vec2{X: -1.0, Y: -1.0}
	g.overworldState.CheckCollisions(g.player)

	// Since it's not solid, player position should have moved
	if g.player.Pos.X >= -100.0 {
		t.Errorf("expected player Pos.X to decrease when moving left in the void, got %f", g.player.Pos.X)
	}
}

func TestVoidCaveCreation(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.ActiveVehicle = nil // Diver on foot

	// Sail/swim out of bounds
	g.player.Pos = gvec.Vec2{X: -10 * config.TileSize, Y: -10 * config.TileSize}

	// Dive by calling EnterCave directly with out-of-bounds coordinates
	tx := int(g.player.Pos.X) / config.TileSize
	ty := int(g.player.Pos.Y) / config.TileSize
	g.EnterCave(tx, ty)

	// Verify that we are in the cave state
	if g.currentState != StateCave {
		t.Errorf("expected state to transition to StateCave, got %v", g.currentState)
	}

	// Verify that the active cave is VoidCave
	if g.caveState.ActiveCave == nil {
		t.Fatal("expected ActiveCave to be set")
	}
	if g.caveState.ActiveCave.GetCaveType() != cave.CaveVoid {
		t.Errorf("expected ActiveCave type to be CaveVoid, got %v", g.caveState.ActiveCave.GetCaveType())
	}

	// Verify that CaveGrid is nil
	if g.caveState.CaveGrid != nil {
		t.Error("expected CaveGrid to be nil for Void Cave")
	}

	// Verify player starting position
	expectedStartX := float64(30 * config.TileSize)
	if g.player.Pos.X != expectedStartX {
		t.Errorf("expected player starting X in Void Cave to be %f, got %f", expectedStartX, g.player.Pos.X)
	}

	// Verify that collision check is solid returns false everywhere (allowing infinite movement)
	if g.caveState.IsSolid(g, 100, 100, g.player.Width, g.player.Height) {
		t.Error("expected isSolid to return false everywhere in Void Cave")
	}
	if g.caveState.IsSolid(g, -5000, 50000, g.player.Width, g.player.Height) {
		t.Error("expected isSolid to return false even deep in Void Cave")
	}

	// Verify that exiting the Void Cave works: swim back up Y <= -8
	g.player.Pos.Y = -10
	err := g.caveState.Update(g)
	if err != nil {
		t.Fatal(err)
	}

	// Verify player returned to overworld
	if g.currentState != StateOverworld {
		t.Errorf("expected to return to StateOverworld when swimming up, got %v", g.currentState)
	}
}

func TestVoidCaveFirstFrameCoords(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.ActiveVehicle = nil // Player on foot swimming on overworld

	// Set player position out of bounds (tx = -10, ty = -10)
	g.player.Pos = gvec.Vec2{X: -10 * config.TileSize, Y: -10 * config.TileSize}

	// Set camera position to track player in overworld
	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)

	// Mock pressing KeyE
	mockInput := NewMockInput()
	mockInput.JustPressedKeys[ebiten.KeyE] = true
	g.Input = mockInput

	t.Logf("Before Update: state = %v, player.Pos = %+v, camera.Pos = %+v", g.currentState, g.player.Pos, g.camera.Pos)

	// Call Update: this should run overworldState.Update, trigger EnterCave, transition to StateCave
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("After Update 1 (Transition Frame): state = %v, player.Pos = %+v, camera.Pos = %+v", g.currentState, g.player.Pos, g.camera.Pos)

	// Draw the screen
	screen := ebiten.NewImage(1280, 720)
	g.Draw(screen)

	t.Logf("Draw 1 uniforms: LightSource = %+v, PersonalRadius = %v", g.caveState.Uniforms["LightSource"], g.caveState.Uniforms["PersonalRadius"])

	// Check if player position is correct (1920, 128)
	if g.player.Pos.X != float64(30 * config.TileSize) || g.player.Pos.Y != config.TileSize * 2 {
		t.Errorf("expected player pos to be (1920, 128), got %+v", g.player.Pos)
	}

	// Check if camera centered properly
	expectedCamX := 1920.0 + 10.0 - 640.0 // 1290
	expectedCamY := 128.0 + 10.0 - 360.0  // -222
	if g.camera.Pos.X != expectedCamX || g.camera.Pos.Y != expectedCamY {
		t.Errorf("expected camera pos to be (%f, %f), got %+v", expectedCamX, expectedCamY, g.camera.Pos)
	}

	// Check if LightSource in shader uniforms is centered [640, 360]
	ls := g.caveState.Uniforms["LightSource"].([]float32)
	if ls[0] != 640.0 || ls[1] != 360.0 {
		t.Errorf("expected LightSource to be [640, 360], got %+v", ls)
	}

	// Simulating the next 10 frames
	for frame := 2; frame <= 11; frame++ {
		// Mock holding or not holding keys (player drifts/swims)
		err = g.Update()
		if err != nil {
			t.Fatal(err)
		}
		g.Draw(screen)
		ls = g.caveState.Uniforms["LightSource"].([]float32)
		t.Logf("Frame %d: player.Pos = %+v, camera.Pos = %+v, LightSource = %+v", frame, g.player.Pos, g.camera.Pos, ls)
	}
}

func TestCaveActiveVehicleAndDebugCaveSetup(t *testing.T) {
	g := NewGame()

	// 1. Verify ActiveVehicle is cleared when entering any cave
	if g.ActiveVehicle == nil {
		t.Fatal("expected default ActiveVehicle to be initialized (skiff)")
	}

	// Dive at coordinates (50, 50)
	g.EnterCave(50, 50)
	if g.ActiveVehicle != nil {
		t.Error("expected ActiveVehicle to be cleared (nil) when transitioning to a cave")
	}

	// 2. Verify debug KeyC setup initializes ActiveCave
	g.TransitionTo(g.overworldState)
	// Mock pressing KeyC
	mockInput := NewMockInput()
	mockInput.JustPressedKeys[ebiten.KeyC] = true
	g.Input = mockInput

	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	if g.currentState != StateCave {
		t.Errorf("expected KeyC to transition to StateCave, got %v", g.currentState)
	}
	if g.caveState.ActiveCave == nil {
		t.Error("expected KeyC transition to initialize ActiveCave in caveState")
	}
}

func TestMiningAndCraftingCompatibility(t *testing.T) {
	g := NewGame()
	node := resource.NewTitaniumNode(0, 0)
	g.player.Inventory.AddItem(node, 1)
	if !g.player.Inventory.Has(&item.Titanium{}, 1) {
		t.Error("game-breaking bug: inventory does not recognize TitaniumNode as item.Titanium")
	}
}

func TestWreckageCaveGeneration(t *testing.T) {
	g := NewGame()

	// Create wreckage cave coordinate
	tx, ty := 10, 10
	g.world.OverworldMap[tx][ty] = world.TileWreckage

	// Transition to wreckage cave
	g.EnterCave(tx, ty)

	// Assert cave dimensions
	grid := g.caveState.CaveGrid
	if len(grid) != 60 || len(grid[0]) != 120 {
		t.Errorf("expected wreckage cave size 60x120, got %dx%d", len(grid), len(grid[0]))
	}

	// Assert active cave type is wreckage
	if g.caveState.ActiveCave.GetCaveType() != cave.CaveWreckage {
		t.Errorf("expected active cave to be CaveWreckage, got %v", g.caveState.ActiveCave.GetCaveType())
	}

	// Assert central elevator shaft is hollow (solid == false)
	for y := 5; y < 100; y++ {
		for x := 27; x <= 32; x++ {
			if grid[x][y] {
				t.Errorf("expected central elevator shaft tile at (%d, %d) to be hollow (false), but was solid (true)", x, y)
			}
		}
	}

	// Assert that we have ScrapMetalNode and ElectronicWasteNode in the cave nodes
	var scrapCount, electronicCount int
	for _, node := range g.caveState.Nodes {
		if node.GetName() == "Scrap Metal" {
			scrapCount++
		} else if node.GetName() == "Electronic Waste" {
			electronicCount++
		}
	}
	t.Logf("Generated %d Scrap Metal nodes and %d Electronic Waste nodes in wreckage cave", scrapCount, electronicCount)
	if scrapCount == 0 && electronicCount == 0 {
		t.Error("expected wreckage cave to generate some scrap metal or electronic waste nodes")
	}
}

func TestCaveExitLandSpawningFix(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)

	// Set up a mock overworld tile grid where (10, 10) is TileWater,
	// and (10, 9) is TileLand (directly above).
	tx, ty := 10, 10
	g.world.OverworldMap[tx][ty] = world.TileWater
	g.world.OverworldMap[tx][ty-1] = world.TileLand

	// Place the player at (10, 10) in pixel coordinates, but close to the top of the tile
	// so that shifting up by TileSize*0.6 would normally push them into (10, 9).
	playerStartX := float64(tx*config.TileSize + 10)
	playerStartY := float64(ty*config.TileSize + 5)
	g.player.Pos = gvec.Vec2{X: playerStartX, Y: playerStartY}

	// Enter the cave. This sets lastOverworldX/Y.
	g.EnterCave(tx, ty)

	// Verify coordinates were stored
	if g.lastOverworldX != playerStartX || g.lastOverworldY != playerStartY {
		t.Errorf("expected last overworld coords to be (%f, %f), got (%f, %f)",
			playerStartX, playerStartY, g.lastOverworldX, g.lastOverworldY)
	}

	// Now exit the cave.
	g.ExitCave()

	// Check if player's position is still safe (i.e. not overlapping TileLand).
	// With the fix, since moving up by 0.6*TileSize would overlap TileLand,
	// it should fall back to the original entry coordinates.
	if g.player.Pos.X != playerStartX || g.player.Pos.Y != playerStartY {
		t.Errorf("expected player position to fall back to start position (%f, %f) because of land above, but got (%f, %f)",
			playerStartX, playerStartY, g.player.Pos.X, g.player.Pos.Y)
	}

	// Verify that the player is indeed not stuck (the position is not solid)
	if g.overworldState.IsSolid(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height) {
		t.Errorf("player spawned in a solid/stuck state at %+v", g.player.Pos)
	}

	// Now test the case where the tile above is NOT land (i.e. it's water).
	// It should shift the player up by TileSize*0.6.
	g.world.OverworldMap[tx][ty-1] = world.TileWater
	g.EnterCave(tx, ty)
	g.ExitCave()

	expectedY := playerStartY - config.TileSize*0.6
	if g.player.Pos.Y != expectedY {
		t.Errorf("expected player Y to be shifted to %f, got %f", expectedY, g.player.Pos.Y)
	}
}

func TestCaveVehicleBeacon(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)

	tx, ty := 10, 10
	g.world.OverworldMap[tx][ty] = world.TileWater

	// Place player and enter cave
	g.player.Pos = gvec.Vec2{X: float64(tx * config.TileSize), Y: float64(ty * config.TileSize)}
	g.EnterCave(tx, ty)

	// Spawn a vehicle in this cave
	sub := vehicle.NewScoutSub(100, 100)
	g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], sub)

	// Exit the cave
	g.ExitCave()

	// Assert warning message contains coordinates and that warning is active
	if g.MineWarningTimer <= 0 {
		t.Error("expected MineWarningTimer to be set upon exiting cave with vehicles")
	}
	expectedWarning := "VEHICLE BEACON ACTIVE AT (10, 10)"
	if g.MineWarning != expectedWarning {
		t.Errorf("expected MineWarning to be %q, got %q", expectedWarning, g.MineWarning)
	}

	// Verify that the drawing of the beacon is triggered when drawing the overworld scene
	// Mock drawing to an image to ensure no panics
	screen := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	g.overworldState.Draw(g, screen)
}


