package scene

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// CosmeticFish represents an individual fish swimming near shorelines.
type CosmeticFish struct {
	Pos       gvec.Vec2
	Vel       gvec.Vec2
	Angle     float64
	BasePos   gvec.Vec2
	WobbleVal float64
	WobbleSpd float64
}

// Update updates the fish position, flocking behavior, and fleeing from the player.
func (f *CosmeticFish) Update(playerCenter gvec.Vec2, isSolid func(x, y float64) bool) {
	dx := f.Pos.X - playerCenter.X
	dy := f.Pos.Y - playerCenter.Y
	dist := math.Hypot(dx, dy)

	f.WobbleVal += f.WobbleSpd

	if dist < 120.0 {
		// Flee rapidly from the player
		angle := math.Atan2(dy, dx)
		targetVelX := math.Cos(angle) * 3.5
		targetVelY := math.Sin(angle) * 3.5
		f.Vel.X = f.Vel.X*0.88 + targetVelX*0.12
		f.Vel.Y = f.Vel.Y*0.88 + targetVelY*0.12
	} else {
		// Wander slowly around its base position
		bdx := f.BasePos.X - f.Pos.X
		bdy := f.BasePos.Y - f.Pos.Y
		bdist := math.Hypot(bdx, bdy)

		var wanderX, wanderY float64
		if bdist > 96.0 {
			// Steer back towards home/base position
			bangle := math.Atan2(bdy, bdx)
			wanderX = math.Cos(bangle) * 0.8
			wanderY = math.Sin(bangle) * 0.8
		} else {
			// Drift/swim gently using a sine wobble
			wanderX = math.Cos(f.WobbleVal) * 0.5
			wanderY = math.Sin(f.WobbleVal*0.5) * 0.5
		}

		f.Vel.X = f.Vel.X*0.94 + wanderX*0.06
		f.Vel.Y = f.Vel.Y*0.94 + wanderY*0.06
	}

	// Move and handle solid/land check separately for X and Y
	newX := f.Pos.X + f.Vel.X
	if isSolid(newX, f.Pos.Y) {
		f.Vel.X = -f.Vel.X * 0.5
	} else {
		f.Pos.X = newX
	}

	newY := f.Pos.Y + f.Vel.Y
	if isSolid(f.Pos.X, newY) {
		f.Vel.Y = -f.Vel.Y * 0.5
	} else {
		f.Pos.Y = newY
	}

	if f.Vel.Length() > 0.05 {
		f.Angle = math.Atan2(f.Vel.Y, f.Vel.X)
	}
}

// Draw renders a tiny procedurally animated fish.
func (f *CosmeticFish) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, mult float64) {
	sx := float32(f.Pos.X - camX)
	sy := float32(f.Pos.Y - camY)

	// Length and width of fish body
	bodyLen := 7.0
	bodyW := 3.5

	cos := math.Cos(f.Angle)
	sin := math.Sin(f.Angle)

	// Animated tail wiggle based on speed and time
	wiggleSpd := 0.22
	if f.Vel.Length() > 1.5 {
		wiggleSpd = 0.45
	}
	wiggle := math.Sin(ticks*wiggleSpd+f.WobbleVal) * 0.5
	tailCos := math.Cos(f.Angle + math.Pi + wiggle)
	tailSin := math.Sin(f.Angle + math.Pi + wiggle)

	// Coordinates for the triangular/diamond fish body
	tipX := sx + float32(bodyLen*0.5*cos)
	tipY := sy + float32(bodyLen*0.5*sin)

	blX := sx + float32(-bodyLen*0.5*cos+bodyW*0.5*-sin)
	blY := sy + float32(-bodyLen*0.5*sin+bodyW*0.5*cos)

	brX := sx + float32(-bodyLen*0.5*cos-bodyW*0.5*-sin)
	brY := sy + float32(-bodyLen*0.5*sin-bodyW*0.5*cos)

	fishColor := color.RGBA{110, 190, 220, 180}
	fishColor = applyLight(fishColor, mult)

	var path vector.Path
	path.MoveTo(tipX, tipY)
	path.LineTo(blX, blY)
	path.LineTo(brX, brY)
	path.Close()

	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(fishColor)
	vector.FillPath(screen, &path, nil, &opts)

	// Triangular wiggling tail fin
	tailBaseX := sx + float32(-bodyLen*0.5*cos)
	tailBaseY := sy + float32(-bodyLen*0.5*sin)
	tailTipL_X := tailBaseX + float32(4.5*tailCos+2.2*-tailSin)
	tailTipL_Y := tailBaseY + float32(4.5*tailSin+2.2*tailCos)
	tailTipR_X := tailBaseX + float32(4.5*tailCos-2.2*-tailSin)
	tailTipR_Y := tailBaseY + float32(4.5*tailSin-2.2*tailCos)

	var tailPath vector.Path
	tailPath.MoveTo(tailBaseX, tailBaseY)
	tailPath.LineTo(tailTipL_X, tailTipL_Y)
	tailPath.LineTo(tailTipR_X, tailTipR_Y)
	tailPath.Close()
	vector.FillPath(screen, &tailPath, nil, &opts)
}

