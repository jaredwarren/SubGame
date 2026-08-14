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
	darkColor := c.Biome.CaveSandDarkColor
	lightColor := c.Biome.CaveSandLightColor
	floorStyle := c.Biome.FloorStyle

	for idx := range c.tileImages {
		img := ebiten.NewImage(config.TileSize, config.TileSize)
		// 1. Fill base rock color
		vector.FillRect(img, 0, 0, config.TileSize, config.TileSize, rockColor, false)
		// 2. Stroke boundary
		vector.StrokeRect(img, 0, 0, config.TileSize, config.TileSize, 0.5, strokeColor, false)

		// Create a local RNG for this variant's generation
		rng := rand.New(rand.NewSource(int64(idx*997 + 17)))

		switch floorStyle {
		case FloorStyleMoss:
			// Moss carpet rendering (Kelp Forest)
			// 1. Irregular rounded moss clumps / patches
			numPatches := rng.Intn(3) + 3
			for i := 0; i < numPatches; i++ {
				mx := float32(rng.Intn(config.TileSize-18)) + 9
				my := float32(rng.Intn(config.TileSize-18)) + 9
				rad := float32(rng.Float64()*6.0 + 4.5)
				vector.FillCircle(img, mx, my, rad, darkColor, false)
				vector.FillCircle(img, mx+float32(rng.Float64()*2.4-1.2), my+float32(rng.Float64()*2.4-1.2), rad*0.65, lightColor, false)
			}

			// 2. Fine moss fibrils / blades
			numTufts := rng.Intn(8) + 10
			for i := 0; i < numTufts; i++ {
				tx := float32(rng.Intn(config.TileSize-6)) + 3
				ty := float32(rng.Intn(config.TileSize-6)) + 3
				vector.FillRect(img, tx, ty, 2, 2, lightColor, false)
				if rng.Float64() < 0.45 {
					vector.StrokeLine(img, tx, ty, tx+float32(rng.Float64()*2-1), ty-float32(rng.Float64()*3+1), 1.0, lightColor, false)
				}
			}

			// 3. Subtle pale chartreuse lichen specks
			lichenColor := color.RGBA{160, 220, 115, 220}
			for i := 0; i < 4; i++ {
				lx := float32(rng.Intn(config.TileSize-8)) + 4
				ly := float32(rng.Intn(config.TileSize-8)) + 4
				vector.FillCircle(img, lx, ly, 1.2, lichenColor, false)
			}

		case FloorStyleBasalt:
			// Basalt rock rendering (Thermal Barrens)
			// 1. Fractured cooling plates & geometric cracks
			numCracks := rng.Intn(3) + 2
			for i := 0; i < numCracks; i++ {
				cx1 := float32(rng.Intn(config.TileSize-12)) + 6
				cy1 := float32(rng.Intn(config.TileSize-12)) + 6
				cx2 := cx1 + float32(rng.Float64()*24.0-12.0)
				cy2 := cy1 + float32(rng.Float64()*24.0-12.0)
				vector.StrokeLine(img, cx1, cy1, cx2, cy2, 1.5, darkColor, false)

				if rng.Float64() < 0.6 {
					midX := (cx1 + cx2) * 0.5
					midY := (cy1 + cy2) * 0.5
					cx3 := midX + float32(rng.Float64()*12.0-6.0)
					cy3 := midY + float32(rng.Float64()*12.0-6.0)
					vector.StrokeLine(img, midX, midY, cx3, cy3, 1.0, darkColor, false)
				}
			}

			// 2. Basalt rough stippling and ash grains
			for range 10 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 2, 2, lightColor, false)
			}
			for range 8 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 2, 2, darkColor, false)
			}

			// 3. Subtle thermal mineral / ember flecks
			if rng.Float64() < 0.55 {
				ex := float32(rng.Intn(config.TileSize-10)) + 5
				ey := float32(rng.Intn(config.TileSize-10)) + 5
				emberColor := color.RGBA{235, 95, 30, 220}
				vector.FillCircle(img, ex, ey, 1.5, emberColor, false)
				vector.FillCircle(img, ex, ey, 0.7, color.RGBA{255, 200, 60, 255}, false)
			}

		case FloorStyleAbyssalSilt:
			// Deep abyssal silt & sediment (Abyssal Trench)
			// 1. Soft horizontal silt drift lines
			numDrifts := rng.Intn(2) + 2
			for i := 0; i < numDrifts; i++ {
				dx1 := float32(rng.Intn(config.TileSize-16)) + 4
				dy := float32(rng.Intn(config.TileSize-12)) + 6
				dx2 := dx1 + float32(rng.Float64()*20.0+10.0)
				vector.StrokeLine(img, dx1, dy, dx2, dy+float32(rng.Float64()*3-1.5), 1.2, darkColor, false)
			}

			// 2. Fine sediment grains
			for range 8 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 2, 2, darkColor, false)
			}
			for range 6 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 2, 2, lightColor, false)
			}

			// 3. Subtle bioluminescent/crystal mineral fleck
			if rng.Float64() < 0.6 {
				cx := float32(rng.Intn(config.TileSize-10)) + 5
				cy := float32(rng.Intn(config.TileSize-10)) + 5
				shimmerColor := color.RGBA{95, 180, 245, 200}
				vector.FillCircle(img, cx, cy, 1.2, shimmerColor, false)
			}

		default: // FloorStyleCoralSand (Shallow Reef)
			// 1. Sand ripple lines
			numRipples := rng.Intn(2) + 1
			for i := 0; i < numRipples; i++ {
				rx1 := float32(rng.Intn(config.TileSize-16)) + 4
				ry := float32(rng.Intn(config.TileSize-12)) + 6
				rx2 := rx1 + float32(rng.Float64()*16.0+8.0)
				vector.StrokeLine(img, rx1, ry, rx2, ry+1, 1.0, darkColor, false)
			}

			// 2. Darker sand grains
			for range 8 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 2, 2, darkColor, false)
			}

			// 3. Lighter sand grains
			for range 8 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 2, 2, lightColor, false)
			}

			// 4. Occasional tiny pebble or coral grain
			if rng.Float64() < 0.35 {
				px := float32(rng.Intn(config.TileSize-8)) + 4
				py := float32(rng.Intn(config.TileSize-8)) + 4
				vector.FillCircle(img, px, py, 1.5, color.RGBA{235, 130, 95, 220}, false)
			}
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
