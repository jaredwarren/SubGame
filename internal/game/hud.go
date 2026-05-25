package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// HUD renders the health, oxygen, and stamina bars on the screen.
type HUD struct{}

// NewHUD creates a new HUD renderer.
func NewHUD() *HUD {
	return &HUD{}
}

// Draw renders the stats panel for the player and any active vehicle.
func (h *HUD) Draw(screen *ebiten.Image, g *Game) {
	p := g.player
	activeVehicle := g.ActiveVehicle

	// Environment Telemetry Panel (top-left)
	const (
		telX = 20
		telY = 20
		telW = 240
		telH = 95
	)

	// Draw Telemetry background panel (glassmorphism feel)
	vector.DrawFilledRect(screen, telX, telY, telW, telH, color.RGBA{18, 24, 38, 200}, false)
	vector.StrokeRect(screen, telX, telY, telW, telH, 1.5, color.RGBA{70, 90, 120, 255}, false)

	// Pulsing telemetry online indicator
	pulseColor := color.RGBA{45, 215, 120, 255} // Safe green
	if g.currentState == StateCave {
		pulseColor = color.RGBA{45, 175, 215, 255} // Cyan for underwater
	}
	
	// Pulsing effect based on g.TimeOfDay
	isPulseOn := int(g.TimeOfDay/15)%2 == 0
	if !isPulseOn {
		pulseColor.A = 100 // dim it
	}

	vector.DrawFilledCircle(screen, telX+15, telY+15, 4, pulseColor, false)

	if g.currentState == StateOverworld {
		ebitenutil.DebugPrintAt(screen, "SYSTEMS MONITOR", telX+26, telY+8)
		
		// Time calculation
		totalMinutes := int(g.TimeOfDay / 14400.0 * 1440.0)
		hour := (totalMinutes / 60 + 6) % 24
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
		
		isDay := g.TimeOfDay < 7200
		dayPhase := "Day"
		if !isDay {
			dayPhase = "Night"
		}

		timeText := fmt.Sprintf("Time: %02d:%02d %s (%s)", displayHour, minute, period, dayPhase)
		posText := fmt.Sprintf("Pos: X:%.0f Y:%.0f", g.player.X, g.player.Y)
		zoneText := "Zone: Surface Ocean"

		ebitenutil.DebugPrintAt(screen, timeText, telX+15, telY+28)
		ebitenutil.DebugPrintAt(screen, posText, telX+15, telY+48)
		ebitenutil.DebugPrintAt(screen, zoneText, telX+15, telY+68)
	} else if g.currentState == StateCave {
		ebitenutil.DebugPrintAt(screen, "DIVE TELEMETRY", telX+26, telY+8)

		// Depth in meters (1 tile = 1 meter)
		depth := (g.player.Y + g.player.Height/2.0) / TileSize
		pressure := 1.0 + depth*0.1

		depthText := fmt.Sprintf("Depth: %.1fm", depth)
		pressText := fmt.Sprintf("Pressure: %.2f atm", pressure)
		trenchText := fmt.Sprintf("Trench Origin: (%d, %d)", g.activeTrenchX, g.activeTrenchY)

		ebitenutil.DebugPrintAt(screen, depthText, telX+15, telY+28)
		ebitenutil.DebugPrintAt(screen, pressText, telX+15, telY+48)
		ebitenutil.DebugPrintAt(screen, trenchText, telX+15, telY+68)

		// If exceeding depth limit
		if g.ActiveVehicle != nil {
			limit := g.ActiveVehicle.GetDepthLimit()
			if limit > 0 {
				limitText := fmt.Sprintf("Hull Limit: %.0fm", limit)
				if depth > limit {
					// Exceeding limit warning color (red/orange)
					vector.DrawFilledRect(screen, telX+140, telY+25, 90, 18, color.RGBA{210, 55, 75, 200}, false)
					ebitenutil.DebugPrintAt(screen, "CRITICAL!", telX+146, telY+27)
				} else {
					ebitenutil.DebugPrintAt(screen, limitText, telX+140, telY+28)
				}
			}
		}
	}

	// Panel dimensions and placement
	const (
		hudX = 20
		hudY = ScreenHeight - 140
		w    = 280
		hBar = 18
	)

	// Draw HUD background panel (glassmorphism feel)
	panelBg := color.RGBA{18, 24, 38, 200}
	vector.DrawFilledRect(screen, hudX, hudY, w, 120, panelBg, false)
	vector.StrokeRect(screen, hudX, hudY, w, 120, 1.5, color.RGBA{70, 90, 120, 255}, false)

	// Draw Health Bar
	hRatio := p.CurrentHealth / p.MaxHealth
	drawStatBar(screen, hudX+15, hudY+15, w-30, hBar, hRatio, color.RGBA{210, 55, 75, 255}, "HP", p.CurrentHealth, p.MaxHealth)

	// Draw Oxygen Bar (indicate if sustained by vehicle)
	oRatio := p.CurrentOxygen / p.MaxOxygen
	oColor := color.RGBA{45, 175, 215, 255}
	oLabel := "O2"
	if activeVehicle != nil && activeVehicle.GetOxygen() > 0.0 {
		oRatio = 1.0
		oColor = color.RGBA{45, 215, 175, 255} // Teal-ish blue for sub life support
		oLabel = "O2 [VEHICLE]"
	}
	drawStatBar(screen, hudX+15, hudY+48, w-30, hBar, oRatio, oColor, oLabel, p.CurrentOxygen, p.MaxOxygen)

	// Draw Stamina Bar
	sRatio := p.CurrentStamina / p.MaxStamina
	drawStatBar(screen, hudX+15, hudY+81, w-30, hBar, sRatio, color.RGBA{45, 190, 110, 255}, "ST", p.CurrentStamina, p.MaxStamina)

	// Draw active vehicle status at bottom-right if piloting
	if activeVehicle != nil {
		const (
			vHudW = 240
			vHudH = 90
			vHudX = ScreenWidth - vHudW - 20
			vHudY = ScreenHeight - vHudH - 20
		)
		vector.DrawFilledRect(screen, vHudX, vHudY, vHudW, vHudH, color.RGBA{18, 24, 38, 200}, false)
		vector.StrokeRect(screen, vHudX, vHudY, vHudW, vHudH, 1.5, color.RGBA{70, 90, 120, 255}, false)

		ebitenutil.DebugPrintAt(screen, activeVehicle.GetName(), vHudX+15, vHudY+8)

		// Draw Hull integrity bar
		hullRatio := activeVehicle.GetHealth() / activeVehicle.GetMaxHealth()
		drawStatBar(screen, float32(vHudX+15), float32(vHudY+30), vHudW-30, 14, hullRatio, color.RGBA{220, 80, 50, 255}, "HULL", activeVehicle.GetHealth(), activeVehicle.GetMaxHealth())

		// Draw Battery bar
		battRatio := activeVehicle.GetBattery() / activeVehicle.GetMaxBattery()
		drawStatBar(screen, float32(vHudX+15), float32(vHudY+54), vHudW-30, 14, battRatio, color.RGBA{220, 180, 40, 255}, "BATT", activeVehicle.GetBattery(), activeVehicle.GetMaxBattery())
	}
}

