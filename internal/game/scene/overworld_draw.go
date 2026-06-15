package scene

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// draw renders the overworld tiles in a viewport centered on the player.
func (o *OverworldScene) draw(g OverworldContext, screen *ebiten.Image) {
	cam := g.GetCamera()
	isPiloting := g.GetActiveVehicle() != nil
	p := g.GetPlayer()

	camX := cam.Pos.X
	camY := cam.Pos.Y

	startTileX := tileAt(camX, config.TileSize)
	endTileX := tileAt(camX+float64(config.ScreenWidth), config.TileSize) + 1
	startTileY := tileAt(camY, config.TileSize)
	endTileY := tileAt(camY+float64(config.ScreenHeight), config.TileSize) + 1

	// Check/update cached static decorations offscreen image
	chunkX := floorDiv(startTileX, 10)
	chunkY := floorDiv(startTileY, 10)

	if o.cachedStaticImage == nil {
		o.cachedStaticImage = ebiten.NewImage(32*config.TileSize, 24*config.TileSize)
	}

	minTileX := chunkX * 10
	maxTileX := chunkX*10 + 32
	minTileY := chunkY * 10
	maxTileY := chunkY*10 + 24

	if !o.hasCache || o.cachedChunkX != chunkX || o.cachedChunkY != chunkY {
		o.cachedStaticImage.Clear()
		o.drawBaseTiles(o.cachedStaticImage, minTileX, maxTileX, minTileY, maxTileY, float64(minTileX*config.TileSize), float64(minTileY*config.TileSize), 1.0)
		o.drawLandDecoration(o.cachedStaticImage, minTileX, maxTileX, minTileY, maxTileY, float64(minTileX*config.TileSize), float64(minTileY*config.TileSize), 1.0)
		o.cachedChunkX = chunkX
		o.cachedChunkY = chunkY
		o.hasCache = true
	}

	// 1. Draw cached static background to screen
	op := &ebiten.DrawImageOptions{}
	dx := float64(minTileX*config.TileSize) - camX
	dy := float64(minTileY*config.TileSize) - camY
	op.GeoM.Translate(dx, dy)

	mult := GetOverworldLightMultiplier(g.GetTimeOfDay())
	op.ColorScale.Scale(float32(mult), float32(mult), float32(mult), 1.0)
	screen.DrawImage(o.cachedStaticImage, op)

	// 2. Draw time-dependent elements: waves and vehicle beacons
	o.drawWaves(screen, startTileX, endTileX, startTileY, endTileY, g, camX, camY)
	o.drawVehicleBeacons(screen, startTileX, endTileX, startTileY, endTileY, g, camX, camY)

	// 3. Draw extras, player, prompts, etc.
	o.DrawExtras(g, screen)

	if o.whirlpool != nil {
		o.whirlpool.Draw(screen, camX, camY)
	}

	if !isPiloting {
		// Check dormant thermal vents proximity for custom diving prompt
		pCenter := gvec.Vec2{X: p.Pos.X + p.Width/2.0, Y: p.Pos.Y + p.Height/2.0}
		var nearVent *oe.ThermalVent
		for _, v := range o.vents {
			dist := math.Hypot(pCenter.X-v.Pos.X, pCenter.Y-v.Pos.Y)
			if dist < 30.0 {
				nearVent = v
				break
			}
		}

		if nearVent != nil && nearVent.State == oe.VentDormant {
			promptText := "Press [E] to Enter Thermo Vent"
			promptX := float32(p.CenterX()) - 95
			promptY := float32(p.CenterY()) - 40
			vector.FillRect(screen, promptX, promptY, 190, 25, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, promptText, int(promptX)+10, int(promptY)+4)
		} else {
			pTileX := tileAt(p.Pos.X+p.Width/2.0, config.TileSize)
			pTileY := tileAt(p.Pos.Y+p.Height/2.0, config.TileSize)
			if pTileX < 0 || pTileX >= o.World.Width || pTileY < 0 || pTileY >= o.World.Height {
				promptText := "Press [E] to Dive into Void"
				promptX := float32(p.CenterX()) - 95
				promptY := float32(p.CenterY()) - 40
				vector.FillRect(screen, promptX, promptY, 190, 25, color.RGBA{0, 0, 0, 180}, false)
				ebitenutil.DebugPrintAt(screen, promptText, int(promptX)+10, int(promptY)+4)
			} else {
				tile := o.World.OverworldMap[pTileX][pTileY]
				info := world.GetTileInfo(tile)
				if info != nil && info.IsDiveable {
					promptText := "Press [E] to Dive"
					if info.DivePrompt != "" {
						promptText = info.DivePrompt
					}
					promptX := float32(p.CenterX()) - 95
					promptY := float32(p.CenterY()) - 40
					vector.FillRect(screen, promptX, promptY, 190, 25, color.RGBA{0, 0, 0, 180}, false)
					ebitenutil.DebugPrintAt(screen, promptText, int(promptX)+10, int(promptY)+4)
				}
			}
		}

		pX := float32(p.CenterX())
		pY := float32(p.CenterY())
		vector.FillCircle(screen, pX, pY, 8.0, color.RGBA{220, 95, 45, 255}, false)
		vx := pX + float32(math.Cos(p.Facing))*5
		vy := pY + float32(math.Sin(p.Facing))*5
		vector.FillCircle(screen, vx, vy, 4.0, color.RGBA{80, 200, 255, 200}, false)
	}
}

