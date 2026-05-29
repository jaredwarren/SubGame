package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type EntityType int

const (
	EntShatterBulb EntityType = iota
	EntFalseBulbSnare
	EntBrimstoneSiphon
	EntThermoclineRammer
	EntNerveMat
	EntElectroWeaver
	EntPassiveFish
	EntPassiveCrab
	EntKelp
)

// CaveEntity represents any plant, predator, or interactive entity inside caves.
type CaveEntity interface {
	Update(gr Runtime, CaveGrid [][]bool)
	Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64)
	IsActive() bool
	SetActive(active bool)
	GetPos() gvec.Vec2
	GetDimensions() gvec.Vec2
	GetType() EntityType
}

// PassiveCreature defines the interface for catchable cave creatures.
type PassiveCreature interface {
	CaveEntity
	GetHarvestedItem() item.Item
	CanCatch(playerPos gvec.Vec2) bool
}

// BaseEntity implements common fields and getters/setters for all entities.
type BaseEntity struct {
	Type       EntityType
	Pos        gvec.Vec2
	Vel        gvec.Vec2
	Dimensions gvec.Vec2
	Active     bool
}

func (b *BaseEntity) IsActive() bool           { return b.Active }
func (b *BaseEntity) SetActive(active bool)    { b.Active = active }
func (b *BaseEntity) GetPos() gvec.Vec2        { return b.Pos }
func (b *BaseEntity) GetDimensions() gvec.Vec2 { return b.Dimensions }
func (b *BaseEntity) GetType() EntityType      { return b.Type }

// GenerateCaveEntities scans the cave grid and spawns biome-specific entities.
func GenerateCaveEntities(grid [][]bool, seed int64, isShallow bool) []CaveEntity {
	r := rand.New(rand.NewSource(seed))
	var entities []CaveEntity

	gridW := len(grid)
	gridH := len(grid[0])

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			// Ensure it's not a solid tile
			if grid[tx][ty] {
				continue
			}

			if isShallow {
				// Only spawn Shatter-bulb (static oxygen plant) on floor/wall tiles
				hasAdjacentWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
				if hasAdjacentWall && r.Float64() < 0.08 {
					entities = append(entities, &ShatterBulb{
						BaseEntity: BaseEntity{
							Type:       EntShatterBulb,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 24, Y: 24},
							Active:     true,
						},
					})
				}

				// Spawn PassiveFish in open water (no adjacent walls)
				isOpenWater := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenWater && r.Float64() < 0.012 {
					entities = append(entities, &PassiveFish{
						BaseEntity: BaseEntity{
							Type:       EntPassiveFish,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-20)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-12)/2.0},
							Dimensions: gvec.Vec2{X: 20, Y: 12},
							Active:     true,
						},
						FacingRight: r.Float64() < 0.5,
						SwimPhase:   r.Float64() * math.Pi * 2,
					})
				}

				// Spawn PassiveCrab on floors (open tile above solid tile)
				if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < 0.015 {
					entities = append(entities, &PassiveCrab{
						BaseEntity: BaseEntity{
							Type:       EntPassiveCrab,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-16)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-10)},
							Dimensions: gvec.Vec2{X: 16, Y: 10},
							Active:     true,
						},
						FacingRight: r.Float64() < 0.5,
					})
				}

				// Spawn Kelp on floors (open tile above solid tile)
				if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < 0.28 {
					height := 32.0 + r.Float64()*48.0
					entities = append(entities, &Kelp{
						BaseEntity: BaseEntity{
							Type:       EntKelp,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-16)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize) - height},
							Dimensions: gvec.Vec2{X: 16, Y: height},
							Active:     true,
						},
						SwayPhase: r.Float64() * math.Pi * 2,
					})
				}

				continue
			}

			// Biome 1: Mid-Depth (0 <= ty < 40) - Grotto
			if ty >= 4 && ty < 40 {
				// Spawn Shatter-bulb (static plant) on walls/ceiling/floor
				hasAdjacentWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
				if hasAdjacentWall && r.Float64() < 0.08 {
					entities = append(entities, &ShatterBulb{
						BaseEntity: BaseEntity{
							Type:       EntShatterBulb,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 24, Y: 24},
							Active:     true,
						},
					})
				}
				// Spawn False-Bulb Snare (mimics Shatter-bulb, hangs on ceiling)
				if grid[tx][ty-1] && r.Float64() < 0.04 {
					entities = append(entities, &FalseBulbSnare{
						BaseEntity: BaseEntity{
							Type:       EntFalseBulbSnare,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-24)/2.0, Y: float64(ty*config.TileSize) + 4}, // Hang near ceiling
							Dimensions: gvec.Vec2{X: 24, Y: 32},
							Active:     true,
						},
						State: 0,
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
						entities = append(entities, &BrimstoneSiphon{
							BaseEntity: BaseEntity{
								Type:       EntBrimstoneSiphon,
								Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-32)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-32)/2.0},
								Dimensions: gvec.Vec2{X: 32, Y: 32},
								Active:     true,
							},
							Direction: dir,
							Timer:     r.Intn(120), // stagger start frames
						})
					}
				}

				// Spawn Thermocline Rammer (swimming predator) in open water
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.015 {
					entities = append(entities, &ThermoclineRammer{
						BaseEntity: BaseEntity{
							Type:       EntThermoclineRammer,
							Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-36)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-24)/2.0},
							Dimensions: gvec.Vec2{X: 36, Y: 24},
							Active:     true,
						},
						Facing: r.Float64() * math.Pi * 2,
					})
				}
			}

			// Biome 3: Abyssal (80 <= ty < 120) - Brine Falls
			if ty >= 80 && ty < gridH-1 {
				// Spawn Pallid Nerve-Mat (ground slow mat) on floors
				if grid[tx][ty+1] && r.Float64() < 0.10 {
					entities = append(entities, &NerveMat{
						BaseEntity: BaseEntity{
							Type:       EntNerveMat,
							Pos:        gvec.Vec2{X: float64(tx * config.TileSize), Y: float64(ty*config.TileSize) + float64(config.TileSize-12)},
							Dimensions: gvec.Vec2{X: float64(config.TileSize), Y: 12},
							Active:     true,
						},
					})
				}

				// Spawn Electro-Weaver (electricity-seeking serpentine predator)
				isOpenSpace := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if isOpenSpace && r.Float64() < 0.012 {
					// Max one Weaver per local coordinates sector to avoid overcrowding
					hasWeaverNearby := false
					for _, ent := range entities {
						if ent.GetType() == EntElectroWeaver && math.Abs(ent.GetPos().X-float64(tx*config.TileSize)) < 500 {
							hasWeaverNearby = true
							break
						}
					}
					if !hasWeaverNearby {
						entities = append(entities, &ElectroWeaver{
							BaseEntity: BaseEntity{
								Type:       EntElectroWeaver,
								Pos:        gvec.Vec2{X: float64(tx*config.TileSize) + float64(config.TileSize-40)/2.0, Y: float64(ty*config.TileSize) + float64(config.TileSize-20)/2.0},
								Dimensions: gvec.Vec2{X: 40, Y: 20},
								Active:     true,
							},
						})
					}
				}
			}
		}
	}

	return entities
}

