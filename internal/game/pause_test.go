package game

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

func TestPause_PreservesOxygenInCave(t *testing.T) {
	g := NewGame()
	mock := NewMockInput()
	g.Input = mock

	// Enter cave and set a specific O2 level
	g.EnterCave(50, 50)
	if g.currentState != StateCave {
		t.Fatalf("expected state to be StateCave, got %v", g.currentState)
	}

	expectedO2 := 45.0
	g.player.CurrentOxygen = expectedO2

	// Press ESC to pause the game
	mock.JustPressedKeys[ebiten.KeyEscape] = true
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}

	if g.currentState != StatePause {
		t.Fatalf("expected state to be StatePause after ESC, got %v", g.currentState)
	}

	if math.Abs(g.player.CurrentOxygen-expectedO2) > 1e-7 {
		t.Fatalf("expected O2 to stay %f on initial pause frame, got %f", expectedO2, g.player.CurrentOxygen)
	}

	// Clear justPressed and advance multiple ticks while paused
	mock.JustPressedKeys = make(map[ebiten.Key]bool)
	for i := 0; i < 30; i++ {
		if err := g.Update(); err != nil {
			t.Fatal(err)
		}
		if g.currentState != StatePause {
			t.Fatalf("expected state to stay StatePause during pause loop, got %v at tick %d", g.currentState, i)
		}
		if math.Abs(g.player.CurrentOxygen-expectedO2) > 1e-7 {
			t.Fatalf("expected O2 to stay %f while paused, got %f at tick %d", expectedO2, g.player.CurrentOxygen, i)
		}
	}

	// Press ESC again to resume
	mock.JustPressedKeys[ebiten.KeyEscape] = true
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}

	if g.currentState != StateCave {
		t.Fatalf("expected state to resume to StateCave, got %v", g.currentState)
	}

	if math.Abs(g.player.CurrentOxygen-expectedO2) > 1e-7 {
		t.Fatalf("expected O2 to remain %f after unpausing, got %f", expectedO2, g.player.CurrentOxygen)
	}
}

func TestPause_ResumeButtonClickPreservesOxygen(t *testing.T) {
	g := NewGame()
	mock := NewMockInput()
	g.Input = mock

	g.EnterCave(50, 50)
	expectedO2 := 33.5
	g.player.CurrentOxygen = expectedO2

	// Pause
	mock.JustPressedKeys[ebiten.KeyEscape] = true
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	mock.JustPressedKeys = make(map[ebiten.Key]bool)

	// Simulate clicking the "RESUME" button
	// Resume button is centered horizontally: centerX = (ScreenWidth - 220)/2, startY = 240, h = 46
	const btnW = 220.0
	centerX := float64(config.ScreenWidth-btnW) / 2.0
	mock.CursorPos.X = centerX + 10
	mock.CursorPos.Y = 250
	mock.JustPressedMouse[ebiten.MouseButtonLeft] = true

	if err := g.Update(); err != nil {
		t.Fatal(err)
	}

	if g.currentState != StateCave {
		t.Fatalf("expected state to resume to StateCave after clicking Resume, got %v", g.currentState)
	}

	if math.Abs(g.player.CurrentOxygen-expectedO2) > 1e-7 {
		t.Fatalf("expected O2 to remain %f after clicking Resume, got %f", expectedO2, g.player.CurrentOxygen)
	}
}

func TestPDA_PreservesOxygenInCave(t *testing.T) {
	g := NewGame()
	mock := NewMockInput()
	g.Input = mock

	g.EnterCave(50, 50)
	expectedO2 := 62.0
	g.player.CurrentOxygen = expectedO2

	// Open PDA via TransitionToPDA
	g.TransitionToPDA()
	if g.currentState != StateBaseMenu {
		t.Fatalf("expected state to be StateBaseMenu after TransitionToPDA, got %v", g.currentState)
	}

	// Advance several frames in PDA menu
	for i := 0; i < 15; i++ {
		if err := g.Update(); err != nil {
			t.Fatal(err)
		}
		if math.Abs(g.player.CurrentOxygen-expectedO2) > 1e-7 {
			t.Fatalf("expected O2 to remain %f while in PDA menu, got %f at frame %d", expectedO2, g.player.CurrentOxygen, i)
		}
	}

	// Close PDA
	g.ClosePDA()
	if g.currentState != StateCave {
		t.Fatalf("expected state to return to StateCave after closing PDA, got %v", g.currentState)
	}

	if math.Abs(g.player.CurrentOxygen-expectedO2) > 1e-7 {
		t.Fatalf("expected O2 to remain %f after closing PDA, got %f", expectedO2, g.player.CurrentOxygen)
	}
}
