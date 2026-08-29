package resource

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

// TileSize matches the global tile size used for rendering and physics.
const TileSize = 64

// AttachDirection defines where a resource attaches on a cave wall.
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
	GetDropCount() int
	// MapColor is the color used for overview / debug map markers.
	MapColor() color.RGBA
}

// BaseResourceNode holds the shared state for all resource node types.
type BaseResourceNode struct {
	Tx, Ty     int // Tile coordinates
	HitsToMine int
	MaxHits    int
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
	if hits > b.MaxHits {
		b.MaxHits = hits
	}
}

func (b *BaseResourceNode) GetDropCount() int {
	if b.MaxHits >= 5 || b.HitsToMine >= 5 {
		return 2
	}
	return 1
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
