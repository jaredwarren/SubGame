package game

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
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
			p := player.NewPlayer(0, 0)
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
	g.TransitionTo(g.overworldState)
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
	g.TransitionTo(g.overworldState)
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

func TestInventory_AddItem_Atomic(t *testing.T) {
	inv := item.NewInventory(1)
	tit := &item.Titanium{}
	maxStack := tit.GetMaxStack()

	// Fill the slot partially, leaving space for exactly 2 items
	if !inv.AddItem(tit, maxStack-2) {
		t.Fatal("expected successfully adding initial items")
	}

	// Try to add 3 items (exceeds available capacity of 2)
	if inv.AddItem(tit, 3) {
		t.Error("expected AddItem to fail when space is insufficient")
	}

	// Assert inventory was not mutated (still has maxStack-2 items)
	count := inv.Count(tit)
	if count != maxStack-2 {
		t.Errorf("expected inventory to retain original count %d, got %d", maxStack-2, count)
	}

	if !item.HasItem[*item.Titanium](inv, maxStack-2) {
		t.Errorf("expected counts cache to have %d Titanium", maxStack-2)
	}
	if item.HasItem[*item.Titanium](inv, maxStack-1) {
		t.Errorf("expected counts cache not to have %d Titanium", maxStack-1)
	}
}

