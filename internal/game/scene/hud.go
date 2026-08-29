package scene

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

// HUD renders the health, oxygen, and stamina bars on the screen.
type HUD struct{}

// NewHUD creates a new HUD renderer.
func NewHUD() *HUD {
	return &HUD{}
}

// Draw renders the stats panel for the player and any active vehicle.
func (h *HUD) Draw(screen *ebiten.Image, g GameContext) {
	p := g.GetPlayer()
	activeVehicle := g.GetActiveVehicle()

	var jx, jy float32
	weaverTimer := g.GetWeaverTrackingTimer()
	if g.GetCurrentState() == StateCave && weaverTimer > 0 {
		mag := float32((weaverTimer / 300.0) * 5.0)
		jx = rand.Float32()*mag - mag/2.0
		jy = rand.Float32()*mag - mag/2.0
	}

	telX := float32(20) + jx
	telY := float32(20) + jy
	telW := float32(240)
	telH := float32(115)

	vector.FillRect(screen, telX, telY, telW, telH, color.RGBA{18, 24, 38, 200}, false)
	vector.StrokeRect(screen, telX, telY, telW, telH, 1.5, color.RGBA{70, 90, 120, 255}, false)

	pulseColor := color.RGBA{45, 215, 120, 255}
	if g.GetCurrentState() == StateCave {
		pulseColor = color.RGBA{45, 175, 215, 255}
	}
	isPulseOn := int(g.GetTimeOfDay()/15)%2 == 0
	if !isPulseOn {
		pulseColor.A = 100
	}
	vector.FillCircle(screen, telX+15, telY+15, 4, pulseColor, false)

	if g.GetCurrentState() == StateOverworld {
		ebitenutil.DebugPrintAt(screen, "SYSTEMS MONITOR", int(telX)+26, int(telY)+8)

		totalMinutes := int(g.GetTimeOfDay() / 14400.0 * 1440.0)
		hour := (totalMinutes/60 + 6) % 24
		minute := totalMinutes % 60
		period := "AM"
		displayHour := hour
		if hour >= 12 {
			period = "PM"
		}
		if hour > 12 {
			displayHour = hour - 12
		}
		if hour == 0 {
			displayHour = 12
		}
		isDay := g.GetTimeOfDay() < 10800
		dayPhase := "Day"
		if !isDay {
			dayPhase = "Night"
		}
		timeText := fmt.Sprintf("Time: %02d:%02d %s (%s)", displayHour, minute, period, dayPhase)

		w := g.GetWorld()
		tx := int(p.Pos.X+p.Width/2) / config.TileSize
		ty := int(p.Pos.Y+p.Height/2) / config.TileSize
		outOfBounds := tx < 0 || tx >= w.Width || ty < 0 || ty >= w.Height

		var posText, biomeText, depthText string
		if outOfBounds {
			posText = "Pos: X:??? Y:???"
			biomeText = "Biome: Ecological Void"
			depthText = "Est. Dive Depth: ???"
		} else {
			posText = fmt.Sprintf("Pos: X:%.0f Y:%.0f", p.Pos.X, p.Pos.Y)

			bID := w.BiomeMap[tx][ty]
			bSpec := world.GetBiomeInfo(bID)
			biomeText = fmt.Sprintf("Biome: %s", bSpec.Name)

			info := world.GetTileInfo(w.OverworldMap[tx][ty])
			if info != nil && info.EstDiveDepth != "" {
				depthText = info.EstDiveDepth
			} else {
				dist := w.DistanceToLand(tx, ty)
				floorY := 6 + int(dist*2.2)
				if floorY < 6 {
					floorY = 6
				}
				if floorY > 60 {
					floorY = 60
				}
				depthText = fmt.Sprintf("Est. Dive Depth: %dm", floorY)
			}
		}

		ebitenutil.DebugPrintAt(screen, timeText, int(telX)+15, int(telY)+28)
		ebitenutil.DebugPrintAt(screen, posText, int(telX)+15, int(telY)+48)
		ebitenutil.DebugPrintAt(screen, biomeText, int(telX)+15, int(telY)+68)
		ebitenutil.DebugPrintAt(screen, depthText, int(telX)+15, int(telY)+88)

	} else if g.GetCurrentState() == StateCave {
		ebitenutil.DebugPrintAt(screen, "DIVE TELEMETRY", int(telX)+26, int(telY)+8)

		var depth float64
		if activeVehicle != nil {
			vPos := activeVehicle.GetPos()
			vDims := activeVehicle.GetDimensions()
			depth = (vPos.Y + vDims.Y/2.0) / config.TileSize
		} else {
			depth = (p.Pos.Y + p.Height/2.0) / config.TileSize
		}

		activeCave := g.GetActiveCave()
		trenchX, trenchY := g.GetActiveTrenchCoords()
		w := g.GetWorld()
		if activeCave != nil && activeCave.GetCaveType() != cave.CaveOrganicShallow {
			if w != nil && trenchX >= 0 && trenchX < w.Width && trenchY >= 0 && trenchY < w.Height {
				tt := w.OverworldMap[trenchX][trenchY]
				if info := world.GetTileInfo(tt); info != nil && info.Subterranean != nil {
					depth += 34.0
				}
			}
		}
		if depth < 0 {
			depth = 0
		}

		pressure := 1.0 + depth*0.1
		depthText := fmt.Sprintf("Depth: %.1fm", depth)
		pressText := fmt.Sprintf("Pressure: %.2f atm", pressure)

		var trenchText, biomeText string
		if activeCave != nil && activeCave.GetCaveType() == cave.CaveVoid {
			trenchText = "Trench Origin: ???"
			biomeText = "Biome: Ecological Void"
		} else {
			trenchText = fmt.Sprintf("Trench Origin: (%d, %d)", trenchX, trenchY)
			if w != nil && trenchX >= 0 && trenchX < w.Width && trenchY >= 0 && trenchY < w.Height {
				bID := w.BiomeMap[trenchX][trenchY]
				bSpec := world.GetBiomeInfo(bID)
				biomeText = fmt.Sprintf("Biome: %s", bSpec.Name)
			} else {
				biomeText = "Biome: Shallow Coral Reef"
			}
		}

		ebitenutil.DebugPrintAt(screen, depthText, int(telX)+15, int(telY)+28)
		ebitenutil.DebugPrintAt(screen, pressText, int(telX)+15, int(telY)+48)
		ebitenutil.DebugPrintAt(screen, trenchText, int(telX)+15, int(telY)+68)
		ebitenutil.DebugPrintAt(screen, biomeText, int(telX)+15, int(telY)+88)

		if activeVehicle != nil {
			limit := activeVehicle.GetDepthLimit()
			if limit > 0 {
				if depth > limit {
					vector.FillRect(screen, telX+140, telY+25, 90, 18, color.RGBA{210, 55, 75, 200}, false)
					ebitenutil.DebugPrintAt(screen, "CRITICAL!", int(telX)+146, int(telY)+27)
				} else {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Hull Limit: %.0fm", limit), int(telX)+140, int(telY)+28)
				}
			}
		}
	}

	hudX := float32(20) + jx
	hudY := float32(config.ScreenHeight-140) + jy
	const (
		w    = 280
		hBar = 18
	)

	vector.FillRect(screen, hudX, hudY, w, 120, color.RGBA{18, 24, 38, 200}, false)
	vector.StrokeRect(screen, hudX, hudY, w, 120, 1.5, color.RGBA{70, 90, 120, 255}, false)

	hRatio := p.CurrentHealth / p.MaxHealth
	drawStatBar(screen, hudX+15, hudY+15, w-30, hBar, hRatio, color.RGBA{210, 55, 75, 255}, "HP", p.CurrentHealth, p.MaxHealth)

	oRatio := p.CurrentOxygen / p.MaxOxygen
	oColor := color.RGBA{45, 175, 215, 255}
	oLabel := "O2"
	if activeVehicle != nil && activeVehicle.GetOxygen() > 0.0 {
		oRatio = 1.0
		oColor = color.RGBA{45, 215, 175, 255}
		oLabel = "O2 [VEHICLE]"
	}
	drawStatBar(screen, hudX+15, hudY+48, w-30, hBar, oRatio, oColor, oLabel, p.CurrentOxygen, p.MaxOxygen)

	sRatio := p.CurrentStamina / p.MaxStamina
	drawStatBar(screen, hudX+15, hudY+81, w-30, hBar, sRatio, color.RGBA{45, 190, 110, 255}, "ST", p.CurrentStamina, p.MaxStamina)

	if activeVehicle != nil {
		const (
			vHudW = 240
			vHudH = 90
		)
		vHudX := float32(config.ScreenWidth-vHudW-20) + jx
		vHudY := float32(config.ScreenHeight-vHudH-20) + jy

		vector.FillRect(screen, vHudX, vHudY, vHudW, vHudH, color.RGBA{18, 24, 38, 200}, false)
		vector.StrokeRect(screen, vHudX, vHudY, vHudW, vHudH, 1.5, color.RGBA{70, 90, 120, 255}, false)
		ebitenutil.DebugPrintAt(screen, activeVehicle.GetName(), int(vHudX)+15, int(vHudY)+8)

		hullRatio := activeVehicle.GetHealth() / activeVehicle.GetMaxHealth()
		drawStatBar(screen, float32(vHudX+15), float32(vHudY+30), vHudW-30, 14, hullRatio, color.RGBA{220, 80, 50, 255}, "HULL", activeVehicle.GetHealth(), activeVehicle.GetMaxHealth())
		battRatio := activeVehicle.GetBattery() / activeVehicle.GetMaxBattery()
		drawStatBar(screen, float32(vHudX+15), float32(vHudY+54), vHudW-30, 14, battRatio, color.RGBA{220, 180, 40, 255}, "BATT", activeVehicle.GetBattery(), activeVehicle.GetMaxBattery())

		if hl, ok := activeVehicle.(vehicle.HeadlightVehicle); ok && hl.HasHeadlights() {
			btnMinX, btnMinY, btnMaxX, btnMaxY := HUDVehicleLightButtonRect()
			btnW := float32(btnMaxX - btnMinX)
			btnH := float32(btnMaxY - btnMinY)
			btnX := float32(btnMinX) + jx
			btnY := float32(btnMinY) + jy

			isOn := hl.IsHeadlightsOn()
			bgClr := color.RGBA{22, 28, 42, 220}
			borderClr := color.RGBA{70, 90, 120, 255}
			label := "[L] LIGHT: OFF"

			if isOn {
				bgClr = color.RGBA{52, 44, 18, 235}
				borderClr = color.RGBA{255, 215, 60, 255}
				label = "[L] LIGHT: ON"
			}

			vector.FillRect(screen, btnX, btnY, btnW, btnH, bgClr, false)
			vector.StrokeRect(screen, btnX, btnY, btnW, btnH, 1.2, borderClr, false)

			// Small indicator bulb dot
			indicatorClr := color.RGBA{120, 140, 160, 200}
			if isOn {
				indicatorClr = color.RGBA{255, 220, 60, 255}
				vector.FillCircle(screen, btnX+12, btnY+btnH/2.0, 4.5, color.RGBA{255, 240, 160, 80}, true)
			}
			vector.FillCircle(screen, btnX+12, btnY+btnH/2.0, 3.0, indicatorClr, false)

			ebitenutil.DebugPrintAt(screen, label, int(btnX)+22, int(btnY)+6)
		}
	}

	if g.GetCurrentState() == StateCave && weaverTimer > 0 {
		gaugeW := float32(260)
		gaugeH := float32(32)
		gx := float32(config.ScreenWidth)/2.0 - gaugeW/2.0 + jx
		gy := float32(14) + jy

		ratio := float32(math.Min(1.0, math.Max(0.0, weaverTimer/300.0)))
		isCritical := ratio >= 0.85
		flashTick := int(g.GetTimeOfDay()/8)%2 == 0

		// Panel Border and Glow colors
		panelBorderClr := color.RGBA{45, 175, 220, 240}
		barFillClr := color.RGBA{35, 215, 245, 255}
		titleText := "ELECTRICAL SURGE"
		titleClr := color.RGBA{160, 225, 255, 255}

		if isCritical {
			if flashTick {
				panelBorderClr = color.RGBA{255, 45, 55, 255}
				barFillClr = color.RGBA{255, 60, 60, 255}
				titleText = "CRITICAL SURGE IMMINENT!"
				titleClr = color.RGBA{255, 80, 80, 255}
			} else {
				panelBorderClr = color.RGBA{255, 180, 30, 255}
				barFillClr = color.RGBA{255, 190, 40, 255}
				titleText = "CRITICAL SURGE IMMINENT!"
				titleClr = color.RGBA{255, 220, 100, 255}
			}
		} else if ratio >= 0.60 {
			panelBorderClr = color.RGBA{255, 190, 40, 240}
			barFillClr = color.RGBA{255, 190, 45, 255}
			titleText = "ELECTRICAL SURGE LOCKING"
			titleClr = color.RGBA{255, 210, 110, 255}
		}

		// Dark high-tech panel background
		vector.FillRect(screen, gx, gy, gaugeW, gaugeH, color.RGBA{12, 16, 26, 230}, false)
		vector.StrokeRect(screen, gx, gy, gaugeW, gaugeH, 1.4, panelBorderClr, false)

		// Header icon & label
		drawColoredDebugText(screen, titleText, int(gx)+12, int(gy)+4, titleClr)

		// Progress bar track
		barX := gx + 12
		barY := gy + 19
		barW := gaugeW - 24
		barH := float32(7)

		vector.FillRect(screen, barX, barY, barW, barH, color.RGBA{6, 8, 14, 255}, false)
		vector.StrokeRect(screen, barX, barY, barW, barH, 1.0, color.RGBA{35, 45, 65, 200}, false)

		// Progress bar fill
		fillW := barW * ratio
		if fillW > 0 {
			vector.FillRect(screen, barX, barY, fillW, barH, barFillClr, false)
			// Glowing head pip
			vector.FillCircle(screen, barX+fillW, barY+barH/2.0, 2.5, color.White, false)
		}

		// Graduation tick marks (25%, 50%, 75%)
		for _, mark := range []float32{0.25, 0.50, 0.75} {
			tx := barX + barW*mark
			vector.StrokeLine(screen, tx, barY, tx, barY+barH, 1.0, color.RGBA{25, 35, 50, 200}, false)
		}
	}

	if g.IsInventoryOpen() {
		if activeVehicle != nil {
			h.DrawVehicleInventory(screen, g, p.Inventory, activeVehicle.GetCargo(), activeVehicle.GetName())
		} else {
			h.DrawInventory(screen, g, p.Inventory)
		}
	} else {
		h.DrawHUDHotbar(screen, g, p)
	}
}

