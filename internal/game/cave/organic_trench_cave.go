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
	// Deep-sea grotto dark background
	screen.Fill(color.RGBA{10, 8, 16, 255})
}

func (c *OrganicTrenchCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float32(tx*config.TileSize - int(camX))
				sy := float32(ty*config.TileSize - int(camY))

				var rockColor, strokeColor color.RGBA
				if ty < 40 {
					// Biome 1: Mid-Depth (Cyan/Teal) - Luminous Pneumatophore Grotto
					bandRatio := float64(ty) / 40.0
					r := uint8(max(8, 22-14*bandRatio))
					g := uint8(max(24, 64-40*bandRatio))
					b := uint8(max(32, 78-46*bandRatio))
					rockColor = color.RGBA{r, g, b, 255}
					strokeColor = color.RGBA{r + 20, g + 40, b + 48, 255}
				} else if ty < 80 {
					// Biome 2: Deep (Dark Grey/Orange) - Silicate Smoker Trenches
					bandRatio := float64(ty-40) / 40.0
					r := uint8(max(25, 45-20*bandRatio))
					g := uint8(max(20, 32-12*bandRatio))
					b := uint8(max(18, 26-8*bandRatio))
					rockColor = color.RGBA{r, g, b, 255}
					strokeColor = color.RGBA{uint8(max(80, 150-70*bandRatio)), 65, 40, 255}
				} else {
					// Biome 3: Abyssal (Vantablack/White) - Benthic Brine-Falls
					rockColor = color.RGBA{5, 5, 8, 255}
					strokeColor = color.RGBA{210, 210, 220, 255}
				}

				h := hashCoords(tx, ty)
				if ty <= 8 {
					blendProb := float64(8-ty) / 9.0 * 0.70
					if float64(h%100)/100.0 < blendProb {
						// Abyssal shallow slate rock
						rockColor = color.RGBA{42, 50, 72, 255}
						strokeColor = color.RGBA{65, 78, 108, 255}
					}
				}

				vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, rockColor, false)
				vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 0.5, strokeColor, false)
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
