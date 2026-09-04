package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestDebugMenu_ToggleAndClose(t *testing.T) {
	g := NewGame()
	mock := NewMockInput()
	g.Input = mock

	if g.showDebugMenu {
		t.Fatal("debug menu should start closed")
	}

	// Press ` / ~ to open
	mock.JustPressedKeys[ebiten.KeyGraveAccent] = true
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if !g.showDebugMenu {
		t.Fatal("expected debug menu to be open after KeyGraveAccent")
	}

	// Clear previous frame input and press F1 to toggle closed
	mock.JustPressedKeys = make(map[ebiten.Key]bool)
	mock.JustPressedKeys[ebiten.KeyF1] = true
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.showDebugMenu {
		t.Fatal("expected debug menu to be closed after KeyF1")
	}

	// Toggle open with F3
	mock.JustPressedKeys = make(map[ebiten.Key]bool)
	mock.JustPressedKeys[ebiten.KeyF3] = true
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if !g.showDebugMenu {
		t.Fatal("expected debug menu to be open after KeyF3")
	}

	// Toggle closed with Ctrl+D
	mock.JustPressedKeys = make(map[ebiten.Key]bool)
	mock.PressedKeys[ebiten.KeyControl] = true
	mock.JustPressedKeys[ebiten.KeyD] = true
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.showDebugMenu {
		t.Fatal("expected debug menu to be closed after Ctrl+D")
	}
}

func TestDebugMenu_GiveItemsAndPresets(t *testing.T) {
	g := NewGame()
	g.player.Inventory.Clear()
	g.player.Hotbar.Clear()

	// Give single item
	g.GiveItem("Titanium", 3)
	if g.player.Inventory.Count(&item.Titanium{}) != 3 {
		t.Errorf("expected 3 titanium, got %d", g.player.Inventory.Count(&item.Titanium{}))
	}

	// Give max stack
	g.GiveItem("Copper", 10)
	if g.player.Inventory.Count(&item.Copper{}) != 10 {
		t.Errorf("expected 10 copper, got %d", g.player.Inventory.Count(&item.Copper{}))
	}

	// Give Mini-Lifepod Kit
	g.GiveItem("Mini-Lifepod Kit", 1)
	if !item.HasItem[*vehicle.MiniLifepodKit](g.player.Inventory, 1) {
		t.Error("expected Mini-Lifepod Kit in player inventory from GiveItem")
	}

	// Clear inventory
	g.ClearPlayerInventory()
	if !g.player.Inventory.IsEmpty() {
		t.Error("expected inventory to be empty after clear")
	}

	// Apply Starter Preset
	g.GivePreset("starter")
	if !item.HasItem[*item.Fins](g.player.Upgrades, 1) {
		t.Error("expected fins equipped from starter preset")
	}
	if !item.HasItem[*item.O2TankHC](g.player.Upgrades, 1) {
		t.Error("expected O2 tank equipped from starter preset")
	}
	if !item.HasItem[*item.Scanner](g.player.Hotbar, 1) {
		t.Error("expected scanner in hotbar from starter preset")
	}

	// Apply All Tools Preset
	g.GivePreset("tools")
	if !item.HasItem[*item.RepairTool](g.player.Hotbar, 1) {
		t.Error("expected repair tool in hotbar from tools preset")
	}
	if !item.HasItem[*item.Flashlight](g.player.Hotbar, 1) {
		t.Error("expected flashlight in hotbar from tools preset")
	}

	// Apply Minerals Preset
	g.GivePreset("minerals")
	if g.player.Inventory.Count(&item.Titanium{}) != 10 || g.player.Inventory.Count(&item.AbyssalOre{}) != 10 {
		t.Error("expected 10x all minerals from minerals preset")
	}
}

func TestDebugMenu_PlayerCheats(t *testing.T) {
	g := NewGame()

	// God Mode
	if g.IsGodMode() {
		t.Fatal("god mode should be off by default")
	}
	g.ToggleGodMode()
	if !g.IsGodMode() {
		t.Fatal("god mode should be on after toggle")
	}

	// When in god mode, health stays at max
	g.player.CurrentHealth = 10.0
	g.applyActiveCheats()
	if g.player.CurrentHealth != g.player.MaxHealth {
		t.Errorf("expected god mode to keep health at max, got %.1f", g.player.CurrentHealth)
	}

	// Infinite O2
	g.ToggleInfiniteO2()
	g.player.CurrentOxygen = 5.0
	g.applyActiveCheats()
	if g.player.CurrentOxygen != g.player.MaxOxygen {
		t.Errorf("expected infinite O2 to keep O2 at max, got %.1f", g.player.CurrentOxygen)
	}

	// Infinite Stamina
	g.ToggleInfiniteStamina()
	g.player.CurrentStamina = 5.0
	g.applyActiveCheats()
	if g.player.CurrentStamina != g.player.MaxStamina {
		t.Errorf("expected infinite stamina to keep stamina at max, got %.1f", g.player.CurrentStamina)
	}

	// Super Speed
	g.ToggleSuperSpeed()
	if !g.IsSuperSpeed() {
		t.Fatal("super speed should be on after toggle")
	}
	if g.player.Speed["cave"].TopSpeed <= 2.5 {
		t.Errorf("expected super speed top speed > 2.5, got %.1f", g.player.Speed["cave"].TopSpeed)
	}

	// Refill actions
	g.ToggleGodMode()
	g.player.CurrentHealth = 40.0
	g.HealPlayerFull()
	if g.player.CurrentHealth != g.player.MaxHealth {
		t.Errorf("expected heal full to restore max HP, got %.1f", g.player.CurrentHealth)
	}
}

