package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/quest"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/save"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/game/sonar"
	"github.com/jaredwarren/SubGame/internal/game/story"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Game implements ebiten.Game and owns all shared game state.
// Scenes interact with Game through the GameContext interface.
type SoundWaveState struct {
	Timer  int
	Radius float64
	X      float64
	Y      float64
}

type WarningBanner struct {
	Message string
	Timer   int
	Level   int
}

type ScreenShake struct {
	Duration  int
	Intensity float64
}

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
	pauseState            *scene.PauseScene

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
	MineWarning      WarningBanner

	// Biome / AI state
	caveEntities        map[string][]entity.CaveEntity
	FlashlightOn        bool
	WeaverTrackingTimer float64
	SoundWave           SoundWaveState
	playerSlowed        bool // reset each tick by entity system
	o2LowAlertPlayed    bool
	o2CritAlertPlayed   bool

	// Effects
	Particles      []*particle.Particle
	Shake          ScreenShake
	deathReason    string

	// Debug
	DebugDisableLightShader bool
	DebugDisableWaterShader bool

	// Story, Quests and Lore
	storyManager       *story.StoryManager
	questManager       *quest.QuestManager
	pdaPriorState      State
	menuOpenedAnywhere bool
	craftingRecipes    []scene.Recipe

	// Exploration / fog-of-war
	explorationTracker *exploration.Tracker

	// Death cargo beacons (inventory dropped at death site; upgrades stay equipped).
	lostCargo []*entity.LostCargoBeacon

	// Tutorial
	TutorialActive bool
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

	sm := story.NewStoryManager()
	qm := quest.NewQuestManager()
	tracker := exploration.NewTracker(w.Width, w.Height)
	spawnTX := int(math.Floor((spawnX + p.Width/2.0) / float64(config.TileSize)))
	spawnTY := int(math.Floor((spawnY + p.Height/2.0) / float64(config.TileSize)))
	tracker.Reveal(spawnTX, spawnTY, exploration.RevealRadius)

	g := &Game{
		currentState:         StateTitle,
		player:               p,
		hud:                  NewHUD(),
		world:                w,
		camera:               cam,
		Input:                NewEbitenInput(),
		caveNodes:            make(map[string][]resource.Resource),
		baseStation:          baseStation,
		ActiveVehicle:        nil,
		OverworldVehicles:    nil,
		CaveVehicles:         make(map[string][]vehicle.Vehicle),
		Sonar:                sonar.NewSonar(),
		caveEntities:         make(map[string][]entity.CaveEntity),
		FlashlightOn:         true,
		storyManager:         sm,
		questManager:         qm,
		craftingRecipes:      scene.DefaultCraftingRecipes(),
		explorationTracker:   tracker,
	}

	g.titleState = NewTitleScene()
	g.introState = NewIntroScene()
	g.overworldState = NewOverworldScene(w)
	g.caveState = NewCaveScene()
	g.baseMenu = NewBaseMenuScene()
	g.gameOverState = NewGameOverScene()
	g.gameWonState = NewGameWonScene()
	g.pauseState = scene.NewPauseScene()

	g.TransitionTo(g.titleState)
	return g
}

// findWaterSpawn finds the ShallowReefBiome water tile nearest the center of the world map.
func findWaterSpawn(w *world.World) (x, y float64) {
	if w == nil {
		return 50.0 * config.TileSize, 50.0 * config.TileSize
	}
	tx, ty := w.FindLifepodSpawn()
	return float64(tx*config.TileSize) + (config.TileSize-20.0)/2.0,
		float64(ty*config.TileSize) + (config.TileSize-20.0)/2.0
}

