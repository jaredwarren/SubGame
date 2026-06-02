package vehicle

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

var (
	heavyMechSheet       *ebiten.Image
	heavyMechSheetOnce   sync.Once
	heavyMechSheetLoaded bool
)

func loadHeavyMechSheetLazy() {
	if heavyMechSheetLoaded {
		return
	}
	heavyMechSheetLoaded = true

	paths := []string{
		"assets/textures/heavy_mech.png",
		"/Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/textures/heavy_mech.png",
		"../../assets/textures/heavy_mech.png",
		"../assets/textures/heavy_mech.png",
		"../../../assets/textures/heavy_mech.png",
	}

	var file *os.File
	var err error
	for _, p := range paths {
		file, err = os.Open(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Printf("Warning: Failed to open assets/textures/heavy_mech.png: %v", err)
		return
	}
	defer func() { _ = file.Close() }()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Warning: Failed to decode assets/textures/heavy_mech.png: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			clr := img.At(x, y)
			r, g, b, a := clr.RGBA()
			ru := uint8(r >> 8)
			gu := uint8(g >> 8)
			bu := uint8(b >> 8)
			au := uint8(a >> 8)

			if gu > 140 && ru < 100 && bu < 100 {
				rgba.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				rgba.SetRGBA(x, y, color.RGBA{ru, gu, bu, au})
			}
		}
	}

	heavyMechSheet = ebiten.NewImageFromImage(rgba)
}

// HeavyMech is a cave-walker with a drill arm, high depth tolerance, and thruster propulsion.
type HeavyMech struct {
	Pos             gvec.Vec2
	Vel             gvec.Vec2
	Dimensions      gvec.Vec2
	Facing          float64
	Health          float64
	MaxHealth       float64
	Battery         float64
	MaxBattery      float64
	Cargo           *item.Inventory
	IsDrilling      bool
	DrillTimer      int
	TargetDrillNode DrillableResource
	ThrustersActive bool
	AnimTick        int
}

// NewHeavyMech creates a HeavyMech at the given world position.
func NewHeavyMech(x, y float64) *HeavyMech {
	return &HeavyMech{
		Pos:        gvec.Vec2{X: x, Y: y},
		Dimensions: gvec.Vec2{X: 48, Y: 48},
		Health:     200.0,
		MaxHealth:  200.0,
		Battery:    100.0,
		MaxBattery: 100.0,
		Cargo:      item.NewInventory(8),
	}
}

func (m *HeavyMech) GetPos() gvec.Vec2            { return m.Pos }
func (m *HeavyMech) SetPos(pos gvec.Vec2)         { m.Pos = pos }
func (m *HeavyMech) GetDimensions() gvec.Vec2     { return m.Dimensions }
func (m *HeavyMech) GetHealth() float64           { return m.Health }
func (m *HeavyMech) GetMaxHealth() float64        { return m.MaxHealth }
func (m *HeavyMech) GetOxygen() float64           { return 100.0 }
func (m *HeavyMech) GetDepthLimit() float64       { return 120.0 }
func (m *HeavyMech) GetCargo() *item.Inventory    { return m.Cargo }
func (m *HeavyMech) GetUpgrades() *item.Inventory { return nil }
func (m *HeavyMech) GetPerspective() string       { return "cave" }
func (m *HeavyMech) GetName() string              { return "Heavy Mech" }
func (m *HeavyMech) GetBattery() float64          { return m.Battery }
func (m *HeavyMech) GetMaxBattery() float64       { return m.MaxBattery }
func (m *HeavyMech) GetFacing() float64           { return m.Facing }

func (m *HeavyMech) TakeDamage(amount float64) {
	m.Health -= amount * 0.6 // 40% damage reduction
	if m.Health < 0 {
		m.Health = 0
	}
}

func (m *HeavyMech) RechargeBattery(amount float64) {
	m.Battery += amount
	if m.Battery > m.MaxBattery {
		m.Battery = m.MaxBattery
	}
}

