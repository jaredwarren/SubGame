package game

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/world"
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

	var jx, jy float32
	if g.currentState == StateCave && g.WeaverTrackingTimer > 0 {
		mag := float32((g.WeaverTrackingTimer / 300.0) * 5.0)
		jx = rand.Float32()*mag - mag/2.0
		jy = rand.Float32()*mag - mag/2.0
	}

	// Environment Telemetry Panel (top-left)
	telX := float32(20) + jx
	telY := float32(20) + jy
	telW := float32(240)
	telH := float32(95)
	if g.currentState == StateOverworld {
		telH = 115
	}

	// Draw Telemetry background panel (glassmorphism feel)
	vector.FillRect(screen, telX, telY, telW, telH, color.RGBA{18, 24, 38, 200}, false)
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

	vector.FillCircle(screen, telX+15, telY+15, 4, pulseColor, false)

	if g.currentState == StateOverworld {
		ebitenutil.DebugPrintAt(screen, "SYSTEMS MONITOR", int(telX)+26, int(telY)+8)

		// Time calculation
		totalMinutes := int(g.TimeOfDay / 14400.0 * 1440.0)
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

		isDay := g.TimeOfDay < 7200
		dayPhase := "Day"
		if !isDay {
			dayPhase = "Night"
		}

		timeText := fmt.Sprintf("Time: %02d:%02d %s (%s)", displayHour, minute, period, dayPhase)

		tx := int(g.player.Pos.X+g.player.Width/2) / config.TileSize
		ty := int(g.player.Pos.Y+g.player.Height/2) / config.TileSize
		outOfBounds := tx < 0 || tx >= g.world.Width || ty < 0 || ty >= g.world.Height

		var posText, zoneText, depthText string
		if outOfBounds {
			posText = "Pos: X:??? Y:???"
			zoneText = "Zone: Ecological Void"
			depthText = "Est. Dive Depth: ???"
		} else {
			posText = fmt.Sprintf("Pos: X:%.0f Y:%.0f", g.player.Pos.X, g.player.Pos.Y)
			zoneText = "Zone: Surface Ocean"
			if g.world.OverworldMap[tx][ty] == world.TileTrench {
				depthText = "Est. Dive Depth: Trench (120m)"
			} else {
				dist := g.world.DistanceToLand(tx, ty)
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
		ebitenutil.DebugPrintAt(screen, zoneText, int(telX)+15, int(telY)+68)
		ebitenutil.DebugPrintAt(screen, depthText, int(telX)+15, int(telY)+88)
	} else if g.currentState == StateCave {
		ebitenutil.DebugPrintAt(screen, "DIVE TELEMETRY", int(telX)+26, int(telY)+8)

		// Depth in meters (1 tile = 1 meter)
		depth := (g.player.Pos.Y + g.player.Height/2.0) / config.TileSize

		var depthText, pressText, trenchText string
		pressure := 1.0 + depth*0.1

		depthText = fmt.Sprintf("Depth: %.1fm", depth)
		pressText = fmt.Sprintf("Pressure: %.2f atm", pressure)

		if g.caveState.ActiveCave != nil && g.caveState.ActiveCave.GetCaveType() == cave.CaveVoid {
			trenchText = "Trench Origin: ???"
		} else {
			trenchText = fmt.Sprintf("Trench Origin: (%d, %d)", g.activeTrenchX, g.activeTrenchY)
		}

		ebitenutil.DebugPrintAt(screen, depthText, int(telX)+15, int(telY)+28)
		ebitenutil.DebugPrintAt(screen, pressText, int(telX)+15, int(telY)+48)
		ebitenutil.DebugPrintAt(screen, trenchText, int(telX)+15, int(telY)+68)

		// If exceeding depth limit
		if g.ActiveVehicle != nil {
			limit := g.ActiveVehicle.GetDepthLimit()
			if limit > 0 {
				limitText := fmt.Sprintf("Hull Limit: %.0fm", limit)
				if depth > limit {
					// Exceeding limit warning color (red/orange)
					vector.FillRect(screen, telX+140, telY+25, 90, 18, color.RGBA{210, 55, 75, 200}, false)
					ebitenutil.DebugPrintAt(screen, "CRITICAL!", int(telX)+146, int(telY)+27)
				} else {
					ebitenutil.DebugPrintAt(screen, limitText, int(telX)+140, int(telY)+28)
				}
			}
		}
	}

	// Panel dimensions and placement
	hudX := float32(20) + jx
	hudY := float32(config.ScreenHeight-140) + jy
	const (
		w    = 280
		hBar = 18
	)

	// Draw HUD background panel (glassmorphism feel)
	panelBg := color.RGBA{18, 24, 38, 200}
	vector.FillRect(screen, hudX, hudY, w, 120, panelBg, false)
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
		)
		vHudX := float32(config.ScreenWidth-vHudW-20) + jx
		vHudY := float32(config.ScreenHeight-vHudH-20) + jy

		vector.FillRect(screen, vHudX, vHudY, vHudW, vHudH, color.RGBA{18, 24, 38, 200}, false)
		vector.StrokeRect(screen, vHudX, vHudY, vHudW, vHudH, 1.5, color.RGBA{70, 90, 120, 255}, false)

		ebitenutil.DebugPrintAt(screen, activeVehicle.GetName(), int(vHudX)+15, int(vHudY)+8)

		// Draw Hull integrity bar
		hullRatio := activeVehicle.GetHealth() / activeVehicle.GetMaxHealth()
		drawStatBar(screen, float32(vHudX+15), float32(vHudY+30), vHudW-30, 14, hullRatio, color.RGBA{220, 80, 50, 255}, "HULL", activeVehicle.GetHealth(), activeVehicle.GetMaxHealth())

		// Draw Battery bar
		battRatio := activeVehicle.GetBattery() / activeVehicle.GetMaxBattery()
		drawStatBar(screen, float32(vHudX+15), float32(vHudY+54), vHudW-30, 14, battRatio, color.RGBA{220, 180, 40, 255}, "BATT", activeVehicle.GetBattery(), activeVehicle.GetMaxBattery())
	}

	// Display warning banner at top center if tracked by Weaver
	if g.currentState == StateCave && g.WeaverTrackingTimer > 0 {
		vector.FillRect(screen, float32(config.ScreenWidth)/2.0-180+jx, 15+jy, 360, 24, color.RGBA{24, 12, 10, 220}, false)
		vector.StrokeRect(screen, float32(config.ScreenWidth)/2.0-180+jx, 15+jy, 360, 24, 1.2, color.RGBA{230, 75, 45, 255}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[WARNING: ELECTRICAL STATIC DETECTED (%.0f%%)]", (g.WeaverTrackingTimer/300.0)*100.0), config.ScreenWidth/2-160+int(jx), 20+int(jy))
	}

	// Render inventory overlay
	if g.showInventory {
		if g.ActiveVehicle != nil {
			g.hud.DrawVehicleInventory(screen, g, g.player.Inventory, g.ActiveVehicle.GetCargo(), g.ActiveVehicle.GetName())
		} else {
			g.hud.DrawInventory(screen, g, g.player.Inventory)
		}
	}
}

