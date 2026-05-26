package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Recipe defines ingredients needed to craft a target item.
type Recipe struct {
	Result      ItemType
	Ingredients []ItemStack
}

// Global list of craftable item upgrades
var CraftingRecipes = []Recipe{
	{
		Result: ItemO2TankHC,
		Ingredients: []ItemStack{
			{Type: ItemTitanium, Quantity: 4},
			{Type: ItemQuartz, Quantity: 2},
		},
	},
	{
		Result: ItemO2TankUHC,
		Ingredients: []ItemStack{
			{Type: ItemO2TankHC, Quantity: 1},
			{Type: ItemTitanium, Quantity: 5},
			{Type: ItemCopper, Quantity: 3},
			{Type: ItemQuartz, Quantity: 2},
		},
	},
	{
		Result: ItemFins,
		Ingredients: []ItemStack{
			{Type: ItemTitanium, Quantity: 3},
			{Type: ItemCopper, Quantity: 2},
		},
	},
	{
		Result: ItemScanner,
		Ingredients: []ItemStack{
			{Type: ItemTitanium, Quantity: 2},
			{Type: ItemCopper, Quantity: 1},
			{Type: ItemQuartz, Quantity: 2},
		},
	},
	{
		Result: ItemScoutSub,
		Ingredients: []ItemStack{
			{Type: ItemTitanium, Quantity: 6},
			{Type: ItemCopper, Quantity: 4},
			{Type: ItemQuartz, Quantity: 2},
		},
	},
	{
		Result: ItemHeavyMech,
		Ingredients: []ItemStack{
			{Type: ItemTitanium, Quantity: 8},
			{Type: ItemCopper, Quantity: 6},
			{Type: ItemQuartz, Quantity: 4},
		},
	},
}

// BaseMenuScene manages tab selections and base management interactions.
type BaseMenuScene struct {
	ActiveTab int
}

// NewBaseMenuScene instantiates a BaseMenuScene.
func NewBaseMenuScene() *BaseMenuScene {
	return &BaseMenuScene{
		ActiveTab: 0,
	}
}

func (m *BaseMenuScene) OnEnter(g *Game) {
	g.currentState = StateBaseMenu
}

func (m *BaseMenuScene) OnExit(g *Game) {}

