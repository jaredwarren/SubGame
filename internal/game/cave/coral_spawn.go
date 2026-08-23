package cave

import (
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
)

const coralSpriteSize = 24

// MaybeSpawnCoral rolls chance against adjacent solid faces and, on success,
// appends a Coral entity positioned for the chosen attachment. Returns the
// (possibly unchanged) entities slice.
//
// RNG order matches the historical inline copies: attachment scan first, then
// chance roll, then variant — so existing cave seeds stay stable.
func MaybeSpawnCoral(
	entities []entity.CaveEntity,
	grid [][]bool,
	tx, ty int,
	chance float64,
	biome int,
	variantCount int,
	r *rand.Rand,
) []entity.CaveEntity {
	gridW, gridH := len(grid), len(grid[0])
	var attachments []string
	if ty+1 < gridH && grid[tx][ty+1] {
		attachments = append(attachments, "floor")
	}
	if ty-1 >= 0 && grid[tx][ty-1] {
		attachments = append(attachments, "ceiling")
	}
	if tx-1 >= 0 && grid[tx-1][ty] {
		attachments = append(attachments, "left")
	}
	if tx+1 < gridW && grid[tx+1][ty] {
		attachments = append(attachments, "right")
	}
	if len(attachments) == 0 || r.Float64() >= chance {
		return entities
	}

	attach := attachments[r.Intn(len(attachments))]
	variants := variantCount
	if variants < 1 {
		variants = 1
	}
	variant := r.Intn(variants)

	ts := config.TileSize
	cx := float64(tx * ts)
	cy := float64(ty * ts)
	switch attach {
	case "floor":
		cx += float64(ts-coralSpriteSize) / 2.0
		cy += float64(ts - coralSpriteSize)
	case "ceiling":
		cx += float64(ts-coralSpriteSize) / 2.0
	case "left":
		cy += float64(ts-coralSpriteSize) / 2.0
	case "right":
		cx += float64(ts - coralSpriteSize)
		cy += float64(ts-coralSpriteSize) / 2.0
	}

	return append(entities, entity.NewCoral(cx, cy, biome, attach, variant, r))
}
