package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/sonar"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Scene represents a distinct game state or view (e.g. Overworld, Cave, Menu, Game Over).
type Scene interface {
	Update(g GameContext) error
	Draw(g GameContext, screen *ebiten.Image)
	OnEnter(g GameContext)
	OnExit(g GameContext)
}

// GameContext is the interface through which scenes interact with the game.
// game.Game implements this; scenes depend on the interface rather than the concrete type,
// breaking the circular import that would result from scenes importing the game package.
type GameContext interface {
	// Scene navigation
	TransitionToOverworld()
	TransitionToGameWon()
	Respawn()
	EnterCave(tx, ty int)
	ExitCave()

	// Input
	GetInput() InputSource

	// Core state
	GetCurrentState() State
	SetCurrentState(s State)

	// Core objects
	GetPlayer() *player.Player
	GetCamera() *camera.Camera
	GetWorld() *world.World
	GetBaseStation() *base.BaseStation

	// Vehicle state
	GetActiveVehicle() vehicle.Vehicle
	GetOverworldVehicles() []vehicle.Vehicle
	GetCaveVehicles(key string) []vehicle.Vehicle
	GetActiveTrenchKey() string
	GetActiveTrenchCoords() (x, y int)

	// Cave state
	GetActiveCave() cave.Cave
	GetCaveNodes(key string) []resource.Resource
	GetCaveEntities(key string) []entity.CaveEntity
	SetCaveEntities(key string, entities []entity.CaveEntity)

	// Entity runtime (adapter is private to game package; returned as the public entity.Runtime interface)
	NewEntityRuntime() entity.Runtime
	DrainEntityCommands(rt entity.Runtime)

	// Particles
	SpawnPlankton(x, y float64)
	SpawnDebris(x, y float64, clr color.RGBA)
	SpawnBubble(x, y float64)
	GetParticles() []*particle.Particle

	// Time / ticks
	GetTimeOfDay() float64
	GetTicks() float64

	// Game state flags
	GetSonar() *sonar.Sonar
	GetSoundWaveState() (timer int, x, y, radius float64)
	SetSoundWaveState(timer int, x, y, radius float64)
	IsPlayerSlowed() bool
	IsFlashlightOn() bool
	GetWeaverTrackingTimer() float64
	SetWeaverTrackingTimer(v float64)

	// HUD / UI
	IsInventoryOpen() bool
	GetMineWarning() (msg string, timer int)
	SetMineWarning(msg string, duration int)

	// Screen effects
	TriggerScreenShake(duration int, intensity float64)

	// Debug toggles
	IsDebugLightShaderDisabled() bool
	IsDebugWaterShaderDisabled() bool
}
