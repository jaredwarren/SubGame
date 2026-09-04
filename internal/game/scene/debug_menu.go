package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/world"
)

// DebugContext defines the interface for interacting with game systems from the debug menu.
type DebugContext interface {
	GetInput() InputSource
	GetPlayer() *player.Player
	GetTicks() float64
	GetCurrentState() State
	GetBaseStation() *base.BaseStation
	GetWorld() *world.World
	GetTimeOfDay() float64
	GetExploration() *exploration.Tracker
	GetActiveVehicle() vehicle.Vehicle

	// Item / Inventory
	GiveItem(name string, qty int)
	GivePreset(presetName string)
	ClearPlayerInventory()
	ClearPlayerHotbar()

	// Cheats State & Toggles
	IsGodMode() bool
	ToggleGodMode()
	IsInfiniteO2() bool
	ToggleInfiniteO2()
	IsInfiniteStamina() bool
	ToggleInfiniteStamina()
	IsSuperSpeed() bool
	ToggleSuperSpeed()
	IsTimeFrozen() bool
	ToggleFreezeTime()
	IsInfiniteVehicleBattery() bool
	ToggleInfiniteVehicleBattery()
	IsInfiniteVehicleHull() bool
	ToggleInfiniteVehicleHull()

	// Player Actions
	HealPlayerFull()
	RefillO2AndStamina()
	KillPlayer()
	TriggerWin()

	// Vehicle Actions
	SpawnVehicle(name string)
	RepairActiveVehicle()
	ChargeActiveVehicle()
	DespawnActiveVehicle()

	// World & Teleport
	SetTimeOfDay(tod float64)
	AdvanceTimeOfDay(hours float64)
	TeleportToLifePod()
	TeleportToPOI(poiType world.TileType)
	TeleportToVoid()
	DirectDiveCave(caveType string)
	SurfaceToOverworld()
	RevealFullMap()
	ResetFogOfWar()

	// Progression
	UnlockAllRecipes()
	UnlockAllLore()
	CompleteCurrentTask()
	CompleteCurrentQuest()
	CompleteAllQuests()
	ResetAllQuests()

	// Status Feedback & Control
	SetMineWarning(msg string, duration, level int)
	CloseDebugMenu()
}

// DebugMenuScene manages the debug testing interface overlay.
type DebugMenuScene struct {
	ActiveTab     int // 0: Items, 1: Cheats, 2: Vehicles, 3: World, 4: Quests
	ItemCategory  int // 0: Minerals, 1: Tools, 2: Upgrades, 3: Vehicles, 4: Food, 5: Base Modules
	SpawnQuantity int // 1, 5, or -1 (max stack)
}

// NewDebugMenuScene creates an instance of the debug menu.
func NewDebugMenuScene() *DebugMenuScene {
	return &DebugMenuScene{
		ActiveTab:     0,
		ItemCategory:  0,
		SpawnQuantity: 1,
	}
}

type debugItemEntry struct {
	Name     string
	Category int // 0: Minerals, 1: Tools, 2: Upgrades, 3: Vehicles, 4: Food, 5: Modules
}

var debugItemList = []debugItemEntry{
	// 0: Minerals & Raw
	{Name: "Titanium", Category: 0},
	{Name: "Copper", Category: 0},
	{Name: "Quartz", Category: 0},
	{Name: "Nickel", Category: 0},
	{Name: "Tungsten", Category: 0},
	{Name: "Abyssal Ore", Category: 0},
	{Name: "Scrap Metal", Category: 0},
	{Name: "Electronic Waste", Category: 0},

	// 1: Tools & Equipment
	{Name: "Scanner", Category: 1},
	{Name: "Flashlight", Category: 1},
	{Name: "Repair Tool", Category: 1},
	{Name: "Propulsion Fins", Category: 1},
	{Name: "High Capacity O2 Tank", Category: 1},
	{Name: "Ultra High Capacity O2 Tank", Category: 1},
	{Name: "Escape Rocket", Category: 1},

	// 2: Vehicle Upgrades & Usables
	{Name: "Sonar Amplifier", Category: 2},
	{Name: "Surface Sonar Module", Category: 2},
	{Name: "Skiff Light Module", Category: 2},
	{Name: "Decoy Launcher", Category: 2},
	{Name: "Chemical Discharger", Category: 2},
	{Name: "Thermal Generator", Category: 2},
	{Name: "Power Cell", Category: 2},
	{Name: "Sonic Decoy", Category: 2},
	{Name: "Chemical Deterrent", Category: 2},

	// 3: Vehicle Kits
	{Name: "Skiff Kit", Category: 3},
	{Name: "Scout Sub Kit", Category: 3},
	{Name: "Heavy Mech Kit", Category: 3},
	{Name: "Mini-Lifepod Kit", Category: 3},

	// 4: Food & Consumables
	{Name: "Cooked Fish", Category: 4},
	{Name: "Raw Fish", Category: 4},
	{Name: "Cooked Crab", Category: 4},
	{Name: "Raw Crab", Category: 4},

	// 5: Base Modules
	{Name: "Solar Array Module", Category: 5},
	{Name: "Solar Array MKII Module", Category: 5},
	{Name: "Storage Vault Module", Category: 5},
	{Name: "Storage Vault MKII Module", Category: 5},
}

