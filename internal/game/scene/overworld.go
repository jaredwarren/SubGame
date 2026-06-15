package scene

import (
	"image/color"
	_ "image/png"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/assets"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

// OverworldContext defines the narrow context interface required by OverworldScene.
type OverworldContext interface {
	GetInput() InputSource
	GetPlayer() *player.Player
	GetCamera() *camera.Camera
	GetWorld() *world.World
	GetBaseStation() *base.BaseStation
	GetTimeOfDay() float64
	GetTicks() float64
	GetActiveVehicle() vehicle.Vehicle
	GetAllCaveVehicles() map[string][]vehicle.Vehicle
	GetActiveTrenchKey() string
	GetActiveTrenchCoords() (x, y int)
	IsInventoryOpen() bool
	SetInventoryOpen(v bool)
	EnterCave(tx, ty int)
	TransitionToPDA()
	SetCurrentState(s State)
	SpawnPlankton(x, y float64)
	SpawnDebris(x, y float64, clr color.RGBA)
	SpawnBubble(x, y float64)
	TriggerScreenShake(duration int, intensity float64)
	SetMineWarning(msg string, duration, level int)
	GetDeathReason() string
	SetDeathReason(reason string)
	DestroyOverworldVehicle(v vehicle.Vehicle)
}

// OverworldScene manages the top-down surface sailing view.
type OverworldScene struct {
	World        *world.World
	whirlpool    *oe.Whirlpool
	crates       []*oe.FloatingCrate
	vents        []*oe.ThermalVent
	fish         []*oe.CosmeticFish
	tileTextures map[world.TileType]*ebiten.Image
	initialized  bool

	// Cached offscreen static image details
	cachedStaticImage *ebiten.Image
	cachedChunkX      int
	cachedChunkY      int
	hasCache          bool
}

// NewOverworldScene creates a new OverworldScene.
func NewOverworldScene(w *world.World) *OverworldScene {
	return &OverworldScene{World: w}
}

func (o *OverworldScene) getTileTexture(tileType world.TileType) *ebiten.Image {
	if o.tileTextures == nil {
		o.tileTextures = map[world.TileType]*ebiten.Image{
			world.TileTrench:   trenchTexture,
			world.TileWreckage: wreckageTexture,
		}
	}
	if tileType == world.TileShockKelpCave && o.tileTextures[world.TileShockKelpCave] == nil {
		o.tileTextures[world.TileShockKelpCave] = cave.GenerateShockKelpReefTexture()
	}
	return o.tileTextures[tileType]
}

func (o *OverworldScene) OnEnter(g GameContext) {
	o.onEnter(g)
}

func (o *OverworldScene) onEnter(g OverworldContext) {
	g.SetCurrentState(StateOverworld)
	if o.whirlpool == nil {
		o.whirlpool = oe.NewWhirlpool(g.GetWorld().Seed)
		rng := rand.New(rand.NewSource(g.GetWorld().Seed + 997))
		pos := o.FindSafeWhirlpoolSpawnPos(g.GetBaseStation().Pos, rng)
		o.whirlpool.Relocate(pos)
	}
}

func (o *OverworldScene) OnExit(g GameContext) {
	o.onExit(g)
}

func (o *OverworldScene) onExit(g OverworldContext) {}

func (o *OverworldScene) Update(g GameContext) error {
	return o.update(g)
}

func (o *OverworldScene) Draw(g GameContext, screen *ebiten.Image) {
	o.draw(g, screen)
}

var (
	trenchTexture         *ebiten.Image
	trenchTextureLoaded   bool
	wreckageTexture       *ebiten.Image
	wreckageTextureLoaded bool
	shockKelpTexture      *ebiten.Image
)


// LoadAssets preloads and chroma-keys all overworld tile textures.
func LoadAssets() {
	// 1. Trench Texture
	{
		img, err := assets.LoadChromaKeyedImage("trench_surface")
		if err != nil {
			log.Printf("Error: Failed to load trench surface: %v", err)
		} else {
			trenchTexture = img
			trenchTextureLoaded = true
		}
	}

	// 2. Wreckage Texture
	{
		img, err := assets.LoadChromaKeyedImage("wreckage_surface")
		if err != nil {
			log.Printf("Error: Failed to load wreckage surface: %v", err)
		} else {
			wreckageTexture = img
			wreckageTextureLoaded = true
		}
	}
}

// IsSolid checks if the proposed bounding box overlaps with solid land.
func (o *OverworldScene) IsSolid(x, y, w, h float64) bool {
	x1, x2, y1, y2 := tileRange(x, y, w, h, config.TileSize)

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
				continue
			}
			if o.World.OverworldMap[tx][ty] == world.TileLand {
				return true
			}
		}
	}
	return false
}
