package cave

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

// PreRenderDepthTileImages creates the pre-rendered tile variants for Background
// and Foreground depth layers.
//
// Background tiles: Darkened by 45-50%, tinted toward ambient water haze, zero hard stroke edges.
// Foreground tiles: Very dark silhouette tones with subtle top rim highlights.
func PreRenderDepthTileImages(baseRock color.RGBA, ambientTint color.RGBA, count int) (bgImages, fgImages []*ebiten.Image) {
	if count < 1 {
		count = 8
	}
	ts := config.TileSize
	bgImages = make([]*ebiten.Image, count)
	fgImages = make([]*ebiten.Image, count)

	// Calculate darkened, ambient-tinted background base color
	bgR := uint8(float64(baseRock.R)*0.40 + float64(ambientTint.R)*0.25)
	bgG := uint8(float64(baseRock.G)*0.40 + float64(ambientTint.G)*0.25)
	bgB := uint8(float64(baseRock.B)*0.40 + float64(ambientTint.B)*0.25)
	bgBase := color.RGBA{R: bgR, G: bgG, B: bgB, A: 255}

	// Foreground silhouette color
	fgBase := color.RGBA{10, 14, 22, 255}
	fgRim := color.RGBA{
		R: uint8(math.Min(255, float64(ambientTint.R)*0.6+25)),
		G: uint8(math.Min(255, float64(ambientTint.G)*0.6+45)),
		B: uint8(math.Min(255, float64(ambientTint.B)*0.6+65)),
		A: 255,
	}

	for idx := 0; idx < count; idx++ {
		r := rand.New(rand.NewSource(int64(idx*1337 + 42)))

		// 1. Background image
		bgImg := ebiten.NewImage(ts, ts)
		vector.FillRect(bgImg, 0, 0, float32(ts), float32(ts), bgBase, false)
		// Subtle soft stippling for texture without harsh lines
		for i := 0; i < 6; i++ {
			px := float32(r.Intn(ts - 8) + 4)
			py := float32(r.Intn(ts - 8) + 4)
			subtleClr := color.RGBA{
				R: uint8(max(0, int(bgR)-10)),
				G: uint8(max(0, int(bgG)-10)),
				B: uint8(max(0, int(bgB)-10)),
				A: 255,
			}
			vector.FillCircle(bgImg, px, py, 2.0, subtleClr, false)
		}
		bgImages[idx] = bgImg

		// 2. Foreground image
		fgImg := ebiten.NewImage(ts, ts)
		vector.FillRect(fgImg, 0, 0, float32(ts), float32(ts), fgBase, false)
		// Subtle top-edge rim highlight
		vector.StrokeLine(fgImg, 0, 1, float32(ts), 1, 1.2, fgRim, false)
		// Subtle side bevel
		if idx%2 == 0 {
			vector.StrokeLine(fgImg, 1, 1, 1, float32(ts), 0.8, color.RGBA{fgRim.R / 2, fgRim.G / 2, fgRim.B / 2, 255}, false)
		}
		fgImages[idx] = fgImg
	}

	return bgImages, fgImages
}