// FloatingCrate represents a lootable cargo crate near shipwrecks.
type FloatingCrate struct {
	Pos          gvec.Vec2
	InitialPos   gvec.Vec2
	BobOffset    float64
	Collected    bool
	RespawnTimer int
}

// Draw renders a wooden cargo crate with a cross board pattern.
func (c *FloatingCrate) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, mult float64) {
	if c.Collected {
		return
	}

	bob := math.Sin(ticks*0.045+c.BobOffset) * 2.0
	sx := float32(c.Pos.X - camX)
	sy := float32(c.Pos.Y - camY + bob)

	const size = 15.0
	const half = size / 2.0

	// Dark wood background
	fillClr := applyLight(color.RGBA{135, 85, 40, 255}, mult)
	vector.FillRect(screen, sx-half, sy-half, size, size, fillClr, false)

	// Lighter wood border & planks cross lines
	strokeClr := applyLight(color.RGBA{200, 130, 60, 255}, mult)
	vector.StrokeRect(screen, sx-half, sy-half, size, size, 1.2, strokeClr, false)
	vector.StrokeLine(screen, sx-half+2, sy-half+2, sx+half-2, sy+half-2, 1.0, strokeClr, false)
	vector.StrokeLine(screen, sx+half-2, sy-half+2, sx-half+2, sy+half-2, 1.0, strokeClr, false)

	// Specular highlight pixel on the top edge
	highlightClr := applyLight(color.RGBA{255, 255, 255, 120}, mult)
	vector.FillRect(screen, sx-half+2, sy-half+2, 1.5, 1.5, highlightClr, false)
}

// ThermalVent represents a volcanic/hydrothermal vent dealing damage & pushing things away.
type ThermalVent struct {
	Pos            gvec.Vec2
	Radius         float64
	BubbleCooldown int
}

// Update ticks the thermal vent bubble particle spawn rate.
func (v *ThermalVent) Update(g GameContext, targetCenter gvec.Vec2, targetDims gvec.Vec2, isPiloting bool) {
	// Spawn rising bubble particles inside the mouth of the vent
	v.BubbleCooldown--
	if v.BubbleCooldown <= 0 {
		v.BubbleCooldown = rand.Intn(12) + 8
		angle := rand.Float64() * 2.0 * math.Pi
		dist := rand.Float64() * 12.0
		bx := v.Pos.X + math.Cos(angle)*dist
		by := v.Pos.Y + math.Sin(angle)*dist
		g.SpawnBubble(bx, by)
	}

	// Calculate distance and check influence radius
	dx := targetCenter.X - v.Pos.X
	dy := targetCenter.Y - v.Pos.Y
	dist := math.Hypot(dx, dy)

	if dist < v.Radius {
		// Calculate outward push force (stronger closer to center)
		ratio := 1.0 - (dist / v.Radius)
		pushStrength := 1.6 * ratio

		var pushX, pushY float64
		if dist > 0.1 {
			pushX = (dx / dist) * pushStrength
			pushY = (dy / dist) * pushStrength
		} else {
			pushX = pushStrength
			pushY = 0
		}

		if isPiloting {
			activeVeh := g.GetActiveVehicle()
			// Apply physical force to the vehicle
			activeVeh.ApplyForce(gvec.Vec2{X: pushX * 0.35, Y: pushY * 0.35})
			// Deal continuous low structural damage to the vehicle
			activeVeh.TakeDamage(0.04)
		} else {
			// Apply force to player velocity
			p := g.GetPlayer()
			p.Vel.X += pushX
			p.Vel.Y += pushY
			// Deal damage to swimming player
			p.CurrentHealth -= 0.10
		}

		// Trigger visual screen shake if very close
		if dist < 30.0 {
			g.TriggerScreenShake(1, 0.4)
		}
	}
}