func (m *HeavyMech) Update(runtime Runtime) {
	m.AnimTick++
	if !runtime.IsActiveVehicle(m) {
		m.Vel.Y += 0.12
		m.Vel.Y *= 0.95
		m.Vel.X = 0
		m.checkCollisions(runtime)
		return
	}

	input := runtime.Input()
	cursor := input.Cursor()
	center := runtime.PlayerScreenCenter()
	m.Facing = math.Atan2(cursor.Y-center.Y, cursor.X-center.X)

	const gravity = 0.12
	const dragH = 0.88
	const dragV = 0.95
	var walkForce = 0.35
	var maxSpeedH = 2.0

	hasPower := m.Battery > 0
	if !hasPower {
		walkForce = 0.08
		maxSpeedH = 0.6
	}
	if runtime.PlayerSlowed() {
		walkForce *= 0.5
		maxSpeedH *= 0.5
	}

	moving := false
	if input.IsKeyPressed(ebiten.KeyA) || input.IsKeyPressed(ebiten.KeyArrowLeft) {
		m.Vel.X -= walkForce
		moving = true
	}
	if input.IsKeyPressed(ebiten.KeyD) || input.IsKeyPressed(ebiten.KeyArrowRight) {
		m.Vel.X += walkForce
		moving = true
	}

	m.Vel.Y += gravity

	const waterline = -12.0
	m.ThrustersActive = false
	if hasPower && (input.IsKeyPressed(ebiten.KeyW) || input.IsKeyPressed(ebiten.KeyArrowUp) || input.IsKeyPressed(ebiten.KeySpace)) {
		if m.Pos.Y > waterline {
			m.Vel.Y -= 0.28
			m.Battery -= 0.08
			if m.Battery < 0 {
				m.Battery = 0
			}
			m.ThrustersActive = true
			if rand.Float64() < 0.4 {
				runtime.Emit(SpawnBubbleCmd{Pos: gvec.Vec2{X: m.Pos.X + 14, Y: m.Pos.Y + m.Dimensions.Y - 14}})
				runtime.Emit(SpawnBubbleCmd{Pos: gvec.Vec2{X: m.Pos.X + m.Dimensions.X - 22, Y: m.Pos.Y + m.Dimensions.Y - 14}})
			}
		}
	}

	if moving && hasPower {
		m.Battery -= 0.01
		if m.Battery < 0 {
			m.Battery = 0
		}
	}

	m.Vel.X *= dragH
	m.Vel.Y *= dragV

	if math.Abs(m.Vel.X) > maxSpeedH {
		m.Vel.X = math.Copysign(maxSpeedH, m.Vel.X)
	}

	if m.Pos.Y < waterline {
		m.Vel.Y += 0.20
	} else if m.ThrustersActive && m.Pos.Y < waterline+16.0 {
		bobY := waterline + 4.0 + math.Sin(float64(runtime.TimeOfDay())*0.05)*2.0
		m.Vel.Y += (bobY - m.Pos.Y) * 0.05
	}

	m.checkCollisions(runtime)

	if m.IsDrilling {
		m.DrillTimer--
		if m.DrillTimer <= 0 {
			m.IsDrilling = false
			if m.TargetDrillNode != nil && m.TargetDrillNode.GetHitsToMine() > 0 {
				m.TargetDrillNode.SetHitsToMine(m.TargetDrillNode.GetHitsToMine() - 1)

				targetTx, targetTy := m.TargetDrillNode.GetTilePos()
				drillPos := gvec.Vec2{
					X: float64(targetTx*TileSize + TileSize/2),
					Y: float64(targetTy*TileSize + TileSize/2),
				}
				nodeColor := color.RGBA{150, 150, 150, 255}
				if cRgba, ok := m.TargetDrillNode.GetColor().(color.RGBA); ok {
					nodeColor = cRgba
				}
				runtime.Emit(SpawnDebrisCmd{Pos: drillPos, Color: nodeColor})

				if m.TargetDrillNode.GetHitsToMine() <= 0 {
					recipeName := m.TargetDrillNode.GetRecipeResultName()
					if recipeName != "" {
						runtime.Emit(UnlockRecipeCmd{RecipeResultName: recipeName})
					} else {
						m.Cargo.AddItem(m.TargetDrillNode, 1)
					}
					runtime.Emit(RemoveCaveNodeCmd{TX: targetTx, TY: targetTy})
				}
			}
			m.TargetDrillNode = nil
		}
	}
}

// DrillStrike initiates a drill animation against the given resource node.
func (m *HeavyMech) DrillStrike(node DrillableResource) {
	m.IsDrilling = true
	m.DrillTimer = 15
	m.TargetDrillNode = node
}

func (m *HeavyMech) checkCollisions(runtime Runtime) {
	newX := m.Pos.X + m.Vel.X
	if m.isSolid(runtime, gvec.Vec2{X: newX, Y: m.Pos.Y}) {
		m.Vel.X = 0
	} else {
		m.Pos.X = newX
	}
	newY := m.Pos.Y + m.Vel.Y
	if m.isSolid(runtime, gvec.Vec2{X: m.Pos.X, Y: newY}) {
		if m.Vel.Y > 4.5 {
			m.TakeDamage((m.Vel.Y - 4.5) * 8.0)
			runtime.Emit(TriggerShakeCmd{Duration: 20, Intensity: (m.Vel.Y - 4.5) * 3.0})
		} else if m.Vel.Y > 2.0 {
			runtime.Emit(TriggerShakeCmd{Duration: 10, Intensity: (m.Vel.Y - 2.0) * 1.5})
		}
		m.Vel.Y = 0
	} else {
		m.Pos.Y = newY
	}
}

func (m *HeavyMech) isSolid(runtime Runtime, pos gvec.Vec2) bool {
	x1 := int(math.Floor(pos.X)) / TileSize
	x2 := int(math.Floor(pos.X+m.Dimensions.X)) / TileSize
	y1 := int(math.Floor(pos.Y)) / TileSize
	y2 := int(math.Floor(pos.Y+m.Dimensions.Y)) / TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if runtime.IsCaveSolidAt(tx, ty) {
				return true
			}
		}
	}
	return false
}

