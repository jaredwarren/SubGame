package scene

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/assets"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/world"
)

// OverworldScene manages the top-down surface sailing view.
type OverworldScene struct {
	World       *world.World
	whirlpool   *oe.Whirlpool
	crates      []*oe.FloatingCrate
	vents       []*oe.ThermalVent
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
	return o.tileTextures[tileType]
}

func (o *OverworldScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateOverworld)
	if o.whirlpool == nil {
		o.whirlpool = oe.NewWhirlpool(g.GetWorld().Seed)
		rng := rand.New(rand.NewSource(g.GetWorld().Seed + 997))
		pos := o.FindSafeWhirlpoolSpawnPos(g.GetBaseStation().Pos, rng)
		o.whirlpool.Relocate(pos)
	}
}

func (o *OverworldScene) OnExit(g GameContext) {}

var (
	trenchTexture         *ebiten.Image
	trenchTextureLoaded   bool
	wreckageTexture       *ebiten.Image
	wreckageTextureLoaded bool
)

func removeChromaKey(img image.Image) *ebiten.Image {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	for i := 0; i < len(rgba.Pix); i += 4 {
		ru := rgba.Pix[i]
		gu := rgba.Pix[i+1]
		bu := rgba.Pix[i+2]

		if gu > 140 && ru < 100 && bu < 100 {
			rgba.Pix[i] = 0
			rgba.Pix[i+1] = 0
			rgba.Pix[i+2] = 0
			rgba.Pix[i+3] = 0
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

// LoadAssets preloads and chroma-keys all overworld tile textures.
func LoadAssets() {
	// 1. Trench Texture
	{
		img, _, err := image.Decode(bytes.NewReader(assets.TrenchSurfacePNG))
		if err != nil {
			log.Printf("Error: Failed to decode trench surface: %v", err)
		} else {
			trenchTexture = removeChromaKey(img)
			trenchTextureLoaded = true
		}
	}

	// 2. Wreckage Texture
	{
		img, _, err := image.Decode(bytes.NewReader(assets.WreckageSurfacePNG))
		if err != nil {
			log.Printf("Error: Failed to decode wreckage surface: %v", err)
		} else {
			wreckageTexture = removeChromaKey(img)
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