func drawStatBar(screen *ebiten.Image, x, y, w, h float32, ratio float64, barColor color.Color, label string, val, max float64) {
	vector.FillRect(screen, x, y, w, h, color.RGBA{32, 40, 52, 255}, false)
	fillW := w * float32(ratio)
	if fillW > 0 {
		vector.FillRect(screen, x, y, fillW, h, barColor, false)
	}
	vector.StrokeRect(screen, x, y, w, h, 1.0, color.RGBA{58, 72, 94, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: %.0f/%.0f", label, val, max), int(x)+8, int(y)+(int(h)-14)/2)
}

// DrawHUDHotbar renders the quick-select hotbar on the active HUD.
func (h *HUD) DrawHUDHotbar(screen *ebiten.Image, g GameContext, p *player.Player) {
	h.drawHotbar(screen, p)
}

const (
	HUDVehicleLightBtnW = 125.0
	HUDVehicleLightBtnH = 24.0
)

// HUDVehicleLightButtonRect returns screen-space bounding box for the driving headlight toggle button.
func HUDVehicleLightButtonRect() (minX, minY, maxX, maxY float64) {
	const vHudW = 240.0
	const vHudH = 90.0
	vHudX := float64(config.ScreenWidth - vHudW - 20)
	vHudY := float64(config.ScreenHeight - vHudH - 20)
	btnX := vHudX + vHudW - HUDVehicleLightBtnW
	btnY := vHudY - HUDVehicleLightBtnH - 6.0
	return btnX, btnY, btnX + HUDVehicleLightBtnW, btnY + HUDVehicleLightBtnH
}

// HUDVehicleLightButtonHit returns true if (x, y) coordinates fall within the driving light button.
func HUDVehicleLightButtonHit(x, y float64) bool {
	minX, minY, maxX, maxY := HUDVehicleLightButtonRect()
	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

const (
	HUDHotbarSlotSize = 40.0
	HUDHotbarGap      = 8.0
	HUDHotbarSlots    = 5
	HUDHotbarBottomY  = 56.0
)

// HUDHotbarSlotRect returns the screen-space bounds for a given hotbar slot.
func HUDHotbarSlotRect(slotIdx int) (minX, minY, maxX, maxY float64) {
	w := float64(HUDHotbarSlots*(HUDHotbarSlotSize+HUDHotbarGap) - HUDHotbarGap)
	startX := (float64(config.ScreenWidth) - w) / 2.0
	startY := float64(config.ScreenHeight - HUDHotbarBottomY)

	sx := startX + float64(slotIdx)*(HUDHotbarSlotSize+HUDHotbarGap)
	return sx - HUDHotbarGap/2.0, startY - 12.0, sx + HUDHotbarSlotSize + HUDHotbarGap/2.0, float64(config.ScreenHeight)
}

// HUDHotbarSlotAt returns the hotbar slot index (0..4) at screen coordinates (x, y), or -1 if none.
func HUDHotbarSlotAt(x, y float64) int {
	for i := 0; i < HUDHotbarSlots; i++ {
		minX, minY, maxX, maxY := HUDHotbarSlotRect(i)
		if x >= minX && x <= maxX && y >= minY && y <= maxY {
			return i
		}
	}
	return -1
}

func (h *HUD) drawHotbar(screen *ebiten.Image, p *player.Player) {
	if p.Hotbar == nil {
		return
	}

	w := float32(HUDHotbarSlots*(HUDHotbarSlotSize+HUDHotbarGap) - HUDHotbarGap)
	// Center horizontally at bottom
	x := (float32(config.ScreenWidth) - w) / 2.0
	y := float32(config.ScreenHeight - HUDHotbarBottomY)

	// Draw container background
	vector.FillRect(screen, x-10, y-10, w+20, HUDHotbarSlotSize+20, color.RGBA{18, 24, 38, 200}, false)
	vector.StrokeRect(screen, x-10, y-10, w+20, HUDHotbarSlotSize+20, 1.5, color.RGBA{70, 90, 120, 255}, false)

	for i := 0; i < HUDHotbarSlots; i++ {
		sx := x + float32(i)*(HUDHotbarSlotSize+HUDHotbarGap)
		sy := y

		// Highlight active slot
		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}
		borderWidth := float32(1.0)
		if p.ActiveSlot == i {
			slotBg = color.RGBA{30, 48, 78, 255}
			slotBorder = color.RGBA{0, 230, 255, 255}
			borderWidth = 1.8
		}

		vector.FillRect(screen, sx, sy, HUDHotbarSlotSize, HUDHotbarSlotSize, slotBg, false)
		vector.StrokeRect(screen, sx, sy, HUDHotbarSlotSize, HUDHotbarSlotSize, borderWidth, slotBorder, false)

		if i < len(p.Hotbar.Slots) {
			slot := p.Hotbar.Slots[i]
			if slot.Item != nil {
				// Draw item icon
				drawItemIcon(screen, sx, sy, HUDHotbarSlotSize, slot.Item)
				if slot.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", slot.Quantity), int(sx)+4, int(sy)+int(HUDHotbarSlotSize)-15)
				}
			} else {
				// Virtual Mining Tool (Default outline)
				drawMiningToolOutline(screen, sx, sy, HUDHotbarSlotSize)
			}
		}

		// Draw slot index label
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", i+1), int(sx)+int(HUDHotbarSlotSize)/2-3, int(sy)-22)
	}
}

func drawMiningToolOutline(screen *ebiten.Image, sx, sy, size float32) {
	// A simple pickaxe vector fallback: a brown handle diagonal, and a grey curved head (drawn with two lines).
	cx := sx + size/2.0
	cy := sy + size/2.0
	r := size * 0.28

	// Handle line (diagonal brown line)
	vector.StrokeLine(screen, cx-r, cy+r, cx+r*0.2, cy-r*0.2, 1.5, color.RGBA{130, 100, 70, 128}, false)
	// Pickaxe head left prong (diagonal grey line)
	vector.StrokeLine(screen, cx-r*0.6, cy-r*0.6, cx, cy, 2.0, color.RGBA{180, 190, 200, 128}, false)
	// Pickaxe head right prong (diagonal grey line)
	vector.StrokeLine(screen, cx+r*0.6, cy-r*0.6, cx, cy, 2.0, color.RGBA{180, 190, 200, 128}, false)
}