func drawStatBar(screen *ebiten.Image, x, y, w, h float32, ratio float64, barColor color.Color, label string, val, max float64) {
	// Bar background
	bg := color.RGBA{32, 40, 52, 255}
	vector.DrawFilledRect(screen, x, y, w, h, bg, false)

	// Bar fill
	fillW := w * float32(ratio)
	if fillW > 0 {
		vector.DrawFilledRect(screen, x, y, fillW, h, barColor, false)
	}

	// Bar border outline
	borderOutline := color.RGBA{58, 72, 94, 255}
	vector.StrokeRect(screen, x, y, w, h, 1.0, borderOutline, false)

	// Text readout on top of the bar using debug printer
	text := fmt.Sprintf("%s: %.0f/%.0f", label, val, max)
	ebitenutil.DebugPrintAt(screen, text, int(x)+8, int(y)+(int(h)-14)/2)
}

// DrawInventory renders the player's grid inventory overlay, icons, counts, and hover tooltips.
func (h *HUD) DrawInventory(screen *ebiten.Image, inv *Inventory) {
	// Center coordinates for the inventory panel overlay
	const (
		panelW = 600
		panelH = 340
		cols   = 8
		rows   = 3
		slotSz = 56
		gap    = 10
	)

	panelX := float32(ScreenWidth-panelW) / 2.0
	panelY := float32(ScreenHeight-panelH) / 2.0

	// Draw panel background panel (translucent glass/steel design)
	panelBg := color.RGBA{14, 20, 32, 238}
	vector.DrawFilledRect(screen, panelX, panelY, panelW, panelH, panelBg, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)

	// Draw Title label
	vector.DrawFilledRect(screen, panelX+15, panelY+12, 110, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " INVENTORY", int(panelX)+20, int(panelY)+16)

	// Get mouse positions to handle slot hover detection
	mx, my := ebiten.CursorPosition()
	var hoveredItemName = "None"

	// Slot grid calculations
	startX := panelX + float32(panelW-float32(cols*(slotSz+gap)-gap))/2.0
	startY := panelY + 60.0

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			slotIdx := r*cols + c
			if slotIdx >= len(inv.Slots) {
				continue
			}

			sx := startX + float32(c*(slotSz+gap))
			sy := startY + float32(r*(slotSz+gap))

			// Slot defaults
			slotBg := color.RGBA{20, 26, 38, 255}
			slotBorder := color.RGBA{48, 60, 80, 255}

			// Hover detection and highlighting
			isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
			if isHovered {
				slotBg = color.RGBA{32, 42, 62, 255}
				slotBorder = color.RGBA{95, 125, 165, 255}
				if inv.Slots[slotIdx].Type != ItemNone {
					hoveredItemName = inv.Slots[slotIdx].Type.String()
				}
			}

			// Render slot container
			vector.DrawFilledRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			// Render item content
			item := inv.Slots[slotIdx]
			if item.Type != ItemNone {
				drawItemIcon(screen, sx, sy, slotSz, item.Type)

				// Draw stack count number
				if item.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.Quantity), int(sx)+6, int(sy)+int(slotSz)-17)
				}
			}
		}
	}

	// Tooltip details bar at bottom
	tooltipY := panelY + panelH - 42
	vector.DrawFilledRect(screen, panelX+20, tooltipY, panelW-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, panelX+20, tooltipY, panelW-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)

	tooltipText := "Hover an item to view details."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