func drawStatBar(screen *ebiten.Image, x, y, w, h float32, ratio float64, barColor color.Color, label string, val, max float64) {
	// Bar background
	bg := color.RGBA{32, 40, 52, 255}
	vector.FillRect(screen, x, y, w, h, bg, false)

	// Bar fill
	fillW := w * float32(ratio)
	if fillW > 0 {
		vector.FillRect(screen, x, y, fillW, h, barColor, false)
	}

	// Bar border outline
	borderOutline := color.RGBA{58, 72, 94, 255}
	vector.StrokeRect(screen, x, y, w, h, 1.0, borderOutline, false)

	// Text readout on top of the bar using debug printer
	text := fmt.Sprintf("%s: %.0f/%.0f", label, val, max)
	ebitenutil.DebugPrintAt(screen, text, int(x)+8, int(y)+(int(h)-14)/2)
}

// DrawInventory renders the player's grid inventory overlay, icons, counts, and hover tooltips.
func (h *HUD) DrawInventory(screen *ebiten.Image, g *Game, inv *item.Inventory) {
	// Center coordinates for the inventory panel overlay
	const (
		panelW = 600
		panelH = 420
		cols   = 8
		rows   = 3
		slotSz = 56
		gap    = 10
	)

	panelX := float32(config.ScreenWidth-panelW) / 2.0
	panelY := float32(config.ScreenHeight-panelH) / 2.0

	// Draw panel background panel (translucent glass/steel design)
	panelBg := color.RGBA{14, 20, 32, 238}
	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelBg, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)

	// Draw Title label
	vector.FillRect(screen, panelX+15, panelY+12, 110, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " INVENTORY", int(panelX)+20, int(panelY)+16)

	// Get mouse positions to handle slot hover detection
	cursor := g.Input.Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
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
				if inv.Slots[slotIdx].Item != nil {
					hoveredItemName = inv.Slots[slotIdx].Item.GetName()
				}
			}

			// Render slot container
			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			// Render item content
			item := inv.Slots[slotIdx]
			if item.Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, item.Item)

				// Draw stack count number
				if item.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.Quantity), int(sx)+6, int(sy)+int(slotSz)-17)
				}
			}
		}
	}

	// Draw Equipped Gear slots section
	gearY := startY + float32(rows*(slotSz+gap)) + 5.0
	ebitenutil.DebugPrintAt(screen, "EQUIPPED GEAR (CLICK ITEM TO EQUIP / UNEQUIP)", int(panelX)+20, int(gearY))

	gearStartX := panelX + (panelW-(4.0*float32(slotSz)+3.0*float32(gap)))/2.0
	gearSlotsY := gearY + 22.0

	for c := 0; c < 4; c++ {
		sx := gearStartX + float32(c*(slotSz+gap))
		sy := gearSlotsY

		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}

		isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
		if isHovered {
			slotBg = color.RGBA{32, 42, 62, 255}
			slotBorder = color.RGBA{95, 125, 165, 255}
			if g.player.Upgrades != nil && c < len(g.player.Upgrades.Slots) {
				itemStack := g.player.Upgrades.Slots[c]
				if itemStack.Item != nil {
					hoveredItemName = itemStack.Item.GetName()
				}
			}
		}

		vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
		vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

		if g.player.Upgrades != nil && c < len(g.player.Upgrades.Slots) {
			itemStack := g.player.Upgrades.Slots[c]
			if itemStack.Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, itemStack.Item)
			}
		}
	}

	// Tooltip details bar at bottom
	tooltipY := panelY + panelH - 42
	vector.FillRect(screen, panelX+20, tooltipY, panelW-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, panelX+20, tooltipY, panelW-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)

	tooltipText := "Hover an item to view details."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

