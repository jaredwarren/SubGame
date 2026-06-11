package scene

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/assets"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// OverworldScene manages the top-down surface sailing view.
type OverworldScene struct {
	World       *world.World
	whirlpool   *oe.Whirlpool
	crates      []*oe.FloatingCrate
	vents       []*oe.ThermalVent
	fish         []*oe.CosmeticFish
	tileTextures map[world.TileType]*ebiten.Image
	initialized  bool
}

// NewOverworldScene creates a new OverworldScene.
func NewOverworldScene(w *world.World) *OverworldScene {
	return &OverworldScene{World: w}
}

func (o *OverworldScene) getTileTexture(tileType world.TileType) *ebiten.Image {
	if o.tileTextures == nil {
		o.tileTextures = map[world.TileType]*ebiten.Image{
			world.TileTrench:   trenchTexture,
			world.TileWreckage: wreckageTexture,
		}
	}
	return o.tileTextures[tileType]
}

func (o *OverworldScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateOverworld)
	if o.whirlpool == nil {
		o.whirlpool = oe.NewWhirlpool(g.GetWorld().Seed)
		rng := rand.New(rand.NewSource(g.GetWorld().Seed + 997))
		pos := o.FindSafeWhirlpoolSpawnPos(g.GetBaseStation().Pos, rng)
		o.whirlpool.Relocate(pos)
	}
}

func (o *OverworldScene) OnExit(g GameContext) {}

