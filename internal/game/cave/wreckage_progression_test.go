package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

func TestTieredWreckageProgression(t *testing.T) {
	seeds := []int64{1, 42, 100, 777, 9999}

	for _, seed := range seeds {
		r := rand.New(rand.NewSource(seed))
		grid := GenerateWreckageGrid(r)

		// Test Ship 0
		cave0 := NewWreckageCorridorCave(grid, 0)
		ents0 := cave0.GenerateEntities(seed)
		res0 := cave0.GenerateResources(seed)

		for _, e := range ents0 {
			if sb, ok := e.(*entity.ShatterBulb); ok {
				ty := int(sb.Pos.Y) / config.TileSize
				if ty >= 40 {
					t.Errorf("seed %d Ship 0 spawned ShatterBulb below 40m: ty=%d", seed, ty)
				}
			}
		}

		hasScoutSub := false
		for _, n := range res0 {
			if rn, ok := n.(*resource.ResourceNode); ok && rn.Type == resource.NodeBlueprint {
				if rn.RecipeResultName == "Scout Sub Kit" {
					hasScoutSub = true
					_, ty := rn.GetTilePos()
					if ty >= 40 {
						t.Errorf("seed %d Ship 0 Scout Sub spawned below 40m: ty=%d", seed, ty)
					}
					if ty < 30 {
						t.Errorf("seed %d Ship 0 Scout Sub spawned too shallow: ty=%d (expected closer to 40m, in [30, 39])", seed, ty)
					}
				}
			}
		}
		if !hasScoutSub {
			t.Errorf("seed %d Ship 0 did not spawn guaranteed Scout Sub Kit", seed)
		}

		// Test Ship 1
		cave1 := NewWreckageCorridorCave(grid, 1)
		ents1 := cave1.GenerateEntities(seed)
		res1 := cave1.GenerateResources(seed)

		for _, e := range ents1 {
			if sb, ok := e.(*entity.ShatterBulb); ok {
				ty := int(sb.Pos.Y) / config.TileSize
				if ty >= 40 {
					t.Errorf("seed %d Ship 1 spawned ShatterBulb below 40m: ty=%d", seed, ty)
				}
			}
		}

		hasHeavyMech := false
		hasDepthModule := false
		for _, n := range res1 {
			if rn, ok := n.(*resource.ResourceNode); ok && rn.Type == resource.NodeBlueprint {
				if rn.RecipeResultName == "Heavy Mech Kit" {
					hasHeavyMech = true
					_, ty := rn.GetTilePos()
					if ty <= 51 {
						t.Errorf("seed %d Ship 1 Heavy Mech spawned on upper deck: ty=%d", seed, ty)
					}
				}
				if rn.RecipeResultName == "Scout Sub Depth Module MK1" {
					hasDepthModule = true
					_, ty := rn.GetTilePos()
					if ty >= 40 {
						t.Errorf("seed %d Ship 1 Depth Module spawned below 40m: ty=%d", seed, ty)
					}
				}
			}
		}
		if !hasHeavyMech {
			t.Errorf("seed %d Ship 1 did not spawn guaranteed Heavy Mech Kit", seed)
		}
		if !hasDepthModule {
			t.Errorf("seed %d Ship 1 did not spawn guaranteed Scout Sub Depth Module MK1", seed)
		}

		// Test Ship 2
		cave2 := NewWreckageCorridorCave(grid, 2)
		ents2 := cave2.GenerateEntities(seed)
		res2 := cave2.GenerateResources(seed)

		for _, e := range ents2 {
			if sb, ok := e.(*entity.ShatterBulb); ok {
				ty := int(sb.Pos.Y) / config.TileSize
				if ty >= 40 {
					t.Errorf("seed %d Ship 2 spawned ShatterBulb below 40m: ty=%d", seed, ty)
				}
			}
		}

		hasEscapeRocket := false
		hasBulkhead := false
		for _, n := range res2 {
			if rn, ok := n.(*resource.ResourceNode); ok {
				if rn.Type == resource.NodeBlueprint && rn.RecipeResultName == "Escape Rocket" {
					hasEscapeRocket = true
					tx, ty := rn.GetTilePos()
					t.Logf("seed %d Escape Rocket at (%d, %d)", seed, tx, ty)
					if ty <= 51 {
						t.Errorf("seed %d Ship 2 Escape Rocket spawned on upper deck: ty=%d", seed, ty)
					}
				}
				if rn.Type == resource.NodeReinforcedBulkhead {
					hasBulkhead = true
					tx, ty := rn.GetTilePos()
					t.Logf("seed %d Bulkhead at (%d, %d)", seed, tx, ty)
					if !rn.RequiresMech() {
						t.Errorf("seed %d Ship 2 Bulkhead should require Mech", seed)
					}
				}
			}
		}
		if !hasEscapeRocket {
			t.Errorf("seed %d Ship 2 did not spawn guaranteed Escape Rocket", seed)
		}
		if !hasBulkhead {
			t.Errorf("seed %d Ship 2 did not spawn Reinforced Blast Bulkhead", seed)
		}
	}
}