// ---------------------------------------------------------
// 1. SHATTER-BULB (Oxygen Plant)
// ---------------------------------------------------------

type ShatterBulb struct {
	BaseEntity
}

func (s *ShatterBulb) Update(gr Runtime, CaveGrid [][]bool) {
	// Static plant. If player/vehicle collides, trigger pop
	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}

	if rectsOverlap(s.Pos.X, s.Pos.Y, s.Dimensions.X, s.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		s.Pop(gr)
	}
}

func (s *ShatterBulb) Pop(gr Runtime) {
	if !s.Active {
		return
	}
	s.Active = false

	// Restore 20 Oxygen to player
	gr.Emit(RestoreOxygenCmd{Amount: 20})

	// Create sound pop coordinates
	gr.Emit(TriggerSoundWaveCmd{
		Pos: gvec.Vec2{
			X: s.Pos.X + s.Dimensions.X/2.0,
			Y: s.Pos.Y + s.Dimensions.Y/2.0,
		},
	})
}

func (s *ShatterBulb) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(s.Pos.X - camera.Pos.X)
	sy := float32(s.Pos.Y - camera.Pos.Y)
	sw := float32(s.Dimensions.X)
	sh := float32(s.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	// Draw plant stem
	vector.StrokeLine(screen, cx, cy, cx, cy+16, 2.0, color.RGBA{45, 95, 75, 255}, false)
	// Draw glowing outer aura with pulsing effect
	phase := s.Pos.X + s.Pos.Y
	pulse := float32(math.Cos(timeOfDay*0.02+phase)) * 2.5
	radius := float32(11.0) + pulse
	if radius < 5.0 {
		radius = 5.0
	}
	vector.FillCircle(screen, cx, cy, radius, color.RGBA{0, 220, 240, 60}, false)
	// Draw central bulb
	vector.FillCircle(screen, cx, cy, 7, color.RGBA{0, 230, 245, 255}, false)
	vector.StrokeCircle(screen, cx, cy, 7, 0.8, color.RGBA{255, 255, 255, 200}, false)
}

// ---------------------------------------------------------
// 2. FALSE-BULB SNARE (Ceiling Mimic Predator)
// ---------------------------------------------------------

type FalseBulbSnare struct {
	BaseEntity
	State int
}

