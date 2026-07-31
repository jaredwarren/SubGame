package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// MaterialDef owns shared identity, stack, and visual data for minerals/materials
// referenced by both inventory items and world resource nodes.
type MaterialDef struct {
	Name           string
	MaxStack       int
	Color          color.RGBA // inventory/HUD
	CoreColor      color.RGBA // icon highlight (zero ok for non-mineral icons)
	WorldColor     color.RGBA // in-world node body (for ores)
	WorldCoreColor color.RGBA // in-world node core
	RequiresMech   bool
}

var (
	MaterialTitanium = &MaterialDef{
		Name:           "Titanium",
		MaxStack:       10,
		Color:          color.RGBA{168, 178, 188, 255},
		CoreColor:      color.RGBA{220, 230, 240, 255},
		WorldColor:     color.RGBA{160, 175, 185, 255},
		WorldCoreColor: color.RGBA{220, 230, 240, 255},
	}
	MaterialCopper = &MaterialDef{
		Name:           "Copper",
		MaxStack:       10,
		Color:          color.RGBA{218, 118, 48, 255},
		CoreColor:      color.RGBA{240, 160, 80, 255},
		WorldColor:     color.RGBA{210, 110, 45, 255},
		WorldCoreColor: color.RGBA{240, 160, 80, 255},
	}
	MaterialQuartz = &MaterialDef{
		Name:           "Quartz",
		MaxStack:       10,
		Color:          color.RGBA{48, 218, 245, 255},
		CoreColor:      color.RGBA{220, 250, 255, 255},
		WorldColor:     color.RGBA{40, 210, 245, 200},
		WorldCoreColor: color.RGBA{220, 250, 255, 255},
	}
	MaterialAbyssalOre = &MaterialDef{
		Name:           "Abyssal Ore",
		MaxStack:       10,
		Color:          color.RGBA{148, 48, 218, 255},
		CoreColor:      color.RGBA{230, 180, 255, 255},
		WorldColor:     color.RGBA{140, 40, 210, 255},
		WorldCoreColor: color.RGBA{230, 180, 255, 255},
		RequiresMech:   true,
	}
	MaterialNickel = &MaterialDef{
		Name:           "Nickel",
		MaxStack:       10,
		Color:          color.RGBA{162, 175, 148, 255},
		CoreColor:      color.RGBA{222, 235, 208, 255},
		WorldColor:     color.RGBA{150, 165, 140, 255},
		WorldCoreColor: color.RGBA{222, 235, 208, 255},
	}
	MaterialScrapMetal = &MaterialDef{
		Name:     "Scrap Metal",
		MaxStack: 10,
		Color:    color.RGBA{140, 110, 95, 255},
	}
	MaterialElectronicWaste = &MaterialDef{
		Name:     "Electronic Waste",
		MaxStack: 10,
		Color:    color.RGBA{70, 130, 90, 255},
	}
)

func mineralMetadata(m *MaterialDef) *ItemMetadata {
	return &ItemMetadata{
		Name:     m.Name,
		MaxStack: m.MaxStack,
		Color:    m.Color,
		DrawIcon: func(screen *ebiten.Image, cx, cy, size float32) {
			drawMineralIcon(screen, cx, cy, size, m.Color, m.CoreColor, m.Name)
		},
	}
}
