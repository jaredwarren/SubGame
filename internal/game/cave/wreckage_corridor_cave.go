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

type WreckageCorridorCave struct {
	Grid      [][]bool
	ShipIndex int
}

func NewWreckageCorridorCave(grid [][]bool, shipIndex int) *WreckageCorridorCave {
	return &WreckageCorridorCave{Grid: grid, ShipIndex: shipIndex}
}

func (c *WreckageCorridorCave) GetCaveType() CaveType { return CaveWreckage }
func (c *WreckageCorridorCave) GetGrid() [][]bool     { return c.Grid }

func (c *WreckageCorridorCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Dark, artificial metallic vessel interior background
	screen.Fill(color.RGBA{14, 15, 20, 255})

	// Optional: render faint background steel pillars or grid lines for industrial feel
	const lineGap = 40.0
	offsetX := float32(math.Mod(camY*0.1, lineGap))
	for x := float32(0); x < float32(config.ScreenWidth); x += lineGap {
		vector.StrokeLine(screen, x, 0, x, float32(config.ScreenHeight), 0.8, color.RGBA{20, 24, 30, 255}, false)
	}
	for y := float32(0); y < float32(config.ScreenHeight); y += lineGap {
		sy := y - offsetX
		vector.StrokeLine(screen, 0, sy, float32(config.ScreenWidth), sy, 0.8, color.RGBA{20, 24, 30, 255}, false)
	}
}

func (c *WreckageCorridorCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float32(tx*config.TileSize - int(camX))
				sy := float32(ty*config.TileSize - int(camY))

				// Steel gray bulkhead panels
				rockColor := color.RGBA{52, 58, 68, 255}
				strokeColor := color.RGBA{85, 80, 75, 255}

				vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, rockColor, false)
				vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 1.2, strokeColor, false)

				// Draw diagonal yellow-and-black hazard warning lines along bulkheads that border corridors
				hasBorder := false
				if tx > 0 && !c.Grid[tx-1][ty] {
					hasBorder = true
				}
				if tx < len(c.Grid)-1 && !c.Grid[tx+1][ty] {
					hasBorder = true
				}
				if ty > 0 && !c.Grid[tx][ty-1] {
					hasBorder = true
				}
				if ty < len(c.Grid[0])-1 && !c.Grid[tx][ty+1] {
					hasBorder = true
				}

				if hasBorder {
					stripeColor := color.RGBA{215, 175, 30, 160} // Rusted safety yellow
					// Draw diagonal warning lines across the bulkhead
					vector.StrokeLine(screen, sx, sy+8, sx+8, sy, 2.0, stripeColor, false)
					vector.StrokeLine(screen, sx, sy+24, sx+24, sy, 2.0, stripeColor, false)
					vector.StrokeLine(screen, sx+16, sy+config.TileSize, sx+config.TileSize, sy+16, 2.0, stripeColor, false)
					vector.StrokeLine(screen, sx+32, sy+config.TileSize, sx+config.TileSize, sy+32, 2.0, stripeColor, false)
				}
			}
		}
	}
}

