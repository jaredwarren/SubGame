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
	const lineGap = 40.0
	offsetX := float32(math.Mod(camY*0.1, lineGap))

	switch c.ShipIndex {
	case 1:
		// Ship 1: Submersible Transport (rusted industrial iron interior, warm amber grid)
		screen.Fill(color.RGBA{18, 16, 15, 255})
		for x := float32(0); x < float32(config.ScreenWidth); x += lineGap {
			vector.StrokeLine(screen, x, 0, x, float32(config.ScreenHeight), 0.8, color.RGBA{35, 28, 22, 255}, false)
		}
		for y := float32(0); y < float32(config.ScreenHeight); y += lineGap {
			sy := y - offsetX
			vector.StrokeLine(screen, 0, sy, float32(config.ScreenWidth), sy, 0.8, color.RGBA{35, 28, 22, 255}, false)
		}
	case 2:
		// Ship 2: AetherCorp Flagship (dark obsidian armor, faint crimson grid lines)
		screen.Fill(color.RGBA{10, 10, 14, 255})
		for x := float32(0); x < float32(config.ScreenWidth); x += lineGap {
			vector.StrokeLine(screen, x, 0, x, float32(config.ScreenHeight), 0.8, color.RGBA{30, 14, 18, 255}, false)
		}
		for y := float32(0); y < float32(config.ScreenHeight); y += lineGap {
			sy := y - offsetX
			vector.StrokeLine(screen, 0, sy, float32(config.ScreenWidth), sy, 0.8, color.RGBA{30, 14, 18, 255}, false)
		}
	default:
		// Ship 0: Research Tender (clinical cyan-steel interior, cool grid)
		screen.Fill(color.RGBA{12, 16, 22, 255})
		for x := float32(0); x < float32(config.ScreenWidth); x += lineGap {
			vector.StrokeLine(screen, x, 0, x, float32(config.ScreenHeight), 0.8, color.RGBA{18, 30, 42, 255}, false)
		}
		for y := float32(0); y < float32(config.ScreenHeight); y += lineGap {
			sy := y - offsetX
			vector.StrokeLine(screen, 0, sy, float32(config.ScreenWidth), sy, 0.8, color.RGBA{18, 30, 42, 255}, false)
		}
	}
}

