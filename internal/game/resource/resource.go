package resource

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

const TileSize = 64

// Resource defines the interface that all mineable nodes/objects must implement.
type Resource interface {
	item.Item
	GetTilePos() (int, int)
	GetHitsToMine() int
	SetHitsToMine(hits int)
	RequiresMech() bool
	Draw(screen *ebiten.Image, camX, camY float64)
}

// BaseResourceNode holds the shared state for all resource node types.
type BaseResourceNode struct {
	Tx, Ty     int // Tile coordinates
	HitsToMine int
}

func (b *BaseResourceNode) GetTilePos() (int, int) {
	return b.Tx, b.Ty
}

func (b *BaseResourceNode) GetHitsToMine() int {
	return b.HitsToMine
}

func (b *BaseResourceNode) SetHitsToMine(hits int) {
	b.HitsToMine = hits
}

// drawNodeBase renders the shared backing block behind all resource nodes.
func drawNodeBase(screen *ebiten.Image, tx, ty int, camX, camY float64) (float32, float32) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))
	vector.FillRect(screen, sx, sy, TileSize, TileSize, color.RGBA{25, 22, 30, 255}, false)
	vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, color.RGBA{45, 40, 52, 255}, false)
	return sx, sy
}

// drawMineral renders the mineral crystal at the center of a node.
func drawMineral(screen *ebiten.Image, sx, sy float32, hitsToMine int, mineralColor, coreColor color.Color, hasExtraShard bool) {
	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0

	// Scale size based on hits left
	size := float32(14.0) * (float32(hitsToMine) / 3.0)
	if size < 4.0 {
		size = 4.0
	}

	// Draw raw minerals as overlapping angled rectangles (crystal facets)
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, mineralColor, false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, coreColor, false)

	// Additional crystal shard for premium minerals
	if hasExtraShard {
		vector.FillRect(screen, cx-size/3.0, cy-size/1.2, size/1.5, size/1.5, coreColor, false)
	}
}

var (
	TitaniumSprite *ebiten.Image
	CopperSprite   *ebiten.Image
	QuartzSprite   *ebiten.Image
	AbyssalSprite  *ebiten.Image
	spritesLoaded  bool
)

func loadSpritesLazy() {
	if spritesLoaded {
		return
	}
	spritesLoaded = true

	paths := []string{
		"assets/textures/ore_sheet.png",
		"/Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/textures/ore_sheet.png",
		"../../assets/textures/ore_sheet.png",
		"../assets/textures/ore_sheet.png",
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
		log.Printf("Warning: Failed to open assets/textures/ore_sheet.png: %v", err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Warning: Failed to decode assets/textures/ore_sheet.png: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	// Chroma keying: set green background to transparent (Alpha 0)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			clr := img.At(x, y)
			r, g, b, a := clr.RGBA()
			ru := uint8(r >> 8)
			gu := uint8(g >> 8)
			bu := uint8(b >> 8)
			au := uint8(a >> 8)

			// Green background keying: dominant green channel, low red and blue channels
			if gu > 140 && ru < 100 && bu < 100 {
				rgba.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				rgba.SetRGBA(x, y, color.RGBA{ru, gu, bu, au})
			}
		}
	}

	sheet := ebiten.NewImageFromImage(rgba)
	if sheet == nil {
		return
	}

	if bounds.Dx() == 3584 && bounds.Dy() == 1184 {
		// Specific coordinate slice for the user's high-res generated ore sheet
		TitaniumSprite = sheet.SubImage(image.Rect(832, 330, 1312, 856)).(*ebiten.Image)
		CopperSprite = sheet.SubImage(image.Rect(1327, 330, 1809, 856)).(*ebiten.Image)
		QuartzSprite = sheet.SubImage(image.Rect(1827, 330, 2308, 856)).(*ebiten.Image)
		AbyssalSprite = sheet.SubImage(image.Rect(2328, 330, 2790, 856)).(*ebiten.Image)
	} else {
		// General fallback: slice into 4 equal columns
		w := bounds.Dx() / 4
		h := bounds.Dy()
		TitaniumSprite = sheet.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
		CopperSprite = sheet.SubImage(image.Rect(w, 0, w*2, h)).(*ebiten.Image)
		QuartzSprite = sheet.SubImage(image.Rect(w*2, 0, w*3, h)).(*ebiten.Image)
		AbyssalSprite = sheet.SubImage(image.Rect(w*3, 0, w*4, h)).(*ebiten.Image)
	}
}