func (ent *FalseBulbSnare) Update(gr Runtime, CaveGrid [][]bool) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2.0
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0
	dist := math.Hypot(px-ex, py-ey)

	if dist > 360.0 {
		ent.State = 0 // Idle/mimic
		return
	}

	// Check if flashlight illuminates this trap
	isLit := false
	if gr.FlashlightOn() {
		facingAngle := gr.PlayerFacing()
		if gr.HasActiveVehicle() {
			facingAngle = gr.ActiveVehicleFacing()
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
		if math.Abs(diff) < 0.42 { // inside flashlight cone
			isLit = true
		}
	}

	// Sound pop alerts this trap globally within 280px
	soundAlerted := gr.SoundWaveTimer() > 0 && math.Hypot(gr.SoundWaveX()-ex, gr.SoundWaveY()-ey) < 280.0
	if soundAlerted {
		ent.State = 1 // Wake up / Aggro
	}

	if isLit {
		// Frozen!
		ent.Vel = gvec.Vec2{}
	} else {
		// Pointing away or dark: lunges if player is close or sound popped
		if dist < 180.0 || ent.State == 1 {
			ent.State = 1 // Aggro state
			dx := px - ex
			dy := py - ey
			dDist := math.Hypot(dx, dy)
			if dDist > 0 {
				ent.Vel.X = (dx / dDist) * 3.5
				ent.Vel.Y = (dy / dDist) * 3.5
			}
			ent.Pos = ent.Pos.Add(ent.Vel)
		} else {
			ent.State = 0
		}
	}

	// Check damage collision
	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}

	if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		// Attack player!
		gr.Emit(DamagePlayerCmd{Amount: 20.0})
		gr.Emit(SetMineWarningCmd{Message: "ATTACKED BY FALSE-BULB SNARE!", Duration: 120})
		ent.Active = false // consume snare
	}
}

func (ent *FalseBulbSnare) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	// Mimics Shatter-bulb but hangs from ceiling
	vector.StrokeLine(screen, cx, sy, cx, cy, 2.0, color.RGBA{45, 95, 75, 255}, false)

	bulbColor := color.RGBA{0, 220, 240, 255}
	// If alert, show reddish core inside
	if ent.State == 1 {
		vector.FillCircle(screen, cx, cy, 12, color.RGBA{230, 75, 45, 80}, false)
		vector.FillCircle(screen, cx, cy, 7, color.RGBA{245, 95, 25, 255}, false)
		// Slit pupil eye
		vector.StrokeLine(screen, cx, cy-4, cx, cy+4, 1.5, color.RGBA{0, 0, 0, 255}, false)
	} else {
		// Mimics Shatter-bulb pulsing outer aura
		phase := ent.Pos.X + ent.Pos.Y
		pulse := float32(math.Cos(timeOfDay*0.02+phase)) * 2.5
		radius := float32(11.0) + pulse
		if radius < 5.0 {
			radius = 5.0
		}
		vector.FillCircle(screen, cx, cy, radius, color.RGBA{0, 220, 240, 60}, false)
		vector.FillCircle(screen, cx, cy, 7, bulbColor, false)
		vector.StrokeCircle(screen, cx, cy, 7, 0.8, color.RGBA{255, 255, 255, 180}, false)
	}
}

// ---------------------------------------------------------
// 3. BRIMSTONE SIPHON (Volcanic Vent)
// ---------------------------------------------------------

type BrimstoneSiphon struct {
	BaseEntity
	Timer     int
	Direction string // "up", "down", "left", "right"
}

func (ent *BrimstoneSiphon) Update(gr Runtime, CaveGrid [][]bool) {
	// Cycles timers
	ent.Timer = (ent.Timer + 1) % 120
	if ent.Timer >= 60 {
		// Erupting steam/fire jet!
		// Check box overlap based on siphon jet direction
		var jx, jy, jw, jh float64
		const jetRange = 160.0 // Jet extends 2.5 tiles

		switch ent.Direction {
		case "up":
			jx, jy, jw, jh = ent.Pos.X, ent.Pos.Y-jetRange, ent.Dimensions.X, jetRange
		case "down":
			jx, jy, jw, jh = ent.Pos.X, ent.Pos.Y+ent.Dimensions.Y, ent.Dimensions.X, jetRange
		case "left":
			jx, jy, jw, jh = ent.Pos.X-jetRange, ent.Pos.Y, jetRange, ent.Dimensions.Y
		default: // "right"
			jx, jy, jw, jh = ent.Pos.X+ent.Dimensions.X, ent.Pos.Y, jetRange, ent.Dimensions.Y
		}

		// Check overlap with player/vehicle
		vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
		targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
		if gr.HasActiveVehicle() {
			vPos := gr.ActiveVehiclePos()
			targetX, targetY = vPos.X, vPos.Y
			vDims := gr.ActiveVehicleDims()
			vWidth, vHeight = vDims.X, vDims.Y
		}

		if rectsOverlap(jx, jy, jw, jh, targetX, targetY, vWidth, vHeight) {
			if gr.HasActiveVehicle() {
				gr.Emit(DamageActiveVehicleCmd{Amount: 0.4})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: 0.6})
			}
		}
	}
}

var (
	entityPath = &vector.Path{}
)

