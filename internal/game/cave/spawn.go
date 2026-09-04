package cave

import (
	"math"
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// SpawnFauna creates a cave entity for the given fauna ID at tile (tx, ty).
// grid may be nil for types that do not need wall attachment; lurker/siphon
// need grid to pick an anchor face/direction.
func SpawnFauna(id FaunaID, tx, ty int, grid [][]bool, r *rand.Rand) entity.CaveEntity {
	def := entity.FaunaDefFor(id)
	if def == nil {
		return nil
	}
	ts := config.TileSize
	cx := float64(tx*ts) + float64(ts)/2.0
	cy := float64(ty*ts) + float64(ts)/2.0
	d := def.Dims

	switch id {
	case FaunaPassiveFish:
		return entity.NewPassiveFish(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
			r.Float64() < 0.5,
			r.Float64()*math.Pi*2,
		)
	case FaunaPassiveCrab:
		return entity.NewPassiveCrab(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y)),
		)
	case FaunaSandViper:
		return entity.NewSandViper(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y)),
		)
	case FaunaFalseBulbSnare:
		return entity.NewFalseBulbSnare(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+4,
		)
	case FaunaThermoclineRammer:
		rammer := entity.NewThermoclineRammer(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
		)
		rammer.Facing = r.Float64() * math.Pi * 2
		return rammer
	case FaunaElectroWeaver:
		return entity.NewElectroWeaver(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
		)
	case FaunaVoltaicLurker:
		face := pickSolidFace(grid, tx, ty, r)
		if face == "" {
			return nil
		}
		return entity.NewVoltaicLurker(float64(tx*ts), float64(ty*ts), face)
	case FaunaBrimstoneSiphon:
		dir := pickSiphonDir(grid, tx, ty, r)
		if dir == "" {
			return nil
		}
		siphon := entity.NewBrimstoneSiphon(
			cx-d.X/2.0, cy-d.Y/2.0, dir,
		)
		siphon.Timer = r.Intn(def.CycleFrames)
		return siphon
	case FaunaInkSquid:
		return entity.NewInkSquid(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
			r.Float64() < 0.5,
		)
	case FaunaLanternfish:
		return entity.NewLanternfish(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
			r.Float64() < 0.5,
			r.Float64()*math.Pi*2,
		)
	case FaunaGlowSquid:
		return entity.NewGlowSquid(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
			r.Float64() < 0.5,
		)
	default:
		return nil
	}
}

func pickSolidFace(grid [][]bool, tx, ty int, r *rand.Rand) string {
	if grid == nil {
		return ""
	}
	var faces []string
	if grid[tx-1][ty] {
		faces = append(faces, "left")
	}
	if grid[tx+1][ty] {
		faces = append(faces, "right")
	}
	if grid[tx][ty-1] {
		faces = append(faces, "top")
	}
	if grid[tx][ty+1] {
		faces = append(faces, "bottom")
	}
	if len(faces) == 0 {
		return ""
	}
	return faces[r.Intn(len(faces))]
}

func pickSiphonDir(grid [][]bool, tx, ty int, r *rand.Rand) string {
	if grid == nil {
		return ""
	}
	var dirs []string
	if grid[tx][ty+1] {
		dirs = append(dirs, "up")
	}
	if grid[tx][ty-1] {
		dirs = append(dirs, "down")
	}
	if grid[tx-1][ty] {
		dirs = append(dirs, "right")
	}
	if grid[tx+1][ty] {
		dirs = append(dirs, "left")
	}
	if len(dirs) == 0 {
		return ""
	}
	return dirs[r.Intn(len(dirs))]
}

// SpawnFlora creates a cave entity for the given flora ID at tile (tx, ty).
// height is used for kelp / shock kelp. Returns nil if the ID is unknown.
func SpawnFlora(id FloraID, tx, ty int, height float64, r *rand.Rand) entity.CaveEntity {
	return SpawnFloraAnchored(id, tx, ty, height, "floor", r)
}

// SpawnFloraAnchored creates flora with an explicit wall/floor anchor.
func SpawnFloraAnchored(id FloraID, tx, ty int, height float64, anchor string, r *rand.Rand) entity.CaveEntity {
	ts := config.TileSize
	switch id {
	case FloraShockKelp:
		w := entity.ShockKelpArchetype.FloorWidth
		var x, y float64
		switch anchor {
		case "left":
			x = float64(tx * ts)
			y = float64(ty*ts) + float64(ts)/2.0 - height
		case "right":
			x = float64(tx*ts) + float64(ts) - 28.0
			y = float64(ty*ts) + float64(ts)/2.0 - height
		default: // floor
			x = float64(tx*ts) + float64(ts-int(w))/2.0
			y = float64(ty*ts) + float64(ts) - height
			anchor = "floor"
		}
		return entity.NewShockKelp(x, y, height, anchor)
	case FloraShatterBulb:
		if height <= 0 {
			height = 42.0 + r.Float64()*12.0
		}
		var x, y float64
		wallW := entity.ShatterBulbArchetype.WallWidth
		if wallW <= 0 {
			wallW = 28.0
		}
		floorW := entity.ShatterBulbArchetype.FloorWidth
		if floorW <= 0 {
			floorW = 24.0
		}
		switch anchor {
		case "left":
			x = float64(tx * ts)
			y = float64(ty*ts) + float64(ts)/2.0 - height
		case "right":
			x = float64(tx*ts) + float64(ts) - wallW
			y = float64(ty*ts) + float64(ts)/2.0 - height
		default: // floor
			x = float64(tx*ts) + float64(ts-int(floorW))/2.0
			y = float64(ty*ts) + float64(ts) - height
			anchor = "floor"
		}
		return entity.NewShatterBulbAnchored(x, y, height, anchor)
	case FloraNerveMat:
		return &entity.NerveMat{
			BaseEntity: entity.BaseEntity{
				Pos:        gvec.Vec2{X: float64(tx * ts), Y: float64(ty*ts) + float64(ts-12)},
				Dimensions: gvec.Vec2{X: float64(ts), Y: 12},
				Active:     true,
			},
		}
	case FloraKelp, FloraCoral:
		// FloraCoral historically fell through to kelp in the floor-flora path.
		return &entity.Kelp{
			BaseEntity: entity.BaseEntity{
				Pos:        gvec.Vec2{X: float64(tx*ts) + float64(ts-16)/2.0, Y: float64(ty*ts) + float64(ts) - height},
				Dimensions: gvec.Vec2{X: 16, Y: height},
				Active:     true,
			},
			SwayPhase: r.Float64() * math.Pi * 2,
		}
	default:
		return nil
	}
}
