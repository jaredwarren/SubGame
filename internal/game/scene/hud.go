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
	"github.com/jaredwarren/SubGame/internal/game/player"
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
	if p.Hotbar == nil {
		return
	}

	const (
		slotSz = 40.0
		gap    = 8.0
		num    = 5
	)
	w := float32(num*(slotSz+gap) - gap)
	// Center horizontally at bottom
	x := (float32(config.ScreenWidth) - w) / 2.0
	y := float32(config.ScreenHeight - 56.0)

	// Draw container background
	vector.FillRect(screen, x-10, y-10, w+20, slotSz+20, color.RGBA{18, 24, 38, 200}, false)
	vector.StrokeRect(screen, x-10, y-10, w+20, slotSz+20, 1.5, color.RGBA{70, 90, 120, 255}, false)

	for i := 0; i < num; i++ {
		sx := x + float32(i)*(slotSz+gap)
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

		vector.FillRect(screen, sx, sy, slotSz, slotSz, slotBg, false)
		vector.StrokeRect(screen, sx, sy, slotSz, slotSz, borderWidth, slotBorder, false)

		slot := p.Hotbar.Slots[i]
		if slot.Item != nil {
			// Draw item icon
			drawItemIcon(screen, sx, sy, slotSz, slot.Item)
			if slot.Quantity > 1 {
				ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", slot.Quantity), int(sx)+4, int(sy)+int(slotSz)-15)
			}
		} else {
			// Virtual Mining Tool (Default outline)
			drawMiningToolOutline(screen, sx, sy, slotSz)
		}

		// Draw slot index label
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", i+1), int(sx)+int(slotSz)/2-3, int(sy)-22)
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
