package cave

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/gvec"
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

				vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, rockColor, false)
				vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 0.5, strokeColor, false)
			}
		}
	}
}

func (c *OrganicTrenchCave) GenerateEntities(seed int64) []entity.CaveEntity {
	grid := c.Grid
	r := rand.New(rand.NewSource(seed))
	var entities []entity.CaveEntity

	gridW := len(grid)
	gridH := len(grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			if grid[tx][ty] {
				continue
			}

			// Biome 1: Mid-Depth (ty 4–40) - Grotto
			if ty >= 4 && ty < 40 {
				hasAdjacentWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
				if hasAdjacentWall && r.Float64() < 0.08 {
					entities = append(entities, &entity.ShatterBulb{
						BaseEntity: entity.BaseEntity{
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 24, Y: 24},
							Active:     true,
						},
					})
				}
				if grid[tx][ty-1] && r.Float64() < 0.04 {
					entities = append(entities, &entity.FalseBulbSnare{
						BaseEntity: entity.BaseEntity{
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + 4},
							Dimensions: gvec.Vec2{X: 24, Y: 32},
							Active:     true,
						},
						State: 0,
					})
				}
			}

			// Biome 2: Deep (ty 40–80) - Smoker Trenches
			if ty >= 40 && ty < 80 {
				if r.Float64() < 0.05 {
					var dir string
					if grid[tx][ty+1] {
						dir = "up"
					} else if grid[tx][ty-1] {
						dir = "down"
					} else if grid[tx-1][ty] {
						dir = "right"
					} else if grid[tx+1][ty] {
						dir = "left"
					}
					if dir != "" {
						entities = append(entities, &entity.BrimstoneSiphon{
							BaseEntity: entity.BaseEntity{
								Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-32)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-32)/2.0},
								Dimensions: gvec.Vec2{X: 32, Y: 32},
								Active:     true,
							},
							Direction: dir,
							Timer:     r.Intn(120),
						})
					}
				}
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.015 {
					entities = append(entities, &entity.ThermoclineRammer{
						BaseEntity: entity.BaseEntity{
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-36)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 36, Y: 24},
							Active:     true,
						},
						Facing: r.Float64() * math.Pi * 2,
					})
				}
			}

			// Spawn ShockKelp in OrganicTrenchCave (Mid-Depth and Deep)
			if (ty >= 4 && ty < 80) && ty < gridH-2 && grid[tx][ty+1] && r.Float64() < 0.18 {
				height := 32.0 + r.Float64()*48.0
				entities = append(entities, entity.NewShockKelp(
					float64(tx*config.TileSize)+float64(config.TileSize-16)/2.0,
					float64(ty*config.TileSize)+float64(config.TileSize)-height,
					height,
					"floor",
				))
			}

			// Biome 3: Abyssal (ty 80+) - Brine Falls
			if ty >= 80 && ty < gridH-1 {
				if grid[tx][ty+1] && r.Float64() < 0.10 {
					entities = append(entities, &entity.NerveMat{
						BaseEntity: entity.BaseEntity{
							Pos:        gvec.Vec2{X: float64(tx * config.TileSize), Y: float64(ty*config.TileSize) + float64(config.TileSize-12)},
							Dimensions: gvec.Vec2{X: float64(config.TileSize), Y: 12},
							Active:     true,
						},
					})
				}
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.012 {
					hasWeaverNearby := false
					for _, ent := range entities {
						ee, ok := ent.(*entity.ElectroWeaver)
						if ok && math.Abs(ee.GetPos().X-float64(tx*config.TileSize)) < 500 {
							hasWeaverNearby = true
							break
						}
					}
					if !hasWeaverNearby {
						entities = append(entities, &entity.ElectroWeaver{
							BaseEntity: entity.BaseEntity{
								Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-40)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-20)/2.0},
								Dimensions: gvec.Vec2{X: 40, Y: 20},
								Active:     true,
							},
						})
					}
				}
			}

			// Spawn decorative trench corals (12% chance near any solid face)
			var coralAttachments []string
			if grid[tx][ty+1] {
				coralAttachments = append(coralAttachments, "floor")
			}
			if grid[tx][ty-1] {
				coralAttachments = append(coralAttachments, "ceiling")
			}
			if grid[tx-1][ty] {
				coralAttachments = append(coralAttachments, "left")
			}
			if grid[tx+1][ty] {
				coralAttachments = append(coralAttachments, "right")
			}

			if len(coralAttachments) > 0 && r.Float64() < 0.12 {
				attach := coralAttachments[r.Intn(len(coralAttachments))]
				variant := r.Intn(2) // 2 variants for trench
				
				cx := float64(tx * config.TileSize)
				cy := float64(ty * config.TileSize)
				
				switch attach {
				case "floor":
					cx += float64(config.TileSize-24) / 2.0
					cy += float64(config.TileSize - 24)
				case "ceiling":
					cx += float64(config.TileSize-24) / 2.0
				case "left":
					cy += float64(config.TileSize-24) / 2.0
				case "right":
					cx += float64(config.TileSize - 24)
					cy += float64(config.TileSize-24) / 2.0
				}
				
				entities = append(entities, entity.NewCoral(cx, cy, entity.CoralBiomeTrench, attach, variant, r))
			}
		}
	}

	return entities
}

func (c *OrganicTrenchCave) GenerateResources(seed int64) []resource.Resource {
	return resource.GenerateResourceNodes(c.Grid, seed)
}

func (c *OrganicTrenchCave) GetAmbientColor(lightMult float64) []float32 {
	return []float32{0.01, 0.01, 0.03, 0.97}
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
