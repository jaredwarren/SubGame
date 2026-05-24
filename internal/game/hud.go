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

// Draw renders the stats panel for the player.
func (h *HUD) Draw(screen *ebiten.Image, p *Player) {
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

	// Draw Oxygen Bar
	oRatio := p.CurrentOxygen / p.MaxOxygen
	drawStatBar(screen, hudX+15, hudY+48, w-30, hBar, oRatio, color.RGBA{45, 175, 215, 255}, "O2", p.CurrentOxygen, p.MaxOxygen)

	// Draw Stamina Bar
	sRatio := p.CurrentStamina / p.MaxStamina
	drawStatBar(screen, hudX+15, hudY+81, w-30, hBar, sRatio, color.RGBA{45, 190, 110, 255}, "ST", p.CurrentStamina, p.MaxStamina)
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
	// Vertically center text offsets
	ebitenutil.DebugPrintAt(screen, text, int(x)+8, int(y)+2)
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
				// Establish visual icons/colors
				var itemClr color.Color
				var drawIconType = "circle"

				switch item.Type {
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
				default:
					itemClr = color.RGBA{98, 198, 148, 255} // Upgrades/Tools
					drawIconType = "circle"
				}

				cx := sx + slotSz/2.0
				cy := sy + slotSz/2.0
				const iSz = 14.0

				if drawIconType == "square" {
					vector.DrawFilledRect(screen, cx-iSz/2.0, cy-iSz/2.0, iSz, iSz, itemClr, false)
				} else if drawIconType == "diamond" {
					// Draw diamond visual
					vector.DrawFilledCircle(screen, cx, cy, iSz/2.0, itemClr, false)
					vector.StrokeCircle(screen, cx, cy, iSz/2.0, 0.5, color.RGBA{255, 255, 255, 200}, false)
				} else {
					vector.DrawFilledCircle(screen, cx, cy, iSz/2.0, itemClr, false)
				}

				// Draw stack count number
				if item.Quantity > 1 {
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.Quantity), int(sx)+6, int(sy)+slotSz-17)
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