// findNearestClearWaterDeployPos returns a top-left vehicle position centered on the
// nearest overworld tile where dims fit entirely on water and clear of the lifepod.
func (g *Game) findNearestClearWaterDeployPos(near gvec.Vec2, dims gvec.Vec2) gvec.Vec2 {
	w := g.world
	if w == nil || w.Width == 0 || w.Height == 0 {
		return near
	}

	ts := float64(config.TileSize)
	startTX := int(math.Floor((near.X + dims.X/2) / ts))
	startTY := int(math.Floor((near.Y + dims.Y/2) / ts))
	startTX = max(0, min(startTX, w.Width-1))
	startTY = max(0, min(startTY, w.Height-1))

	type tilePos struct{ x, y int }
	visited := make([][]bool, w.Width)
	for x := range visited {
		visited[x] = make([]bool, w.Height)
	}
	queue := []tilePos{{startTX, startTY}}
	visited[startTX][startTY] = true
	dirs := []tilePos{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		pos := gvec.Vec2{
			X: float64(cur.x)*ts + ts/2 - dims.X/2,
			Y: float64(cur.y)*ts + ts/2 - dims.Y/2,
		}
		if g.isClearOverworldDeploy(pos, dims) {
			return pos
		}

		for _, d := range dirs {
			nx, ny := cur.x+d.x, cur.y+d.y
			if nx < 0 || ny < 0 || nx >= w.Width || ny >= w.Height || visited[nx][ny] {
				continue
			}
			visited[nx][ny] = true
			queue = append(queue, tilePos{nx, ny})
		}
	}
	return near
}

// isClearOverworldDeploy reports whether a vehicle bbox at pos is fully on water
// and does not overlap the lifepod collision box.
func (g *Game) isClearOverworldDeploy(pos, dims gvec.Vec2) bool {
	w := g.world
	x1, x2, y1, y2 := gvec.TileRange(pos, dims, config.TileSize)
	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || ty < 0 || tx >= w.Width || ty >= w.Height {
				return false
			}
			info := world.GetTileInfo(w.OverworldMap[tx][ty])
			if info == nil || !info.IsWater {
				return false
			}
		}
	}
	if g.baseStation != nil {
		bPos, bSize := g.baseStation.Pos, g.baseStation.Size
		if pos.X < bPos.X+bSize.X && pos.X+dims.X > bPos.X &&
			pos.Y < bPos.Y+bSize.Y && pos.Y+dims.Y > bPos.Y {
			return false
		}
	}
	return true
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
	g.updateSceneAudio(next)
}

func (g *Game) updateSceneAudio(next Scene) {
	audioMgr := audio.Get()
	if audioMgr == nil {
		return
	}

	switch next {
	case g.titleState:
		audioMgr.SetSubmerged(false)
		audioMgr.PlayMusic("music/main_title.mp3", 0.6)
	case g.introState:
		audioMgr.SetSubmerged(false)
		audioMgr.PlayMusic("music/intro_cinematic.mp3", 0.7)
	case g.overworldState:
		audioMgr.SetSubmerged(false)
		audioMgr.PlayMusic("music/overworld_surface.mp3", 0.5)
	case g.caveState:
		audioMgr.SetSubmerged(true)
		musicTrack := "music/cave_shallow.mp3"
		if g.caveState != nil && g.caveState.ActiveCave != nil {
			switch g.caveState.ActiveCave.GetCaveType() {
			case cave.CaveThermo:
				musicTrack = "music/cave_volcanic.mp3"
			case cave.CaveShockKelp:
				musicTrack = "music/cave_kelp.mp3"
			case cave.CaveWreckage:
				musicTrack = "music/cave_wreckage.mp3"
			case cave.CaveVoid, cave.CaveOrganicTrench:
				musicTrack = "music/cave_abyssal.mp3"
			}
		}
		audioMgr.PlayMusic(musicTrack, 0.6)
	case g.gameOverState:
		audioMgr.SetSubmerged(false)
		audioMgr.PlayMusic("music/game_over_theme.mp3", 0.7)
	case g.gameWonState:
		audioMgr.SetSubmerged(false)
		audioMgr.PlayMusic("music/escape_outro.mp3", 0.75)
	}
}