func (m *HeavyMech) Draw(screen *ebiten.Image, camX, camY float64) {
	sx := float32(m.Pos.X - camX)
	sy := float32(m.Pos.Y - camY)
	w := float32(m.Dimensions.X)
	h := float32(m.Dimensions.Y)

	isFacingRight := math.Cos(m.Facing) >= 0

	loadHeavyMechSheetLazy()

	if heavyMechSheet != nil {
		col := 0
		row := 0

		if m.Battery > 0 && m.ThrustersActive {
			row = 2
			col = 0
		} else if m.IsDrilling {
			row = 1
			col = (m.AnimTick / 4) % 4
		} else if math.Abs(m.Vel.X) > 0.1 {
			row = 0
			col = (m.AnimTick / 8) % 4
		} else {
			row = 0
			col = 0
		}

		rect := image.Rect(col*704, row*512, (col+1)*704, (row+1)*512)
		activeFrame := heavyMechSheet.SubImage(rect).(*ebiten.Image)

		op := &ebiten.DrawImageOptions{}

		// Center horizontally around 352, and align feet at 492
		op.GeoM.Translate(-352, -492)

		facingSign := 1.0
		if !isFacingRight {
			facingSign = -1.0
		}

		// Scale so the frame has draw width of 120.0
		const frameScale = 120.0 / 704.0
		op.GeoM.Scale(facingSign*frameScale, frameScale)

		// Translate to screen coordinates: horizontal center at sx + w/2, feet aligned at sy + h - 2 (slight overlap)
		spriteX := float64(sx) + float64(w)/2.0
		spriteY := float64(sy) + float64(h) - 2.0
		if m.IsDrilling {
			spriteX += float64(rand.Intn(3) - 1)
			spriteY += float64(rand.Intn(3) - 1)
		}
		op.GeoM.Translate(spriteX, spriteY)

		screen.DrawImage(activeFrame, op)
		return
	}

	// Fallback to original vector drawing code
	mechBodyColor := color.RGBA{218, 98, 16, 255}
	visorColor := color.RGBA{80, 200, 255, 230}
	frameColor := color.RGBA{58, 68, 78, 255}

	vector.FillRect(screen, sx+8, sy+h-14, 8, 14, frameColor, false)
	vector.FillRect(screen, sx+w-16, sy+h-14, 8, 14, frameColor, false)
	vector.FillCircle(screen, sx+12, sy+h-14, 5, color.RGBA{38, 48, 58, 255}, false)
	vector.FillCircle(screen, sx+w-12, sy+h-14, 5, color.RGBA{38, 48, 58, 255}, false)

	vector.FillRect(screen, sx+4, sy+4, w-8, h-16, mechBodyColor, false)
	vector.StrokeRect(screen, sx+4, sy+4, w-8, h-16, 1.5, color.RGBA{250, 160, 50, 255}, false)

	if isFacingRight {
		vector.FillRect(screen, sx+w-14, sy+8, 8, 12, visorColor, false)
		vector.FillRect(screen, sx-6, sy+14, 10, 6, frameColor, false)
		vector.FillRect(screen, sx-6, sy+10, 4, 14, frameColor, false)

		drillX := sx + w
		drillY := sy + 18
		if m.IsDrilling {
			drillX += float32(rand.Intn(3) - 1)
			drillY += float32(rand.Intn(3) - 1)
		}
		dx1, dy1 := drillX, drillY
		dx2, dy2 := drillX+12, drillY+4
		dx3, dy3 := drillX, drillY+8
		drawFilledTriangle(screen, dx1, dy1, dx2, dy2, dx3, dy3, color.RGBA{140, 150, 160, 255})
		vector.StrokeLine(screen, dx1, dy1, dx2, dy2, 1.0, color.RGBA{255, 255, 255, 255}, false)
	} else {
		vector.FillRect(screen, sx+6, sy+8, 8, 12, visorColor, false)
		vector.FillRect(screen, sx+w-4, sy+14, 10, 6, frameColor, false)
		vector.FillRect(screen, sx+w+2, sy+10, 4, 14, frameColor, false)

		drillX := sx - 12
		drillY := sy + 18
		if m.IsDrilling {
			drillX += float32(rand.Intn(3) - 1)
			drillY += float32(rand.Intn(3) - 1)
		}
		dx1, dy1 := drillX, drillY+4
		dx2, dy2 := drillX+12, drillY
		dx3, dy3 := drillX+12, drillY+8
		drawFilledTriangle(screen, dx1, dy1, dx2, dy2, dx3, dy3, color.RGBA{140, 150, 160, 255})
		vector.StrokeLine(screen, dx1, dy1, dx2, dy2, 1.0, color.RGBA{255, 255, 255, 255}, false)
	}

	if m.Battery > 0 && m.ThrustersActive {
		flameX1 := sx + 14
		flameY1 := sy + h - 14
		vector.FillCircle(screen, flameX1, flameY1+float32(rand.Intn(6)), 4, color.RGBA{240, 110, 30, 220}, false)
		flameX2 := sx + w - 22
		vector.FillCircle(screen, flameX2, flameY1+float32(rand.Intn(6)), 4, color.RGBA{240, 110, 30, 220}, false)
	}
}