func (o *OverworldScene) drawBaseTiles(target *ebiten.Image, startTileX, endTileX, startTileY, endTileY int, offsetX, offsetY float64, mult float64) {
	op := &ebiten.DrawImageOptions{}

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			sx := float32(tx*config.TileSize) - float32(offsetX)
			sy := float32(ty*config.TileSize) - float32(offsetY)

			clampedTx := tx
			if clampedTx < 0 {
				clampedTx = 0
			} else if clampedTx >= o.World.Width {
				clampedTx = o.World.Width - 1
			}
			clampedTy := ty
			if clampedTy < 0 {
				clampedTy = 0
			} else if clampedTy >= o.World.Height {
				clampedTy = o.World.Height - 1
			}

			tileType := o.World.OverworldMap[clampedTx][clampedTy]
			landDist := o.World.LandDist[clampedTx][clampedTy]
			waterDist := o.World.WaterDist[clampedTx][clampedTy]

			tileClr, strokeClr := ComputeTileColors(tx, ty, tileType, landDist, waterDist, o.World.Width, o.World.Height, mult)

			// Always draw the base background tile (e.g. water color) first
			vector.FillRect(target, sx, sy, config.TileSize, config.TileSize, tileClr, false)
			vector.StrokeRect(target, sx, sy, config.TileSize, config.TileSize, 0.5, strokeClr, false)

			drawTexture := o.getTileTexture(tileType)
			if drawTexture != nil {
				op.GeoM.Reset()
				op.ColorScale.Reset()
				wImg, hImg := drawTexture.Bounds().Dx(), drawTexture.Bounds().Dy()
				if wImg > 0 && hImg > 0 {
					op.GeoM.Scale(float64(config.TileSize)/float64(wImg), float64(config.TileSize)/float64(hImg))
				}
				op.GeoM.Translate(float64(sx), float64(sy))

				// Apply fade and time-of-day lighting
				var distFromBorder float64
				if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
					dx := 0.0
					if tx < 0 {
						dx = float64(-tx)
					} else if tx >= o.World.Width {
						dx = float64(tx - o.World.Width + 1)
					}
					dy := 0.0
					if ty < 0 {
						dy = float64(-ty)
					} else if ty >= o.World.Height {
						dy = float64(ty - o.World.Height + 1)
					}
					distFromBorder = math.Max(dx, dy)
				} else {
					dxInside := float64(tx)
					if float64(o.World.Width-1-tx) < dxInside {
						dxInside = float64(o.World.Width - 1 - tx)
					}
					dyInside := float64(ty)
					if float64(o.World.Height-1-ty) < dyInside {
						dyInside = float64(o.World.Height - 1 - ty)
					}
					distFromBorder = -math.Min(dxInside, dyInside)
				}

				t := (distFromBorder + 1.0) / 4.0
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}

				totalMult := float32((1.0 - t) * mult)
				op.ColorScale.Scale(totalMult, totalMult, totalMult, 1.0)

				target.DrawImage(drawTexture, op)
			}
		}
	}
}

