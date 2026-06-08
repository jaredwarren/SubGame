package cave

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

type VoidCave struct{}

func NewVoidCave() *VoidCave {
	return &VoidCave{}
}

func (c *VoidCave) GetCaveType() CaveType { return CaveVoid }
func (c *VoidCave) GetGrid() [][]bool     { return nil }

func (c *VoidCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
	// Endless crushing dark void
	screen.Fill(color.RGBA{2, 3, 6, 255})
}

func (c *VoidCave) DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int) {
	// No tiles in the void
}

func (c *VoidCave) GenerateEntities(seed int64) []entity.CaveEntity {
	return nil
}

func (c *VoidCave) GenerateResources(seed int64) []resource.Resource {
	return nil
}