func drawCracks(screen *ebiten.Image, sx, sy float32, hitsToMine int) {
	if hitsToMine >= 3 {
		// No cracks at full health (3 hits remaining)
		return
	}

	// Crack color: dark charcoal/black to represent fracture lines overlaying the mineral
	crackColor := color.RGBA{15, 15, 20, 235}

	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0

	// Stage 1 (1 hit taken, 2 remaining): draw primary cracks radiating from center
	if hitsToMine <= 2 {
		vector.StrokeLine(screen, cx-14, cy-12, cx-3, cy-2, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx-3, cy-2, cx+12, cy-14, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx-3, cy-2, cx-6, cy+14, 3.0, crackColor, false)
	}

	// Stage 2 (2 hits taken, 1 remaining): add detailed, longer branching fractures
	if hitsToMine <= 1 {
		// Branches from Stage 1 cracks
		vector.StrokeLine(screen, cx-14, cy-12, cx-22, cy-8, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx+12, cy-14, cx+20, cy-20, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx-6, cy+14, cx-14, cy+22, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx-6, cy+14, cx+8, cy+12, 2.0, crackColor, false)

		// Second independent fracture cluster
		vector.StrokeLine(screen, cx+14, cy+4, cx+3, cy-8, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx+3, cy-8, cx-8, cy-5, 3.0, crackColor, false)
	}
}

func drawNodeSprite(screen *ebiten.Image, tx, ty int, camX, camY float64, sprite *ebiten.Image, hitsToMine int) bool {
	if sprite == nil {
		return false
	}
	sx := float64(tx*TileSize - int(camX))
	sy := float64(ty*TileSize - int(camY))

	// Draw the node background block under the sprite
	drawNodeBase(screen, tx, ty, camX, camY)

	op := &ebiten.DrawImageOptions{}

	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())

	// Scale the sprite to match the full TileSize (64x64) without shrinking
	baseScaleX := float64(TileSize) / spriteW
	baseScaleY := float64(TileSize) / spriteH

	// Center the sprite on the origin (0,0) before scaling
	op.GeoM.Translate(-spriteW/2.0, -spriteH/2.0)
	// Scale to full tile size
	op.GeoM.Scale(baseScaleX, baseScaleY)
	// Translate to screen tile coordinates + center offset
	op.GeoM.Translate(sx+float64(TileSize)/2.0, sy+float64(TileSize)/2.0)

	screen.DrawImage(sprite, op)

	// Draw overlay cracks representing node damage state
	drawCracks(screen, float32(sx), float32(sy), hitsToMine)

	return true
}

func drawNodeIconSprite(screen *ebiten.Image, cx, cy, size float32, sprite *ebiten.Image) bool {
	if sprite == nil {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())

	op.GeoM.Translate(-spriteW/2.0, -spriteH/2.0)
	op.GeoM.Scale(float64(size)/spriteW, float64(size)/spriteH)
	op.GeoM.Translate(float64(cx), float64(cy))

	screen.DrawImage(sprite, op)
	return true
}

// ---------------------------------------------------------
// TitaniumNode
// ---------------------------------------------------------

type TitaniumNode struct {
	BaseResourceNode
}

func (n *TitaniumNode) GetName() string    { return "Titanium" }
func (n *TitaniumNode) GetMaxStack() int   { return 10 }
func (n *TitaniumNode) RequiresMech() bool { return false }
func (n *TitaniumNode) GetColor() color.Color { return color.RGBA{168, 178, 188, 255} }
func (n *TitaniumNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	loadSpritesLazy()
	if drawNodeIconSprite(screen, cx, cy, size, TitaniumSprite) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, n.GetColor(), false)
}

func (n *TitaniumNode) Draw(screen *ebiten.Image, camX, camY float64) {
	loadSpritesLazy()
	if drawNodeSprite(screen, n.Tx, n.Ty, camX, camY, TitaniumSprite, n.HitsToMine) {
		return
	}
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{160, 175, 185, 255} // Metallic silver
	coreColor := color.RGBA{220, 230, 240, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, false)
}

// ---------------------------------------------------------
// CopperNode
// ---------------------------------------------------------

type CopperNode struct {
	BaseResourceNode
}

func (n *CopperNode) GetName() string    { return "Copper" }
func (n *CopperNode) GetMaxStack() int   { return 10 }
func (n *CopperNode) RequiresMech() bool { return false }
func (n *CopperNode) GetColor() color.Color { return color.RGBA{218, 118, 48, 255} }
func (n *CopperNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	loadSpritesLazy()
	if drawNodeIconSprite(screen, cx, cy, size, CopperSprite) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, n.GetColor(), false)
}

func (n *CopperNode) Draw(screen *ebiten.Image, camX, camY float64) {
	loadSpritesLazy()
	if drawNodeSprite(screen, n.Tx, n.Ty, camX, camY, CopperSprite, n.HitsToMine) {
		return
	}
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{210, 110, 45, 255} // Copper orange
	coreColor := color.RGBA{240, 160, 80, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, false)
}

// ---------------------------------------------------------
// QuartzNode
// ---------------------------------------------------------

type QuartzNode struct {
	BaseResourceNode
}

func (n *QuartzNode) GetName() string    { return "Quartz" }
func (n *QuartzNode) GetMaxStack() int   { return 10 }
func (n *QuartzNode) RequiresMech() bool { return false }
func (n *QuartzNode) GetColor() color.Color { return color.RGBA{48, 218, 245, 255} }
func (n *QuartzNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	loadSpritesLazy()
	if drawNodeIconSprite(screen, cx, cy, size, QuartzSprite) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, n.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
}

