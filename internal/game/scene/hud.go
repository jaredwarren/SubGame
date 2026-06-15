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

		var posText, zoneText, depthText string
		if outOfBounds {
			posText = "Pos: X:??? Y:???"
			zoneText = "Zone: Ecological Void"
			depthText = "Est. Dive Depth: ???"
		} else {
			posText = fmt.Sprintf("Pos: X:%.0f Y:%.0f", p.Pos.X, p.Pos.Y)
			zoneText = "Zone: Surface Ocean"

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
	layout := SoloInventoryGridDescriptor
	panelX := (float64(config.ScreenWidth) - layout.PanelW) / 2.0
	panelY := (float64(config.ScreenHeight) - layout.PanelH) / 2.0

	vector.FillRect(screen, float32(panelX), float32(panelY), float32(layout.PanelW), float32(layout.PanelH), color.RGBA{14, 20, 32, 238}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(layout.PanelW), float32(layout.PanelH), 1.5, color.RGBA{68, 88, 120, 255}, false)
	vector.FillRect(screen, float32(panelX)+15, float32(panelY)+12, 110, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " INVENTORY", int(panelX)+20, int(panelY)+16)

	cursor := g.GetInput().Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
	var hoveredItemName = "None"

	for slotIdx := 0; slotIdx < len(inv.Slots); slotIdx++ {
		rect := layout.SlotRect(panelX, panelY, slotIdx)
		sx := float32(rect.Min.X)
		sy := float32(rect.Min.Y)
		slotSz := float32(layout.SlotSz)

		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}
		isHovered := mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
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

	p := g.GetPlayer()
	gearLayout := SoloGearGridDescriptor
	gearY := float32(panelY + layout.StartY + float64(layout.Rows)*(layout.SlotSz+layout.Gap) - layout.Gap + 5.0)
	ebitenutil.DebugPrintAt(screen, "EQUIPPED GEAR (CLICK ITEM TO EQUIP / UNEQUIP)", int(panelX)+20, int(gearY))

	for c := 0; c < 4; c++ {
		rect := gearLayout.SlotRect(panelX, panelY, c)
		sx := float32(rect.Min.X)
		sy := float32(rect.Min.Y)
		slotSz := float32(gearLayout.SlotSz)

		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}
		isHovered := mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
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

	tooltipY := float32(panelY + layout.PanelH - 42)
	vector.FillRect(screen, float32(panelX)+20, tooltipY, float32(layout.PanelW)-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, float32(panelX)+20, tooltipY, float32(layout.PanelW)-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)
	tooltipText := "Hover an item to view details."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

// DrawVehicleInventory renders a split UI showing player inventory and vehicle cargo.
func (h *HUD) DrawVehicleInventory(screen *ebiten.Image, g GameContext, pInv *item.Inventory, vInv *item.Inventory, vName string) {
	layoutP := VehiclePlayerInvLayout
	panelX := (float64(config.ScreenWidth) - layoutP.PanelW) / 2.0
	panelY := (float64(config.ScreenHeight) - layoutP.PanelH) / 2.0

	vector.FillRect(screen, float32(panelX), float32(panelY), float32(layoutP.PanelW), float32(layoutP.PanelH), color.RGBA{14, 20, 32, 238}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(layoutP.PanelW), float32(layoutP.PanelH), 1.5, color.RGBA{68, 88, 120, 255}, false)
	vector.FillRect(screen, float32(panelX)+30, float32(panelY)+12, 160, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, " DIVER INVENTORY", int(panelX)+35, int(panelY)+16)
	vector.FillRect(screen, float32(panelX)+510, float32(panelY)+12, 200, 24, color.RGBA{22, 32, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(" %s CARGO", vName), int(panelX)+515, int(panelY)+16)

	cursor := g.GetInput().Cursor()
	mx, my := int(cursor.X), int(cursor.Y)

	activeVehicle := g.GetActiveVehicle()
	if activeVehicle != nil && activeVehicle.GetKit() != nil {
		btnX := panelX + 730
		btnY := panelY + 12
		const btnW, btnH = 200.0, 24.0
		btnBg := color.RGBA{22, 50, 80, 255}
		btnBorder := color.RGBA{45, 175, 215, 255}
		if mx >= int(btnX) && mx < int(btnX+btnW) && my >= int(btnY) && my < int(btnY+btnH) {
			btnBg = color.RGBA{30, 80, 120, 255}
			btnBorder = color.RGBA{60, 200, 255, 255}
		}
		vector.FillRect(screen, float32(btnX), float32(btnY), btnW, btnH, btnBg, false)
		vector.StrokeRect(screen, float32(btnX), float32(btnY), btnW, btnH, 1.0, btnBorder, false)
		ebitenutil.DebugPrintAt(screen, "  PICK UP VEHICLE", int(btnX)+10, int(btnY)+4)
	}

	var hoveredItemName = "None"

	// Player Inventory Grid
	for slotIdx := 0; slotIdx < len(pInv.Slots); slotIdx++ {
		rect := layoutP.SlotRect(panelX, panelY, slotIdx)
		sx := float32(rect.Min.X)
		sy := float32(rect.Min.Y)
		slotSz := float32(layoutP.SlotSz)

		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}
		isHovered := mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
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

	// Vehicle Cargo Grid
	layoutV := GetVehicleCargoLayout(len(vInv.Slots))
	for slotIdx := 0; slotIdx < len(vInv.Slots); slotIdx++ {
		rect := layoutV.SlotRect(panelX, panelY, slotIdx)
		sx := float32(rect.Min.X)
		sy := float32(rect.Min.Y)
		slotSz := float32(layoutV.SlotSz)

		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}
		isHovered := mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
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

	// Vehicle Upgrades Grid
	vUpg := activeVehicle.GetUpgrades()
	if vUpg != nil {
		startX_V := float32(panelX + 510)
		upgY := float32(panelY + 220)
		vector.FillRect(screen, startX_V, upgY, 200, 18, color.RGBA{22, 32, 50, 255}, false)
		ebitenutil.DebugPrintAt(screen, " VEHICLE UPGRADES", int(startX_V), int(upgY)+2)

		layoutUpg := GetVehicleUpgradeLayout(len(vUpg.Slots))
		for c := 0; c < len(vUpg.Slots); c++ {
			rect := layoutUpg.SlotRect(panelX, panelY, c)
			sx := float32(rect.Min.X)
			sy := float32(rect.Min.Y)
			slotSz := float32(layoutUpg.SlotSz)

			slotBg := color.RGBA{20, 26, 38, 255}
			slotBorder := color.RGBA{48, 60, 80, 255}
			isHovered := mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
			if isHovered {
				slotBg = color.RGBA{32, 42, 62, 255}
				slotBorder = color.RGBA{95, 125, 165, 255}
				if vUpg.Slots[c].Item != nil {
					hoveredItemName = vUpg.Slots[c].Item.GetName()
				}
			}
			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)
			if vUpg.Slots[c].Item != nil {
				drawItemIcon(screen, sx, sy, slotSz, vUpg.Slots[c].Item)
			}
		}
	}

	tooltipY := float32(panelY + 360 - 42)
	vector.FillRect(screen, float32(panelX)+20, tooltipY, float32(layoutP.PanelW)-40, 24, color.RGBA{8, 12, 18, 255}, false)
	vector.StrokeRect(screen, float32(panelX)+20, tooltipY, float32(layoutP.PanelW)-40, 24, 0.8, color.RGBA{40, 52, 70, 255}, false)
	tooltipText := "Hover an item to view details. Click item to transfer."
	if hoveredItemName != "None" {
		tooltipText = hoveredItemName
	}
	ebitenutil.DebugPrintAt(screen, tooltipText, int(panelX)+30, int(tooltipY)+4)
}

// HandlePlayerInventoryClicks processes clicks on the player inventory and gear grid.
func (h *HUD) HandlePlayerInventoryClicks(g GameContext) {
	cursor := g.GetInput().Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
	p := g.GetPlayer()

	panelX := float64(config.ScreenWidth-600) / 2.0
	panelY := float64(config.ScreenHeight-420) / 2.0

	// Main inventory grid
	hoveredIdx := SoloInventoryGridDescriptor.HoveredSlot(panelX, panelY, len(p.Inventory.Slots), mx, my)
	if hoveredIdx != -1 {
		slot := &p.Inventory.Slots[hoveredIdx]
		if slot.Item != nil {
			g.ActivatePlayerItem(slot.Item)
		}
		return
	}

	// Equipped gear slots (uninstall on click)
	if p.Upgrades != nil {
		hoveredGearIdx := SoloGearGridDescriptor.HoveredSlot(panelX, panelY, len(p.Upgrades.Slots), mx, my)
		if hoveredGearIdx != -1 {
			slot := &p.Upgrades.Slots[hoveredGearIdx]
			if slot.Item != nil && p.Inventory.AddItem(item.Clone(slot.Item), 1) {
				p.Upgrades.Remove(slot.Item, 1)
				p.RecalculateUpgrades()
			}
		}
	}
}

// HandleVehicleInventoryClicks processes clicks on the player + vehicle inventory split panel.
func (h *HUD) HandleVehicleInventoryClicks(g GameContext) {
	cursor := g.GetInput().Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
	p := g.GetPlayer()
	v := g.GetActiveVehicle()
	if v == nil {
		return
	}

	panelX := float64(config.ScreenWidth-960) / 2.0
	panelY := float64(config.ScreenHeight-360) / 2.0

	// Check if Pick Up button is clicked
	if v.GetKit() != nil {
		btnX := panelX + 730
		btnY := panelY + 12
		const btnW, btnH = 200.0, 24.0
		if float64(mx) >= btnX && float64(mx) < btnX+btnW && float64(my) >= btnY && float64(my) < btnY+btnH {
			g.PickUpActiveVehicle()
			return
		}
	}

	// Transfer from player inventory to vehicle
	hoveredPlayerIdx := VehiclePlayerInvLayout.HoveredSlot(panelX, panelY, len(p.Inventory.Slots), mx, my)
	if hoveredPlayerIdx != -1 {
		slot := &p.Inventory.Slots[hoveredPlayerIdx]
		if slot.Item != nil {
			g.TransferToVehicle(slot.Item)
		}
		return
	}

	// Transfer from vehicle cargo to player
	vInv := v.GetCargo()
	cargoLayoutDesc := GetVehicleCargoLayout(len(vInv.Slots))
	hoveredCargoIdx := cargoLayoutDesc.HoveredSlot(panelX, panelY, len(vInv.Slots), mx, my)
	if hoveredCargoIdx != -1 {
		slot := &vInv.Slots[hoveredCargoIdx]
		if slot.Item != nil && p.Inventory.AddItem(item.Clone(slot.Item), 1) {
			vInv.Remove(slot.Item, 1)
			p.RecalculateUpgrades()
		}
		return
	}

	// Transfer from vehicle upgrades to player
	if vUpg := v.GetUpgrades(); vUpg != nil {
		upgradeLayoutDesc := GetVehicleUpgradeLayout(len(vUpg.Slots))
		hoveredUpgradeIdx := upgradeLayoutDesc.HoveredSlot(panelX, panelY, len(vUpg.Slots), mx, my)
		if hoveredUpgradeIdx != -1 {
			slot := &vUpg.Slots[hoveredUpgradeIdx]
			if slot.Item != nil && p.Inventory.AddItem(item.Clone(slot.Item), 1) {
				vUpg.Remove(slot.Item, 1)
				p.RecalculateUpgrades()
			}
		}
	}
}

func drawItemIcon(screen *ebiten.Image, sx, sy, slotSz float32, i item.Item) {
	if i == nil {
		return
	}
	i.DrawIcon(screen, sx+slotSz/2.0, sy+slotSz/2.0, slotSz*0.45)
}
