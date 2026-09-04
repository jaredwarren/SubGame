package game

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/data"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/save"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestMiniLifepod_CraftingRecipe(t *testing.T) {
	recipes := data.DefaultCraftingRecipes()
	var found *data.Recipe
	for i := range recipes {
		if recipes[i].ResultName() == "Mini-Lifepod Kit" {
			found = &recipes[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected Mini-Lifepod Kit recipe in DefaultCraftingRecipes")
	}
	if found.Tier != 2 {
		t.Errorf("got recipe tier %d, want 2", found.Tier)
	}
	if found.Unlocked {
		t.Error("expected recipe to be locked initially")
	}

	ingCount := map[string]int{}
	for _, ing := range found.Ingredients {
		it := ing.NewItem()
		ingCount[it.GetName()] = ing.Quantity
	}
	if ingCount["Titanium"] != 8 {
		t.Errorf("Titanium count = %d, want 8", ingCount["Titanium"])
	}
	if ingCount["Nickel"] != 4 {
		t.Errorf("Nickel count = %d, want 4", ingCount["Nickel"])
	}
	if ingCount["Power Cell"] != 2 {
		t.Errorf("Power Cell count = %d, want 2", ingCount["Power Cell"])
	}
	if ingCount["Quartz"] != 4 {
		t.Errorf("Quartz count = %d, want 4", ingCount["Quartz"])
	}
}

func TestMiniLifepod_SurfaceDeploymentAndCavePrevention(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil
	g.ActiveVehicle = nil

	// Place player on clear water in overworld
	waterPos, ok := g.overworldState.FindNearestWater(gvec.Vec2{X: 300, Y: 300}, 24, 24, g.baseStation)
	if !ok {
		t.Fatal("could not find clear water in test world")
	}
	g.player.Pos = waterPos

	// 1. Deploy Mini-Lifepod Kit on surface
	kit := &vehicle.MiniLifepodKit{}
	g.player.Inventory.AddItem(kit, 1)
	g.ActivatePlayerItem(kit)

	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected 1 overworld vehicle after deploy, got %d", len(g.OverworldVehicles))
	}
	pod, ok := g.OverworldVehicles[0].(*vehicle.MiniLifepod)
	if !ok {
		t.Fatalf("expected vehicle in OverworldVehicles to be *vehicle.MiniLifepod, got %T", g.OverworldVehicles[0])
	}
	if pod.GetID() != vehicle.VehicleMiniLifepod {
		t.Errorf("got vehicle ID %s, want %s", pod.GetID(), vehicle.VehicleMiniLifepod)
	}

	// 2. Attempt to deploy Mini-Lifepod inside cave
	g.currentState = StateCave
	g.activeTrenchKey = "0,0"
	caveKit := &vehicle.MiniLifepodKit{}
	g.player.Inventory.AddItem(caveKit, 1)
	g.ActivatePlayerItem(caveKit)

	if len(g.CaveVehicles[g.activeTrenchKey]) != 0 {
		t.Errorf("expected Mini-Lifepod NOT to deploy in caves, got %d vehicles in cave", len(g.CaveVehicles[g.activeTrenchKey]))
	}
	if !item.HasItem[*vehicle.MiniLifepodKit](g.player.Inventory, 1) {
		t.Error("expected MiniLifepodKit to remain in player inventory when cave deploy is rejected")
	}
}

func TestMiniLifepod_SkiffDockingAndRecall(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil

	skiff := vehicle.NewSkiff(400, 400)
	g.OverworldVehicles = append(g.OverworldVehicles, skiff)
	g.ActiveVehicle = skiff

	// Dock a Mini-Lifepod in Skiff Bay 2
	pod := vehicle.NewMiniLifepod(0, 0)
	pod.Battery = 50.0
	bayIdx, ok := skiff.Dock(pod)
	if !ok || bayIdx != 2 {
		t.Fatalf("expected pod to dock in bay 2, got %d (ok=%v)", bayIdx, ok)
	}

	// 1. Deploy Mini-Lifepod from Skiff onto surface
	g.DeploySubFromSkiff(2)

	if len(g.OverworldVehicles) != 2 {
		t.Fatalf("expected 2 overworld vehicles (Skiff + MiniLifepod), got %d", len(g.OverworldVehicles))
	}
	if skiff.GetDocked(2) != nil {
		t.Error("expected Skiff Bay 2 to be empty after deploy")
	}

	// 2. Recall Mini-Lifepod back into Skiff dock
	g.WinchRecallSub(2)

	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected 1 overworld vehicle after winch recall, got %d", len(g.OverworldVehicles))
	}
	docked := skiff.GetDocked(2)
	if docked == nil || docked.ID != vehicle.VehicleMiniLifepod {
		t.Fatalf("expected Bay 2 to hold VehicleMiniLifepod, got %+v", docked)
	}
}

