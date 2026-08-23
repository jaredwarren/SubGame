package scene

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/quest"
)

func (m *BaseMenuScene) updateQuestsTab(g MenuContext, panelX, panelY float64, mx, my int, leftClicked bool) {
	qm := g.GetQuestManager()
	if qm == nil {
		return
	}

	inp := g.GetInput()
	_, wy := inp.Wheel()
	if wy != 0 {
		m.ScrollY -= wy * 18
		if m.ScrollY < 0 {
			m.ScrollY = 0
		}
	}

	if !leftClicked {
		return
	}

	listX := int(panelX) + 30
	listY := int(panelY) + 95
	listW := 270
	listH := 335

	if mx >= listX && mx < listX+listW && my >= listY && my < listY+listH {
		viewportMinY := float64(listY + 5)
		currentY := viewportMinY - m.ScrollY

		for _, cat := range qm.Categories {
			// Category header row
			headerH := 30.0
			if float64(my) >= currentY && float64(my) < currentY+headerH {
				qm.ToggleCategory(cat.ID)
				audio.Get().PlaySFX("sfx/ui_hover.wav")
				return
			}
			currentY += headerH + 4.0

			if !cat.Collapsed {
				for _, q := range cat.Quests {
					questH := 28.0
					if float64(my) >= currentY && float64(my) < currentY+questH {
						m.SelectedQuestID = q.ID
						audio.Get().PlaySFX("sfx/ui_hover.wav")
						return
					}
					currentY += questH + 3.0
				}
			}
			currentY += 4.0 // Margin between categories
		}
	}
}

