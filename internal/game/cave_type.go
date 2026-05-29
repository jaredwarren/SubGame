package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type CaveType int

const (
	CaveOrganicShallow CaveType = iota
	CaveOrganicTrench
	CaveWreckage
	CaveVoid
)

// Cave defines the interface for different cave types.
type Cave interface {
	GetCaveType() CaveType
	GetGrid() [][]bool
	DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64)
	DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int)
	GenerateEntities(seed int64) []CaveEntity
	GenerateResources(seed int64) []resource.Resource
}

// -----------------------------------------------------------------------------
// 1. VOID CAVE (Ecological Dead Zone Cavern)
// -----------------------------------------------------------------------------

type VoidCave struct{}

func NewVoidCave() *VoidCave {
	return &VoidCave{}
}

func (c *VoidCave) GetCaveType() CaveType { return CaveVoid }
func (c *VoidCave) GetGrid() [][]bool     { return nil }

func (c *VoidCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Endless crushing dark void
	screen.Fill(color.RGBA{2, 3, 6, 255})
}

func (c *VoidCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	// No tiles in the void
}

func (c *VoidCave) GenerateEntities(seed int64) []CaveEntity {
	return nil
}

func (c *VoidCave) GenerateResources(seed int64) []resource.Resource {
	return nil
}

// -----------------------------------------------------------------------------
// 2. SHALLOW SEABED CAVE (Sandy Reefs)
// -----------------------------------------------------------------------------

type ShallowSeabedCave struct {
	Grid [][]bool
}

func NewShallowSeabedCave(grid [][]bool) *ShallowSeabedCave {
	return &ShallowSeabedCave{Grid: grid}
}

func (c *ShallowSeabedCave) GetCaveType() CaveType { return CaveOrganicShallow }
func (c *ShallowSeabedCave) GetGrid() [][]bool     { return c.Grid }

func (c *ShallowSeabedCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Surface base color
	baseR := float64(10) + float64(30)*lightMult  // 10 (night) → 40 (day)
	baseG := float64(40) + float64(80)*lightMult  // 40 (night) → 120 (day)
	baseB := float64(100) + float64(80)*lightMult // 100 (night) → 180 (day)

	maxDarken := 0.45 + (1.0-lightMult)*0.45

	const stripH = float32(6)
	for sy := float32(0); sy < float32(ScreenHeight); sy += stripH {
		worldY := camY + float64(sy)
		depthFrac := 0.0
		if worldY > 0 {
			depthFrac = worldY / maxDepth
			if depthFrac > 1 {
				depthFrac = 1
			}
		}
		darkFactor := 1.0 - depthFrac*maxDarken
		sc := color.RGBA{
			R: uint8(baseR * darkFactor),
			G: uint8(baseG * darkFactor),
			B: uint8(baseB * darkFactor),
			A: 255,
		}
		vector.FillRect(screen, 0, sy, float32(ScreenWidth), stripH, sc, false)
	}
}

func (c *ShallowSeabedCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float32(tx*TileSize - int(camX))
				sy := float32(ty*TileSize - int(camY))
				rockColor := color.RGBA{180, 155, 100, 255}
				strokeColor := color.RGBA{210, 185, 120, 255}
				vector.FillRect(screen, sx, sy, TileSize, TileSize, rockColor, false)
				vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, strokeColor, false)
			}
		}
	}
}

func (c *ShallowSeabedCave) GenerateEntities(seed int64) []CaveEntity {
	return GenerateCaveEntities(c.Grid, seed, true)
}


func (c *ShallowSeabedCave) GenerateResources(seed int64) []resource.Resource {
	// Standard depth generation delegate (depth tier 0)
	return resource.GenerateResourceNodes(c.Grid, seed)
}

// -----------------------------------------------------------------------------
// 3. ORGANIC TRENCH CAVE (Deep Crevice / Abyssal)
// -----------------------------------------------------------------------------

type OrganicTrenchCave struct {
	Grid [][]bool
}

func NewOrganicTrenchCave(grid [][]bool) *OrganicTrenchCave {
	return &OrganicTrenchCave{Grid: grid}
}

func (c *OrganicTrenchCave) GetCaveType() CaveType { return CaveOrganicTrench }
func (c *OrganicTrenchCave) GetGrid() [][]bool     { return c.Grid }

func (c *OrganicTrenchCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Deep-sea grotto dark background
	screen.Fill(color.RGBA{10, 8, 16, 255})
}

func (c *OrganicTrenchCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float32(tx*TileSize - int(camX))
				sy := float32(ty*TileSize - int(camY))

				var rockColor, strokeColor color.RGBA
				if ty < 40 {
					// Biome 1: Mid-Depth (Cyan/Teal) - Luminous Pneumatophore Grotto
					bandRatio := float64(ty) / 40.0
					r := uint8(math.Max(8, 22-14*bandRatio))
					g := uint8(math.Max(24, 64-40*bandRatio))
					b := uint8(math.Max(32, 78-46*bandRatio))
					rockColor = color.RGBA{r, g, b, 255}
					strokeColor = color.RGBA{r + 20, g + 40, b + 48, 255}
				} else if ty < 80 {
					// Biome 2: Deep (Dark Grey/Orange) - Silicate Smoker Trenches
					bandRatio := float64(ty-40) / 40.0
					r := uint8(math.Max(25, 45-20*bandRatio))
					g := uint8(math.Max(20, 32-12*bandRatio))
					b := uint8(math.Max(18, 26-8*bandRatio))
					rockColor = color.RGBA{r, g, b, 255}
					strokeColor = color.RGBA{uint8(math.Max(80, 150-70*bandRatio)), 65, 40, 255}
				} else {
					// Biome 3: Abyssal (Vantablack/White) - Benthic Brine-Falls
					rockColor = color.RGBA{5, 5, 8, 255}
					strokeColor = color.RGBA{210, 210, 220, 255}
				}

				vector.FillRect(screen, sx, sy, TileSize, TileSize, rockColor, false)
				vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, strokeColor, false)
			}
		}
	}
}

