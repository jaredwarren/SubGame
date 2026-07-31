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
)

type ShallowSeabedCave struct {
	Grid       [][]bool
	Biome      *CaveBiomeSpec
	tileImages []*ebiten.Image
}

func NewShallowSeabedCave(grid [][]bool) *ShallowSeabedCave {
	return NewShallowSeabedCaveWithBiome(grid, nil)
}

func NewShallowSeabedCaveWithBiome(grid [][]bool, spec *CaveBiomeSpec) *ShallowSeabedCave {
	if spec == nil {
		spec = DefaultShallowReefBiome
	}
	c := &ShallowSeabedCave{
		Grid:       grid,
		Biome:      spec,
		tileImages: make([]*ebiten.Image, 8),
	}
	c.preRenderTiles()
	return c
}

func (c *ShallowSeabedCave) preRenderTiles() {
	rockColor := c.Biome.CaveRockColor
	strokeColor := c.Biome.CaveStrokeColor
	darkSandColor := c.Biome.CaveSandDarkColor
	lightSandColor := c.Biome.CaveSandLightColor

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
	// Surface base color tinted by Biome Ambient Tint
	tintR := float64(10)
	tintG := float64(40)
	tintB := float64(100)
	if c.Biome != nil {
		tintR = float64(c.Biome.CaveAmbientTint.R) * 0.5
		tintG = float64(c.Biome.CaveAmbientTint.G) * 0.5
		tintB = float64(c.Biome.CaveAmbientTint.B) * 0.5
	}

	baseR := tintR + float64(30)*lightMult
	baseG := tintG + float64(80)*lightMult
	baseB := tintB + float64(80)*lightMult

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
			R: uint8(max(0, min(255, baseR*darkFactor))),
			G: uint8(max(0, min(255, baseG*darkFactor))),
			B: uint8(max(0, min(255, baseB*darkFactor))),
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
	rules := c.Biome.SpawnRulesOrDefault()

	gridW := len(grid)
	gridH := len(grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			if grid[tx][ty] {
				continue
			}

			hasAdjacentWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
			if hasAdjacentWall && r.Float64() < rules.ShatterBulbChance {
				entities = append(entities, entity.NewShatterBulb(
					float64(tx*config.TileSize)+float64(config.TileSize-24)/2.0,
					float64(ty*config.TileSize)+float64(config.TileSize-24)/2.0,
				))
			}
			isOpenWater := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
			if isOpenWater && r.Float64() < rules.OpenWaterFishChance {
				entities = append(entities, entity.NewPassiveFish(
					float64(tx*config.TileSize)+float64(config.TileSize-20)/2.0,
					float64(ty*config.TileSize)+float64(config.TileSize-12)/2.0,
					r.Float64() < 0.5,
					r.Float64()*math.Pi*2,
				))
			}
			if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < rules.FaunaChance {
				faunaType := FaunaPassiveFish
				if c.Biome != nil && len(c.Biome.FaunaSpawns) > 0 {
					faunaType = SelectWeightedEntry(c.Biome.FaunaSpawns, r.Float64())
				}
				if ent := SpawnFauna(faunaType, tx, ty, r); ent != nil {
					entities = append(entities, ent)
				}
			}
			if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < rules.FloraChance {
				height := 32.0 + r.Float64()*48.0
				floraType := FloraKelp
				if c.Biome != nil && len(c.Biome.FloraSpawns) > 0 {
					floraType = SelectWeightedEntry(c.Biome.FloraSpawns, r.Float64())
				}
				if ent := SpawnFlora(floraType, tx, ty, height, r); ent != nil {
					entities = append(entities, ent)
				}
			}

			// Spawn decorative corals near any solid face
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

			if len(coralAttachments) > 0 && r.Float64() < rules.CoralChance {
				attach := coralAttachments[r.Intn(len(coralAttachments))]
				variant := r.Intn(3) // 3 variants for shallow seabed
				
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
				
				entities = append(entities, entity.NewCoral(cx, cy, entity.CoralBiomeShallow, attach, variant, r))
			}
		}
	}

	return entities
}

func (c *ShallowSeabedCave) GenerateResources(seed int64) []resource.Resource {
	if c.Biome != nil && len(c.Biome.MineralSpawns) > 0 {
		spawns := make([]resource.ResourceSpawnEntry, len(c.Biome.MineralSpawns))
		for i, s := range c.Biome.MineralSpawns {
			spawns[i] = resource.ResourceSpawnEntry{Type: s.Type, Weight: s.Weight}
		}
		return resource.GenerateResourceNodesWithBiome(c.Grid, seed, spawns)
	}
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
