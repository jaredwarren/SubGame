package scene

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/world"
)

// DiverDrawWidth defines the targeted width of the diver sprite on screen.
var DiverDrawWidth = 36.0

var (
	DiverGridCols     = 8
	DiverGridRows     = 4
	DiverMineGridCols = 6
)

var (
	DiverIdleCols   = []int{0, 1, 2, 3}
	DiverSwimCols   = []int{0, 1, 2, 3, 4, 5, 6, 7}
	DiverMineCols   = []int{0, 1, 2, 3}
	DiverDamageCols = []int{0}
)

var (
	DiverXOffset = 0
	DiverYOffset = 0
)

// CaveScene manages the side-view cave swimming controls, collision, and rendering.
type CaveScene struct {
	ActiveCave cave.Cave
	CaveGrid   [][]bool
	Nodes      []resource.Resource
	Entities   []entity.CaveEntity
	IsShallow  bool

	shaderOpts    ebiten.DrawRectShaderOptions
	Uniforms      map[string]any
	lightSource   []float32
	flashlightDir []float32
	sonarSource   []float32
	entranceLight []float32

	offscreen *ebiten.Image

	diverSheet       *ebiten.Image
	diverIdleFrames  []*ebiten.Image
	diverSwimFrames  []*ebiten.Image
	diverMineFrames  []*ebiten.Image
	diverDamageFrame *ebiten.Image

	// Scroll transition fields
	scrollActive bool
	scrollTimer  int
	scrollDir    int // -1 for left, 1 for right
	oldCave      cave.Cave
	newCave      cave.Cave
	oldCaveGrid  [][]bool
	newCaveGrid  [][]bool
	oldNodes     []resource.Resource
	newNodes     []resource.Resource
	oldEntities  []entity.CaveEntity
	newEntities  []entity.CaveEntity
	oldTrenchX   int
	oldTrenchY   int
	newTrenchX   int
	newTrenchY   int
	oldTrenchKey string
	newTrenchKey string
	oldCamX      float64
	oldCamY      float64
	newCamX      float64
	newCamY      float64
	offscreenOld *ebiten.Image
	offscreenNew *ebiten.Image
}

// NewCaveScene creates a new CaveScene instance.
func NewCaveScene() *CaveScene {
	cs := &CaveScene{
		Nodes:         []resource.Resource{},
		Entities:      []entity.CaveEntity{},
		Uniforms:      make(map[string]any),
		lightSource:   make([]float32, 2),
		flashlightDir: make([]float32, 2),
		sonarSource:   make([]float32, 2),
		entranceLight: make([]float32, 2),
	}
	cs.shaderOpts.Uniforms = cs.Uniforms
	cs.loadDiverSheet()
	return cs
}

func (c *CaveScene) loadDiverSheet() {
	paths := []string{
		"assets/textures/diver_sheet.png",
		"/Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/textures/diver_sheet.png",
		"../../assets/textures/diver_sheet.png",
		"../assets/textures/diver_sheet.png",
	}

	var file *os.File
	var err error
	for _, p := range paths {
		file, err = os.Open(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Printf("Warning: Failed to open assets/textures/diver_sheet.png: %v", err)
		return
	}
	defer func() { _ = file.Close() }()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Warning: Failed to decode assets/textures/diver_sheet.png: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			clr := img.At(x, y)
			r, g, b, a := clr.RGBA()
			ru := uint8(r >> 8)
			gu := uint8(g >> 8)
			bu := uint8(b >> 8)
			au := uint8(a >> 8)
			if gu > 140 && ru < 100 && bu < 100 {
				rgba.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				rgba.SetRGBA(x, y, color.RGBA{ru, gu, bu, au})
			}
		}
	}

	sheet := ebiten.NewImageFromImage(rgba)
	c.diverSheet = sheet

	frameW := bounds.Dx() / DiverGridCols
	frameH := bounds.Dy() / DiverGridRows

	for _, col := range DiverIdleCols {
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH)
		c.diverIdleFrames = append(c.diverIdleFrames, sheet.SubImage(rect).(*ebiten.Image))
	}
	for _, col := range DiverSwimCols {
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset+frameH, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH*2)
		c.diverSwimFrames = append(c.diverSwimFrames, sheet.SubImage(rect).(*ebiten.Image))
	}
	mineFrameW := bounds.Dx() / DiverMineGridCols
	for _, col := range DiverMineCols {
		rect := image.Rect(DiverXOffset+col*mineFrameW, DiverYOffset+frameH*2, DiverXOffset+(col+1)*mineFrameW, DiverYOffset+frameH*3)
		c.diverMineFrames = append(c.diverMineFrames, sheet.SubImage(rect).(*ebiten.Image))
	}
	if len(DiverDamageCols) > 0 {
		col := DiverDamageCols[0]
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset+frameH*3, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH*4)
		c.diverDamageFrame = sheet.SubImage(rect).(*ebiten.Image)
	}
}

