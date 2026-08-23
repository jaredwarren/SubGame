package game

import (
	"fmt"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestShockKelpTwoTierTransitions(t *testing.T) {
	g := NewGame()
	g.world.Width = 30
	g.world.Height = 30

	// Set a tile to TileShockKelpCave
	shockTx, shockTy := 15, 15
	g.world.OverworldMap[shockTx][shockTy] = world.TileShockKelpCave

	// Step 1: Dive on the shock kelp tile
	g.EnterCave(shockTx, shockTy)

	if g.currentState != StateCave {
		t.Fatalf("expected current state to be StateCave, got %v", g.currentState)
	}
	if !g.caveState.IsShallow {
		t.Error("expected initial cave layer to be shallow seabed")
	}
	if g.caveState.ActiveCave == nil || g.caveState.ActiveCave.GetCaveType() != cave.CaveOrganicShallow {
		t.Fatalf("expected active cave to be CaveOrganicShallow, got %v", g.caveState.ActiveCave.GetCaveType())
	}

	chasm, ok := g.caveState.ActiveCave.(cave.ChasmProvider)
	if !ok || !chasm.HasFloorChasm() {
		t.Fatal("expected active cave to implement ChasmProvider and have floor chasm")
	}

	minX, maxX, triggerY := chasm.GetChasmBounds()

	// Step 2: Deploy Scout Sub and navigate down into the chasm
	sub := vehicle.NewScoutSub((minX+maxX)/2.0, triggerY-4.0)
	g.ActiveVehicle = sub
	g.CaveVehicles[g.activeTrenchKey] = []vehicle.Vehicle{sub}

	// Trigger boundary update
	g.caveState.Update(g)

	// Verify scroll transition started
	if !g.caveState.IsScrollActive() {
		t.Fatal("expected vertical scroll transition downward to be active")
	}

	// Advance scroll timer to completion (45 ticks)
	for i := 0; i < 45; i++ {
		_ = g.caveState.Update(g)
	}

	// Verify we are now in subterranean Shock Kelp Cave
	if g.caveState.IsScrollActive() {
		t.Fatal("expected scroll transition to have completed")
	}
	if g.caveState.IsShallow {
		t.Error("expected deep Shock Kelp cave to NOT be shallow")
	}
	if g.caveState.ActiveCave.GetCaveType() != cave.CaveShockKelp {
		t.Fatalf("expected deep cave to be CaveShockKelp, got %v", g.caveState.ActiveCave.GetCaveType())
	}

	// Verify vehicle mapping updated to deep trench key
	deepKey := fmt.Sprintf("%d_%d_shock", shockTx, shockTy)
	if g.activeTrenchKey != deepKey {
		t.Fatalf("expected activeTrenchKey to be %q, got %q", deepKey, g.activeTrenchKey)
	}
	if len(g.CaveVehicles[deepKey]) != 1 || g.CaveVehicles[deepKey][0] != sub {
		t.Fatalf("expected Scout Sub to be tracked in CaveVehicles[%q]", deepKey)
	}

	// Step 3: Ascend Scout Sub to ceiling (Y <= 0)
	sub.SetPos(gvec.Vec2{X: float64(len(g.caveState.CaveGrid) / 2 * config.TileSize), Y: -2.0})

	// Trigger boundary update
	_ = g.caveState.Update(g)

	if !g.caveState.IsScrollActive() {
		t.Fatal("expected upward scroll transition to be active")
	}

	// Advance scroll timer to completion
	for i := 0; i < 45; i++ {
		_ = g.caveState.Update(g)
	}

	// Verify we are back in the shallow seabed cave
	if !g.caveState.IsShallow {
		t.Error("expected returned cave to be shallow seabed")
	}
	if g.caveState.ActiveCave.GetCaveType() != cave.CaveOrganicShallow {
		t.Fatalf("expected returned cave to be CaveOrganicShallow, got %v", g.caveState.ActiveCave.GetCaveType())
	}

	shallowKey := fmt.Sprintf("%d_%d", shockTx, shockTy)
	if g.activeTrenchKey != shallowKey {
		t.Fatalf("expected activeTrenchKey to be %q, got %q", shallowKey, g.activeTrenchKey)
	}
	if len(g.CaveVehicles[shallowKey]) != 1 || g.CaveVehicles[shallowKey][0] != sub {
		t.Fatalf("expected Scout Sub to be tracked in CaveVehicles[%q]", shallowKey)
	}

	// Step 4: Dismount sub and swim up past surface (Y < -8)
	g.ActiveVehicle = nil
	g.player.Pos = gvec.Vec2{X: (minX + maxX) / 2.0, Y: -10.0}

	_ = g.caveState.Update(g)

	// Verify we returned to the overworld
	if g.currentState != StateOverworld {
		t.Fatalf("expected current state to be StateOverworld, got %v", g.currentState)
	}
}

func TestRegularShallowCaveNoChasm(t *testing.T) {
	g := NewGame()
	g.world.Width = 30
	g.world.Height = 30

	waterTx, waterTy := 10, 10
	g.world.OverworldMap[waterTx][waterTy] = world.TileWater

	// Dive in regular water tile
	g.EnterCave(waterTx, waterTy)

	if g.currentState != StateCave {
		t.Fatalf("expected current state to be StateCave, got %v", g.currentState)
	}

	// Verify ActiveCave does not provide a chasm
	if chasm, ok := g.caveState.ActiveCave.(cave.ChasmProvider); ok && chasm.HasFloorChasm() {
		t.Errorf("regular shallow cave should NOT have a floor chasm")
	}

	// Move player to bottom floor
	g.player.Pos = gvec.Vec2{X: 1000, Y: 1800}
	_ = g.caveState.Update(g)

	// Verify no scroll transition started
	if g.caveState.IsScrollActive() {
		t.Error("regular shallow cave should NOT trigger scroll transition downward")
	}
	if g.caveState.ActiveCave.GetCaveType() == cave.CaveShockKelp {
		t.Error("regular shallow cave should NOT transition into CaveShockKelp")
	}
}

