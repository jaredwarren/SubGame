package resource

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/assets"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

const TileSize = 64

type AttachDirection int

const (
	AttachNone AttachDirection = iota
	AttachTop
	AttachBottom
	AttachLeft
	AttachRight
)

// Resource defines the interface that all mineable nodes/objects must implement.
type Resource interface {
	item.Item
	GetTilePos() (int, int)
	GetHitsToMine() int
	SetHitsToMine(hits int)
	RequiresMech() bool
	Draw(screen *ebiten.Image, camX, camY float64)
	GetRecipeResultName() string
	SetAttachDir(dir AttachDirection)
	GetAttachDir() AttachDirection
}

// BaseResourceNode holds the shared state for all resource node types.
type BaseResourceNode struct {
	Tx, Ty     int // Tile coordinates
	HitsToMine int
	AttachDir  AttachDirection
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

func (b *BaseResourceNode) SetAttachDir(dir AttachDirection) {
	b.AttachDir = dir
}

func (b *BaseResourceNode) GetAttachDir() AttachDirection {
	return b.AttachDir
}

// drawNodeBase renders the shared backing block behind all resource nodes.
func drawNodeBase(screen *ebiten.Image, tx, ty int, camX, camY float64) (float32, float32) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))
	vector.FillRect(screen, sx, sy, TileSize, TileSize, color.RGBA{25, 22, 30, 255}, false)
	vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, color.RGBA{45, 40, 52, 255}, false)
	return sx, sy
}

var gPath vector.Path

// Helper to darken a color
func darkenColor(c color.Color, factor float32) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(float32(r>>8) * factor),
		G: uint8(float32(g>>8) * factor),
		B: uint8(float32(b>>8) * factor),
		A: uint8(a >> 8),
	}
}

// Helper to blend two colors
func blendColor(c1, c2 color.Color, t float32) color.RGBA {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return color.RGBA{
		R: uint8((1.0-t)*float32(r1>>8) + t*float32(r2>>8)),
		G: uint8((1.0-t)*float32(g1>>8) + t*float32(g2>>8)),
		B: uint8((1.0-t)*float32(b1>>8) + t*float32(b2>>8)),
		A: uint8((1.0-t)*float32(a1>>8) + t*float32(a2>>8)),
	}
}

// rotateVec rotates a 2D vector by an angle in radians
func rotateVec(v [2]float32, angle float32) [2]float32 {
	cosA := float32(math.Cos(float64(angle)))
	sinA := float32(math.Sin(float64(angle)))
	return [2]float32{
		v[0]*cosA - v[1]*sinA,
		v[0]*sinA + v[1]*cosA,
	}
}

// localToScreen transforms a local point (lx, ly) to screen space based on growth basis
func localToScreen(cx, cy float32, lx, ly float32, dirVec, perpVec [2]float32) (float32, float32) {
	return cx + lx*perpVec[0] + ly*dirVec[0], cy + lx*perpVec[1] + ly*dirVec[1]
}

// drawShard draws a single 3D crystal shard growing in local space
func drawShard(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, length, width float32, shadowColor, highlightColor color.Color) {
	// Left Face
	gPath.Reset()
	lx0, ly0 := localToScreen(cx, cy, 0, 0, dirVec, perpVec)
	lx1, ly1 := localToScreen(cx, cy, -width/2, 0, dirVec, perpVec)
	lx2, ly2 := localToScreen(cx, cy, -width/2, length*0.4, dirVec, perpVec)
	lx3, ly3 := localToScreen(cx, cy, 0, length, dirVec, perpVec)
	lx4, ly4 := localToScreen(cx, cy, 0, length*0.45, dirVec, perpVec)

	gPath.MoveTo(lx0, ly0)
	gPath.LineTo(lx1, ly1)
	gPath.LineTo(lx2, ly2)
	gPath.LineTo(lx3, ly3)
	gPath.LineTo(lx4, ly4)
	gPath.Close()
	var shadowOpts vector.DrawPathOptions
	shadowOpts.ColorScale.ScaleWithColor(shadowColor)
	vector.FillPath(screen, &gPath, nil, &shadowOpts)

	// Right Face
	gPath.Reset()
	rx0, ry0 := localToScreen(cx, cy, 0, 0, dirVec, perpVec)
	rx1, ry1 := localToScreen(cx, cy, 0, length*0.45, dirVec, perpVec)
	rx2, ry2 := localToScreen(cx, cy, 0, length, dirVec, perpVec)
	rx3, ry3 := localToScreen(cx, cy, width/2, length*0.4, dirVec, perpVec)
	rx4, ry4 := localToScreen(cx, cy, width/2, 0, dirVec, perpVec)

	gPath.MoveTo(rx0, ry0)
	gPath.LineTo(rx1, ry1)
	gPath.LineTo(rx2, ry2)
	gPath.LineTo(rx3, ry3)
	gPath.LineTo(rx4, ry4)
	gPath.Close()
	var highlightOpts vector.DrawPathOptions
	highlightOpts.ColorScale.ScaleWithColor(highlightColor)
	vector.FillPath(screen, &gPath, nil, &highlightOpts)
}

