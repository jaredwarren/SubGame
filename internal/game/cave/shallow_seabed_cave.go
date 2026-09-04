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
	Grid            [][]bool
	Biome           *CaveBiomeSpec
	tileImages      []*ebiten.Image
	HasChasm        bool
	ChasmX          int
	ChasmWidth      int
	BasinFloorY     int
	ChasmBottomY    int
	ChasmTarget     CaveType
	chasmTargetSet  bool
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

// NewChasmShallowCave creates a shallow seabed cave with a subterranean chasm targeting a deep cave type.
func NewChasmShallowCave(grid [][]bool, spec *CaveBiomeSpec, target CaveType) *ShallowSeabedCave {
	c := NewShallowSeabedCaveWithBiome(grid, spec)
	c.HasChasm = true
	c.ChasmTarget = target
	c.chasmTargetSet = true
	c.detectChasm()
	return c
}

// NewShockKelpShallowCave creates a shallow Kelp Forest seabed cave with a subterranean shock chasm.
func NewShockKelpShallowCave(grid [][]bool) *ShallowSeabedCave {
	return NewChasmShallowCave(grid, KelpForestBiome, CaveShockKelp)
}

// NewTrenchShallowCave creates a shallow Abyssal seabed cave with a subterranean trench fissure.
func NewTrenchShallowCave(grid [][]bool) *ShallowSeabedCave {
	return NewChasmShallowCave(grid, AbyssalBlueBiome, CaveOrganicTrench)
}

func (c *ShallowSeabedCave) detectChasm() {
	if len(c.Grid) == 0 || !c.HasChasm {
		return
	}
	gridW := len(c.Grid)
	gridH := len(c.Grid[0])

	// 1. Find the seabed surface floor level at the depression basin
	basinFloorY := 0
	for x := 2; x < gridW-2; x++ {
		for y := 0; y < gridH; y++ {
			if c.Grid[x][y] {
				if y > basinFloorY {
					basinFloorY = y
				}
				break
			}
		}
	}
	if basinFloorY < 8 {
		basinFloorY = 16
	}
	c.BasinFloorY = basinFloorY

	// 2. Position trigger 14 tiles down into the crevice shaft
	triggerTileY := min(basinFloorY+14, gridH-4)
	startX := -1
	endX := -1
	for x := 2; x < gridW-2; x++ {
		if !c.Grid[x][triggerTileY] {
			if startX == -1 {
				startX = x
			}
			endX = x
		}
	}

	if startX != -1 && endX >= startX {
		c.ChasmX = startX
		c.ChasmWidth = endX - startX + 1
		c.ChasmBottomY = triggerTileY
		if !c.chasmTargetSet {
			c.ChasmTarget = CaveShockKelp
			c.chasmTargetSet = true
		}
		return
	}

	c.HasChasm = false
}

func (c *ShallowSeabedCave) HasFloorChasm() bool {
	return c.HasChasm
}

func (c *ShallowSeabedCave) GetChasmBounds() (minX, maxX, triggerY float64) {
	minX = float64(c.ChasmX * config.TileSize)
	maxX = float64((c.ChasmX + c.ChasmWidth) * config.TileSize)
	triggerY = float64(c.ChasmBottomY * config.TileSize)
	return minX, maxX, triggerY
}

func (c *ShallowSeabedCave) GetChasmTarget() CaveType {
	return c.ChasmTarget
}