func TestMiniLifepod_FieldInteractionAndPackUp(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil
	g.ActiveVehicle = nil

	// Place player far from main Life Pod so proximity check tests Mini-Lifepod
	g.baseStation.Pos = gvec.Vec2{X: 1000, Y: 1000}
	waterPos, ok := g.overworldState.FindNearestWater(gvec.Vec2{X: 300, Y: 300}, 24, 24, g.baseStation)
	if !ok {
		t.Fatal("could not find clear water in test world")
	}
	g.player.Pos = waterPos

	pod := vehicle.NewMiniLifepod(waterPos.X+10, waterPos.Y+10)
	pod.Battery = 60.0
	g.OverworldVehicles = append(g.OverworldVehicles, pod)

	// Verify canEnterLifePodNearby is true near the Mini-Lifepod
	if !g.canEnterLifePodNearby() {
		t.Error("expected canEnterLifePodNearby to be true when near Mini-Lifepod")
	}

	// Press E to enter Mini-Lifepod Terminal
	mockInput := NewMockInput()
	mockInput.JustPressedKeys[ebiten.KeyE] = true
	g.Input = mockInput
	g.handleInput()

	if g.currentState != StateBaseMenu {
		t.Fatalf("expected transition to StateBaseMenu, got %v", g.currentState)
	}
	if g.activeMiniLifepod != pod {
		t.Fatalf("expected activeMiniLifepod to be pod, got %v", g.activeMiniLifepod)
	}
	if !g.CanPackUpActiveBase() {
		t.Error("expected CanPackUpActiveBase to be true")
	}
	if g.GetActiveBaseStationName() != "FIELD OUTPOST - MINI-LIFEPOD" {
		t.Errorf("got terminal name %q, want 'FIELD OUTPOST - MINI-LIFEPOD'", g.GetActiveBaseStationName())
	}

	// Install Solar Array module into Mini-Lifepod
	activeBase := g.GetBaseStation()
	if activeBase == nil {
		t.Fatal("expected GetBaseStation to return non-nil station for MiniLifepod")
	}
	if !activeBase.InstallUpgrade(&item.UpgradeSolar{}) {
		t.Error("expected Solar Array upgrade installation to succeed on Mini-Lifepod")
	}

	// Pack Up Pod
	g.PackUpActiveBase()

	if g.currentState != StateOverworld {
		t.Errorf("expected transition back to StateOverworld after Pack Up, got %v", g.currentState)
	}
	if len(g.OverworldVehicles) != 0 {
		t.Errorf("expected OverworldVehicles to be empty after Pack Up, got %d", len(g.OverworldVehicles))
	}
	if !item.HasItem[*vehicle.MiniLifepodKit](g.player.Inventory, 1) {
		t.Error("expected player inventory to contain 1 MiniLifepodKit")
	}

	// Redeploy the kit and verify the installed Solar module was preserved
	g.player.Pos = waterPos

	var redeployKit *vehicle.MiniLifepodKit
	for _, slot := range g.player.Inventory.Slots {
		if k, ok := slot.Item.(*vehicle.MiniLifepodKit); ok {
			redeployKit = k
			break
		}
	}
	if redeployKit == nil {
		t.Fatal("expected to find MiniLifepodKit in inventory")
	}
	g.ActivatePlayerItem(redeployKit)

	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected 1 overworld vehicle after redeploy, got %d", len(g.OverworldVehicles))
	}
	repositionedPod := g.OverworldVehicles[0].(*vehicle.MiniLifepod)
	if !repositionedPod.HasModule(item.ModuleSolar) {
		t.Error("expected redeployed Mini-Lifepod to retain installed Solar module")
	}
}

