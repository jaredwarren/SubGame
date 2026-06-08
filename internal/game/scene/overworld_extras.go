package scene

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// InitializeExtras populates the overworld with fish, crates, and vents if not already initialized.
func (o *OverworldScene) InitializeExtras(g GameContext) {
	if o.initialized {
		return
	}
	o.initialized = true

	r := rand.New(rand.NewSource(o.World.Seed + 12345))

	// Scan the overworld map to place extras
	for tx := 0; tx < o.World.Width; tx++ {
		for ty := 0; ty < o.World.Height; ty++ {
			tile := o.World.OverworldMap[tx][ty]

			switch tile {
			case world.TileWreckage:
				// Spawn 2-3 cargo crates around each wreckage tile (within a radius of 2-3 tiles)
				numCrates := r.Intn(2) + 2
				for i := 0; i < numCrates; i++ {
					offsetX := (r.Float64() - 0.5) * 192.0
					offsetY := (r.Float64() - 0.5) * 192.0

					px := float64(tx*config.TileSize) + float64(config.TileSize)/2.0 + offsetX
					py := float64(ty*config.TileSize) + float64(config.TileSize)/2.0 + offsetY

					gtx := int(px) / config.TileSize
					gty := int(py) / config.TileSize
					if gtx >= 0 && gtx < o.World.Width && gty >= 0 && gty < o.World.Height {
						if o.World.OverworldMap[gtx][gty] != world.TileLand {
							o.crates = append(o.crates, &oe.FloatingCrate{
								Pos:        gvec.Vec2{X: px, Y: py},
								InitialPos: gvec.Vec2{X: px, Y: py},
								BobOffset:  r.Float64() * 100.0,
							})
						}
					}
				}

			case world.TileWater:
				// If this tile is close to land, spawn a school of fish
				dist := o.World.LandDist[tx][ty]
				if dist >= 1 && dist <= 3 {
					// 4% chance per eligible tile to spawn a school of fish
					if r.Float64() < 0.04 {
						schoolSize := r.Intn(3) + 3
						schoolBaseX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
						schoolBaseY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0

						for i := 0; i < schoolSize; i++ {
							fx := schoolBaseX + (r.Float64()-0.5)*32.0
							fy := schoolBaseY + (r.Float64()-0.5)*32.0

							o.fish = append(o.fish, &oe.CosmeticFish{
								Pos:       gvec.Vec2{X: fx, Y: fy},
								BasePos:   gvec.Vec2{X: schoolBaseX, Y: schoolBaseY},
								WobbleVal: r.Float64() * 100.0,
								WobbleSpd: r.Float64()*0.05 + 0.02,
							})
						}
					}
				}
			}
		}
	}

	// Spawn random vents in open waters far from the base station
	var candidates []gvec.Vec2
	basePos := g.GetBaseStation().Pos
	for tx := 0; tx < o.World.Width; tx++ {
		for ty := 0; ty < o.World.Height; ty++ {
			if o.World.OverworldMap[tx][ty] == world.TileWater && o.World.LandDist[tx][ty] >= 3 {
				tileX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
				tileY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0
				dist := math.Hypot(tileX-basePos.X, tileY-basePos.Y)
				if dist >= 960.0 {
					candidates = append(candidates, gvec.Vec2{X: tileX, Y: tileY})
				}
			}
		}
	}

	if len(candidates) < 6 {
		// Fallback: check any water tile further than 400px from base station
		for tx := 0; tx < o.World.Width; tx++ {
			for ty := 0; ty < o.World.Height; ty++ {
				if o.World.OverworldMap[tx][ty] == world.TileWater {
					tileX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
					tileY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0
					dist := math.Hypot(tileX-basePos.X, tileY-basePos.Y)
					if dist >= 400.0 {
						candidates = append(candidates, gvec.Vec2{X: tileX, Y: tileY})
					}
				}
			}
		}
	}

	if len(candidates) > 0 {
		// Shuffle candidates using deterministic local PRNG
		r.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		numVents := 6
		if len(candidates) < numVents {
			numVents = len(candidates)
		}

		for i := 0; i < numVents; i++ {
			pos := candidates[i]
			// Initialize with randomized state timer so they don't all erupt at once
			o.vents = append(o.vents, oe.NewThermalVent(pos, int64(i*12345)))
		}
	}
}

// UpdateExtras runs update logic for cosmetic fish, crates, and vents.
func (o *OverworldScene) UpdateExtras(g GameContext) {
	o.InitializeExtras(g)

	p := g.GetPlayer()
	pPos := p.Pos
	pDims := gvec.Vec2{X: p.Width, Y: p.Height}
	pCenter := gvec.Vec2{X: pPos.X + pDims.X/2.0, Y: pPos.Y + pDims.Y/2.0}

	var targetCenter gvec.Vec2

	activeVeh := g.GetActiveVehicle()
	if activeVeh != nil {
		vPos := activeVeh.GetPos()
		vDims := activeVeh.GetDimensions()
		targetCenter = gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
	} else {
		targetCenter = pCenter
	}

	isSolid := func(x, y float64) bool {
		return o.IsSolid(x-2, y-2, 4, 4)
	}

	// Update cosmetic fish
	for _, f := range o.fish {
		f.Update(targetCenter, isSolid)
	}

	// Update vents (push force, heat damage, bubble particle spawning)
	for _, v := range o.vents {
		v.Update(g)
	}

	// Update crates (collisions, looting, respawn timer)
	for _, c := range o.crates {
		c.Update(g)
	}
}

// DrawExtras draws cosmetic fish, crates, and vents.
func (o *OverworldScene) DrawExtras(g GameContext, screen *ebiten.Image) {
	cam := g.GetCamera()
	camX, camY := cam.Pos.X, cam.Pos.Y
	ticks := g.GetTicks()
	mult := GetOverworldLightMultiplier(g.GetTimeOfDay())

	// Draw vents (on the bottom layer)
	for _, v := range o.vents {
		if v.Pos.X >= camX-100 && v.Pos.X <= camX+float64(config.ScreenWidth)+100 &&
			v.Pos.Y >= camY-100 && v.Pos.Y <= camY+float64(config.ScreenHeight)+100 {
			v.Draw(screen, camX, camY, ticks, mult)
		}
	}

	// Draw crates
	for _, c := range o.crates {
		if !c.Collected && c.Pos.X >= camX-50 && c.Pos.X <= camX+float64(config.ScreenWidth)+50 &&
			c.Pos.Y >= camY-50 && c.Pos.Y <= camY+float64(config.ScreenHeight)+50 {
			c.Draw(screen, camX, camY, ticks, mult)
		}
	}

	// Draw fish
	for _, f := range o.fish {
		if f.Pos.X >= camX-50 && f.Pos.X <= camX+float64(config.ScreenWidth)+50 &&
			f.Pos.Y >= camY-50 && f.Pos.Y <= camY+float64(config.ScreenHeight)+50 {
			f.Draw(screen, camX, camY, ticks, mult)
		}
	}
}
