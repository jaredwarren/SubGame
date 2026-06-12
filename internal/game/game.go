package game

import (
	"image/color"

	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/game/sonar"
	"github.com/jaredwarren/SubGame/internal/game/story"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Game implements ebiten.Game and owns all shared game state.
// Scenes interact with Game through the GameContext interface.
type Game struct {
	// Scene management
	currentState          State
	currentScene          Scene
	nextScene             Scene // scheduled deferred transition
	transitionedThisFrame bool
	titleState            *TitleScene
	introState            *IntroScene
	overworldState        *OverworldScene
	caveState             *CaveScene
	baseMenu              *BaseMenuScene
	gameOverState         *GameOverScene
	gameWonState          *GameWonScene

	// Core objects
	player *player.Player
	hud    *HUD
	world  *world.World
	camera *camera.Camera
	Input  InputSource

	// Navigation
	lastOverworldX  float64
	lastOverworldY  float64
	activeTrenchX   int
	activeTrenchY   int
	activeTrenchKey string
	justExited      bool

	// Inventory / cave resources
	caveNodes     map[string][]resource.Resource
	showInventory bool
	baseStation   *base.BaseStation

	// Vehicles
	ActiveVehicle     vehicle.Vehicle
	OverworldVehicles []vehicle.Vehicle
	CaveVehicles      map[string][]vehicle.Vehicle // keyed by trenchKey

	// World time
	TimeOfDay float64 // 0–14400 ticks per 4-min day/night cycle
	Ticks     float64

	// Sonar and alerts
	Sonar            *sonar.Sonar
	MineWarning      string
	MineWarningTimer int
	MineWarningLevel int

	// Biome / AI state
	caveEntities        map[string][]entity.CaveEntity
	FlashlightOn        bool
	WeaverTrackingTimer float64
	SoundWaveTimer      int
	SoundWaveRadius     float64
	SoundWaveX          float64
	SoundWaveY          float64
	playerSlowed        bool // reset each tick by entity system

	// Effects
	Particles      []*particle.Particle
	shakeDuration  int
	shakeIntensity float64
	deathReason    string

	// Debug
	DebugDisableLightShader bool
	DebugDisableWaterShader bool

	// Story and Lore
	storyManager       *story.StoryManager
	pdaPriorState      State
	menuOpenedAnywhere bool
	craftingRecipes    []scene.Recipe
}

// NewGame creates a fully initialized Game ready to run.
func NewGame() *Game {
	vehicle.LoadAssets()
	resource.LoadAssets()
	item.LoadAssets()
	base.LoadAssets()
	scene.LoadAssets()
	w := world.NewWorld(12345)

	spawnX, spawnY := findWaterSpawn(w)

	p := player.NewPlayer(spawnX, spawnY)
	cam := camera.NewCamera(spawnX, spawnY)
	cam.CenterOn(spawnX, spawnY, p.Width, p.Height)

	baseStation := base.NewBaseStation(spawnX+96.0, spawnY-64.0)
	skiff := vehicle.NewSkiff(spawnX, spawnY)

	sm := story.NewStoryManager()
	_ = sm.Load(story.LoreJSONBytes)

	g := &Game{
		currentState:      StateTitle,
		player:            p,
		hud:               NewHUD(),
		world:             w,
		camera:            cam,
		Input:             NewEbitenInput(),
		caveNodes:         make(map[string][]resource.Resource),
		baseStation:       baseStation,
		ActiveVehicle:     skiff,
		OverworldVehicles: []vehicle.Vehicle{skiff},
		CaveVehicles:      make(map[string][]vehicle.Vehicle),
		Sonar:             sonar.NewSonar(),
		caveEntities:      make(map[string][]entity.CaveEntity),
		FlashlightOn:      true,
		storyManager:      sm,
		craftingRecipes:   scene.DefaultCraftingRecipes(),
	}

	g.titleState = NewTitleScene()
	g.introState = NewIntroScene()
	g.overworldState = NewOverworldScene(w)
	g.caveState = NewCaveScene()
	g.baseMenu = NewBaseMenuScene()
	g.gameOverState = NewGameOverScene()
	g.gameWonState = NewGameWonScene()

	g.TransitionTo(g.titleState)
	return g
}

// findWaterSpawn scans the center region of the world map for a suitable water tile.
func findWaterSpawn(w *world.World) (x, y float64) {
	x = 50.0 * config.TileSize
	y = 50.0 * config.TileSize
	for tx := 45; tx < 55; tx++ {
		for ty := 45; ty < 55; ty++ {
			if w.OverworldMap[tx][ty] == world.TileWater {
				return float64(tx*config.TileSize) + (config.TileSize-20.0)/2.0,
					float64(ty*config.TileSize) + (config.TileSize-20.0)/2.0
			}
		}
	}
	return x, y
}

// TransitionTo switches the active scene, calling lifecycle hooks on the old and new scenes.
func (g *Game) TransitionTo(next Scene) {
	if g.currentScene != nil {
		g.currentScene.OnExit(g)
	}
	g.currentScene = next
	if next != nil {
		next.OnEnter(g)
	}
	g.transitionedThisFrame = true
}

// Respawn resets the player after death and returns to the overworld.
func (g *Game) Respawn() {
	g.player.Pos = gvec.Vec2{X: g.baseStation.Pos.X - 96.0, Y: g.baseStation.Pos.Y + 64.0}
	g.player.Vel = gvec.Vec2{}
	g.player.CurrentHealth = g.player.MaxHealth
	g.player.LastHealth = g.player.MaxHealth
	g.player.CurrentOxygen = g.player.MaxOxygen
	g.player.CurrentStamina = g.player.MaxStamina
	g.player.Inventory.Clear()
	g.ActiveVehicle = nil
	g.deathReason = ""
	g.shakeDuration = 0
	g.shakeIntensity = 0
	g.showInventory = false
	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
	g.TransitionTo(g.overworldState)
}

// DestroyOverworldVehicle removes a vehicle from the overworld list and resets active vehicle.
func (g *Game) DestroyOverworldVehicle(v vehicle.Vehicle) {
	for i, ov := range g.OverworldVehicles {
		if ov == v {
			g.OverworldVehicles = append(g.OverworldVehicles[:i], g.OverworldVehicles[i+1:]...)
			break
		}
	}
	if g.ActiveVehicle == v {
		g.ActiveVehicle = nil
	}
}

// TriggerScreenShake registers a screen shake — higher intensity/longer duration wins.
func (g *Game) TriggerScreenShake(duration int, intensity float64) {
	if intensity > g.shakeIntensity || g.shakeDuration <= 0 {
		g.shakeIntensity = intensity
	}
	if duration > g.shakeDuration {
		g.shakeDuration = duration
	}
}

// Layout returns the logical screen size for ebiten.
func (g *Game) Layout(_, _ int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

// SpawnPlankton adds a plankton particle at the given world position.
func (g *Game) SpawnPlankton(x, y float64) {
	g.Particles = append(g.Particles, particle.NewPlanktonParticle(x, y))
}

// SpawnDebris adds debris particles at the given world position.
func (g *Game) SpawnDebris(x, y float64, clr color.RGBA) {
	g.Particles = append(g.Particles, particle.NewDebrisParticles(x, y, clr)...)
}