func TestBaseMenu_OpenClose(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
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
	g.TransitionTo(g.overworldState)
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

func TestBaseMenu_InstallUpgrade(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// Add an upgrade item to the player's inventory
	upg := &item.UpgradeSolar{}
	g.player.Inventory.AddItem(upg, 1)

	// Transition to baseMenu
	g.TransitionTo(g.baseMenu)
	g.baseMenu.ActiveTab = 0

	// Check if upgrade is in player's inventory
	if !item.HasItem[*item.UpgradeSolar](g.player.Inventory, 1) {
		t.Fatal("expected player to have UpgradeSolar")
	}

	// We click on the first slot of the player's inventory in Case 0.
	// Player Inventory is drawn starting at panelX + 45, panelY + 140.
	const (
		panelW = 800
		panelH = 500
	)
	panelX := float64(config.ScreenWidth-panelW) / 2.0
	panelY := float64(config.ScreenHeight-panelH) / 2.0

	mockInput.CursorPos = gvec.Vec2{X: panelX + 65, Y: panelY + 160}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	// Call Update
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Upgrade should now be removed from player inventory and installed in Base Upgrades
	if item.HasItem[*item.UpgradeSolar](g.player.Inventory, 1) {
		t.Errorf("expected UpgradeSolar to be removed from player inventory")
	}
	if !g.baseStation.HasModule(item.ModuleSolar) {
		t.Errorf("expected BaseStation to have ModuleSolar active")
	}

	// Now uninstall it by clicking on the first base upgrade slot.
	// Base upgrades start at panelX + 445, panelY + 140.
	mockInput.CursorPos = gvec.Vec2{X: panelX + 465, Y: panelY + 160}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	// Call Update
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Upgrade should be back in player inventory and uninstalled from Base
	if !item.HasItem[*item.UpgradeSolar](g.player.Inventory, 1) {
		t.Errorf("expected UpgradeSolar to be back in player inventory")
	}
	if g.baseStation.HasModule(item.ModuleSolar) {
		t.Errorf("expected BaseStation to not have ModuleSolar active")
	}
}

func TestBaseMenu_FabricatorScrollAndCraft(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// Transition to baseMenu Fabricator tab
	g.TransitionTo(g.baseMenu)
	g.baseMenu.ActiveTab = 1

	// Scroll down so that we can test scrolling.
	mockInput.WheelY = -2
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Check if ScrollY increased (wy = -2, so ScrollY should increase by -(-2)*15 = 30)
	if g.baseMenu.ScrollY != 30.0 {
		t.Errorf("expected ScrollY to be 30.0, got %f", g.baseMenu.ScrollY)
	}

	// Now reset wheel
	mockInput.WheelY = 0

	// Let's check a craft click.
	// We want to craft the first item (O2TankHC).
	// Let's give player ingredients (4 Titanium, 2 Quartz).
	g.player.Inventory.AddItem(&item.Titanium{}, 4)
	g.player.Inventory.AddItem(&item.Quartz{}, 2)

	// Since we scrolled by 30px, the first recipe visual row is shifted up by 30px.
	// Row Y starts at viewportMinY = startY + 25 = panelY + 120.
	// With ScrollY = 30, the first row (i=0) visual position is:
	// ry = panelY + 120 + 0 - 30 = panelY + 90.
	// Since Y = panelY + 98 is less than panelY + 120, it is outside the viewport and should NOT register click!
	const (
		panelW = 800
		panelH = 500
	)
	panelX := float64(config.ScreenWidth-panelW) / 2.0
	panelY := float64(config.ScreenHeight-panelH) / 2.0
	startX := panelX + 30

	btnX := startX + 560
	btnY := panelY + 98 // visual button Y for scrolled-out first item

	mockInput.CursorPos = gvec.Vec2{X: btnX + 10, Y: btnY + 10}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Player should NOT have crafted it because it was clipped out
	if item.HasItem[*item.O2TankHC](g.player.Inventory, 1) {
		t.Errorf("expected first item to not be crafted because it was scrolled out of viewport")
	}

	// Now scroll back to top (ScrollY = 0)
	mockInput.WheelY = 2
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if g.baseMenu.ScrollY != 0.0 {
		t.Errorf("expected ScrollY to be 0.0 after scroll up, got %f", g.baseMenu.ScrollY)
	}
	mockInput.WheelY = 0

	// Now click first item (O2TankHC) button:
	btnY = panelY + 128
	mockInput.CursorPos = gvec.Vec2{X: btnX + 10, Y: btnY + 10}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Player should have successfully crafted O2TankHC!
	if !item.HasItem[*item.O2TankHC](g.player.Inventory, 1) {
		t.Errorf("expected O2TankHC to be crafted when scrolled into view")
	}
}

func TestPlayer_EquipUnequipUpgrades(t *testing.T) {
	p := player.NewPlayer(0, 0)

	// Initially, player stats should be default
	if item.HasItem[*item.Fins](p.Upgrades, 1) {
		t.Error("expected player to not have fins initially")
	}
	if p.MaxOxygen != 100.0 {
		t.Errorf("expected player initial MaxOxygen to be 100.0, got %f", p.MaxOxygen)
	}

	// Add Fins and High Capacity O2 tank to player inventory
	fins := &item.Fins{}
	tank := &item.O2TankHC{}
	p.Inventory.AddItem(fins, 1)
	p.Inventory.AddItem(tank, 1)

	// Try equipping raw titanium - should fail
	titanium := &item.Titanium{}
	p.Inventory.AddItem(titanium, 1)
	if p.EquipUpgrade(titanium) {
		t.Error("expected equipping titanium to fail")
	}

	// Equip Fins
	if !p.EquipUpgrade(fins) {
		t.Fatal("expected equipping fins to succeed")
	}
	p.Inventory.Remove(fins, 1)

	// HasFins should be true now
	if !item.HasItem[*item.Fins](p.Upgrades, 1) {
		t.Error("expected Fins to be equipped after equipping")
	}

	// Equip O2 Tank
	if !p.EquipUpgrade(tank) {
		t.Fatal("expected equipping High Capacity O2 Tank to succeed")
	}
	p.Inventory.Remove(tank, 1)

	// MaxOxygen should be 160.0 now
	if p.MaxOxygen != 160.0 {
		t.Errorf("expected MaxOxygen to be 160.0 after equipping HC tank, got %f", p.MaxOxygen)
	}

	// Check upgrades inventory contents
	if !item.HasItem[*item.Fins](p.Upgrades, 1) {
		t.Error("expected upgrades inventory to have Fins")
	}
	if !item.HasItem[*item.O2TankHC](p.Upgrades, 1) {
		t.Error("expected upgrades inventory to have High Capacity O2 Tank")
	}

	// Unequip Fins (remove from Upgrades and add back to Inventory)
	p.Upgrades.Remove(fins, 1)
	p.Inventory.AddItem(fins, 1)
	p.RecalculateUpgrades()

	if item.HasItem[*item.Fins](p.Upgrades, 1) {
		t.Error("expected Fins to be not equipped after unequipping")
	}
	if !item.HasItem[*item.Fins](p.Inventory, 1) {
		t.Error("expected Fins to be back in main inventory")
	}
}

func TestPlayer_InventoryClickEquip(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// Open player inventory screen
	g.showInventory = true
	g.ActiveVehicle = nil

	// Add Fins to player inventory
	fins := &item.Fins{}
	g.player.Inventory.AddItem(fins, 1)
	g.player.RecalculateUpgrades() // sync stats

	// Verify player does not have fins equipped initially
	if item.HasItem[*item.Fins](g.player.Upgrades, 1) {
		t.Fatal("expected player to not have fins equipped initially")
	}

	// Click on the first slot of player inventory: (panelX + 65, panelY + 160)
	// panelX = (1280 - 600) / 2 = 340.
	// panelY = (720 - 420) / 2 = 150.
	// startX = panelX + (600 - (8*66-10))/2 = 340 + (600 - 518)/2 = 340 + 41 = 381.
	// startY = panelY + 60 = 150 + 60 = 210.
	// Click center of first slot (r=0, c=0): (startX + 28, startY + 28) = (409, 238).
	mockInput.CursorPos = gvec.Vec2{X: 409, Y: 238}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Verify Fins are removed from main inventory and equipped in upgrades slots
	if g.player.Inventory.Slots[0].Item != nil {
		t.Errorf("expected main inventory slot 0 to be empty after equipping")
	}
	if !item.HasItem[*item.Fins](g.player.Upgrades, 1) {
		t.Errorf("expected player to have fins equipped after click")
	}
	if !item.HasItem[*item.Fins](g.player.Upgrades, 1) {
		t.Errorf("expected upgrades inventory to contain Fins")
	}

	// Now unequip by clicking on the first equipped gear slot:
	// gearStartX = panelX + (600 - 254)/2 = 340 + 173 = 513.
	// gearSlotsY = startY + 3*66 + 5 + 22 = 210 + 198 + 27 = 435.
	// Click center of first gear slot: (gearStartX + 28, gearSlotsY + 28) = (541, 463).
	mockInput.CursorPos = gvec.Vec2{X: 541, Y: 463}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Verify Fins are unequipped (removed from upgrades and added back to main inventory)
	if item.HasItem[*item.Fins](g.player.Upgrades, 1) {
		t.Errorf("expected player to not have fins equipped after unequipping")
	}
	if !item.HasItem[*item.Fins](g.player.Inventory, 1) {
		t.Errorf("expected Fins to be returned to player inventory")
	}
}

func TestTitleScene_Transitions(t *testing.T) {
	// 1. Verify initially in StateTitle
	g := NewGame()
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	if g.currentState != StateTitle {
		t.Errorf("expected initial state to be StateTitle, got %s", g.currentState)
	}
	if g.currentScene != g.titleState {
		t.Errorf("expected initial scene to be titleState, got %+v", g.currentScene)
	}

	// 2. Press Enter to transition to intro
	mockInput.JustPressedKeys[ebiten.KeyEnter] = true
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	if g.currentState != StateIntro {
		t.Errorf("expected state to transition to StateIntro on Enter, got %s", g.currentState)
	}

	// 2.5. Press Enter inside Intro to go to Overworld
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	mockInput.JustPressedKeys[ebiten.KeyEnter] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if g.currentState != StateOverworld {
		t.Errorf("expected state to transition to StateOverworld from Intro on Enter, got %s", g.currentState)
	}

	// 3. Reset back to title scene
	g.TransitionTo(g.titleState)
	if g.currentState != StateTitle {
		t.Errorf("expected state to be StateTitle after resetting, got %s", g.currentState)
	}

	// 4. Click outside the "Dive" button -> should NOT transition
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	mockInput.CursorPos = gvec.Vec2{X: 10, Y: 10}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if g.currentState != StateTitle {
		t.Errorf("expected state to remain StateTitle after clicking outside button, got %s", g.currentState)
	}

	// 5. Click inside the "Dive" button -> should transition to Intro
	// Dive button is at X: 520, Y: 460, W: 240, H: 60
	mockInput.CursorPos = gvec.Vec2{X: 640, Y: 490}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if g.currentState != StateIntro {
		t.Errorf("expected state to transition to StateIntro on button click, got %s", g.currentState)
	}

	// 5.5. Press Enter inside Intro to go to Overworld
	mockInput.JustPressedMouse = make(map[ebiten.MouseButton]bool)
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	mockInput.JustPressedKeys[ebiten.KeyEnter] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}
	if g.currentState != StateOverworld {
		t.Errorf("expected state to transition to StateOverworld from Intro, got %s", g.currentState)
	}
}