func (ent *BrimstoneSiphon) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	// Draw static volcanic vent cone
	entityPath.Reset()
	entityPath.MoveTo(cx-16, sy+32)
	entityPath.LineTo(cx+16, sy+32)
	entityPath.LineTo(cx+8, sy+12)
	entityPath.LineTo(cx-8, sy+12)
	entityPath.Close()

	var ventColor color.RGBA
	if ent.Timer >= 60 {
		ventColor = color.RGBA{185, 85, 45, 255} // Glowing red hot
	} else {
		ventColor = color.RGBA{65, 55, 50, 255} // Cool rock
	}

	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(ventColor)
	vector.FillPath(screen, entityPath, nil, &opts)

	// Draw fire/steam jet particles if erupting
	if ent.Timer >= 60 {
		jetLen := float32(120.0) // Eruption visible length
		var jx, jy float32
		switch ent.Direction {
		case "up":
			jx, jy = cx, cy-jetLen/2.0
			vector.FillRect(screen, cx-8, cy-jetLen, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-3, cy-jetLen-10, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
		case "down":
			jx, jy = cx, cy+jetLen/2.0
			vector.FillRect(screen, cx-8, cy+16, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-3, cy+16, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
		case "left":
			jx, jy = cx-jetLen/2.0, cy
			vector.FillRect(screen, cx-jetLen-16, cy-8, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-jetLen-26, cy-3, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
		default: // "right"
			jx, jy = cx+jetLen/2.0, cy
			vector.FillRect(screen, cx+16, cy-8, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx+16, cy-3, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
		}
		_ = jx
		_ = jy
	}
}

// ---------------------------------------------------------
// 4. THERMOCLINE RAMMER (Swimming Predator)
// ---------------------------------------------------------

type ThermoclineRammer struct {
	BaseEntity
	State     int
	Timer     int
	Facing    float64
	StunTimer int
}

func (ent *ThermoclineRammer) Update(gr Runtime, CaveGrid [][]bool) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2.0
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0
	dist := math.Hypot(px-ex, py-ey)

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
		if !gr.HasActiveVehicle() && gr.IsPlayerSprinting() && (math.Abs(gr.PlayerVel().X) > 1.2 || math.Abs(gr.PlayerVel().Y) > 1.2) {
			isAggroTrigger = true
		}
		// Mech thrusters or vehicle moving fast
		if gr.HasActiveVehicle() && gr.ActiveVehicleMoving() {
			isAggroTrigger = true
		}
	}
	// Also triggered by sound pops
	if gr.SoundWaveTimer() > 0 && math.Hypot(gr.SoundWaveX()-ex, gr.SoundWaveY()-ey) < 250.0 {
		isAggroTrigger = true
	}

	if ent.State == 0 { // Patrol
		if isAggroTrigger {
			ent.State = 1 // Alert/Charge state
			// Lock charge direction to closest orthogonal axis
			dx := px - ex
			dy := py - ey
			if math.Abs(dx) > math.Abs(dy) {
				ent.Vel.Y = 0
				if dx > 0 {
					ent.Vel.X = 6.2
					ent.Facing = 0.0
				} else {
					ent.Vel.X = -6.2
					ent.Facing = math.Pi
				}
			} else {
				ent.Vel.X = 0
				if dy > 0 {
					ent.Vel.Y = 6.2
					ent.Facing = math.Pi / 2.0
				} else {
					ent.Vel.Y = -6.2
					ent.Facing = -math.Pi / 2.0
				}
			}
		} else {
			// Swim idle patrol back and forth
			ent.Timer++
			if ent.Timer%120 == 0 {
				ent.Facing += math.Pi // turn around
			}
			ent.Vel.X = math.Cos(ent.Facing) * 0.8
			ent.Vel.Y = math.Sin(ent.Facing) * 0.4
			if !isSolid(CaveGrid, ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
				ent.Pos = ent.Pos.Add(ent.Vel)
			} else {
				ent.Facing += math.Pi // turn around on wall bump
			}
		}
	} else if ent.State == 1 { // Charging
		// Move in straight line
		nextX := ent.Pos.X + ent.Vel.X
		nextY := ent.Pos.Y + ent.Vel.Y

		// Check wall collision
		if isSolid(CaveGrid, nextX, nextY, ent.Dimensions.X, ent.Dimensions.Y) {
			ent.State = 2 // Stunned!
			ent.StunTimer = 180
			ent.Vel = gvec.Vec2{}
		} else {
			ent.Pos.X = nextX
			ent.Pos.Y = nextY
		}

		// Check player/vehicle damage overlap
		vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
		targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
		if gr.HasActiveVehicle() {
			vPos := gr.ActiveVehiclePos()
			targetX, targetY = vPos.X, vPos.Y
			vDims := gr.ActiveVehicleDims()
			vWidth, vHeight = vDims.X, vDims.Y
		}

		if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
			if gr.HasActiveVehicle() {
				gr.Emit(DamageActiveVehicleCmd{Amount: 30.0})
				gr.Emit(SetMineWarningCmd{Message: "VEHICLE RAMMED BY THERMOCLINE RAMMER!", Duration: 120})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: 25.0})
				gr.Emit(SetMineWarningCmd{Message: "RAMMED BY THERMOCLINE RAMMER!", Duration: 120})
			}
			ent.State = 0 // return to patrol
			ent.Vel = gvec.Vec2{}
		}
	}
}

