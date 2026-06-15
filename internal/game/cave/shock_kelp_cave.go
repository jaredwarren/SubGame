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

type ShockKelpCave struct {
	Grid [][]bool
}

func NewShockKelpCave(grid [][]bool) *ShockKelpCave {
	return &ShockKelpCave{Grid: grid}
}

func (c *ShockKelpCave) GetCaveType() CaveType { return CaveShockKelp }
func (c *ShockKelpCave) GetGrid() [][]bool     { return c.Grid }

func (c *ShockKelpCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Dark purple-grey background representing a charged, deep reef grotto
	screen.Fill(color.RGBA{15, 12, 22, 255})

	// Parallax scrolling electric cyan & purple floating particles
	const numParticles = 40
	for i := 0; i < numParticles; i++ {
		rng := rand.New(rand.NewSource(int64(i * 1234)))
		px := rng.Float64() * float64(config.ScreenWidth)
		pyInitial := rng.Float64() * float64(config.ScreenHeight * 2)

		// Parallax scroll factor: 0.35x camera speed
		py := math.Mod(pyInitial-camY*0.35, float64(config.ScreenHeight*2))
		if py < 0 {
			py += float64(config.ScreenHeight * 2)
		}

		if py < float64(config.ScreenHeight) {
			size := 1.0 + rng.Float64()*2.0
			var pClr color.RGBA
			if rng.Float64() < 0.75 {
				pClr = color.RGBA{0, 220, 255, 160} // Electric Cyan
			} else {
				pClr = color.RGBA{160, 45, 230, 160} // Purple
			}
			vector.FillCircle(screen, float32(px), float32(py), float32(size), pClr, false)
		}
	}
}

func (c *ShockKelpCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if tx >= 0 && tx < len(c.Grid) && ty >= 0 && ty < len(c.Grid[0]) {
				if c.Grid[tx][ty] {
					sx := float32(tx*config.TileSize - int(camX))
					sy := float32(ty*config.TileSize - int(camY))

					// Slate-grey rock colors
					rockColor := color.RGBA{55, 60, 68, 255}
					strokeColor := color.RGBA{82, 88, 98, 255}

					vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, rockColor, false)
					vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 0.5, strokeColor, false)

					// Deterministic vein cracks inside the rock
					h := hashCoords(tx, ty)
					if h%3 == 0 {
						var veinClr color.RGBA
						if h%2 == 0 {
							veinClr = color.RGBA{140, 50, 210, 90} // Electric purple vein
						} else {
							veinClr = color.RGBA{0, 210, 240, 75}  // Glowing cyan vein
						}
						// Draw a diagonal crack
						vector.StrokeLine(screen, sx+6, sy+6, sx+16, sy+16, 1.2, veinClr, false)
						vector.StrokeLine(screen, sx+16, sy+16, sx+12, sy+24, 1.0, veinClr, false)
					}
					if h%5 == 0 {
						// Subtle second crack/vein
						veinClr := color.RGBA{0, 210, 240, 60}
						vector.StrokeLine(screen, sx+float32(config.TileSize)-8, sy+8, sx+float32(config.TileSize)-16, sy+20, 1.0, veinClr, false)
					}
				}
			}
		}
	}
}