// Draw renders a glowing circular volcanic mouth on the seafloor.
func (v *ThermalVent) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, mult float64) {
	sx := float32(v.Pos.X - camX)
	sy := float32(v.Pos.Y - camY)

	// Pulse outer radius slightly
	pulse := float32(math.Sin(ticks*0.05)) * 1.8
	outerRad := float32(22.0) + pulse

	// Draw glowing concentric magma-like rings
	glowRed := applyLight(color.RGBA{175, 45, 15, 160}, mult)
	glowOrange := applyLight(color.RGBA{215, 110, 25, 220}, mult)
	glowYellow := applyLight(color.RGBA{240, 200, 45, 255}, mult)
	abyssBlack := applyLight(color.RGBA{14, 6, 8, 255}, mult)

	vector.FillCircle(screen, sx, sy, outerRad, glowRed, false)
	vector.FillCircle(screen, sx, sy, outerRad-4.0, glowOrange, false)
	vector.FillCircle(screen, sx, sy, outerRad-8.0, glowYellow, false)
	vector.FillCircle(screen, sx, sy, outerRad-12.0, abyssBlack, false)
}

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
							o.crates = append(o.crates, &FloatingCrate{
								Pos:        gvec.Vec2{X: px, Y: py},
								InitialPos: gvec.Vec2{X: px, Y: py},
								BobOffset:  r.Float64() * 100.0,
							})
						}
					}
				}

			case world.TileTrench:
				// Spawn a geothermal vent near the trench center
				offsetX := (r.Float64() - 0.5) * 128.0
				offsetY := (r.Float64() - 0.5) * 128.0
				px := float64(tx*config.TileSize) + float64(config.TileSize)/2.0 + offsetX
				py := float64(ty*config.TileSize) + float64(config.TileSize)/2.0 + offsetY

				gtx := int(px) / config.TileSize
				gty := int(py) / config.TileSize
				if gtx >= 0 && gtx < o.World.Width && gty >= 0 && gty < o.World.Height {
					if o.World.OverworldMap[gtx][gty] != world.TileLand {
						o.vents = append(o.vents, &ThermalVent{
							Pos:    gvec.Vec2{X: px, Y: py},
							Radius: 70.0,
						})
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

							o.fish = append(o.fish, &CosmeticFish{
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
}

// UpdateExtras runs update logic for cosmetic fish, crates, and vents.
func (o *OverworldScene) UpdateExtras(g GameContext) {
	o.InitializeExtras(g)

	p := g.GetPlayer()
	pPos := p.Pos
	pDims := gvec.Vec2{X: p.Width, Y: p.Height}
	pCenter := gvec.Vec2{X: pPos.X + pDims.X/2.0, Y: pPos.Y + pDims.Y/2.0}

	isPiloting := false
	var targetCenter gvec.Vec2
	var targetDims gvec.Vec2

	activeVeh := g.GetActiveVehicle()
	if activeVeh != nil {
		isPiloting = true
		vPos := activeVeh.GetPos()
		targetDims = activeVeh.GetDimensions()
		targetCenter = gvec.Vec2{X: vPos.X + targetDims.X/2.0, Y: vPos.Y + targetDims.Y/2.0}
	} else {
		targetCenter = pCenter
		targetDims = pDims
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
		v.Update(g, targetCenter, targetDims, isPiloting)
	}

	// Update crates (collisions, looting, respawn timer)
	for _, c := range o.crates {
		if c.Collected {
			c.RespawnTimer--
			if c.RespawnTimer <= 0 {
				distToPlayer := math.Hypot(targetCenter.X-c.InitialPos.X, targetCenter.Y-c.InitialPos.Y)
				if distToPlayer > 400.0 {
					c.Collected = false
					c.Pos = c.InitialPos
				} else {
					c.RespawnTimer = 60
				}
			}
			continue
		}

		// Proximity check for collection
		dist := math.Hypot(targetCenter.X-c.Pos.X, targetCenter.Y-c.Pos.Y)
		colRadius := 16.0 + math.Max(targetDims.X, targetDims.Y)/2.0
		if dist < colRadius {
			// Loot generation
			r := rand.Float64()
			var loot item.Item
			if r < 0.50 {
				loot = &item.ScrapMetal{}
			} else if r < 0.75 {
				loot = &item.Titanium{}
			} else if r < 0.90 {
				loot = &item.Copper{}
			} else if r < 0.96 {
				loot = &item.ElectronicWaste{}
			} else {
				loot = &item.PowerCell{}
			}

			added := false
			if isPiloting {
				added = activeVeh.GetCargo().AddItem(loot, 1)
			} else {
				added = p.Inventory.AddItem(loot, 1)
			}

			if added {
				c.Collected = true
				c.RespawnTimer = 7200 // 2 minutes (120 seconds * 60 ticks)

				// FX
				g.SpawnDebris(c.Pos.X, c.Pos.Y, color.RGBA{139, 90, 43, 255})
				g.TriggerScreenShake(10, 1.5)
				g.SetMineWarning("Salvaged: "+loot.GetName()+"!", 120)

				if !isPiloting {
					p.RecalculateUpgrades()
				}
			} else {
				g.SetMineWarning("Inventory full! Cannot salvage crate.", 60)
			}
		}
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
