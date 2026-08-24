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

type ThermoCave struct {
	Grid  [][]bool
	ticks int
}

func NewThermoCave(grid [][]bool) *ThermoCave {
	return &ThermoCave{Grid: grid}
}

func (c *ThermoCave) GetCaveType() CaveType { return CaveThermo }
func (c *ThermoCave) GetGrid() [][]bool     { return c.Grid }

func (c *ThermoCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Dark, warm basalt grey background
	screen.Fill(color.RGBA{16, 10, 10, 255})

	c.ticks++

	// Parallax scrolling floating sparks/embers rising upwards
	const numParticles = 40
	for i := 0; i < numParticles; i++ {
		seed := uint64(i * 9876)
		px := fracHash(seed) * float64(config.ScreenWidth)
		pyInitial := fracHash(seed+1) * float64(config.ScreenHeight*2)

		// Parallax scroll factor: 0.35x camera speed, plus steady upward float
		riseOffset := float64(c.ticks) * (0.3 + fracHash(seed+2)*0.4)
		py := math.Mod(pyInitial-camY*0.35-riseOffset, float64(config.ScreenHeight*2))
		if py < 0 {
			py += float64(config.ScreenHeight * 2)
		}

		if py < float64(config.ScreenHeight) {
			size := 1.2 + fracHash(seed+3)*1.8
			var pClr color.RGBA
			roll := fracHash(seed + 4)
			if roll < 0.50 {
				pClr = color.RGBA{255, 95, 10, 170} // Hot Orange
			} else if roll < 0.85 {
				pClr = color.RGBA{225, 25, 5, 170} // Glowing Red
			} else {
				pClr = color.RGBA{255, 210, 30, 190} // Bright Yellow Spark
			}
			vector.FillCircle(screen, float32(px), float32(py), float32(size), pClr, false)
		}
	}
}

func (c *ThermoCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if tx >= 0 && tx < len(c.Grid) && ty >= 0 && ty < len(c.Grid[0]) {
				if c.Grid[tx][ty] {
					sx := float32(tx*config.TileSize - int(camX))
					sy := float32(ty*config.TileSize - int(camY))

					// Dark cooled-basalt stone colors
					rockColor := color.RGBA{30, 26, 26, 255}
					strokeColor := color.RGBA{46, 38, 38, 255}

					vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, rockColor, false)
					vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 0.5, strokeColor, false)

					// Deterministic vein cracks inside the rock
					h := hashCoords(tx, ty)
					if h%3 == 0 {
						var veinClr color.RGBA
						if h%2 == 0 {
							veinClr = color.RGBA{235, 65, 15, 120} // Hot orange crack
						} else {
							veinClr = color.RGBA{205, 20, 5, 90} // Deep magma red crack
						}
						// Draw a diagonal crack
						vector.StrokeLine(screen, sx+6, sy+6, sx+16, sy+16, 1.2, veinClr, false)
						vector.StrokeLine(screen, sx+16, sy+16, sx+12, sy+24, 1.0, veinClr, false)
					}
					if h%5 == 0 {
						// Subtle second crack/vein (brighter orange)
						veinClr := color.RGBA{255, 125, 20, 100}
						vector.StrokeLine(screen, sx+float32(config.TileSize)-8, sy+8, sx+float32(config.TileSize)-16, sy+20, 1.0, veinClr, false)
					}
				}
			}
		}
	}
}

func (c *ThermoCave) GenerateEntities(seed int64) []entity.CaveEntity {
	return GenerateDeepEntitiesFromSpec(Spec(CaveThermo), c.Grid, seed)
}

func (c *ThermoCave) GenerateResources(seed int64) []resource.Resource {
	return resource.GenerateResourceNodes(c.Grid, seed)
}

func (c *ThermoCave) GetAmbientColor(lightMult float64) [4]float32 {
	return AmbientColor(CaveThermo)
}

// GenerateThermoCaveGrid generates a shallow 60x60 grid with magma chambers and narrow crevices.
func GenerateThermoCaveGrid(r *rand.Rand) [][]bool {
	const (
		w = 60
		h = 60
	)

	// Starts completely solid
	grid := make([][]bool, w)
	for x := range w {
		grid[x] = make([]bool, h)
		for y := range h {
			grid[x][y] = true
		}
	}

	// Helper to carve a circular region
	carveCircle := func(cx, cy, radius int) {
		for dx := -radius; dx <= radius; dx++ {
			for dy := -radius; dy <= radius; dy++ {
				if dx*dx+dy*dy <= radius*radius {
					nx, ny := cx+dx, cy+dy
					if nx > 0 && nx < w-1 && ny > 0 && ny < h-1 {
						grid[nx][ny] = false
					}
				}
			}
		}
	}

	// Helper to carve a brush point (2x2)
	carveBrush := func(cx, cy int) {
		for dx := 0; dx <= 1; dx++ {
			for dy := 0; dy <= 1; dy++ {
				nx, ny := cx+dx, cy+dy
				if nx > 0 && nx < w-1 && ny > 0 && ny < h-1 {
					grid[nx][ny] = false
				}
			}
		}
	}

	// 1. Carve 3 large magma chambers (basins)
	// Chamber 1: Upper-Mid-Left
	ch1X, ch1Y := 18, 16
	carveCircle(ch1X, ch1Y, 7)

	// Chamber 2: Mid-Right
	ch2X, ch2Y := 42, 30
	carveCircle(ch2X, ch2Y, 8)

	// Chamber 3: Bottom-Left
	ch3X, ch3Y := 22, 45
	carveCircle(ch3X, ch3Y, 7)

	// 2. Run random walkers to carve winding connecting crevices (lava tubes)
	// We run 3 walkers starting from the entrance.
	const numWalkers = 3
	for i := 0; i < numWalkers; i++ {
		cx := w/2 + r.Intn(4) - 2
		cy := 2

		// Walk down to the bottom
		for cy < h-2 {
			roll := r.Float64()
			switch {
			case roll < 0.54:
				cy++ // Down (biased)
			case roll < 0.76:
				cx-- // Left
			case roll < 0.98:
				cx++ // Right
			default:
				cy-- // Up
			}

			// Attraction forces to pull walkers through our magma chambers at depth thresholds
			if cy == 14 && r.Float64() < 0.8 {
				cx = ch1X + r.Intn(4) - 2
			} else if cy == 28 && r.Float64() < 0.8 {
				cx = ch2X + r.Intn(4) - 2
			} else if cy == 42 && r.Float64() < 0.8 {
				cx = ch3X + r.Intn(4) - 2
			}

			// Clamp
			if cx < 2 {
				cx = 2
			}
			if cx > w-3 {
				cx = w - 3
			}
			if cy < 2 {
				cy = 2
			}

			carveBrush(cx, cy)
		}
	}

	// 3. Ensure entrance at top center is fully open and cleared
	for y := 0; y < 5; y++ {
		for x := (w / 2) - 3; x <= (w/2)+3; x++ {
			if x > 0 && x < w-1 && y < h-1 {
				grid[x][y] = false
			}
		}
	}

	return grid
}