// Update handles mouse interactions within the menu tabs, crafting, and vault transfers.
func (m *BaseMenuScene) Update(g *Game) error {
	p := g.player
	b := g.baseStation

	cursor := g.Input.Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
	leftClicked := g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	// Panel placement calculations
	const (
		panelW = 800
		panelH = 500
	)
	panelX := float64(ScreenWidth-panelW) / 2.0
	panelY := float64(ScreenHeight-panelH) / 2.0

	// 1. Check tab button clicks
	if leftClicked {
		ty := int(panelY) + 40
		if my >= ty && my < ty+30 {
			for i := 0; i < 4; i++ {
				tx := int(panelX) + 30 + i*150
				if mx >= tx && mx < tx+140 {
					// Lock storage tab until storage vault module is built
					if i == 2 && !b.Modules[ModuleStorage] {
						continue
					}
					m.ActiveTab = i
				}
			}
		}
	}

	// 2. Perform updates depending on the active tab
	switch m.ActiveTab {
	case 0: // Overview tab (upgrades)
		if leftClicked {
			// Solar Upgrade Button click check
			solarX := int(panelX) + 480
			solarY := int(panelY) + 110
			if mx >= solarX && mx < solarX+260 && my >= solarY && my < solarY+45 {
				if !b.Modules[ModuleSolar] && p.Inventory.HasItem(ItemTitanium, 5) && p.Inventory.HasItem(ItemCopper, 3) {
					p.Inventory.RemoveItem(ItemTitanium, 5)
					p.Inventory.RemoveItem(ItemCopper, 3)
					b.Modules[ModuleSolar] = true
				}
			}

			// Storage Upgrade Button click check
			vaultX := int(panelX) + 480
			vaultY := int(panelY) + 175
			if mx >= vaultX && mx < vaultX+260 && my >= vaultY && my < vaultY+45 {
				if !b.Modules[ModuleStorage] && p.Inventory.HasItem(ItemTitanium, 4) && p.Inventory.HasItem(ItemQuartz, 2) {
					p.Inventory.RemoveItem(ItemTitanium, 4)
					p.Inventory.RemoveItem(ItemQuartz, 2)
					b.Modules[ModuleStorage] = true
				}
			}
		}

	case 1: // Fabricator tab (crafting)
		if leftClicked {
			startX := int(panelX) + 40
			startY := int(panelY) + 90
			rowH := 58

			for i, rcp := range CraftingRecipes {
				btnX := startX + 540
				btnY := startY + i*rowH + 8

				// Check if clicked the Craft button for this recipe
				if mx >= btnX && mx < btnX+140 && my >= btnY && my < btnY+35 {
					// Check if base power is available (consuming 10 power per craft)
					if b.Power >= 10.0 {
						// Verify player has all ingredients
						hasAll := true
						for _, ing := range rcp.Ingredients {
							if !p.Inventory.HasItem(ing.Type, ing.Quantity) {
								hasAll = false
								break
							}
						}

						if hasAll {
							// Check if inventory has slot space for result
							if p.Inventory.AddItem(rcp.Result, 1) {
								// Consume ingredients
								for _, ing := range rcp.Ingredients {
									p.Inventory.RemoveItem(ing.Type, ing.Quantity)
								}
								// Consume power
								b.Power -= 10.0

								// Apply upgrades directly to player metrics immediately
								if rcp.Result == ItemO2TankHC {
									p.MaxOxygen = 160.0
								} else if rcp.Result == ItemO2TankUHC {
									// Remove O2TankHC from inventory since it's upgraded to UHC
									p.Inventory.RemoveItem(ItemO2TankHC, 1)
									p.MaxOxygen = 240.0
								} else if rcp.Result == ItemFins {
									// Swim speed boost will be checked in movement physics
								}
							}
						}
					}
				}
			}
		}

	case 2: // Storage Tab (Vault item transfer)
		if leftClicked && b.Modules[ModuleStorage] {
			// Layout offsets for inventories
			const (
				cols   = 8
				rows   = 3
				slotSz = 40
				gap    = 6
			)

			// Player Inventory Grid Bounds (Left)
			pStartX := panelX + 30
			pStartY := panelY + 110

			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					idx := r*cols + c
					sx := int(pStartX) + c*(slotSz+gap)
					sy := int(pStartY) + r*(slotSz+gap)

					if mx >= sx && mx < sx+slotSz && my >= sy && my < sy+slotSz {
						item := &p.Inventory.Slots[idx]
						if item.Type != ItemNone {
							// Transfer 1 item to base storage
							if b.Storage.AddItem(item.Type, 1) {
								p.Inventory.RemoveItem(item.Type, 1)
							}
						}
					}
				}
			}

			// Base Storage Grid Bounds (Right)
			bStartX := panelX + 430
			bStartY := panelY + 110

			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					idx := r*cols + c
					sx := int(bStartX) + c*(slotSz+gap)
					sy := int(bStartY) + r*(slotSz+gap)

					if mx >= sx && mx < sx+slotSz && my >= sy && my < sy+slotSz {
						item := &b.Storage.Slots[idx]
						if item.Type != ItemNone {
							// Transfer 1 item to player inventory
							if p.Inventory.AddItem(item.Type, 1) {
								b.Storage.RemoveItem(item.Type, 1)
							}
						}
					}
				}
			}
		}

	case 3: // Medical Tab (healing)
		if leftClicked {
			healBtnX := int(panelX) + 260
			healBtnY := int(panelY) + 160
			if mx >= healBtnX && mx < healBtnX+280 && my >= healBtnY && my < healBtnY+60 {
				// Costs 15 base power, heals 40 HP
				if b.Power >= 15.0 && p.CurrentHealth < p.MaxHealth {
					b.Power -= 15.0
					p.CurrentHealth += 40.0
					if p.CurrentHealth > p.MaxHealth {
						p.CurrentHealth = p.MaxHealth
					}
				}
			}
		}
	}

	return nil
}

