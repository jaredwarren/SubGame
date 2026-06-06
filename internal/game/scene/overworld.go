package scene

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// OverworldScene manages the top-down surface sailing view.
type OverworldScene struct {
	World     *world.World
	whirlpool *Whirlpool
}

// NewOverworldScene creates a new OverworldScene.
func NewOverworldScene(w *world.World) *OverworldScene {
	return &OverworldScene{World: w}
}

func (o *OverworldScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateOverworld)
	if o.whirlpool == nil {
		o.whirlpool = NewWhirlpool(g.GetWorld().Seed)
		o.whirlpool.Relocate(o.World, g.GetBaseStation().Pos)
	}
}

func (o *OverworldScene) OnExit(g GameContext) {}

// Update handles input, movement physics, and checks state transition triggers.
func (o *OverworldScene) Update(g GameContext) error {
	if o.whirlpool == nil {
		o.whirlpool = NewWhirlpool(g.GetWorld().Seed)
		o.whirlpool.Relocate(o.World, g.GetBaseStation().Pos)
	}

	o.whirlpool.Update(o.World, g.GetBaseStation().Pos)

	// If piloting a vehicle, apply forces to the vehicle and handle death.
	if v := g.GetActiveVehicle(); v != nil {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		vCenter := gvec.Vec2{
			X: vPos.X + vDims.X/2.0,
			Y: vPos.Y + vDims.Y/2.0,
		}

		// Check death center distance (within 15 pixels of whirlpool eye)
		wpDx := o.whirlpool.Pos.X - vCenter.X
		wpDy := o.whirlpool.Pos.Y - vCenter.Y
		wpDist := math.Hypot(wpDx, wpDy)

		if wpDist < 15.0 {
			g.SetDeathReason("Your vehicle was dragged into the abyss by a violent whirlpool!")
			g.DestroyOverworldVehicle(v)
			g.GetPlayer().CurrentHealth = 0
			return nil
		}

		// Apply pulling force
		force := o.whirlpool.PullForce(vCenter)
		if force.X != 0 || force.Y != 0 {
			newPos := gvec.Vec2{
				X: vPos.X + force.X,
				Y: vPos.Y + force.Y,
			}
			// Collision checks before updating position
			if !o.IsSolid(newPos.X, newPos.Y, vDims.X, vDims.Y) {
				v.SetPos(newPos)
				// Immediately sync player position to center of vehicle to avoid lag
				p := g.GetPlayer()
				p.Pos.X = newPos.X + (vDims.X-p.Width)/2.0
				p.Pos.Y = newPos.Y + (vDims.Y-p.Height)/2.0
			}
		}
		return nil
	}

	p := g.GetPlayer()
	inp := g.GetInput()

	// Calculate and apply whirlpool force to player velocity
	pCenter := gvec.Vec2{
		X: p.Pos.X + p.Width/2.0,
		Y: p.Pos.Y + p.Height/2.0,
	}

	// Check death center distance (within 15 pixels of whirlpool eye)
	wpDx := o.whirlpool.Pos.X - pCenter.X
	wpDy := o.whirlpool.Pos.Y - pCenter.Y
	wpDist := math.Hypot(wpDx, wpDy)

	if wpDist < 15.0 {
		g.SetDeathReason("You were sucked into the center of a swirling vortex!")
		g.GetPlayer().CurrentHealth = 0
		return nil
	}

	// Apply pulling force to player velocity
	force := o.whirlpool.PullForce(pCenter)
	p.Vel.X += force.X
	p.Vel.Y += force.Y

	speedProp := p.Speed["overworld"]
	accel := speedProp.Acceleration
	maxSpeed := speedProp.TopSpeed
	isSprinting := inp.IsKeyPressed(ebiten.KeyShift)

	if isSprinting && p.CurrentStamina > 0 {
		accel *= 1.5
		maxSpeed *= 1.5
	}

	moving := false
	var dx, dy float64
	if inp.IsKeyPressed(ebiten.KeyW) || inp.IsKeyPressed(ebiten.KeyArrowUp) {
		dy -= 1.0
		moving = true
	}
	if inp.IsKeyPressed(ebiten.KeyS) || inp.IsKeyPressed(ebiten.KeyArrowDown) {
		dy += 1.0
		moving = true
	}
	if inp.IsKeyPressed(ebiten.KeyA) || inp.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx -= 1.0
		moving = true
	}
	if inp.IsKeyPressed(ebiten.KeyD) || inp.IsKeyPressed(ebiten.KeyArrowRight) {
		dx += 1.0
		moving = true
	}

	if moving {
		angle := math.Atan2(dy, dx)
		p.Facing = angle
		p.Vel.X += math.Cos(angle) * accel
		p.Vel.Y += math.Sin(angle) * accel
	}

	const drag = 0.88
	p.Vel = p.Vel.Scale(drag)

	speed := p.Vel.Length()
	if speed > maxSpeed {
		p.Vel = p.Vel.Scale(maxSpeed / speed)
	}

	o.CheckCollisions(p)

	isMoving := speed > 0.1
	p.UpdateStats(false, isSprinting && isMoving && moving)

	tx := int(p.Pos.X+p.Width/2) / config.TileSize
	ty := int(p.Pos.Y+p.Height/2) / config.TileSize
	if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
		if inp.IsKeyJustPressed(ebiten.KeyE) {
			g.EnterCave(tx, ty)
			return nil
		}
	} else {
		tile := o.World.OverworldMap[tx][ty]
		if tile == world.TileTrench || tile == world.TileWater || tile == world.TileWreckage {
			if inp.IsKeyJustPressed(ebiten.KeyE) && g.GetBaseStation().DistanceToPlayer(p) >= 100.0 {
				g.EnterCave(tx, ty)
				return nil
			}
		}
	}

	return nil
}

