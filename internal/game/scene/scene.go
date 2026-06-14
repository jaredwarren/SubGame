package scene

import (
	"image/color"
	"math"

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
	"github.com/jaredwarren/SubGame/internal/game/story"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Scene represents a distinct game state or view (e.g. Overworld, Cave, Menu, Game Over).
type Scene interface {
	Update(g GameContext) error
	Draw(g GameContext, screen *ebiten.Image)
	OnEnter(g GameContext)
	OnExit(g GameContext)
}

type Navigator interface {
	StartGame(seed int64)
	TransitionToOverworld()
	TransitionToGameWon()
	Respawn()
	EnterCave(tx, ty int)
	ExitCave()
	HorizontalTransition(newTx, newTy int, newTrenchKey string, newCave cave.Cave, newGrid [][]bool, newNodes []resource.Resource, newEntities []entity.CaveEntity)
	TransitionToPDA()
	ClosePDA()
	TransitionToIntro(seed int64)
	SetCurrentState(s State)
	GetCurrentState() State
}

type InputAccess interface {
	GetInput() InputSource
}

type PlayerAccess interface {
	GetPlayer() *player.Player
	IsPlayerSlowed() bool
	IsFlashlightOn() bool
	PickUpActiveVehicle()
	TransferToVehicle(it item.Item)
	ActivatePlayerItem(it item.Item)
}

type WorldAccess interface {
	GetWorld() *world.World
	GetBaseStation() *base.BaseStation
	GetCamera() *camera.Camera
}

type CaveStateAccess interface {
	GetActiveCave() cave.Cave
	GetCaveNodes(key string) []resource.Resource
	SetCaveNodes(key string, nodes []resource.Resource)
	GetCaveEntities(key string) []entity.CaveEntity
	SetCaveEntities(key string, entities []entity.CaveEntity)
	GetActiveVehicle() vehicle.Vehicle
	GetOverworldVehicles() []vehicle.Vehicle
	GetCaveVehicles(key string) []vehicle.Vehicle
	GetAllCaveVehicles() map[string][]vehicle.Vehicle
	GetActiveTrenchKey() string
	GetActiveTrenchCoords() (x, y int)
}

type EffectsEmitter interface {
	NewEntityRuntime() entity.Runtime
	DrainEntityCommands(rt entity.Runtime)
	SpawnPlankton(x, y float64)
	SpawnDebris(x, y float64, clr color.RGBA)
	SpawnBubble(x, y float64)
	GetParticles() []*particle.Particle
	TriggerScreenShake(duration int, intensity float64)
	GetSonar() *sonar.Sonar
	GetSoundWaveState() (timer int, x, y, radius float64)
	SetSoundWaveState(timer int, x, y, radius float64)
}

type StoryAccess interface {
	GetStoryManager() *story.StoryManager
	GetCraftingRecipes() []Recipe
}

type TimeAccess interface {
	GetTimeOfDay() float64
	GetTicks() float64
}

type UIAccess interface {
	IsInventoryOpen() bool
	SetInventoryOpen(v bool)
	GetMineWarning() (msg string, timer int)
	SetMineWarning(msg string, duration, level int)
	IsMenuOpenedAnywhere() bool
}

type DebugAccess interface {
	IsDebugLightShaderDisabled() bool
	IsDebugWaterShaderDisabled() bool
}

// GameContext is the interface through which scenes interact with the game.
// game.Game implements this; scenes depend on the interface rather than the concrete type,
// breaking the circular import that would result from scenes importing the game package.
type GameContext interface {
	Navigator
	InputAccess
	PlayerAccess
	WorldAccess
	CaveStateAccess
	EffectsEmitter
	StoryAccess
	TimeAccess
	UIAccess
	DebugAccess

	GetDeathReason() string
	SetDeathReason(reason string)
	DestroyOverworldVehicle(v vehicle.Vehicle)
	GetWeaverTrackingTimer() float64
	SetWeaverTrackingTimer(v float64)
}

// tileAt calculates the tile index for a given single coordinate using floor division to handle negative bounds.
func tileAt(coord float64, tileSize int) int {
	return int(math.Floor(coord / float64(tileSize)))
}

// tileRange calculates the tile index range spanned by a bounding box.
// Subtracts a small epsilon of 0.001 from the maximum bounds to prevent flush boundaries probing an extra tile.
func tileRange(x, y, w, h float64, tileSize int) (x1, x2, y1, y2 int) {
	return gvec.TileRange(gvec.Vec2{X: x, Y: y}, gvec.Vec2{X: w, Y: h}, tileSize)
}