func (o *OverworldScene) drawLandDecoration(target *ebiten.Image, startTileX, endTileX, startTileY, endTileY int, offsetX, offsetY float64, mult float64) {
	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			sx := float32(tx*config.TileSize) - float32(offsetX)
			sy := float32(ty*config.TileSize) - float32(offsetY)

			clampedTx := tx
			if clampedTx < 0 {
				clampedTx = 0
			} else if clampedTx >= o.World.Width {
				clampedTx = o.World.Width - 1
			}
			clampedTy := ty
			if clampedTy < 0 {
				clampedTy = 0
			} else if clampedTy >= o.World.Height {
				clampedTy = o.World.Height - 1
			}

			// Draw procedurally generated trees, plants, and grass texture for land tiles
			if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height && o.World.OverworldMap[tx][ty] == world.TileLand {
				// Seed generator deterministically based on tile coords
				rngVal := hashCoords(clampedTx, clampedTy)
				if rngVal == 0 {
					rngVal = 1
				}
				rng := statelessRNG(rngVal)
				dist := o.World.WaterDist[clampedTx][clampedTy]
				isSand := dist == 1

				if isSand {
					// 1. Draw sand ripples (subtle darker lines)
					numRipples := rng.intn(2) + 1
					rippleClr := applyLight(color.RGBA{220, 200, 150, 255}, mult)
					for i := 0; i < numRipples; i++ {
						rx := float32(rng.float64()*40.0) + 12.0
						ry := float32(rng.float64()*40.0) + 12.0
						vector.StrokeLine(target, sx+rx, sy+ry, sx+rx+8, sy+ry+2, 1.0, rippleClr, false)
					}

					// 2. Draw occasional pebble or starfish (40% chance)
					if rng.float64() < 0.40 {
						px := float32(rng.float64()*44.0) + 10.0
						py := float32(rng.float64()*44.0) + 10.0

						if rng.float64() < 0.15 {
							// Rare starfish! (Orange cross/star)
							starClr := applyLight(color.RGBA{235, 110, 50, 255}, mult)
							vector.StrokeLine(target, sx+px-3, sy+py, sx+px+3, sy+py, 1.2, starClr, false)
							vector.StrokeLine(target, sx+px, sy+py-3, sx+px, sy+py+3, 1.2, starClr, false)
						} else {
							// Grey/white seashell/pebble
							pebbleClr := applyLight(color.RGBA{235, 230, 220, 255}, mult)
							vector.FillCircle(target, sx+px, sy+py, 2.0, pebbleClr, false)
							vector.StrokeCircle(target, sx+px, sy+py, 2.0, 0.8, applyLight(color.RGBA{180, 175, 165, 255}, mult), false)
						}
					}
				} else {
					// 1. Grass tufts/blades for texture
					numGrass := rng.intn(3) + 2
					grassClr := applyLight(color.RGBA{60, 195, 120, 255}, mult)
					for i := 0; i < numGrass; i++ {
						gx := float32(rng.float64()*50.0) + 7.0
						gy := float32(rng.float64()*50.0) + 7.0
						lenG := float32(rng.float64()*4.0 + 3.0)
						angleG := (rng.float64() - 0.5) * 0.4 // slight tilt

						// Draw two blades per tuft
						gx2 := gx + lenG*float32(math.Sin(float64(angleG)))
						gy2 := gy - lenG*float32(math.Cos(float64(angleG)))
						vector.StrokeLine(target, sx+gx, sy+gy, sx+gx2, sy+gy2, 1.0, grassClr, false)

						gx3 := gx + lenG*0.7*float32(math.Sin(float64(angleG+0.3)))
						gy3 := gy - lenG*0.7*float32(math.Cos(float64(angleG+0.3)))
						vector.StrokeLine(target, sx+gx, sy+gy, sx+gx3, sy+gy3, 1.0, grassClr, false)
					}

					// 2. Flowering plants (35% chance)
					if rng.float64() < 0.35 {
						px := float32(rng.float64()*40.0) + 12.0
						py := float32(rng.float64()*40.0) + 12.0

						// Stem
						stemClr := applyLight(color.RGBA{40, 150, 80, 255}, mult)
						vector.StrokeLine(target, sx+px, sy+py, sx+px, sy+py-5, 1.0, stemClr, false)

						// Flower petals (red, yellow, or blue)
						var petalClr color.RGBA
						switch rng.intn(3) {
						case 0:
							petalClr = color.RGBA{235, 80, 80, 255} // red
						case 1:
							petalClr = color.RGBA{240, 205, 45, 255} // yellow
						default:
							petalClr = color.RGBA{80, 160, 235, 255} // blue
						}
						petalClr = applyLight(petalClr, mult)

						// Small center dot
						vector.FillCircle(target, sx+px, sy+py-6, 2.0, petalClr, false)
					}

					// 3. Leafy Trees (22% chance)
					if rng.float64() < 0.22 {
						tx := float32(rng.float64()*24.0) + 20.0
						ty := float32(rng.float64()*24.0) + 20.0

						shadowClr := color.RGBA{4, 12, 8, 55}
						trunkClr := applyLight(color.RGBA{100, 65, 35, 255}, mult)
						canopyClr := applyLight(color.RGBA{24, 115, 62, 255}, mult)
						canopyStroke := applyLight(color.RGBA{18, 90, 50, 255}, mult)

						// Subtle canopy shadow underneath
						vector.FillCircle(target, sx+tx+3, sy+ty+6, 8.5, shadowClr, false)

						// Trunk
						vector.FillRect(target, sx+tx-2, sy+ty, 4, 8, trunkClr, false)

						// Main leafy canopy
						vector.FillCircle(target, sx+tx, sy+ty-4, 9.0, canopyClr, false)
						vector.StrokeCircle(target, sx+tx, sy+ty-4, 9.0, 1.0, canopyStroke, false)

						// Highlight on the top-left of the canopy
						highlightClr := applyLight(color.RGBA{65, 175, 110, 255}, mult)
						vector.FillCircle(target, sx+tx-3, sy+ty-7, 3.5, highlightClr, false)
					}
				}
			}
		}
	}
}

