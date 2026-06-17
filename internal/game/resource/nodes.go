package resource

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

type NodeType int

const (
	NodeTitanium NodeType = iota
	NodeCopper
	NodeQuartz
	NodeAbyssalOre
	NodeNickel
	NodeScrapMetal
	NodeElectronicWaste
	NodeBlueprint
)

type NodeTypeInfo struct {
	Name         string
	MaxStack     int
	RequiresMech bool
	BaseItem     func() item.Item
	Color        color.Color
	DrawIcon     func(screen *ebiten.Image, cx, cy, size float32)
	Draw         func(screen *ebiten.Image, node *ResourceNode, camX, camY float64)
}

var nodeRegistry = map[NodeType]*NodeTypeInfo{
	NodeTitanium: {
		Name:         "Titanium",
		MaxStack:     10,
		RequiresMech: false,
		BaseItem:     func() item.Item { return &item.Titanium{} },
		Color:        color.RGBA{168, 178, 188, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{220, 230, 240, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{168, 178, 188, 255}, coreColor, "Titanium")
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			mineralColor := color.RGBA{160, 175, 185, 255} // Metallic silver
			coreColor := color.RGBA{220, 230, 240, 255}
			drawMineral(screen, node.Tx, node.Ty, camX, camY, node.HitsToMine, mineralColor, coreColor, node.AttachDir, "Titanium")
		},
	},
	NodeCopper: {
		Name:         "Copper",
		MaxStack:     10,
		RequiresMech: false,
		BaseItem:     func() item.Item { return &item.Copper{} },
		Color:        color.RGBA{218, 118, 48, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{240, 160, 80, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{218, 118, 48, 255}, coreColor, "Copper")
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			mineralColor := color.RGBA{210, 110, 45, 255} // Copper orange
			coreColor := color.RGBA{240, 160, 80, 255}
			drawMineral(screen, node.Tx, node.Ty, camX, camY, node.HitsToMine, mineralColor, coreColor, node.AttachDir, "Copper")
		},
	},
	NodeQuartz: {
		Name:         "Quartz",
		MaxStack:     10,
		RequiresMech: false,
		BaseItem:     func() item.Item { return &item.Quartz{} },
		Color:        color.RGBA{48, 218, 245, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{220, 250, 255, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{48, 218, 245, 255}, coreColor, "Quartz")
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			mineralColor := color.RGBA{40, 210, 245, 200} // Cyan bioluminescent quartz
			coreColor := color.RGBA{220, 250, 255, 255}
			drawMineral(screen, node.Tx, node.Ty, camX, camY, node.HitsToMine, mineralColor, coreColor, node.AttachDir, "Quartz")
		},
	},
	NodeAbyssalOre: {
		Name:         "Abyssal Ore",
		MaxStack:     10,
		RequiresMech: true,
		BaseItem:     func() item.Item { return &item.AbyssalOre{} },
		Color:        color.RGBA{148, 48, 218, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{230, 180, 255, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{148, 48, 218, 255}, coreColor, "Abyssal Ore")
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			mineralColor := color.RGBA{140, 40, 210, 255} // Glowing purple abyssal ore
			coreColor := color.RGBA{230, 180, 255, 255}
			drawMineral(screen, node.Tx, node.Ty, camX, camY, node.HitsToMine, mineralColor, coreColor, node.AttachDir, "Abyssal Ore")
		},
	},
	NodeNickel: {
		Name:         "Nickel",
		MaxStack:     10,
		RequiresMech: false,
		BaseItem:     func() item.Item { return &item.Nickel{} },
		Color:        color.RGBA{162, 175, 148, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			coreColor := color.RGBA{222, 235, 208, 255}
			drawMineralIcon(screen, cx, cy, size, color.RGBA{162, 175, 148, 255}, coreColor, "Nickel")
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			mineralColor := color.RGBA{150, 165, 140, 255} // Warm greenish-silver
			coreColor := color.RGBA{222, 235, 208, 255}
			drawMineral(screen, node.Tx, node.Ty, camX, camY, node.HitsToMine, mineralColor, coreColor, node.AttachDir, "Nickel")
		},
	},
	NodeScrapMetal: {
		Name:         "Scrap Metal",
		MaxStack:     10,
		RequiresMech: false,
		BaseItem:     func() item.Item { return &item.ScrapMetal{} },
		Color:        color.RGBA{140, 110, 95, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			(&item.ScrapMetal{}).DrawIcon(screen, cx, cy, size)
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			sx, sy := drawNodeBase(screen, node.Tx, node.Ty, camX, camY)
			cx := sx + float32(TileSize)/2.0
			cy := sy + float32(TileSize)/2.0
			size := float32(14.0) * (float32(node.HitsToMine) / 3.0)
			if size < 4.0 {
				size = 4.0
			}
			// Scrap crate: brown/rust box with dark gray outline
			vector.FillRect(screen, cx-size, cy-size, size*2.0, size*2.0, color.RGBA{130, 95, 75, 255}, false)
			vector.StrokeRect(screen, cx-size, cy-size, size*2.0, size*2.0, 1.5, color.RGBA{90, 80, 75, 255}, false)
			vector.StrokeLine(screen, cx-size, cy-size, cx+size, cy+size, 1.5, color.RGBA{90, 80, 75, 255}, false)

			drawCracks(screen, float32(sx), float32(sy), node.HitsToMine)
		},
	},
	NodeElectronicWaste: {
		Name:         "Electronic Waste",
		MaxStack:     10,
		RequiresMech: false,
		BaseItem:     func() item.Item { return &item.ElectronicWaste{} },
		Color:        color.RGBA{70, 130, 90, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			(&item.ElectronicWaste{}).DrawIcon(screen, cx, cy, size)
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			sx, sy := drawNodeBase(screen, node.Tx, node.Ty, camX, camY)
			cx := sx + float32(TileSize)/2.0
			cy := sy + float32(TileSize)/2.0
			size := float32(14.0) * (float32(node.HitsToMine) / 3.0)
			if size < 4.0 {
				size = 4.0
			}
			// Electronic green container crate
			vector.FillRect(screen, cx-size, cy-size, size*2.0, size*2.0, color.RGBA{50, 110, 70, 255}, false)
			vector.StrokeRect(screen, cx-size, cy-size, size*2.0, size*2.0, 1.5, color.RGBA{110, 190, 130, 255}, false)
			vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, color.RGBA{30, 30, 30, 255}, false)

			drawCracks(screen, float32(sx), float32(sy), node.HitsToMine)
		},
	},
	NodeBlueprint: {
		Name:         "Blueprint",
		MaxStack:     1,
		RequiresMech: false,
		BaseItem:     func() item.Item { return nil }, // immediately unlocks, no item in inventory
		Color:        color.RGBA{0, 180, 255, 255},
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			// Simple blueprint icon: blue circle with outline
			vector.FillCircle(screen, cx, cy, size/2.0, color.RGBA{0, 180, 255, 255}, false)
			vector.StrokeCircle(screen, cx, cy, size/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			sx, sy := drawNodeBase(screen, node.Tx, node.Ty, camX, camY)
			cx := sx + float32(TileSize)/2.0
			cy := sy + float32(TileSize)/2.0

			size := float32(14.0) * (float32(node.HitsToMine) / 3.0)
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

			drawCracks(screen, sx, sy, node.HitsToMine)
		},
	},
}

