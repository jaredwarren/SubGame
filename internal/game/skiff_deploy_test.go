package game

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func TestActivatePlayerItem_SkiffDeploysOnClearWater(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.OverworldVehicles = nil
	g.ActiveVehicle = nil

	dims := vehicle.SkiffArchetype.Dims
	kit := &vehicle.SkiffKit{}

	cases := []struct {
		name string
		pos  gvec.Vec2
	}{
		{
			name: "on lifepod",
			pos:  g.baseStation.Pos,
		},
		{
			name: "on land near water",
			pos:  landTileNearWater(g),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g.OverworldVehicles = nil
			g.player.Pos = tc.pos
			g.player.Inventory.Clear()
			g.player.Inventory.AddItem(kit, 1)

			g.ActivatePlayerItem(kit)

			if len(g.OverworldVehicles) != 1 {
				t.Fatalf("expected 1 overworld vehicle, got %d", len(g.OverworldVehicles))
			}
			skiff := g.OverworldVehicles[0]
			pos := skiff.GetPos()
			if !g.isClearOverworldDeploy(pos, dims) {
				t.Fatalf("skiff deployed into blocked space at %+v (player was at %+v)", pos, tc.pos)
			}

			// Nearest clear tile should still be close to the player.
			cx := pos.X + dims.X/2
			cy := pos.Y + dims.Y/2
			dist := hypot(cx-tc.pos.X, cy-tc.pos.Y)
			if dist > float64(config.TileSize)*6 {
				t.Fatalf("deploy pos too far from player: dist=%.1f player=%+v skiffCenter=(%.1f,%.1f)", dist, tc.pos, cx, cy)
			}
		})
	}
}

func landTileNearWater(g *Game) gvec.Vec2 {
	w := g.world
	// Prefer a land tile adjacent to water near the spawn/lifepod area.
	baseTX := int(g.baseStation.Pos.X) / config.TileSize
	baseTY := int(g.baseStation.Pos.Y) / config.TileSize
	for r := 0; r < 40; r++ {
		for dx := -r; dx <= r; dx++ {
			for dy := -r; dy <= r; dy++ {
				if abs(dx) != r && abs(dy) != r {
					continue
				}
				tx, ty := baseTX+dx, baseTY+dy
				if tx < 1 || ty < 1 || tx >= w.Width-1 || ty >= w.Height-1 {
					continue
				}
				if w.OverworldMap[tx][ty] != world.TileLand {
					continue
				}
				for _, n := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := tx+n[0], ty+n[1]
					info := world.GetTileInfo(w.OverworldMap[nx][ny])
					if info != nil && info.IsWater {
						return gvec.Vec2{
							X: float64(tx*config.TileSize) + float64(config.TileSize)/2,
							Y: float64(ty*config.TileSize) + float64(config.TileSize)/2,
						}
					}
				}
			}
		}
	}
	// Fallback: force a land tile next to known water spawn.
	sx, sy := findWaterSpawn(w)
	stx := int(sx) / config.TileSize
	sty := int(sy) / config.TileSize
	w.OverworldMap[stx+1][sty] = world.TileLand
	return gvec.Vec2{
		X: float64((stx+1)*config.TileSize) + float64(config.TileSize)/2,
		Y: float64(sty*config.TileSize) + float64(config.TileSize)/2,
	}
}

func hypot(a, b float64) float64 {
	return gvec.Vec2{X: a, Y: b}.Length()
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
