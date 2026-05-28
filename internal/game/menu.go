package game

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

// Ingredient represents an item constructor and quantity required for a recipe.
type Ingredient struct {
	NewItem  func() item.Item
	Quantity int
}

// Recipe defines ingredients needed to craft a target item.
type Recipe struct {
	NewResult      func() item.Item
	ResultQuantity int
	Ingredients    []Ingredient
}

// Global list of craftable item upgrades
var CraftingRecipes = []Recipe{
	{
		NewResult: func() item.Item { return &item.O2TankHC{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.O2TankUHC{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.O2TankHC{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.Fins{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.Scanner{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.ScoutSubKit{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.HeavyMechKit{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 8},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
	},
	{
		NewResult: func() item.Item { return &item.UpgradeSolar{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
		},
	},
	{
		NewResult: func() item.Item { return &item.UpgradeSolarMKII{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.UpgradeSolar{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.UpgradeStorage{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.UpgradeStorageMKII{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.UpgradeStorage{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.UpgradeSolar{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 3},
		},
	},
	{
		NewResult: func() item.Item { return &item.SonarAmplifier{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
	},
	{
		NewResult: func() item.Item { return &item.PowerCell{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 1},
		},
	},
	{
		NewResult: func() item.Item { return &item.ThermalGenerator{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
	},
	{
		NewResult: func() item.Item { return &item.EscapeRocket{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.AbyssalOre{} }, Quantity: 10},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 10},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 5},
		},
	},
	{
		NewResult: func() item.Item { return &item.CookedFish{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.RawFish{} }, Quantity: 1},
		},
	},
	{
		NewResult: func() item.Item { return &item.CookedCrab{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.RawCrab{} }, Quantity: 1},
		},
	},
	{
		NewResult: func() item.Item { return &item.Titanium{} },
		ResultQuantity: 2,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ScrapMetal{} }, Quantity: 1},
		},
	},
	{
		NewResult: func() item.Item { return &item.Copper{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ElectronicWaste{} }, Quantity: 1},
		},
	},
}

// BaseMenuScene manages tab selections and base management interactions.
type BaseMenuScene struct {
	ActiveTab int
	ScrollY   float64
}

// NewBaseMenuScene instantiates a BaseMenuScene.
func NewBaseMenuScene() *BaseMenuScene {
	return &BaseMenuScene{
		ActiveTab: 0,
		ScrollY:   0,
	}
}

func (m *BaseMenuScene) OnEnter(g *Game) {
	g.currentState = StateBaseMenu
	m.ScrollY = 0
}

func (m *BaseMenuScene) OnExit(g *Game) {}

// Update handles mouse interactions within the menu tabs, crafting, and vault transfers.
func (m *BaseMenuScene) Update(g *Game) error {
	// Exit terminal check
	if g.Input.IsKeyJustPressed(ebiten.KeyE) || g.Input.IsKeyJustPressed(ebiten.KeyO) {
		g.TransitionTo(g.overworldState)
		return nil
	}

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
					if i == 2 && !b.HasModule(item.ModuleStorage) {
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
			// 1. Click Player Inventory to install upgrade module
			pStartX := panelX + 45
			pStartY := panelY + 140
			for r := 0; r < 3; r++ {
				for c := 0; c < 8; c++ {
					idx := r*8 + c
					if idx >= len(p.Inventory.Slots) {
						continue
					}
					sx := int(pStartX) + c*(40+6)
					sy := int(pStartY) + r*(40+6)
					if mx >= sx && mx < sx+40 && my >= sy && my < sy+40 {
						slot := &p.Inventory.Slots[idx]
						if slot.Item != nil {
							if b.InstallUpgrade(slot.Item) {
								p.Inventory.Remove(slot.Item, 1)
								p.RecalculateUpgrades()
							}
						}
					}
				}
			}

			// 2. Click Base Upgrades to uninstall upgrade module
			bStartX := panelX + 445
			bStartY := panelY + 140
			for c := 0; c < 4; c++ {
				sx := int(bStartX) + c*(40+6)
				sy := int(bStartY)
				if mx >= sx && mx < sx+40 && my >= sy && my < sy+40 {
					slot := &b.Upgrades.Slots[c]
					if slot.Item != nil {
						if p.Inventory.AddItem(item.Clone(slot.Item), 1) {
							b.Upgrades.Remove(slot.Item, 1)
							b.RecalculateProperties()
							p.RecalculateUpgrades()
						}
					}
				}
			}
		}

	case 1: // Fabricator tab (crafting)
		// Update scroll position using mouse wheel
		_, wy := g.Input.Wheel()
		if wy != 0 {
			m.ScrollY -= wy * 15
			maxScroll := float64(len(CraftingRecipes)*58 - 310)
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.ScrollY < 0 {
				m.ScrollY = 0
			} else if m.ScrollY > maxScroll {
				m.ScrollY = maxScroll
			}
		}

		if leftClicked {
			startX := int(panelX) + 30
			startY := int(panelY) + 95
			rowH := 58
			viewportMinY := startY + 25
			viewportMaxY := startY + 25 + 310

			// Only register clicks inside the viewport bounds
			if my >= viewportMinY && my < viewportMaxY {
				for i, rcp := range CraftingRecipes {
					// Apply scroll offset to find visual position
					ry := float64(viewportMinY) + float64(i*rowH) - m.ScrollY
					btnX := startX + 560
					btnY := int(ry) + 8

					// Verify if clicked the Craft button for this recipe
					if mx >= btnX && mx < btnX+160 && my >= btnY && my < btnY+35 {
						// Check if base power is available (consuming 10 power per craft)
						if b.Power >= 10.0 {
							// Verify player has all ingredients
							hasAll := true
							for _, ing := range rcp.Ingredients {
								if !p.Inventory.Has(ing.NewItem(), ing.Quantity) {
									hasAll = false
									break
								}
							}

							if hasAll {
								newItem := rcp.NewResult()
								if _, isRocket := newItem.(*item.EscapeRocket); isRocket {
									for _, ing := range rcp.Ingredients {
										p.Inventory.Remove(ing.NewItem(), ing.Quantity)
									}
									b.Power -= 10.0
									p.RecalculateUpgrades()
									g.TransitionTo(g.gameWonState)
									return nil
								}
								resQty := rcp.ResultQuantity
								if resQty <= 0 {
									resQty = 1
								}
								if p.Inventory.AddItem(newItem, resQty) {
									// Consume ingredients
									for _, ing := range rcp.Ingredients {
										p.Inventory.Remove(ing.NewItem(), ing.Quantity)
									}
									// Consume power
									b.Power -= 10.0
									p.RecalculateUpgrades()
								}
							}
						}
					}
				}
			}
		}

	case 2: // Storage Tab (Vault item transfer)
		if leftClicked && b.HasModule(item.ModuleStorage) {
			// Layout offsets for inventories
			const (
				cols   = 8
				slotSz = 40
				gap    = 6
			)
			pRows := len(p.Inventory.Slots) / cols
			bRows := len(b.Storage.Slots) / cols

			// Player Inventory Grid Bounds (Left)
			pStartX := panelX + 30
			pStartY := panelY + 110

			for r := 0; r < pRows; r++ {
				for c := 0; c < cols; c++ {
					idx := r*cols + c
					if idx >= len(p.Inventory.Slots) {
						continue
					}
					sx := int(pStartX) + c*(slotSz+gap)
					sy := int(pStartY) + r*(slotSz+gap)

					if mx >= sx && mx < sx+slotSz && my >= sy && my < sy+slotSz {
						slot := &p.Inventory.Slots[idx]
						if slot.Item != nil {
							// Transfer 1 item to base storage
							if b.Storage.AddItem(item.Clone(slot.Item), 1) {
								p.Inventory.Remove(slot.Item, 1)
								p.RecalculateUpgrades()
							}
						}
					}
				}
			}

			// Base Storage Grid Bounds (Right)
			bStartX := panelX + 430
			bStartY := panelY + 110

			for r := 0; r < bRows; r++ {
				for c := 0; c < cols; c++ {
					idx := r*cols + c
					if idx >= len(b.Storage.Slots) {
						continue
					}
					sx := int(bStartX) + c*(slotSz+gap)
					sy := int(bStartY) + r*(slotSz+gap)

					if mx >= sx && mx < sx+slotSz && my >= sy && my < sy+slotSz {
						slot := &b.Storage.Slots[idx]
						if slot.Item != nil {
							// Transfer 1 item to player inventory
							if p.Inventory.AddItem(item.Clone(slot.Item), 1) {
								b.Storage.Remove(slot.Item, 1)
								p.RecalculateUpgrades()
							}
						}
					}
				}
			}
		}

	case 3: // Medical Tab (healing)
		if leftClicked {
			healBtnX := int(panelX) + 260
			healBtnY := int(panelY) + 245
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
	if b.HasModule(item.ModuleSolar) {
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
		if i == 2 && !b.HasModule(item.ModuleStorage) {
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
		// Left Side: Player Inventory for slotting upgrades
		schematicX := panelX + 30
		schematicY := panelY + 95

		vector.FillRect(screen, schematicX, schematicY, 380, 335, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, schematicX, schematicY, 380, 335, 1.0, color.RGBA{48, 62, 85, 255}, false)

		ebitenutil.DebugPrintAt(screen, "PLAYER INVENTORY (CLICK UPGRADE TO INSTALL)", int(schematicX)+15, int(schematicY)+15)
		drawInventoryGrid(g, screen, float32(schematicX)+15, float32(schematicY)+45, p.Inventory)

		// Right Side: Installed Upgrades Grid & Module status
		upgradeX := panelX + 430
		upgradeY := panelY + 95

		vector.FillRect(screen, upgradeX, upgradeY, 340, 335, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, upgradeX, upgradeY, 340, 335, 1.0, color.RGBA{48, 62, 85, 255}, false)

		ebitenutil.DebugPrintAt(screen, "INSTALLED MODULES (CLICK TO UNINSTALL)", int(upgradeX)+15, int(upgradeY)+15)

		// Draw base upgrades slots (4 slots)
		const (
			slotSz = 40
			gap    = 6
		)
		cursor := g.Input.Cursor()
		mx, my := int(cursor.X), int(cursor.Y)

		for c := 0; c < 4; c++ {
			sx := upgradeX + 15 + float32(c*(slotSz+gap))
			sy := upgradeY + 45

			slotBg := color.RGBA{24, 30, 44, 255}
			slotBorder := color.RGBA{54, 68, 92, 255}

			isHovered := mx >= int(sx) && mx < int(sx+slotSz) && my >= int(sy) && my < int(sy+slotSz)
			if isHovered {
				slotBg = color.RGBA{38, 48, 70, 255}
				slotBorder = color.RGBA{100, 130, 180, 255}
			}

			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			if b.Upgrades != nil && c < len(b.Upgrades.Slots) {
				itemStack := b.Upgrades.Slots[c]
				if itemStack.Item != nil {
					itemClr := itemStack.Item.GetColor()
					cx := sx + slotSz/2.0
					cy := sy + slotSz/2.0
					vector.FillCircle(screen, cx, cy, 10, itemClr, false)
				}
			}
		}

		// Draw Schematic Category status below upgrades grid
		statusStartY := upgradeY + 110
		ebitenutil.DebugPrintAt(screen, "BASE SCHEMATIC SYSTEMS STATUS", int(upgradeX)+15, int(statusStartY))

		modulesList := []struct {
			mod  item.BaseModule
			name string
		}{
			{item.ModuleFabricator, "Fabricator Module"},
			{item.ModuleMedical, "Medical Bay"},
			{item.ModuleSolar, "Solar Array"},
			{item.ModuleStorage, "Storage Vault"},
		}

		for idx, modInfo := range modulesList {
			status := "NOT INSTALLED"
			statusColor := color.RGBA{220, 80, 80, 255}
			displayName := modInfo.name

			if b.HasModule(modInfo.mod) {
				status = "ACTIVE"
				statusColor = color.RGBA{60, 210, 110, 255}
				if modInfo.mod == item.ModuleSolar && b.HasModule(item.ModuleSolarMKII) {
					displayName = "Solar Array MKII"
				} else if modInfo.mod == item.ModuleStorage && b.HasModule(item.ModuleStorageMKII) {
					displayName = "Storage Vault MKII"
				}
			}

			sy := statusStartY + 25 + float32(idx*50)
			vector.FillRect(screen, upgradeX+15, sy, 310, 42, color.RGBA{24, 32, 48, 255}, false)
			vector.StrokeRect(screen, upgradeX+15, sy, 310, 42, 0.8, color.RGBA{58, 75, 100, 255}, false)

			ebitenutil.DebugPrintAt(screen, displayName, int(upgradeX)+25, int(sy)+6)
			vector.FillRect(screen, upgradeX+25, sy+22, 95, 14, statusColor, false)
			ebitenutil.DebugPrintAt(screen, status, int(upgradeX)+28, int(sy)+23)
		}

	case 1: // FABRICATOR (Crafting menu)
		startX := panelX + 30
		startY := panelY + 95
		rowH := float32(58)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FABRICATOR MENU (POWER COST: 10 HP PER CRAFT - CURRENT: %.0f HP)", b.Power), int(startX)+10, int(startY))

		// Viewport parameters
		viewportY := startY + 25
		viewportH := float32(310)

		// Draw clipped recipes
		rect := image.Rect(int(startX), int(viewportY), int(startX)+740, int(viewportY+viewportH))
		subImg := screen.SubImage(rect)
		if subImg != nil {
			clippedScreen := subImg.(*ebiten.Image)
			for i, rcp := range CraftingRecipes {
				ry := viewportY + float32(i)*rowH - float32(m.ScrollY)

				// Draw row background panel
				vector.FillRect(clippedScreen, float32(startX), ry, 740, 52, color.RGBA{18, 24, 38, 255}, false)
				vector.StrokeRect(clippedScreen, float32(startX), ry, 740, 52, 0.8, color.RGBA{45, 58, 78, 255}, false)

				// Draw output name
				resultName := rcp.NewResult().GetName()
				resQty := rcp.ResultQuantity
				if resQty > 1 {
					resultName = fmt.Sprintf("%s x%d", resultName, resQty)
				}
				ebitenutil.DebugPrintAt(clippedScreen, resultName, int(startX)+15, int(ry)+6)

				// Draw ingredients checklist
				ingText := "Ingredients: "
				hasAll := true
				for j, ing := range rcp.Ingredients {
					ingredient := ing.NewItem()
					qtyInInv := p.Inventory.Count(ingredient)

					checkChar := "X"
					if qtyInInv >= ing.Quantity {
						checkChar = "✓"
					} else {
						hasAll = false
					}

					ingName := ingredient.GetName()
					ingText += fmt.Sprintf("[%s] %s (%d/%d)  ", checkChar, ingName, qtyInInv, ing.Quantity)
					if j < len(rcp.Ingredients)-1 {
						ingText += "|  "
					}
				}
				ebitenutil.DebugPrintAt(clippedScreen, ingText, int(startX)+15, int(ry)+28)

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

				vector.FillRect(clippedScreen, float32(startX)+560, ry+8, 160, 35, btnBg, false)
				vector.StrokeRect(clippedScreen, float32(startX)+560, ry+8, 160, 35, 1.0, color.RGBA{80, 100, 130, 255}, false)
				ebitenutil.DebugPrintAt(clippedScreen, btnLabel, int(startX)+608, int(ry)+18)
			}
		}

		// Draw scrollbar on the right side if content overflows
		totalH := float32(len(CraftingRecipes) * 58)
		if totalH > viewportH {
			scrollBarX := startX + 740 + 6
			vector.FillRect(screen, float32(scrollBarX), viewportY, 6, viewportH, color.RGBA{24, 30, 44, 128}, false)

			handleH := viewportH * (viewportH / totalH)
			if handleH < 15 {
				handleH = 15 // minimum size
			}
			maxScroll := totalH - viewportH
			var handleY float32
			if maxScroll > 0 {
				handleY = viewportY + (float32(m.ScrollY)/maxScroll)*(viewportH-handleH)
			} else {
				handleY = viewportY
			}

			vector.FillRect(screen, float32(scrollBarX), handleY, 6, handleH, color.RGBA{100, 130, 180, 255}, false)
			vector.StrokeRect(screen, float32(scrollBarX), handleY, 6, handleH, 0.8, color.RGBA{140, 170, 220, 255}, false)
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
func drawInventoryGrid(g *Game, screen *ebiten.Image, startX, startY float32, inv *item.Inventory) {
	const (
		cols   = 8
		slotSz = 40
		gap    = 6
	)
	rows := len(inv.Slots) / cols

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

			itemStack := inv.Slots[idx]
			if itemStack.Item != nil {
				itemClr := itemStack.Item.GetColor()

				cx := sx + slotSz/2.0
				cy := sy + slotSz/2.0
				vector.FillCircle(screen, cx, cy, 10, itemClr, false)

				if itemStack.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", itemStack.Quantity), int(sx)+4, int(sy)+slotSz-15)
				}
			}
		}
	}
}
