package entity

import (
	"math"
	"math/rand"

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

// GenerateCaveEntities scans the cave grid and spawns biome-specific entities.
func GenerateCaveEntities(grid [][]bool, seed int64, isShallow bool) []CaveEntity {
	r := rand.New(rand.NewSource(seed))
	var entities []CaveEntity

	gridW := len(grid)
	gridH := len(grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			if grid[tx][ty] {
				continue
			}

			if isShallow {
				hasAdjacentWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
				if hasAdjacentWall && r.Float64() < 0.08 {
					entities = append(entities, &ShatterBulb{
						BaseEntity: BaseEntity{
							Type:       EntShatterBulb,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 24, Y: 24},
							Active:     true,
						},
					})
				}
				isOpenWater := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenWater && r.Float64() < 0.012 {
					entities = append(entities, &PassiveFish{
						BaseEntity: BaseEntity{
							Type:       EntPassiveFish,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-20)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-12)/2.0},
							Dimensions: gvec.Vec2{X: 20, Y: 12},
							Active:     true,
						},
						FacingRight: r.Float64() < 0.5,
						SwimPhase:   r.Float64() * math.Pi * 2,
					})
				}
				if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < 0.015 {
					entities = append(entities, &PassiveCrab{
						BaseEntity: BaseEntity{
							Type:       EntPassiveCrab,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-16)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-10)},
							Dimensions: gvec.Vec2{X: 16, Y: 10},
							Active:     true,
						},
						FacingRight: r.Float64() < 0.5,
					})
				}
				if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < 0.28 {
					height := 32.0 + r.Float64()*48.0
					entities = append(entities, &Kelp{
						BaseEntity: BaseEntity{
							Type:       EntKelp,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-16)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize) - height},
							Dimensions: gvec.Vec2{X: 16, Y: height},
							Active:     true,
						},
						SwayPhase: r.Float64() * math.Pi * 2,
					})
				}
				continue
			}

			// Biome 1: Mid-Depth (ty 4–40) - Grotto
			if ty >= 4 && ty < 40 {
				hasAdjacentWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
				if hasAdjacentWall && r.Float64() < 0.08 {
					entities = append(entities, &ShatterBulb{
						BaseEntity: BaseEntity{
							Type:       EntShatterBulb,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 24, Y: 24},
							Active:     true,
						},
					})
				}
				if grid[tx][ty-1] && r.Float64() < 0.04 {
					entities = append(entities, &FalseBulbSnare{
						BaseEntity: BaseEntity{
							Type:       EntFalseBulbSnare,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + 4},
							Dimensions: gvec.Vec2{X: 24, Y: 32},
							Active:     true,
						},
						State: 0,
					})
				}
			}

			// Biome 2: Deep (ty 40–80) - Smoker Trenches
			if ty >= 40 && ty < 80 {
				if r.Float64() < 0.05 {
					var dir string
					if grid[tx][ty+1] {
						dir = "up"
					} else if grid[tx][ty-1] {
						dir = "down"
					} else if grid[tx-1][ty] {
						dir = "right"
					} else if grid[tx+1][ty] {
						dir = "left"
					}
					if dir != "" {
						entities = append(entities, &BrimstoneSiphon{
							BaseEntity: BaseEntity{
								Type:       EntBrimstoneSiphon,
								Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-32)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-32)/2.0},
								Dimensions: gvec.Vec2{X: 32, Y: 32},
								Active:     true,
							},
							Direction: dir,
							Timer:     r.Intn(120),
						})
					}
				}
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.015 {
					entities = append(entities, &ThermoclineRammer{
						BaseEntity: BaseEntity{
							Type:       EntThermoclineRammer,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-36)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 36, Y: 24},
							Active:     true,
						},
						Facing: r.Float64() * math.Pi * 2,
					})
				}
			}

			// Biome 3: Abyssal (ty 80+) - Brine Falls
			if ty >= 80 && ty < gridH-1 {
				if grid[tx][ty+1] && r.Float64() < 0.10 {
					entities = append(entities, &NerveMat{
						BaseEntity: BaseEntity{
							Type:       EntNerveMat,
							Pos:        gvec.Vec2{X: float64(tx * config.TileSize), Y: float64(ty*config.TileSize) + float64(config.TileSize-12)},
							Dimensions: gvec.Vec2{X: float64(config.TileSize), Y: 12},
							Active:     true,
						},
					})
				}
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.012 {
					hasWeaverNearby := false
					for _, ent := range entities {
						if ent.GetType() == EntElectroWeaver && math.Abs(ent.GetPos().X-float64(tx*config.TileSize)) < 500 {
							hasWeaverNearby = true
							break
						}
					}
					if !hasWeaverNearby {
						entities = append(entities, &ElectroWeaver{
							BaseEntity: BaseEntity{
								Type:       EntElectroWeaver,
								Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-40)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-20)/2.0},
								Dimensions: gvec.Vec2{X: 40, Y: 20},
								Active:     true,
							},
						})
					}
				}
			}
		}
	}

	return entities
}

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