var debugCategoryNames = []string{
	"Minerals", "Equipment", "Upgrades", "Vehicles", "Food", "Base Modules",
}

// Update handles mouse interactions and key navigation for the debug menu.
func (d *DebugMenuScene) Update(g DebugContext) error {
	inp := g.GetInput()
	ctrlPressed := inp.IsKeyPressed(ebiten.KeyControl) || inp.IsKeyPressed(ebiten.KeyMeta)
	if inp.IsKeyJustPressed(ebiten.KeyGraveAccent) || inp.IsKeyJustPressed(ebiten.KeyF1) || inp.IsKeyJustPressed(ebiten.KeyF3) || inp.IsKeyJustPressed(ebiten.KeyF12) || (ctrlPressed && inp.IsKeyJustPressed(ebiten.KeyD)) || inp.IsKeyJustPressed(ebiten.KeyEscape) {
		g.CloseDebugMenu()
		return nil
	}

	if !inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return nil
	}

	cursor := inp.Cursor()
	mx, my := int(cursor.X), int(cursor.Y)

	const (
		panelW = 860.0
		panelH = 480.0
	)
	panelX := float64(config.ScreenWidth-int(panelW)) / 2.0
	panelY := float64(config.ScreenHeight-int(panelH)) / 2.0

	// Close Button [X] at top-right
	if mx >= int(panelX+panelW-35) && mx <= int(panelX+panelW-10) && my >= int(panelY+10) && my <= int(panelY+30) {
		audio.Get().PlaySFX("sfx/ui_click.wav")
		g.CloseDebugMenu()
		return nil
	}

	// Main Tab Bar
	tabNames := []string{"ITEMS & SPAWN", "PLAYER CHEATS", "VEHICLES", "WORLD / WARP", "QUESTS & LORE"}
	for i := range tabNames {
		tx := int(panelX) + 15 + i*165
		ty := int(panelY) + 38
		if mx >= tx && mx <= tx+155 && my >= ty && my <= ty+26 {
			if d.ActiveTab != i {
				audio.Get().PlaySFX("sfx/ui_hover.wav")
				d.ActiveTab = i
			}
			return nil
		}
	}

	contentX := panelX + 15
	contentY := panelY + 75

	switch d.ActiveTab {
	case 0:
		d.updateItemsTab(g, contentX, contentY, mx, my)
	case 1:
		d.updateCheatsTab(g, contentX, contentY, mx, my)
	case 2:
		d.updateVehiclesTab(g, contentX, contentY, mx, my)
	case 3:
		d.updateWorldTab(g, contentX, contentY, mx, my)
	case 4:
		d.updateQuestsTab(g, contentX, contentY, mx, my)
	}

	return nil
}

