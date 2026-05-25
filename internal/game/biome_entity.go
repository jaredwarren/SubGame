package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type EntityType int

const (
	EntShatterBulb EntityType = iota
	EntFalseBulbSnare
	EntBrimstoneSiphon
	EntThermoclineRammer
	EntNerveMat
	EntElectroWeaver
)

// CaveEntity represents any plant, predator, or interactive entity inside caves.
type CaveEntity struct {
	Type          EntityType
	X, Y          float64
	Vx, Vy        float64
	Width, Height float64

	// AI and animation states
	Timer     int
	State     int // 0: Idle/Mimic, 1: Active/Charging, 2: Stunned
	Facing    float64
	StunTimer int

	// Brimstone Siphon specific
	Direction string // "up", "down", "left", "right"

	Active bool
}

// GenerateCaveEntities scans the cave grid and spawns biome-specific entities.
func GenerateCaveEntities(grid [][]bool, seed int64) []*CaveEntity {
	r := rand.New(rand.NewSource(seed))
	var entities []*CaveEntity

	gridW := len(grid)
	gridH := len(grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			// Ensure it's not a solid tile
			if grid[tx][ty] {
				continue
			}

			// Biome 1: Mid-Depth (0 <= ty < 40) - Grotto
			if ty >= 4 && ty < 40 {
				// Spawn Shatter-bulb (static plant) on walls/ceiling/floor
				hasAdjacentWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
				if hasAdjacentWall && r.Float64() < 0.08 {
					entities = append(entities, &CaveEntity{
						Type:   EntShatterBulb,
						X:      float64(tx*TileSize) + float64(TileSize-24)/2.0,
						Y:      float64(ty*TileSize) + float64(TileSize-24)/2.0,
						Width:  24,
						Height: 24,
						Active: true,
					})
				}
				// Spawn False-Bulb Snare (mimics Shatter-bulb, hangs on ceiling)
				if grid[tx][ty-1] && r.Float64() < 0.04 {
					entities = append(entities, &CaveEntity{
						Type:   EntFalseBulbSnare,
						X:      float64(tx*TileSize) + float64(TileSize-24)/2.0,
						Y:      float64(ty*TileSize) + 4, // Hang near ceiling
						Width:  24,
						Height: 32,
						Active: true,
					})
				}
			}

			// Biome 2: Deep (40 <= ty < 80) - Smoker Trenches
			if ty >= 40 && ty < 80 {
				// Spawn Brimstone Siphon (thermal hazard) on walls
				if r.Float64() < 0.05 {
					var dir string
					if grid[tx][ty+1] { // floor
						dir = "up"
					} else if grid[tx][ty-1] { // ceiling
						dir = "down"
					} else if grid[tx-1][ty] { // left wall
						dir = "right"
					} else if grid[tx+1][ty] { // right wall
						dir = "left"
					}

					if dir != "" {
						entities = append(entities, &CaveEntity{
							Type:      EntBrimstoneSiphon,
							X:         float64(tx*TileSize) + float64(TileSize-32)/2.0,
							Y:         float64(ty*TileSize) + float64(TileSize-32)/2.0,
							Width:     32,
							Height:    32,
							Direction: dir,
							Timer:     r.Intn(120), // stagger start frames
							Active:    true,
						})
					}
				}

				// Spawn Thermocline Rammer (swimming predator) in open water
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.015 {
					entities = append(entities, &CaveEntity{
						Type:   EntThermoclineRammer,
						X:      float64(tx*TileSize) + float64(TileSize-36)/2.0,
						Y:      float64(ty*TileSize) + float64(TileSize-24)/2.0,
						Width:  36,
						Height: 24,
						Facing: r.Float64() * math.Pi * 2,
						Active: true,
					})
				}
			}

			// Biome 3: Abyssal (80 <= ty < 120) - Brine Falls
			if ty >= 80 && ty < gridH-1 {
				// Spawn Pallid Nerve-Mat (ground slow mat) on floors
				if grid[tx][ty+1] && r.Float64() < 0.10 {
					entities = append(entities, &CaveEntity{
						Type:   EntNerveMat,
						X:      float64(tx*TileSize),
						Y:      float64(ty*TileSize) + float64(TileSize-12),
						Width:  float64(TileSize),
						Height: 12,
						Active: true,
					})
				}

				// Spawn Electro-Weaver (electricity-seeking serpentine predator)
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.012 {
					// Max one Weaver per local coordinates sector to avoid overcrowding
					hasWeaverNearby := false
					for _, ent := range entities {
						if ent.Type == EntElectroWeaver && math.Abs(ent.X-float64(tx*TileSize)) < 500 {
							hasWeaverNearby = true
							break
						}
					}
					if !hasWeaverNearby {
						entities = append(entities, &CaveEntity{
							Type:   EntElectroWeaver,
							X:      float64(tx*TileSize) + float64(TileSize-40)/2.0,
							Y:      float64(ty*TileSize) + float64(TileSize-20)/2.0,
							Width:  40,
							Height: 20,
							Active: true,
						})
					}
				}
			}
		}
	}

	return entities
}