func (o *OverworldScene) drawWaves(screen *ebiten.Image, startTileX, endTileX, startTileY, endTileY int, g OverworldContext, camX, camY float64) {
	mult := GetOverworldLightMultiplier(g.GetTimeOfDay())
	ticks := g.GetTicks()

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			clampedTx := tx
			if clampedTx < 0 {
				clampedTx = 0
			} else if clampedTx >= o.World.Width {
				clampedTx = o.World.Width - 1
			}
			clampedTy := ty
			if clampedTy < 0 {
				clampedTy = 0
			} else if clampedTy >= o.World.Height {
				clampedTy = o.World.Height - 1
			}

			currTile := o.World.OverworldMap[clampedTx][clampedTy]
			currInfo := world.GetTileInfo(currTile)
			if currInfo != nil && currInfo.IsWater {
				if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height {
					// Only generate waves on a subset of water tiles (50% density)
					tileRngVal := hashCoords(tx, ty)
					if tileRngVal == 0 {
						tileRngVal = 1
					}
					tileRng := statelessRNG(tileRngVal)
					if tileRng.float64() < 0.5 {
						rngVal := hashCoords(tx, ty) ^ 0x5555555555555555
						if rngVal == 0 {
							rngVal = 1
						}
						rng := statelessRNG(rngVal)

						// Static base offset inside the tile
						baseOffsetX := rng.float64() * float64(config.TileSize)
						baseOffsetY := rng.float64() * float64(config.TileSize)

						// Life cycle configuration (140 ticks per cycle)
						const cycleLength = 140.0
						phaseOffset := rng.float64() * cycleLength
						cycleTime := math.Mod(ticks+phaseOffset, cycleLength)
						lifeFrac := cycleTime / cycleLength

						// Fade in and out
						opacity := math.Sin(lifeFrac * math.Pi)
						if opacity < 0 {
							opacity = 0
						}

						// Slow wind drift to the left and slightly down
						driftX := -lifeFrac * 14.0
						driftY := lifeFrac * 4.0

						// Wave center in world space
						wcx := float64(tx*config.TileSize) + baseOffsetX + driftX
						wcy := float64(ty*config.TileSize) + baseOffsetY + driftY

						// Dynamic bobbing: all points of the wave bob together smoothly
						wcyBobbed := wcy + math.Sin(ticks*0.06+wcx*0.02)*1.2

						// Variety in wave type, length, and color
						waveType := rng.intn(3)

						switch waveType {
						case 0: // Small white foam cap
							halfLen := 3.0 + rng.float64()*2.0
							clr := color.RGBA{245, 250, 255, uint8(opacity * 170.0)}
							drawWaveCurve(screen, camX, camY, wcx, wcyBobbed, halfLen, 1.0, 1.2, clr, mult)

						case 1: // Medium light-blue wave
							halfLen := 6.0 + rng.float64()*4.0
							clr := color.RGBA{135, 215, 255, uint8(opacity * 130.0)}
							drawWaveCurve(screen, camX, camY, wcx, wcyBobbed, halfLen, 1.5, 1.2, clr, mult)

						case 2: // Double crest (light-blue base with white cap)
							halfLen := 8.0 + rng.float64()*5.0
							clrBlue := color.RGBA{130, 205, 255, uint8(opacity * 120.0)}
							clrWhite := color.RGBA{245, 250, 255, uint8(opacity * 180.0)}

							// 1. Blue base crest
							drawWaveCurve(screen, camX, camY, wcx, wcyBobbed, halfLen, 1.8, 1.2, clrBlue, mult)

							// 2. White cap (offset upwards and shorter)
							drawWaveCurve(screen, camX, camY, wcx, wcyBobbed-2.0, halfLen*0.5, 1.0, 1.0, clrWhite, mult)
						}
					}
				}
			}
		}
	}
}