func (c *WreckageCorridorCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	var rockColor, strokeColor, stripeColor color.RGBA
	switch c.ShipIndex {
	case 1:
		// Oxidized heavy iron hull, weathered industrial hazard amber
		rockColor = color.RGBA{58, 50, 44, 255}
		strokeColor = color.RGBA{92, 78, 68, 255}
		stripeColor = color.RGBA{225, 155, 25, 190}
	case 2:
		// Reinforced dark charcoal / obsidian armor, high-security crimson hazard stripes
		rockColor = color.RGBA{28, 30, 36, 255}
		strokeColor = color.RGBA{55, 58, 68, 255}
		stripeColor = color.RGBA{215, 45, 40, 190}
	default:
		// Research Tender: cool steel-cyan bulkhead panels, clean science teal stripes
		rockColor = color.RGBA{42, 60, 75, 255}
		strokeColor = color.RGBA{65, 90, 110, 255}
		stripeColor = color.RGBA{40, 195, 185, 180}
	}

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.Grid[tx][ty] {
				sx := float32(tx*config.TileSize - int(camX))
				sy := float32(ty*config.TileSize - int(camY))

				vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, rockColor, false)
				vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 1.2, strokeColor, false)

				// Bulkhead structural corner rivets
				vector.FillCircle(screen, sx+2.5, sy+2.5, 0.8, strokeColor, false)
				vector.FillCircle(screen, sx+config.TileSize-2.5, sy+2.5, 0.8, strokeColor, false)
				vector.FillCircle(screen, sx+2.5, sy+config.TileSize-2.5, 0.8, strokeColor, false)
				vector.FillCircle(screen, sx+config.TileSize-2.5, sy+config.TileSize-2.5, 0.8, strokeColor, false)

				// Draw diagonal hazard warning lines along bulkheads that border open corridors or rooms
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
	ents := GenerateDeepEntitiesFromSpec(Spec(CaveWreckage), c.Grid, seed)

	var filtered []entity.CaveEntity
	for _, ent := range ents {
		if sb, ok := ent.(*entity.ShatterBulb); ok {
			ty := int(sb.Pos.Y) / config.TileSize
			// Zero O2 bubbles spawn below 40m depth for all wrecked ships
			if ty >= 40 {
				continue
			}
		}
		filtered = append(filtered, ent)
	}

	r := rand.New(rand.NewSource(seed + 99))
	gridW := len(c.Grid)
	gridH := len(c.Grid[0])
	ts := config.TileSize

	// Identify room floors, ceilings, and walls (excluding central shaft)
	var floorTiles [][2]int
	var ceilingTiles [][2]int
	var wallTiles [][2]int

	for tx := 4; tx < gridW-4; tx++ {
		for ty := 4; ty < gridH-4; ty++ {
			if !c.Grid[tx][ty] && !isWreckageCorridorOrShaftTile(tx, ty) {
				// Floor tile
				if c.Grid[tx][ty+1] {
					floorTiles = append(floorTiles, [2]int{tx, ty})
				}
				// Ceiling tile
				if c.Grid[tx][ty-1] {
					ceilingTiles = append(ceilingTiles, [2]int{tx, ty})
				}
				// Wall tile
				if c.Grid[tx-1][ty] || c.Grid[tx+1][ty] {
					wallTiles = append(wallTiles, [2]int{tx, ty})
				}
			}
		}
	}

	// 1. Spawn Scrap Hermit Crabs on room floors
	numCrabs := 3 + r.Intn(2) // 3 to 4 crabs
	if len(floorTiles) > 0 {
		shuffledFloors := make([][2]int, len(floorTiles))
		copy(shuffledFloors, floorTiles)
		r.Shuffle(len(shuffledFloors), func(i, j int) {
			shuffledFloors[i], shuffledFloors[j] = shuffledFloors[j], shuffledFloors[i]
		})

		for i := 0; i < numCrabs && i < len(shuffledFloors); i++ {
			pt := shuffledFloors[i]
			var shell entity.ShellVariant
			switch c.ShipIndex {
			case 1:
				if r.Float64() < 0.6 {
					shell = entity.ShellPipeElbow
				} else {
					shell = entity.ShellCogGear
				}
			case 2:
				if r.Float64() < 0.65 {
					shell = entity.ShellCogGear
				} else {
					shell = entity.ShellPipeElbow
				}
			default:
				if r.Float64() < 0.6 {
					shell = entity.ShellTinCan
				} else {
					shell = entity.ShellPipeElbow
				}
			}
			crab := entity.NewScrapHermitCrabWithShell(
				float64(pt[0]*ts)+2.0,
				float64(pt[1]*ts)+float64(ts-12),
				shell,
			)
			filtered = append(filtered, crab)
		}
	}

	// 2. Spawn Wreck Terminal
	var targetMinY, targetMaxY int
	switch c.ShipIndex {
	case 1: // Transport: Mid deck (cargo/engineering)
		targetMinY, targetMaxY = 28, 55
	case 2: // Flagship: Command deck
		targetMinY, targetMaxY = 60, 95
	default: // Research: Science deck
		targetMinY, targetMaxY = 6, 25
	}

	var termCandidates [][2]int
	for _, pt := range wallTiles {
		if pt[1] >= targetMinY && pt[1] <= targetMaxY {
			termCandidates = append(termCandidates, pt)
		}
	}
	if len(termCandidates) == 0 {
		termCandidates = wallTiles
	}
	if len(termCandidates) > 0 {
		chosen := termCandidates[r.Intn(len(termCandidates))]
		term := entity.NewWreckTerminal(
			float64(chosen[0]*ts)+4.0,
			float64(chosen[1]*ts)+4.0,
			c.ShipIndex,
		)
		filtered = append(filtered, term)
	}

	// 3. Spawn Sparking Conduits
	numConduits := 1
	if c.ShipIndex == 1 {
		numConduits = 3 + r.Intn(2) // 3-4
	} else if c.ShipIndex == 2 {
		numConduits = 5 + r.Intn(2) // 5-6
	}

	if len(ceilingTiles) > 0 {
		shuffledCeilings := make([][2]int, len(ceilingTiles))
		copy(shuffledCeilings, ceilingTiles)
		r.Shuffle(len(shuffledCeilings), func(i, j int) {
			shuffledCeilings[i], shuffledCeilings[j] = shuffledCeilings[j], shuffledCeilings[i]
		})

		for i := 0; i < numConduits && i < len(shuffledCeilings); i++ {
			pt := shuffledCeilings[i]
			conduit := entity.NewSparkingConduit(
				float64(pt[0]*ts)+8.0,
				float64(pt[1]*ts),
				seed+int64(i*37),
			)
			filtered = append(filtered, conduit)
		}
	}

	return filtered
}