func (c *CaveScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateCave)
}

func (c *CaveScene) OnExit(g GameContext) {}

func (c *CaveScene) checkCollisions(g GameContext, p *player.Player) {
	newX := p.Pos.X + p.Vel.X
	if c.IsSolid(g, newX, p.Pos.Y, p.Width, p.Height) {
		p.Vel.X = 0
	} else {
		p.Pos.X = newX
	}
	newY := p.Pos.Y + p.Vel.Y
	if c.IsSolid(g, p.Pos.X, newY, p.Width, p.Height) {
		p.Vel.Y = 0
	} else {
		p.Pos.Y = newY
	}
}

// IsSolid checks if the proposed bounding box overlaps with solid cave tiles.
func (c *CaveScene) IsSolid(g GameContext, x, y, w, h float64) bool {
	if c.CaveGrid == nil {
		return false
	}
	gridW := len(c.CaveGrid)
	gridH := len(c.CaveGrid[0])

	x1 := int(math.Floor(x)) / config.TileSize
	x2 := int(math.Floor(x+w)) / config.TileSize
	y1 := int(math.Floor(y)) / config.TileSize
	y2 := int(math.Floor(y+h)) / config.TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= gridW {
				if c.IsShallow {
					worldObj := g.GetWorld()
					currentTx, currentTy := g.GetActiveTrenchCoords()
					var neighborTx int
					if tx < 0 {
						neighborTx = currentTx - 1
					} else {
						neighborTx = currentTx + 1
					}
					if neighborTx >= 0 && neighborTx < worldObj.Width && worldObj.OverworldMap[neighborTx][currentTy] == world.TileWater {
						continue
					}
				}
				return true
			}
			if ty < 0 {
				continue
			}
			if ty >= gridH {
				return true
			}
			if c.CaveGrid[tx][ty] {
				return true
			}
		}
	}
	return false
}

func getSkyColor(timeOfDay float64) color.RGBA {
	nightSky := [3]float64{20, 30, 70}
	dawnSky := [3]float64{255, 160, 80}
	daySky := [3]float64{140, 200, 255}
	duskSky := [3]float64{220, 100, 60}

	lerpF := func(a, b [3]float64, t float64) color.RGBA {
		return color.RGBA{
			R: uint8(a[0] + (b[0]-a[0])*t),
			G: uint8(a[1] + (b[1]-a[1])*t),
			B: uint8(a[2] + (b[2]-a[2])*t),
			A: 255,
		}
	}

	switch {
	case timeOfDay < 1200:
		return lerpF(nightSky, dawnSky, timeOfDay/1200.0)
	case timeOfDay < 2400:
		return lerpF(dawnSky, daySky, (timeOfDay-1200.0)/1200.0)
	case timeOfDay < 8400:
		return lerpF(daySky, daySky, 0)
	case timeOfDay < 9600:
		return lerpF(daySky, duskSky, (timeOfDay-8400.0)/1200.0)
	case timeOfDay < 10800:
		return lerpF(duskSky, nightSky, (timeOfDay-9600.0)/1200.0)
	default:
		return lerpF(nightSky, nightSky, 0)
	}
}

func (c *CaveScene) getAmbientColor(isShallow bool, timeOfDay float64) []float32 {
	if config.LightCaveForDebug {
		return []float32{0.02, 0.02, 0.03, 0.15}
	}
	if isShallow {
		mult := GetOverworldLightMultiplier(timeOfDay)
		alpha := float32(0.75 - (mult-0.2)/0.8*0.60)
		return []float32{0.04, 0.06, 0.12, alpha}
	}
	return []float32{0.01, 0.01, 0.03, 0.97}
}

func max0(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min0(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *CaveScene) IsScrollActive() bool {
	return c.scrollActive
}
