package game

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPlayer_UpdateStats(t *testing.T) {
	tests := []struct {
		name         string
		inCave       string
		isSprinting  bool
		vel          Vec2
		initialO2    float64
		initialHp    float64
		initialSt    float64
		expectedO2   float64
		expectedHp   float64
		expectedSt   float64
	}{
		{
			name:        "O2 depletes in cave",
			inCave:      "true",
			isSprinting: false,
			vel:         Vec2{},
			initialO2:   100.0,
			initialHp:   100.0,
			initialSt:   100.0,
			// O2DrainRate is 1.0 per second. At 60 FPS, updates by 1/60.
			expectedO2:   100.0 - (1.0 / 60.0),
			expectedHp:   100.0,
			expectedSt:   100.0,
		},
		{
			name:        "O2 refills on surface",
			inCave:      "false",
			isSprinting: false,
			vel:         Vec2{},
			initialO2:   50.0,
			initialHp:   100.0,
			initialSt:   100.0,
			expectedO2:   100.0, // immediately refilled
			expectedHp:   100.0,
			expectedSt:   100.0,
		},
		{
			name:        "Drowning damage applied when O2 is 0",
			inCave:      "true",
			isSprinting: false,
			vel:         Vec2{},
			initialO2:   0.0,
			initialHp:   100.0,
			initialSt:   100.0,
			expectedO2:   0.0,
			// DrownDamageRate is 30.0 per second. At 60 FPS, updates by 30/60 = 0.5.
			expectedHp:   99.5,
			expectedSt:   100.0,
		},
		{
			name:        "Stamina depletes when sprinting and moving",
			inCave:      "false",
			isSprinting: true,
			vel:         Vec2{X: 1.0, Y: 0.0},
			initialO2:   100.0,
			initialHp:   100.0,
			initialSt:   100.0,
			expectedO2:   100.0,
			expectedHp:   100.0,
			// StaminaDrainRate is 1.5 per second. At 60 FPS, updates by 1.5/60 = 0.025.
			expectedSt:   99.975,
		},
		{
			name:        "Stamina regens when idle",
			inCave:      "false",
			isSprinting: false,
			vel:         Vec2{},
			initialO2:   100.0,
			initialHp:   100.0,
			initialSt:   50.0,
			expectedO2:   100.0,
			expectedHp:   100.0,
			// StaminaRegenRate is 1.0 per second. At 60 FPS, updates by 1/60.
			expectedSt:   50.0 + (1.0 / 60.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlayer(0, 0)
			p.CurrentOxygen = tt.initialO2
			p.CurrentHealth = tt.initialHp
			p.CurrentStamina = tt.initialSt
			p.Vel = tt.vel

			inCaveBool := tt.inCave == "true"
			p.UpdateStats(inCaveBool, tt.isSprinting)

			if math.Abs(p.CurrentOxygen-tt.expectedO2) > 1e-7 {
				t.Errorf("expected O2 %f, got %f", tt.expectedO2, p.CurrentOxygen)
			}
			if math.Abs(p.CurrentHealth-tt.expectedHp) > 1e-7 {
				t.Errorf("expected HP %f, got %f", tt.expectedHp, p.CurrentHealth)
			}
			if math.Abs(p.CurrentStamina-tt.expectedSt) > 1e-7 {
				t.Errorf("expected Stamina %f, got %f", tt.expectedSt, p.CurrentStamina)
			}
		})
	}
}

func TestPlayer_Movement(t *testing.T) {
	g := NewGame()
	g.Input = NewMockInput()

	// Inject keypress mock input: D key to move right
	mockInput := g.Input.(*MockInput)
	mockInput.PressedKeys[ebiten.KeyD] = true

	// Ensure ActiveVehicle is nil so the player is swimming on foot
	g.ActiveVehicle = nil

	// Update player movement via OverworldScene
	startX := g.player.Pos.X
	err := g.overworldState.Update(g)
	if err != nil {
		t.Fatal(err)
	}

	// Player should have positive X velocity and position X should have increased
	if g.player.Vel.X <= 0 {
		t.Errorf("expected player Vel.X to be positive, got %f", g.player.Vel.X)
	}
	if g.player.Pos.X <= startX {
		t.Errorf("expected player Pos.X to be greater than startX (%f), got %f", startX, g.player.Pos.X)
	}
}