func (d *DebugMenuScene) updateItemsTab(g DebugContext, cx, cy float64, mx, my int) {
	// Category Pills
	for i := range debugCategoryNames {
		px := int(cx) + i*135
		py := int(cy)
		if mx >= px && mx <= px+125 && my >= py && my <= py+22 {
			audio.Get().PlaySFX("sfx/ui_hover.wav")
			d.ItemCategory = i
			return
		}
	}

	// Quantity Selector Buttons (+1, +5, +MAX)
	qtyX := int(cx) + 580
	qtyY := int(cy)
	quantities := []int{1, 5, -1}
	qtyLabels := []string{"+1", "+5", "+MAX"}
	for i, q := range quantities {
		bx := qtyX + i*75
		if mx >= bx && mx <= bx+65 && my >= qtyY && my <= qtyY+22 {
			audio.Get().PlaySFX("sfx/ui_hover.wav")
			d.SpawnQuantity = q
			return
		}
		_ = qtyLabels[i]
	}

	// Items Grid for Current Category
	var itemsInCat []debugItemEntry
	for _, it := range debugItemList {
		if it.Category == d.ItemCategory {
			itemsInCat = append(itemsInCat, it)
		}
	}

	gridY := cy + 32.0
	for i, it := range itemsInCat {
		row := i / 4
		col := i % 4
		bx := int(cx) + col*205
		by := int(gridY) + row*42

		if mx >= bx && mx <= bx+195 && my >= by && my <= by+36 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			qty := d.SpawnQuantity
			if qty == -1 {
				dummy := item.NewItemByName(it.Name)
				if dummy != nil {
					qty = dummy.GetMaxStack()
				} else {
					qty = 1
				}
			}
			g.GiveItem(it.Name, qty)
			return
		}
	}

	// Presets Row at Bottom
	presetY := cy + 340.0
	presets := []struct {
		Label string
		Name  string
	}{
		{"[ Starter Kit ]", "starter"},
		{"[ All Tools ]", "tools"},
		{"[ All Minerals x10 ]", "minerals"},
		{"[ All Upgrades ]", "upgrades"},
		{"[ Rocket Parts ]", "rocket"},
		{"[ Clear Inv ]", "clear_inv"},
		{"[ Clear Hotbar ]", "clear_hotbar"},
	}

	for i, p := range presets {
		bx := int(cx) + i*118
		by := int(presetY)
		if mx >= bx && mx <= bx+112 && my >= by && my <= by+26 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			switch p.Name {
			case "clear_inv":
				g.ClearPlayerInventory()
			case "clear_hotbar":
				g.ClearPlayerHotbar()
			default:
				g.GivePreset(p.Name)
			}
			return
		}
	}
}

func (d *DebugMenuScene) updateCheatsTab(g DebugContext, cx, cy float64, mx, my int) {
	// Cheats Toggles
	toggles := []struct {
		Label  string
		Active bool
		Action func()
	}{
		{"God Mode (Invulnerable)", g.IsGodMode(), g.ToggleGodMode},
		{"Infinite Oxygen (100% Lock)", g.IsInfiniteO2(), g.ToggleInfiniteO2},
		{"Infinite Stamina (No Fatigue)", g.IsInfiniteStamina(), g.ToggleInfiniteStamina},
		{"Super Speed (2.5x Velocity)", g.IsSuperSpeed(), g.ToggleSuperSpeed},
		{"Freeze Time of Day", g.IsTimeFrozen(), g.ToggleFreezeTime},
		{"Flashlight: Follow Mouse", config.FlashlightFollowsMouse, func() { config.FlashlightFollowsMouse = !config.FlashlightFollowsMouse }},
	}

	for i, t := range toggles {
		bx := int(cx) + 20
		by := int(cy) + 15 + i*42
		if mx >= bx && mx <= bx+350 && my >= by && my <= by+32 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			t.Action()
			return
		}
	}

	// Immediate Action Buttons
	actions := []struct {
		Label  string
		Action func()
	}{
		{"Heal Player to Full HP", g.HealPlayerFull},
		{"Refill Oxygen & Stamina", g.RefillO2AndStamina},
		{"Kill Player (Test Death Beacon)", g.KillPlayer},
		{"Trigger Rocket Victory", g.TriggerWin},
	}

	for i, act := range actions {
		bx := int(cx) + 420
		by := int(cy) + 15 + i*42
		if mx >= bx && mx <= bx+350 && my >= by && my <= by+32 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			act.Action()
			return
		}
	}
}

