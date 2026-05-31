package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// EntityType identifies a specific biome entity variant.
type EntityType int

const (
	EntShatterBulb EntityType = iota
	EntFalseBulbSnare
	EntBrimstoneSiphon
	EntThermoclineRammer
	EntNerveMat
	EntElectroWeaver
	EntPassiveFish
	EntPassiveCrab
	EntKelp
)

// CaveEntity represents any plant, predator, or interactive entity inside caves.
type CaveEntity interface {
	Update(gr Runtime, CaveGrid [][]bool)
	Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64)
	IsActive() bool
	SetActive(active bool)
	GetPos() gvec.Vec2
	GetDimensions() gvec.Vec2
	GetType() EntityType
}

// PassiveCreature defines the interface for catchable cave creatures.
type PassiveCreature interface {
	CaveEntity
	GetHarvestedItem() item.Item
	CanCatch(playerPos gvec.Vec2) bool
}

// BaseEntity implements common fields and getters/setters for all entities.
type BaseEntity struct {
	Type       EntityType
	Pos        gvec.Vec2
	Vel        gvec.Vec2
	Dimensions gvec.Vec2
	Active     bool
}

func (b *BaseEntity) IsActive() bool           { return b.Active }
func (b *BaseEntity) SetActive(active bool)    { b.Active = active }
func (b *BaseEntity) GetPos() gvec.Vec2        { return b.Pos }
func (b *BaseEntity) GetDimensions() gvec.Vec2 { return b.Dimensions }
func (b *BaseEntity) GetType() EntityType      { return b.Type }

// entityPath is a pre-allocated path reused across entity Draw calls to avoid allocations.
var entityPath = &vector.Path{}

// rectsOverlap reports whether two axis-aligned rectangles intersect.
func rectsOverlap(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// isSolid checks whether the proposed bounding box overlaps a solid tile.
func isSolid(grid [][]bool, x, y, w, h float64) bool {
	if grid == nil {
		return false
	}
	gridW := len(grid)
	gridH := len(grid[0])

	x1 := int(math.Floor(x)) / config.TileSize
	x2 := int(math.Floor(x+w)) / config.TileSize
	y1 := int(math.Floor(y)) / config.TileSize
	y2 := int(math.Floor(y+h)) / config.TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= gridW {
				return true
			}
			if ty < 0 {
				continue
			}
			if ty >= gridH {
				return true
			}
			if grid[tx][ty] {
				return true
			}
		}
	}
	return false
}