func (ent *ThermoclineRammer) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	// Draw fish body (grey-orange)
	bodyColor := color.RGBA{195, 95, 45, 255}
	vector.FillCircle(screen, cx, cy, 8.0, bodyColor, false)

	// Hard head skull (grey triangle)
	cosF := float32(math.Cos(ent.Facing))
	sinF := float32(math.Sin(ent.Facing))
	entityPath.Reset()
	hx := cx + cosF*12
	hy := cy + sinF*12
	entityPath.MoveTo(hx, hy)
	entityPath.LineTo(cx-sinF*6, cy+cosF*6)
	entityPath.LineTo(cx+sinF*6, cy-cosF*6)
	entityPath.Close()

	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(color.RGBA{120, 130, 140, 255})
	vector.FillPath(screen, entityPath, nil, &opts)

	// Tail fin
	tx := cx - cosF*10
	ty := cy - sinF*10
	vector.StrokeLine(screen, tx, ty, tx-sinF*8, ty+cosF*8, 2.0, color.RGBA{195, 95, 45, 255}, false)
	vector.StrokeLine(screen, tx, ty, tx+sinF*8, ty-cosF*8, 2.0, color.RGBA{195, 95, 45, 255}, false)

	// Stun stars above head if stunned
	if ent.State == 2 {
		starAng := float64(ent.StunTimer) * 0.15
		sx1 := cx + float32(math.Cos(starAng))*14
		sy1 := cy - 14 + float32(math.Sin(starAng))*5
		sx2 := cx + float32(math.Cos(starAng+math.Pi))*14
		sy2 := cy - 14 + float32(math.Sin(starAng+math.Pi))*5
		vector.FillCircle(screen, sx1, sy1, 2.5, color.RGBA{255, 230, 40, 255}, false)
		vector.FillCircle(screen, sx2, sy2, 2.5, color.RGBA{255, 230, 40, 255}, false)
	}
}

// ---------------------------------------------------------
// 5. PALID NERVE-MAT (Flora Carpet)
// ---------------------------------------------------------

type NerveMat struct {
	BaseEntity
}

func (ent *NerveMat) Update(gr Runtime, CaveGrid [][]bool) {
	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}

	if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		gr.Emit(SetPlayerSlowedCmd{Slowed: true})
	}
}

func (ent *NerveMat) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)

	// Purple nerve carpet on floor
	vector.FillRect(screen, sx, sy+sh-4, sw, 4, color.RGBA{80, 25, 120, 255}, false)
	// Small vertical fibers
	for o := float32(4); o < sw; o += 12 {
		vector.StrokeLine(screen, sx+o, sy+sh, sx+o, sy+sh-8, 1.5, color.RGBA{130, 40, 180, 255}, false)
		vector.FillCircle(screen, sx+o, sy+sh-8, 2.0, color.RGBA{180, 60, 220, 255}, false)
	}
}

// ---------------------------------------------------------
// 6. ELECTRO-WEAVER (Electricity-Seeking Serpentine Predator)
// ---------------------------------------------------------

type ElectroWeaver struct {
	BaseEntity
	Timer  int
	Facing float64
}

func (ent *ElectroWeaver) Update(gr Runtime, CaveGrid [][]bool) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2.0
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2.0
	ex := ent.Pos.X + ent.Dimensions.X/2.0
	ey := ent.Pos.Y + ent.Dimensions.Y/2.0
	dist := math.Hypot(px-ex, py-ey)

	inAbyssal := (py / config.TileSize) >= 80
	if !inAbyssal {
		ent.Timer = 0
		return
	}

	isElectricity := gr.FlashlightOn() || gr.SonarActive() || gr.HasActiveVehicle()
	if isElectricity && dist < 500.0 {
		ent.Timer++
		// Feed tracking value to game screen static/jitter
		gr.Emit(UpdateWeaverTrackingTimerCmd{Value: float64(ent.Timer)})

		// Strike check (5 seconds at 60 FPS)
		if ent.Timer >= 300 {
			gr.Emit(DamagePlayerCmd{Amount: 45.0})
			gr.Emit(SetMineWarningCmd{Message: "ELECTRO-WEAVER STRIKE! SEVERE DAMAGE!", Duration: 180})

			// Teleport Weaver near player to simulate lunging strike
			ent.Pos.X = gr.PlayerPos().X + float64(rand.Intn(120)-60)
			ent.Pos.Y = gr.PlayerPos().Y + float64(rand.Intn(120)-60)
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
			ent.Vel.X = (dx / dDist) * 1.5
			ent.Vel.Y = (dy / dDist) * 1.5
		} else {
			ent.Vel.X = math.Cos(gr.TimeOfDay()/30.0) * 1.2
			ent.Vel.Y = math.Sin(gr.TimeOfDay()/30.0) * 1.2
		}
	} else {
		ent.Vel.X = math.Cos(gr.TimeOfDay()/40.0) * 0.8
		ent.Vel.Y = math.Sin(gr.TimeOfDay()/40.0) * 0.8
	}

	// Soft collisions for slithering
	if !isSolid(CaveGrid, ent.Pos.X+ent.Vel.X, ent.Pos.Y+ent.Vel.Y, ent.Dimensions.X, ent.Dimensions.Y) {
		ent.Pos = ent.Pos.Add(ent.Vel)
	}
}

