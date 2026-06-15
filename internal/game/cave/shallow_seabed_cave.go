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

type ShallowSeabedCave struct {
	Grid       [][]bool
	tileImages []*ebiten.Image
}

func NewShallowSeabedCave(grid [][]bool) *ShallowSeabedCave {
	c := &ShallowSeabedCave{
		Grid:       grid,
		tileImages: make([]*ebiten.Image, 8),
	}
	c.preRenderTiles()
	return c
}

func (c *ShallowSeabedCave) preRenderTiles() {
	rockColor := color.RGBA{180, 155, 100, 255}
	strokeColor := color.RGBA{210, 185, 120, 255}
	darkSandColor := color.RGBA{150, 130, 80, 255}
	lightSandColor := color.RGBA{215, 190, 125, 255}

	for idx := range c.tileImages {
		img := ebiten.NewImage(config.TileSize, config.TileSize)
		// 1. Fill base rock color
		vector.FillRect(img, 0, 0, config.TileSize, config.TileSize, rockColor, false)
		// 2. Stroke boundary
		vector.StrokeRect(img, 0, 0, config.TileSize, config.TileSize, 0.5, strokeColor, false)

		// Create a local RNG for this variant's generation
		rng := rand.New(rand.NewSource(int64(idx * 997)))

		// 3. Draw darker sand grains
		for range 6 {
			px := float32(rng.Intn(config.TileSize-4)) + 2
			py := float32(rng.Intn(config.TileSize-4)) + 2
			vector.FillRect(img, px, py, 2, 2, darkSandColor, false)
		}

		// 4. Draw lighter sand grains
		for range 6 {
			px := float32(rng.Intn(config.TileSize-4)) + 2
			py := float32(rng.Intn(config.TileSize-4)) + 2
			vector.FillRect(img, px, py, 2, 2, lightSandColor, false)
		}

		c.tileImages[idx] = img
	}
}

func hashCoords(tx, ty int) uint64 {
	// Injective combination of 32-bit coordinates into 64-bit int
	x := (int64(tx) << 32) | (int64(uint32(ty)))
	// SplitMix64 finalizer
	u := uint64(x)
	u ^= u >> 33
	u *= 0xff51afd7ed558ccd
	u ^= u >> 33
	u *= 0xc4ceb9fe1a85ec53
	u ^= u >> 33
	return u
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
	for sy := float32(0); sy < float32(config.ScreenHeight); sy += stripH {
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
		vector.FillRect(screen, 0, sy, float32(config.ScreenWidth), stripH, sc, false)
	}
}

func (c *ShallowSeabedCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	op := ebiten.DrawImageOptions{}
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float64(tx*config.TileSize - int(camX))
				sy := float64(ty*config.TileSize - int(camY))

				op.GeoM.Reset()
				op.GeoM.Translate(sx, sy)

				// Injective stateless hash to choose the variant
				h := hashCoords(tx, ty)
				variantIdx := h % uint64(len(c.tileImages))

				screen.DrawImage(c.tileImages[variantIdx], &op)
			}
		}
	}
}

func (c *ShallowSeabedCave) GenerateEntities(seed int64) []entity.CaveEntity {
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
			isOpenWater := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
			if isOpenWater && r.Float64() < 0.012 {
				entities = append(entities, entity.NewPassiveFish(
					float64(tx*config.TileSize)+float64(config.TileSize-20)/2.0,
					float64(ty*config.TileSize)+float64(config.TileSize-12)/2.0,
					r.Float64() < 0.5,
					r.Float64()*math.Pi*2,
				))
			}
			if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < 0.015 {
				entities = append(entities, &entity.PassiveCrab{
					BaseEntity: entity.BaseEntity{
						Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-16)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-10)},
						Dimensions: gvec.Vec2{X: 16, Y: 10},
						Active:     true,
					},
					FacingRight: r.Float64() < 0.5,
				})
			}
			if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < 0.28 {
				height := 32.0 + r.Float64()*48.0
				if r.Float64() < 0.20 {
					entities = append(entities, entity.NewShockKelp(
						float64(tx*config.TileSize)+float64(config.TileSize-16)/2.0,
						float64(ty*config.TileSize)+float64(config.TileSize)-height,
						height,
						"floor",
					))
				} else {
					entities = append(entities, &entity.Kelp{
						BaseEntity: entity.BaseEntity{
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-16)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize) - height},
							Dimensions: gvec.Vec2{X: 16, Y: height},
							Active:     true,
						},
						SwayPhase: r.Float64() * math.Pi * 2,
					})
				}
			}
		}
	}

	return entities
}

func (c *ShallowSeabedCave) GenerateResources(seed int64) []resource.Resource {
	// Standard depth generation delegate (depth tier 0)
	return resource.GenerateResourceNodes(c.Grid, seed)
}

func (c *ShallowSeabedCave) GetAmbientColor(lightMult float64) []float32 {
	alpha := float32(0.75 - (lightMult-0.2)/0.8*0.60)
	return []float32{0.04, 0.06, 0.12, alpha}
}

func GenerateShallowSeabedGrid(r *rand.Rand, distToLand float64, hasLeftWater, hasRightWater bool) [][]bool {
	const (
		w = CaveWidth
		h = CaveHeight
	)
	floorY := min(max(6+int(distToLand*2.2), 6), 60)

	freq1 := 0.15 + r.Float64()*0.2
	freq2 := 0.05 + r.Float64()*0.1
	amp1 := 2.0 + r.Float64()*4.0
	amp2 := 1.0 + r.Float64()*3.0

	caveGrid := make([][]bool, w)
	for x := range w {
		caveGrid[x] = make([]bool, h)
		colFloorY := max(floorY+int(math.Sin(float64(x)*freq1)*amp1+math.Cos(float64(x)*freq2)*amp2), 6)

		// Apply slope to the left edge if the neighbor is not water
		if !hasLeftWater && x < 15 {
			t := float64(x) / 15.0
			t = math.Sin(t * math.Pi / 2.0)
			blendY := 4.0 + (float64(colFloorY)-4.0)*t
			colFloorY = int(blendY)
		}

		// Apply slope to the right edge if the neighbor is not water
		if !hasRightWater && x >= w-15 {
			t := float64(w-1-x) / 15.0
			t = math.Sin(t * math.Pi / 2.0)
			blendY := 4.0 + (float64(colFloorY)-4.0)*t
			colFloorY = int(blendY)
		}

		for y := range h {
			isLeftBorderSolid := !hasLeftWater && x == 0
			isRightBorderSolid := !hasRightWater && x == w-1
			if isLeftBorderSolid || isRightBorderSolid || y >= colFloorY {
				caveGrid[x][y] = true
			} else {
				caveGrid[x][y] = false
			}
		}
	}
	return caveGrid
}