func TestPlayer_EquipUpgrade_VehicleKits(t *testing.T) {
	p := player.NewPlayer(0, 0)
	subKit := &vehicle.ScoutSubKit{}
	mechKit := &vehicle.HeavyMechKit{}

	// Attempting to equip vehicle kits as upgrades should return false
	if p.EquipUpgrade(subKit) {
		t.Error("expected equipping ScoutSubKit as player upgrade to return false")
	}
	if p.EquipUpgrade(mechKit) {
		t.Error("expected equipping HeavyMechKit as player upgrade to return false")
	}

	// Verify upgrades inventory is empty
	if item.HasItem[*vehicle.ScoutSubKit](p.Upgrades, 1) || item.HasItem[*vehicle.HeavyMechKit](p.Upgrades, 1) {
		t.Error("expected upgrades inventory to be empty of vehicle kits")
	}
}

func TestTitleScene_SeedInput(t *testing.T) {
	g := NewGame()
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// Verify initially in StateTitle
	if g.currentState != StateTitle {
		t.Fatalf("expected initial state to be StateTitle, got %s", g.currentState)
	}

	// 1. Click on the seed input field to focus it
	// Seed input box is at X: 520, Y: 535, W: 240, H: 40
	mockInput.CursorPos = gvec.Vec2{X: 640, Y: 555}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true
	
	// Update to process focus
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify focus is not transitioned to overworld yet
	if g.currentState != StateTitle {
		t.Fatalf("expected to remain in Title, transitioned to %s", g.currentState)
	}

	// Reset inputs for next frame
	mockInput.JustPressedMouse = make(map[ebiten.MouseButton]bool)

	// 2. Mock typing some characters (e.g. "9876") and then press Enter
	mockInput.InputChars = []rune{'9', '8', '7', '6'}
	mockInput.JustPressedKeys[ebiten.KeyEnter] = true

	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Should transition to StateIntro
	if g.currentState != StateIntro {
		t.Fatalf("expected state to transition to StateIntro, got %s", g.currentState)
	}

	// Press Enter to transition to StateOverworld
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	mockInput.JustPressedKeys[ebiten.KeyEnter] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	if g.currentState != StateOverworld {
		t.Fatalf("expected state to transition to StateOverworld, got %s", g.currentState)
	}

	// The generated world seed should be parsed as 123459876 because "12345" + "9876" = "123459876"
	expectedSeed := int64(123459876)
	if g.world.Seed != expectedSeed {
		t.Errorf("expected world seed to be %d, got %d", expectedSeed, g.world.Seed)
	}
}

