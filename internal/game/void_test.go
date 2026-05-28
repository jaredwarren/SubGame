package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestOverworldVoidBoundaries(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.ActiveVehicle = nil // Player on foot
	g.player.Pos = gvec.Vec2{X: -100, Y: -100} // Place player out of bounds (top-left)

	// Verify that the position is solid check returns false (i.e. we can sail past boundary)
	if g.overworldState.isSolid(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height) {
		t.Error("expected out of bounds coordinates to NOT be solid")
	}

	// Update movement: player has velocity
	g.player.Vel = gvec.Vec2{X: -1.0, Y: -1.0}
	g.overworldState.checkCollisions(g.player)

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
	g.player.Pos = gvec.Vec2{X: -10 * TileSize, Y: -10 * TileSize}

	// Dive by calling EnterCave directly with out-of-bounds coordinates
	tx := int(g.player.Pos.X) / TileSize
	ty := int(g.player.Pos.Y) / TileSize
	g.EnterCave(tx, ty)

	// Verify that we are in the cave state
	if g.currentState != StateCave {
		t.Errorf("expected state to transition to StateCave, got %v", g.currentState)
	}

	// Verify that the active cave is VoidCave
	if g.caveState.ActiveCave == nil {
		t.Fatal("expected ActiveCave to be set")
	}
	if g.caveState.ActiveCave.GetCaveType() != CaveVoid {
		t.Errorf("expected ActiveCave type to be CaveVoid, got %v", g.caveState.ActiveCave.GetCaveType())
	}

	// Verify that CaveGrid is nil
	if g.caveState.CaveGrid != nil {
		t.Error("expected CaveGrid to be nil for Void Cave")
	}

	// Verify player starting position
	expectedStartX := float64(30 * TileSize)
	if g.player.Pos.X != expectedStartX {
		t.Errorf("expected player starting X in Void Cave to be %f, got %f", expectedStartX, g.player.Pos.X)
	}

	// Verify that collision check is solid returns false everywhere (allowing infinite movement)
	if g.caveState.isSolid(100, 100, g.player.Width, g.player.Height) {
		t.Error("expected isSolid to return false everywhere in Void Cave")
	}
	if g.caveState.isSolid(-5000, 50000, g.player.Width, g.player.Height) {
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
	g.player.Pos = gvec.Vec2{X: -10 * TileSize, Y: -10 * TileSize}

	// Set camera position to track player in overworld
	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)

	// Mock pressing KeyE
	mockInput := NewMockInput()
	mockInput.PressedKeys[ebiten.KeyE] = true
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

	t.Logf("Draw 1 uniforms: LightSource = %+v, PersonalRadius = %v", g.caveState.uniforms["LightSource"], g.caveState.uniforms["PersonalRadius"])

	// Check if player position is correct (1920, 128)
	if g.player.Pos.X != float64(30 * TileSize) || g.player.Pos.Y != TileSize * 2 {
		t.Errorf("expected player pos to be (1920, 128), got %+v", g.player.Pos)
	}

	// Check if camera centered properly
	expectedCamX := 1920.0 + 10.0 - 640.0 // 1290
	expectedCamY := 128.0 + 10.0 - 360.0  // -222
	if g.camera.Pos.X != expectedCamX || g.camera.Pos.Y != expectedCamY {
		t.Errorf("expected camera pos to be (%f, %f), got %+v", expectedCamX, expectedCamY, g.camera.Pos)
	}

	// Check if LightSource in shader uniforms is centered [640, 360]
	ls := g.caveState.uniforms["LightSource"].([]float32)
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
		ls = g.caveState.uniforms["LightSource"].([]float32)
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