// DrawVehicleInventory renders a split UI showing player inventory on the left and vehicle cargo on the right.
func (h *HUD) DrawVehicleInventory(screen *ebiten.Image, pInv *Inventory, vInv *Inventory, vName string) {
	const (
		panelW = 960
		panelH = 360
		colsP  = 8
		rowsP  = 3
		slotSz = 48
		gap    = 8
	)

	panelX := float32(ScreenWidth-panelW) / 2.0
	panelY := float32(ScreenHeight-panelH) / 2.0

	// Draw panel background
	panelBg := color.RGBA{14, 20, 32, 238}
	vector.DrawFilledRect(screen, panelX, panelY, panelW, panelH, panelBg, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)

	// Titles
	vector.DrawFilledRect(screen, panelX+30, panelY+12, 160, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " DIVER INVENTORY", int(panelX)+35, int(panelY)+16)

	vector.DrawFilledRect(screen, panelX+510, panelY+12, 200, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" %s CARGO", vName), int(panelX)+515, int(panelY)+16)

	mx, my := ebiten.CursorPosition()
	var hoveredItemName = "None"

	// 1. Draw Player Inventory Grid (Left)
	startX_P := panelX + 30
	startY_P := panelY + 60

	for r := 0; r < rowsP; r++ {
		for c := 0; c < colsP; c++ {
			slotIdx := r*colsP + c
			if slotIdx >= len(pInv.Slots) {
				continue
			}

			sx := startX_P + float32(c*(slotSz+gap))
			sy := startY_P + float32(r*(slotSz+gap))

			slotBg := color.RGBA{20, 26, 38, 255}
			slotBorder := color.RGBA{48, 60, 80, 255}

			isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
			if isHovered {
				slotBg = color.RGBA{32, 42, 62, 255}
				slotBorder = color.RGBA{95, 125, 165, 255}
				if pInv.Slots[slotIdx].Type != ItemNone {
					hoveredItemName = pInv.Slots[slotIdx].Type.String()
				}
			}

			vector.DrawFilledRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			item := pInv.Slots[slotIdx]
			if item.Type != ItemNone {
				drawItemIcon(screen, sx, sy, slotSz, item.Type)
				if item.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.Quantity), int(sx)+5, int(sy)+int(slotSz)-16)
				}
			}
		}
	}

	// 2. Draw Vehicle Inventory Grid (Right)
	numSlots := len(vInv.Slots)
	var vCols, vRows int
	if numSlots == 24 {
		vCols = 8
		vRows = 3
	} else if numSlots == 12 {
		vCols = 6
		vRows = 2
	} else { // 8 slots
		vCols = 4
		vRows = 2
	}

	startX_V := panelX + 510
	startY_V := panelY + 60

	for r := 0; r < vRows; r++ {
		for c := 0; c < vCols; c++ {
			slotIdx := r*vCols + c
			if slotIdx >= numSlots {
				continue
			}

			sx := startX_V + float32(c*(slotSz+gap))
			sy := startY_V + float32(r*(slotSz+gap))

			slotBg := color.RGBA{20, 26, 38, 255}
			slotBorder := color.RGBA{48, 60, 80, 255}

			isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
			if isHovered {
				slotBg = color.RGBA{32, 42, 62, 255}
				slotBorder = color.RGBA{95, 125, 165, 255}
				if vInv.Slots[slotIdx].Type != ItemNone {
					hoveredItemName = vInv.Slots[slotIdx].Type.String()
				}
			}

			vector.DrawFilledRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			item := vInv.Slots[slotIdx]
			if item.Type != ItemNone {
				drawItemIcon(screen, sx, sy, slotSz, item.Type)
				if item.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.Quantity), int(sx)+5, int(sy)+int(slotSz)-16)
				}
			}
		}
	}

	// Tooltip
	tooltipY := panelY + panelH - 42
	vector.DrawFilledRect(screen, panelX+20, tooltipY, panelW-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, panelX+20, tooltipY, panelW-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)

	tooltipText := "Hover an item to view details. Click item to transfer."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

