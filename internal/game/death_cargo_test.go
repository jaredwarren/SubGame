package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/data"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestDeathDropsCargoKeepsUpgrades(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)

	// Equip fins (stays on death) and carry titanium + hotbar copper.
	fins := &item.Fins{}
	if !g.player.Upgrades.AddItem(fins, 1) {
		t.Fatal("failed to equip fins")
	}
	g.player.RecalculateUpgrades()
	if !g.player.Inventory.AddItem(&item.Titanium{}, 5) {
		t.Fatal("failed to add titanium")
	}
	if !g.player.Hotbar.AddItem(&item.Copper{}, 3) {
		t.Fatal("failed to add copper to hotbar")
	}

	deathPos := gvec.Vec2{X: g.player.Pos.X + 200, Y: g.player.Pos.Y + 50}
	g.player.Pos = deathPos
	g.player.CurrentHealth = 0

	// One update: drop + game over transition.
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.currentState != StateGameOver {
		t.Fatalf("expected GameOver, got %v", g.currentState)
	}
	if !g.player.Inventory.IsEmpty() {
		t.Error("inventory should be empty after cargo drop")
	}
	if !g.player.Hotbar.IsEmpty() {
		t.Error("hotbar should be empty after cargo drop")
	}
	if !item.HasItem[*item.Fins](g.player.Upgrades, 1) {
		t.Error("equipped fins should remain on the player")
	}
	if len(g.lostCargo) != 1 {
		t.Fatalf("expected 1 cargo beacon, got %d", len(g.lostCargo))
	}
	b := g.lostCargo[0]
	if b.Cargo.Count(&item.Titanium{}) != 5 {
		t.Errorf("beacon titanium = %d, want 5", b.Cargo.Count(&item.Titanium{}))
	}
	if b.Cargo.Count(&item.Copper{}) != 3 {
		t.Errorf("beacon copper = %d, want 3", b.Cargo.Count(&item.Copper{}))
	}
	if b.LifetimeTicks != entity.CargoBeaconLifetimeTicks {
		t.Errorf("lifetime = %d, want %d", b.LifetimeTicks, entity.CargoBeaconLifetimeTicks)
	}

	// Second death frame while on game-over must not double-drop.
	g.player.CurrentHealth = 0
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if len(g.lostCargo) != 1 {
		t.Fatalf("expected still 1 beacon after game-over frame, got %d", len(g.lostCargo))
	}

	// Respawn keeps upgrades and leaves cargo at site.
	g.Respawn()
	if !item.HasItem[*item.Fins](g.player.Upgrades, 1) {
		t.Error("fins should still be equipped after respawn")
	}
	if len(g.lostCargo) != 1 {
		t.Fatal("beacon should still exist after respawn")
	}

	// Return to death site and recover.
	g.player.Pos = gvec.Vec2{X: b.Pos.X - g.player.Width/2, Y: b.Pos.Y - g.player.Height/2}
	g.updateLostCargo()
	if len(g.lostCargo) != 0 {
		t.Fatalf("beacon should be fully recovered, still have %d", len(g.lostCargo))
	}
	if g.player.Inventory.Count(&item.Titanium{}) != 5 {
		t.Errorf("recovered titanium = %d, want 5", g.player.Inventory.Count(&item.Titanium{}))
	}
	if g.player.Inventory.Count(&item.Copper{}) != 3 {
		t.Errorf("recovered copper = %d, want 3", g.player.Inventory.Count(&item.Copper{}))
	}
}