func (c *ShockKelpCave) GenerateEntities(seed int64) []entity.CaveEntity {
	grid := c.Grid
	r := rand.New(rand.NewSource(seed))
	var entities []entity.CaveEntity

	gridW := len(grid)
	gridH := len(grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-2; ty++ {
			if grid[tx][ty] {
				continue
			}

			// 1. Floor-anchored kelp (60% chance on solid bottom tile)
			if grid[tx][ty+1] && r.Float64() < 0.60 {
				height := 24.0 + r.Float64()*28.0
				entities = append(entities, entity.NewShockKelp(
					float64(tx*config.TileSize)+float64(config.TileSize-16)/2.0,
					float64(ty*config.TileSize)+float64(config.TileSize)-height,
					height,
					"floor",
				))
			}

			// 2. Left-wall anchored kelp (45% chance on solid left tile)
			if grid[tx-1][ty] && r.Float64() < 0.45 {
				height := 24.0 + r.Float64()*28.0
				entities = append(entities, entity.NewShockKelp(
					float64(tx*config.TileSize),
					float64(ty*config.TileSize)+float64(config.TileSize)/2.0-height,
					height,
					"left",
				))
			}

			// 3. Right-wall anchored kelp (45% chance on solid right tile)
			if grid[tx+1][ty] && r.Float64() < 0.45 {
				height := 24.0 + r.Float64()*28.0
				entities = append(entities, entity.NewShockKelp(
					float64(tx*config.TileSize)+float64(config.TileSize-28.0),
					float64(ty*config.TileSize)+float64(config.TileSize)/2.0-height,
					height,
					"right",
				))
			}

			// 4. VoltaicLurker (8% chance on solid faces)
			var lurkerFaces []string
			if grid[tx-1][ty] {
				lurkerFaces = append(lurkerFaces, "left")
			}
			if grid[tx+1][ty] {
				lurkerFaces = append(lurkerFaces, "right")
			}
			if grid[tx][ty-1] {
				lurkerFaces = append(lurkerFaces, "top")
			}
			if grid[tx][ty+1] {
				lurkerFaces = append(lurkerFaces, "bottom")
			}

			if len(lurkerFaces) > 0 && r.Float64() < 0.08 {
				face := lurkerFaces[r.Intn(len(lurkerFaces))]
				entities = append(entities, entity.NewVoltaicLurker(
					float64(tx*config.TileSize),
					float64(ty*config.TileSize),
					face,
				))
			}
		}
	}

	// Spawn 0-2 (random) ElectroWeavers in the shock kelp cave
	var weaverCandidates []gvec.Vec2
	for tx := 2; tx < gridW-2; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			if !grid[tx][ty] {
				weaverCandidates = append(weaverCandidates, gvec.Vec2{X: float64(tx), Y: float64(ty)})
			}
		}
	}

	if len(weaverCandidates) > 0 {
		numToSpawn := r.Intn(3) // 0, 1, or 2
		if numToSpawn > len(weaverCandidates) {
			numToSpawn = len(weaverCandidates)
		}
		for i := 0; i < numToSpawn; i++ {
			idx := r.Intn(len(weaverCandidates))
			c1 := weaverCandidates[idx]
			weaverCandidates = append(weaverCandidates[:idx], weaverCandidates[idx+1:]...)

			entities = append(entities, &entity.ElectroWeaver{
				BaseEntity: entity.BaseEntity{
					Pos: gvec.Vec2{
						X: c1.X*float64(config.TileSize) + (float64(config.TileSize)-40.0)/2.0,
						Y: c1.Y*float64(config.TileSize) + (float64(config.TileSize)-20.0)/2.0,
					},
					Dimensions: gvec.Vec2{X: 40, Y: 20},
					Active:     true,
				},
			})
		}
	}

	return entities
}

func (c *ShockKelpCave) GenerateResources(seed int64) []resource.Resource {
	grid := c.Grid
	r := rand.New(rand.NewSource(seed))
	var nodes []resource.Resource

	gridW := len(grid)
	gridH := len(grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			if grid[tx][ty] {
				continue
			}

			// Gather possible attachment wall directions
			var possibleDirs []resource.AttachDirection
			if grid[tx][ty-1] {
				possibleDirs = append(possibleDirs, resource.AttachTop)
			}
			if grid[tx][ty+1] {
				possibleDirs = append(possibleDirs, resource.AttachBottom)
			}
			if grid[tx-1][ty] {
				possibleDirs = append(possibleDirs, resource.AttachLeft)
			}
			if grid[tx+1][ty] {
				possibleDirs = append(possibleDirs, resource.AttachRight)
			}

			// High resource density check (8.5% spawn chance near walls)
			if len(possibleDirs) > 0 && r.Float64() < 0.085 {
				attachDir := possibleDirs[r.Intn(len(possibleDirs))]
				roll := r.Float64()

				var node resource.Resource
				if roll < 0.45 {
					// 45% Copper Node
					node = resource.NewCopperNode(tx, ty)
				} else if roll < 0.85 {
					// 40% Quartz Node
					node = resource.NewQuartzNode(tx, ty)
				} else {
					// 15% Titanium Node
					node = resource.NewTitaniumNode(tx, ty)
				}

				node.SetAttachDir(attachDir)
				nodes = append(nodes, node)
			}
		}
	}

	return nodes
}

func (c *ShockKelpCave) GetAmbientColor(lightMult float64) []float32 {
	return []float32{0.03, 0.02, 0.06, 0.68}
}

