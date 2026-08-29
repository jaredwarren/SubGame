package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestRespawnTeleportsSkiffNearLifePod(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)

	farX := g.baseStation.Pos.X + 800
	farY := g.baseStation.Pos.Y + 800
	skiff := vehicle.NewSkiff(farX, farY)
	g.OverworldVehicles = []vehicle.Vehicle{skiff}

	g.player.CurrentHealth = 0
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.currentState != StateGameOver {
		t.Fatalf("expected GameOver, got %v", g.currentState)
	}

	g.Respawn()

	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected skiff to still exist, got %d vehicles", len(g.OverworldVehicles))
	}
	respawned, ok := g.OverworldVehicles[0].(*vehicle.Skiff)
	if !ok {
		t.Fatalf("expected overworld vehicle to be a skiff, got %T", g.OverworldVehicles[0])
	}
	assertSkiffNearLifePod(t, g, respawned)
}

func TestRespawnTeleportsSkiffWhenPlayerDiedInSkiff(t *testing.T) {
	g := NewGame()
	g.TransitionTo(g.overworldState)

	skiff := vehicle.NewSkiff(g.baseStation.Pos.X+500, g.baseStation.Pos.Y+500)
	g.OverworldVehicles = []vehicle.Vehicle{skiff}
	g.ActiveVehicle = skiff

	g.player.CurrentHealth = 0
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.currentState != StateGameOver {
		t.Fatalf("expected GameOver, got %v", g.currentState)
	}

	g.Respawn()

	if g.ActiveVehicle != nil {
		t.Fatal("expected active vehicle cleared after respawn")
	}
	if len(g.OverworldVehicles) != 1 {
		t.Fatalf("expected skiff to still exist, got %d vehicles", len(g.OverworldVehicles))
	}
	respawned, ok := g.OverworldVehicles[0].(*vehicle.Skiff)
	if !ok {
		t.Fatalf("expected overworld vehicle to be a skiff, got %T", g.OverworldVehicles[0])
	}
	assertSkiffNearLifePod(t, g, respawned)
}

func assertSkiffNearLifePod(t *testing.T, g *Game, skiff *vehicle.Skiff) {
	t.Helper()

	pos := skiff.GetPos()
	dims := skiff.GetDimensions()
	if !g.isClearOverworldDeploy(pos, dims) {
		t.Fatalf("skiff respawned into blocked space at %+v", pos)
	}

	near := gvec.Vec2{X: g.baseStation.Pos.X, Y: g.baseStation.Pos.Y + 64.0}
	cx := pos.X + dims.X/2
	cy := pos.Y + dims.Y/2
	dist := hypot(cx-near.X, cy-near.Y)
	if dist > float64(config.TileSize)*6 {
		t.Fatalf("skiff too far from life pod: dist=%.1f", dist)
	}
	if skiff.Vel.X != 0 || skiff.Vel.Y != 0 {
		t.Fatalf("expected skiff velocity cleared, got %+v", skiff.Vel)
	}
}
