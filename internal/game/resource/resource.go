package resource

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/assets"
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
	GetRecipeResultName() string
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

func (b *BaseResourceNode) GetRecipeResultName() string {
	return ""
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

// LoadAssets preloads and chroma-keys all resource crystal sprites.
func LoadAssets() {
	img, _, err := image.Decode(bytes.NewReader(assets.OreSheetPNG))
	if err != nil {
		log.Printf("Error: Failed to decode ore sheet: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Chroma-key green pixels using fast direct byte slice manipulation
	for i := 0; i < len(rgba.Pix); i += 4 {
		ru := rgba.Pix[i]
		gu := rgba.Pix[i+1]
		bu := rgba.Pix[i+2]

		if gu > 140 && ru < 100 && bu < 100 {
			rgba.Pix[i] = 0
			rgba.Pix[i+1] = 0
			rgba.Pix[i+2] = 0
			rgba.Pix[i+3] = 0
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
	spritesLoaded = true
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

func (n *TitaniumNode) GetName() string        { return "Titanium" }
func (n *TitaniumNode) GetMaxStack() int       { return 10 }
func (n *TitaniumNode) RequiresMech() bool     { return false }
func (n *TitaniumNode) GetBaseItem() item.Item { return &item.Titanium{} }
func (n *TitaniumNode) GetColor() color.Color  { return color.RGBA{168, 178, 188, 255} }
func (n *TitaniumNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawNodeIconSprite(screen, cx, cy, size, TitaniumSprite) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, n.GetColor(), false)
}

func (n *TitaniumNode) Draw(screen *ebiten.Image, camX, camY float64) {
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

func (n *CopperNode) GetName() string        { return "Copper" }
func (n *CopperNode) GetMaxStack() int       { return 10 }
func (n *CopperNode) RequiresMech() bool     { return false }
func (n *CopperNode) GetBaseItem() item.Item { return &item.Copper{} }
func (n *CopperNode) GetColor() color.Color  { return color.RGBA{218, 118, 48, 255} }
func (n *CopperNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawNodeIconSprite(screen, cx, cy, size, CopperSprite) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, n.GetColor(), false)
}

func (n *CopperNode) Draw(screen *ebiten.Image, camX, camY float64) {
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

func (n *QuartzNode) GetName() string        { return "Quartz" }
func (n *QuartzNode) GetMaxStack() int       { return 10 }
func (n *QuartzNode) RequiresMech() bool     { return false }
func (n *QuartzNode) GetBaseItem() item.Item { return &item.Quartz{} }
func (n *QuartzNode) GetColor() color.Color  { return color.RGBA{48, 218, 245, 255} }
func (n *QuartzNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawNodeIconSprite(screen, cx, cy, size, QuartzSprite) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, n.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
}

func (n *QuartzNode) Draw(screen *ebiten.Image, camX, camY float64) {
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

func (n *AbyssalOreNode) GetName() string        { return "Abyssal Ore" }
func (n *AbyssalOreNode) GetMaxStack() int       { return 10 }
func (n *AbyssalOreNode) RequiresMech() bool     { return true }
func (n *AbyssalOreNode) GetBaseItem() item.Item { return &item.AbyssalOre{} }
func (n *AbyssalOreNode) GetColor() color.Color  { return color.RGBA{148, 48, 218, 255} }
func (n *AbyssalOreNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawNodeIconSprite(screen, cx, cy, size, AbyssalSprite) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, n.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
}

func (n *AbyssalOreNode) Draw(screen *ebiten.Image, camX, camY float64) {
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
// ScrapMetalNode
// ---------------------------------------------------------

type ScrapMetalNode struct {
	BaseResourceNode
}

func (n *ScrapMetalNode) GetName() string        { return "Scrap Metal" }
func (n *ScrapMetalNode) GetMaxStack() int       { return 10 }
func (n *ScrapMetalNode) RequiresMech() bool     { return false }
func (n *ScrapMetalNode) GetBaseItem() item.Item { return &item.ScrapMetal{} }
func (n *ScrapMetalNode) GetColor() color.Color  { return color.RGBA{140, 110, 95, 255} }
func (n *ScrapMetalNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	(&item.ScrapMetal{}).DrawIcon(screen, cx, cy, size)
}

func (n *ScrapMetalNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0
	size := float32(14.0) * (float32(n.HitsToMine) / 3.0)
	if size < 4.0 {
		size = 4.0
	}
	// Scrap crate: brown/rust box with dark gray outline
	vector.FillRect(screen, cx-size, cy-size, size*2.0, size*2.0, color.RGBA{130, 95, 75, 255}, false)
	vector.StrokeRect(screen, cx-size, cy-size, size*2.0, size*2.0, 1.5, color.RGBA{90, 80, 75, 255}, false)
	vector.StrokeLine(screen, cx-size, cy-size, cx+size, cy+size, 1.5, color.RGBA{90, 80, 75, 255}, false)

	drawCracks(screen, float32(sx), float32(sy), n.HitsToMine)
}

func NewScrapMetalNode(tx, ty int) *ScrapMetalNode {
	return &ScrapMetalNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// ---------------------------------------------------------
// ElectronicWasteNode
// ---------------------------------------------------------

type ElectronicWasteNode struct {
	BaseResourceNode
}

func (n *ElectronicWasteNode) GetName() string        { return "Electronic Waste" }
func (n *ElectronicWasteNode) GetMaxStack() int       { return 10 }
func (n *ElectronicWasteNode) RequiresMech() bool     { return false }
func (n *ElectronicWasteNode) GetBaseItem() item.Item { return &item.ElectronicWaste{} }
func (n *ElectronicWasteNode) GetColor() color.Color  { return color.RGBA{70, 130, 90, 255} }
func (n *ElectronicWasteNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	(&item.ElectronicWaste{}).DrawIcon(screen, cx, cy, size)
}

func (n *ElectronicWasteNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0
	size := float32(14.0) * (float32(n.HitsToMine) / 3.0)
	if size < 4.0 {
		size = 4.0
	}
	// Electronic green container crate
	vector.FillRect(screen, cx-size, cy-size, size*2.0, size*2.0, color.RGBA{50, 110, 70, 255}, false)
	vector.StrokeRect(screen, cx-size, cy-size, size*2.0, size*2.0, 1.5, color.RGBA{110, 190, 130, 255}, false)
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, color.RGBA{30, 30, 30, 255}, false)

	drawCracks(screen, float32(sx), float32(sy), n.HitsToMine)
}

func NewElectronicWasteNode(tx, ty int) *ElectronicWasteNode {
	return &ElectronicWasteNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// ---------------------------------------------------------
// World Generation Configuration
// ---------------------------------------------------------

// ResourceTier defines configuration for resource spawning at a specific depth range.
type ResourceTier struct {
	MaxDepth         int     // The maximum depth (exclusive threshold, e.g. ty < MaxDepth)
	SpawnChance      float64 // The density/chance of spawning a resource on an exposed rock tile
	TitaniumWeight   float64 // Relative weight for Titanium spawning
	CopperWeight     float64 // Relative weight for Copper spawning
	QuartzWeight     float64 // Relative weight for Quartz spawning
	AbyssalOreWeight float64 // Relative weight for Abyssal Ore spawning
}

// ResourceGenConfig holds the configuration parameters for resource generation.
type ResourceGenConfig struct {
	FallbackSpawnChance float64        // Fallback spawn chance if no tier matches
	BaseHitsToMine      int            // Base health/hits to mine a node
	HitsDepthScale      int            // Scaling factor for health: depth / HitsDepthScale
	Tiers               []ResourceTier // List of depth-based configuration tiers (ordered by MaxDepth ascending)
	WreckageSpawnChance float64        // Spawn chance for wreckage resources on wreckage floor tiles
	ScrapMetalWeight    float64        // Relative weight for Scrap Metal in wreckage
	ElecWasteWeight     float64        // Relative weight for Electronic Waste in wreckage
}

// DefaultGenConfig represents the default generation settings matching the original game balance.
var DefaultGenConfig = ResourceGenConfig{
	FallbackSpawnChance: 0.05,
	BaseHitsToMine:      3,
	HitsDepthScale:      30,
	Tiers: []ResourceTier{
		{
			MaxDepth:         30,
			SpawnChance:      0.04,
			TitaniumWeight:   70.0,
			CopperWeight:     30.0,
			QuartzWeight:     0.0,
			AbyssalOreWeight: 0.0,
		},
		{
			MaxDepth:         60,
			SpawnChance:      0.055,
			TitaniumWeight:   40.0,
			CopperWeight:     40.0,
			QuartzWeight:     20.0,
			AbyssalOreWeight: 0.0,
		},
		{
			MaxDepth:         90,
			SpawnChance:      0.07,
			TitaniumWeight:   30.0,
			CopperWeight:     30.0,
			QuartzWeight:     30.0,
			AbyssalOreWeight: 10.0,
		},
		{
			MaxDepth:         999999, // Catch-all for super deep zones
			SpawnChance:      0.085,
			TitaniumWeight:   20.0,
			CopperWeight:     20.0,
			QuartzWeight:     35.0,
			AbyssalOreWeight: 25.0,
		},
	},
	WreckageSpawnChance: 0.08,
	ScrapMetalWeight:    65.0,
	ElecWasteWeight:     35.0,
}

// GenConfig is the active resource generation configuration.
// It can be adjusted at runtime to easily change spawning behavior.
var GenConfig = DefaultGenConfig

// ---------------------------------------------------------
// World Generation Functions
// ---------------------------------------------------------

// BlueprintNode represents a blueprint node containing a recipe.
type BlueprintNode struct {
	BaseResourceNode
	RecipeResultName string
}

func (n *BlueprintNode) GetName() string        { return "Blueprint: " + n.RecipeResultName }
func (n *BlueprintNode) GetMaxStack() int       { return 1 }
func (n *BlueprintNode) RequiresMech() bool     { return false }
func (n *BlueprintNode) GetBaseItem() item.Item { return nil } // immediately unlocks, no item in inventory
func (n *BlueprintNode) GetColor() color.Color  { return color.RGBA{0, 180, 255, 255} }
func (n *BlueprintNode) GetRecipeResultName() string { return n.RecipeResultName }

func (n *BlueprintNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	// Simple blueprint icon: blue circle with outline
	vector.FillCircle(screen, cx, cy, size/2.0, n.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
}

func (n *BlueprintNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0
	
	size := float32(14.0) * (float32(n.HitsToMine) / 3.0)
	if size < 4.0 {
		size = 4.0
	}
	
	// Blueprint backing sheet
	vector.FillRect(screen, cx-size, cy-size, size*2.0, size*2.0, color.RGBA{10, 40, 90, 255}, false)
	vector.StrokeRect(screen, cx-size, cy-size, size*2.0, size*2.0, 1.5, color.RGBA{0, 160, 255, 255}, false)
	
	// Blueprint layout details
	vector.StrokeLine(screen, cx-size+4, cy-size+4, cx+size-4, cy-size+4, 1.0, color.RGBA{0, 120, 220, 180}, false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{0, 180, 255, 180}, false)
	vector.StrokeLine(screen, cx-size/2.0, cy+size/2.0, cx+size/2.0, cy+size/2.0, 1.0, color.RGBA{0, 120, 220, 180}, false)

	drawCracks(screen, sx, sy, n.HitsToMine)
}

func NewBlueprintNode(tx, ty int, recipeResultName string) *BlueprintNode {
	return &BlueprintNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		RecipeResultName: recipeResultName,
	}
}

// Helper to shuffle a slice of strings using r rand.Rand
func shuffleRecipes(slice []string, r *rand.Rand) []string {
	shuffled := make([]string, len(slice))
	copy(shuffled, slice)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

// GenerateWreckageResources spawns scrap metal and electronic waste nodes on room floors in wreckage caves,
// and also spawns appropriate recipe blueprints depending on the shipIndex (0, 1, or 2).
func GenerateWreckageResources(grid [][]bool, seed int64, shipIndex int) []Resource {
	nodes := []Resource{}
	if grid == nil {
		return nodes
	}
	gridW := len(grid)
	gridH := len(grid[0])
	r := rand.New(rand.NewSource(seed))

	// Find room floor tiles (open space above a solid tile, not in the central elevator shaft)
	var upperFloors [][2]int
	var lowerFloors [][2]int

	for tx := 1; tx < gridW-1; tx++ {
		// Central elevator shaft is tx 27..32
		if tx >= 27 && tx <= 32 {
			continue
		}
		for ty := 1; ty < gridH-2; ty++ {
			if !grid[tx][ty] { // open tile
				if grid[tx][ty+1] { // solid tile below (floor)
					if ty <= 51 {
						upperFloors = append(upperFloors, [2]int{tx, ty})
					} else {
						lowerFloors = append(lowerFloors, [2]int{tx, ty})
					}

					if r.Float64() < GenConfig.WreckageSpawnChance {
						var node Resource
						totalW := GenConfig.ScrapMetalWeight + GenConfig.ElecWasteWeight
						var isScrap bool
						if totalW > 0 {
							isScrap = r.Float64()*totalW < GenConfig.ScrapMetalWeight
						} else {
							isScrap = true
						}

						if isScrap {
							node = NewScrapMetalNode(tx, ty)
						} else {
							node = NewElectronicWasteNode(tx, ty)
						}
						// Scale hits with depth
						node.SetHitsToMine(GenConfig.BaseHitsToMine + (ty / GenConfig.HitsDepthScale))
						nodes = append(nodes, node)
					}
				}
			}
		}
	}

	// Spawn Blueprints
	t1Recipes := []string{
		"Ultra High Capacity O2 Tank",
		"Scout Sub Kit",
		"Solar Array MKII Module",
		"Storage Vault MKII Module",
		"Sonar Amplifier",
		"Thermal Generator",
	}
	t2Recipes := []string{
		"Heavy Mech Kit",
		"Escape Rocket",
	}

	var selected []string
	if shipIndex == 0 {
		shuffled := shuffleRecipes(t1Recipes, r)
		numToSpawn := 3 + r.Intn(2) // 3 or 4
		if numToSpawn > len(shuffled) {
			numToSpawn = len(shuffled)
		}
		selected = shuffled[:numToSpawn]
	} else if shipIndex == 1 {
		allRecipes := append([]string{}, t1Recipes...)
		allRecipes = append(allRecipes, t2Recipes...)
		shuffled := shuffleRecipes(allRecipes, r)
		numToSpawn := 4 + r.Intn(2) // 4 or 5
		if numToSpawn > len(shuffled) {
			numToSpawn = len(shuffled)
		}
		selected = shuffled[:numToSpawn]
	} else if shipIndex == 2 {
		selected = append([]string{}, t2Recipes...)
	}

	// Helper to check if a tile is already occupied by a spawned node
	isOccupied := func(tx, ty int) bool {
		for _, n := range nodes {
			ntx, nty := n.GetTilePos()
			if ntx == tx && nty == ty {
				return true
			}
		}
		return false
	}

	for _, recipeName := range selected {
		// Determine tier
		isTier2 := false
		for _, name := range t2Recipes {
			if name == recipeName {
				isTier2 = true
				break
			}
		}

		var floorList *[][2]int
		if isTier2 {
			floorList = &lowerFloors
		} else {
			floorList = &upperFloors
		}

		if len(*floorList) > 0 {
			// Find a non-occupied random floor tile
			shuffledIndices := r.Perm(len(*floorList))
			var chosenTile [2]int
			found := false
			for _, idx := range shuffledIndices {
				tile := (*floorList)[idx]
				if !isOccupied(tile[0], tile[1]) {
					chosenTile = tile
					found = true
					// Remove the chosen tile from the list to avoid duplicate blueprint placement
					*floorList = append((*floorList)[:idx], (*floorList)[idx+1:]...)
					break
				}
			}

			if found {
				bpNode := NewBlueprintNode(chosenTile[0], chosenTile[1], recipeName)
				nodes = append(nodes, bpNode)
			}
		}
	}

	return nodes
}

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
					kind := kindTitanium
					var spawnChance = GenConfig.FallbackSpawnChance

					// Find the matching tier based on depth ty
					var activeTier *ResourceTier
					for i := range GenConfig.Tiers {
						if ty < GenConfig.Tiers[i].MaxDepth {
							activeTier = &GenConfig.Tiers[i]
							break
						}
					}

					// TODO: fix below!
					if activeTier != nil {
						spawnChance = activeTier.SpawnChance
						totalWeight := activeTier.TitaniumWeight + activeTier.CopperWeight + activeTier.QuartzWeight + activeTier.AbyssalOreWeight
						if totalWeight > 0 {
							roll := r.Float64() * totalWeight
							if roll < activeTier.TitaniumWeight {
								kind = kindTitanium
							} else if roll < activeTier.TitaniumWeight+activeTier.CopperWeight {
								kind = kindCopper
							} else if roll < activeTier.TitaniumWeight+activeTier.CopperWeight+activeTier.QuartzWeight {
								kind = kindQuartz
							} else {
								kind = kindAbyssalOre
							}
						}
					}

					if spawnRoll < spawnChance {
						var node Resource
						switch kind {
						case kindTitanium:
							node = NewTitaniumNode(tx, ty)
						case kindCopper:
							node = NewCopperNode(tx, ty)
						case kindQuartz:
							node = NewQuartzNode(tx, ty)
						case kindAbyssalOre:
							node = NewAbyssalOreNode(tx, ty)
						}
						// Scale node hits (health) with depth: base + depth / scale
						node.SetHitsToMine(GenConfig.BaseHitsToMine + (ty / GenConfig.HitsDepthScale))
						nodes = append(nodes, node)
					}
				}
			}
		}
	}

	return nodes
}
