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

func TestAbyssalTrenchTwoTierTransitions(t *testing.T) {
	g := NewGame()
	g.world.Width = 30
	g.world.Height = 30

	trenchTx, trenchTy := 12, 12
	g.world.OverworldMap[trenchTx][trenchTy] = world.TileTrench

	// Step 1: Dive on the Abyssal Trench tile
	g.EnterCave(trenchTx, trenchTy)

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
		t.Fatal("expected active trench shallow cave to implement ChasmProvider and have floor chasm")
	}
	if chasm.GetChasmTarget() != cave.CaveOrganicTrench {
		t.Fatalf("expected chasm target CaveOrganicTrench, got %v", chasm.GetChasmTarget())
	}

	minX, maxX, triggerY := chasm.GetChasmBounds()

	// Step 2: Deploy Scout Sub and descend into the chasm
	sub := vehicle.NewScoutSub((minX+maxX)/2.0, triggerY-4.0)
	g.ActiveVehicle = sub
	g.CaveVehicles[g.activeTrenchKey] = []vehicle.Vehicle{sub}

	// Trigger boundary update
	g.caveState.Update(g)

	if !g.caveState.IsScrollActive() {
		t.Fatal("expected vertical scroll transition downward to be active")
	}

	// Advance scroll timer to completion (45 ticks)
	for i := 0; i < 45; i++ {
		_ = g.caveState.Update(g)
	}

	// Verify we are now in subterranean Organic Trench Cave
	if g.caveState.IsScrollActive() {
		t.Fatal("expected scroll transition to have completed")
	}
	if g.caveState.IsShallow {
		t.Error("expected deep trench cave to NOT be shallow")
	}
	if g.caveState.ActiveCave.GetCaveType() != cave.CaveOrganicTrench {
		t.Fatalf("expected deep cave to be CaveOrganicTrench, got %v", g.caveState.ActiveCave.GetCaveType())
	}

	// Verify vehicle mapping updated to deep trench key
	deepKey := fmt.Sprintf("%d_%d_trench", trenchTx, trenchTy)
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

	shallowKey := fmt.Sprintf("%d_%d", trenchTx, trenchTy)
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

func TestThermoCaveDirectAccess(t *testing.T) {
	// Verify TileThermoCave has Subterranean == nil (direct overworld access)
	info := world.GetTileInfo(world.TileThermoCave)
	if info == nil {
		t.Fatal("expected TileThermoCave info to exist")
	}
	if info.Subterranean != nil {
		t.Errorf("expected TileThermoCave to NOT use SubterraneanSpec (direct access)")
	}

	g := NewGame()
	g.world.Width = 30
	g.world.Height = 30

	thermoTx, thermoTy := 8, 8
	g.world.OverworldMap[thermoTx][thermoTy] = world.TileThermoCave

	g.EnterCave(thermoTx, thermoTy)

	if g.currentState != StateCave {
		t.Fatalf("expected current state to be StateCave, got %v", g.currentState)
	}
	if g.caveState.ActiveCave.GetCaveType() != cave.CaveThermo {
		t.Fatalf("expected active cave to be CaveThermo directly, got %v", g.caveState.ActiveCave.GetCaveType())
	}
}