func (c *ShallowSeabedCave) chasmRim() *ChasmRimSpec {
	return ChasmRimSpecFor(c.ChasmTarget)
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
			// 1. Fluid volcanic silt drift bands (dark obsidian shadows)
			numDrifts := rng.Intn(2) + 2
			for i := 0; i < numDrifts; i++ {
				dx1 := float32(rng.Intn(config.TileSize-18)) + 3
				dy := float32(rng.Intn(config.TileSize-12)) + 6
				dx2 := dx1 + float32(rng.Float64()*22.0+10.0)
				vector.StrokeLine(img, dx1, dy, dx2, dy+float32(rng.Float64()*3-1.5), 1.6, darkColor, false)
			}

			// 2. Bioluminescent sediment ripple veins (glowing cyan energy hairline fractures)
			if rng.Float64() < 0.75 {
				vx1 := float32(rng.Intn(config.TileSize-16)) + 4
				vy := float32(rng.Intn(config.TileSize-14)) + 7
				vx2 := vx1 + float32(rng.Float64()*18.0+8.0)
				veinColor := color.RGBA{45, 195, 235, 175}
				vector.StrokeLine(img, vx1, vy, vx2, vy+float32(rng.Float64()*2-1.0), 0.9, veinColor, false)
			}

			// 3. Fine volcanic sediment and obsidian granules
			for range 10 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 2, 2, darkColor, false)
			}
			for range 6 {
				px := float32(rng.Intn(config.TileSize-4)) + 2
				py := float32(rng.Intn(config.TileSize-4)) + 2
				vector.FillRect(img, px, py, 1.5, 1.5, lightColor, false)
			}

			// 4. Luminous crystalline mineral flecks with radiant halo
			numFlecks := rng.Intn(2) + 1
			for i := 0; i < numFlecks; i++ {
				cx := float32(rng.Intn(config.TileSize-10)) + 5
				cy := float32(rng.Intn(config.TileSize-10)) + 5
				haloClr := color.RGBA{40, 160, 240, 110}
				sparkleClr := color.RGBA{130, 230, 255, 240}
				vector.FillCircle(img, cx, cy, 2.0, haloClr, false)
				vector.FillCircle(img, cx, cy, 1.0, sparkleClr, false)
				vector.FillCircle(img, cx, cy, 0.4, color.White, false)
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

// fracHash returns a deterministic float in [0, 1) from a seed — used for
// background particles so DrawBackground does not allocate RNG objects.
func fracHash(seed uint64) float64 {
	u := seed
	u ^= u >> 33
	u *= 0xff51afd7ed558ccd
	u ^= u >> 33
	u *= 0xc4ceb9fe1a85ec53
	u ^= u >> 33
	return float64(u>>11) / float64(1<<53)
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
			if tx < 0 || tx >= len(c.Grid) || ty < 0 || ty >= len(c.Grid[0]) || !c.Grid[tx][ty] {
				continue
			}

			sx := float64(tx*config.TileSize - int(camX))
			sy := float64(ty*config.TileSize - int(camY))

			h := hashCoords(tx, ty)

			// If this cave has a chasm, blend deep rock tiles near the shaft/fissure
			isChasmTile := false
			if c.HasChasm {
				distFromChasm := 999
				if tx < c.ChasmX {
					distFromChasm = c.ChasmX - tx
				} else if tx >= c.ChasmX+c.ChasmWidth {
					distFromChasm = tx - (c.ChasmX + c.ChasmWidth - 1)
				} else {
					distFromChasm = 0
				}

				startBlendY := max(c.BasinFloorY-4, 8)
				if distFromChasm <= 4 && ty >= startBlendY {
					depthSpan := float64(c.ChasmBottomY - startBlendY)
					if depthSpan < 6.0 {
						depthSpan = 6.0
					}
					depthFrac := float64(ty-startBlendY) / depthSpan
					if depthFrac > 1.0 {
						depthFrac = 1.0
					}
					proxFrac := 1.0 - float64(distFromChasm)/5.0
					blendChance := depthFrac * proxFrac * 0.95
					if float64(h%100)/100.0 < blendChance {
						isChasmTile = true
					}
				}
			}

			if isChasmTile {
				rim := c.chasmRim()
				rockColor := rim.RockColor
				strokeColor := rim.StrokeColor

				vector.FillRect(screen, float32(sx), float32(sy), config.TileSize, config.TileSize, rockColor, false)
				vector.StrokeRect(screen, float32(sx), float32(sy), config.TileSize, config.TileSize, 0.5, strokeColor, false)

				// Deterministic cracks/veins
				if h%3 == 0 {
					veinClr := rim.primaryVeinColor(h)
					vector.StrokeLine(screen, float32(sx)+6, float32(sy)+6, float32(sx)+16, float32(sy)+16, 1.5, veinClr, false)
					vector.StrokeLine(screen, float32(sx)+16, float32(sy)+16, float32(sx)+12, float32(sy)+24, 1.2, veinClr, false)
				}
				if h%5 == 0 {
					vector.StrokeLine(screen, float32(sx)+float32(config.TileSize)-8, float32(sy)+8, float32(sx)+float32(config.TileSize)-16, float32(sy)+20, 1.2, rim.VeinSecondary, false)
				}
			} else {
				op.GeoM.Reset()
				op.GeoM.Translate(sx, sy)
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

			hasFloor := ty < gridH-2 && grid[tx][ty+1]
			if hasFloor && r.Float64() < rules.ShatterBulbChance {
				height := 42.0 + r.Float64()*16.0
				entities = append(entities, entity.NewShatterBulb(
					float64(tx*config.TileSize)+float64(config.TileSize-24)/2.0,
					float64(ty*config.TileSize)+float64(config.TileSize)-height,
					height,
				))
			} else if !hasFloor && ty > 1 && ty < gridH-2 && !grid[tx][ty-1] && (grid[tx-1][ty] || grid[tx+1][ty]) && r.Float64() < rules.ShatterBulbChance*0.6 {
				height := 42.0 + r.Float64()*16.0
				anchor := "left"
				if grid[tx-1][ty] && grid[tx+1][ty] {
					if r.Float64() < 0.5 {
						anchor = "right"
					}
				} else if grid[tx+1][ty] {
					anchor = "right"
				}
				if ent := SpawnFloraAnchored(FloraShatterBulb, tx, ty, height, anchor, r); ent != nil {
					entities = append(entities, ent)
				}
			}
			isOpenWater := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
			if isOpenWater {
				roll := r.Float64()
				if roll < rules.OpenWaterFishChance {
					if c.Biome != nil && c.Biome.ID == "abyssal_blue" {
						entities = append(entities, entity.NewLanternfish(
							float64(tx*config.TileSize)+float64(config.TileSize-18)/2.0,
							float64(ty*config.TileSize)+float64(config.TileSize-12)/2.0,
							r.Float64() < 0.5,
							r.Float64()*math.Pi*2,
						))
					} else {
						entities = append(entities, entity.NewPassiveFish(
							float64(tx*config.TileSize)+float64(config.TileSize-20)/2.0,
							float64(ty*config.TileSize)+float64(config.TileSize-12)/2.0,
							r.Float64() < 0.5,
							r.Float64()*math.Pi*2,
						))
					}
				} else if roll < rules.OpenWaterFishChance+0.006 {
					entities = append(entities, entity.NewInkSquid(
						float64(tx*config.TileSize)+float64(config.TileSize-22)/2.0,
						float64(ty*config.TileSize)+float64(config.TileSize-16)/2.0,
						r.Float64() < 0.5,
					))
				}
			}
			if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < rules.FaunaChance {
				faunaType := FaunaPassiveFish
				if c.Biome != nil && len(c.Biome.FaunaSpawns) > 0 {
					faunaType = SelectWeightedEntry(c.Biome.FaunaSpawns, r.Float64())
				}
				if ent := SpawnFauna(faunaType, tx, ty, grid, r); ent != nil {
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
			} else if ty > 1 && ty < gridH-2 && !grid[tx][ty+1] && !grid[tx][ty-1] && (grid[tx-1][ty] || grid[tx+1][ty]) && r.Float64() < rules.FloraChance {
				if c.Biome != nil && len(c.Biome.FloraSpawns) > 0 {
					floraType := SelectWeightedEntry(c.Biome.FloraSpawns, r.Float64())
					if floraType == FloraShatterBulb || floraType == FloraShockKelp {
						height := 32.0 + r.Float64()*48.0
						anchor := "left"
						if grid[tx-1][ty] && grid[tx+1][ty] {
							if r.Float64() < 0.5 {
								anchor = "right"
							}
						} else if grid[tx+1][ty] {
							anchor = "right"
						}
						if ent := SpawnFloraAnchored(floraType, tx, ty, height, anchor, r); ent != nil {
							entities = append(entities, ent)
						}
					}
				}
			}

			// Spawn decorative corals near any solid face
			entities = MaybeSpawnCoral(entities, grid, tx, ty, rules.CoralChance, entity.CoralBiomeShallow, entity.CoralShallowVariantCount, r)
		}
	}

	// Ensure at least 1 InkSquid is present in every shallow seabed cave
	squidCount := 0
	for _, ent := range entities {
		if _, ok := ent.(*entity.InkSquid); ok {
			squidCount++
		}
	}
	for squidCount < 1 {
		found := false
		for attempts := 0; attempts < 100; attempts++ {
			tx := 2 + r.Intn(gridW-4)
			ty := 2 + r.Intn(gridH-4)
			if !grid[tx][ty] && !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1] {
				entities = append(entities, entity.NewInkSquid(
					float64(tx*config.TileSize)+float64(config.TileSize-22)/2.0,
					float64(ty*config.TileSize)+float64(config.TileSize-16)/2.0,
					r.Float64() < 0.5,
				))
				squidCount++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// Spawn special flora and entities along chasm rims and winding crevice walls
	if c.HasChasm && c.ChasmX > 0 {
		entities = append(entities, spawnChasmRimEntities(c.chasmRim(), grid, c.ChasmX, c.ChasmWidth, r)...)
	}

	return entities
}

func (c *ShallowSeabedCave) GenerateResources(seed int64) []resource.Resource {
	var nodes []resource.Resource
	if c.Biome != nil && len(c.Biome.MineralSpawns) > 0 {
		spawns := make([]resource.ResourceSpawnEntry, len(c.Biome.MineralSpawns))
		for i, s := range c.Biome.MineralSpawns {
			spawns[i] = resource.ResourceSpawnEntry{Type: s.Type, Weight: s.Weight}
		}
		nodes = resource.GenerateResourceNodesWithBiome(c.Grid, seed, spawns)
	} else {
		nodes = resource.GenerateResourceNodes(c.Grid, seed)
	}

	// Guarantee at least 1 Titanium node in ShallowReefBiome (and default shallow seabed caves)
	if c.Biome == nil || c.Biome.ID == "shallow_reef" {
		hasTitanium := false
		for _, n := range nodes {
			if rn, ok := n.(*resource.ResourceNode); ok && rn.Type == resource.NodeTitanium {
				hasTitanium = true
				break
			}
		}

		if !hasTitanium {
			if len(nodes) > 0 {
				if rn, ok := nodes[0].(*resource.ResourceNode); ok {
					rn.Type = resource.NodeTitanium
					hasTitanium = true
				}
			}

			if !hasTitanium && c.Grid != nil {
				r := rand.New(rand.NewSource(seed))
				gridW := len(c.Grid)
				gridH := len(c.Grid[0])
				type candidatePos struct {
					tx, ty int
					dirs   []resource.AttachDirection
				}
				var candidates []candidatePos
				for tx := 1; tx < gridW-1; tx++ {
					for ty := 1; ty < gridH-1; ty++ {
						if !c.Grid[tx][ty] {
							var dirs []resource.AttachDirection
							if c.Grid[tx][ty-1] {
								dirs = append(dirs, resource.AttachTop)
							}
							if c.Grid[tx][ty+1] {
								dirs = append(dirs, resource.AttachBottom)
							}
							if c.Grid[tx-1][ty] {
								dirs = append(dirs, resource.AttachLeft)
							}
							if c.Grid[tx+1][ty] {
								dirs = append(dirs, resource.AttachRight)
							}
							if len(dirs) > 0 {
								candidates = append(candidates, candidatePos{tx, ty, dirs})
							}
						}
					}
				}
				if len(candidates) > 0 {
					pick := candidates[r.Intn(len(candidates))]
					node := resource.NewNode(resource.NodeTitanium, pick.tx, pick.ty)
					node.SetAttachDir(pick.dirs[r.Intn(len(pick.dirs))])
					node.SetHitsToMine(resource.GenConfig.BaseHitsToMine)
					nodes = append(nodes, node)
				}
			}
		}
	}

	return nodes
}


func (c *ShallowSeabedCave) GetAmbientColor(lightMult float64) [4]float32 {
	alpha := float32(0.75 - (lightMult-0.2)/0.8*0.60)
	if c.HasChasm {
		rgb := c.chasmRim().AmbientRGB
		if c.chasmRim().Target == CaveOrganicTrench {
			alpha = max(0.85, alpha)
		}
		return [4]float32{rgb[0], rgb[1], rgb[2], alpha}
	}
	return [4]float32{0.04, 0.06, 0.12, alpha}
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

// GenerateChasmShallowGrid generates a shallow seabed grid with a funneled basin
// and an organic winding crevice / fissure descending into a subterranean deep cave.
func GenerateChasmShallowGrid(r *rand.Rand, distToLand float64, hasLeftWater, hasRightWater bool, funnelRadius, initialWidth, minWidth float64) [][]bool {
	const (
		w = CaveWidth  // 60
		h = CaveHeight // 120
	)
	baseFloorY := min(max(6+int(distToLand*2.2), 6), 60)

	freq1 := 0.15 + r.Float64()*0.2
	freq2 := 0.05 + r.Float64()*0.1
	amp1 := 2.0 + r.Float64()*4.0
	amp2 := 1.0 + r.Float64()*3.0

	// Center of the funneled basin & chasm (between x=24 and x=34)
	chasmCenter := 26 + r.Intn(7) // e.g. 26..32

	caveGrid := make([][]bool, w)
	colFloorHeights := make([]int, w)

	for x := range w {
		caveGrid[x] = make([]bool, h)
		colFloorY := max(baseFloorY+int(math.Sin(float64(x)*freq1)*amp1+math.Cos(float64(x)*freq2)*amp2), 6)

		// 1. Funnel effect: smoothly depress seabed floor as we get closer to chasmCenter
		distFromCenter := math.Abs(float64(x - chasmCenter))
		if distFromCenter < funnelRadius {
			t := (funnelRadius - distFromCenter) / funnelRadius
			dip := math.Sin(t * math.Pi / 2.0) * 8.0
			colFloorY += int(dip)
		}

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

		colFloorHeights[x] = colFloorY

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

	// 2. Carve the organic winding crevice down through the rock
	basinFloorY := colFloorHeights[chasmCenter]
	meanderFreq := 0.20 + r.Float64()*0.15
	meanderPhase := r.Float64() * math.Pi * 2.0

	for y := basinFloorY - 1; y < h; y++ {
		depthProgress := float64(y - basinFloorY)
		// Smooth horizontal meandering (-2.0 to +2.0 tiles)
		meanderOffset := math.Sin(depthProgress*meanderFreq+meanderPhase) * 2.0
		roughness := (r.Float64() - 0.5) * 0.8
		cx := float64(chasmCenter) + meanderOffset + roughness

		// Crevice width: smoothly tapering from initialWidth to minWidth
		creviceWidth := initialWidth - math.Min(depthProgress/30.0, 1.0)*(initialWidth-minWidth) + (r.Float64()-0.5)*0.5
		if creviceWidth < minWidth {
			creviceWidth = minWidth
		}

		minX := int(math.Floor(cx - creviceWidth/2.0))
		maxX := int(math.Ceil(cx + creviceWidth/2.0))

		for x := minX; x <= maxX; x++ {
			if x > 1 && x < w-2 && y >= 0 && y < h {
				caveGrid[x][y] = false
			}
		}
	}

	return caveGrid
}

// GenerateShockKelpShallowGrid generates a shallow seabed grid with a funneled kelp basin
// and an organic winding crevice (4.2 to 5.2 tiles wide) descending toward the subterranean Shock Kelp Cave.
func GenerateShockKelpShallowGrid(r *rand.Rand, distToLand float64, hasLeftWater, hasRightWater bool) [][]bool {
	return GenerateChasmShallowGrid(r, distToLand, hasLeftWater, hasRightWater, 11.0, 5.2, 4.2)
}

// GenerateTrenchShallowGrid generates a shallow seabed grid with a funneled abyssal basin
// and an organic winding fissure (4.2 to 5.6 tiles wide) descending toward the subterranean Organic Trench Cave.
func GenerateTrenchShallowGrid(r *rand.Rand, distToLand float64, hasLeftWater, hasRightWater bool) [][]bool {
	return GenerateChasmShallowGrid(r, distToLand, hasLeftWater, hasRightWater, 12.0, 5.6, 4.2)
}