type ResourceNode struct {
	BaseResourceNode
	Type             NodeType
	RecipeResultName string
}

func (n *ResourceNode) GetName() string {
	if n.Type == NodeBlueprint {
		return "Blueprint: " + n.RecipeResultName
	}
	return nodeRegistry[n.Type].Name
}

func (n *ResourceNode) GetMaxStack() int {
	return nodeRegistry[n.Type].MaxStack
}

func (n *ResourceNode) RequiresMech() bool {
	return nodeRegistry[n.Type].RequiresMech
}

func (n *ResourceNode) GetBaseItem() item.Item {
	return nodeRegistry[n.Type].BaseItem()
}

func (n *ResourceNode) GetColor() color.Color {
	return nodeRegistry[n.Type].Color
}

func (n *ResourceNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	nodeRegistry[n.Type].DrawIcon(screen, cx, cy, size)
}

func (n *ResourceNode) Draw(screen *ebiten.Image, camX, camY float64) {
	nodeRegistry[n.Type].Draw(screen, n, camX, camY)
}

func (n *ResourceNode) GetRecipeResultName() string {
	if n.Type == NodeBlueprint {
		return n.RecipeResultName
	}
	return ""
}

// ---------------------------------------------------------
// Wrappers embedding ResourceNode for type-assertions compatibility
// ---------------------------------------------------------

type TitaniumNode struct{ ResourceNode }

func (n *TitaniumNode) GetBaseItem() item.Item { return &item.Titanium{} }

func NewTitaniumNode(tx, ty int) *TitaniumNode {
	return &TitaniumNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeTitanium,
	}}
}

type CopperNode struct{ ResourceNode }

func (n *CopperNode) GetBaseItem() item.Item { return &item.Copper{} }

func NewCopperNode(tx, ty int) *CopperNode {
	return &CopperNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeCopper,
	}}
}

type QuartzNode struct{ ResourceNode }

func (n *QuartzNode) GetBaseItem() item.Item { return &item.Quartz{} }

func NewQuartzNode(tx, ty int) *QuartzNode {
	return &QuartzNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeQuartz,
	}}
}

type AbyssalOreNode struct{ ResourceNode }

func (n *AbyssalOreNode) GetBaseItem() item.Item { return &item.AbyssalOre{} }

func NewAbyssalOreNode(tx, ty int) *AbyssalOreNode {
	return &AbyssalOreNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeAbyssalOre,
	}}
}

type NickelNode struct{ ResourceNode }

func (n *NickelNode) GetBaseItem() item.Item { return &item.Nickel{} }

func NewNickelNode(tx, ty int) *NickelNode {
	return &NickelNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeNickel,
	}}
}

type ScrapMetalNode struct{ ResourceNode }

func (n *ScrapMetalNode) GetBaseItem() item.Item { return &item.ScrapMetal{} }

func NewScrapMetalNode(tx, ty int) *ScrapMetalNode {
	return &ScrapMetalNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeScrapMetal,
	}}
}

type ElectronicWasteNode struct{ ResourceNode }

func (n *ElectronicWasteNode) GetBaseItem() item.Item { return &item.ElectronicWaste{} }

func NewElectronicWasteNode(tx, ty int) *ElectronicWasteNode {
	return &ElectronicWasteNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeElectronicWaste,
	}}
}

type BlueprintNode struct{ ResourceNode }

func (n *BlueprintNode) GetBaseItem() item.Item { return nil }

func NewBlueprintNode(tx, ty int, recipeResultName string) *BlueprintNode {
	return &BlueprintNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeBlueprint,
		RecipeResultName: recipeResultName,
	}}
}