// Respawn resets the player after death and returns to the overworld.
// Equipped upgrades are kept. Inventory/hotbar cargo was already moved to a
// lost-cargo beacon at the death site (see dropLostCargo).
func (g *Game) Respawn() {
	g.player.Pos = gvec.Vec2{X: g.baseStation.Pos.X - 96.0, Y: g.baseStation.Pos.Y + 64.0}
	g.player.Vel = gvec.Vec2{}
	g.player.CurrentHealth = g.player.MaxHealth
	g.player.LastHealth = g.player.MaxHealth
	g.player.CurrentOxygen = g.player.MaxOxygen
	g.player.CurrentStamina = g.player.MaxStamina
	// Safety: cargo should already be empty after dropLostCargo; never wipe upgrades.
	g.player.Inventory.Clear()
	if g.player.Hotbar != nil {
		g.player.Hotbar.Clear()
	}
	if g.TutorialActive {
		g.player.Inventory.AddItem(&item.Titanium{}, 9)
	}
	g.ActiveVehicle = nil
	g.deathReason = ""
	g.Shake.Duration = 0
	g.Shake.Intensity = 0
	g.showInventory = false
	g.camera.CenterOn(g.player.Pos.X, g.player.Pos.Y, g.player.Width, g.player.Height)
	g.TransitionTo(g.overworldState)
}

// dropLostCargo deposits inventory + hotbar as a surface crate at the death area.
// Cave deaths pin cargo to the dive-site overworld coords so the return trip is chartable.
// Equipped upgrade slots are left untouched.
func (g *Game) dropLostCargo() {
	var stacks []item.ItemStack
	if g.player.Inventory != nil {
		stacks = append(stacks, g.player.Inventory.ExtractAll()...)
	}
	if g.player.Hotbar != nil {
		stacks = append(stacks, g.player.Hotbar.ExtractAll()...)
	}
	if len(stacks) == 0 {
		return
	}

	pos := g.lostCargoOverworldPos()
	beacon := entity.NewLostCargoBeacon(pos, stacks)
	g.lostCargo = append(g.lostCargo, beacon)
	g.SetMineWarning("Cargo crate left on the surface — find it on the map!", 240, 2)
}

// lostCargoOverworldPos returns where a new death crate should sit on the overworld.
// Always overworld so the recovery expedition is visible after life-pod respawn.
func (g *Game) lostCargoOverworldPos() gvec.Vec2 {
	// Cave death → pin to last surface position (dive site). Fall back to trench tile center.
	if g.currentState == StateCave {
		if g.lastOverworldX != 0 || g.lastOverworldY != 0 {
			return gvec.Vec2{
				X: g.lastOverworldX + g.player.Width/2.0,
				Y: g.lastOverworldY + g.player.Height/2.0,
			}
		}
		return gvec.Vec2{
			X: float64(g.activeTrenchX*config.TileSize) + float64(config.TileSize)/2.0,
			Y: float64(g.activeTrenchY*config.TileSize) + float64(config.TileSize)/2.0,
		}
	}

	if g.ActiveVehicle != nil {
		vPos := g.ActiveVehicle.GetPos()
		vDims := g.ActiveVehicle.GetDimensions()
		return gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
	}
	return gvec.Vec2{
		X: g.player.Pos.X + g.player.Width/2.0,
		Y: g.player.Pos.Y + g.player.Height/2.0,
	}
}

// GetLostCargo returns active surface cargo beacons for map markers / debugging.
func (g *Game) GetLostCargo() []*entity.LostCargoBeacon {
	return g.lostCargo
}