func (c *OrganicTrenchCave) GenerateEntities(seed int64) []CaveEntity {
	// Re-uses original GenerateCaveEntities but forces organic trench spawning (isShallow = false)
	return GenerateCaveEntities(c.Grid, seed, false)
}

func (c *OrganicTrenchCave) GenerateResources(seed int64) []resource.Resource {
	return resource.GenerateResourceNodes(c.Grid, seed)
}

// -----------------------------------------------------------------------------
// 4. WRECKAGE CORRIDOR CAVE (Sunken Vessel Decks)
// -----------------------------------------------------------------------------

type WreckageCorridorCave struct {
	Grid [][]bool
}

func NewWreckageCorridorCave(grid [][]bool) *WreckageCorridorCave {
	return &WreckageCorridorCave{Grid: grid}
}

func (c *WreckageCorridorCave) GetCaveType() CaveType { return CaveWreckage }
func (c *WreckageCorridorCave) GetGrid() [][]bool     { return c.Grid }

func (c *WreckageCorridorCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Dark, artificial metallic vessel interior background
	screen.Fill(color.RGBA{14, 15, 20, 255})

	// Optional: render faint background steel pillars or grid lines for industrial feel
	const lineGap = 40.0
	offsetX := float32(math.Mod(camY*0.1, lineGap))
	for x := float32(0); x < float32(ScreenWidth); x += lineGap {
		vector.StrokeLine(screen, x, 0, x, float32(ScreenHeight), 0.8, color.RGBA{20, 24, 30, 255}, false)
	}
	for y := float32(0); y < float32(ScreenHeight); y += lineGap {
		sy := y - offsetX
		vector.StrokeLine(screen, 0, sy, float32(ScreenWidth), sy, 0.8, color.RGBA{20, 24, 30, 255}, false)
	}
}

func (c *WreckageCorridorCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float32(tx*TileSize - int(camX))
				sy := float32(ty*TileSize - int(camY))

				// Steel gray bulkhead panels
				rockColor := color.RGBA{52, 58, 68, 255}
				strokeColor := color.RGBA{85, 80, 75, 255}

				vector.FillRect(screen, sx, sy, TileSize, TileSize, rockColor, false)
				vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 1.2, strokeColor, false)

				// Draw diagonal yellow-and-black hazard warning lines along bulkheads that border corridors
				hasBorder := false
				if tx > 0 && !c.Grid[tx-1][ty] { hasBorder = true }
				if tx < len(c.Grid)-1 && !c.Grid[tx+1][ty] { hasBorder = true }
				if ty > 0 && !c.Grid[tx][ty-1] { hasBorder = true }
				if ty < len(c.Grid[0])-1 && !c.Grid[tx][ty+1] { hasBorder = true }

				if hasBorder {
					stripeColor := color.RGBA{215, 175, 30, 160} // Rusted safety yellow
					// Draw diagonal warning lines across the bulkhead
					vector.StrokeLine(screen, sx, sy+8, sx+8, sy, 2.0, stripeColor, false)
					vector.StrokeLine(screen, sx, sy+24, sx+24, sy, 2.0, stripeColor, false)
					vector.StrokeLine(screen, sx+16, sy+TileSize, sx+TileSize, sy+16, 2.0, stripeColor, false)
					vector.StrokeLine(screen, sx+32, sy+TileSize, sx+TileSize, sy+32, 2.0, stripeColor, false)
				}
			}
		}
	}
}

func (c *WreckageCorridorCave) GenerateEntities(seed int64) []CaveEntity {
	r := rand.New(rand.NewSource(seed))
	var entities []CaveEntity

	gridW := len(c.Grid)
	gridH := len(c.Grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			if c.Grid[tx][ty] {
				continue
			}
			// Wreckage caves only spawn static Shatter-bulb plants (emergency lighting bulbs) on walls
			hasAdjacentWall := c.Grid[tx-1][ty] || c.Grid[tx+1][ty] || c.Grid[tx][ty-1] || c.Grid[tx][ty+1]
			if hasAdjacentWall && r.Float64() < 0.05 {
				entities = append(entities, &ShatterBulb{
					BaseEntity: BaseEntity{
						Type:       EntShatterBulb,
						Pos:        gvec.Vec2{X: float64(tx*TileSize) + float64(TileSize-24)/2.0, Y: float64(ty*TileSize) + float64(TileSize-24)/2.0},
						Dimensions: gvec.Vec2{X: 24, Y: 24},
						Active:     true,
					},
				})
			}
		}
	}
	return entities
}

func (c *WreckageCorridorCave) GenerateResources(seed int64) []resource.Resource {
	// Scrap nodes are generated inside wreckage caves instead of mineral nodes
	return resource.GenerateWreckageResources(c.Grid, seed)
}
