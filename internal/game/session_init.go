package game

import (
	"math"

	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/world"
)

// sessionConfig drives the unified new-game bootstrap shared by StartGame and tests.
type sessionConfig struct {
	Seed                  int64
	Tutorial              bool
	GrantStarterInventory bool
}

// initSessionFromSeed resets gameplay state for a new run from a world seed.
// Used by StartGame; loadSaveFromPath uses its own hydration path but shares cache resets.
func (g *Game) initSessionFromSeed(cfg sessionConfig) *world.World {
	w := world.NewWorld(cfg.Seed)
	spawnX, spawnY := findWaterSpawn(w)

	g.world = w
	g.player = player.NewPlayer(spawnX, spawnY)
	if cfg.GrantStarterInventory {
		g.grantStarterInventory()
	}

	g.camera = camera.NewCamera(spawnX, spawnY)
	g.camera.CenterOn(spawnX, spawnY, g.player.Width, g.player.Height)
	g.baseStation = base.NewBaseStation(spawnX+96.0, spawnY-64.0)

	spawnTX := int(math.Floor((spawnX + g.player.Width/2.0) / float64(config.TileSize)))
	spawnTY := int(math.Floor((spawnY + g.player.Height/2.0) / float64(config.TileSize)))

	g.resetNavigation()
	g.resetVehicles()
	g.resetCaveCache()
	g.resetEffects()
	g.resetProgressManagers(w, spawnTX, spawnTY)

	g.TutorialActive = cfg.Tutorial
	g.TimeOfDay = 0
	g.Ticks = 0
	g.showInventory = false
	g.menuOpenedAnywhere = false
	g.pdaPriorState = StateTitle

	if g.baseMenu != nil {
		g.baseMenu.ActiveTab = 0
		g.baseMenu.ScrollY = 0
		g.baseMenu.SelectedLoreIndex = 0
		g.baseMenu.ResetMapCache()
	}

	g.overworldState = NewOverworldScene(w)
	return w
}

// resetSessionCaches clears cave snapshots and vehicles when loading a save
// without rebuilding the whole world.
func (g *Game) resetSessionCaches() {
	g.resetCaveCache()
	g.resetVehicles()
	g.showInventory = false
	g.menuOpenedAnywhere = false
}
