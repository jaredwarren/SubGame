package game

import (
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/game/sonar"
	"github.com/jaredwarren/SubGame/internal/game/story"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

// compile-time assertion: *Game must satisfy scene.GameContext
var _ scene.GameContext = (*Game)(nil)

// --- Scene navigation ---

func (g *Game) StartGame(seed int64) {
	w := world.NewWorld(seed)
	g.world = w

	spawnX, spawnY := findWaterSpawn(w)

	g.player = player.NewPlayer(spawnX, spawnY)
	g.camera = camera.NewCamera(spawnX, spawnY)
	g.camera.CenterOn(spawnX, spawnY, g.player.Width, g.player.Height)

	g.baseStation = base.NewBaseStation(spawnX+96.0, spawnY-64.0)
	skiff := vehicle.NewSkiff(spawnX, spawnY)

	g.ActiveVehicle = skiff
	g.OverworldVehicles = []vehicle.Vehicle{skiff}
	g.CaveVehicles = make(map[string][]vehicle.Vehicle)
	g.caveNodes = make(map[string][]resource.Resource)
	g.caveEntities = make(map[string][]entity.CaveEntity)
	g.Sonar = sonar.NewSonar()

	// Reset navigation and progression state
	g.lastOverworldX = 0
	g.lastOverworldY = 0
	g.activeTrenchX = 0
	g.activeTrenchY = 0
	g.activeTrenchKey = ""
	g.justExited = false
	g.showInventory = false
	g.TimeOfDay = 0
	g.Ticks = 0
	g.WeaverTrackingTimer = 0
	g.SoundWaveTimer = 0
	g.SoundWaveRadius = 0
	g.SoundWaveX = 0
	g.SoundWaveY = 0
	g.playerSlowed = false
	g.FlashlightOn = true
	g.Particles = nil
	g.shakeDuration = 0
	g.shakeIntensity = 0
	g.deathReason = ""
	g.MineWarning = ""
	g.MineWarningTimer = 0
	g.MineWarningLevel = 0

	g.overworldState = NewOverworldScene(w)

	g.TransitionToOverworld()
}

func (g *Game) TransitionToOverworld() { g.TransitionTo(g.overworldState) }
func (g *Game) TransitionToGameWon()   { g.TransitionTo(g.gameWonState) }

// EnterCave and ExitCave are defined in transition.go.

func (g *Game) HorizontalTransition(newTx, newTy int, newTrenchKey string, newCave cave.Cave, newGrid [][]bool, newNodes []resource.Resource, newEntities []entity.CaveEntity) {
	oldKey := g.activeTrenchKey

	// Save old cave state
	g.caveNodes[oldKey] = g.caveState.Nodes
	g.caveEntities[oldKey] = g.caveState.Entities

	// Set new trench coordinates and key
	g.activeTrenchX = newTx
	g.activeTrenchY = newTy
	g.activeTrenchKey = newTrenchKey

	// Apply new cave scene state
	g.caveState.ActiveCave = newCave
	g.caveState.CaveGrid = newGrid
	g.caveState.Nodes = newNodes
	g.caveState.Entities = newEntities

	// Update the player's last overworld emergence coordinates to match new location
	playerWidth := g.player.Width
	playerHeight := g.player.Height
	g.lastOverworldX = float64(newTx*config.TileSize) + (config.TileSize-playerWidth)/2
	g.lastOverworldY = float64(newTy*config.TileSize) + (config.TileSize-playerHeight)/2

	// Update vehicle mapping
	if g.ActiveVehicle != nil {
		oldList := g.CaveVehicles[oldKey]
		for i, v := range oldList {
			if v == g.ActiveVehicle {
				g.CaveVehicles[oldKey] = append(oldList[:i], oldList[i+1:]...)
				break
			}
		}
		g.CaveVehicles[newTrenchKey] = append(g.CaveVehicles[newTrenchKey], g.ActiveVehicle)
	}
}

// --- Input ---

func (g *Game) GetInput() scene.InputSource { return g.Input }

// --- Core state ---

func (g *Game) GetCurrentState() scene.State  { return g.currentState }
func (g *Game) SetCurrentState(s scene.State) { g.currentState = s }

// --- Core objects ---

func (g *Game) GetPlayer() *player.Player         { return g.player }
func (g *Game) GetCamera() *camera.Camera         { return g.camera }
func (g *Game) GetWorld() *world.World            { return g.world }
func (g *Game) GetBaseStation() *base.BaseStation { return g.baseStation }

// --- Vehicle state ---

func (g *Game) GetActiveVehicle() vehicle.Vehicle       { return g.ActiveVehicle }
func (g *Game) GetOverworldVehicles() []vehicle.Vehicle { return g.OverworldVehicles }
func (g *Game) GetCaveVehicles(key string) []vehicle.Vehicle {
	return g.CaveVehicles[key]
}
func (g *Game) GetActiveTrenchKey() string { return g.activeTrenchKey }
func (g *Game) GetActiveTrenchCoords() (x, y int) {
	return g.activeTrenchX, g.activeTrenchY
}

// --- Cave state ---

func (g *Game) GetActiveCave() cave.Cave { return g.caveState.ActiveCave }
func (g *Game) GetCaveNodes(key string) []resource.Resource {
	return g.caveNodes[key]
}
func (g *Game) SetCaveNodes(key string, nodes []resource.Resource) {
	g.caveNodes[key] = nodes
}
func (g *Game) GetCaveEntities(key string) []entity.CaveEntity {
	return g.caveEntities[key]
}
func (g *Game) SetCaveEntities(key string, entities []entity.CaveEntity) {
	g.caveEntities[key] = entities
}

// --- Entity runtime ---

func (g *Game) NewEntityRuntime() entity.Runtime {
	return &entityRuntimeAdapter{g: g}
}

func (g *Game) DrainEntityCommands(rt entity.Runtime) {
	if adapter, ok := rt.(*entityRuntimeAdapter); ok {
		g.drainEntityCommands(adapter)
	}
}

// --- Particles ---

func (g *Game) SpawnBubble(x, y float64) {
	g.Particles = append(g.Particles, particle.NewBubbleParticle(x, y))
}

func (g *Game) GetParticles() []*particle.Particle { return g.Particles }

// SpawnPlankton and SpawnDebris are defined in game.go.

// --- Time / ticks ---

func (g *Game) GetTimeOfDay() float64 { return g.TimeOfDay }
func (g *Game) GetTicks() float64     { return g.Ticks }

// --- Game state flags ---

func (g *Game) GetSonar() *sonar.Sonar { return g.Sonar }

func (g *Game) GetSoundWaveState() (timer int, x, y, radius float64) {
	return g.SoundWaveTimer, g.SoundWaveX, g.SoundWaveY, g.SoundWaveRadius
}

func (g *Game) SetSoundWaveState(timer int, x, y, radius float64) {
	g.SoundWaveTimer = timer
	g.SoundWaveX = x
	g.SoundWaveY = y
	g.SoundWaveRadius = radius
}

func (g *Game) IsPlayerSlowed() bool             { return g.playerSlowed }
func (g *Game) IsFlashlightOn() bool             { return g.FlashlightOn }
func (g *Game) GetWeaverTrackingTimer() float64  { return g.WeaverTrackingTimer }
func (g *Game) SetWeaverTrackingTimer(v float64) { g.WeaverTrackingTimer = v }

// --- HUD / UI ---

func (g *Game) IsInventoryOpen() bool   { return g.showInventory }
func (g *Game) SetInventoryOpen(v bool) { g.showInventory = v }

func (g *Game) GetMineWarning() (msg string, timer int) {
	return g.MineWarning, g.MineWarningTimer
}

func (g *Game) SetMineWarning(msg string, duration, level int) {
	g.MineWarning = msg
	g.MineWarningTimer = duration
	g.MineWarningLevel = level
}

// TriggerScreenShake is defined in game.go.

// --- Death state ---

func (g *Game) GetDeathReason() string       { return g.deathReason }
func (g *Game) SetDeathReason(reason string) { g.deathReason = reason }

// --- Debug ---

func (g *Game) IsDebugLightShaderDisabled() bool { return g.DebugDisableLightShader }
func (g *Game) IsDebugWaterShaderDisabled() bool { return g.DebugDisableWaterShader }

// --- Story and Lore ---

func (g *Game) GetStoryManager() *story.StoryManager { return g.storyManager }

func (g *Game) TransitionToPDA() {
	if g.currentState == scene.StateOverworld || g.currentState == scene.StateCave {
		g.pdaPriorState = g.currentState
	}
	g.menuOpenedAnywhere = true
	g.baseMenu.ActiveTab = 4
	g.TransitionTo(g.baseMenu)
}

func (g *Game) ClosePDA() {
	if g.pdaPriorState == scene.StateCave {
		g.TransitionTo(g.caveState)
	} else {
		g.TransitionTo(g.overworldState)
	}
	g.pdaPriorState = 0
	g.menuOpenedAnywhere = false
}

func (g *Game) IsMenuOpenedAnywhere() bool {
	return g.menuOpenedAnywhere
}

func (g *Game) TransitionToIntro(seed int64) {
	g.introState.SetSeed(seed)
	g.TransitionTo(g.introState)
}