// Update handles input, movement physics, and checks state transition triggers.
func (o *OverworldScene) Update(g GameContext) error {
	if o.whirlpool == nil {
		o.whirlpool = oe.NewWhirlpool(g.GetWorld().Seed)
		rng := rand.New(rand.NewSource(g.GetWorld().Seed + 997))
		pos := o.FindSafeWhirlpoolSpawnPos(g.GetBaseStation().Pos, rng)
		o.whirlpool.Relocate(pos)
	}

	o.whirlpool.Update(whirlpoolContext{scene: o, g: g})
	o.UpdateExtras(g)

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

	o.CheckCollisions(p, g.GetBaseStation())

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

func (o *OverworldScene) CheckCollisions(p *player.Player, baseStation *base.BaseStation) {
	hasBase := baseStation != nil && baseStation.Size.X > 0 && baseStation.Size.Y > 0

	newX := p.Pos.X + p.Vel.X
	collidesX := o.IsSolid(newX, p.Pos.Y, p.Width, p.Height)
	if !collidesX && hasBase {
		bPos, bSize := baseStation.Pos, baseStation.Size
		collidesX = newX < bPos.X+bSize.X && newX+p.Width > bPos.X &&
			p.Pos.Y < bPos.Y+bSize.Y && p.Pos.Y+p.Height > bPos.Y
	}

	if collidesX {
		p.Vel.X = 0
	} else {
		p.Pos.X = newX
	}

	newY := p.Pos.Y + p.Vel.Y
	collidesY := o.IsSolid(p.Pos.X, newY, p.Width, p.Height)
	if !collidesY && hasBase {
		bPos, bSize := baseStation.Pos, baseStation.Size
		collidesY = p.Pos.X < bPos.X+bSize.X && p.Pos.X+p.Width > bPos.X &&
			newY < bPos.Y+bSize.Y && newY+p.Height > bPos.Y
	}

	if collidesY {
		p.Vel.Y = 0
	} else {
		p.Pos.Y = newY
	}
}

var (
	trenchTexture         *ebiten.Image
	trenchTextureLoaded   bool
	wreckageTexture       *ebiten.Image
	wreckageTextureLoaded bool
)

func removeChromaKey(img image.Image) *ebiten.Image {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	for i := 0; i < len(rgba.Pix); i += 4 {
		ru := rgba.Pix[i]
		gu := rgba.Pix[i+1]
		bu := rgba.Pix[i+2]

		if gu > 140 && ru < 100 && bu < 100 {
			rgba.Pix[i] = 0
			rgba.Pix[i+1] = 0
			rgba.Pix[i+2] = 0
			rgba.Pix[i+3] = 0
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

// LoadAssets preloads and chroma-keys all overworld tile textures.
func LoadAssets() {
	// 1. Trench Texture
	{
		img, _, err := image.Decode(bytes.NewReader(assets.TrenchSurfacePNG))
		if err != nil {
			log.Printf("Error: Failed to decode trench surface: %v", err)
		} else {
			trenchTexture = removeChromaKey(img)
			trenchTextureLoaded = true
		}
	}

	// 2. Wreckage Texture
	{
		img, _, err := image.Decode(bytes.NewReader(assets.WreckageSurfacePNG))
		if err != nil {
			log.Printf("Error: Failed to decode wreckage surface: %v", err)
		} else {
			wreckageTexture = removeChromaKey(img)
			wreckageTextureLoaded = true
		}
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

	op := &ebiten.DrawImageOptions{}

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
			case world.TileWater, world.TileTrench, world.TileWreckage:
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
				dist := o.World.WaterDist[clampedTx][clampedTy]
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

			mult := GetOverworldLightMultiplier(g.GetTimeOfDay())
			tileClr = applyLight(tileClr, mult)
			strokeClr = applyLight(strokeClr, mult)

			tileType := o.World.OverworldMap[clampedTx][clampedTy]
			drawTexture := o.getTileTexture(tileType)

			// Always draw the base background tile (e.g. water color) first
			vector.FillRect(screen, sx, sy, config.TileSize, config.TileSize, tileClr, false)
			vector.StrokeRect(screen, sx, sy, config.TileSize, config.TileSize, 0.5, strokeClr, false)

			if drawTexture != nil {
				op.GeoM.Reset()
				op.ColorScale.Reset()
				wImg, hImg := drawTexture.Bounds().Dx(), drawTexture.Bounds().Dy()
				if wImg > 0 && hImg > 0 {
					op.GeoM.Scale(float64(config.TileSize)/float64(wImg), float64(config.TileSize)/float64(hImg))
				}
				op.GeoM.Translate(float64(sx), float64(sy))

				// Apply fade and time-of-day lighting
				totalMult := float32((1.0 - t) * mult)
				op.ColorScale.Scale(totalMult, totalMult, totalMult, 1.0)

				screen.DrawImage(drawTexture, op)
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
						vector.StrokeLine(screen, sx+rx, sy+ry, sx+rx+8, sy+ry+2, 1.0, rippleClr, false)
					}

					// 2. Draw occasional pebble or starfish (40% chance)
					if rng.float64() < 0.40 {
						px := float32(rng.float64()*44.0) + 10.0
						py := float32(rng.float64()*44.0) + 10.0

						if rng.float64() < 0.15 {
							// Rare starfish! (Orange cross/star)
							starClr := applyLight(color.RGBA{235, 110, 50, 255}, mult)
							vector.StrokeLine(screen, sx+px-3, sy+py, sx+px+3, sy+py, 1.2, starClr, false)
							vector.StrokeLine(screen, sx+px, sy+py-3, sx+px, sy+py+3, 1.2, starClr, false)
						} else {
							// Grey/white seashell/pebble
							pebbleClr := applyLight(color.RGBA{235, 230, 220, 255}, mult)
							vector.FillCircle(screen, sx+px, sy+py, 2.0, pebbleClr, false)
							vector.StrokeCircle(screen, sx+px, sy+py, 2.0, 0.8, applyLight(color.RGBA{180, 175, 165, 255}, mult), false)
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
						vector.StrokeLine(screen, sx+gx, sy+gy, sx+gx2, sy+gy2, 1.0, grassClr, false)

						gx3 := gx + lenG*0.7*float32(math.Sin(float64(angleG+0.3)))
						gy3 := gy - lenG*0.7*float32(math.Cos(float64(angleG+0.3)))
						vector.StrokeLine(screen, sx+gx, sy+gy, sx+gx3, sy+gy3, 1.0, grassClr, false)
					}

					// 2. Flowering plants (35% chance)
					if rng.float64() < 0.35 {
						px := float32(rng.float64()*40.0) + 12.0
						py := float32(rng.float64()*40.0) + 12.0

						// Stem
						stemClr := applyLight(color.RGBA{40, 150, 80, 255}, mult)
						vector.StrokeLine(screen, sx+px, sy+py, sx+px, sy+py-5, 1.0, stemClr, false)

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
						vector.FillCircle(screen, sx+px, sy+py-6, 2.0, petalClr, false)
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
						vector.FillCircle(screen, sx+tx+3, sy+ty+6, 8.5, shadowClr, false)

						// Trunk
						vector.FillRect(screen, sx+tx-2, sy+ty, 4, 8, trunkClr, false)

						// Main leafy canopy
						vector.FillCircle(screen, sx+tx, sy+ty-4, 9.0, canopyClr, false)
						vector.StrokeCircle(screen, sx+tx, sy+ty-4, 9.0, 1.0, canopyStroke, false)

						// Highlight on the top-left of the canopy
						highlightClr := applyLight(color.RGBA{65, 175, 110, 255}, mult)
						vector.FillCircle(screen, sx+tx-3, sy+ty-7, 3.5, highlightClr, false)
					}
				}
			}

			// Draw procedurally generated waves for standard water tiles
			currTile := o.World.OverworldMap[clampedTx][clampedTy]
			if currTile == world.TileWater || currTile == world.TileTrench || currTile == world.TileWreckage {
				if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height {
					ticks := g.GetTicks()

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

			hasVehicle := false
			trenchKey := g.GetActiveTrenchKey()
			trenchX, trenchY := g.GetActiveTrenchCoords()
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

	o.DrawExtras(g, screen)

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
