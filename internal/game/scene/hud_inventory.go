package scene

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

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

	// Player Hotbar Grid
	hotbarLayout := SoloHotbarGridDescriptor
	hotbarLabelY := gearY + float32(gearLayout.SlotSz) + 15.0
	ebitenutil.DebugPrintAt(screen, "HOTBAR / ACTIVE COUNTERMEASURES (CLICK TO ASSIGN)", int(panelX)+20, int(hotbarLabelY))

	for c := 0; c < 5; c++ {
		rect := hotbarLayout.SlotRect(panelX, panelY, c)
		sx := float32(rect.Min.X)
		sy := float32(rect.Min.Y)
		slotSz := float32(hotbarLayout.SlotSz)

		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}
		borderWidth := float32(1.0)
		isHovered := mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
		if isHovered {
			slotBg = color.RGBA{32, 42, 62, 255}
			slotBorder = color.RGBA{95, 125, 165, 255}
			if p.Hotbar != nil && p.Hotbar.Slots[c].Item != nil {
				hoveredItemName = p.Hotbar.Slots[c].Item.GetName()
			}
		}
		if p.ActiveSlot == c {
			slotBg = color.RGBA{26, 40, 64, 255}
			slotBorder = color.RGBA{0, 230, 255, 255}
			borderWidth = 1.8
		}

		vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
		vector.StrokeRect(screen, sx, sy, slotSz, slotSz, borderWidth, slotBorder, false)

		if p.Hotbar != nil && p.Hotbar.Slots[c].Item != nil {
			drawItemIcon(screen, sx, sy, slotSz, p.Hotbar.Slots[c].Item)
			if p.Hotbar.Slots[c].Quantity > 1 {
				ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", p.Hotbar.Slots[c].Quantity), int(sx)+6, int(sy)+int(slotSz)-17)
			}
		} else {
			drawMiningToolOutline(screen, sx, sy, slotSz)
		}

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[%d]", c+1), int(sx)+int(slotSz)/2-9, int(sy)-18)
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
	p := g.GetPlayer()
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

	// Player Hotbar Grid in Vehicle View
	hotbarVLayout := VehicleHotbarGridDescriptor
	for c := 0; c < 5; c++ {
		rect := hotbarVLayout.SlotRect(panelX, panelY, c)
		sx := float32(rect.Min.X)
		sy := float32(rect.Min.Y)
		slotSz := float32(hotbarVLayout.SlotSz)

		slotBg := color.RGBA{20, 26, 38, 255}
		slotBorder := color.RGBA{48, 60, 80, 255}
		borderWidth := float32(1.0)
		isHovered := mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
		if isHovered {
			slotBg = color.RGBA{32, 42, 62, 255}
			slotBorder = color.RGBA{95, 125, 165, 255}
			if p.Hotbar != nil && p.Hotbar.Slots[c].Item != nil {
				hoveredItemName = p.Hotbar.Slots[c].Item.GetName()
			}
		}
		if p.ActiveSlot == c {
			slotBg = color.RGBA{26, 40, 64, 255}
			slotBorder = color.RGBA{0, 230, 255, 255}
			borderWidth = 1.8
		}

		vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
		vector.StrokeRect(screen, sx, sy, slotSz, slotSz, borderWidth, slotBorder, false)

		if p.Hotbar != nil && p.Hotbar.Slots[c].Item != nil {
			drawItemIcon(screen, sx, sy, slotSz, p.Hotbar.Slots[c].Item)
			if p.Hotbar.Slots[c].Quantity > 1 {
				ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", p.Hotbar.Slots[c].Quantity), int(sx)+5, int(sy)+int(slotSz)-16)
			}
		} else {
			drawMiningToolOutline(screen, sx, sy, slotSz)
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

	tooltipY := float32(panelY + layoutP.PanelH - 42)
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
	panelY := float64(config.ScreenHeight-470) / 2.0

	// Main inventory grid
	hoveredIdx := SoloInventoryGridDescriptor.HoveredSlot(panelX, panelY, len(p.Inventory.Slots), mx, my)
	if hoveredIdx != -1 {
		slot := &p.Inventory.Slots[hoveredIdx]
		if slot.Item != nil {
			// Player upgrades (O2 tank, fins) go to equipment upgrades
			if upg, ok := slot.Item.(item.PlayerUpgradeItem); ok && upg.IsPlayerUpgrade() {
				g.ActivatePlayerItem(slot.Item)
				return
			}
			// Deployable vehicle kits should also deploy directly!
			if _, isDeployable := slot.Item.(vehicle.Deployable); isDeployable {
				g.ActivatePlayerItem(slot.Item)
				return
			}
			// Other items try to go to Hotbar
			if p.Hotbar.AddItem(slot.Item, 1) {
				p.Inventory.Remove(slot.Item, 1)
				p.RecalculateUpgrades()
			} else {
				// Fallback to active tool/deploy direct use
				g.ActivatePlayerItem(slot.Item)
			}
		}
		return
	}

	// Hotbar grid
	hoveredHotbarIdx := SoloHotbarGridDescriptor.HoveredSlot(panelX, panelY, len(p.Hotbar.Slots), mx, my)
	if hoveredHotbarIdx != -1 {
		slot := &p.Hotbar.Slots[hoveredHotbarIdx]
		if slot.Item != nil {
			if _, isDeployable := slot.Item.(vehicle.Deployable); isDeployable {
				g.ActivatePlayerItem(slot.Item)
				return
			}
			// Move item from hotbar back to main inventory
			if p.Inventory.AddItem(item.Clone(slot.Item), 1) {
				p.Hotbar.Remove(slot.Item, 1)
				p.RecalculateUpgrades()
			}
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
	panelY := float64(config.ScreenHeight-410) / 2.0

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
			// Player upgrades go directly to upgrades/vehicle if vehicle open
			if upg, ok := slot.Item.(item.PlayerUpgradeItem); ok && upg.IsPlayerUpgrade() {
				g.TransferToVehicle(slot.Item)
				return
			}
			// Try to move to hotbar first
			if p.Hotbar.AddItem(slot.Item, 1) {
				p.Inventory.Remove(slot.Item, 1)
				p.RecalculateUpgrades()
			} else {
				// Otherwise transfer to vehicle
				g.TransferToVehicle(slot.Item)
			}
		}
		return
	}

	// Player hotbar grid clicked (in vehicle view)
	hoveredHotbarIdx := VehicleHotbarGridDescriptor.HoveredSlot(panelX, panelY, len(p.Hotbar.Slots), mx, my)
	if hoveredHotbarIdx != -1 {
		slot := &p.Hotbar.Slots[hoveredHotbarIdx]
		if slot.Item != nil {
			// Transfer from hotbar back to player inventory
			if p.Inventory.AddItem(item.Clone(slot.Item), 1) {
				p.Hotbar.Remove(slot.Item, 1)
				p.RecalculateUpgrades()
			}
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