// CarveSubGroundTunnels inspects seabed floor heights and carves organic
// sub-ground tunnels connecting surface valleys or branching into deep grottoes.
func CarveSubGroundTunnels(grid [][]bool, colFloorHeights []int, r *rand.Rand) {
	if r.Float64() >= ActiveCaveDepthConfig.TunnelOccurrenceChance {
		return // Tunnels skipped per adjustable occurrence rate
	}

	gridW := len(grid)
	gridH := len(grid[0])
	if gridW < 20 || gridH < 30 {
		return
	}

	// Helper to carve circular brush
	radius := ActiveCaveDepthConfig.TunnelRadius
	if radius < 1 {
		radius = 2
	}
	carveTunnel := func(cx, cy int) {
		for dx := -radius; dx <= radius; dx++ {
			for dy := -radius; dy <= radius; dy++ {
				if dx*dx+dy*dy <= radius*radius+1 {
					nx, ny := cx+dx, cy+dy
					if nx > 1 && nx < gridW-2 && ny > 2 && ny < gridH-3 {
						grid[nx][ny] = false // Open water
					}
				}
			}
		}
	}

	// Helper to carve smooth line connecting two points
	carveLine := func(x0, y0, x1, y1 int) {
		dx := math.Abs(float64(x1 - x0))
		dy := math.Abs(float64(y1 - y0))
		steps := int(math.Max(dx, dy)) * 2
		if steps <= 0 {
			steps = 1
		}
		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(steps)
			jitter := math.Sin(t*math.Pi*3.0) * 1.5
			cx := int(float64(x0) + t*float64(x1-x0) + jitter)
			cy := int(float64(y0) + t*float64(y1-y0))
			carveTunnel(cx, cy)
		}
	}

	// Find surface valley depressions (left zone: x=10..26, right zone: x=34..50)
	valley1X, minFl1 := 16, 999
	for x := 10; x <= 26 && x < gridW; x++ {
		if colFloorHeights[x] < minFl1 {
			minFl1 = colFloorHeights[x]
			valley1X = x
		}
	}

	valley2X, minFl2 := 42, 999
	for x := 34; x <= 50 && x < gridW; x++ {
		if colFloorHeights[x] < minFl2 {
			minFl2 = colFloorHeights[x]
			valley2X = x
		}
	}

	fl1 := colFloorHeights[valley1X]
	fl2 := colFloorHeights[valley2X]

	// Determine subterranean belly depth (at least 12 tiles down)
	bellyY := max(fl1, fl2) + 10 + r.Intn(16)
	if bellyY > gridH-8 {
		bellyY = gridH - 8
	}

	midBellyX1 := valley1X + 4 + r.Intn(4)
	midBellyX2 := valley2X - 4 - r.Intn(4)
	if midBellyX2 < midBellyX1 {
		midBellyX2 = midBellyX1 + 2
	}

	// 1. Carve descent from valley 1
	carveLine(valley1X, fl1-1, midBellyX1, bellyY)
	// 2. Carve horizontal subterranean corridor
	carveLine(midBellyX1, bellyY, midBellyX2, bellyY)
	// 3. Carve ascent back up to valley 2
	carveLine(midBellyX2, bellyY, valley2X, fl2-1)

	// 4. Optional branch into a deep high-reward grotto chamber
	if r.Float64() < ActiveCaveDepthConfig.DeepGrottoChance {
		grottoX := min(gridW-10, max(valley2X+6, midBellyX2+6))
		grottoY := min(gridH-8, bellyY+6)
		carveLine(midBellyX2, bellyY, grottoX, grottoY)

		// Carve circular/elliptical grotto chamber
		grottoRadX := 5 + r.Intn(3)
		grottoRadY := 4 + r.Intn(3)
		for dx := -grottoRadX; dx <= grottoRadX; dx++ {
			for dy := -grottoRadY; dy <= grottoRadY; dy++ {
				normDist := float64(dx*dx)/float64(grottoRadX*grottoRadX) + float64(dy*dy)/float64(grottoRadY*grottoRadY)
				if normDist <= 1.0 {
					nx, ny := grottoX+dx, grottoY+dy
					if nx > 1 && nx < gridW-2 && ny > 1 && ny < gridH-2 {
						grid[nx][ny] = false
					}
				}
			}
		}
	}
}

