package cave

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

const (
	CaveWidth  = 60
	CaveHeight = 120
	SplitY     = 60
)

type CaveType int

const (
	CaveOrganicShallow CaveType = iota
	CaveOrganicTrench
	CaveWreckage
	CaveVoid
	CaveShockKelp
	CaveThermo
)

// Cave defines the interface for different cave types.
type Cave interface {
	GetCaveType() CaveType
	GetGrid() [][]bool
	DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64)
	DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int)
	GenerateEntities(seed int64) []entity.CaveEntity
	GenerateResources(seed int64) []resource.Resource
	GetAmbientColor(lightMult float64) []float32
}