func (o *OverworldScene) drawVehicleBeacons(screen *ebiten.Image, startTileX, endTileX, startTileY, endTileY int, g OverworldContext, camX, camY float64) {
	// Precompute active cave vehicle coordinates to avoid Sprintf allocations per tile.
	var cavesWithVehicles map[[2]int]bool
	hasVoidDiveVehicle := false
	if vehiclesMap := g.GetAllCaveVehicles(); len(vehiclesMap) > 0 {
		cavesWithVehicles = make(map[[2]int]bool)
		for key, vehicles := range vehiclesMap {
			if len(vehicles) > 0 {
				if key == "void_dive" {
					hasVoidDiveVehicle = true
				} else {
					var cx, cy int
					if _, err := fmt.Sscanf(key, "%d_%d", &cx, &cy); err == nil {
						cavesWithVehicles[[2]int{cx, cy}] = true
					}
				}
			}
		}
	}

	trenchKey := g.GetActiveTrenchKey()
	trenchX, trenchY := g.GetActiveTrenchCoords()

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			sx := float32(tx*config.TileSize - int(camX))
			sy := float32(ty*config.TileSize - int(camY))

			hasVehicle := false
			if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height {
				if cavesWithVehicles != nil && cavesWithVehicles[[2]int{tx, ty}] {
					hasVehicle = true
				}
			} else if trenchKey == "void_dive" && tx == trenchX && ty == trenchY {
				if hasVoidDiveVehicle {
					hasVehicle = true
				}
			}

			if hasVehicle {
				cx := sx + float32(config.TileSize)/2.0
				cy := sy + float32(config.TileSize)/2.0
				pulse := float32(math.Sin(g.GetTicks()*0.08)) * 3.5
				radius := float32(12.0) + pulse
				vector.StrokeCircle(screen, cx, cy, radius, 1.5, color.RGBA{0, 220, 255, 140}, false)
				vector.FillCircle(screen, cx, cy, 5.0, color.RGBA{0, 120, 180, 220}, false)
				vector.StrokeCircle(screen, cx, cy, 5.0, 1.0, color.RGBA{0, 240, 255, 255}, false)
				vector.FillCircle(screen, cx, cy, 1.5, color.RGBA{255, 255, 255, 255}, false)
			}
		}
	}
}

