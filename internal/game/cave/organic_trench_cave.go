package cave

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

type OrganicTrenchCave struct {
	Grid [][]bool
}

func NewOrganicTrenchCave(grid [][]bool) *OrganicTrenchCave {
	return &OrganicTrenchCave{Grid: grid}
}

func (c *OrganicTrenchCave) GetCaveType() CaveType { return CaveOrganicTrench }
func (c *OrganicTrenchCave) GetGrid() [][]bool     { return c.Grid }

func (c *OrganicTrenchCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Deep-sea grotto dark background: subtle depth darkening from deep indigo at top to dark abyss near bottom
	depthFrac := 0.0
	if maxDepth > 0 {
		depthFrac = min(1.0, max(0.0, camY/maxDepth))
	}
	r := uint8(max(3, int(10-7*depthFrac)))
	g := uint8(max(4, int(12-8*depthFrac)))
	b := uint8(max(10, int(22-12*depthFrac)))
	screen.Fill(color.RGBA{r, g, b, 255})
}

func (c *OrganicTrenchCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float32(tx*config.TileSize - int(camX))
				sy := float32(ty*config.TileSize - int(camY))

				var rockColor, strokeColor, seamColor color.RGBA
				if ty < 40 {
					// Biome 1: Mid-Depth (Cyan/Teal) - Luminous Pneumatophore Grotto
					bandRatio := float64(ty) / 40.0
					r := uint8(max(10, 20-10*bandRatio))
					g := uint8(max(26, 52-26*bandRatio))
					b := uint8(max(38, 70-32*bandRatio))
					rockColor = color.RGBA{r, g, b, 255}
					strokeColor = color.RGBA{r + 20, g + 45, b + 55, 255}
					seamColor = color.RGBA{60, 210, 235, 220}
				} else if ty < 80 {
					// Biome 2: Deep (Dark Grey/Orange) - Silicate Smoker Trenches
					bandRatio := float64(ty-40) / 40.0
					r := uint8(max(24, 40-16*bandRatio))
					g := uint8(max(20, 30-10*bandRatio))
					b := uint8(max(22, 28-6*bandRatio))
					rockColor = color.RGBA{r, g, b, 255}
					strokeColor = color.RGBA{uint8(max(90, 150-60*bandRatio)), 70, 45, 255}
					seamColor = color.RGBA{235, 125, 55, 220}
				} else {
					// Biome 3: Abyssal (Midnight Basalt / Electric Azure) - Benthic Brine-Falls
					rockColor = color.RGBA{16, 18, 32, 255}
					strokeColor = color.RGBA{38, 85, 150, 255}
					seamColor = color.RGBA{75, 160, 255, 230}
				}

				h := hashCoords(tx, ty)
				if ty <= 8 {
					blendProb := float64(8-ty) / 9.0 * 0.70
					if float64(h%100)/100.0 < blendProb {
						// Abyssal shallow slate rock
						rockColor = color.RGBA{32, 38, 58, 255}
						strokeColor = color.RGBA{50, 110, 150, 255}
						seamColor = color.RGBA{65, 180, 225, 220}
					}
				}

				// 1. Fill base basalt block
				tileSize := float32(config.TileSize)
				vector.FillRect(screen, sx, sy, tileSize, tileSize, rockColor, false)

				// 2. Geological bedding strata
				strataY := sy + float32((h%10)+3)
				strataDark := color.RGBA{
					uint8(max(0, int(rockColor.R)-6)),
					uint8(max(0, int(rockColor.G)-6)),
					uint8(max(0, int(rockColor.B)-6)),
					255,
				}
				vector.StrokeLine(screen, sx+2, strataY, sx+tileSize-2, strataY+float32((h>>4)%3-1), 1.0, strataDark, false)

				// 3. Embedded luminous crystalline speck (on ~25% of tiles)
				if (h % 4) == 0 {
					fx := sx + float32((h>>8)%10+3)
					fy := sy + float32((h>>12)%10+3)
					vector.FillCircle(screen, fx, fy, 1.0, seamColor, false)
					vector.FillCircle(screen, fx, fy, 0.4, color.White, false)
				}

				// 4. Outer tile grid stroke
				vector.StrokeRect(screen, sx, sy, tileSize, tileSize, 0.5, strokeColor, false)

				// 5. Luminous crystalline edge seams on faces exposed to open water
				gridH := len(c.Grid[0])
				gridW := len(c.Grid)
				if ty > 0 && !c.Grid[tx][ty-1] {
					vector.StrokeLine(screen, sx, sy, sx+tileSize, sy, 1.2, seamColor, false)
				}
				if ty < gridH-1 && !c.Grid[tx][ty+1] {
					vector.StrokeLine(screen, sx, sy+tileSize, sx+tileSize, sy+tileSize, 1.2, seamColor, false)
				}
				if tx > 0 && !c.Grid[tx-1][ty] {
					vector.StrokeLine(screen, sx, sy, sx, sy+tileSize, 1.2, seamColor, false)
				}
				if tx < gridW-1 && !c.Grid[tx+1][ty] {
					vector.StrokeLine(screen, sx+tileSize, sy, sx+tileSize, sy+tileSize, 1.2, seamColor, false)
				}
			}
		}
	}
}

func (c *OrganicTrenchCave) GenerateEntities(seed int64) []entity.CaveEntity {
	return GenerateDeepEntitiesFromSpec(Spec(CaveOrganicTrench), c.Grid, seed)
}

func (c *OrganicTrenchCave) GenerateResources(seed int64) []resource.Resource {
	return resource.GenerateResourceNodes(c.Grid, seed)
}

func (c *OrganicTrenchCave) GetAmbientColor(lightMult float64) [4]float32 {
	return AmbientColor(CaveOrganicTrench)
}

func GenerateOrganicTrenchGrid(r *rand.Rand) [][]bool {
	const (
		w = CaveWidth
		h = CaveHeight
	)
	// 1. Generate upper shallow cave (Cellular Automata)
	shallowCave := GenerateCellularCave(w, SplitY, 0.42, 4, r)

	// 2. Generate lower deep crevice cave (Drunkard's Walk)
	deepCave := GenerateDrunkardCave(w, h-SplitY, r)

	// 3. Instantiate full cave grid
	caveGrid := make([][]bool, w)
	for x := range w {
		caveGrid[x] = make([]bool, h)
	}

	// 4. Merge upper and lower caves
	for x := range w {
		for y := range h {
			if y < SplitY {
				caveGrid[x][y] = shallowCave[x][y]
			} else {
				caveGrid[x][y] = deepCave[x][y-SplitY]
			}
		}
	}

	// 5. Connect the two halves at the split boundary
	// Carve a vertical connecting shaft in the middle to ensure pathability
	const shaftHalfWidth = 2
	for y := SplitY - 8; y < SplitY+8; y++ {
		for x := (w / 2) - shaftHalfWidth; x <= (w/2)+shaftHalfWidth; x++ {
			if x > 0 && x < w-1 && y > 0 && y < h-1 {
				caveGrid[x][y] = false // Carve path
			}
		}
	}

	// 6. Ensure entrance at top center is open for diving player
	for y := 0; y < 5; y++ {
		for x := (w / 2) - 3; x <= (w/2)+3; x++ {
			if x > 0 && x < w-1 && y < h-1 {
				caveGrid[x][y] = false
			}
		}
	}

	return caveGrid
}
