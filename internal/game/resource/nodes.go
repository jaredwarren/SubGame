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
	Material     *item.MaterialDef // nil for Blueprint
	Name         string            // used when Material is nil
	MaxStack     int               // used when Material is nil
	RequiresMech bool              // used when Material is nil; otherwise derived from Material
	BaseItem     func() item.Item
	Color        color.Color // used when Material is nil
	DrawIcon     func(screen *ebiten.Image, cx, cy, size float32)
	Draw         func(screen *ebiten.Image, node *ResourceNode, camX, camY float64)
}

func oreNodeInfo(m *item.MaterialDef, baseItem func() item.Item) *NodeTypeInfo {
	return &NodeTypeInfo{
		Material: m,
		BaseItem: baseItem,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			drawMineralIcon(screen, cx, cy, size, m.Color, m.CoreColor, m.Name)
		},
		Draw: func(screen *ebiten.Image, node *ResourceNode, camX, camY float64) {
			drawMineral(screen, node.Tx, node.Ty, camX, camY, node.HitsToMine, m.WorldColor, m.WorldCoreColor, node.AttachDir, m.Name)
		},
	}
}

var nodeRegistry = map[NodeType]*NodeTypeInfo{
	NodeTitanium:   oreNodeInfo(item.MaterialTitanium, func() item.Item { return &item.Titanium{} }),
	NodeCopper:     oreNodeInfo(item.MaterialCopper, func() item.Item { return &item.Copper{} }),
	NodeQuartz:     oreNodeInfo(item.MaterialQuartz, func() item.Item { return &item.Quartz{} }),
	NodeAbyssalOre: oreNodeInfo(item.MaterialAbyssalOre, func() item.Item { return &item.AbyssalOre{} }),
	NodeNickel:     oreNodeInfo(item.MaterialNickel, func() item.Item { return &item.Nickel{} }),
	NodeScrapMetal: {
		Material: item.MaterialScrapMetal,
		BaseItem: func() item.Item { return &item.ScrapMetal{} },
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
		Material: item.MaterialElectronicWaste,
		BaseItem: func() item.Item { return &item.ElectronicWaste{} },
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

func (n *ResourceNode) info() *NodeTypeInfo {
	return nodeRegistry[n.Type]
}

func (n *ResourceNode) GetName() string {
	if n.Type == NodeBlueprint {
		return "Blueprint: " + n.RecipeResultName
	}
	info := n.info()
	if info.Material != nil {
		return info.Material.Name
	}
	return info.Name
}

func (n *ResourceNode) GetMaxStack() int {
	info := n.info()
	if info.Material != nil {
		return info.Material.MaxStack
	}
	return info.MaxStack
}

func (n *ResourceNode) RequiresMech() bool {
	info := n.info()
	if info.Material != nil {
		return info.Material.RequiresMech
	}
	return info.RequiresMech
}

func (n *ResourceNode) GetBaseItem() item.Item {
	return n.info().BaseItem()
}

func (n *ResourceNode) GetColor() color.Color {
	info := n.info()
	if info.Material != nil {
		return info.Material.Color
	}
	return info.Color
}

func (n *ResourceNode) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	n.info().DrawIcon(screen, cx, cy, size)
}

func (n *ResourceNode) Draw(screen *ebiten.Image, camX, camY float64) {
	n.info().Draw(screen, n, camX, camY)
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
