package scene

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/story"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
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
	Tier           int
	Unlocked       bool
}

// CraftingRecipes is the global list of craftable item upgrades.
var CraftingRecipes = []Recipe{
	{
		NewResult: func() item.Item { return &item.O2TankHC{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.O2TankUHC{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.O2TankHC{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.Fins{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.Scanner{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &vehicle.ScoutSubKit{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &vehicle.HeavyMechKit{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 8},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
		Tier:     2,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.UpgradeSolar{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.UpgradeSolarMKII{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.UpgradeSolar{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.UpgradeStorage{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
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
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.SonarAmplifier{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.PowerCell{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.ThermalGenerator{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.EscapeRocket{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.AbyssalOre{} }, Quantity: 10},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 10},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 5},
		},
		Tier:     2,
		Unlocked: false,
	},
	{
		NewResult:      func() item.Item { return &item.CookedFish{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.RawFish{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.CookedCrab{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.RawCrab{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.Titanium{} },
		ResultQuantity: 2,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ScrapMetal{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.Copper{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ElectronicWaste{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.SonicDecoy{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ElectronicWaste{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.ChemicalDeterrent{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult:      func() item.Item { return &item.DecoyLauncher{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult:      func() item.Item { return &item.ChemicalDischarger{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
		Tier:     1,
		Unlocked: false,
	},
}

// DefaultCraftingRecipes returns a fresh copy of the default CraftingRecipes slice.
func DefaultCraftingRecipes() []Recipe {
	recipes := make([]Recipe, len(CraftingRecipes))
	for i, rcp := range CraftingRecipes {
		recipes[i] = rcp
	}
	return recipes
}

// MenuContext defines the narrow context interface required by BaseMenuScene.
type MenuContext interface {
	GetInput() InputSource
	GetPlayer() *player.Player
	GetBaseStation() *base.BaseStation
	GetCraftingRecipes() []Recipe
	GetStoryManager() *story.StoryManager
	IsMenuOpenedAnywhere() bool
	ClosePDA()
	TransitionToOverworld()
	TransitionToGameWon()
	SetCurrentState(s State)
	SetMineWarning(msg string, duration, level int)
}

// BaseMenuScene manages tab selections and base management interactions.
type BaseMenuScene struct {
	ActiveTab         int
	ScrollY           float64
	SelectedLoreIndex int
}

// NewBaseMenuScene instantiates a BaseMenuScene.
func NewBaseMenuScene() *BaseMenuScene {
	return &BaseMenuScene{ActiveTab: 0, ScrollY: 0, SelectedLoreIndex: 0}
}

func (m *BaseMenuScene) OnEnter(g GameContext) {
	m.onEnter(g)
}

func (m *BaseMenuScene) onEnter(g MenuContext) {
	g.SetCurrentState(StateBaseMenu)
	m.ScrollY = 0
}

func (m *BaseMenuScene) OnExit(g GameContext) {
	m.onExit(g)
}

func (m *BaseMenuScene) onExit(g MenuContext) {}

// Update handles mouse interactions within the menu tabs, crafting, and vault transfers.
func (m *BaseMenuScene) Update(g GameContext) error {
	return m.update(g)
}

func (m *BaseMenuScene) update(g MenuContext) error {
	inp := g.GetInput()

	if inp.IsKeyJustPressed(ebiten.KeyE) || inp.IsKeyJustPressed(ebiten.KeyO) {
		if g.IsMenuOpenedAnywhere() {
			g.ClosePDA()
		} else {
			g.TransitionToOverworld()
		}
		return nil
	}

	p := g.GetPlayer()
	b := g.GetBaseStation()

	cursor := inp.Cursor()
	mx, my := int(cursor.X), int(cursor.Y)
	leftClicked := inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	const (
		panelW = 800
		panelH = 500
	)
	panelX := float64(config.ScreenWidth-panelW) / 2.0
	panelY := float64(config.ScreenHeight-panelH) / 2.0

	if leftClicked {
		ty := int(panelY) + 40
		if my >= ty && my < ty+30 && !g.IsMenuOpenedAnywhere() {
			for i := 0; i < 5; i++ {
				tx := int(panelX) + 30 + i*150
				if mx >= tx && mx < tx+140 {
					if i == 2 && !b.HasModule(item.ModuleStorage) {
						continue
					}
					m.ActiveTab = i
					m.ScrollY = 0
					m.SelectedLoreIndex = 0
				}
			}
		}
	}

	switch m.ActiveTab {
	case 0:
		if leftClicked {
			hoveredIdx := BaseMenuPlayerInvLayout.HoveredSlot(panelX, panelY, len(p.Inventory.Slots), mx, my)
			if hoveredIdx != -1 {
				slot := &p.Inventory.Slots[hoveredIdx]
				if slot.Item != nil {
					if b.InstallUpgrade(slot.Item) {
						p.Inventory.Remove(slot.Item, 1)
						p.RecalculateUpgrades()
					}
				}
			}

			if b.Upgrades != nil {
				hoveredModuleIdx := BaseMenuInstalledModulesLayout.HoveredSlot(panelX, panelY, len(b.Upgrades.Slots), mx, my)
				if hoveredModuleIdx != -1 {
					slot := &b.Upgrades.Slots[hoveredModuleIdx]
					if slot.Item != nil {
						if b.WouldUninstallOverflow(hoveredModuleIdx) {
							g.SetMineWarning("Vault has too many items to uninstall storage upgrade!", 120, 2)
						} else if p.Inventory.AddItem(item.Clone(slot.Item), 1) {
							b.Upgrades.Remove(slot.Item, 1)
							b.RecalculateProperties()
							p.RecalculateUpgrades()
						}
					}
				}
			}
		}

	case 1:
		_, wy := inp.Wheel()
		unlockedCount := 0
		recipes := g.GetCraftingRecipes()
		for _, rcp := range recipes {
			if rcp.Unlocked {
				unlockedCount++
			}
		}
		if wy != 0 {
			m.ScrollY -= wy * 15
			maxScroll := float64(unlockedCount*58 - 310)
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

			if my >= viewportMinY && my < viewportMaxY {
				visibleIndex := 0
				for _, rcp := range recipes {
					if !rcp.Unlocked {
						continue
					}
					ry := float64(viewportMinY) + float64(visibleIndex*rowH) - m.ScrollY
					btnX := startX + 560
					btnY := int(ry) + 8

					if mx >= btnX && mx < btnX+160 && my >= btnY && my < btnY+35 {
						if b.Power >= 10.0 {
							hasAll := true
							for _, ing := range rcp.Ingredients {
								if !p.Inventory.Has(ing.NewItem(), ing.Quantity) {
									hasAll = false
									break
								}
							}

							if hasAll {
								newItm := rcp.NewResult()
								if _, isRocket := newItm.(*item.EscapeRocket); isRocket {
									for _, ing := range rcp.Ingredients {
										p.Inventory.Remove(ing.NewItem(), ing.Quantity)
									}
									b.Power -= 10.0
									p.RecalculateUpgrades()
									g.TransitionToGameWon()
									return nil
								}
								resQty := rcp.ResultQuantity
								if resQty <= 0 {
									resQty = 1
								}
								if p.Inventory.AddItem(newItm, resQty) {
									for _, ing := range rcp.Ingredients {
										p.Inventory.Remove(ing.NewItem(), ing.Quantity)
									}
									b.Power -= 10.0
									p.RecalculateUpgrades()
								}
							}
						}
					}
					visibleIndex++
				}
			}
		}

	case 2:
		if leftClicked && b.HasModule(item.ModuleStorage) {
			hoveredIdx := BaseVaultPlayerInvLayout.HoveredSlot(panelX, panelY, len(p.Inventory.Slots), mx, my)
			if hoveredIdx != -1 {
				slot := &p.Inventory.Slots[hoveredIdx]
				if slot.Item != nil {
					if b.Storage.AddItem(item.Clone(slot.Item), 1) {
						p.Inventory.Remove(slot.Item, 1)
						p.RecalculateUpgrades()
					}
				}
			}

			vaultLayout := GetBaseVaultStorageLayout(len(b.Storage.Slots))
			hoveredVaultIdx := vaultLayout.HoveredSlot(panelX, panelY, len(b.Storage.Slots), mx, my)
			if hoveredVaultIdx != -1 {
				slot := &b.Storage.Slots[hoveredVaultIdx]
				if slot.Item != nil {
					if p.Inventory.AddItem(item.Clone(slot.Item), 1) {
						b.Storage.Remove(slot.Item, 1)
						p.RecalculateUpgrades()
					}
				}
			}
		}

	case 3:
		if leftClicked {
			healBtnX := int(panelX) + 260
			healBtnY := int(panelY) + 245
			if mx >= healBtnX && mx < healBtnX+280 && my >= healBtnY && my < healBtnY+60 {
				if b.Power >= 15.0 && p.CurrentHealth < p.MaxHealth {
					b.Power -= 15.0
					p.CurrentHealth += 40.0
					if p.CurrentHealth > p.MaxHealth {
						p.CurrentHealth = p.MaxHealth
					}
				}
			}
		}

	case 4:
		unlocked := g.GetStoryManager().GetUnlockedEntries()
		if len(unlocked) > 0 {
			_, wy := inp.Wheel()
			if wy != 0 {
				m.ScrollY -= wy * 15
				maxScroll := float64(len(unlocked)*32 - 300)
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
				listX := int(panelX) + 30
				listY := int(panelY) + 95
				listW := 260
				listH := 335

				if mx >= listX && mx < listX+listW && my >= listY && my < listY+listH {
					viewportMinY := listY + 5
					clickY := float64(my-viewportMinY) + m.ScrollY
					clickedIdx := int(clickY) / 32
					if clickedIdx >= 0 && clickedIdx < len(unlocked) {
						m.SelectedLoreIndex = clickedIdx
					}
				}
			}
		}
	}

	return nil
}

// Draw renders the management base menu tabs and UI controls.
func (m *BaseMenuScene) Draw(g GameContext, screen *ebiten.Image) {
	m.draw(g, screen)
}

func (m *BaseMenuScene) draw(g MenuContext, screen *ebiten.Image) {
	p := g.GetPlayer()
	b := g.GetBaseStation()

	const (
		panelW = 800
		panelH = 500
	)
	panelX := float32(config.ScreenWidth-panelW) / 2.0
	panelY := float32(config.ScreenHeight-panelH) / 2.0

	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{12, 16, 26, 242}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{68, 88, 120, 255}, false)

	ebitenutil.DebugPrintAt(screen, "BASE ANCHOR TERMINAL - LIFE POD 5", int(panelX)+20, int(panelY)+12)
	powerText := fmt.Sprintf("BASE POWER: %.0f/%.0f HP (Recharge: solar panels)", b.Power, b.MaxPower)
	if b.HasModule(item.ModuleSolar) {
		powerText = fmt.Sprintf("BASE POWER: %.0f/%.0f HP (Recharging: +Solar Active)", b.Power, b.MaxPower)
	}
	ebitenutil.DebugPrintAt(screen, powerText, int(panelX)+420, int(panelY)+12)

	tabLabels := []string{"1. OVERVIEW", "2. FABRICATOR", "3. BASE VAULT", "4. MEDICAL", "5. PDA LOGS"}
	if g.IsMenuOpenedAnywhere() {
		tx := panelX + 30
		ty := panelY + 40
		vector.FillRect(screen, tx, ty, 160, 30, color.RGBA{32, 45, 68, 255}, false)
		vector.StrokeRect(screen, tx, ty, 160, 30, 1.0, color.RGBA{95, 125, 165, 255}, false)
		ebitenutil.DebugPrintAt(screen, "★ DETACHED PDA LOGS ★", int(tx)+10, int(ty)+6)
	} else {
		for i := 0; i < 5; i++ {
			tx := panelX + 30 + float32(i*150)
			ty := panelY + 40

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
			ebitenutil.DebugPrintAt(screen, label, int(tx)+12, int(ty)+6)
		}
	}

	vector.StrokeLine(screen, panelX+20, panelY+75, panelX+panelW-20, panelY+75, 1.0, color.RGBA{68, 88, 120, 255}, false)

	switch m.ActiveTab {
	case 0:
		schematicX := panelX + 30
		schematicY := panelY + 95
		vector.FillRect(screen, schematicX, schematicY, 380, 335, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, schematicX, schematicY, 380, 335, 1.0, color.RGBA{48, 62, 85, 255}, false)
		ebitenutil.DebugPrintAt(screen, "PLAYER INVENTORY (CLICK UPGRADE TO INSTALL)", int(schematicX)+15, int(schematicY)+15)
		drawInventoryGrid(g, screen, float64(panelX), float64(panelY), BaseMenuPlayerInvLayout, p.Inventory)

		upgradeX := panelX + 430
		upgradeY := panelY + 95
		vector.FillRect(screen, upgradeX, upgradeY, 340, 335, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, upgradeX, upgradeY, 340, 335, 1.0, color.RGBA{48, 62, 85, 255}, false)
		ebitenutil.DebugPrintAt(screen, "INSTALLED MODULES (CLICK TO UNINSTALL)", int(upgradeX)+15, int(upgradeY)+15)

		cursor := g.GetInput().Cursor()
		mx, my := int(cursor.X), int(cursor.Y)

		for c := 0; c < 4; c++ {
			rect := BaseMenuInstalledModulesLayout.SlotRect(float64(panelX), float64(panelY), c)
			sx := float32(rect.Min.X)
			sy := float32(rect.Min.Y)
			slotSz := float32(BaseMenuInstalledModulesLayout.SlotSz)

			slotBg := color.RGBA{24, 30, 44, 255}
			slotBorder := color.RGBA{54, 68, 92, 255}
			isHovered := BaseMenuInstalledModulesLayout.InSlot(float64(panelX), float64(panelY), c, mx, my)
			if isHovered {
				slotBg = color.RGBA{38, 48, 70, 255}
				slotBorder = color.RGBA{100, 130, 180, 255}
			}

			vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
			vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

			if b.Upgrades != nil && c < len(b.Upgrades.Slots) {
				itemStack := b.Upgrades.Slots[c]
				if itemStack.Item != nil {
					cx := sx + slotSz/2.0
					cy := sy + slotSz/2.0
					itemStack.Item.DrawIcon(screen, cx, cy, slotSz*0.7)
				}
			}
		}

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

	case 1:
		startX := panelX + 30
		startY := panelY + 95
		rowH := float32(58)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FABRICATOR MENU (POWER COST: 10 HP PER CRAFT - CURRENT: %.0f HP)", b.Power), int(startX)+10, int(startY))

		viewportY := startY + 25
		viewportH := float32(310)

		rect := image.Rect(int(startX), int(viewportY), int(startX)+740, int(viewportY+viewportH))
		subImg := screen.SubImage(rect)
		if subImg != nil {
			clippedScreen := subImg.(*ebiten.Image)
			visibleIndex := 0
			recipes := g.GetCraftingRecipes()
			for _, rcp := range recipes {
				if !rcp.Unlocked {
					continue
				}
				ry := viewportY + float32(visibleIndex)*rowH - float32(m.ScrollY)
				vector.FillRect(clippedScreen, float32(startX), ry, 740, 52, color.RGBA{18, 24, 38, 255}, false)
				vector.StrokeRect(clippedScreen, float32(startX), ry, 740, 52, 0.8, color.RGBA{45, 58, 78, 255}, false)

				resultName := rcp.NewResult().GetName()
				resQty := rcp.ResultQuantity
				if resQty > 1 {
					resultName = fmt.Sprintf("%s x%d", resultName, resQty)
				}
				ebitenutil.DebugPrintAt(clippedScreen, resultName, int(startX)+15, int(ry)+6)

				currentX := int(startX) + 15
				drawColoredDebugText(clippedScreen, "Ingredients:", currentX, int(ry)+28, color.RGBA{180, 190, 200, 255})
				currentX += len("Ingredients:") * 6

				hasAll := true
				for j, ing := range rcp.Ingredients {
					ingredient := ing.NewItem()
					qtyInInv := p.Inventory.Count(ingredient)

					var isEnough bool
					if qtyInInv >= ing.Quantity {
						isEnough = true
					} else {
						hasAll = false
						isEnough = false
					}

					if j > 0 {
						drawColoredDebugText(clippedScreen, " |", currentX, int(ry)+28, color.RGBA{100, 110, 120, 255})
						currentX += len(" |") * 6
					}

					ingStr := fmt.Sprintf(" %s (%d/%d)", ingredient.GetName(), qtyInInv, ing.Quantity)
					var textColor color.RGBA
					if isEnough {
						textColor = color.RGBA{60, 210, 110, 255} // Green
					} else {
						textColor = color.RGBA{240, 80, 80, 255}  // Red
					}

					drawColoredDebugText(clippedScreen, ingStr, currentX, int(ry)+28, textColor)
					currentX += len(ingStr) * 6
				}

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

				visibleIndex++
			}
		}

		unlockedCount := 0
		recipes := g.GetCraftingRecipes()
		for _, rcp := range recipes {
			if rcp.Unlocked {
				unlockedCount++
			}
		}
		totalH := float32(unlockedCount * 58)
		if totalH > viewportH {
			scrollBarX := startX + 740 + 6
			vector.FillRect(screen, float32(scrollBarX), viewportY, 6, viewportH, color.RGBA{24, 30, 44, 128}, false)
			handleH := viewportH * (viewportH / totalH)
			if handleH < 15 {
				handleH = 15
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

	case 2:
		pStartX := panelX + 30
		pStartY := panelY + 110
		ebitenutil.DebugPrintAt(screen, "PLAYER INVENTORY (CLICK TO STORE)", int(pStartX), int(pStartY)-25)
		drawInventoryGrid(g, screen, float64(panelX), float64(panelY), BaseVaultPlayerInvLayout, p.Inventory)

		bStartX := panelX + 430
		bStartY := panelY + 110
		ebitenutil.DebugPrintAt(screen, "BASE VAULT (CLICK TO TAKE)", int(bStartX), int(bStartY)-25)
		vaultLayout := GetBaseVaultStorageLayout(len(b.Storage.Slots))
		drawInventoryGrid(g, screen, float64(panelX), float64(panelY), vaultLayout, b.Storage)

		arrowX := panelX + 395
		arrowY := panelY + 180
		ebitenutil.DebugPrintAt(screen, "<-->", int(arrowX), int(arrowY))

	case 3:
		medX := panelX + 30
		medY := panelY + 95
		vector.FillRect(screen, medX, medY, 740, 335, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, medX, medY, 740, 335, 1.0, color.RGBA{48, 62, 85, 255}, false)
		ebitenutil.DebugPrintAt(screen, "INFIRMARY / MEDICAL SCANNER UNIT", int(medX)+220, int(medY)+40)

		statusHp := fmt.Sprintf("DIVER HEALTH STATUS: %.0f / %.0f HP", p.CurrentHealth, p.MaxHealth)
		ebitenutil.DebugPrintAt(screen, statusHp, int(medX)+240, int(medY)+100)

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

	case 4:
		// Left Panel: Entry List
		listX := panelX + 30
		listY := panelY + 95
		listW := float32(260)
		listH := float32(335)
		vector.FillRect(screen, listX, listY, listW, listH, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, listX, listY, listW, listH, 1.0, color.RGBA{48, 62, 85, 255}, false)
		ebitenutil.DebugPrintAt(screen, "DECRYPTED LOGS LIST", int(listX)+15, int(listY)-20)

		// Right Panel: Detail View
		rightX := panelX + 310
		rightY := panelY + 95
		rightW := float32(460)
		rightH := float32(335)
		vector.FillRect(screen, rightX, rightY, rightW, rightH, color.RGBA{16, 22, 34, 255}, false)
		vector.StrokeRect(screen, rightX, rightY, rightW, rightH, 1.0, color.RGBA{48, 62, 85, 255}, false)

		unlocked := g.GetStoryManager().GetUnlockedEntries()
		if len(unlocked) == 0 {
			ebitenutil.DebugPrintAt(screen, "No database entries decrypted.", int(listX)+15, int(listY)+40)
			ebitenutil.DebugPrintAt(screen, "Harvest resources or catch wildlife", int(listX)+15, int(listY)+60)
			ebitenutil.DebugPrintAt(screen, "to decrypt logs automatically.", int(listX)+15, int(listY)+80)

			ebitenutil.DebugPrintAt(screen, "★ PDA LOG SYSTEM OFFLINE ★", int(rightX)+110, int(rightY)+40)
			ebitenutil.DebugPrintAt(screen, "Data packets are locked behind active", int(rightX)+50, int(rightY)+100)
			ebitenutil.DebugPrintAt(screen, "environmental telemetry grids. Explore,", int(rightX)+50, int(rightY)+120)
			ebitenutil.DebugPrintAt(screen, "mine, and recover items to unlock logs.", int(rightX)+50, int(rightY)+140)
		} else {
			if m.SelectedLoreIndex >= len(unlocked) {
				m.SelectedLoreIndex = len(unlocked) - 1
			}
			if m.SelectedLoreIndex < 0 {
				m.SelectedLoreIndex = 0
			}

			// Render left panel entries
			rowH := float32(32)
			viewportMinY := listY + 5
			viewportH := listH - 10

			rect := image.Rect(int(listX), int(viewportMinY), int(listX+listW), int(viewportMinY+viewportH))
			subImg := screen.SubImage(rect)
			if subImg != nil {
				clippedScreen := subImg.(*ebiten.Image)
				for idx, entry := range unlocked {
					ry := viewportMinY + float32(idx)*rowH - float32(m.ScrollY)
					
					bgClr := color.RGBA{22, 28, 42, 255}
					borderClr := color.RGBA{42, 54, 76, 255}
					if idx == m.SelectedLoreIndex {
						bgClr = color.RGBA{38, 54, 86, 255}
						borderClr = color.RGBA{100, 140, 200, 255}
					}
					
					vector.FillRect(clippedScreen, listX+5, ry, listW-10, 28, bgClr, false)
					vector.StrokeRect(clippedScreen, listX+5, ry, listW-10, 28, 0.8, borderClr, false)

					titleText := fmt.Sprintf("[%s] %s", entry.Category, entry.Title)
					if len(titleText) > 28 {
						titleText = titleText[:25] + "..."
					}
					drawColoredDebugText(clippedScreen, titleText, int(listX)+12, int(ry)+6, color.RGBA{220, 220, 220, 255})
				}
			}

			// Render scrollbar for left panel if needed
			totalH := float32(len(unlocked) * 32)
			if totalH > viewportH {
				scrollBarX := listX + listW - 8
				vector.FillRect(screen, scrollBarX, viewportMinY, 4, viewportH, color.RGBA{24, 30, 44, 128}, false)
				handleH := viewportH * (viewportH / totalH)
				if handleH < 15 {
					handleH = 15
				}
				maxScroll := totalH - viewportH
				var handleY float32
				if maxScroll > 0 {
					handleY = viewportMinY + (float32(m.ScrollY)/maxScroll)*(viewportH-handleH)
				} else {
					handleY = viewportMinY
				}
				vector.FillRect(screen, scrollBarX, handleY, 4, handleH, color.RGBA{100, 130, 180, 255}, false)
			}

			// Render right panel selected entry
			entry := unlocked[m.SelectedLoreIndex]
			drawColoredDebugText(screen, entry.Title, int(rightX)+15, int(rightY)+15, color.RGBA{220, 180, 50, 255})
			drawColoredDebugText(screen, "Category: "+entry.Category, int(rightX)+15, int(rightY)+35, color.RGBA{100, 180, 220, 255})
			vector.StrokeLine(screen, rightX+15, rightY+55, rightX+rightW-15, rightY+55, 0.8, color.RGBA{48, 62, 85, 255}, false)

			textStartY := rightY + 65
			for _, pGraph := range entry.Paragraphs {
				drawColoredDebugText(screen, pGraph.Header, int(rightX)+15, int(textStartY), color.RGBA{140, 170, 220, 255})
				textStartY += 18

				wrappedLines := wrapText(pGraph.Text, 62)
				for _, line := range wrappedLines {
					drawColoredDebugText(screen, line, int(rightX)+15, int(textStartY), color.RGBA{200, 200, 200, 255})
					textStartY += 16
				}
				textStartY += 12 // margin between paragraphs
			}
		}
	}

	closeText := "Press [E] or [O] to Close Terminal Interface"
	if g.IsMenuOpenedAnywhere() {
		closeText = "Press [J], [E], or [O] to Close PDA logs"
	}
	ebitenutil.DebugPrintAt(screen, closeText, int(panelX)+260, int(panelY)+panelH-25)
}

// wrapText splits a single long string of text into multiple lines of maxChars length.
func wrapText(text string, maxChars int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) > maxChars {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

func drawInventoryGrid(g MenuContext, screen *ebiten.Image, panelX, panelY float64, d LayoutDescriptor, inv *item.Inventory) {
	cursor := g.GetInput().Cursor()
	mx, my := int(cursor.X), int(cursor.Y)

	for idx := 0; idx < len(inv.Slots); idx++ {
		rect := d.SlotRect(panelX, panelY, idx)
		sx := float32(rect.Min.X)
		sy := float32(rect.Min.Y)
		slotSz := float32(d.SlotSz)

		slotBg := color.RGBA{24, 30, 44, 255}
		slotBorder := color.RGBA{54, 68, 92, 255}
		isHovered := d.InSlot(panelX, panelY, idx, mx, my)
		if isHovered {
			slotBg = color.RGBA{38, 48, 70, 255}
			slotBorder = color.RGBA{100, 130, 180, 255}
		}

		vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
		vector.StrokeRect(screen, sx, sy, slotSz, slotSz, 1.0, slotBorder, false)

		itemStack := inv.Slots[idx]
		if itemStack.Item != nil {
			cx := sx + slotSz/2.0
			cy := sy + slotSz/2.0
			itemStack.Item.DrawIcon(screen, cx, cy, slotSz*0.7)
			if itemStack.Quantity > 1 {
				ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", itemStack.Quantity), int(sx)+4, int(sy)+int(slotSz)-15)
			}
		}
	}
}

var textTempImage *ebiten.Image

func drawColoredDebugText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	w := len(str) * 6
	if w <= 0 {
		return
	}
	if textTempImage == nil || textTempImage.Bounds().Dx() < w {
		textTempImage = ebiten.NewImage(w+100, 20)
	}

	subRect := image.Rect(0, 0, w, 16)
	textTempImage.SubImage(subRect).(*ebiten.Image).Clear()

	ebitenutil.DebugPrintAt(textTempImage, str, 0, 0)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)

	screen.DrawImage(textTempImage.SubImage(subRect).(*ebiten.Image), op)
}