// Draw renders the management base menu tabs and UI controls.
func (m *BaseMenuScene) Draw(g *Game, screen *ebiten.Image) {
	p := g.player
	b := g.baseStation

	const (
		panelW = 800
		panelH = 500
	)
	panelX := float32(ScreenWidth-panelW) / 2.0
	panelY := float32(ScreenHeight-panelH) / 2.0

	// Draw main window panel (transparent dark slate)
	panelBg := color.RGBA{12, 16, 26, 242}
	vector.FillRect(screen, panelX, panelY, panelW, panelH, panelBg, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)

	// Draw base title and stats
	ebitenutil.DebugPrintAt(screen, "BASE ANCHOR TERMINAL - LIFE POD 5", int(panelX)+20, int(panelY)+12)
	powerText := fmt.Sprintf("BASE POWER: %.0f/%.0f HP (Recharge: solar panels)", b.Power, b.MaxPower)
	if b.Modules[ModuleSolar] {
		powerText = fmt.Sprintf("BASE POWER: %.0f/%.0f HP (Recharging: +Solar Active)", b.Power, b.MaxPower)
	}
	ebitenutil.DebugPrintAt(screen, powerText, int(panelX)+420, int(panelY)+12)

	// 1. Draw Tab Buttons
	tabLabels := []string{"1. OVERVIEW", "2. FABRICATOR", "3. BASE VAULT", "4. MEDICAL"}
	for i := 0; i < 4; i++ {
		tx := panelX + 30 + float32(i*150)
		ty := panelY + 40

		// Locked vault indicator
		label := tabLabels[i]
		if i == 2 && !b.Modules[ModuleStorage] {
			label = "3. VAULT [LOCKED]"
		}

		tabBg := color.RGBA{18, 24, 38, 255}
		tabBorder := color.RGBA{45, 58, 78, 255}

		if m.ActiveTab == i {
			tabBg = color.RGBA{32, 45, 68, 255}
			tabBorder = color.RGBA{95, 125, 165, 255}
		}

		vector.FillRect(screen, tx, ty, 140, 30, tabBg, false)
		vector.StrokeRect(screen, tx, ty, 140, 30, 1.0, tabBorder, false)
		
		// Draw label center offset (manual visual approximation)
		ebitenutil.DebugPrintAt(screen, label, int(tx)+12, int(ty)+6)
	}

	// Draw Tab Divider Line
	vector.StrokeLine(screen, panelX+20, panelY+75, panelX+panelW-20, panelY+75, 1.0, color.RGBA{68, 88, 120, 255}, false)

	// 2. Draw active tab container content
	switch m.ActiveTab {
	case 0: // OVERVIEW (Modular Schematic & Base Station upgrades)
		// Draw Left: modular slots grid
		schematicX := panelX + 30
		schematicY := panelY + 95

		vector.FillRect(screen, schematicX, schematicY, 380, 335, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, schematicX, schematicY, 380, 335, 1.0, color.RGBA{48, 62, 85, 255}, false)

		ebitenutil.DebugPrintAt(screen, "BASE MODULE SCHEMATICS STATUS", int(schematicX)+15, int(schematicY)+15)

		modulesList := []BaseModule{ModuleFabricator, ModuleMedical, ModuleSolar, ModuleStorage}
		for idx, mod := range modulesList {
			status := "NOT INSTALLED"
			statusColor := color.RGBA{220, 80, 80, 255}
			if b.Modules[mod] {
				status = "INSTALLED"
				statusColor = color.RGBA{60, 210, 110, 255}
			}

			sy := schematicY + 50 + float32(idx*65)
			vector.FillRect(screen, schematicX+15, sy, 350, 50, color.RGBA{24, 32, 48, 255}, false)
			vector.StrokeRect(screen, schematicX+15, sy, 350, 50, 0.8, color.RGBA{58, 75, 100, 255}, false)

			ebitenutil.DebugPrintAt(screen, mod.String(), int(schematicX)+25, int(sy)+8)
			vector.FillRect(screen, schematicX+25, sy+28, 90, 16, statusColor, false)
			ebitenutil.DebugPrintAt(screen, status, int(schematicX)+28, int(sy)+29)
		}

		// Draw Right: Upgrade installer shops
		upgradeX := panelX + 440
		upgradeY := panelY + 95

		ebitenutil.DebugPrintAt(screen, "AVAILABLE INSTALLATIONS", int(upgradeX), int(upgradeY)+15)

		// 1. Solar installation
		sy := upgradeY + 45
		vector.FillRect(screen, upgradeX, sy, 320, 100, color.RGBA{20, 26, 38, 255}, false)
		vector.StrokeRect(screen, upgradeX, sy, 320, 100, 1.0, color.RGBA{50, 68, 92, 255}, false)

		ebitenutil.DebugPrintAt(screen, "SOLAR ARRAY INSTALLATION", int(upgradeX)+12, int(sy)+10)
		ebitenutil.DebugPrintAt(screen, "Recharges base power during surface stays.", int(upgradeX)+12, int(sy)+30)
		ebitenutil.DebugPrintAt(screen, "Cost: 5 Titanium, 3 Copper", int(upgradeX)+12, int(sy)+50)

		solBtnBg := color.RGBA{38, 52, 75, 255}
		solBtnTxt := "BUILD MODULE"
		if b.Modules[ModuleSolar] {
			solBtnBg = color.RGBA{30, 80, 50, 255}
			solBtnTxt = "CONSTRUCTED"
		}
		vector.FillRect(screen, upgradeX+12, sy+70, 296, 22, solBtnBg, false)
		ebitenutil.DebugPrintAt(screen, solBtnTxt, int(upgradeX)+110, int(sy)+73)

		// 2. Storage vault installation
		vy := upgradeY + 165
		vector.FillRect(screen, upgradeX, vy, 320, 100, color.RGBA{20, 26, 38, 255}, false)
		vector.StrokeRect(screen, upgradeX, vy, 320, 100, 1.0, color.RGBA{50, 68, 92, 255}, false)

		ebitenutil.DebugPrintAt(screen, "STORAGE VAULT INSTALLATION", int(upgradeX)+12, int(vy)+10)
		ebitenutil.DebugPrintAt(screen, "Unlocks base storage chest tab for hoarding.", int(upgradeX)+12, int(vy)+30)
		ebitenutil.DebugPrintAt(screen, "Cost: 4 Titanium, 2 Quartz", int(upgradeX)+12, int(vy)+50)

		vBtnBg := color.RGBA{38, 52, 75, 255}
		vBtnTxt := "BUILD MODULE"
		if b.Modules[ModuleStorage] {
			vBtnBg = color.RGBA{30, 80, 50, 255}
			vBtnTxt = "CONSTRUCTED"
		}
		vector.FillRect(screen, upgradeX+12, vy+70, 296, 22, vBtnBg, false)
		ebitenutil.DebugPrintAt(screen, vBtnTxt, int(upgradeX)+110, int(vy)+73)

	case 1: // FABRICATOR (Crafting menu)
		startX := panelX + 30
		startY := panelY + 95
		rowH := float32(58)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FABRICATOR MENU (POWER COST: 10 HP PER CRAFT - CURRENT: %.0f HP)", b.Power), int(startX)+10, int(startY))

		for i, rcp := range CraftingRecipes {
			ry := startY + 25 + float32(i)*rowH

			// Draw row background panel
			vector.FillRect(screen, startX, ry, 740, 52, color.RGBA{18, 24, 38, 255}, false)
			vector.StrokeRect(screen, startX, ry, 740, 52, 0.8, color.RGBA{45, 58, 78, 255}, false)

			// Draw output name
			ebitenutil.DebugPrintAt(screen, rcp.Result.String(), int(startX)+15, int(ry)+6)

			// Draw ingredients checklist
			ingText := "Ingredients: "
			hasAll := true
			for j, ing := range rcp.Ingredients {
				qtyInInv := 0
				for _, slot := range p.Inventory.Slots {
					if slot.Type == ing.Type {
						qtyInInv += slot.Quantity
					}
				}

				checkChar := "X"
				if qtyInInv >= ing.Quantity {
					checkChar = "✓"
				} else {
					hasAll = false
				}

				ingText += fmt.Sprintf("[%s] %s (%d/%d)  ", checkChar, ing.Type.String(), qtyInInv, ing.Quantity)
				if j < len(rcp.Ingredients)-1 {
					ingText += "|  "
				}
			}
			ebitenutil.DebugPrintAt(screen, ingText, int(startX)+15, int(ry)+28)

			// Craft button
			btnBg := color.RGBA{50, 70, 100, 255}
			btnLabel := "CRAFT ITEM"
			if !hasAll {
				btnBg = color.RGBA{38, 42, 50, 255}
				btnLabel = "NO MATERIALS"
			} else if b.Power < 10.0 {
				btnBg = color.RGBA{50, 20, 20, 255}
				btnLabel = "NO POWER"
			}

			vector.FillRect(screen, startX+560, ry+8, 160, 35, btnBg, false)
			vector.StrokeRect(screen, startX+560, ry+8, 160, 35, 1.0, color.RGBA{80, 100, 130, 255}, false)
			ebitenutil.DebugPrintAt(screen, btnLabel, int(startX)+608, int(ry)+18)
		}

	case 2: // BASE VAULT STORAGE (Items transfer)
		// Left: Player Inventory
		pStartX := panelX + 30
		pStartY := panelY + 110
		ebitenutil.DebugPrintAt(screen, "PLAYER INVENTORY (CLICK TO STORE)", int(pStartX), int(pStartY)-25)

		drawInventoryGrid(g, screen, pStartX, pStartY, p.Inventory)

		// Right: Base Vault Storage
		bStartX := panelX + 430
		bStartY := panelY + 110
		ebitenutil.DebugPrintAt(screen, "BASE VAULT (CLICK TO TAKE)", int(bStartX), int(bStartY)-25)

		drawInventoryGrid(g, screen, bStartX, bStartY, b.Storage)

		// Middle arrow graphic
		arrowX := panelX + 395
		arrowY := panelY + 180
		ebitenutil.DebugPrintAt(screen, "<-->", int(arrowX), int(arrowY))

	case 3: // MEDICAL BAY (Healing)
		medX := panelX + 30
		medY := panelY + 95

		vector.FillRect(screen, medX, medY, 740, 335, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, medX, medY, 740, 335, 1.0, color.RGBA{48, 62, 85, 255}, false)

		ebitenutil.DebugPrintAt(screen, "INFIRMARY / MEDICAL SCANNER UNIT", int(medX)+220, int(medY)+40)
		
		statusHp := fmt.Sprintf("DIVER HEALTH STATUS: %.0f / %.0f HP", p.CurrentHealth, p.MaxHealth)
		ebitenutil.DebugPrintAt(screen, statusHp, int(medX)+240, int(medY)+100)

		// Medical scan panel
		vector.FillRect(screen, medX+230, medY+150, 280, 60, color.RGBA{22, 38, 55, 255}, false)
		
		healText := "ACTIVATE DECONTAMINATION HEAL"
		healSubText := "Costs 15 base energy (Heals +40 HP)"
		if p.CurrentHealth >= p.MaxHealth {
			healText = "DIVER HEALTH SECURED"
			healSubText = "Maximum health level achieved."
		} else if b.Power < 15.0 {
			healText = "INSUFFICIENT BASE POWER"
			healSubText = "Charge base power to activate infirmary."
		}

		ebitenutil.DebugPrintAt(screen, healText, int(medX)+260, int(medY)+166)
		ebitenutil.DebugPrintAt(screen, healSubText, int(medX)+260, int(medY)+186)
		vector.StrokeRect(screen, medX+230, medY+150, 280, 60, 1.0, color.RGBA{70, 100, 140, 255}, false)
	}

	// Exit instructions overlay at bottom
	ebitenutil.DebugPrintAt(screen, "Press [E] or [O] to Close Terminal Interface", int(panelX)+260, int(panelY)+panelH-25)
}

