package game

import (
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/quest"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/game/sonar"
	"github.com/jaredwarren/SubGame/internal/game/story"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Session holds player, world, navigation, vehicles, and world clock state.
type Session struct {
	player *player.Player
	world  *world.World
	camera *camera.Camera
	Input  InputSource

	lastOverworldX  float64
	lastOverworldY  float64
	activeTrenchX   int
	activeTrenchY   int
	activeTrenchKey string
	justExited      bool

	baseStation *base.BaseStation

	ActiveVehicle     vehicle.Vehicle
	OverworldVehicles []vehicle.Vehicle
	CaveVehicles      map[string][]vehicle.Vehicle

	TimeOfDay float64
	Ticks     float64
}

// CaveCache holds per-trench cave resource and entity snapshots.
type CaveCache struct {
	caveNodes    map[string][]resource.Resource
	caveEntities map[string][]entity.CaveEntity
}

// Effects holds transient presentation and alert state.
type Effects struct {
	Sonar               *sonar.Sonar
	MineWarning         WarningBanner
	FlashlightOn        bool
	WeaverTrackingTimer float64
	SoundWave           SoundWaveState
	playerSlowed        bool
	o2LowAlertPlayed    bool
	o2CritAlertPlayed   bool
	Particles           []*particle.Particle
	Shake               ScreenShake
	deathReason         string
}

// Progress holds story, quests, exploration, and meta-progression.
type Progress struct {
	storyManager       *story.StoryManager
	questManager       *quest.QuestManager
	pdaPriorState      State
	menuOpenedAnywhere bool
	craftingRecipes    []scene.Recipe
	explorationTracker *exploration.Tracker
	lostCargo          []*entity.LostCargoBeacon
	TutorialActive     bool
}

// resetCaveCache clears per-world cave snapshots (new game / load).
func (g *Game) resetCaveCache() {
	g.caveNodes = make(map[string][]resource.Resource)
	g.caveEntities = make(map[string][]entity.CaveEntity)
}

// resetEffects clears transient FX and alert state for a fresh session.
func (g *Game) resetEffects() {
	g.Sonar = sonar.NewSonar()
	g.MineWarning = WarningBanner{}
	g.FlashlightOn = true
	g.WeaverTrackingTimer = 0
	g.SoundWave = SoundWaveState{}
	g.playerSlowed = false
	g.o2LowAlertPlayed = false
	g.o2CritAlertPlayed = false
	g.Particles = nil
	g.Shake = ScreenShake{}
	g.deathReason = ""
}

// resetProgressManagers replaces story/quest/exploration for a new run.
func (g *Game) resetProgressManagers(w *world.World, spawnTX, spawnTY int) {
	g.storyManager = story.NewStoryManager()
	g.questManager = quest.NewQuestManager()
	g.craftingRecipes = scene.DefaultCraftingRecipes()
	g.explorationTracker = exploration.NewTracker(w.Width, w.Height)
	g.explorationTracker.Reveal(spawnTX, spawnTY, exploration.RevealRadius)
	g.lostCargo = nil
}

// resetNavigation clears trench / surface navigation bookkeeping.
func (g *Game) resetNavigation() {
	g.lastOverworldX = 0
	g.lastOverworldY = 0
	g.activeTrenchX = 0
	g.activeTrenchY = 0
	g.activeTrenchKey = ""
	g.justExited = false
}

// resetVehicles clears all deployed craft.
func (g *Game) resetVehicles() {
	g.ActiveVehicle = nil
	g.OverworldVehicles = nil
	g.CaveVehicles = make(map[string][]vehicle.Vehicle)
}

// grantStarterInventory adds tutorial starter items when applicable.
func (g *Game) grantStarterInventory() {
	if g.player == nil || g.player.Inventory == nil {
		return
	}
	g.player.Inventory.AddItem(&item.Titanium{}, 9)
}
