package game

import (
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/game/sonar"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

// compile-time assertion: *Game must satisfy scene.GameContext
var _ scene.GameContext = (*Game)(nil)

// --- Scene navigation ---

func (g *Game) TransitionToOverworld() { g.TransitionTo(g.overworldState) }
func (g *Game) TransitionToGameWon()   { g.TransitionTo(g.gameWonState) }

// EnterCave and ExitCave are defined in transition.go.

// --- Input ---

func (g *Game) GetInput() scene.InputSource { return g.Input }

// --- Core state ---

func (g *Game) GetCurrentState() scene.State  { return g.currentState }
func (g *Game) SetCurrentState(s scene.State) { g.currentState = s }

// --- Core objects ---

func (g *Game) GetPlayer() *player.Player        { return g.player }
func (g *Game) GetCamera() *camera.Camera        { return g.camera }
func (g *Game) GetWorld() *world.World           { return g.world }
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

func (g *Game) IsPlayerSlowed() bool          { return g.playerSlowed }
func (g *Game) IsFlashlightOn() bool          { return g.FlashlightOn }
func (g *Game) GetWeaverTrackingTimer() float64 { return g.WeaverTrackingTimer }
func (g *Game) SetWeaverTrackingTimer(v float64) { g.WeaverTrackingTimer = v }

// --- HUD / UI ---

func (g *Game) IsInventoryOpen() bool { return g.showInventory }
func (g *Game) SetInventoryOpen(v bool) { g.showInventory = v }

func (g *Game) GetMineWarning() (msg string, timer int) {
	return g.MineWarning, g.MineWarningTimer
}

func (g *Game) SetMineWarning(msg string, duration int) {
	g.MineWarning = msg
	g.MineWarningTimer = duration
}

// TriggerScreenShake is defined in game.go.

// --- Death state ---

func (g *Game) GetDeathReason() string                 { return g.deathReason }
func (g *Game) SetDeathReason(reason string)          { g.deathReason = reason }

// --- Debug ---

func (g *Game) IsDebugLightShaderDisabled() bool { return g.DebugDisableLightShader }
func (g *Game) IsDebugWaterShaderDisabled() bool  { return g.DebugDisableWaterShader }