// drawInventoryGrid helper draws standard slots grid for items storage transfer.
func drawInventoryGrid(g *Game, screen *ebiten.Image, startX, startY float32, inv *Inventory) {
	const (
		cols   = 8
		rows   = 3
		slotSz = 40
		gap    = 6
	)

	cursor := g.Input.Cursor()
	mx, my := int(cursor.X), int(cursor.Y)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx >= len(inv.Slots) {
				continue
			}

			sx := startX + float32(c*(slotSz+gap))
			sy := startY + float32(r*(slotSz+gap))

			slotBg := color.RGBA{24, 30, 44, 255}
			slotBorder := color.RGBA{54, 68, 92, 255}

			isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
			if isHovered {
				slotBg = color.RGBA{38, 48, 70, 255}
				slotBorder = color.RGBA{100, 130, 180, 255}
			}

			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			item := inv.Slots[idx]
			if item.Type != ItemNone {
				var itemClr color.Color
				switch item.Type {
				case ItemTitanium:
					itemClr = color.RGBA{168, 178, 188, 255}
				case ItemCopper:
					itemClr = color.RGBA{218, 118, 48, 255}
				case ItemQuartz:
					itemClr = color.RGBA{48, 218, 245, 255}
				case ItemAbyssalOre:
					itemClr = color.RGBA{148, 48, 218, 255}
				default:
					itemClr = color.RGBA{98, 198, 148, 255}
				}

				cx := sx + slotSz/2.0
				cy := sy + slotSz/2.0
				vector.FillCircle(screen, cx, cy, 10, itemClr, false)

				if item.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.Quantity), int(sx)+4, int(sy)+slotSz-15)
				}
			}
		}
	}
}