func (m *BaseMenuScene) drawQuestsTab(g MenuContext, screen *ebiten.Image, panelX, panelY float32) {
	qm := g.GetQuestManager()
	if qm == nil {
		ebitenutil.DebugPrintAt(screen, "Quest telemetry systems offline.", int(panelX)+40, int(panelY)+120)
		return
	}

	// Left Panel: Categories and Quest List
	listX := panelX + 30
	listY := panelY + 95
	listW := float32(270)
	listH := float32(335)
	vector.FillRect(screen, listX, listY, listW, listH, color.RGBA{16, 22, 34, 255}, false)
	vector.StrokeRect(screen, listX, listY, listW, listH, 1.0, color.RGBA{48, 62, 85, 255}, false)
	ebitenutil.DebugPrintAt(screen, "QUESTS & OBJECTIVES", int(listX)+15, int(listY)-20)

	// Right Panel: Selected Quest Checklist Details
	rightX := panelX + 315
	rightY := panelY + 95
	rightW := float32(455)
	rightH := float32(335)
	vector.FillRect(screen, rightX, rightY, rightW, rightH, color.RGBA{16, 22, 34, 255}, false)
	vector.StrokeRect(screen, rightX, rightY, rightW, rightH, 1.0, color.RGBA{48, 62, 85, 255}, false)

	// Render Left Panel (Clipped viewport)
	viewportMinY := listY + 5
	viewportH := listH - 10
	rect := image.Rect(int(listX), int(viewportMinY), int(listX+listW), int(viewportMinY+viewportH))
	subImg := screen.SubImage(rect)

	// Find active quest or fallback to first available
	var activeQuest *quest.Quest
	if m.SelectedQuestID != "" {
		activeQuest = qm.FindQuest(m.SelectedQuestID)
	}
	if activeQuest == nil && len(qm.Categories) > 0 && len(qm.Categories[0].Quests) > 0 {
		activeQuest = qm.Categories[0].Quests[0]
		m.SelectedQuestID = activeQuest.ID
	}

	if subImg != nil {
		clippedScreen := subImg.(*ebiten.Image)
		currentY := viewportMinY - float32(m.ScrollY)

		for _, cat := range qm.Categories {
			comp, tot := cat.CompletedRatio()
			headerText := fmt.Sprintf("[-] %s", cat.Title)
			if cat.Collapsed {
				headerText = fmt.Sprintf("[+] %s", cat.Title)
			}
			headerH := float32(30)

			// Category Header Box
			catBg := color.RGBA{26, 36, 54, 255}
			catBorder := color.RGBA{65, 88, 125, 255}
			vector.FillRect(clippedScreen, listX+5, currentY, listW-10, headerH, catBg, false)
			vector.StrokeRect(clippedScreen, listX+5, currentY, listW-10, headerH, 1.0, catBorder, false)

			drawColoredDebugText(clippedScreen, headerText, int(listX)+12, int(currentY)+8, color.RGBA{220, 200, 100, 255})
			ratioText := fmt.Sprintf("%d/%d", comp, tot)
			drawColoredDebugText(clippedScreen, ratioText, int(listX+listW)-len(ratioText)*6-18, int(currentY)+8, color.RGBA{140, 190, 220, 255})

			currentY += headerH + 4

			if !cat.Collapsed {
				for _, q := range cat.Quests {
					questH := float32(28)
					qBg := color.RGBA{20, 26, 40, 255}
					qBorder := color.RGBA{45, 58, 80, 255}
					qTextColor := color.RGBA{200, 210, 220, 255}

					if q.ID == m.SelectedQuestID {
						qBg = color.RGBA{36, 52, 82, 255}
						qBorder = color.RGBA{90, 130, 190, 255}
						qTextColor = color.RGBA{255, 255, 255, 255}
					}

					vector.FillRect(clippedScreen, listX+12, currentY, listW-24, questH, qBg, false)
					vector.StrokeRect(clippedScreen, listX+12, currentY, listW-24, questH, 0.8, qBorder, false)

					titleLabel := q.Title
					if len(titleLabel) > 22 {
						titleLabel = titleLabel[:20] + "..."
					}
					drawColoredDebugText(clippedScreen, "• "+titleLabel, int(listX)+18, int(currentY)+7, qTextColor)

					statusBadge := fmt.Sprintf("[%d/%d]", q.CompletedCount(), len(q.Tasks))
					badgeColor := color.RGBA{140, 190, 220, 255}
					if q.Completed {
						statusBadge = "[✓]"
						badgeColor = color.RGBA{60, 220, 110, 255}
					}
					drawColoredDebugText(clippedScreen, statusBadge, int(listX+listW)-len(statusBadge)*6-24, int(currentY)+7, badgeColor)

					currentY += questH + 3
				}
			}
			currentY += 4 // Margin between categories
		}
	}

	// Render Right Panel (Quest details & Task checklist)
	if activeQuest == nil {
		ebitenutil.DebugPrintAt(screen, "Select a quest from the left panel.", int(rightX)+20, int(rightY)+40)
		return
	}

	// Quest Title & Category Tag
	drawColoredDebugText(screen, fmt.Sprintf("[%s]", activeQuest.Category), int(rightX)+15, int(rightY)+15, color.RGBA{100, 180, 230, 255})
	drawColoredDebugText(screen, activeQuest.Title, int(rightX)+15+len(activeQuest.Category)*6+24, int(rightY)+15, color.RGBA{240, 200, 60, 255})

	statusLabel := "[IN PROGRESS]"
	statusColor := color.RGBA{80, 190, 240, 255}
	if activeQuest.Completed {
		statusLabel = "[✓ COMPLETED]"
		statusColor = color.RGBA{60, 220, 110, 255}
	}
	drawColoredDebugText(screen, statusLabel, int(rightX+rightW)-len(statusLabel)*6-20, int(rightY)+15, statusColor)

	vector.StrokeLine(screen, rightX+15, rightY+38, rightX+rightW-15, rightY+38, 0.8, color.RGBA{48, 62, 85, 255}, false)

	// Description
	descY := rightY + 46
	wrappedDesc := wrapText(activeQuest.Description, 58)
	for _, line := range wrappedDesc {
		drawColoredDebugText(screen, line, int(rightX)+15, int(descY), color.RGBA{175, 190, 205, 255})
		descY += 16
	}

	descY += 6
	vector.StrokeLine(screen, rightX+15, descY, rightX+rightW-15, descY, 0.8, color.RGBA{48, 62, 85, 255}, false)

	// Tasks Checklist Header
	taskStartY := descY + 12
	drawColoredDebugText(screen, "CHECKLIST OBJECTIVES:", int(rightX)+15, int(taskStartY), color.RGBA{220, 220, 220, 255})
	taskStartY += 22

	for _, t := range activeQuest.Tasks {
		boxBg := color.RGBA{20, 26, 40, 255}
		boxBorder := color.RGBA{45, 58, 80, 255}
		checkMark := "[ ]"
		checkColor := color.RGBA{100, 180, 230, 255}
		textColor := color.RGBA{210, 220, 230, 255}

		if t.Completed {
			boxBg = color.RGBA{20, 36, 28, 255}
			boxBorder = color.RGBA{50, 120, 70, 255}
			checkMark = "[✓]"
			checkColor = color.RGBA{60, 220, 110, 255}
			textColor = color.RGBA{180, 240, 190, 255}
		}

		boxH := float32(36)
		if t.RequiredCount > 1 {
			boxH = float32(46)
		}

		vector.FillRect(screen, rightX+15, taskStartY, rightW-30, boxH, boxBg, false)
		vector.StrokeRect(screen, rightX+15, taskStartY, rightW-30, boxH, 0.8, boxBorder, false)

		// Checkbox indicator
		drawColoredDebugText(screen, checkMark, int(rightX)+25, int(taskStartY)+8, checkColor)
		// Task description
		drawColoredDebugText(screen, t.Description, int(rightX)+54, int(taskStartY)+8, textColor)

		// Progress meter if multi-count
		if t.RequiredCount > 1 {
			meterX := rightX + 54
			meterY := taskStartY + 26
			meterW := float32(180)
			meterH := float32(10)

			vector.FillRect(screen, meterX, meterY, meterW, meterH, color.RGBA{10, 14, 22, 255}, false)
			vector.StrokeRect(screen, meterX, meterY, meterW, meterH, 0.8, color.RGBA{50, 65, 90, 255}, false)

			progressFraction := float32(t.CurrentCount) / float32(t.RequiredCount)
			if progressFraction > 1.0 {
				progressFraction = 1.0
			}
			if progressFraction > 0 {
				fillColor := color.RGBA{0, 190, 240, 255}
				if t.Completed {
					fillColor = color.RGBA{60, 220, 110, 255}
				}
				vector.FillRect(screen, meterX+1, meterY+1, (meterW-2)*progressFraction, meterH-2, fillColor, false)
			}

			countLabel := fmt.Sprintf("%d / %d", t.CurrentCount, t.RequiredCount)
			drawColoredDebugText(screen, countLabel, int(meterX+meterW)+12, int(meterY)-1, color.RGBA{180, 200, 220, 255})
		}

		taskStartY += boxH + 6
	}
}