func (c *WreckageCorridorCave) GenerateResources(seed int64) []resource.Resource {
	// Scrap nodes are generated inside wreckage caves instead of mineral nodes
	return resource.GenerateWreckageResources(c.Grid, seed, c.ShipIndex)
}

func (c *WreckageCorridorCave) GetAmbientColor(lightMult float64) [4]float32 {
	switch c.ShipIndex {
	case 1:
		return [4]float32{0.022, 0.016, 0.012, 0.96}
	case 2:
		return [4]float32{0.026, 0.008, 0.012, 0.97}
	default:
		return [4]float32{0.015, 0.022, 0.035, 0.95}
	}
}

func isWreckageCorridorOrShaftTile(tx, ty int) bool {
	// Central elevator shaft
	if tx >= 27 && tx <= 32 {
		return true
	}
	// Horizontal corridors (decks)
	if (ty >= 2 && ty <= 3) ||
		(ty >= 24 && ty <= 27) ||
		(ty >= 52 && ty <= 55) ||
		(ty >= 80 && ty <= 83) ||
		(ty >= 108 && ty <= 111) {
		return true
	}
	return false
}

// GenerateWreckageGrid generates the baseline wreckage cave layout (Ship 0).
func GenerateWreckageGrid(r *rand.Rand) [][]bool {
	return GenerateWreckageGridWithShip(r, 0)
}

// GenerateWreckageGridWithShip generates the wreckage cave layout with structural degradation per tier.
func GenerateWreckageGridWithShip(r *rand.Rand, shipIndex int) [][]bool {
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

	// Top bridge corridor connecting upper deck 0 rooms to central shaft
	for y := 2; y <= 3; y++ {
		for x := 4; x < w-4; x++ {
			grid[x][y] = false
		}
	}

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

	// 5. Tier-specific structural degradation
	if shipIndex == 1 {
		// Ship 1 (Transport): Collapsed deck floor sections (vertical drop-through shortcuts)
		for _, dy := range []int{23, 51} {
			for x := 8; x < w-8; x += 5 {
				if x >= shaftX1-4 && x <= shaftX2+4 {
					continue
				}
				if !grid[x][dy-1] && grid[x][dy] && !grid[x][dy+1] {
					if r.Float64() < 0.5 {
						grid[x][dy] = false
						if x+1 < w && grid[x+1][dy] && !grid[x+1][dy-1] && !grid[x+1][dy+1] {
							grid[x+1][dy] = false
						}
					}
				}
			}
		}
	} else if shipIndex == 2 {
		// Ship 2 (Flagship): Structural hull ruptures and outer abyss breaches
		for _, dy := range []int{23, 51, 79} {
			for x := 8; x < w-8; x += 5 {
				if x >= shaftX1-4 && x <= shaftX2+4 {
					continue
				}
				if !grid[x][dy-1] && grid[x][dy] && !grid[x][dy+1] {
					if r.Float64() < 0.6 {
						grid[x][dy] = false
						if x+1 < w && grid[x+1][dy] && !grid[x+1][dy-1] && !grid[x+1][dy+1] {
							grid[x+1][dy] = false
						}
					}
				}
			}
		}

		// Outer hull breaches in upper/mid bays (never in bottom bay 3 where Deep Vault is)
		breachYs := []int{36, 68}
		for _, by := range breachYs {
			if !grid[4][by] {
				grid[3][by] = false
				grid[2][by] = false
				grid[1][by] = false
				grid[2][by+1] = false
			}
			if !grid[w-5][by] {
				grid[w-4][by] = false
				grid[w-3][by] = false
				grid[w-2][by] = false
				grid[w-3][by+1] = false
			}
		}
	}

	return grid
}