func TestMiniLifepod_SaveAndLoadPersistence(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil

	// Deploy Skiff with Mini-Lifepod in Bay 2 and a separate deployed Mini-Lifepod in ocean
	skiff := vehicle.NewSkiff(300, 300)
	dockedPod := vehicle.NewMiniLifepod(0, 0)
	dockedPod.Battery = 45.0
	dockedPod.Upgrades.AddItem(&item.UpgradeSolar{}, 1)
	skiff.Dock(dockedPod)

	deployedPod := vehicle.NewMiniLifepod(500, 500)
	deployedPod.Battery = 65.0

	g.OverworldVehicles = append(g.OverworldVehicles, skiff, deployedPod)

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "save_slot_1.json")

	saveData := save.SaveData{
		Version:   save.CurrentSaveVersion,
		WorldSeed: 42,
		Player: save.SavedPlayer{
			PosX:      100,
			PosY:      100,
			Health:    100,
			MaxHealth: 100,
			Inventory: g.player.Inventory.SerializeState(),
			Upgrades:  g.player.Upgrades.SerializeState(),
			Hotbar:    g.player.Hotbar.SerializeState(),
		},
		BaseStation: save.SavedBaseStation{
			PosX:     g.baseStation.Pos.X,
			PosY:     g.baseStation.Pos.Y,
			Power:    g.baseStation.Power,
			MaxPower: g.baseStation.MaxPower,
			Storage:  g.baseStation.Storage.SerializeState(),
			Upgrades: g.baseStation.Upgrades.SerializeState(),
		},
		Vehicles: []save.SavedVehicle{
			serializeVehicle(skiff, "overworld", false),
			serializeVehicle(deployedPod, "overworld", false),
		},
	}

	if err := save.SaveToFile(savePath, &saveData); err != nil {
		t.Fatalf("failed to save test state: %v", err)
	}

	// Create a new game instance and restore from file
	g2 := NewGame()
	if err := g2.loadSaveFromPath(savePath); err != nil {
		t.Fatalf("failed to load save: %v", err)
	}

	if len(g2.OverworldVehicles) != 2 {
		t.Fatalf("expected 2 overworld vehicles restored, got %d", len(g2.OverworldVehicles))
	}

	var restoredSkiff *vehicle.Skiff
	var restoredPod *vehicle.MiniLifepod
	for _, v := range g2.OverworldVehicles {
		if s, ok := v.(*vehicle.Skiff); ok {
			restoredSkiff = s
		}
		if p, ok := v.(*vehicle.MiniLifepod); ok {
			restoredPod = p
		}
	}

	if restoredSkiff == nil {
		t.Fatal("expected restored Skiff")
	}
	if restoredPod == nil {
		t.Fatal("expected restored deployed MiniLifepod")
	}
	if math.Abs(restoredPod.Battery-65.0) > 0.1 {
		t.Errorf("restored deployed pod battery = %.1f, want 65.0", restoredPod.Battery)
	}

	docked := restoredSkiff.GetDocked(2)
	if docked == nil || docked.ID != vehicle.VehicleMiniLifepod {
		t.Fatalf("expected restored Skiff to hold VehicleMiniLifepod in Bay 2, got %+v", docked)
	}
	if math.Abs(docked.Battery-45.0) > 0.1 {
		t.Errorf("restored docked pod battery = %.1f, want 45.0", docked.Battery)
	}
	if docked.Upgrades == nil || !item.HasItem[*item.UpgradeSolar](docked.Upgrades, 1) {
		t.Error("expected docked pod to retain installed Solar module")
	}
}