// updateLostCargo ticks beacon lifetimes and handles recovery when the player is nearby.
func (g *Game) updateLostCargo() {
	if len(g.lostCargo) == 0 {
		return
	}
	// Lifetime and recovery only while playing the overworld / cave (surface crates
	// are recoverable only on the overworld).
	if g.currentState != StateOverworld && g.currentState != StateCave {
		return
	}

	// Lifetime still ticks in both states so 3-day clocks keep running while diving.
	live := g.lostCargo[:0]
	for _, b := range g.lostCargo {
		if b == nil {
			continue
		}
		if b.TickLifetime() {
			g.SetMineWarning("A cargo crate sank into the deep…", 150, 2)
			continue
		}
		// Recovery only on the surface (crate is always overworld-placed).
		if g.currentState == StateOverworld {
			pCenter := gvec.Vec2{
				X: g.player.Pos.X + g.player.Width/2.0,
				Y: g.player.Pos.Y + g.player.Height/2.0,
			}
			reach := 36.0
			if g.ActiveVehicle != nil {
				vPos := g.ActiveVehicle.GetPos()
				vDims := g.ActiveVehicle.GetDimensions()
				pCenter = gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
				reach = 20.0 + math.Max(vDims.X, vDims.Y)/2.0
			}
			dist := math.Hypot(pCenter.X-b.Pos.X, pCenter.Y-b.Pos.Y)
			if dist < reach {
				n, empty := b.TryRecover(g.player.Inventory)
				if !empty {
					var n2 int
					n2, empty = b.TryRecover(g.player.Hotbar)
					n += n2
				}
				if n > 0 {
					g.SetMineWarning(fmt.Sprintf("Recovered %d items from lost cargo!", n), 150, 1)
				}
				if empty {
					continue
				}
				if n == 0 {
					g.SetMineWarning("Inventory full — free space to recover cargo.", 90, 2)
				}
			}
		}
		live = append(live, b)
	}
	g.lostCargo = live
}