// DrawVehicleInventory renders a split UI showing player inventory on the left and vehicle cargo on the right.
func (h *HUD) DrawVehicleInventory(screen *ebiten.Image, g *Game, pInv *item.Inventory, vInv *item.Inventory, vName string) {
	const (
		panelW = 960
		panelH = 360
		colsP  = 8
		rowsP  = 3
		slotSz = 48
		gap    = 8
	)

	panelX := float32(config.ScreenWidth-panelW) / 2.0
	panelY := float32(config.ScreenHeight-panelH) / 2.0

	// Draw panel background
	panelBg := color.RGBA{14, 20, 32, 238}
	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelBg, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)

	// Titles
	vector.FillRect(screen, panelX+30, panelY+12, 160, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " DIVER INVENTORY", int(panelX)+35, int(panelY)+16)

	vector.FillRect(screen, panelX+510, panelY+12, 200, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" %s CARGO", vName), int(panelX)+515, int(panelY)+16)

	cursor := g.Input.Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
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
				if pInv.Slots[slotIdx].Item != nil {
					hoveredItemName = pInv.Slots[slotIdx].Item.GetName()
				}
			}

			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			item := pInv.Slots[slotIdx]
			if item.Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, item.Item)
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
				if vInv.Slots[slotIdx].Item != nil {
					hoveredItemName = vInv.Slots[slotIdx].Item.GetName()
				}
			}

			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			item := vInv.Slots[slotIdx]
			if item.Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, item.Item)
				if item.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.Quantity), int(sx)+5, int(sy)+int(slotSz)-16)
				}
			}
		}
	}

	// 3. Draw Vehicle Upgrades Grid (Below Cargo)
	vUpg := g.ActiveVehicle.GetUpgrades()
	if vUpg != nil {
		upgY := panelY + 220
		vector.FillRect(screen, startX_V, upgY, 200, 18, color.RGBA{22, 32, 50, 255}, false)
		ebitenutil.DebugPrintAt(screen, " VEHICLE UPGRADES", int(startX_V), int(upgY)+2)

		upgSlotsY := upgY + 24
		for c := 0; c < len(vUpg.Slots); c++ {
			sx := startX_V + float32(c*(slotSz+gap))
			sy := upgSlotsY

			slotBg := color.RGBA{20, 26, 38, 255}
			slotBorder := color.RGBA{48, 60, 80, 255}

			isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
			if isHovered {
				slotBg = color.RGBA{32, 42, 62, 255}
				slotBorder = color.RGBA{95, 125, 165, 255}
				if c < len(vUpg.Slots) {
					itemStack := vUpg.Slots[c]
					if itemStack.Item != nil {
						hoveredItemName = itemStack.Item.GetName()
					}
				}
			}

			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			if c < len(vUpg.Slots) {
				itemStack := vUpg.Slots[c]
				if itemStack.Item != nil {
					drawItemIcon(screen, sx, sy, slotSz, itemStack.Item)
				}
			}
		}
	}

	// Tooltip
	tooltipY := panelY + panelH - 42
	vector.FillRect(screen, panelX+20, tooltipY, panelW-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, panelX+20, tooltipY, panelW-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)

	tooltipText := "Hover an item to view details. Click item to transfer."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

// drawItemIcon helper renders the item's icon in the given inventory slot.
func drawItemIcon(screen *ebiten.Image, sx, sy, slotSz float32, i item.Item) {
	if i == nil {
		return
	}
	cx := sx + slotSz/2.0
	cy := sy + slotSz/2.0
	iSz := slotSz * 0.45
	i.DrawIcon(screen, cx, cy, iSz)
}