func TestTitleScene_SeedInputBackspace(t *testing.T) {
	g := NewGame()
	g.Input = NewMockInput()
	mockInput := g.Input.(*MockInput)

	// Click to focus
	mockInput.CursorPos = gvec.Vec2{X: 640, Y: 555}
	mockInput.JustPressedMouse[ebiten.MouseButtonLeft] = true
	err := g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Press Backspace twice to delete "4" and "5" from "12345"
	mockInput.JustPressedMouse = make(map[ebiten.MouseButton]bool)
	mockInput.JustPressedKeys[ebiten.KeyBackspace] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Backspace again
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// Now press Enter to transition to StateIntro
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	mockInput.JustPressedKeys[ebiten.KeyEnter] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	if g.currentState != StateIntro {
		t.Fatalf("expected state to transition to StateIntro, got %s", g.currentState)
	}

	// Press Enter inside Intro to splash down to Overworld
	mockInput.JustPressedKeys = make(map[ebiten.Key]bool)
	mockInput.JustPressedKeys[ebiten.KeyEnter] = true
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	if g.currentState != StateOverworld {
		t.Fatalf("expected state to transition to StateOverworld, got %s", g.currentState)
	}

	// Seed should be 123
	if g.world.Seed != 123 {
		t.Errorf("expected world seed to be 123, got %d", g.world.Seed)
	}
}