func (c *WreckageCorridorCave) GenerateEntities(seed int64) []entity.CaveEntity {
	r := rand.New(rand.NewSource(seed))
	var entities []entity.CaveEntity

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
				entities = append(entities, &entity.ShatterBulb{
					BaseEntity: entity.BaseEntity{
						Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
						Dimensions: gvec.Vec2{X: 24, Y: 24},
						Active:     true,
					},
				})
			}

			// Spawn decorative wreckage barnacles/growths (8% chance near any solid face)
			var coralAttachments []string
			if c.Grid[tx][ty+1] {
				coralAttachments = append(coralAttachments, "floor")
			}
			if c.Grid[tx][ty-1] {
				coralAttachments = append(coralAttachments, "ceiling")
			}
			if c.Grid[tx-1][ty] {
				coralAttachments = append(coralAttachments, "left")
			}
			if c.Grid[tx+1][ty] {
				coralAttachments = append(coralAttachments, "right")
			}

			if len(coralAttachments) > 0 && r.Float64() < 0.08 {
				attach := coralAttachments[r.Intn(len(coralAttachments))]
				variant := r.Intn(2) // 2 variants for wreckage
				
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
				
				entities = append(entities, entity.NewCoral(cx, cy, entity.CoralBiomeWreckage, attach, variant, r))
			}
		}
	}

	// Spawn 1-2 (random) ElectroWeavers deep in the wreckage cave (depth ty >= 80)
	var candidates []gvec.Vec2
	for tx := 2; tx < gridW-2; tx++ {
		for ty := 80; ty < gridH-2; ty++ {
			if c.Grid[tx][ty] {
				continue
			}
			isOpenSpace := !c.Grid[tx-1][ty] && !c.Grid[tx+1][ty] && !c.Grid[tx][ty-1] && !c.Grid[tx][ty+1]
			if isOpenSpace {
				candidates = append(candidates, gvec.Vec2{X: float64(tx), Y: float64(ty)})
			}
		}
	}

	if len(candidates) > 0 {
		numToSpawn := min(
			// 1 or 2
			r.Intn(2)+1, len(candidates))

		idx1 := r.Intn(len(candidates))
		c1 := candidates[idx1]
		entities = append(entities, &entity.ElectroWeaver{
			BaseEntity: entity.BaseEntity{
				Pos:        gvec.Vec2{X: c1.X*float64(config.TileSize) + (float64(config.TileSize)-40)/2.0, Y: c1.Y*float64(config.TileSize) + (float64(config.TileSize)-20)/2.0},
				Dimensions: gvec.Vec2{X: 40, Y: 20},
				Active:     true,
			},
		})

		if numToSpawn == 2 && len(candidates) > 1 {
			idx2 := idx1
			for range 10 {
				candIdx := r.Intn(len(candidates))
				if candIdx != idx1 {
					dist := math.Hypot(candidates[candIdx].X-c1.X, candidates[candIdx].Y-c1.Y)
					if dist >= 5 {
						idx2 = candIdx
						break
					}
					idx2 = candIdx
				}
			}
			if idx2 != idx1 {
				c2 := candidates[idx2]
				entities = append(entities, &entity.ElectroWeaver{
					BaseEntity: entity.BaseEntity{
						Pos:        gvec.Vec2{X: c2.X*float64(config.TileSize) + (float64(config.TileSize)-40)/2.0, Y: c2.Y*float64(config.TileSize) + (float64(config.TileSize)-20)/2.0},
						Dimensions: gvec.Vec2{X: 40, Y: 20},
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
	return resource.GenerateWreckageResources(c.Grid, seed, c.ShipIndex)
}

func (c *WreckageCorridorCave) GetAmbientColor(lightMult float64) []float32 {
	return []float32{0.01, 0.01, 0.03, 0.97}
}

func GenerateWreckageGrid(r *rand.Rand) [][]bool {
	const (
		w = CaveWidth
		h = CaveHeight
	)
	grid := make([][]bool, w)
	for x := range w {
		grid[x] = make([]bool, h)
		for y := range h {
			grid[x][y] = true
		}
	}

	// 1. Central elevator shaft
	shaftX1 := w/2 - 3 // 27
	shaftX2 := w/2 + 2 // 32
	for y := 0; y < h-4; y++ {
		for x := shaftX1; x <= shaftX2; x++ {
			grid[x][y] = false
		}
	}

	// 2. Horizontal corridors (decks)
	deckYs := []int{24, 52, 80, 108}
	deckHeight := 4

	for _, dy := range deckYs {
		for y := dy; y < dy+deckHeight; y++ {
			for x := 4; x < w-4; x++ {
				grid[x][y] = false
			}
		}
	}

	carveRoom := func(x1, y1, x2, y2 int) {
		for x := x1; x <= x2; x++ {
			for y := y1; y <= y2; y++ {
				grid[x][y] = false
			}
		}
	}

	carveDoor := func(doorX, y1, y2 int) {
		for y := y1; y <= y2; y++ {
			grid[doorX][y] = false
			grid[doorX+1][y] = false
		}
	}

	// 3. Generate rooms branching off corridors
	bays := []struct {
		yMin int
		yMax int
	}{
		{4, deckYs[0] - 1},
		{deckYs[0] + deckHeight, deckYs[1] - 1},
		{deckYs[1] + deckHeight, deckYs[2] - 1},
		{deckYs[2] + deckHeight, deckYs[3] - 1},
	}

	for _, bay := range bays {
		bayH := bay.yMax - bay.yMin + 1
		if bayH < 6 {
			continue
		}

		leftXMin := 4
		leftXMax := shaftX1 - 2
		rightXMin := shaftX2 + 2
		rightXMax := w - 5

		generateRoomsInBay := func(xMin, xMax int, yMin, yMax int, doorToY int) {
			width := xMax - xMin + 1
			if width < 8 {
				return
			}

			numRooms := 2
			if width >= 18 && r.Float64() < 0.6 {
				numRooms = 3
			}

			roomWidth := width / numRooms
			for i := 0; i < numRooms; i++ {
				rx1 := xMin + i*roomWidth + 1
				rx2 := rx1 + roomWidth - 3
				if i == numRooms-1 {
					rx2 = xMax - 1
				}

				ry1 := yMin + 1
				ry2 := yMax - 1

				if rx2 > rx1 && ry2 > ry1 {
					carveRoom(rx1, ry1, rx2, ry2)

					doorX := (rx1 + rx2) / 2
					if doorToY > ry2 {
						carveDoor(doorX, ry2+1, doorToY)
					} else {
						carveDoor(doorX, doorToY, ry1-1)
					}
				}
			}
		}

		if bayH >= 18 {
			midY := (bay.yMin + bay.yMax) / 2
			// Upper half
			generateRoomsInBay(leftXMin, leftXMax, bay.yMin, midY-1, bay.yMin-1)
			generateRoomsInBay(rightXMin, rightXMax, bay.yMin, midY-1, bay.yMin-1)
			// Lower half
			generateRoomsInBay(leftXMin, leftXMax, midY+1, bay.yMax, bay.yMax+1)
			generateRoomsInBay(rightXMin, rightXMax, midY+1, bay.yMax, bay.yMax+1)
		} else {
			doorToY := bay.yMax + 1
			generateRoomsInBay(leftXMin, leftXMax, bay.yMin, bay.yMax, doorToY)
			generateRoomsInBay(rightXMin, rightXMax, bay.yMin, bay.yMax, doorToY)
		}
	}

	// 4. Ensure entrance at top center is open for diving player
	for y := range 5 {
		for x := w/2 - 3; x <= w/2+3; x++ {
			grid[x][y] = false
		}
	}

	return grid
}
