package game

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestPlayer_UpdateStats(t *testing.T) {
	tests := []struct {
		name        string
		inCave      string
		isSprinting bool
		vel         gvec.Vec2
		initialO2   float64
		initialHp   float64
		initialSt   float64
		expectedO2  float64
		expectedHp  float64
		expectedSt  float64
	}{
		{
			name:        "O2 depletes in cave",
			inCave:      "true",
			isSprinting: false,
			vel:         gvec.Vec2{},
			initialO2:   100.0,
			initialHp:   100.0,
			initialSt:   100.0,
			// O2DrainRate is 1.0 per second. At 60 FPS, updates by 1/60.
			expectedO2: 100.0 - (1.0 / 60.0),
			expectedHp: 100.0,
			expectedSt: 100.0,
		},
		{
			name:        "O2 refills on surface",
			inCave:      "false",
			isSprinting: false,
			vel:         gvec.Vec2{},
			initialO2:   50.0,
			initialHp:   100.0,
			initialSt:   100.0,
			expectedO2:  100.0, // immediately refilled
			expectedHp:  100.0,
			expectedSt:  100.0,
		},
		{
			name:        "Drowning damage applied when O2 is 0",
			inCave:      "true",
			isSprinting: false,
			vel:         gvec.Vec2{},
			initialO2:   0.0,
			initialHp:   100.0,
			initialSt:   100.0,
			expectedO2:  0.0,
			// DrownDamageRate is 30.0 per second. At 60 FPS, updates by 30/60 = 0.5.
			expectedHp: 99.5,
			expectedSt: 100.0,
		},
		{
			name:        "Stamina depletes when sprinting and moving",
			inCave:      "false",
			isSprinting: true,
			vel:         gvec.Vec2{X: 1.0, Y: 0.0},
			initialO2:   100.0,
			initialHp:   100.0,
			initialSt:   100.0,
			expectedO2:  100.0,
			expectedHp:  100.0,
			// StaminaDrainRate is 1.5 per second. At 60 FPS, updates by 1.5/60 = 0.025.
			expectedSt: 99.975,
		},
		{
			name:        "Stamina regens when idle",
			inCave:      "false",
			isSprinting: false,
			vel:         gvec.Vec2{},
			initialO2:   100.0,
			initialHp:   100.0,
			initialSt:   50.0,
			expectedO2:  100.0,
			expectedHp:  100.0,
			// StaminaRegenRate is 1.0 per second. At 60 FPS, updates by 1/60.
			expectedSt: 50.0 + (1.0 / 60.0),
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

func TestVehicle_EntryExit(t *testing.T) {
	g := NewGame()
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// In NewGame, player starts inside a skiff, so ActiveVehicle is the skiff.
	if g.ActiveVehicle == nil {
		t.Fatal("expected player to start inside a vehicle (skiff)")
	}

	skiff := g.ActiveVehicle

	// Mock F key pressed to exit vehicle
	mockInput.JustPressedKeys[ebiten.KeyF] = true

	// Call Game.Update
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Now ActiveVehicle should be nil
	if g.ActiveVehicle != nil {
		t.Errorf("expected ActiveVehicle to be nil after exit, got %+v", g.ActiveVehicle)
	}

	// Verify player position was offset on exit (in overworld: vPos.X - 24)
	expectedX := skiff.GetPos().X - 24
	if g.player.Pos.X != expectedX {
		t.Errorf("expected player Pos.X to be %f, got %f", expectedX, g.player.Pos.X)
	}

	// Now let's try to enter the skiff again.
	// First, let's place the player close to the skiff.
	g.player.Pos = gvec.Vec2{X: skiff.GetPos().X + 5, Y: skiff.GetPos().Y + 5}

	// Reset inputs for the next frame
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	mockInput.JustPressedKeys[ebiten.KeyF] = true

	// Call Game.Update (which will process entry)
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Now ActiveVehicle should be back to skiff
	if g.ActiveVehicle != skiff {
		t.Errorf("expected player to enter skiff, ActiveVehicle is %+v", g.ActiveVehicle)
	}
}

func TestInventory_AddItem(t *testing.T) {
	inv := item.NewInventory(5)

	// Test adding a raw Titanium item
	if !inv.AddItem(&item.Titanium{}, 3) {
		t.Error("expected successfully adding Titanium")
	}
	if !item.HasItem[*item.Titanium](inv, 3) {
		t.Errorf("expected inventory to have 3 Titanium")
	}

	// Test adding a ResourceNode (implements Item via Resource)
	node := resource.NewCopperNode(10, 10)
	if !inv.AddItem(node, 2) {
		t.Error("expected successfully adding Copper resource node")
	}
	if !item.HasItem[*resource.CopperNode](inv, 2) {
		t.Errorf("expected inventory to have 2 Copper resource nodes")
	}
}

func TestBaseMenu_OpenClose(t *testing.T) {
	g := NewGame()
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// Exit active vehicle so player is swimming
	g.ActiveVehicle = nil

	// Place player close to the base station
	g.player.Pos = gvec.Vec2{
		X: g.baseStation.Pos.X + g.baseStation.Size.X/2.0 - g.player.Width/2.0,
		Y: g.baseStation.Pos.Y + g.baseStation.Size.Y/2.0 - g.player.Height/2.0 + 30, // 30px below base center
	}

	// Verify distance is less than 100.0
	dist := g.baseStation.DistanceToPlayer(g.player)
	if dist >= 100.0 {
		t.Fatalf("expected player to be near base, got distance %f", dist)
	}

	// Press E
	mockInput.JustPressedKeys[ebiten.KeyE] = true

	// Call Update
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Verify scene transitioned to baseMenu
	if g.currentScene != g.baseMenu {
		t.Errorf("expected current scene to be baseMenu, got %+v", g.currentScene)
	}

	// In the next frame, E should NOT immediately close the menu (the frame-double-trigger fix)
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if g.currentScene != g.baseMenu {
		t.Errorf("expected current scene to remain baseMenu, got %+v", g.currentScene)
	}

	// Now press E in the base menu to close it
	mockInput.JustPressedKeys[ebiten.KeyE] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if g.currentScene != g.overworldState {
		t.Errorf("expected current scene to transition back to overworldState, got %+v", g.currentScene)
	}
}

func TestBaseMenu_OpenCloseFromVehicle(t *testing.T) {
	g := NewGame()
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// Player is in skiff (the default active vehicle)
	if g.ActiveVehicle == nil {
		t.Fatal("expected player to start in a vehicle")
	}

	// Position active vehicle (skiff) close to base station
	skiff := g.ActiveVehicle
	skiff.SetPos(gvec.Vec2{
		X: g.baseStation.Pos.X + g.baseStation.Size.X/2.0 - skiff.GetDimensions().X/2.0,
		Y: g.baseStation.Pos.Y + g.baseStation.Size.Y/2.0 - skiff.GetDimensions().Y/2.0 + 40,
	})

	// Run update to sync player position inside vehicle
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Check distance is less than 100.0
	dist := g.baseStation.DistanceToPlayer(g.player)
	if dist >= 100.0 {
		t.Fatalf("expected vehicle/player to be near base, got distance %f", dist)
	}

	// Press E
	mockInput.JustPressedKeys[ebiten.KeyE] = true

	// Call Update
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Verify scene transitioned to baseMenu
	if g.currentScene != g.baseMenu {
		t.Errorf("expected current scene to be baseMenu, got %+v", g.currentScene)
	}
}