func TestWreckageCaveReachability(t *testing.T) {
	for seed := int64(1); seed <= 100; seed++ {
		r := rand.New(rand.NewSource(seed))
		grid := GenerateWreckageGrid(r)
		w := len(grid)
		h := len(grid[0])

		visited := make([][]bool, w)
		for x := range w {
			visited[x] = make([]bool, h)
		}

		// BFS from entrance (30, 2)
		type pt struct{ x, y int }
		queue := []pt{{30, 2}}
		visited[30][2] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
			for _, d := range dirs {
				nx, ny := curr.x+d[0], curr.y+d[1]
				if nx >= 0 && nx < w && ny >= 0 && ny < h && !grid[nx][ny] && !visited[nx][ny] {
					visited[nx][ny] = true
					queue = append(queue, pt{nx, ny})
				}
			}
		}

		// Count unreachable open tiles
		unreachableCount := 0
		var sampleUnreachable pt
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				if !grid[x][y] && !visited[x][y] {
					if unreachableCount == 0 {
						sampleUnreachable = pt{x, y}
					}
					unreachableCount++
				}
			}
		}
		if unreachableCount > 0 {
			t.Errorf("seed %d: found %d unreachable open tiles! Sample at (%d, %d)", seed, unreachableCount, sampleUnreachable.x, sampleUnreachable.y)
		}
	}
}

func TestShip2EscapeRocketBlockedByBulkhead(t *testing.T) {
	for seed := int64(1); seed <= 100; seed++ {
		r := rand.New(rand.NewSource(seed))
		grid := GenerateWreckageGrid(r)
		w := len(grid)
		h := len(grid[0])

		cave2 := NewWreckageCorridorCave(grid, 2)
		res2 := cave2.GenerateResources(seed)

		var rocketPos [2]int
		var bulkheads [][2]int
		hasRocket := false

		for _, n := range res2 {
			if rn, ok := n.(*resource.ResourceNode); ok {
				if rn.Type == resource.NodeBlueprint && rn.RecipeResultName == "Escape Rocket" {
					hasRocket = true
					rocketPos[0], rocketPos[1] = rn.GetTilePos()
				}
				if rn.Type == resource.NodeReinforcedBulkhead {
					x, y := rn.GetTilePos()
					bulkheads = append(bulkheads, [2]int{x, y})
				}
			}
		}

		if !hasRocket {
			t.Fatalf("seed %d: Escape Rocket did not spawn", seed)
		}
		if len(bulkheads) == 0 {
			t.Fatalf("seed %d: No bulkheads spawned in Ship 2", seed)
		}

		// Verify 1: Escape Rocket is in the deep vault (Y >= 80)
		if rocketPos[1] < 80 {
			t.Errorf("seed %d: Escape Rocket spawned too shallow: Y = %d", seed, rocketPos[1])
		}

		// Verify 2: With bulkheads impassable (solid), BFS from entrance (30, 2) CANNOT reach the Escape Rocket!
		isSolidWithBulkhead := func(x, y int) bool {
			if x < 0 || x >= w || y < 0 || y >= h {
				return true
			}
			if grid[x][y] {
				return true
			}
			for _, bh := range bulkheads {
				if bh[0] == x && bh[1] == y {
					return true
				}
			}
			return false
		}

		visited := make([][]bool, w)
		for x := range visited {
			visited[x] = make([]bool, h)
		}

		type pt struct{ x, y int }
		queue := []pt{{30, 2}}
		visited[30][2] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
			for _, d := range dirs {
				nx, ny := curr.x+d[0], curr.y+d[1]
				if nx >= 0 && nx < w && ny >= 0 && ny < h && !isSolidWithBulkhead(nx, ny) && !visited[nx][ny] {
					visited[nx][ny] = true
					queue = append(queue, pt{nx, ny})
				}
			}
		}

		if visited[rocketPos[0]][rocketPos[1]] {
			t.Fatalf("seed %d: Escape Rocket at (%d, %d) is REACHABLE WITHOUT BREAKING BULKHEAD! Bulkheads at %v",
				seed, rocketPos[0], rocketPos[1], bulkheads)
		}

		// Verify 3: When bulkheads are destroyed (passable), BFS from entrance DOES reach the Escape Rocket!
		for x := range visited {
			for y := range visited[x] {
				visited[x][y] = false
			}
		}
		queue = []pt{{30, 2}}
		visited[30][2] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
			for _, d := range dirs {
				nx, ny := curr.x+d[0], curr.y+d[1]
				if nx >= 0 && nx < w && ny >= 0 && ny < h && !grid[nx][ny] && !visited[nx][ny] {
					visited[nx][ny] = true
					queue = append(queue, pt{nx, ny})
				}
			}
		}

		if !visited[rocketPos[0]][rocketPos[1]] {
			t.Fatalf("seed %d: Escape Rocket at (%d, %d) is UNREACHABLE even after breaking bulkheads!",
				seed, rocketPos[0], rocketPos[1])
		}
	}
}