// GenerateShockKelpCaveGrid generates a shallow, narrow maze-like 60x60 grid.
func GenerateShockKelpCaveGrid(r *rand.Rand) [][]bool {
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

	// Helper to carve a 1-tile brush (exact position)
	carve := func(cx, cy int) {
		if cx > 0 && cx < w-1 && cy > 0 && cy < h-1 {
			grid[cx][cy] = false
		}
	}

	// Helper to carve a small pocket room (2x2 or 3x3)
	carvePocket := func(cx, cy, sz int) {
		for dx := -sz; dx <= sz; dx++ {
			for dy := -sz; dy <= sz; dy++ {
				nx, ny := cx+dx, cy+dy
				if nx > 0 && nx < w-1 && ny > 0 && ny < h-1 {
					grid[nx][ny] = false
				}
			}
		}
	}

	// Run multiple narrow walkers to carve winding vertical crevices
	const numWalkers = 4
	for i := 0; i < numWalkers; i++ {
		// Distribute initial horizontal positions slightly around the middle
		cx := w/2 + r.Intn(10) - 5
		cy := 2

		// Walk until reaching near the bottom border
		for cy < h-2 {
			roll := r.Float64()
			switch {
			case roll < 0.55:
				cy++ // Down (biased)
			case roll < 0.77:
				cx-- // Left
			case roll < 0.95:
				cx++ // Right
			default:
				cy-- // Up
			}

			// Clamp inside boundaries
			if cx < 2 {
				cx = 2
			}
			if cx > w-3 {
				cx = w - 3
			}
			if cy < 2 {
				cy = 2
			}

			// Carve exact coordinate (brush size = 0, creating 1-tile wide paths)
			carve(cx, cy)

			// Occasional pocket chamber (7% chance per step)
			if r.Float64() < 0.07 {
				carvePocket(cx, cy, r.Intn(2)+1) // 1x1 or 2x2 extension
			}
		}
	}

	// Ensure entrance at top center is fully open and cleared for spawn safety
	for y := 0; y < 5; y++ {
		for x := (w / 2) - 3; x <= (w/2)+3; x++ {
			if x > 0 && x < w-1 && y < h-1 {
				grid[x][y] = false
			}
		}
	}

	return grid
}

// GenerateShockKelpReefTexture generates the slate-grey stone reef texture with electric purple veins
// and glowing cyan spores, centered around a dark entrance, for use in the overworld view.
func GenerateShockKelpReefTexture() *ebiten.Image {
	img := ebiten.NewImage(config.TileSize, config.TileSize)
	cx := float32(config.TileSize) / 2.0
	cy := float32(config.TileSize) / 2.0

	// Draw grey stone reef circle
	vector.FillCircle(img, cx, cy, 28, color.RGBA{60, 65, 75, 255}, false)
	vector.StrokeCircle(img, cx, cy, 28, 2.0, color.RGBA{95, 100, 110, 255}, false)

	// Draw stony bumps
	for angle := 0.0; angle < 2.0*math.Pi; angle += math.Pi / 4.0 {
		bx := cx + float32(math.Cos(angle))*20.0
		by := cy + float32(math.Sin(angle))*20.0
		vector.FillCircle(img, bx, by, 6.0, color.RGBA{75, 80, 90, 255}, false)
		vector.StrokeCircle(img, bx, by, 6.0, 1.0, color.RGBA{105, 110, 120, 255}, false)
	}

	// Draw electric purple swirls/lines
	vector.StrokeLine(img, cx-15, cy-15, cx+15, cy+15, 2.0, color.RGBA{140, 50, 210, 255}, false)
	vector.StrokeLine(img, cx+15, cy-15, cx-15, cy+15, 2.0, color.RGBA{140, 50, 210, 255}, false)

	// Draw cyan glowing spots (kelp spores)
	vector.FillCircle(img, cx-12, cy-10, 3.0, color.RGBA{0, 230, 255, 255}, false)
	vector.FillCircle(img, cx+12, cy+10, 3.0, color.RGBA{0, 230, 255, 255}, false)
	vector.FillCircle(img, cx+10, cy-12, 2.5, color.RGBA{0, 230, 255, 255}, false)
	vector.FillCircle(img, cx-10, cy+12, 2.5, color.RGBA{0, 230, 255, 255}, false)

	// Draw cave entrance in the center
	vector.FillCircle(img, cx, cy, 8.0, color.RGBA{15, 15, 18, 255}, false)
	vector.StrokeCircle(img, cx, cy, 8.0, 1.5, color.RGBA{140, 50, 210, 255}, false)

	return img
}