func TestCaveDeathDropsCargoOnOverworldDiveSite(t *testing.T) {
	g := NewGame()
	// Simulate having dived from a known surface position into a cave.
	surfaceX := 120.0 * float64(config.TileSize)
	surfaceY := 80.0 * float64(config.TileSize)
	g.lastOverworldX = surfaceX
	g.lastOverworldY = surfaceY
	g.activeTrenchX = 120
	g.activeTrenchY = 80
	g.activeTrenchKey = "120_80"
	g.TransitionTo(g.caveState)
	g.currentState = StateCave

	if !g.player.Inventory.AddItem(&item.Titanium{}, 4) {
		t.Fatal("add cargo failed")
	}
	// Player position deep in the side-scroller cave (irrelevant for surface marker).
	g.player.Pos = gvec.Vec2{X: 9999, Y: 9999}
	g.player.CurrentHealth = 0

	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if len(g.lostCargo) != 1 {
		t.Fatalf("expected cargo on cave death, got %d", len(g.lostCargo))
	}
	b := g.lostCargo[0]
	wantX := surfaceX + g.player.Width/2.0
	wantY := surfaceY + g.player.Height/2.0
	if mathAbs(b.Pos.X-wantX) > 0.5 || mathAbs(b.Pos.Y-wantY) > 0.5 {
		t.Fatalf("cargo should pin to dive site overworld pos got (%.1f,%.1f) want (%.1f,%.1f)",
			b.Pos.X, b.Pos.Y, wantX, wantY)
	}
}

func TestCargoBeaconExpires(t *testing.T) {
	b := entity.NewLostCargoBeacon(gvec.Vec2{}, []item.ItemStack{
		{Item: &item.Titanium{}, Quantity: 1},
	})
	b.LifetimeTicks = 1
	if !b.TickLifetime() {
		t.Fatal("expected beacon to expire after last tick")
	}
}

func TestLostCargoRecoversIntoHotbarWhenInventoryFull(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)
	g.ActiveVehicle = nil
	g.player.Inventory.Clear()
	g.player.Hotbar.Clear()

	for i := range g.player.Inventory.Slots {
		g.player.Inventory.Slots[i] = item.ItemStack{Item: &item.Quartz{}, Quantity: 1}
	}

	beacon := entity.NewLostCargoBeacon(g.player.Pos, []item.ItemStack{
		{Item: &item.Titanium{}, Quantity: 2},
	})
	g.lostCargo = []*entity.LostCargoBeacon{beacon}
	g.player.Pos = gvec.Vec2{X: beacon.Pos.X - g.player.Width/2, Y: beacon.Pos.Y - g.player.Height/2}

	g.updateLostCargo()
	if len(g.lostCargo) != 0 {
		t.Fatalf("expected cargo to recover into hotbar, still have %d beacons", len(g.lostCargo))
	}
	if g.player.Hotbar.Count(&item.Titanium{}) != 2 {
		t.Errorf("hotbar titanium = %d, want 2", g.player.Hotbar.Count(&item.Titanium{}))
	}
}

func TestApplyUnlockedRecipesPrefersNames(t *testing.T) {
	recipes := data.DefaultCraftingRecipes()
	applyUnlockedRecipes(recipes, []string{"Scout Sub Kit"}, []int{4})
	var scoutUnlocked bool
	for _, rcp := range recipes {
		if rcp.ResultName() == "Scout Sub Kit" {
			scoutUnlocked = rcp.Unlocked
		}
	}
	if !scoutUnlocked {
		t.Fatal("expected Scout Sub Kit to unlock by name")
	}
}

func TestLostCargoSaveRoundTrip(t *testing.T) {
	original := []*entity.LostCargoBeacon{
		entity.NewLostCargoBeacon(gvec.Vec2{X: 120, Y: 80}, []item.ItemStack{
			{Item: &item.Titanium{}, Quantity: 4},
			{Item: &item.Copper{}, Quantity: 2},
		}),
	}
	original[0].LifetimeTicks = 999

	restored := deserializeLostCargo(serializeLostCargo(original))
	if len(restored) != 1 {
		t.Fatalf("expected 1 beacon, got %d", len(restored))
	}
	if restored[0].Pos.X != 120 || restored[0].Pos.Y != 80 {
		t.Errorf("pos = %+v, want (120,80)", restored[0].Pos)
	}
	if restored[0].LifetimeTicks != 999 {
		t.Errorf("lifetime = %d, want 999", restored[0].LifetimeTicks)
	}
	if restored[0].Cargo.Count(&item.Titanium{}) != 4 {
		t.Errorf("titanium = %d, want 4", restored[0].Cargo.Count(&item.Titanium{}))
	}
	if restored[0].Cargo.Count(&item.Copper{}) != 2 {
		t.Errorf("copper = %d, want 2", restored[0].Cargo.Count(&item.Copper{}))
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