func (d *DebugMenuScene) updateVehiclesTab(g DebugContext, cx, cy float64, mx, my int) {
	// Spawn Vehicles
	spawns := []struct {
		Label string
		Name  string
	}{
		{"Spawn Surface Skiff", "The Skiff"},
		{"Spawn Scout Submarine", "Scout Sub"},
		{"Spawn Heavy Mech Walker", "Heavy Mech"},
		{"Spawn Mini-Lifepod", "Mini-Lifepod"},
	}

	for i, sp := range spawns {
		bx := int(cx) + 20
		by := int(cy) + 15 + i*42
		if mx >= bx && mx <= bx+350 && my >= by && my <= by+32 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			g.SpawnVehicle(sp.Name)
			return
		}
	}

	// Active Vehicle Modifiers
	vActions := []struct {
		Label  string
		Action func()
	}{
		{"Repair Active Hull to 100%", g.RepairActiveVehicle},
		{"Recharge Battery to 100%", g.ChargeActiveVehicle},
		{"Despawn Active Vehicle", g.DespawnActiveVehicle},
	}

	for i, act := range vActions {
		bx := int(cx) + 420
		by := int(cy) + 15 + i*42
		if mx >= bx && mx <= bx+350 && my >= by && my <= by+32 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			act.Action()
			return
		}
	}

	// Vehicle God Toggles
	vToggles := []struct {
		Label  string
		Active bool
		Action func()
	}{
		{"Infinite Vehicle Battery", g.IsInfiniteVehicleBattery(), g.ToggleInfiniteVehicleBattery},
		{"Infinite Vehicle Hull (God Mode)", g.IsInfiniteVehicleHull(), g.ToggleInfiniteVehicleHull},
	}

	for i, t := range vToggles {
		bx := int(cx) + 20
		by := int(cy) + 160 + i*42
		if mx >= bx && mx <= bx+350 && my >= by && my <= by+32 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			t.Action()
			return
		}
	}
}

func (d *DebugMenuScene) updateWorldTab(g DebugContext, cx, cy float64, mx, my int) {
	// Time of Day
	times := []struct {
		Label string
		TOD   float64
	}{
		{"Dawn (06:00)", 600},
		{"Noon (12:00)", 5400},
		{"Dusk (18:00)", 10200},
		{"Midnight (00:00)", 12600},
	}

	for i, t := range times {
		bx := int(cx) + 20 + i*195
		by := int(cy) + 15
		if mx >= bx && mx <= bx+185 && my >= by && my <= by+28 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			g.SetTimeOfDay(t.TOD)
			return
		}
	}

	// Advance Time & Fog of War
	fogActions := []struct {
		Label  string
		Action func()
	}{
		{"Advance Time +1h", func() { g.AdvanceTimeOfDay(1.0) }},
		{"Reveal Full Overworld Map", g.RevealFullMap},
		{"Reset Fog of War", g.ResetFogOfWar},
	}

	for i, act := range fogActions {
		bx := int(cx) + 20 + i*260
		by := int(cy) + 55
		if mx >= bx && mx <= bx+245 && my >= by && my <= by+28 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			act.Action()
			return
		}
	}

	// Teleport to Overworld POIs
	warps := []struct {
		Label  string
		Action func()
	}{
		{"Warp to Life Pod Base", g.TeleportToLifePod},
		{"Warp to Nearest Trench", func() { g.TeleportToPOI(world.TileTrench) }},
		{"Warp to Nearest Wreckage", func() { g.TeleportToPOI(world.TileWreckage) }},
		{"Warp to Shock Kelp Cave", func() { g.TeleportToPOI(world.TileShockKelpCave) }},
		{"Warp to Thermal Cave / Vent", func() { g.TeleportToPOI(world.TileThermoCave) }},
		{"Warp to Void Border", g.TeleportToVoid},
	}

	for i, w := range warps {
		col := i % 2
		row := i / 2
		bx := int(cx) + 20 + col*400
		by := int(cy) + 105 + row*38
		if mx >= bx && mx <= bx+380 && my >= by && my <= by+30 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			w.Action()
			return
		}
	}

	// Direct Cave Diving
	caveWarps := []struct {
		Label  string
		Action func()
	}{
		{"Direct Dive: Shallow Seabed", func() { g.DirectDiveCave("shallow") }},
		{"Direct Dive: Deep Trench", func() { g.DirectDiveCave("trench") }},
		{"Direct Dive: Shock Kelp", func() { g.DirectDiveCave("kelp") }},
		{"Direct Dive: Thermal Caverns", func() { g.DirectDiveCave("thermo") }},
		{"Surface to Overworld", g.SurfaceToOverworld},
	}

	for i, cw := range caveWarps {
		bx := int(cx) + 20 + i*155
		by := int(cy) + 245
		if mx >= bx && mx <= bx+145 && my >= by && my <= by+32 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			cw.Action()
			return
		}
	}
}