func (n *QuartzNode) Draw(screen *ebiten.Image, camX, camY float64) {
	loadSpritesLazy()
	if drawNodeSprite(screen, n.Tx, n.Ty, camX, camY, QuartzSprite, n.HitsToMine) {
		return
	}
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{40, 210, 245, 200} // Cyan bioluminescent quartz
	coreColor := color.RGBA{220, 250, 255, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, true)
}

// ---------------------------------------------------------
// AbyssalOreNode
// ---------------------------------------------------------

type AbyssalOreNode struct {
	BaseResourceNode
}

func (n *AbyssalOreNode) GetName() string    { return "Abyssal Ore" }
func (n *AbyssalOreNode) GetMaxStack() int   { return 10 }
func (n *AbyssalOreNode) RequiresMech() bool { return true }
func (n *AbyssalOreNode) GetColor() color.Color { return color.RGBA{148, 48, 218, 255} }
func (n *AbyssalOreNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	loadSpritesLazy()
	if drawNodeIconSprite(screen, cx, cy, size, AbyssalSprite) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, n.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
}

func (n *AbyssalOreNode) Draw(screen *ebiten.Image, camX, camY float64) {
	loadSpritesLazy()
	if drawNodeSprite(screen, n.Tx, n.Ty, camX, camY, AbyssalSprite, n.HitsToMine) {
		return
	}
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{140, 40, 210, 255} // Glowing purple abyssal ore
	coreColor := color.RGBA{230, 180, 255, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, true)
}

// ---------------------------------------------------------
// Constructor helpers
// ---------------------------------------------------------

// NewTitaniumNode creates a new titanium resource node.
func NewTitaniumNode(tx, ty int) *TitaniumNode {
	return &TitaniumNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// NewCopperNode creates a new copper resource node.
func NewCopperNode(tx, ty int) *CopperNode {
	return &CopperNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// NewQuartzNode creates a new quartz resource node.
func NewQuartzNode(tx, ty int) *QuartzNode {
	return &QuartzNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// NewAbyssalOreNode creates a new abyssal ore resource node.
func NewAbyssalOreNode(tx, ty int) *AbyssalOreNode {
	return &AbyssalOreNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// ---------------------------------------------------------
// World Generation
// ---------------------------------------------------------

// GenerateResourceNodes scans the cave tile grid and generates mineral nodes on exposed wall surfaces.
func GenerateResourceNodes(grid [][]bool, seed int64) []Resource {
	nodes := []Resource{}
	if grid == nil {
		return nodes
	}
	gridW := len(grid)
	gridH := len(grid[0])

	r := rand.New(rand.NewSource(seed))

	// Helper to check if a tile is adjacent to empty water (open path)
	isAdjacentToEmpty := func(tx, ty int) bool {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := tx+dx, ty+dy
				if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH {
					if !grid[nx][ny] {
						return true
					}
				}
			}
		}
		return false
	}

	// newNode creates the appropriate concrete resource node for a given depth.
	type nodeKind int
	const (
		kindTitanium nodeKind = iota
		kindCopper
		kindQuartz
		kindAbyssalOre
	)

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			// Place nodes inside solid rock walls
			if grid[tx][ty] {
				if isAdjacentToEmpty(tx, ty) {
					spawnRoll := r.Float64()
					var kind nodeKind
					var spawnChance = 0.05 // 5% chance per eligible tile face

					// Determine resource distribution based on depth
					if ty > 90 {
						// Abyssal depths: mostly Titanium/Quartz, but introduces Abyssal Ore
						roll := r.Float64()
						if roll < 0.18 {
							kind = kindAbyssalOre
							spawnChance = 0.035
						} else if roll < 0.45 {
							kind = kindQuartz
						} else if roll < 0.75 {
							kind = kindCopper
						} else {
							kind = kindTitanium
						}
					} else if ty > 45 {
						// Mid depths: Quartz and Copper common
						roll := r.Float64()
						if roll < 0.35 {
							kind = kindQuartz
						} else if roll < 0.70 {
							kind = kindCopper
						} else {
							kind = kindTitanium
						}
					} else {
						// Shallow depths: Titanium and Copper only
						roll := r.Float64()
						if roll < 0.28 {
							kind = kindCopper
						} else {
							kind = kindTitanium
						}
					}

					if spawnRoll < spawnChance {
						switch kind {
						case kindTitanium:
							nodes = append(nodes, NewTitaniumNode(tx, ty))
						case kindCopper:
							nodes = append(nodes, NewCopperNode(tx, ty))
						case kindQuartz:
							nodes = append(nodes, NewQuartzNode(tx, ty))
						case kindAbyssalOre:
							nodes = append(nodes, NewAbyssalOreNode(tx, ty))
						}
					}
				}
			}
		}
	}

	return nodes
}