func drawWaveSegment(screen *ebiten.Image, camX, camY float64, x1, y1, x2, y2 float64, thickness float32, clr color.RGBA, mult float64) {
	lx1 := float32(x1 - camX)
	ly1 := float32(y1 - camY)
	lx2 := float32(x2 - camX)
	ly2 := float32(y2 - camY)
	litColor := applyLight(clr, mult)
	vector.StrokeLine(screen, lx1, ly1, lx2, ly2, thickness, litColor, true)
}

func drawWaveCurve(screen *ebiten.Image, camX, camY float64, targetWcx, targetWcy, halfLen, maxArcHeight float64, thickness float32, clr color.RGBA, mult float64) {
	const segments = 6
	xStep := (halfLen * 2.0) / segments
	var lastX, lastY float64

	for i := 0; i <= segments; i++ {
		frac := float64(i) / float64(segments)
		wx := targetWcx - halfLen + float64(i)*xStep
		arc := 4.0 * frac * (1.0 - frac) * maxArcHeight
		wy := targetWcy - arc

		if i > 0 {
			drawWaveSegment(screen, camX, camY, lastX, lastY, wx, wy, thickness, clr, mult)
		}
		lastX, lastY = wx, wy
	}
}

// ComputeTileColors calculates fill and stroke colors for a tile, applying the void-border fade and light multiplier.
func ComputeTileColors(tx, ty int, tileType world.TileType, landDist, waterDist int, worldWidth, worldHeight int, mult float64) (color.RGBA, color.RGBA) {
	var distFromBorder float64
	if tx < 0 || tx >= worldWidth || ty < 0 || ty >= worldHeight {
		dx := 0.0
		if tx < 0 {
			dx = float64(-tx)
		} else if tx >= worldWidth {
			dx = float64(tx - worldWidth + 1)
		}
		dy := 0.0
		if ty < 0 {
			dy = float64(-ty)
		} else if ty >= worldHeight {
			dy = float64(ty - worldHeight + 1)
		}
		distFromBorder = math.Max(dx, dy)
	} else {
		dxInside := float64(tx)
		if float64(worldWidth-1-tx) < dxInside {
			dxInside = float64(worldWidth - 1 - tx)
		}
		dyInside := float64(ty)
		if float64(worldHeight-1-ty) < dyInside {
			dyInside = float64(worldHeight - 1 - ty)
		}
		distFromBorder = -math.Min(dxInside, dyInside)
	}

	t := (distFromBorder + 1.0) / 4.0
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	var baseClr color.RGBA
	var baseStrokeClr color.RGBA

	tileInfo := world.GetTileInfo(tileType)
	if tileInfo != nil && tileInfo.IsWater {
		dist := landDist
		const maxDist = 15
		lerpT := float64(dist) / float64(maxDist)
		if lerpT > 1.0 {
			lerpT = 1.0
		}
		r := uint8(28 - lerpT*20)
		gr := uint8(85 - lerpT*53)
		b := uint8(165 - lerpT*83)
		baseClr = color.RGBA{r, gr, b, 255}
		baseStrokeClr = color.RGBA{r + 8, gr + 10, b + 15, 255}
	} else if tileType == world.TileLand {
		dist := waterDist
		isSand := dist == 1 // exactly adjacent to water

		if isSand {
			baseClr = color.RGBA{232, 212, 165, 255} // beach sand
			baseStrokeClr = color.RGBA{215, 195, 145, 255}
		} else {
			// Grass gradient from shore edge to inland
			const maxLandDist = 4
			lerpT := float64(dist-2) / float64(maxLandDist-2)
			if lerpT > 1.0 {
				lerpT = 1.0
			}
			if lerpT < 0.0 {
				lerpT = 0.0
			}
			r := uint8(20 + lerpT*18)
			gr := uint8(100 + lerpT*42)
			b := uint8(60 + lerpT*25)
			baseClr = color.RGBA{r, gr, b, 255}

			sr := uint8(25 + lerpT*23)
			sgr := uint8(115 + lerpT*45)
			sb := uint8(70 + lerpT*28)
			baseStrokeClr = color.RGBA{sr, sgr, sb, 255}
		}
	}

	voidClr := color.RGBA{4, 6, 12, 255}
	voidStrokeClr := color.RGBA{8, 12, 20, 255}

	rv := uint8(float64(baseClr.R)*(1.0-t) + float64(voidClr.R)*t)
	gv := uint8(float64(baseClr.G)*(1.0-t) + float64(voidClr.G)*t)
	bv := uint8(float64(baseClr.B)*(1.0-t) + float64(voidClr.B)*t)
	av := uint8(float64(baseClr.A)*(1.0-t) + float64(voidClr.A)*t)
	tileClr := color.RGBA{rv, gv, bv, av}

	srv := uint8(float64(baseStrokeClr.R)*(1.0-t) + float64(voidStrokeClr.R)*t)
	sgv := uint8(float64(baseStrokeClr.G)*(1.0-t) + float64(voidStrokeClr.G)*t)
	sbv := uint8(float64(baseStrokeClr.B)*(1.0-t) + float64(voidStrokeClr.B)*t)
	sav := uint8(float64(baseStrokeClr.A)*(1.0-t) + float64(voidStrokeClr.A)*t)
	strokeClr := color.RGBA{srv, sgv, sbv, sav}

	return applyLight(tileClr, mult), applyLight(strokeClr, mult)
}