func (d *DebugMenuScene) updateQuestsTab(g DebugContext, cx, cy float64, mx, my int) {
	qActions := []struct {
		Label  string
		Action func()
	}{
		{"Unlock All Crafting Blueprints", g.UnlockAllRecipes},
		{"Unlock All 25 PDA Lore Logs", g.UnlockAllLore},
		{"Complete Current Objective Task", g.CompleteCurrentTask},
		{"Complete Entire Current Quest", g.CompleteCurrentQuest},
		{"Complete ALL Quests (Endgame State)", g.CompleteAllQuests},
		{"Reset All Quests to Beginning", g.ResetAllQuests},
	}

	for i, act := range qActions {
		col := i % 2
		row := i / 2
		bx := int(cx) + 20 + col*400
		by := int(cy) + 20 + row*55
		if mx >= bx && mx <= bx+380 && my >= by && my <= by+42 {
			audio.Get().PlaySFX("sfx/ui_click.wav")
			act.Action()
			return
		}
	}
}

// Draw renders the debug console overlay.
func (d *DebugMenuScene) Draw(screen *ebiten.Image, g DebugContext) {
	const (
		panelW = 860.0
		panelH = 480.0
	)
	panelX := float32(config.ScreenWidth-int(panelW)) / 2.0
	panelY := float32(config.ScreenHeight-int(panelH)) / 2.0

	// Dim backdrop
	vector.FillRect(screen, 0, 0, float32(config.ScreenWidth), float32(config.ScreenHeight), color.RGBA{0, 4, 10, 160}, false)

	// Main Panel Box
	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{10, 15, 24, 250}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{0, 220, 255, 200}, false)

	// Header Bar
	vector.FillRect(screen, panelX+2, panelY+2, panelW-4, 30, color.RGBA{14, 25, 42, 255}, false)
	drawColoredDebugText(screen, "✦ SUBGAME DEBUG & TESTING CONSOLE ✦  [~ / F1 / F3 / Ctrl+D / ESC to close]", int(panelX)+16, int(panelY)+8, color.RGBA{0, 240, 255, 255})

	// Close Button [X]
	closeX := panelX + panelW - 30
	closeY := panelY + 6
	vector.FillRect(screen, closeX, closeY, 20, 18, color.RGBA{80, 20, 30, 255}, false)
	vector.StrokeRect(screen, closeX, closeY, 20, 18, 1.0, color.RGBA{240, 80, 80, 255}, false)
	drawColoredDebugText(screen, "X", int(closeX)+6, int(closeY)+2, color.RGBA{255, 255, 255, 255})

	// Main Tab Bar
	tabNames := []string{"ITEMS & SPAWN", "PLAYER CHEATS", "VEHICLES", "WORLD / WARP", "QUESTS & LORE"}
	for i, name := range tabNames {
		tx := panelX + 15 + float32(i*165)
		ty := panelY + 38
		tabBg := color.RGBA{18, 26, 38, 255}
		tabBorder := color.RGBA{45, 60, 80, 255}
		tabTextClr := color.RGBA{160, 180, 200, 255}
		if d.ActiveTab == i {
			tabBg = color.RGBA{0, 75, 120, 255}
			tabBorder = color.RGBA{0, 220, 255, 255}
			tabTextClr = color.RGBA{255, 255, 255, 255}
		}
		vector.FillRect(screen, tx, ty, 155, 26, tabBg, false)
		vector.StrokeRect(screen, tx, ty, 155, 26, 1.0, tabBorder, false)
		drawColoredDebugText(screen, name, int(tx)+12, int(ty)+6, tabTextClr)
	}

	contentX := panelX + 15
	contentY := panelY + 75

	switch d.ActiveTab {
	case 0:
		d.drawItemsTab(screen, g, contentX, contentY)
	case 1:
		d.drawCheatsTab(screen, g, contentX, contentY)
	case 2:
		d.drawVehiclesTab(screen, g, contentX, contentY)
	case 3:
		d.drawWorldTab(screen, g, contentX, contentY)
	case 4:
		d.drawQuestsTab(screen, g, contentX, contentY)
	}
}