// drawItemIcon helper renders customized vector icons for each item type.
func drawItemIcon(screen *ebiten.Image, sx, sy, slotSz float32, itemType ItemType) {
	var itemClr color.Color
	var drawIconType = "circle"

	switch itemType {
	case ItemTitanium:
		itemClr = color.RGBA{168, 178, 188, 255}
		drawIconType = "square"
	case ItemCopper:
		itemClr = color.RGBA{218, 118, 48, 255}
		drawIconType = "square"
	case ItemQuartz:
		itemClr = color.RGBA{48, 218, 245, 255}
		drawIconType = "diamond"
	case ItemAbyssalOre:
		itemClr = color.RGBA{148, 48, 218, 255}
		drawIconType = "diamond"
	case ItemScoutSub:
		itemClr = color.RGBA{15, 160, 185, 255}
		drawIconType = "sub"
	case ItemHeavyMech:
		itemClr = color.RGBA{218, 98, 16, 255}
		drawIconType = "mech"
	default:
		itemClr = color.RGBA{98, 198, 148, 255} // Upgrades/Tools
		drawIconType = "circle"
	}

	cx := sx + slotSz/2.0
	cy := sy + slotSz/2.0
	iSz := slotSz * 0.45

	if drawIconType == "square" {
		vector.DrawFilledRect(screen, cx-iSz/2.0, cy-iSz/2.0, iSz, iSz, itemClr, false)
	} else if drawIconType == "diamond" {
		vector.DrawFilledCircle(screen, cx, cy, iSz/2.0, itemClr, false)
		vector.StrokeCircle(screen, cx, cy, iSz/2.0, 1.0, color.RGBA{255, 255, 255, 200}, false)
	} else if drawIconType == "sub" {
		// Draw a small sub capsule silhouette
		vector.DrawFilledRect(screen, cx-iSz/2.0, cy-iSz/4.0, iSz, iSz/2.0, itemClr, false)
		vector.DrawFilledCircle(screen, cx+iSz/4.0, cy, iSz/4.0, color.RGBA{80, 205, 255, 255}, false)
	} else if drawIconType == "mech" {
		// Draw a tiny mech torso silhouette
		vector.DrawFilledRect(screen, cx-iSz/3.0, cy-iSz/3.0, iSz/1.5, iSz/1.5, itemClr, false)
		vector.DrawFilledRect(screen, cx-iSz/2.0, cy+iSz/6.0, iSz, iSz/6.0, color.RGBA{60, 70, 80, 255}, false)
	} else {
		vector.DrawFilledCircle(screen, cx, cy, iSz/2.0, itemClr, false)
	}
}