func floorDiv(a, b int) int {
	d := a / b
	if (a%b) != 0 && (a < 0) != (b < 0) {
		d--
	}
	return d
}

// GetOverworldLightMultiplier returns a light multiplier based on TimeOfDay.
func GetOverworldLightMultiplier(timeOfDay float64) float64 {
	if timeOfDay >= 0 && timeOfDay < 1200 {
		return 0.2 + (timeOfDay/1200.0)*0.8
	}
	if timeOfDay >= 1200 && timeOfDay < 9600 {
		return 1.0
	}
	if timeOfDay >= 9600 && timeOfDay < 10800 {
		return 1.0 - ((timeOfDay-9600.0)/1200.0)*0.8
	}
	return 0.2
}

func applyLight(c color.RGBA, mult float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * mult),
		G: uint8(float64(c.G) * mult),
		B: uint8(float64(c.B) * mult),
		A: c.A,
	}
}

type statelessRNG uint64

func (r *statelessRNG) next() uint64 {
	*r ^= *r >> 12
	*r ^= *r << 25
	*r ^= *r >> 27
	return uint64(*r) * 0x2545F4914F6CDD1D
}

func (r *statelessRNG) float64() float64 {
	return float64(r.next()) / float64(math.MaxUint64)
}

func (r *statelessRNG) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() % uint64(n))
}

func hashCoords(tx, ty int) uint64 {
	// Injective combination of 32-bit coordinates into 64-bit int
	x := (int64(tx) << 32) | (int64(uint32(ty)))
	// SplitMix64 finalizer
	u := uint64(x)
	u ^= u >> 33
	u *= 0xff51afd7ed558ccd
	u ^= u >> 33
	u *= 0xc4ceb9fe1a85ec53
	u ^= u >> 33
	return u
}