func (d *DebugMenuScene) drawItemsTab(screen *ebiten.Image, g DebugContext, cx, cy float32) {
	// Category Pills
	for i, catName := range debugCategoryNames {
		px := cx + float32(i*135)
		py := cy
		bg := color.RGBA{22, 32, 48, 255}
		border := color.RGBA{50, 70, 95, 255}
		txtClr := color.RGBA{180, 200, 220, 255}
		if d.ItemCategory == i {
			bg = color.RGBA{0, 120, 180, 255}
			border = color.RGBA{0, 240, 255, 255}
			txtClr = color.RGBA{255, 255, 255, 255}
		}
		vector.FillRect(screen, px, py, 125, 22, bg, false)
		vector.StrokeRect(screen, px, py, 125, 22, 1.0, border, false)
		drawColoredDebugText(screen, catName, int(px)+10, int(py)+4, txtClr)
	}

	// Quantity Selectors
	qtyX := cx + 580
	qtyY := cy
	quantities := []int{1, 5, -1}
	qtyLabels := []string{"+1", "+5", "+MAX"}
	for i, q := range quantities {
		bx := qtyX + float32(i*75)
		bg := color.RGBA{22, 32, 48, 255}
		border := color.RGBA{50, 70, 95, 255}
		txtClr := color.RGBA{180, 200, 220, 255}
		if d.SpawnQuantity == q {
			bg = color.RGBA{200, 110, 30, 255}
			border = color.RGBA{255, 170, 60, 255}
			txtClr = color.RGBA{255, 255, 255, 255}
		}
		vector.FillRect(screen, bx, qtyY, 65, 22, bg, false)
		vector.StrokeRect(screen, bx, qtyY, 65, 22, 1.0, border, false)
		drawColoredDebugText(screen, qtyLabels[i], int(bx)+14, int(qtyY)+4, txtClr)
	}

	// Item Grid
	var itemsInCat []debugItemEntry
	for _, it := range debugItemList {
		if it.Category == d.ItemCategory {
			itemsInCat = append(itemsInCat, it)
		}
	}

	gridY := cy + 32.0
	for i, it := range itemsInCat {
		row := i / 4
		col := i % 4
		bx := cx + float32(col*205)
		by := gridY + float32(row*42)

		vector.FillRect(screen, bx, by, 195, 36, color.RGBA{18, 24, 36, 255}, false)
		vector.StrokeRect(screen, bx, by, 195, 36, 1.0, color.RGBA{45, 60, 80, 255}, false)

		// Draw icon preview if available
		dummyItem := item.NewItemByName(it.Name)
		if dummyItem != nil {
			dummyItem.DrawIcon(screen, bx+18, by+18, 14.0)
		}

		drawColoredDebugText(screen, it.Name, int(bx)+38, int(by)+10, color.RGBA{220, 230, 240, 255})
	}

	// Presets Row at Bottom
	presetY := cy + 340.0
	presets := []string{
		"[ Starter Kit ]",
		"[ All Tools ]",
		"[ Minerals x10 ]",
		"[ All Upgrades ]",
		"[ Rocket Parts ]",
		"[ Clear Inv ]",
		"[ Clear Hotbar ]",
	}

	for i, label := range presets {
		bx := cx + float32(i*118)
		by := presetY
		bg := color.RGBA{28, 38, 54, 255}
		border := color.RGBA{60, 80, 110, 255}
		if label == "[ Clear Inv ]" || label == "[ Clear Hotbar ]" {
			bg = color.RGBA{48, 22, 28, 255}
			border = color.RGBA{140, 50, 60, 255}
		}
		vector.FillRect(screen, bx, by, 112, 26, bg, false)
		vector.StrokeRect(screen, bx, by, 112, 26, 1.0, border, false)
		drawColoredDebugText(screen, label, int(bx)+6, int(by)+6, color.RGBA{220, 230, 240, 255})
	}
}