// drawLostCargo draws surface cargo crates. Overworld only.
func (g *Game) drawLostCargo(screen *ebiten.Image) {
	if g.currentState != StateOverworld || len(g.lostCargo) == 0 {
		return
	}
	camX, camY := g.camera.Pos.X, g.camera.Pos.Y
	mult := GetOverworldLightMultiplier(g.TimeOfDay)
	for _, b := range g.lostCargo {
		if b == nil || !b.Active() {
			continue
		}
		b.Draw(screen, camX, camY, g.Ticks, mult)
	}
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
	if intensity > g.Shake.Intensity || g.Shake.Duration <= 0 {
		g.Shake.Intensity = intensity
	}
	if duration > g.Shake.Duration {
		g.Shake.Duration = duration
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

func (g *Game) hasSkiffInWorld() bool {
	if _, ok := g.ActiveVehicle.(*vehicle.Skiff); ok {
		return true
	}
	for _, v := range g.OverworldVehicles {
		if _, ok := v.(*vehicle.Skiff); ok {
			return true
		}
	}
	return false
}

// HasSaveFile returns true if a save file exists on disk.
func (g *Game) HasSaveFile() bool {
	return save.HasSaveFile()
}

// effectivePlayState returns the gameplay scene that should be recorded in a
// save, resolving the state that was active underneath the pause overlay.
func (g *Game) effectivePlayState() State {
	if g.currentState == StatePause && g.pauseState != nil {
		return g.pauseState.PriorState
	}
	return g.currentState
}

// SaveGame serializes and saves current game state to disk.
func (g *Game) SaveGame() error {
	sceneName := "Overworld"
	if g.effectivePlayState() == StateCave {
		sceneName = "Cave"
	}

	var savedVehicles []save.SavedVehicle
	for _, v := range g.OverworldVehicles {
		if v == nil {
			continue
		}
		var cargo, upg item.SavedInventory
		if v.GetCargo() != nil {
			cargo = v.GetCargo().SerializeState()
		}
		if v.GetUpgrades() != nil {
			upg = v.GetUpgrades().SerializeState()
		}
		savedVehicles = append(savedVehicles, save.SavedVehicle{
			Type:       v.GetName(),
			PosX:       v.GetPos().X,
			PosY:       v.GetPos().Y,
			Facing:     v.GetFacing(),
			Health:     v.GetHealth(),
			MaxHealth:  v.GetMaxHealth(),
			Battery:    v.GetBattery(),
			MaxBattery: v.GetMaxBattery(),
			Cargo:      cargo,
			Upgrades:   upg,
			IsActive:   (v == g.ActiveVehicle),
			Location:   "overworld",
		})
	}
	for trenchKey, vList := range g.CaveVehicles {
		for _, v := range vList {
			if v == nil {
				continue
			}
			var cargo, upg item.SavedInventory
			if v.GetCargo() != nil {
				cargo = v.GetCargo().SerializeState()
			}
			if v.GetUpgrades() != nil {
				upg = v.GetUpgrades().SerializeState()
			}
			savedVehicles = append(savedVehicles, save.SavedVehicle{
				Type:       v.GetName(),
				PosX:       v.GetPos().X,
				PosY:       v.GetPos().Y,
				Facing:     v.GetFacing(),
				Health:     v.GetHealth(),
				MaxHealth:  v.GetMaxHealth(),
				Battery:    v.GetBattery(),
				MaxBattery: v.GetMaxBattery(),
				Cargo:      cargo,
				Upgrades:   upg,
				IsActive:   (v == g.ActiveVehicle),
				Location:   trenchKey,
			})
		}
	}

	var unlockedRecipes []int
	var unlockedNames []string
	for i, rcp := range g.craftingRecipes {
		if rcp.Unlocked {
			unlockedRecipes = append(unlockedRecipes, i)
			if name := rcp.ResultName(); name != "" {
				unlockedNames = append(unlockedNames, name)
			}
		}
	}

	data := &save.SaveData{
		WorldSeed:       int64(g.world.Seed),
		TimeOfDay:       g.TimeOfDay,
		Ticks:           g.Ticks,
		TutorialActive:  g.TutorialActive,
		SceneState:      sceneName,
		LastOverworldX:  g.lastOverworldX,
		LastOverworldY:  g.lastOverworldY,
		ActiveTrenchX:   g.activeTrenchX,
		ActiveTrenchY:   g.activeTrenchY,
		ActiveTrenchKey: g.activeTrenchKey,
		Player: save.SavedPlayer{
			PosX:       g.player.Pos.X,
			PosY:       g.player.Pos.Y,
			Facing:     g.player.Facing,
			Health:     g.player.CurrentHealth,
			MaxHealth:  g.player.MaxHealth,
			Oxygen:     g.player.CurrentOxygen,
			MaxOxygen:  g.player.MaxOxygen,
			Stamina:    g.player.CurrentStamina,
			MaxStamina: g.player.MaxStamina,
			Energy:     g.player.CurrentEnergy,
			MaxEnergy:  g.player.MaxEnergy,
			ActiveSlot: g.player.ActiveSlot,
			Inventory:  g.player.Inventory.SerializeState(),
			Upgrades:   g.player.Upgrades.SerializeState(),
			Hotbar:     g.player.Hotbar.SerializeState(),
		},
		BaseStation: save.SavedBaseStation{
			PosX:     g.baseStation.Pos.X,
			PosY:     g.baseStation.Pos.Y,
			Storage:  g.baseStation.Storage.SerializeState(),
			Upgrades: g.baseStation.Upgrades.SerializeState(),
		},
		Vehicles:        savedVehicles,
		Story:           g.storyManager.SerializeState(),
		Exploration:         g.explorationTracker.SerializeState(),
		UnlockedRecipes:     unlockedRecipes,
		UnlockedRecipeNames: unlockedNames,
		LostCargo:           serializeLostCargo(g.lostCargo),
		Quests:              g.questManager.SerializeState(),
	}

	return save.SaveToFile(save.GetSavePath(), data)
}

// LoadSaveGame deserializes and restores game state from disk.
func (g *Game) LoadSaveGame() error {
	data, err := save.LoadFromFile(save.GetSavePath())
	if err != nil {
		return err
	}

	w := world.NewWorld(data.WorldSeed)
	g.world = w

	p := player.NewPlayer(data.Player.PosX, data.Player.PosY)
	p.Facing = data.Player.Facing
	if data.Player.MaxHealth > 0 {
		p.MaxHealth = data.Player.MaxHealth
	}
	p.CurrentHealth = data.Player.Health
	if data.Player.MaxOxygen > 0 {
		p.MaxOxygen = data.Player.MaxOxygen
	}
	p.CurrentOxygen = data.Player.Oxygen
	if data.Player.MaxStamina > 0 {
		p.MaxStamina = data.Player.MaxStamina
	}
	p.CurrentStamina = data.Player.Stamina
	if data.Player.MaxEnergy > 0 {
		p.MaxEnergy = data.Player.MaxEnergy
	}
	p.CurrentEnergy = data.Player.Energy
	p.ActiveSlot = data.Player.ActiveSlot

	p.Inventory = item.DeserializeInventory(data.Player.Inventory)
	p.Upgrades = item.DeserializeInventory(data.Player.Upgrades)
	p.Hotbar = item.DeserializeInventory(data.Player.Hotbar)
	p.RecalculateUpgrades()
	g.player = p

	baseStation := base.NewBaseStation(data.BaseStation.PosX, data.BaseStation.PosY)
	if data.BaseStation.Storage.Size > 0 || len(data.BaseStation.Storage.Slots) > 0 {
		baseStation.Storage = item.DeserializeInventory(data.BaseStation.Storage)
	}
	if data.BaseStation.Upgrades.Size > 0 || len(data.BaseStation.Upgrades.Slots) > 0 {
		baseStation.Upgrades = item.DeserializeInventory(data.BaseStation.Upgrades)
	}
	baseStation.RecalculateProperties()
	g.baseStation = baseStation

	g.ActiveVehicle = nil
	g.OverworldVehicles = nil
	g.CaveVehicles = make(map[string][]vehicle.Vehicle)

	for _, vData := range data.Vehicles {
		v := vehicle.NewVehicleByName(vData.Type, vData.PosX, vData.PosY)
		if v != nil {
			if vData.MaxHealth > 0 && vData.Health < vData.MaxHealth {
				v.TakeDamage(vData.MaxHealth - vData.Health)
			}
			if vData.MaxBattery > 0 && vData.Battery < vData.MaxBattery {
				v.RechargeBattery(vData.Battery - v.GetBattery())
			}
			if v.GetCargo() != nil && (vData.Cargo.Size > 0 || len(vData.Cargo.Slots) > 0) {
				*v.GetCargo() = *item.DeserializeInventory(vData.Cargo)
			}
			if v.GetUpgrades() != nil && (vData.Upgrades.Size > 0 || len(vData.Upgrades.Slots) > 0) {
				*v.GetUpgrades() = *item.DeserializeInventory(vData.Upgrades)
			}

			if vData.Location == "overworld" || vData.Location == "" {
				g.OverworldVehicles = append(g.OverworldVehicles, v)
			} else {
				g.CaveVehicles[vData.Location] = append(g.CaveVehicles[vData.Location], v)
			}

			if vData.IsActive {
				g.ActiveVehicle = v
			}
		}
	}

	if g.storyManager == nil {
		g.storyManager = story.NewStoryManager()
	}
	g.storyManager.DeserializeState(data.Story)

	if g.questManager == nil {
		g.questManager = quest.NewQuestManager()
	}
	if len(data.Quests.Categories) > 0 || len(data.Quests.Quests) > 0 {
		g.questManager.DeserializeState(data.Quests)
	}

	g.craftingRecipes = scene.DefaultCraftingRecipes()
	applyUnlockedRecipes(g.craftingRecipes, data.UnlockedRecipeNames, data.UnlockedRecipes)
	g.lostCargo = deserializeLostCargo(data.LostCargo)

	// Always reset per-world cave caches so a prior session cannot leak in.
	g.caveNodes = make(map[string][]resource.Resource)
	g.caveEntities = make(map[string][]entity.CaveEntity)
	g.showInventory = false
	g.menuOpenedAnywhere = false

	g.explorationTracker = exploration.NewTracker(w.Width, w.Height)
	g.explorationTracker.DeserializeState(data.Exploration)

	g.TimeOfDay = data.TimeOfDay
	g.Ticks = data.Ticks
	g.TutorialActive = data.TutorialActive

	g.lastOverworldX = data.LastOverworldX
	g.lastOverworldY = data.LastOverworldY
	g.activeTrenchX = data.ActiveTrenchX
	g.activeTrenchY = data.ActiveTrenchY
	g.activeTrenchKey = data.ActiveTrenchKey

	g.camera = camera.NewCamera(p.Pos.X, p.Pos.Y)
	g.camera.CenterOn(p.Pos.X, p.Pos.Y, p.Width, p.Height)

	g.overworldState = NewOverworldScene(w)

	if data.SceneState == "Cave" {
		// Rebuild the cave under the player, keeping their saved position.
		savedX, savedY := p.Pos.X, p.Pos.Y
		savedFacing := p.Facing
		g.hydrateCave(data.ActiveTrenchX, data.ActiveTrenchY)
		if data.ActiveTrenchKey != "" {
			g.activeTrenchKey = data.ActiveTrenchKey
		}
		p.Pos.X = savedX
		p.Pos.Y = savedY
		p.Facing = savedFacing
		p.Vel = gvec.Vec2{}
		g.camera.CenterOn(p.Pos.X, p.Pos.Y, p.Width, p.Height)
		g.TransitionTo(g.caveState)
	} else {
		g.TransitionTo(g.overworldState)
	}

	g.SetMineWarning("GAME LOADED", 120, 1)
	return nil
}

func applyUnlockedRecipes(recipes []Recipe, names []string, indexes []int) {
	if len(names) > 0 {
		want := make(map[string]struct{}, len(names))
		for _, name := range names {
			want[name] = struct{}{}
		}
		for i := range recipes {
			if _, ok := want[recipes[i].ResultName()]; ok {
				recipes[i].Unlocked = true
			}
		}
		return
	}
	for _, idx := range indexes {
		if idx >= 0 && idx < len(recipes) {
			recipes[idx].Unlocked = true
		}
	}
}

func serializeLostCargo(beacons []*entity.LostCargoBeacon) []save.SavedLostCargo {
	if len(beacons) == 0 {
		return nil
	}
	out := make([]save.SavedLostCargo, 0, len(beacons))
	for _, b := range beacons {
		if b == nil || !b.Active() {
			continue
		}
		out = append(out, save.SavedLostCargo{
			PosX:          b.Pos.X,
			PosY:          b.Pos.Y,
			LifetimeTicks: b.LifetimeTicks,
			Cargo:         b.Cargo.SerializeState(),
		})
	}
	return out
}

func deserializeLostCargo(saved []save.SavedLostCargo) []*entity.LostCargoBeacon {
	if len(saved) == 0 {
		return nil
	}
	out := make([]*entity.LostCargoBeacon, 0, len(saved))
	for _, s := range saved {
		cargo := item.DeserializeInventory(s.Cargo)
		if cargo == nil || cargo.IsEmpty() || s.LifetimeTicks <= 0 {
			continue
		}
		out = append(out, &entity.LostCargoBeacon{
			Pos:           gvec.Vec2{X: s.PosX, Y: s.PosY},
			Cargo:         cargo,
			LifetimeTicks: s.LifetimeTicks,
		})
	}
	return out
}
