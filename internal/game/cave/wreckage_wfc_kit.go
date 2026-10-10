package cave

import (
	"github.com/jaredwarren/pcg/relaxwfc"
)

// Wreckage module-kit TileIDs — adapter-only mapping into bool solid grids.
// Organic CA biomes never use this path.
const (
	wfcWall     relaxwfc.TileID = 0
	wfcRoom     relaxwfc.TileID = 1
	wfcCorridor relaxwfc.TileID = 2
	wfcDoor     relaxwfc.TileID = 3
	wfcFail     relaxwfc.TileID = 4
)

const (
	wfcModuleW = 3
	wfcModuleH = 3
)

// wreckageKitRules returns hard adjacency for a ship-interior module kit plus
// soft preferences that may drop under Family-102 relaxation.
func wreckageKitRules() []relaxwfc.Rule {
	var rules []relaxwfc.Rule
	add := func(a, b relaxwfc.TileID, d relaxwfc.Direction, hard bool, w float32) {
		rules = append(rules, relaxwfc.Rule{A: a, B: b, Dir: d, Hard: hard, Weight: w})
	}
	both := func(a, b relaxwfc.TileID, hard bool, w float32) {
		for _, d := range []relaxwfc.Direction{relaxwfc.North, relaxwfc.East, relaxwfc.South, relaxwfc.West} {
			add(a, b, d, hard, w)
			add(b, a, d, hard, w)
		}
	}

	// Hard: structural kit vocabulary.
	both(wfcWall, wfcWall, true, 1)
	both(wfcWall, wfcRoom, true, 1)
	both(wfcWall, wfcCorridor, true, 1)
	both(wfcWall, wfcDoor, true, 1)
	both(wfcRoom, wfcRoom, true, 1)
	both(wfcRoom, wfcDoor, true, 1)
	both(wfcCorridor, wfcCorridor, true, 1)
	both(wfcCorridor, wfcDoor, true, 1)
	both(wfcDoor, wfcDoor, true, 1)

	// Soft: prefer rooms not abutting raw corridor (door in between) — droppable.
	both(wfcRoom, wfcCorridor, false, 0.15)

	return rules
}

// carveBayWithRelaxWFC fills [xMin,xMax]×[yMin,yMax] using a coarse module WFC
// kit, then ensures a 2-wide link to doorToY (deck corridor). Returns false if
// the bay is too small or Solve fails — caller should use legacy carving.
func carveBayWithRelaxWFC(grid [][]bool, seed uint64, xMin, xMax, yMin, yMax, doorToY int) bool {
	width := xMax - xMin + 1
	height := yMax - yMin + 1
	mw := width / wfcModuleW
	mh := height / wfcModuleH
	if mw < 2 || mh < 2 {
		return false
	}

	// MaxSteps only (no wall-clock Budget) so the same world seed stays
	// deterministic across machines. Module grids are tiny (~≤8×6).
	res, err := relaxwfc.Solve(relaxwfc.Config{
		Width:        mw,
		Height:       mh,
		Tiles:        []relaxwfc.TileID{wfcWall, wfcRoom, wfcCorridor, wfcDoor, wfcFail},
		Rules:        wreckageKitRules(),
		Seed:         seed,
		MaxSteps:     8_000,
		SoftFailTile: wfcFail,
	})
	if err != nil {
		return false
	}

	// Stamp modules (fail → wall).
	for my := 0; my < mh; my++ {
		for mx := 0; mx < mw; mx++ {
			tile := res.Grid[my][mx]
			ox := xMin + mx*wfcModuleW
			oy := yMin + my*wfcModuleH
			stampWreckageModule(grid, ox, oy, tile)
		}
	}

	// Connect an open module cell to the deck corridor at doorToY.
	connectBayToDeck(grid, xMin, xMax, yMin, yMax, doorToY)
	return true
}

func stampWreckageModule(grid [][]bool, ox, oy int, tile relaxwfc.TileID) {
	open := tile == wfcRoom || tile == wfcCorridor || tile == wfcDoor
	for dy := 0; dy < wfcModuleH; dy++ {
		for dx := 0; dx < wfcModuleW; dx++ {
			x, y := ox+dx, oy+dy
			if x < 0 || y < 0 || x >= len(grid) || y >= len(grid[0]) {
				continue
			}
			if open {
				grid[x][y] = false
			} else {
				grid[x][y] = true
			}
		}
	}
}

func connectBayToDeck(grid [][]bool, xMin, xMax, yMin, yMax, doorToY int) {
	cx := (xMin + xMax) / 2
	bestX := -1
	bestDist := int(^uint(0) >> 1)
	for x := xMin; x <= xMax; x++ {
		open := false
		for y := yMin; y <= yMax; y++ {
			if !grid[x][y] {
				open = true
				break
			}
		}
		if !open {
			continue
		}
		d := absInt(x - cx)
		if d < bestDist {
			bestDist = d
			bestX = x
		}
	}
	if bestX < 0 {
		// Force a small room if WFC produced all walls.
		rx1, rx2 := cx-2, cx+2
		if rx1 < xMin {
			rx1 = xMin
		}
		if rx2 > xMax {
			rx2 = xMax
		}
		ry1, ry2 := yMin+1, yMax-1
		if ry2 <= ry1 {
			ry1, ry2 = yMin, yMax
		}
		for x := rx1; x <= rx2; x++ {
			for y := ry1; y <= ry2; y++ {
				grid[x][y] = false
			}
		}
		bestX = (rx1 + rx2) / 2
	}

	doorX := bestX
	if doorX+1 > xMax {
		doorX = xMax - 1
	}
	if doorX < xMin {
		doorX = xMin
	}

	yStart, yEnd := yMin, doorToY
	if doorToY < yMin {
		yStart, yEnd = doorToY, yMax
	}
	if yStart > yEnd {
		yStart, yEnd = yEnd, yStart
	}
	for y := yStart; y <= yEnd; y++ {
		if y < 0 || y >= len(grid[0]) {
			continue
		}
		grid[doorX][y] = false
		if doorX+1 <= xMax {
			grid[doorX+1][y] = false
		}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