func (d *DebugMenuScene) drawCheatsTab(screen *ebiten.Image, g DebugContext, cx, cy float32) {
	// Cheats Toggles
	toggles := []struct {
		Label  string
		Active bool
	}{
		{"God Mode (Invulnerable)", g.IsGodMode()},
		{"Infinite Oxygen (100% Lock)", g.IsInfiniteO2()},
		{"Infinite Stamina (No Fatigue)", g.IsInfiniteStamina()},
		{"Super Speed (2.5x Velocity)", g.IsSuperSpeed()},
		{"Freeze Time of Day", g.IsTimeFrozen()},
		{"Flashlight: Follow Mouse", config.FlashlightFollowsMouse},
	}

	for i, t := range toggles {
		bx := cx + 20
		by := cy + 15 + float32(i*42)

		bg := color.RGBA{18, 26, 38, 255}
		border := color.RGBA{45, 60, 80, 255}
		badgeBg := color.RGBA{60, 20, 25, 255}
		badgeBorder := color.RGBA{180, 40, 50, 255}
		badgeText := "OFF"
		if t.Active {
			badgeBg = color.RGBA{20, 70, 40, 255}
			badgeBorder = color.RGBA{40, 200, 100, 255}
			badgeText = "ON"
		}

		vector.FillRect(screen, bx, by, 350, 32, bg, false)
		vector.StrokeRect(screen, bx, by, 350, 32, 1.0, border, false)
		drawColoredDebugText(screen, t.Label, int(bx)+12, int(by)+8, color.RGBA{220, 230, 240, 255})

		// Status Badge
		badgeX := bx + 280
		badgeY := by + 6
		vector.FillRect(screen, badgeX, badgeY, 55, 20, badgeBg, false)
		vector.StrokeRect(screen, badgeX, badgeY, 55, 20, 1.0, badgeBorder, false)
		drawColoredDebugText(screen, badgeText, int(badgeX)+18, int(badgeY)+3, color.RGBA{255, 255, 255, 255})
	}

	// Immediate Action Buttons
	actions := []string{
		"Heal Player to Full HP",
		"Refill Oxygen & Stamina",
		"Kill Player (Test Death Beacon)",
		"Trigger Rocket Victory",
	}

	for i, act := range actions {
		bx := cx + 420
		by := cy + 15 + float32(i*42)
		bg := color.RGBA{22, 34, 52, 255}
		border := color.RGBA{50, 80, 120, 255}
		if act == "Kill Player (Test Death Beacon)" {
			bg = color.RGBA{50, 20, 25, 255}
			border = color.RGBA{160, 40, 50, 255}
		} else if act == "Trigger Rocket Victory" {
			bg = color.RGBA{40, 40, 15, 255}
			border = color.RGBA{200, 180, 30, 255}
		}

		vector.FillRect(screen, bx, by, 350, 32, bg, false)
		vector.StrokeRect(screen, bx, by, 350, 32, 1.0, border, false)
		drawColoredDebugText(screen, act, int(bx)+16, int(by)+8, color.RGBA{220, 235, 250, 255})
	}
}

func (d *DebugMenuScene) drawVehiclesTab(screen *ebiten.Image, g DebugContext, cx, cy float32) {
	// Spawns
	spawns := []string{
		"Spawn Surface Skiff",
		"Spawn Scout Submarine",
		"Spawn Heavy Mech Walker",
		"Spawn Mini-Lifepod",
	}
	for i, sp := range spawns {
		bx := cx + 20
		by := cy + 15 + float32(i*42)
		vector.FillRect(screen, bx, by, 350, 32, color.RGBA{22, 34, 52, 255}, false)
		vector.StrokeRect(screen, bx, by, 350, 32, 1.0, color.RGBA{50, 80, 120, 255}, false)
		drawColoredDebugText(screen, sp, int(bx)+16, int(by)+8, color.RGBA{220, 235, 250, 255})
	}

	// Active Vehicle Modifiers
	vActions := []string{
		"Repair Active Hull to 100%",
		"Recharge Battery to 100%",
		"Despawn Active Vehicle",
	}
	for i, act := range vActions {
		bx := cx + 420
		by := cy + 15 + float32(i*42)
		bg := color.RGBA{22, 34, 52, 255}
		border := color.RGBA{50, 80, 120, 255}
		if act == "Despawn Active Vehicle" {
			bg = color.RGBA{50, 20, 25, 255}
			border = color.RGBA{160, 40, 50, 255}
		}
		vector.FillRect(screen, bx, by, 350, 32, bg, false)
		vector.StrokeRect(screen, bx, by, 350, 32, 1.0, border, false)
		drawColoredDebugText(screen, act, int(bx)+16, int(by)+8, color.RGBA{220, 235, 250, 255})
	}

	// Vehicle God Toggles
	vToggles := []struct {
		Label  string
		Active bool
	}{
		{"Infinite Vehicle Battery", g.IsInfiniteVehicleBattery()},
		{"Infinite Vehicle Hull (God Mode)", g.IsInfiniteVehicleHull()},
	}

	for i, t := range vToggles {
		bx := cx + 20
		by := cy + 160 + float32(i*42)

		bg := color.RGBA{18, 26, 38, 255}
		border := color.RGBA{45, 60, 80, 255}
		badgeBg := color.RGBA{60, 20, 25, 255}
		badgeBorder := color.RGBA{180, 40, 50, 255}
		badgeText := "OFF"
		if t.Active {
			badgeBg = color.RGBA{20, 70, 40, 255}
			badgeBorder = color.RGBA{40, 200, 100, 255}
			badgeText = "ON"
		}

		vector.FillRect(screen, bx, by, 350, 32, bg, false)
		vector.StrokeRect(screen, bx, by, 350, 32, 1.0, border, false)
		drawColoredDebugText(screen, t.Label, int(bx)+12, int(by)+8, color.RGBA{220, 230, 240, 255})

		badgeX := bx + 280
		badgeY := by + 6
		vector.FillRect(screen, badgeX, badgeY, 55, 20, badgeBg, false)
		vector.StrokeRect(screen, badgeX, badgeY, 55, 20, 1.0, badgeBorder, false)
		drawColoredDebugText(screen, badgeText, int(badgeX)+18, int(badgeY)+3, color.RGBA{255, 255, 255, 255})
	}
}