func (ent *ElectroWeaver) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	// Slithering snake bodies
	bodyParts := 5
	for i := 0; i < bodyParts; i++ {
		lag := float64(i) * 0.3
		tVal := timeOfDay*0.08 - lag
		offX := math.Cos(tVal) * 6
		offY := math.Sin(tVal) * 4

		segmentX := cx - float32(math.Cos(ent.Facing)*float64(i)*8.0) + float32(offX)
		segmentY := cy - float32(math.Sin(ent.Facing)*float64(i)*8.0) + float32(offY)

		segColor := color.RGBA{140 - uint8(i*18), 45, 205 - uint8(i*12), 255}
		vector.FillCircle(screen, segmentX, segmentY, 6.0-float32(i)*0.8, segColor, false)

		// Glow indicator on head
		if i == 0 {
			vector.FillCircle(screen, segmentX+float32(math.Cos(ent.Facing))*4, segmentY+float32(math.Sin(ent.Facing))*4, 2.0, color.RGBA{255, 255, 80, 255}, false)
		}
	}

	// Draw electrical discharge sparks if tracking
	if ent.Timer > 0 {
		sparkRatio := float64(ent.Timer) / 300.0
		numSparks := int(sparkRatio * 5)
		for s := 0; s < numSparks; s++ {
			spx := cx + float32(rand.Intn(40)-20)
			spy := cy + float32(rand.Intn(40)-20)
			vector.StrokeLine(screen, cx, cy, spx, spy, 1.0, color.RGBA{160, 220, 255, 255}, false)
		}
	}
}

// ---------------------------------------------------------
// 7. PASSIVE FISH (Catchable Swimming Creature)
// ---------------------------------------------------------

type PassiveFish struct {
	BaseEntity
	FacingRight bool
	SwimPhase   float64
	FleeTimer   int
}

func (f *PassiveFish) GetHarvestedItem() item.Item {
	return &item.RawFish{}
}

func (f *PassiveFish) CanCatch(playerPos gvec.Vec2) bool {
	cx := f.Pos.X + f.Dimensions.X/2
	cy := f.Pos.Y + f.Dimensions.Y/2
	dist := math.Hypot(playerPos.X-cx, playerPos.Y-cy)
	return dist <= 80.0
}

func (f *PassiveFish) Update(gr Runtime, CaveGrid [][]bool) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2
	fx := f.Pos.X + f.Dimensions.X/2
	fy := f.Pos.Y + f.Dimensions.Y/2
	dist := math.Hypot(px-fx, py-fy)

	f.SwimPhase += 0.04

	if f.FleeTimer > 0 {
		f.FleeTimer--
		// Dart away quickly
		speed := 3.5
		if f.FacingRight {
			f.Vel.X = speed
		} else {
			f.Vel.X = -speed
		}
		f.Vel.Y = math.Sin(f.SwimPhase*2) * 1.0
	} else if dist < 120 {
		// Flee from player
		f.FleeTimer = 60
		f.FacingRight = px < fx // swim away from player
	} else {
		// Gentle sine-wave swimming
		speed := 0.6
		if f.FacingRight {
			f.Vel.X = speed
		} else {
			f.Vel.X = -speed
		}
		f.Vel.Y = math.Sin(f.SwimPhase) * 0.4
	}

	nextX := f.Pos.X + f.Vel.X
	nextY := f.Pos.Y + f.Vel.Y

	if !isSolid(CaveGrid, nextX, nextY, f.Dimensions.X, f.Dimensions.Y) {
		f.Pos.X = nextX
		f.Pos.Y = nextY
	} else {
		// Turn around on wall bump
		f.FacingRight = !f.FacingRight
		f.FleeTimer = 0
	}
}

