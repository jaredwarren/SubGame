package cave

import (
	"image/color"
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/game/entity"
)

// ChasmVeinStyle controls how crack/vein colors are picked on chasm-blended tiles.
type ChasmVeinStyle int

const (
	ChasmVeinSingle ChasmVeinStyle = iota
	ChasmVeinDual   // alternate two colors by tile hash parity
)

// ChasmRimSpawn places flora/fauna along the shallow chasm rim when topology matches.
type ChasmRimSpawn struct {
	Chance     float64
	SideWall   bool // grid[x-1][y] || grid[x+1][y]
	Ceiling    bool // grid[x][y-1]
	LeftWall   bool // grid[x-1][y]
	RightWall  bool // grid[x+1][y]
	FloorBelow bool // grid[x][y+1] — spawn anchored flora on the floor tile below

	IsFlora bool
	IsFauna bool
	Flora   FloraID
	Fauna   FaunaID
	MinH    float64
	MaxH    float64
	Anchor  string // left, right, floor (flora only)
}

// ChasmRimSpec describes draw palette and rim spawns for a shallow chasm leading to target.
type ChasmRimSpec struct {
	Target       CaveType
	RockColor    color.RGBA
	StrokeColor  color.RGBA
	VeinStyle    ChasmVeinStyle
	VeinPrimary  color.RGBA
	VeinAltA     color.RGBA
	VeinAltB     color.RGBA
	VeinSecondary color.RGBA
	AmbientRGB   [3]float32 // RGB; alpha applied at runtime from lightMult
	Spawns       []ChasmRimSpawn
}

var chasmRimRegistry = map[CaveType]*ChasmRimSpec{}

func registerChasmRim(s *ChasmRimSpec) {
	chasmRimRegistry[s.Target] = s
}

// ChasmRimSpecFor returns rim data for a subterranean target cave, or the shock-kelp default.
func ChasmRimSpecFor(target CaveType) *ChasmRimSpec {
	if s := chasmRimRegistry[target]; s != nil {
		return s
	}
	return chasmRimRegistry[CaveShockKelp]
}

func init() {
	registerChasmRim(&ChasmRimSpec{
		Target:      CaveOrganicTrench,
		RockColor:   color.RGBA{22, 28, 44, 255},
		StrokeColor: color.RGBA{38, 48, 70, 255},
		VeinStyle:   ChasmVeinSingle,
		VeinPrimary: color.RGBA{0, 190, 220, 160},
		VeinSecondary: color.RGBA{0, 230, 255, 120},
		AmbientRGB:  [3]float32{0.02, 0.04, 0.08},
		Spawns: []ChasmRimSpawn{
			{Chance: 0.25, SideWall: true, IsFlora: true, Flora: FloraShatterBulb},
			{Chance: 0.20, Ceiling: true, IsFauna: true, Fauna: FaunaFalseBulbSnare},
		},
	})
	registerChasmRim(&ChasmRimSpec{
		Target:      CaveShockKelp,
		RockColor:   color.RGBA{55, 60, 68, 255},
		StrokeColor: color.RGBA{82, 88, 98, 255},
		VeinStyle:   ChasmVeinDual,
		VeinAltA:    color.RGBA{140, 50, 210, 160},
		VeinAltB:    color.RGBA{0, 220, 255, 140},
		VeinSecondary: color.RGBA{0, 220, 255, 110},
		AmbientRGB:  [3]float32{0.03, 0.05, 0.09},
		Spawns: []ChasmRimSpawn{
			{Chance: 0.35, LeftWall: true, IsFlora: true, Flora: FloraShockKelp, MinH: 28, MaxH: 28, Anchor: "left"},
			{Chance: 0.35, RightWall: true, IsFlora: true, Flora: FloraShockKelp, MinH: 28, MaxH: 28, Anchor: "right"},
			{Chance: 0.30, FloorBelow: true, IsFlora: true, Flora: FloraShockKelp, MinH: 32, MaxH: 32, Anchor: "floor"},
		},
	})
}

// primaryVeinColor picks the main vein tint for a chasm-blended tile.
func (s *ChasmRimSpec) primaryVeinColor(tileHash uint64) color.RGBA {
	if s == nil {
		return color.RGBA{}
	}
	switch s.VeinStyle {
	case ChasmVeinDual:
		if tileHash%2 == 0 {
			return s.VeinAltA
		}
		return s.VeinAltB
	default:
		return s.VeinPrimary
	}
}

// spawnChasmRimEntities places rim flora/fauna from spec within the chasm x-span.
func spawnChasmRimEntities(spec *ChasmRimSpec, grid [][]bool, chasmX, chasmWidth int, r *rand.Rand) []entity.CaveEntity {
	if spec == nil || len(spec.Spawns) == 0 || chasmX <= 0 {
		return nil
	}
	gridW := len(grid)
	gridH := len(grid[0])
	var entities []entity.CaveEntity

	for y := 10; y < gridH-2; y += 3 {
		for x := chasmX - 2; x <= chasmX+chasmWidth+2; x++ {
			if x <= 1 || x >= gridW-1 || grid[x][y] {
				continue
			}
			for _, sp := range spec.Spawns {
				if r.Float64() >= sp.Chance {
					continue
				}
				if sp.SideWall && !(grid[x-1][y] || grid[x+1][y]) {
					continue
				}
				if sp.Ceiling && !grid[x][y-1] {
					continue
				}
				if sp.LeftWall && !grid[x-1][y] {
					continue
				}
				if sp.RightWall && !grid[x+1][y] {
					continue
				}
				if sp.FloorBelow {
					if y >= gridH-2 || !grid[x][y+1] {
						continue
					}
				}
				if sp.IsFauna {
					spawnTX, spawnTY := x, y
					if ent := SpawnFauna(sp.Fauna, spawnTX, spawnTY, grid, r); ent != nil {
						entities = append(entities, ent)
					}
					continue
				}
				if sp.IsFlora {
					spawnTX, spawnTY := x, y
					if sp.FloorBelow {
						spawnTY = y + 1
					}
					minH, maxH := sp.MinH, sp.MaxH
					if maxH <= minH {
						minH, maxH = 24, 52
					}
					height := minH
					if maxH > minH {
						height = minH + r.Float64()*(maxH-minH)
					}
					anchor := sp.Anchor
					if anchor == "" {
						anchor = "floor"
					}
					if ent := SpawnFloraAnchored(sp.Flora, spawnTX, spawnTY, height, anchor, r); ent != nil {
						entities = append(entities, ent)
					}
				}
			}
		}
	}
	return entities
}
