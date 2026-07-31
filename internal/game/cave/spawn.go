package cave

import (
	"math"
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// SpawnFauna creates a cave entity for the given fauna ID at tile (tx, ty).
// Returns nil if the ID is unknown.
func SpawnFauna(id FaunaID, tx, ty int, r *rand.Rand) entity.CaveEntity {
	ts := config.TileSize
	switch id {
	case FaunaPassiveFish:
		d := entity.PassiveFishArchetype.Dims
		return entity.NewPassiveFish(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
			r.Float64() < 0.5,
			r.Float64()*math.Pi*2,
		)
	case FaunaPassiveCrab:
		d := entity.PassiveCrabArchetype.Dims
		return entity.NewPassiveCrab(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y)),
		)
	case FaunaSandViper:
		d := entity.SandViperArchetype.Dims
		return entity.NewSandViper(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y)),
		)
	default:
		return nil
	}
}

// SpawnFlora creates a cave entity for the given flora ID at tile (tx, ty).
// height is used for kelp / shock kelp. Returns nil if the ID is unknown.
func SpawnFlora(id FloraID, tx, ty int, height float64, r *rand.Rand) entity.CaveEntity {
	ts := config.TileSize
	switch id {
	case FloraShockKelp:
		w := entity.ShockKelpArchetype.FloorWidth
		return entity.NewShockKelp(
			float64(tx*ts)+float64(ts-int(w))/2.0,
			float64(ty*ts)+float64(ts)-height,
			height,
			"floor",
		)
	case FloraShatterBulb:
		d := entity.ShatterBulbArchetype.Dims
		return entity.NewShatterBulb(
			float64(tx*ts)+float64(ts-int(d.X))/2.0,
			float64(ty*ts)+float64(ts-int(d.Y))/2.0,
		)
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