func (f *PassiveFish) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(f.Pos.X - camera.Pos.X)
	sy := float32(f.Pos.Y - camera.Pos.Y)
	sw := float32(f.Dimensions.X)
	sh := float32(f.Dimensions.Y)
	cx := sx + sw/2
	cy := sy + sh/2

	// Body colors: blue-green iridescent
	bodyColor := color.RGBA{60, 160, 200, 255}
	finColor := color.RGBA{40, 130, 180, 200}

	// Body oval
	vector.FillCircle(screen, cx, cy, 6.0, bodyColor, false)

	// Tail fin (triangle pointing opposite to facing)
	var tailX float32
	if f.FacingRight {
		tailX = cx - 8
	} else {
		tailX = cx + 8
	}
	wiggle := float32(math.Sin(timeOfDay*0.12+float64(f.SwimPhase))) * 3
	entityPath.Reset()
	entityPath.MoveTo(tailX, cy)
	entityPath.LineTo(tailX-4+wiggle, cy-5)
	entityPath.LineTo(tailX-4+wiggle, cy+5)
	entityPath.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(finColor)
	vector.FillPath(screen, entityPath, nil, &opts)

	// Eye
	var eyeX float32
	if f.FacingRight {
		eyeX = cx + 3
	} else {
		eyeX = cx - 3
	}
	vector.FillCircle(screen, eyeX, cy-1.5, 1.5, color.White, false)
	vector.FillCircle(screen, eyeX, cy-1.5, 0.8, color.Black, false)

	// Subtle dorsal stripe
	vector.StrokeLine(screen, cx-4, cy-3, cx+4, cy-3, 0.8, color.RGBA{80, 200, 240, 180}, false)
}

// ---------------------------------------------------------
// 8. PASSIVE CRAB (Catchable Floor Creature)
// ---------------------------------------------------------

type PassiveCrab struct {
	BaseEntity
	FacingRight bool
	InShell     bool
	ShellTimer  int
	WalkTimer   int
}

func (c *PassiveCrab) GetHarvestedItem() item.Item {
	return &item.RawCrab{}
}

func (c *PassiveCrab) CanCatch(playerPos gvec.Vec2) bool {
	cx := c.Pos.X + c.Dimensions.X/2
	cy := c.Pos.Y + c.Dimensions.Y/2
	dist := math.Hypot(playerPos.X-cx, playerPos.Y-cy)
	return dist <= 64.0
}

func (c *PassiveCrab) Update(gr Runtime, CaveGrid [][]bool) {
	px := gr.PlayerPos().X + gr.PlayerDims().X/2
	py := gr.PlayerPos().Y + gr.PlayerDims().Y/2
	cx := c.Pos.X + c.Dimensions.X/2
	cy := c.Pos.Y + c.Dimensions.Y/2
	dist := math.Hypot(px-cx, py-cy)

	// Check if flashlight is pointing at the crab
	isLit := false
	if gr.FlashlightOn() && dist < 300 {
		facingAngle := gr.PlayerFacing()
		dx := cx - px
		dy := cy - py
		angleToEnt := math.Atan2(dy, dx)
		diff := angleToEnt - facingAngle
		for diff > math.Pi {
			diff -= 2 * math.Pi
		}
		for diff < -math.Pi {
			diff += 2 * math.Pi
		}
		if math.Abs(diff) < 0.42 {
			isLit = true
		}
	}

	// Retract into shell if player is close or flashlight is on them
	if dist < 100 || isLit {
		c.InShell = true
		c.ShellTimer = 90
		c.Vel.X = 0
	} else if c.ShellTimer > 0 {
		c.ShellTimer--
		if c.ShellTimer <= 0 {
			c.InShell = false
		}
	}

	if !c.InShell {
		c.WalkTimer++
		// Change direction periodically
		if c.WalkTimer%180 == 0 {
			c.FacingRight = !c.FacingRight
		}

		speed := 0.35
		if c.FacingRight {
			c.Vel.X = speed
		} else {
			c.Vel.X = -speed
		}

		// Apply gravity
		c.Vel.Y += 0.3
		if c.Vel.Y > 4.0 {
			c.Vel.Y = 4.0
		}
	} else {
		c.Vel.Y += 0.3
		if c.Vel.Y > 4.0 {
			c.Vel.Y = 4.0
		}
	}

	// Horizontal movement
	nextX := c.Pos.X + c.Vel.X
	if !isSolid(CaveGrid, nextX, c.Pos.Y, c.Dimensions.X, c.Dimensions.Y) {
		c.Pos.X = nextX
	} else {
		c.FacingRight = !c.FacingRight
		c.Vel.X = 0
	}

	// Vertical movement (gravity)
	nextY := c.Pos.Y + c.Vel.Y
	if !isSolid(CaveGrid, c.Pos.X, nextY, c.Dimensions.X, c.Dimensions.Y) {
		c.Pos.Y = nextY
	} else {
		c.Vel.Y = 0
	}
}