// GenerateDepthGrids builds BgGrid and FgGrid for a natural cave:
// - All sub-ground open water (below surface seabed) is ALWAYS Background.
// - Surface valley mouths receive flaring background entry alcoves.
// - Natural arches and crossing pillars form Foreground framing.
func GenerateDepthGrids(grid [][]bool, isShallow bool) (bgGrid, fgGrid [][]bool) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return nil, nil
	}

	w := len(grid)
	h := len(grid[0])

	bgGrid = make([][]bool, w)
	for x := range w {
		bgGrid[x] = make([]bool, h)
	}

	// 1. Detect seabed surface floor contour across the cave.
	// Find raw first solid block in each column.
	rawFloor := make([]int, w)
	for x := 0; x < w; x++ {
		fl := h
		for y := 0; y < h; y++ {
			if grid[x][y] {
				fl = y
				break
			}
		}
		rawFloor[x] = fl
	}

	// Reconstruct smooth surface floor ignoring narrow tunnel entrance openings
	surfaceFloor := make([]int, w)
	const window = 8
	for x := 0; x < w; x++ {
		minFl := rawFloor[x]
		for dx := -window; dx <= window; dx++ {
			nx := x + dx
			if nx >= 0 && nx < w {
				if rawFloor[nx] < minFl {
					minFl = rawFloor[nx]
				}
			}
		}
		surfaceFloor[x] = minFl + 2
	}

	// 2. Identify Sub-Ground Tunnels & Surface Dips
	for x := 0; x < w; x++ {
		fl := surfaceFloor[x]
		hasRockAbove := false
		for y := 0; y < h; y++ {
			if grid[x][y] {
				hasRockAbove = true
			} else {
				// Sub-ground open water: under a solid rock ceiling OR below surface floor line!
				if hasRockAbove || y >= fl {
					bgGrid[x][y] = true
				} else if isShallow && y >= fl-4 {
					// Surface alcoves in dips
					leftFl := fl
					if x > 0 {
						leftFl = surfaceFloor[x-1]
					}
					rightFl := fl
					if x < w-1 {
						rightFl = surfaceFloor[x+1]
					}
					if fl >= leftFl && fl >= rightFl {
						bgGrid[x][y] = true
					}
				}
			}
		}
	}

	// In deep caves (non-shallow), open areas with solid rock above/below get 35% background backdrops
	if !isShallow {
		for x := 2; x < w-2; x++ {
			for y := 4; y < h-4; y++ {
				if !grid[x][y] && (grid[x][y-1] || grid[x][y+1] || grid[x-1][y] || grid[x+1][y]) {
					hVal := hashCoords(x, y)
					if hVal%100 < 45 {
						bgGrid[x][y] = true
					}
				}
			}
		}
	}

	// Foreground tiles removed for now per design feedback (return nil fgGrid)
	return bgGrid, nil
}

// CalcDepthTileColors computes the darkened, ambient-tinted background base and stroke colors.
func CalcDepthTileColors(baseRock, ambientTint color.RGBA) (bgBase, bgStroke color.RGBA) {
	bgR := uint8(float64(baseRock.R)*0.40 + float64(ambientTint.R)*0.25)
	bgG := uint8(float64(baseRock.G)*0.40 + float64(ambientTint.G)*0.25)
	bgB := uint8(float64(baseRock.B)*0.40 + float64(ambientTint.B)*0.25)
	bgBase = color.RGBA{R: bgR, G: bgG, B: bgB, A: 255}
	bgStroke = color.RGBA{
		R: uint8(max(0, int(bgR)-12)),
		G: uint8(max(0, int(bgG)-12)),
		B: uint8(max(0, int(bgB)-12)),
		A: 255,
	}
	return bgBase, bgStroke
}

// DrawBackgroundTileLayer renders the background depth tiles with organic curves and flowing corners.
func DrawBackgroundTileLayer(screen *ebiten.Image, bgGrid [][]bool, bgBase, bgStroke color.RGBA, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	if bgGrid == nil {
		return
	}
	minTX := max(0, startTileX-1)
	maxTX := min(len(bgGrid), endTileX+1)
	minTY := max(0, startTileY-1)
	maxTY := min(len(bgGrid[0]), endTileY+1)

	for tx := minTX; tx < maxTX; tx++ {
		for ty := minTY; ty < maxTY; ty++ {
			DrawOrganicCaveTile(screen, bgGrid, tx, ty, camX, camY, bgBase, bgStroke, nil)
		}
	}
}

// DrawForegroundTileLayer renders the foreground framing arches and pillars.
func DrawForegroundTileLayer(screen *ebiten.Image, fgGrid [][]bool, fgImages []*ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int, fgAlpha float32) {
	if fgGrid == nil || len(fgImages) == 0 || fgAlpha <= 0.005 {
		return
	}
	ts := config.TileSize
	op := ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1.0, 1.0, 1.0, fgAlpha)

	for tx := startTileX; tx < endTileX; tx++ {
		if tx < 0 || tx >= len(fgGrid) {
			continue
		}
		for ty := startTileY; ty < endTileY; ty++ {
			if ty < 0 || ty >= len(fgGrid[0]) || !fgGrid[tx][ty] {
				continue
			}

			sx := float64(tx*ts - int(camX))
			sy := float64(ty*ts - int(camY))
			h := hashCoords(tx, ty)
			variantIdx := h % uint64(len(fgImages))

			op.GeoM.Reset()
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(fgImages[variantIdx], &op)
		}
	}
}