// Update evaluates entity AI behavior, movement, and collision effects.
func (ent *CaveEntity) Update(g *Game, cave *CaveState) {
	if !ent.Active {
		return
	}

	// 1. Get player center and general vector distance
	px := g.player.X + g.player.Width/2.0
	py := g.player.Y + g.player.Height/2.0
	ex := ent.X + ent.Width/2.0
	ey := ent.Y + ent.Height/2.0
	dist := math.Hypot(px-ex, py-ey)

	switch ent.Type {
	case EntShatterBulb:
		// Static plant. If player/vehicle collides, trigger pop
		vWidth, vHeight := g.player.Width, g.player.Height
		targetX, targetY := g.player.X, g.player.Y
		if g.ActiveVehicle != nil {
			targetX, targetY = g.ActiveVehicle.GetPos()
			vWidth, vHeight = g.ActiveVehicle.GetDimensions()
		}

		if rectsOverlap(ent.X, ent.Y, ent.Width, ent.Height, targetX, targetY, vWidth, vHeight) {
			ent.Pop(g, cave)
		}

	case EntFalseBulbSnare:
		// Ceiling mimic trap. Freezes in flashlight, charges from dark.
		if dist > 360.0 {
			ent.State = 0 // Idle/hidden
			return
		}

		// Check if flashlight illuminates this trap
		isLit := false
		if g.FlashlightOn {
			facingAngle := g.player.Facing
			if g.ActiveVehicle != nil {
				facingAngle = g.ActiveVehicle.GetFacing()
			}
			dx := ex - px
			dy := ey - py
			angleToEnt := math.Atan2(dy, dx)

			// Normalize angle diff
			diff := angleToEnt - facingAngle
			for diff > math.Pi {
				diff -= 2 * math.Pi
			}
			for diff < -math.Pi {
				diff += 2 * math.Pi
			}
			if math.Abs(diff) < 0.42 { // inside flashlight cone (approx 24 degree half-angle)
				isLit = true
			}
		}

		// Sound pop alerts this trap globally within 280px
		soundAlerted := g.SoundWaveTimer > 0 && math.Hypot(g.SoundWaveX-ex, g.SoundWaveY-ey) < 280.0
		if soundAlerted {
			ent.State = 1 // Wake up / Aggro
		}

		if isLit {
			// Frozen!
			ent.Vx = 0
			ent.Vy = 0
		} else {
			// Pointing away or dark: lunges if player is close or sound popped
			if dist < 180.0 || ent.State == 1 {
				ent.State = 1 // Aggro state
				dx := px - ex
				dy := py - ey
				dDist := math.Hypot(dx, dy)
				if dDist > 0 {
					ent.Vx = (dx / dDist) * 3.5
					ent.Vy = (dy / dDist) * 3.5
				}
				ent.X += ent.Vx
				ent.Y += ent.Vy
			} else {
				ent.State = 0
			}
		}

		// Check damage collision
		vWidth, vHeight := g.player.Width, g.player.Height
		targetX, targetY := g.player.X, g.player.Y
		if g.ActiveVehicle != nil {
			targetX, targetY = g.ActiveVehicle.GetPos()
			vWidth, vHeight = g.ActiveVehicle.GetDimensions()
		}

		if rectsOverlap(ent.X, ent.Y, ent.Width, ent.Height, targetX, targetY, vWidth, vHeight) {
			// Attack player!
			g.player.CurrentHealth -= 20.0
			g.MineWarning = "ATTACKED BY FALSE-BULB SNARE!"
			g.MineWarningTimer = 120
			ent.Active = false // consume snare
		}

	case EntBrimstoneSiphon:
		// Cycles timers
		ent.Timer = (ent.Timer + 1) % 120
		if ent.Timer >= 60 {
			// Erupting steam/fire jet!
			// Check box overlap based on siphon jet direction
			var jx, jy, jw, jh float64
			const jetRange = 160.0 // Jet extends 2.5 tiles

			switch ent.Direction {
			case "up":
				jx, jy, jw, jh = ent.X, ent.Y-jetRange, ent.Width, jetRange
			case "down":
				jx, jy, jw, jh = ent.X, ent.Y+ent.Height, ent.Width, jetRange
			case "left":
				jx, jy, jw, jh = ent.X-jetRange, ent.Y, jetRange, ent.Height
			default: // "right"
				jx, jy, jw, jh = ent.X+ent.Width, ent.Y, jetRange, ent.Height
			}

			// Check overlap with player/vehicle
			vWidth, vHeight := g.player.Width, g.player.Height
			targetX, targetY := g.player.X, g.player.Y
			if g.ActiveVehicle != nil {
				targetX, targetY = g.ActiveVehicle.GetPos()
				vWidth, vHeight = g.ActiveVehicle.GetDimensions()
			}

			if rectsOverlap(jx, jy, jw, jh, targetX, targetY, vWidth, vHeight) {
				if g.ActiveVehicle != nil {
					g.ActiveVehicle.TakeDamage(0.4)
				} else {
					g.player.CurrentHealth -= 0.6
				}
			}
		}

	case EntThermoclineRammer:
		if ent.State == 2 { // Stunned
			ent.StunTimer--
			if ent.StunTimer <= 0 {
				ent.State = 0
			}
			return
		}

		// Check player noise aggro triggers
		isAggroTrigger := false
		if dist < 250.0 {
			// Player sprinting
			if g.ActiveVehicle == nil && ebiten.IsKeyPressed(ebiten.KeyShift) && (math.Abs(g.player.Vx) > 1.2 || math.Abs(g.player.Vy) > 1.2) {
				isAggroTrigger = true
			}
			// Mech thrusters or vehicle moving fast
			if g.ActiveVehicle != nil {
				vx, vy := g.ActiveVehicle.GetPos()
				// Approximate check if active vehicle movement buttons pressed
				if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeySpace) {
					isAggroTrigger = true
				}
				_ = vx
				_ = vy
			}
		}
		// Also triggered by sound pops
		if g.SoundWaveTimer > 0 && math.Hypot(g.SoundWaveX-ex, g.SoundWaveY-ey) < 250.0 {
			isAggroTrigger = true
		}

		if ent.State == 0 { // Patrol
			if isAggroTrigger {
				ent.State = 1 // Alert/Charge state
				// Lock charge direction to closest orthogonal axis
				dx := px - ex
				dy := py - ey
				if math.Abs(dx) > math.Abs(dy) {
					ent.Vy = 0
					if dx > 0 {
						ent.Vx = 6.2
						ent.Facing = 0.0
					} else {
						ent.Vx = -6.2
						ent.Facing = math.Pi
					}
				} else {
					ent.Vx = 0
					if dy > 0 {
						ent.Vy = 6.2
						ent.Facing = math.Pi / 2.0
					} else {
						ent.Vy = -6.2
						ent.Facing = -math.Pi / 2.0
					}
				}
			} else {
				// Swim idle patrol back and forth
				ent.Timer++
				if ent.Timer%120 == 0 {
					ent.Facing += math.Pi // turn around
				}
				ent.Vx = math.Cos(ent.Facing) * 0.8
				ent.Vy = math.Sin(ent.Facing) * 0.4
				if !cave.isSolid(ent.X+ent.Vx, ent.Y+ent.Vy, ent.Width, ent.Height) {
					ent.X += ent.Vx
					ent.Y += ent.Vy
				} else {
					ent.Facing += math.Pi // turn around on wall bump
				}
			}
		} else if ent.State == 1 { // Charging
			// Move in straight line
			nextX := ent.X + ent.Vx
			nextY := ent.Y + ent.Vy

			// Check wall collision
			if cave.isSolid(nextX, nextY, ent.Width, ent.Height) {
				ent.State = 2 // Stunned!
				ent.StunTimer = 180
				ent.Vx = 0
				ent.Vy = 0
			} else {
				ent.X = nextX
				ent.Y = nextY
			}

			// Check player/vehicle damage overlap
			vWidth, vHeight := g.player.Width, g.player.Height
			targetX, targetY := g.player.X, g.player.Y
			if g.ActiveVehicle != nil {
				targetX, targetY = g.ActiveVehicle.GetPos()
				vWidth, vHeight = g.ActiveVehicle.GetDimensions()
			}

			if rectsOverlap(ent.X, ent.Y, ent.Width, ent.Height, targetX, targetY, vWidth, vHeight) {
				if g.ActiveVehicle != nil {
					g.ActiveVehicle.TakeDamage(30.0)
					g.MineWarning = "VEHICLE RAMMED BY THERMOCLINE RAMMER!"
				} else {
					g.player.CurrentHealth -= 25.0
					g.MineWarning = "RAMMED BY THERMOCLINE RAMMER!"
				}
				g.MineWarningTimer = 120
				ent.State = 0 // return to patrol
				ent.Vx, ent.Vy = 0, 0
			}
		}

	case EntNerveMat:
		// Static ground plant. Slows player/vehicle on overlap
		vWidth, vHeight := g.player.Width, g.player.Height
		targetX, targetY := g.player.X, g.player.Y
		if g.ActiveVehicle != nil {
			targetX, targetY = g.ActiveVehicle.GetPos()
			vWidth, vHeight = g.ActiveVehicle.GetDimensions()
		}

		if rectsOverlap(ent.X, ent.Y, ent.Width, ent.Height, targetX, targetY, vWidth, vHeight) {
			g.playerSlowed = true
		}

	case EntElectroWeaver:
		// Serpentine stalker. Tracks electricity in Abyssal biome
		inAbyssal := (py / TileSize) >= 80
		if !inAbyssal {
			ent.Timer = 0
			return
		}

		isElectricity := g.FlashlightOn || g.SonarTimer > 0 || g.ActiveVehicle != nil
		if isElectricity && dist < 500.0 {
			ent.Timer++
			// Feed tracking value to game screen static/jitter
			g.WeaverTrackingTimer = math.Max(g.WeaverTrackingTimer, float64(ent.Timer))

			// Strike check (5 seconds at 60 FPS)
			if ent.Timer >= 300 {
				g.player.CurrentHealth -= 45.0
				g.MineWarning = "ELECTRO-WEAVER STRIKE! SEVERE DAMAGE!"
				g.MineWarningTimer = 180

				// Teleport Weaver near player to simulate lunging strike
				ent.X = g.player.X + float64(rand.Intn(120)-60)
				ent.Y = g.player.Y + float64(rand.Intn(120)-60)
				ent.Timer = 0
			}
		} else {
			// Fade tracking
			if ent.Timer > 0 {
				ent.Timer -= 2
				if ent.Timer < 0 {
					ent.Timer = 0
				}
			}
		}

		// Slowly wander or slither towards player if tracking
		if ent.Timer > 60 {
			dx := px - ex
			dy := py - ey
			dDist := math.Hypot(dx, dy)
			if dDist > 100 {
				ent.Vx = (dx / dDist) * 1.5
				ent.Vy = (dy / dDist) * 1.5
			} else {
				ent.Vx = math.Cos(float64(g.TimeOfDay)/30.0) * 1.2
				ent.Vy = math.Sin(float64(g.TimeOfDay)/30.0) * 1.2
			}
		} else {
			ent.Vx = math.Cos(float64(g.TimeOfDay)/40.0) * 0.8
			ent.Vy = math.Sin(float64(g.TimeOfDay)/40.0) * 0.8
		}

		// Soft collisions for slithering
		if !cave.isSolid(ent.X+ent.Vx, ent.Y+ent.Vy, ent.Width, ent.Height) {
			ent.X += ent.Vx
			ent.Y += ent.Vy
		}
	}
}