func (o *OverworldScene) CheckCollisions(p *player.Player) {
	newX := p.Pos.X + p.Vel.X
	if o.IsSolid(newX, p.Pos.Y, p.Width, p.Height) {
		p.Vel.X = 0
	} else {
		p.Pos.X = newX
	}

	newY := p.Pos.Y + p.Vel.Y
	if o.IsSolid(p.Pos.X, newY, p.Width, p.Height) {
		p.Vel.Y = 0
	} else {
		p.Pos.Y = newY
	}
}

// IsSolid checks if the proposed bounding box overlaps with solid land.
func (o *OverworldScene) IsSolid(x, y, w, h float64) bool {
	x1 := int(math.Floor(x)) / config.TileSize
	x2 := int(math.Floor(x+w)) / config.TileSize
	y1 := int(math.Floor(y)) / config.TileSize
	y2 := int(math.Floor(y+h)) / config.TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
				continue
			}
			if o.World.OverworldMap[tx][ty] == world.TileLand {
				return true
			}
		}
	}
	return false
}

// Draw renders the overworld tiles in a viewport centered on the player.
func (o *OverworldScene) Draw(g GameContext, screen *ebiten.Image) {
	cam := g.GetCamera()
	isPiloting := g.GetActiveVehicle() != nil
	p := g.GetPlayer()

	camX := cam.Pos.X
	camY := cam.Pos.Y

	startTileX := int(camX) / config.TileSize
	endTileX := (int(camX)+config.ScreenWidth)/config.TileSize + 1
	startTileY := int(camY) / config.TileSize
	endTileY := (int(camY)+config.ScreenHeight)/config.TileSize + 1

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			sx := float32(tx*config.TileSize - int(camX))
			sy := float32(ty*config.TileSize - int(camY))

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

			var baseClr color.RGBA
			var baseStrokeClr color.RGBA

			switch o.World.OverworldMap[clampedTx][clampedTy] {
			case world.TileWater:
				dist := o.World.LandDist[clampedTx][clampedTy]
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
			case world.TileLand:
				baseClr = color.RGBA{38, 142, 85, 255}
				baseStrokeClr = color.RGBA{48, 160, 98, 255}
			case world.TileTrench:
				baseClr = color.RGBA{6, 18, 42, 255}
				baseStrokeClr = color.RGBA{10, 26, 58, 255}
			case world.TileWreckage:
				baseClr = color.RGBA{45, 52, 60, 255}
				baseStrokeClr = color.RGBA{110, 80, 50, 255}
			}

			voidClr := color.RGBA{4, 6, 12, 255}
			voidStrokeClr := color.RGBA{8, 12, 20, 255}

			rv := uint8(float64(baseClr.R)*(1.0-t) + float64(voidClr.R)*t)
			gv := uint8(float64(baseClr.G)*(1.0-t) + float64(voidClr.G)*t)
			bv := uint8(float64(baseClr.B)*(1.0-t) + float64(voidClr.B)*t)
			tileClr := color.RGBA{rv, gv, bv, 255}

			srv := uint8(float64(baseStrokeClr.R)*(1.0-t) + float64(voidStrokeClr.R)*t)
			sgv := uint8(float64(baseStrokeClr.G)*(1.0-t) + float64(voidStrokeClr.G)*t)
			sbv := uint8(float64(baseStrokeClr.B)*(1.0-t) + float64(voidStrokeClr.B)*t)
			strokeClr := color.RGBA{srv, sgv, sbv, 255}

			mult := GetOverworldLightMultiplier(g.GetTimeOfDay())
			tileClr = applyLight(tileClr, mult)
			strokeClr = applyLight(strokeClr, mult)

			vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, tileClr, false)
			vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 0.5, strokeClr, false)

			// Draw procedurally generated waves for standard water tiles
			if o.World.OverworldMap[clampedTx][clampedTy] == world.TileWater {
				ticks := g.GetTicks()

				for nx := tx - 1; nx <= tx+1; nx++ {
					for ny := ty - 1; ny <= ty+1; ny++ {
						if nx < 0 || nx >= o.World.Width || ny < 0 || ny >= o.World.Height {
							continue
						}
						if o.World.OverworldMap[nx][ny] == world.TileWater {
							// Only generate waves on a subset of water tiles (50% density)
							tileRng := rand.New(rand.NewSource(int64(nx*997 + ny*1009)))
							if tileRng.Float64() < 0.5 {
								rng := rand.New(rand.NewSource(int64(nx*773 + ny*877)))

								// Static base offset inside the tile
								baseOffsetX := rng.Float64() * float64(config.TileSize)
								baseOffsetY := rng.Float64() * float64(config.TileSize)

								// Life cycle configuration (140 ticks per cycle)
								const cycleLength = 140.0
								phaseOffset := rng.Float64() * cycleLength
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
								wcx := float64(nx*config.TileSize) + baseOffsetX + driftX
								wcy := float64(ny*config.TileSize) + baseOffsetY + driftY

								// Dynamic bobbing: all points of the wave bob together smoothly
								wcyBobbed := wcy + math.Sin(ticks*0.06+wcx*0.02)*1.2

								// Variety in wave type, length, and color
								waveType := rng.Intn(3)

								// Draw function for a wave line segment, clipped precisely to tile bounds using math.Floor
								drawWaveSegment := func(x1, y1, x2, y2 float64, thickness float32, clr color.RGBA) {
									t1x := int(math.Floor(x1 / float64(config.TileSize)))
									t1y := int(math.Floor(y1 / float64(config.TileSize)))
									t2x := int(math.Floor(x2 / float64(config.TileSize)))
									t2y := int(math.Floor(y2 / float64(config.TileSize)))

									if t1x == tx && t1y == ty && t2x == tx && t2y == ty {
										lx1 := float32(x1 - camX)
										ly1 := float32(y1 - camY)
										lx2 := float32(x2 - camX)
										ly2 := float32(y2 - camY)
										litColor := applyLight(clr, mult)
										vector.StrokeLine(screen, lx1, ly1, lx2, ly2, thickness, litColor, true)
									}
								}

								// Draw function for a parabolic wave curve
								drawWaveCurve := func(targetWcx, targetWcy, halfLen, maxArcHeight float64, thickness float32, clr color.RGBA) {
									const segments = 6
									xStep := (halfLen * 2.0) / segments
									var lastX, lastY float64

									for i := 0; i <= segments; i++ {
										frac := float64(i) / float64(segments)
										wx := targetWcx - halfLen + float64(i)*xStep
										arc := 4.0 * frac * (1.0 - frac) * maxArcHeight
										wy := targetWcy - arc

										if i > 0 {
											drawWaveSegment(lastX, lastY, wx, wy, thickness, clr)
										}
										lastX, lastY = wx, wy
									}
								}

								switch waveType {
								case 0: // Small white foam cap
									halfLen := 3.0 + rng.Float64()*2.0
									clr := color.RGBA{245, 250, 255, uint8(opacity * 170.0)}
									drawWaveCurve(wcx, wcyBobbed, halfLen, 1.0, 1.2, clr)

								case 1: // Medium light-blue wave
									halfLen := 6.0 + rng.Float64()*4.0
									clr := color.RGBA{135, 215, 255, uint8(opacity * 130.0)}
									drawWaveCurve(wcx, wcyBobbed, halfLen, 1.5, 1.2, clr)

								case 2: // Double crest (light-blue base with white cap)
									halfLen := 8.0 + rng.Float64()*5.0
									clrBlue := color.RGBA{130, 205, 255, uint8(opacity * 120.0)}
									clrWhite := color.RGBA{245, 250, 255, uint8(opacity * 180.0)}

									// 1. Blue base crest
									drawWaveCurve(wcx, wcyBobbed, halfLen, 1.8, 1.2, clrBlue)

									// 2. White cap (offset upwards and shorter)
									drawWaveCurve(wcx, wcyBobbed-2.0, halfLen*0.5, 1.0, 1.0, clrWhite)
								}
							}
						}
					}
				}
			}

			hasVehicle := false
			trenchKey := g.GetActiveTrenchKey()
			trenchX, trenchY := g.GetActiveTrenchCoords()
			if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height {
				key := fmt.Sprintf("%d_%d", tx, ty)
				if vehicles := g.GetCaveVehicles(key); len(vehicles) > 0 {
					hasVehicle = true
				}
			} else if trenchKey == "void_dive" && tx == trenchX && ty == trenchY {
				if vehicles := g.GetCaveVehicles("void_dive"); len(vehicles) > 0 {
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

	if o.whirlpool != nil {
		o.whirlpool.Draw(screen, camX, camY)
	}

	if !isPiloting {
		pTileX := int(p.Pos.X+p.Width/2) / config.TileSize
		pTileY := int(p.Pos.Y+p.Height/2) / config.TileSize
		if pTileX < 0 || pTileX >= o.World.Width || pTileY < 0 || pTileY >= o.World.Height {
			promptText := "Press [E] to Dive into Void"
			promptX := float32(p.CenterX()) - 95
			promptY := float32(p.CenterY()) - 40
			vector.FillRect(screen, promptX, promptY, 190, 25, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, promptText, int(promptX)+10, int(promptY)+4)
		} else {
			tile := o.World.OverworldMap[pTileX][pTileY]
			if tile == world.TileTrench || tile == world.TileWater || tile == world.TileWreckage {
				promptText := "Press [E] to Dive"
				if tile == world.TileWreckage {
					promptText = "Press [E] to Salvage Wreckage"
				}
				promptX := float32(p.CenterX()) - 95
				promptY := float32(p.CenterY()) - 40
				vector.FillRect(screen, promptX, promptY, 190, 25, color.RGBA{0, 0, 0, 180}, false)
				ebitenutil.DebugPrintAt(screen, promptText, int(promptX)+10, int(promptY)+4)
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