// drawCrystalCluster draws 3 crystal shards in a cluster
func drawCrystalCluster(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color, isSpiky bool) {
	baseLength := float32(28.0) * scale
	baseWidth := float32(11.0) * scale
	if isSpiky {
		baseLength = float32(34.0) * scale
		baseWidth = float32(7.0) * scale
	}

	// 1. Left shard (rotated by -0.42 radians, slightly shorter)
	leftDir := rotateVec(dirVec, -0.42)
	leftPerp := rotateVec(perpVec, -0.42)
	drawShard(screen, cx, cy, leftDir, leftPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	// 2. Right shard (rotated by +0.42 radians, slightly shorter)
	rightDir := rotateVec(dirVec, 0.42)
	rightPerp := rotateVec(perpVec, 0.42)
	drawShard(screen, cx, cy, rightDir, rightPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	// 3. Center shard (straight, full size)
	drawShard(screen, cx, cy, dirVec, perpVec, baseLength, baseWidth, shadowColor, highlightColor)
}

// drawNodule draws 4 overlapping sphere-like bumps to represent a nodule
func drawNodule(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, mineralColor, coreColor color.Color) {
	R := float32(12.0) * scale
	if R < 2.0 {
		R = 2.0
	}

	type bump struct {
		lx, ly float32
		r      float32
	}

	// Layering order: back to front (Left & Right, then Center, then Top/Tip)
	bumps := []bump{
		{-R * 0.45, R * 0.3, R * 0.75}, // Left
		{R * 0.45, R * 0.3, R * 0.75},  // Right
		{0, R * 0.5, R},                // Center
		{0, R * 0.85, R * 0.6},         // Top/Tip
	}

	for _, b := range bumps {
		bx, by := localToScreen(cx, cy, b.lx, b.ly, dirVec, perpVec)
		// 1. Fill base dark bump
		vector.FillCircle(screen, bx, by, b.r, darkenColor(mineralColor, 0.9), false)
		// 2. Stroke outline
		vector.StrokeCircle(screen, bx, by, b.r, 1.0, darkenColor(mineralColor, 0.5), false)

		// 3. Draw shiny highlight offset
		hx, hy := localToScreen(cx, cy, b.lx-b.r*0.25, b.ly+b.r*0.25, dirVec, perpVec)
		hr := b.r * 0.28
		if hr < 1.0 {
			hr = 1.0
		}
		vector.FillCircle(screen, hx, hy, hr, blendColor(coreColor, color.White, 0.6), false)
	}
}

// drawQuartzNeedles draws thin, long glowing quartz needles
func drawQuartzNeedles(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color) {
	baseLength := float32(36.0) * scale
	baseWidth := float32(4.5) * scale

	// Draw 4 needles pointing at various angles
	angles := []float32{-0.5, -0.18, 0.2, 0.55}
	lengths := []float32{0.7, 1.0, 0.85, 0.65}

	for i, angle := range angles {
		d := rotateVec(dirVec, angle)
		p := rotateVec(perpVec, angle)
		drawShard(screen, cx, cy, d, p, baseLength*lengths[i], baseWidth, shadowColor, highlightColor)
	}
}

// drawMineral renders the mineral based on its type and attachment direction
func drawMineral(screen *ebiten.Image, tx, ty int, camX, camY float64, hitsToMine int, mineralColor, coreColor color.Color, attachDir AttachDirection, mineralName string) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))

	// Determine basis vectors based on AttachDirection
	var cx, cy float32
	var dirVec, perpVec [2]float32

	switch attachDir {
	case AttachTop:
		cx = sx + float32(TileSize)/2.0
		cy = sy
		dirVec = [2]float32{0, 1}
		perpVec = [2]float32{-1, 0}
	case AttachLeft:
		cx = sx
		cy = sy + float32(TileSize)/2.0
		dirVec = [2]float32{1, 0}
		perpVec = [2]float32{0, -1}
	case AttachRight:
		cx = sx + float32(TileSize)
		cy = sy + float32(TileSize)/2.0
		dirVec = [2]float32{-1, 0}
		perpVec = [2]float32{0, 1}
	case AttachBottom:
		cx = sx + float32(TileSize)/2.0
		cy = sy + float32(TileSize)
		dirVec = [2]float32{0, -1}
		perpVec = [2]float32{1, 0}
	default: // AttachNone (fallback / icon center)
		cx = sx + float32(TileSize)/2.0
		cy = sy + float32(TileSize)/2.0 + 8.0 // offset slightly down so growing up centers it
		dirVec = [2]float32{0, -1}
		perpVec = [2]float32{1, 0}
	}

	// Scale size based on hits left
	scale := float32(hitsToMine) / 3.0
	if scale < 0.35 {
		scale = 0.35
	}

	shadowColor := darkenColor(mineralColor, 0.82)
	highlightColor := blendColor(mineralColor, coreColor, 0.65)

	switch mineralName {
	case "Copper":
		drawNodule(screen, cx, cy, dirVec, perpVec, scale, mineralColor, coreColor)
	case "Quartz":
		drawQuartzNeedles(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor)
	case "Abyssal Ore":
		// Glowing purple crystal shards
		drawSpikyCrystal := true
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, drawSpikyCrystal)
	default: // Titanium / default
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, false)
	}
}

