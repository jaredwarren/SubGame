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

type VentState int

const (
	VentDormant VentState = iota
	VentWarning
	VentErupting
)

// ThermalVent represents a volcanic/hydrothermal vent dealing damage & pushing things away.
type ThermalVent struct {
	Pos            gvec.Vec2
	Radius         float64
	BubbleCooldown int
	State          VentState
	StateTimer     int
	SeedOffset     int64
	Intensity      float64
}

// Update ticks the thermal vent bubble particle spawn rate and geyser state machine.
func (v *ThermalVent) Update(g GameContext, targetCenter gvec.Vec2, targetDims gvec.Vec2, isPiloting bool) {
	// Tick down StateTimer and handle transitions
	if v.StateTimer <= 0 {
		r := rand.New(rand.NewSource(int64(g.GetTicks()) + v.SeedOffset))
		switch v.State {
		case VentDormant:
			v.State = VentWarning
			v.StateTimer = r.Intn(31) + 60 // 60-90 ticks (1 - 1.5 seconds)
		case VentWarning:
			v.State = VentErupting
			v.StateTimer = r.Intn(61) + 120 // 120-180 ticks (2 - 3 seconds)
			v.Intensity = 1.2               // INSTANT ERUPTION BURST!
		case VentErupting:
			v.State = VentDormant
			v.StateTimer = r.Intn(221) + 180 // 180-400 ticks (3 - 6.6 seconds)
		}
	}
	v.StateTimer--

	// Smoothly transition intensity based on current state
	switch v.State {
	case VentDormant:
		v.Intensity += (0.0 - v.Intensity) * 0.03 // slowly fade to dormant
	case VentWarning:
		target := 0.4 + 0.1*math.Sin(float64(g.GetTicks())*0.2)
		v.Intensity += (target - v.Intensity) * 0.08 // quick pulse warning
	case VentErupting:
		v.Intensity += (0.5 - v.Intensity) * 0.01 // slowly fade/decay during eruption
	}

	// Spawn rising bubble particles based on state
	v.BubbleCooldown--
	if v.BubbleCooldown <= 0 {
		switch v.State {
		case VentDormant:
			v.BubbleCooldown = rand.Intn(40) + 40 // very low bubbles
		case VentWarning:
			v.BubbleCooldown = rand.Intn(15) + 10 // moderate bubbles
		case VentErupting:
			v.BubbleCooldown = rand.Intn(4) + 2 // constant bubble eruption!
		}

		angle := rand.Float64() * 2.0 * math.Pi
		dist := rand.Float64() * 12.0
		bx := v.Pos.X + math.Cos(angle)*dist
		by := v.Pos.Y + math.Sin(angle)*dist
		g.SpawnBubble(bx, by)
	}

	// Calculate distance to player/vehicle
	dx := targetCenter.X - v.Pos.X
	dy := targetCenter.Y - v.Pos.Y
	dist := math.Hypot(dx, dy)

	// Screen rumble during Warning state
	if v.State == VentWarning && dist < v.Radius {
		g.TriggerScreenShake(1, 0.2)
	}

	// Push forces and damage ONLY apply during Erupting state
	if v.State == VentErupting && dist < v.Radius {
		// Calculate outward push force (stronger closer to center, scaled by intensity)
		ratio := 1.0 - (dist / v.Radius)
		intensityScale := math.Max(0.0, math.Min(1.0, v.Intensity))
		pushStrength := 1.8 * ratio * intensityScale

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
			// Deal continuous structural damage to the vehicle
			activeVeh.TakeDamage(0.06 * intensityScale)
		} else {
			// Apply force to player velocity
			p := g.GetPlayer()
			p.Vel.X += pushX
			p.Vel.Y += pushY
			// Deal damage to swimming player
			p.CurrentHealth -= 0.15 * intensityScale
		}

		// Trigger visual screen shake if close, scaled by intensity
		if dist < 40.0 {
			g.TriggerScreenShake(1, 1.2*intensityScale)
		}
	}
}

// Draw renders a glowing circular volcanic mouth on the seafloor with states.
func (v *ThermalVent) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, mult float64) {
	sx := float32(v.Pos.X - camX)
	sy := float32(v.Pos.Y - camY)

	drawIntensity := v.Intensity
	if v.State == VentWarning {
		// Add rapid warning flash to drawIntensity
		flash := (int(ticks) % 16) < 8
		if flash {
			drawIntensity += 0.25
		}
	}

	// Clamp drawIntensity to [0, 1] for color/size interpolations
	t := math.Max(0.0, math.Min(1.0, drawIntensity))

	// Interpolate pulse speed and amplitude based on state
	var pulseSpeed, pulseAmp float64
	switch v.State {
	case VentDormant:
		pulseSpeed = 0.02
		pulseAmp = 0.8
	case VentWarning:
		pulseSpeed = 0.12
		pulseAmp = 1.5
	case VentErupting:
		pulseSpeed = 0.25
		pulseAmp = 3.5
	}
	pulse := float32(math.Sin(ticks*pulseSpeed)) * float32(pulseAmp)
	outerRad := float32(16.0) + float32(8.0*t) + pulse

	// Define colors for dormant (t = 0) vs fully erupted (t = 1)
	dormantRed := color.RGBA{70, 15, 5, 120}
	eruptedRed := color.RGBA{220, 55, 10, 220}

	dormantOrange := color.RGBA{90, 25, 5, 150}
	eruptedOrange := color.RGBA{255, 145, 25, 255}

	dormantYellow := color.RGBA{110, 45, 10, 180}
	eruptedYellow := color.RGBA{255, 230, 70, 255}

	dormantBlack := color.RGBA{8, 4, 5, 255}
	eruptedBlack := color.RGBA{18, 8, 10, 255}

	// Interpolate and apply lighting multiplier
	glowRed := applyLight(lerpColor(dormantRed, eruptedRed, t), mult)
	glowOrange := applyLight(lerpColor(dormantOrange, eruptedOrange, t), mult)
	glowYellow := applyLight(lerpColor(dormantYellow, eruptedYellow, t), mult)
	abyssBlack := applyLight(lerpColor(dormantBlack, eruptedBlack, t), mult)

	vector.FillCircle(screen, sx, sy, outerRad, glowRed, false)
	vector.FillCircle(screen, sx, sy, outerRad-4.0, glowOrange, false)
	vector.FillCircle(screen, sx, sy, outerRad-8.0, glowYellow, false)
	vector.FillCircle(screen, sx, sy, outerRad-12.0, abyssBlack, false)
}

func lerpColor(c1, c2 color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(c1.R) + float64(int(c2.R)-int(c1.R))*t),
		G: uint8(float64(c1.G) + float64(int(c2.G)-int(c1.G))*t),
		B: uint8(float64(c1.B) + float64(int(c2.B)-int(c1.B))*t),
		A: uint8(float64(c1.A) + float64(int(c2.A)-int(c1.A))*t),
	}
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
			o.vents = append(o.vents, &ThermalVent{
				Pos:            pos,
				Radius:         70.0,
				State:          VentDormant,
				StateTimer:     r.Intn(300) + 120, // stagger initial transitions
				SeedOffset:     int64(i * 12345),
				BubbleCooldown: r.Intn(10) + 5,
			})
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

				g.SetMineWarning("Salvaged: "+loot.GetName()+"!", 120, 1)

				if !isPiloting {
					p.RecalculateUpgrades()
				}
			} else {
				g.SetMineWarning("Inventory full! Cannot salvage crate.", 60, 2)
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
