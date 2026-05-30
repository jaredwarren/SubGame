package scene

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
	telH := float32(95)
	if g.GetCurrentState() == StateOverworld {
		telH = 115
	}

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
		isDay := g.GetTimeOfDay() < 7200
		dayPhase := "Day"
		if !isDay {
			dayPhase = "Night"
		}
		timeText := fmt.Sprintf("Time: %02d:%02d %s (%s)", displayHour, minute, period, dayPhase)

		w := g.GetWorld()
		tx := int(p.Pos.X+p.Width/2) / config.TileSize
		ty := int(p.Pos.Y+p.Height/2) / config.TileSize
		outOfBounds := tx < 0 || tx >= w.Width || ty < 0 || ty >= w.Height

		var posText, zoneText, depthText string
		if outOfBounds {
			posText = "Pos: X:??? Y:???"
			zoneText = "Zone: Ecological Void"
			depthText = "Est. Dive Depth: ???"
		} else {
			posText = fmt.Sprintf("Pos: X:%.0f Y:%.0f", p.Pos.X, p.Pos.Y)
			zoneText = "Zone: Surface Ocean"
			if w.OverworldMap[tx][ty] == world.TileTrench {
				depthText = "Est. Dive Depth: Trench (120m)"
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
		ebitenutil.DebugPrintAt(screen, zoneText, int(telX)+15, int(telY)+68)
		ebitenutil.DebugPrintAt(screen, depthText, int(telX)+15, int(telY)+88)

	} else if g.GetCurrentState() == StateCave {
		ebitenutil.DebugPrintAt(screen, "DIVE TELEMETRY", int(telX)+26, int(telY)+8)

		depth := (p.Pos.Y + p.Height/2.0) / config.TileSize
		pressure := 1.0 + depth*0.1
		depthText := fmt.Sprintf("Depth: %.1fm", depth)
		pressText := fmt.Sprintf("Pressure: %.2f atm", pressure)

		trenchX, trenchY := g.GetActiveTrenchCoords()
		var trenchText string
		activeCave := g.GetActiveCave()
		if activeCave != nil && activeCave.GetCaveType() == cave.CaveVoid {
			trenchText = "Trench Origin: ???"
		} else {
			trenchText = fmt.Sprintf("Trench Origin: (%d, %d)", trenchX, trenchY)
		}

		ebitenutil.DebugPrintAt(screen, depthText, int(telX)+15, int(telY)+28)
		ebitenutil.DebugPrintAt(screen, pressText, int(telX)+15, int(telY)+48)
		ebitenutil.DebugPrintAt(screen, trenchText, int(telX)+15, int(telY)+68)

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
	}

	if g.GetCurrentState() == StateCave && weaverTimer > 0 {
		vector.FillRect(screen, float32(config.ScreenWidth)/2.0-180+jx, 15+jy, 360, 24, color.RGBA{24, 12, 10, 220}, false)
		vector.StrokeRect(screen, float32(config.ScreenWidth)/2.0-180+jx, 15+jy, 360, 24, 1.2, color.RGBA{230, 75, 45, 255}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[WARNING: ELECTRICAL STATIC DETECTED (%.0f%%)]", (weaverTimer/300.0)*100.0), config.ScreenWidth/2-160+int(jx), 20+int(jy))
	}

	if g.IsInventoryOpen() {
		if activeVehicle != nil {
			h.DrawVehicleInventory(screen, g, p.Inventory, activeVehicle.GetCargo(), activeVehicle.GetName())
		} else {
			h.DrawInventory(screen, g, p.Inventory)
		}
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

// DrawInventory renders the player's grid inventory overlay.
func (h *HUD) DrawInventory(screen *ebiten.Image, g GameContext, inv *item.Inventory) {
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

	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{14, 20, 32, 238}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)
	vector.FillRect(screen, panelX+15, panelY+12, 110, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " INVENTORY", int(panelX)+20, int(panelY)+16)

	cursor := g.GetInput().Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
	var hoveredItemName = "None"

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

			slotBg := color.RGBA{20, 26, 38, 255}
			slotBorder := color.RGBA{48, 60, 80, 255}
			isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
			if isHovered {
				slotBg = color.RGBA{32, 42, 62, 255}
				slotBorder = color.RGBA{95, 125, 165, 255}
				if inv.Slots[slotIdx].Item != nil {
					hoveredItemName = inv.Slots[slotIdx].Item.GetName()
				}
			}

			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			slot := inv.Slots[slotIdx]
			if slot.Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, slot.Item)
				if slot.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", slot.Quantity), int(sx)+6, int(sy)+int(slotSz)-17)
				}
			}
		}
	}

	p := g.GetPlayer()
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
			if p.Upgrades != nil && c < len(p.Upgrades.Slots) && p.Upgrades.Slots[c].Item != nil {
				hoveredItemName = p.Upgrades.Slots[c].Item.GetName()
			}
		}

		vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
		vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

		if p.Upgrades != nil && c < len(p.Upgrades.Slots) && p.Upgrades.Slots[c].Item != nil {
			drawItemIcon(screen, sx, sy, slotSz, p.Upgrades.Slots[c].Item)
		}
	}

	tooltipY := panelY + panelH - 42
	vector.FillRect(screen, panelX+20, tooltipY, panelW-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, panelX+20, tooltipY, panelW-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)
	tooltipText := "Hover an item to view details."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

// DrawVehicleInventory renders a split UI showing player inventory and vehicle cargo.
func (h *HUD) DrawVehicleInventory(screen *ebiten.Image, g GameContext, pInv *item.Inventory, vInv *item.Inventory, vName string) {
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

	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{14, 20, 32, 238}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)
	vector.FillRect(screen, panelX+30, panelY+12, 160, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " DIVER INVENTORY", int(panelX)+35, int(panelY)+16)
	vector.FillRect(screen, panelX+510, panelY+12, 200, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" %s CARGO", vName), int(panelX)+515, int(panelY)+16)

	cursor := g.GetInput().Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
	var hoveredItemName = "None"

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
			slot := pInv.Slots[slotIdx]
			if slot.Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, slot.Item)
				if slot.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", slot.Quantity), int(sx)+5, int(sy)+int(slotSz)-16)
				}
			}
		}
	}

	numSlots := len(vInv.Slots)
	var vCols, vRows int
	if numSlots == 24 {
		vCols, vRows = 8, 3
	} else if numSlots == 12 {
		vCols, vRows = 6, 2
	} else {
		vCols, vRows = 4, 2
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
			slot := vInv.Slots[slotIdx]
			if slot.Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, slot.Item)
				if slot.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", slot.Quantity), int(sx)+5, int(sy)+int(slotSz)-16)
				}
			}
		}
	}

	activeVehicle := g.GetActiveVehicle()
	vUpg := activeVehicle.GetUpgrades()
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
				if c < len(vUpg.Slots) && vUpg.Slots[c].Item != nil {
					hoveredItemName = vUpg.Slots[c].Item.GetName()
				}
			}
			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)
			if c < len(vUpg.Slots) && vUpg.Slots[c].Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, vUpg.Slots[c].Item)
			}
		}
	}

	tooltipY := panelY + panelH - 42
	vector.FillRect(screen, panelX+20, tooltipY, panelW-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, panelX+20, tooltipY, panelW-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)
	tooltipText := "Hover an item to view details. Click item to transfer."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

func drawItemIcon(screen *ebiten.Image, sx, sy, slotSz float32, i item.Item) {
	if i == nil {
		return
	}
	i.DrawIcon(screen, sx+slotSz/2.0, sy+slotSz/2.0, slotSz*0.45)
}