// drawMineralIcon renders the mineral crystal or nodule centered at cx, cy with a custom size.
func drawMineralIcon(screen *ebiten.Image, cx, cy, size float32, mineralColor, coreColor color.Color, mineralName string) {
	scale := size / 40.0
	if scale < 0.2 {
		scale = 0.2
	}

	dirVec := [2]float32{0, -1}
	perpVec := [2]float32{1, 0}

	shadowColor := darkenColor(mineralColor, 0.82)
	highlightColor := blendColor(mineralColor, coreColor, 0.65)

	switch mineralName {
	case "Copper":
		drawNodule(screen, cx, cy, dirVec, perpVec, scale, mineralColor, coreColor)
	case "Quartz":
		drawQuartzNeedles(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor)
	case "Abyssal Ore":
		drawSpikyCrystal := true
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, drawSpikyCrystal)
	default: // Titanium / default
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, false)
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
	sheet, err := assets.LoadChromaKeyedImage("ore_sheet")
	if err != nil {
		log.Printf("Error: Failed to load ore sheet: %v", err)
		return
	}

	bounds := sheet.Bounds()

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
	coreColor := color.RGBA{220, 230, 240, 255}
	drawMineralIcon(screen, cx, cy, size, n.GetColor(), coreColor, "Titanium")
}

func (n *TitaniumNode) Draw(screen *ebiten.Image, camX, camY float64) {
	mineralColor := color.RGBA{160, 175, 185, 255} // Metallic silver
	coreColor := color.RGBA{220, 230, 240, 255}
	drawMineral(screen, n.Tx, n.Ty, camX, camY, n.HitsToMine, mineralColor, coreColor, n.AttachDir, "Titanium")
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
	coreColor := color.RGBA{240, 160, 80, 255}
	drawMineralIcon(screen, cx, cy, size, n.GetColor(), coreColor, "Copper")
}

func (n *CopperNode) Draw(screen *ebiten.Image, camX, camY float64) {
	mineralColor := color.RGBA{210, 110, 45, 255} // Copper orange
	coreColor := color.RGBA{240, 160, 80, 255}
	drawMineral(screen, n.Tx, n.Ty, camX, camY, n.HitsToMine, mineralColor, coreColor, n.AttachDir, "Copper")
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
	coreColor := color.RGBA{220, 250, 255, 255}
	drawMineralIcon(screen, cx, cy, size, n.GetColor(), coreColor, "Quartz")
}

func (n *QuartzNode) Draw(screen *ebiten.Image, camX, camY float64) {
	mineralColor := color.RGBA{40, 210, 245, 200} // Cyan bioluminescent quartz
	coreColor := color.RGBA{220, 250, 255, 255}
	drawMineral(screen, n.Tx, n.Ty, camX, camY, n.HitsToMine, mineralColor, coreColor, n.AttachDir, "Quartz")
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
	coreColor := color.RGBA{230, 180, 255, 255}
	drawMineralIcon(screen, cx, cy, size, n.GetColor(), coreColor, "Abyssal Ore")
}

func (n *AbyssalOreNode) Draw(screen *ebiten.Image, camX, camY float64) {
	mineralColor := color.RGBA{140, 40, 210, 255} // Glowing purple abyssal ore
	coreColor := color.RGBA{230, 180, 255, 255}
	drawMineral(screen, n.Tx, n.Ty, camX, camY, n.HitsToMine, mineralColor, coreColor, n.AttachDir, "Abyssal Ore")
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

	type nodeKind int
	const (
		kindTitanium nodeKind = iota
		kindCopper
		kindQuartz
		kindAbyssalOre
	)

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			// Place nodes in open (water) tiles that are adjacent to solid walls
			if !grid[tx][ty] {
				// Check which cardinal neighbors are solid blocks
				var possibleDirs []AttachDirection
				if grid[tx][ty-1] {
					possibleDirs = append(possibleDirs, AttachTop)
				}
				if grid[tx][ty+1] {
					possibleDirs = append(possibleDirs, AttachBottom)
				}
				if grid[tx-1][ty] {
					possibleDirs = append(possibleDirs, AttachLeft)
				}
				if grid[tx+1][ty] {
					possibleDirs = append(possibleDirs, AttachRight)
				}

				if len(possibleDirs) > 0 {
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
						// Pick one of the adjacent solid wall directions to attach to
						attachDir := possibleDirs[r.Intn(len(possibleDirs))]
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
						node.SetAttachDir(attachDir)
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