// Pop ruptures a Shatter-Bulb, restoring oxygen and emitting sound waves.
func (ent *CaveEntity) Pop(g *Game, cave *CaveState) {
	if !ent.Active {
		return
	}
	ent.Active = false

	// Restore 20 Oxygen to player
	g.player.CurrentOxygen = math.Min(g.player.MaxOxygen, g.player.CurrentOxygen+20)

	// Create sound pop coordinates
	g.SoundWaveTimer = 60
	g.SoundWaveRadius = 0.0
	g.SoundWaveX = ent.X + ent.Width/2.0
	g.SoundWaveY = ent.Y + ent.Height/2.0

	// Remove from cave state active entity list
	for i, e := range cave.Entities {
		if e == ent {
			cave.Entities = append(cave.Entities[:i], cave.Entities[i+1:]...)
			break
		}
	}
}

// Draw renders beautiful customized vector shapes for the biome entities.
func (ent *CaveEntity) Draw(screen *ebiten.Image, camera *Camera, timeOfDay float64) {
	if !ent.Active {
		return
	}

	sx := float32(ent.X - camera.X)
	sy := float32(ent.Y - camera.Y)
	sw := float32(ent.Width)
	sh := float32(ent.Height)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	switch ent.Type {
	case EntShatterBulb:
		// Draw plant stem
		vector.StrokeLine(screen, cx, cy, cx, cy+16, 2.0, color.RGBA{45, 95, 75, 255}, false)
		// Draw glowing outer aura
		vector.DrawFilledCircle(screen, cx, cy, 11, color.RGBA{0, 220, 240, 60}, false)
		// Draw central bulb
		vector.DrawFilledCircle(screen, cx, cy, 7, color.RGBA{0, 230, 245, 255}, false)
		vector.StrokeCircle(screen, cx, cy, 7, 0.8, color.RGBA{255, 255, 255, 200}, false)

	case EntFalseBulbSnare:
		// Mimics Shatter-bulb but hangs from ceiling
		vector.StrokeLine(screen, cx, sy, cx, cy, 2.0, color.RGBA{45, 95, 75, 255}, false)
		
		bulbColor := color.RGBA{0, 220, 240, 255}
		// If alert, show reddish core inside
		if ent.State == 1 {
			vector.DrawFilledCircle(screen, cx, cy, 12, color.RGBA{230, 75, 45, 80}, false)
			vector.DrawFilledCircle(screen, cx, cy, 7, color.RGBA{245, 95, 25, 255}, false)
			// Slit pupil eye
			vector.StrokeLine(screen, cx, cy-4, cx, cy+4, 1.5, color.RGBA{0, 0, 0, 255}, false)
		} else {
			vector.DrawFilledCircle(screen, cx, cy, 11, color.RGBA{0, 220, 240, 60}, false)
			vector.DrawFilledCircle(screen, cx, cy, 7, bulbColor, false)
			vector.StrokeCircle(screen, cx, cy, 7, 0.8, color.RGBA{255, 255, 255, 180}, false)
		}

	case EntBrimstoneSiphon:
		// Draw static volcanic vent cone
		path := new(vector.Path)
		path.MoveTo(cx - 16, sy + 32)
		path.LineTo(cx + 16, sy + 32)
		path.LineTo(cx + 8, sy + 12)
		path.LineTo(cx - 8, sy + 12)
		path.Close()
		
		var ventColor color.RGBA
		if ent.Timer >= 60 {
			ventColor = color.RGBA{185, 85, 45, 255} // Glowing red hot
		} else {
			ventColor = color.RGBA{65, 55, 50, 255} // Cool rock
		}
		
		op := &ebiten.DrawTrianglesOptions{FillRule: ebiten.EvenOdd}
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vs {
			vs[i].ColorR = float32(ventColor.R) / 255.0
			vs[i].ColorG = float32(ventColor.G) / 255.0
			vs[i].ColorB = float32(ventColor.B) / 255.0
			vs[i].ColorA = 1.0
		}
		screen.DrawTriangles(vs, is, emptyImage, op)

		// Draw fire/steam jet particles if erupting
		if ent.Timer >= 60 {
			jetLen := float32(120.0) // Eruption visible length
			// Draw particle trails
			var jx, jy float32
			switch ent.Direction {
			case "up":
				jx, jy = cx, cy - jetLen/2.0
				vector.DrawFilledRect(screen, cx-8, cy-jetLen, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
				vector.DrawFilledRect(screen, cx-3, cy-jetLen-10, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
			case "down":
				jx, jy = cx, cy + jetLen/2.0
				vector.DrawFilledRect(screen, cx-8, cy+16, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
				vector.DrawFilledRect(screen, cx-3, cy+16, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
			case "left":
				jx, jy = cx - jetLen/2.0, cy
				vector.DrawFilledRect(screen, cx-jetLen-16, cy-8, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
				vector.DrawFilledRect(screen, cx-jetLen-26, cy-3, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
			default: // "right"
				jx, jy = cx + jetLen/2.0, cy
				vector.DrawFilledRect(screen, cx+16, cy-8, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
				vector.DrawFilledRect(screen, cx+16, cy-3, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
			}
			_ = jx
			_ = jy
		}

	case EntThermoclineRammer:
		// Draw eyeless fish pointing in Facing angle
		cosF := float32(math.Cos(ent.Facing))
		sinF := float32(math.Sin(ent.Facing))

		// Draw fish body (grey-orange)
		bodyColor := color.RGBA{195, 95, 45, 255}
		vector.DrawFilledCircle(screen, cx, cy, 8.0, bodyColor, false)
		
		// Hard head skull (grey triangle)
		path := new(vector.Path)
		hx := cx + cosF*12
		hy := cy + sinF*12
		path.MoveTo(hx, hy)
		path.LineTo(cx - sinF*6, cy + cosF*6)
		path.LineTo(cx + sinF*6, cy - cosF*6)
		path.Close()
		
		op := &ebiten.DrawTrianglesOptions{FillRule: ebiten.EvenOdd}
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vs {
			vs[i].ColorR = 120.0 / 255.0
			vs[i].ColorG = 130.0 / 255.0
			vs[i].ColorB = 140.0 / 255.0
			vs[i].ColorA = 1.0
		}
		screen.DrawTriangles(vs, is, emptyImage, op)

		// Tail fin
		tx := cx - cosF*10
		ty := cy - sinF*10
		vector.StrokeLine(screen, tx, ty, tx - sinF*8, ty + cosF*8, 2.0, color.RGBA{195, 95, 45, 255}, false)
		vector.StrokeLine(screen, tx, ty, tx + sinF*8, ty - cosF*8, 2.0, color.RGBA{195, 95, 45, 255}, false)

		// Stun stars above head if stunned
		if ent.State == 2 {
			starAng := float64(ent.StunTimer) * 0.15
			sx1 := cx + float32(math.Cos(starAng))*14
			sy1 := cy - 14 + float32(math.Sin(starAng))*5
			sx2 := cx + float32(math.Cos(starAng+math.Pi))*14
			sy2 := cy - 14 + float32(math.Sin(starAng+math.Pi))*5
			vector.DrawFilledCircle(screen, sx1, sy1, 2.5, color.RGBA{255, 230, 40, 255}, false)
			vector.DrawFilledCircle(screen, sx2, sy2, 2.5, color.RGBA{255, 230, 40, 255}, false)
		}

	case EntNerveMat:
		// Purple nerve carpet on floor
		vector.DrawFilledRect(screen, sx, sy+sh-4, sw, 4, color.RGBA{80, 25, 120, 255}, false)
		// Small vertical fibers
		for o := float32(4); o < sw; o += 12 {
			vector.StrokeLine(screen, sx+o, sy+sh, sx+o, sy+sh-8, 1.5, color.RGBA{130, 40, 180, 255}, false)
			vector.DrawFilledCircle(screen, sx+o, sy+sh-8, 2.0, color.RGBA{180, 60, 220, 255}, false)
		}

	case EntElectroWeaver:
		// Slithering snake bodies
		bodyParts := 5
		for i := 0; i < bodyParts; i++ {
			// Offset coordinates chronologically along body path
			lag := float64(i) * 0.3
			tVal := timeOfDay*0.08 - lag
			offX := math.Cos(tVal) * 6
			offY := math.Sin(tVal) * 4

			segmentX := cx - float32(math.Cos(ent.Facing)*float64(i)*8.0) + float32(offX)
			segmentY := cy - float32(math.Sin(ent.Facing)*float64(i)*8.0) + float32(offY)

			segColor := color.RGBA{140 - uint8(i*18), 45, 205 - uint8(i*12), 255}
			vector.DrawFilledCircle(screen, segmentX, segmentY, 6.0-float32(i)*0.8, segColor, false)

			// Glow indicator on head
			if i == 0 {
				vector.DrawFilledCircle(screen, segmentX + float32(math.Cos(ent.Facing))*4, segmentY + float32(math.Sin(ent.Facing))*4, 2.0, color.RGBA{255, 255, 80, 255}, false)
			}
		}

		// Draw electrical discharge sparks if tracking
		if ent.Timer > 0 {
			sparkRatio := float64(ent.Timer) / 300.0
			numSparks := int(sparkRatio * 5)
			for s := 0; s < numSparks; s++ {
				// Random spark line from head outwards
				spx := cx + float32(rand.Intn(40)-20)
				spy := cy + float32(rand.Intn(40)-20)
				vector.StrokeLine(screen, cx, cy, spx, spy, 1.0, color.RGBA{160, 220, 255, 255}, false)
			}
		}
	}
}

func rectsOverlap(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}
