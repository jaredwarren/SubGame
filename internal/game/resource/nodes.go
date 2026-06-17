package resource

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

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
// BlueprintNode
// ---------------------------------------------------------

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