func TestVehicle_PickUp(t *testing.T) {
	g := NewGame()
	
	// Set state to Cave and prepare a trench key
	g.currentState = StateCave
	g.activeTrenchKey = "0_0"
	sub := vehicle.NewScoutSub(100, 100)
	g.CaveVehicles[g.activeTrenchKey] = append(g.CaveVehicles[g.activeTrenchKey], sub)
	g.ActiveVehicle = sub

	// 1. Try picking up sub with cargo inside -> should fail.
	sub.GetCargo().AddItem(&item.Titanium{}, 1)
	g.PickUpActiveVehicle()

	// Active vehicle should still be sub, and inventory shouldn't have sub kit.
	if g.ActiveVehicle != sub {
		t.Error("expected active vehicle to remain sub when cargo is not empty")
	}
	if item.HasItem[*vehicle.ScoutSubKit](g.player.Inventory, 1) {
		t.Error("expected player inventory to not contain ScoutSubKit when pickup failed")
	}

	// 2. Clear cargo and upgrades, but player inventory is full -> should fail.
	sub.GetCargo().Clear()
	if sub.GetUpgrades() != nil {
		sub.GetUpgrades().Clear()
	}
	// Fill player inventory (size is 24)
	for i := 0; i < 24; i++ {
		g.player.Inventory.Slots[i] = item.ItemStack{Item: &item.Titanium{}, Quantity: 10}
	}
	g.PickUpActiveVehicle()
	if g.ActiveVehicle != sub {
		t.Error("expected active vehicle to remain sub when player inventory is full")
	}

	// 3. Free up one slot -> should succeed.
	g.player.Inventory.Slots[0] = item.ItemStack{}
	g.PickUpActiveVehicle()
	
	if g.ActiveVehicle != nil {
		t.Error("expected active vehicle to be nil after successful pickup")
	}
	// Verify sub kit is in inventory
	if !item.HasItem[*vehicle.ScoutSubKit](g.player.Inventory, 1) {
		t.Error("expected player inventory to contain ScoutSubKit after successful pickup")
	}
	// Verify sub is removed from CaveVehicles list
	for _, cv := range g.CaveVehicles[g.activeTrenchKey] {
		if cv == sub {
			t.Error("expected sub to be removed from CaveVehicles list")
		}
	}
}