func (d *DebugMenuScene) drawWorldTab(screen *ebiten.Image, g DebugContext, cx, cy float32) {
	// Time of Day
	times := []string{"Dawn (06:00)", "Noon (12:00)", "Dusk (18:00)", "Midnight (00:00)"}
	for i, t := range times {
		bx := cx + 20 + float32(i*195)
		by := cy + 15
		vector.FillRect(screen, bx, by, 185, 28, color.RGBA{22, 34, 52, 255}, false)
		vector.StrokeRect(screen, bx, by, 185, 28, 1.0, color.RGBA{50, 80, 120, 255}, false)
		drawColoredDebugText(screen, t, int(bx)+12, int(by)+6, color.RGBA{220, 235, 250, 255})
	}

	// Fog & Advance Time
	fogActions := []string{"Advance Time +1h", "Reveal Full Map", "Reset Fog of War"}
	for i, act := range fogActions {
		bx := cx + 20 + float32(i*260)
		by := cy + 55
		vector.FillRect(screen, bx, by, 245, 28, color.RGBA{22, 34, 52, 255}, false)
		vector.StrokeRect(screen, bx, by, 245, 28, 1.0, color.RGBA{50, 80, 120, 255}, false)
		drawColoredDebugText(screen, act, int(bx)+16, int(by)+6, color.RGBA{220, 235, 250, 255})
	}

	// Teleport POIs
	warps := []string{
		"Warp to Life Pod Base",
		"Warp to Nearest Trench",
		"Warp to Nearest Wreckage",
		"Warp to Shock Kelp Cave",
		"Warp to Thermal Cave / Vent",
		"Warp to Void Border",
	}

	for i, w := range warps {
		col := i % 2
		row := i / 2
		bx := cx + 20 + float32(col*400)
		by := cy + 105 + float32(row*38)
		vector.FillRect(screen, bx, by, 380, 30, color.RGBA{22, 34, 52, 255}, false)
		vector.StrokeRect(screen, bx, by, 380, 30, 1.0, color.RGBA{50, 80, 120, 255}, false)
		drawColoredDebugText(screen, w, int(bx)+16, int(by)+7, color.RGBA{220, 235, 250, 255})
	}

	// Direct Cave Diving
	caveWarps := []string{
		"Dive: Shallow",
		"Dive: Trench",
		"Dive: Kelp",
		"Dive: Thermo",
		"Surface to Top",
	}

	for i, cw := range caveWarps {
		bx := cx + 20 + float32(i*155)
		by := cy + 245
		vector.FillRect(screen, bx, by, 145, 32, color.RGBA{30, 48, 70, 255}, false)
		vector.StrokeRect(screen, bx, by, 145, 32, 1.0, color.RGBA{0, 190, 230, 255}, false)
		drawColoredDebugText(screen, cw, int(bx)+10, int(by)+8, color.RGBA{240, 245, 255, 255})
	}
}

func (d *DebugMenuScene) drawQuestsTab(screen *ebiten.Image, g DebugContext, cx, cy float32) {
	qActions := []string{
		"Unlock All Crafting Blueprints",
		"Unlock All 25 PDA Lore Logs",
		"Complete Current Objective Task",
		"Complete Entire Current Quest",
		"Complete ALL Quests (Endgame State)",
		"Reset All Quests to Beginning",
	}

	for i, act := range qActions {
		col := i % 2
		row := i / 2
		bx := cx + 20 + float32(col*400)
		by := cy + 20 + float32(row*55)
		bg := color.RGBA{22, 34, 52, 255}
		border := color.RGBA{50, 80, 120, 255}
		if act == "Reset All Quests to Beginning" {
			bg = color.RGBA{50, 20, 25, 255}
			border = color.RGBA{160, 40, 50, 255}
		}

		vector.FillRect(screen, bx, by, 380, 42, bg, false)
		vector.StrokeRect(screen, bx, by, 380, 42, 1.0, border, false)
		drawColoredDebugText(screen, act, int(bx)+16, int(by)+13, color.RGBA{220, 235, 250, 255})
	}
}