func (c *PassiveCrab) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	sw := float32(c.Dimensions.X)
	sh := float32(c.Dimensions.Y)
	ccx := sx + sw/2
	ccy := sy + sh/2

	shellColor := color.RGBA{180, 60, 50, 255}
	legColor := color.RGBA{160, 45, 40, 255}

	if c.InShell {
		// Retracted shell: just a circle
		vector.FillCircle(screen, ccx, ccy, 6.0, shellColor, false)
		vector.StrokeCircle(screen, ccx, ccy, 6.0, 1.0, color.RGBA{140, 40, 35, 255}, false)
		// Shell lines
		vector.StrokeLine(screen, ccx-3, ccy-2, ccx+3, ccy-2, 0.8, color.RGBA{130, 35, 30, 255}, false)
		vector.StrokeLine(screen, ccx-2, ccy+1, ccx+2, ccy+1, 0.8, color.RGBA{130, 35, 30, 255}, false)
		return
	}

	// Body
	vector.FillCircle(screen, ccx, ccy, 5.0, shellColor, false)

	// Legs (animated with walk cycle)
	legWiggle := float32(math.Sin(timeOfDay*0.15+float64(c.WalkTimer)*0.1)) * 2

	// Left legs
	vector.StrokeLine(screen, ccx-4, ccy+2, ccx-7, ccy+5+legWiggle, 1.2, legColor, false)
	vector.StrokeLine(screen, ccx-3, ccy+3, ccx-6, ccy+6-legWiggle, 1.2, legColor, false)

	// Right legs
	vector.StrokeLine(screen, ccx+4, ccy+2, ccx+7, ccy+5-legWiggle, 1.2, legColor, false)
	vector.StrokeLine(screen, ccx+3, ccy+3, ccx+6, ccy+6+legWiggle, 1.2, legColor, false)

	// Claws
	clawColor := color.RGBA{200, 70, 55, 255}
	if c.FacingRight {
		vector.FillCircle(screen, ccx+7, ccy-2, 2.5, clawColor, false)
		vector.FillCircle(screen, ccx-6, ccy-1, 2.0, clawColor, false)
	} else {
		vector.FillCircle(screen, ccx-7, ccy-2, 2.5, clawColor, false)
		vector.FillCircle(screen, ccx+6, ccy-1, 2.0, clawColor, false)
	}

	// Eyes (little stalks)
	vector.StrokeLine(screen, ccx-2, ccy-4, ccx-2, ccy-7, 0.8, legColor, false)
	vector.FillCircle(screen, ccx-2, ccy-7, 1.0, color.White, false)
	vector.StrokeLine(screen, ccx+2, ccy-4, ccx+2, ccy-7, 0.8, legColor, false)
	vector.FillCircle(screen, ccx+2, ccy-7, 1.0, color.White, false)
}

// ---------------------------------------------------------
// 9. KELP (Decorative Swaying Plant)
// ---------------------------------------------------------

type Kelp struct {
	BaseEntity
	SwayPhase float64
}

func (k *Kelp) Update(gr Runtime, CaveGrid [][]bool) {
	k.SwayPhase += 0.03
}

func (k *Kelp) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(k.Pos.X - camera.Pos.X)
	sy := float32(k.Pos.Y - camera.Pos.Y)
	sw := float32(k.Dimensions.X)
	sh := float32(k.Dimensions.Y)
	cx := sx + sw/2.0
	bottomY := sy + sh

	numSegments := int(sh / 8.0)
	if numSegments < 3 {
		numSegments = 3
	}
	segmentHeight := sh / float32(numSegments)

	lastX := cx
	lastY := bottomY

	for i := 0; i < numSegments; i++ {
		factor := float64(i+1) / float64(numSegments)
		swayOffset := float32(math.Sin(k.SwayPhase+float64(i)*0.4)) * 8.0 * float32(factor)

		nextX := cx + swayOffset
		nextY := bottomY - float32(i+1)*segmentHeight

		// Draw stalk segment
		vector.StrokeLine(screen, lastX, lastY, nextX, nextY, 2.5-float32(factor)*1.0, color.RGBA{34, 139, 34, 255}, false)

		// Draw leaves on the sides of the segment junction
		leafColor := color.RGBA{46, 150, 60, 220}
		leafSize := (5.0 - float32(factor)*2.0)
		if leafSize < 2.0 {
			leafSize = 2.0
		}

		vector.FillCircle(screen, nextX-4.0, nextY, leafSize, leafColor, false)
		vector.FillCircle(screen, nextX+4.0, nextY, leafSize, leafColor, false)

		lastX = nextX
		lastY = nextY
	}
}

// ---------------------------------------------------------
// HELPERS
// ---------------------------------------------------------

func rectsOverlap(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

func isSolid(grid [][]bool, x, y, w, h float64) bool {
	if grid == nil {
		return false
	}
	gridW := len(grid)
	gridH := len(grid[0])

	x1 := int(math.Floor(x)) / config.TileSize
	x2 := int(math.Floor(x+w)) / config.TileSize
	y1 := int(math.Floor(y)) / config.TileSize
	y2 := int(math.Floor(y+h)) / config.TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= gridW {
				return true
			}
			if ty < 0 {
				continue
			}
			if ty >= gridH {
				return true
			}
			if grid[tx][ty] {
				return true
			}
		}
	}
	return false
}