func TestTutorial_Flow(t *testing.T) {
	g := NewGame()
	g.Input = NewMockInput()
	g.StartGame(12345) // Starts the actual game/tutorial

	// 1. Verify tutorial is active and starts with 9 titanium and no vehicle.
	if !g.TutorialActive {
		t.Fatal("expected tutorial to be active at start")
	}
	if g.ActiveVehicle != nil {
		t.Errorf("expected no starting active vehicle, got: %v", g.ActiveVehicle)
	}
	if len(g.OverworldVehicles) != 0 {
		t.Errorf("expected no overworld vehicles at start, got: %d", len(g.OverworldVehicles))
	}
	titCount := g.player.Inventory.Count(&item.Titanium{})
	if titCount != 9 {
		t.Errorf("expected player to start with 9 Titanium, got: %d", titCount)
	}

	// 2. Check tutorial instruction is correct for the start stage in Overworld
	instr := g.getTutorialInstruction()
	if instr != "Tutorial: Swim away from Lifepod and press [E] to dive (9/10)" {
		t.Errorf("unexpected starting overworld near-base instruction: %q", instr)
	}

	// Move player away from Lifepod
	g.player.Pos = gvec.Vec2{X: g.baseStation.Pos.X + 200, Y: g.baseStation.Pos.Y}
	instr = g.getTutorialInstruction()
	if instr != "Tutorial: Press [E] to dive and find Titanium (9/10)" {
		t.Errorf("unexpected starting overworld far-base instruction: %q", instr)
	}

	// Transition to Cave
	g.TransitionTo(g.caveState)
	instr = g.getTutorialInstruction()
	if instr != "Tutorial: Mine 1 Titanium from the cave wall (9/10)" {
		t.Errorf("unexpected starting cave instruction: %q", instr)
	}

	// 3. Add 1 titanium (bringing count to 10)
	g.player.Inventory.AddItem(&item.Titanium{}, 1)
	instr = g.getTutorialInstruction()
	if instr != "Tutorial: Swim to top of cave (Y < 0) and press [E] to surface" {
		t.Errorf("unexpected cave return instruction: %q", instr)
	}

	// Surface (Overworld)
	g.TransitionTo(g.overworldState)
	instr = g.getTutorialInstruction()
	if instr != "Tutorial: Return to the Lifepod (follow the HUD marker)" {
		t.Errorf("unexpected overworld return instruction: %q", instr)
	}

	// Move near Lifepod
	g.player.Pos = gvec.Vec2{X: g.baseStation.Pos.X, Y: g.baseStation.Pos.Y}
	instr = g.getTutorialInstruction()
	if instr != "Tutorial: Press [E] at the Lifepod terminal to craft a Skiff" {
		t.Errorf("unexpected overworld craft instruction: %q", instr)
	}

	// 4. Craft Skiff Kit
	// Verify Skiff Kit recipe is in the recipes list
	recipes := g.GetCraftingRecipes()
	var skiffRecipe *Recipe
	for i := range recipes {
		if _, ok := recipes[i].NewResult().(*vehicle.SkiffKit); ok {
			skiffRecipe = &recipes[i]
			break
		}
	}
	if skiffRecipe == nil {
		t.Fatal("expected to find Skiff Kit recipe")
	}

	// Simulate crafting click: remove ingredients, add result
	if !g.player.Inventory.Has(&item.Titanium{}, 10) {
		t.Fatal("expected player to have 10 Titanium before craft")
	}
	for _, ing := range skiffRecipe.Ingredients {
		g.player.Inventory.Remove(ing.NewItem(), ing.Quantity)
	}
	g.player.Inventory.AddItem(skiffRecipe.NewResult(), 1)

	if g.player.Inventory.Count(&item.Titanium{}) != 0 {
		t.Errorf("expected Titanium to be consumed after crafting, got %d", g.player.Inventory.Count(&item.Titanium{}))
	}
	if !g.player.Inventory.Has(&vehicle.SkiffKit{}, 1) {
		t.Fatal("expected player to have Skiff Kit in inventory after crafting")
	}

	instr = g.getTutorialInstruction()
	if instr != "Tutorial: Press [TAB] for inventory, click Skiff Kit to deploy it" {
		t.Errorf("unexpected deploy instruction: %q", instr)
	}

	// 5. Deploy Skiff via Hotbar click
	kit := &vehicle.SkiffKit{}
	// Remove from main inventory, put in hotbar
	g.player.Inventory.Remove(kit, 1)
	g.player.Hotbar.Slots[0] = item.ItemStack{Item: kit, Quantity: 1}
	g.player.ActiveSlot = 0

	// Mock left-click
	mInput := g.Input.(*MockInput)
	mInput.JustPressedMouse[ebiten.MouseButtonLeft] = true

	// Update the overworld scene to trigger hotbar deployment
	err := g.overworldState.Update(g)
	if err != nil {
		t.Fatal(err)
	}

	// Verify Skiff is in the world and active vehicle is updated correctly (player needs to press F manually, so it's not active automatically)
	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected 1 overworld vehicle after deployment, got: %d", len(g.OverworldVehicles))
	}
	if _, ok := g.OverworldVehicles[0].(*vehicle.Skiff); !ok {
		t.Error("expected deployed vehicle to be a Skiff")
	}
	if g.player.Hotbar.Has(&vehicle.SkiffKit{}, 1) {
		t.Error("expected Skiff Kit to be removed from hotbar")
	}

	// Run Update to process tutorial completion check
	err = g.Update()
	if err != nil {
		t.Fatal(err)
	}

	// 6. Verify tutorial is complete
	if g.TutorialActive {
		t.Error("expected tutorial to be completed after deploying the Skiff")
	}
}