func TestDebugMenu_Vehicles(t *testing.T) {
	g := NewGame()
	g.currentState = StateCave
	g.activeTrenchKey = "0_0"

	// Spawn Scout Sub
	g.SpawnVehicle("Scout Sub")
	if g.ActiveVehicle == nil || g.ActiveVehicle.GetName() != "Scout Sub" {
		t.Fatalf("expected active Scout Sub, got %v", g.ActiveVehicle)
	}

	// Damage and repair
	g.ActiveVehicle.TakeDamage(50.0)
	if g.ActiveVehicle.GetHealth() >= g.ActiveVehicle.GetMaxHealth() {
		t.Fatal("expected vehicle to have damage")
	}
	g.RepairActiveVehicle()
	if g.ActiveVehicle.GetHealth() != g.ActiveVehicle.GetMaxHealth() {
		t.Errorf("expected vehicle health 100%%, got %.1f", g.ActiveVehicle.GetHealth())
	}

	// Infinite Vehicle Battery cheat
	g.ToggleInfiniteVehicleBattery()
	g.ActiveVehicle.RechargeBattery(-50.0)
	g.applyActiveCheats()
	if g.ActiveVehicle.GetBattery() != g.ActiveVehicle.GetMaxBattery() {
		t.Errorf("expected infinite battery to hold 100%%, got %.1f", g.ActiveVehicle.GetBattery())
	}

	// Despawn
	g.DespawnActiveVehicle()
	if g.ActiveVehicle != nil {
		t.Fatal("expected active vehicle to be nil after despawn")
	}

	// Spawn Mini-Lifepod on surface
	g.currentState = StateOverworld
	g.OverworldVehicles = nil
	g.SpawnVehicle("Mini-Lifepod")
	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected 1 overworld vehicle, got %d", len(g.OverworldVehicles))
	}
	if _, ok := g.OverworldVehicles[0].(*vehicle.MiniLifepod); !ok {
		t.Fatalf("expected spawned vehicle to be *vehicle.MiniLifepod, got %T", g.OverworldVehicles[0])
	}
}

func TestDebugMenu_WorldAndTeleport(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld

	// Time of day
	g.SetTimeOfDay(5400) // Noon
	if g.GetTimeOfDay() != 5400 {
		t.Errorf("expected TOD 5400, got %.0f", g.GetTimeOfDay())
	}
	g.AdvanceTimeOfDay(1.0) // +600 ticks
	if g.GetTimeOfDay() != 6000 {
		t.Errorf("expected TOD 6000, got %.0f", g.GetTimeOfDay())
	}

	// Freeze time
	g.ToggleFreezeTime()
	g.advanceTimers()
	if g.GetTimeOfDay() != 6000 {
		t.Errorf("expected time to be frozen at 6000, got %.0f", g.GetTimeOfDay())
	}

	// Teleport to Life Pod
	g.TeleportToLifePod()
	if g.player.Pos.X != g.baseStation.Pos.X {
		t.Errorf("player X = %.1f, want %.1f", g.player.Pos.X, g.baseStation.Pos.X)
	}

	// Teleport to POI
	g.TeleportToPOI(world.TileTrench)
	if g.currentState != StateOverworld {
		t.Errorf("expected overworld state after POI warp, got %v", g.currentState)
	}

	// Reveal full map
	g.RevealFullMap()
	if !g.explorationTracker.IsExplored(10, 10) || !g.explorationTracker.IsExplored(40, 40) {
		t.Error("expected map tiles to be explored after RevealFullMap")
	}
}

func TestDebugMenu_QuestsAndLore(t *testing.T) {
	g := NewGame()

	// Unlock all recipes
	g.UnlockAllRecipes()
	for _, rcp := range g.craftingRecipes {
		if !rcp.Unlocked {
			t.Errorf("expected recipe %s to be unlocked", rcp.ResultName())
		}
	}

	// Complete current task and quest
	g.CompleteCurrentTask()
	g.CompleteCurrentQuest()

	// Complete all quests
	g.CompleteAllQuests()
	for _, cat := range g.questManager.Categories {
		for _, q := range cat.Quests {
			if !q.Completed {
				t.Errorf("expected quest %s to be completed", q.ID)
			}
		}
	}

	// Reset quests
	g.ResetAllQuests()
	if len(g.questManager.Categories) == 0 {
		t.Fatal("expected reset quest manager to have categories")
	}
}
