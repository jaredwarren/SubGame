package scene

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	"log"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/save"
)

type titlePanelMode int

const (
	panelNone titlePanelMode = iota
	panelSlots
)

// TitleScene manages the title screen.
type TitleScene struct {
	backgroundImage            *ebiten.Image
	backgroundLoadErr          error
	titleText                  string
	btnX, btnY, btnW, btnH     float64
	seedText                   string
	seedX, seedY, seedW, seedH float64
	seedFocused                bool
	runesBuffer                []rune

	panelMode         titlePanelMode
	slotInfos         []save.SlotInfo
	deleteConfirmSlot int

	slotPanelX, slotPanelW float64
	slotRowY               [save.NumSlots]float64
	slotRowH               float64
	backBtnX, backBtnY     float64
	deleteBtnW             float64

	fallbackBackground *ebiten.Image
	titleImg           *ebiten.Image
	btnTextImg         *ebiten.Image
	seedTextImg        *ebiten.Image
	lastDisplayText    string
}

// NewTitleScene creates a new TitleScene.
func NewTitleScene() *TitleScene {
	s := &TitleScene{
		titleText: "S U B G A M E",
		btnW:      240,
		btnH:      50,
		seedText:  "12345",
		seedW:     240,
		seedH:     40,
	}
	s.btnX = (float64(config.ScreenWidth) - s.btnW) / 2.0
	s.btnY = 460.0

	s.seedX = (float64(config.ScreenWidth) - s.seedW) / 2.0
	s.seedY = 525.0

	s.slotPanelW = 520
	s.slotPanelX = (float64(config.ScreenWidth) - s.slotPanelW) / 2.0
	s.slotRowH = 52
	s.deleteBtnW = 110
	s.backBtnX = s.btnX
	s.layoutSlotPanel() // slot row geometry
	s.layoutMain()      // start-screen seed/button Y (must come last)

	// Pre-render fallback background gradient
	s.fallbackBackground = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	for y := 0; y < config.ScreenHeight; y++ {
		ratio := float64(y) / float64(config.ScreenHeight)
		r := uint8(5 - 5*ratio)
		gr := uint8(20 - 15*ratio)
		b := uint8(45 - 25*ratio)
		vector.StrokeLine(s.fallbackBackground, 0, float32(y), float32(config.ScreenWidth), float32(y), 1.0, color.RGBA{R: r, G: gr, B: b, A: 255}, false)
	}

	// Pre-render static text images
	s.titleImg = ebiten.NewImage(200, 20)
	ebitenutil.DebugPrintAt(s.titleImg, s.titleText, 40, 2)

	s.btnTextImg = ebiten.NewImage(80, 16)
	ebitenutil.DebugPrintAt(s.btnTextImg, "START", 22, 0)

	// Pre-allocate dynamic seed text image
	s.seedTextImg = ebiten.NewImage(int(s.seedW), 20)

	paths := []string{
		"StartBackground.jpeg",
		"/Users/jaredwarren/src/github.com/jaredwarren/SubGame/StartBackground.jpeg",
		"../../StartBackground.jpeg",
		"../StartBackground.jpeg",
	}

	var img *ebiten.Image
	var err error
	for _, p := range paths {
		img, _, err = ebitenutil.NewImageFromFile(p)
		if err == nil {
			s.backgroundImage = img
			break
		}
	}
	if err != nil {
		s.backgroundLoadErr = err
		log.Printf("Warning: Failed to load title background image: %v", err)
	}

	return s
}

// TitleContext defines the narrow context interface required by TitleScene.
type TitleContext interface {
	GetInput() InputSource
	TransitionToIntro(seed int64)
	GetTicks() float64
	SetCurrentState(s State)
	HasSaveFile() bool
	ListSaveSlots() []save.SlotInfo
	LoadSaveSlot(slot int) error
	DeleteSaveSlot(slot int) error
	SetActiveSaveSlot(slot int)
}

func (s *TitleScene) OnEnter(g GameContext) {
	s.onEnter(g)
}

func (s *TitleScene) onEnter(g TitleContext) {
	g.SetCurrentState(StateTitle)
	s.panelMode = panelNone
	s.deleteConfirmSlot = 0
	s.seedFocused = false
	s.layoutMain()
}

func (s *TitleScene) OnExit(g GameContext) {}

func (s *TitleScene) Update(g GameContext) error {
	return s.update(g)
}

func (s *TitleScene) layoutMain() {
	s.btnY = 460.0
	s.seedY = 525.0
}

func (s *TitleScene) layoutSlotPanel() {
	const (
		firstSlotY = 295.0
		slotGap    = 12.0
	)
	for i := 0; i < save.NumSlots; i++ {
		s.slotRowY[i] = firstSlotY + float64(i)*(s.slotRowH+slotGap)
	}
	lastBottom := s.slotRowY[save.NumSlots-1] + s.slotRowH
	s.seedY = lastBottom + 18
	s.backBtnY = s.seedY + s.seedH + 14
}

func (s *TitleScene) openSlotPanel(g TitleContext) {
	s.panelMode = panelSlots
	s.deleteConfirmSlot = 0
	s.slotInfos = g.ListSaveSlots()
	s.seedFocused = false
	s.layoutSlotPanel()
}

func (s *TitleScene) closePanel() {
	s.panelMode = panelNone
	s.deleteConfirmSlot = 0
	s.layoutMain()
	audio.Get().PlaySFX("sfx/ui_hover.wav")
}

func (s *TitleScene) startFresh(g TitleContext, slot int) {
	g.SetActiveSaveSlot(slot)
	g.TransitionToIntro(parseSeed(s.seedText))
}

func (s *TitleScene) handleStart(g TitleContext) error {
	if !g.HasSaveFile() {
		audio.Get().PlaySFX("sfx/ui_confirm.wav")
		s.startFresh(g, 1)
		return nil
	}
	audio.Get().PlaySFX("sfx/ui_confirm.wav")
	s.openSlotPanel(g)
	return nil
}

func (s *TitleScene) update(g TitleContext) error {
	inp := g.GetInput()

	if s.panelMode != panelNone {
		return s.updateSlotPanel(g, inp)
	}

	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := inp.Cursor()
		mx, my := cursor.X, cursor.Y

		if mx >= s.seedX && mx < s.seedX+s.seedW && my >= s.seedY && my < s.seedY+s.seedH {
			s.seedFocused = true
			audio.Get().PlaySFX("sfx/ui_hover.wav")
		} else {
			s.seedFocused = false
		}

		if mx >= s.btnX && mx < s.btnX+s.btnW && my >= s.btnY && my < s.btnY+s.btnH {
			return s.handleStart(g)
		}
	}

	if s.seedFocused {
		s.runesBuffer = inp.AppendInputChars(s.runesBuffer[:0])
		for _, r := range s.runesBuffer {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				if len(s.seedText) < 20 {
					s.seedText += string(r)
					audio.Get().PlaySFX("sfx/pda_typewriter_tick.wav")
				}
			}
		}
		if inp.IsKeyJustPressed(ebiten.KeyBackspace) {
			if len(s.seedText) > 0 {
				s.seedText = s.seedText[:len(s.seedText)-1]
				audio.Get().PlaySFX("sfx/pda_typewriter_tick.wav")
			}
		}
	}

	if inp.IsKeyJustPressed(ebiten.KeyEnter) {
		return s.handleStart(g)
	}

	return nil
}

func (s *TitleScene) slotInfo(slot int) save.SlotInfo {
	for _, info := range s.slotInfos {
		if info.Slot == slot {
			return info
		}
	}
	return save.SlotInfo{Slot: slot}
}

func (s *TitleScene) updateSlotPanel(g TitleContext, inp InputSource) error {
	if inp.IsKeyJustPressed(ebiten.KeyEscape) {
		s.closePanel()
		return nil
	}

	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := inp.Cursor()
		mx, my := cursor.X, cursor.Y

		if mx >= s.backBtnX && mx < s.backBtnX+s.btnW && my >= s.backBtnY && my < s.backBtnY+s.btnH {
			audio.Get().PlaySFX("sfx/ui_confirm.wav")
			s.closePanel()
			return nil
		}

		for i := 0; i < save.NumSlots; i++ {
			slot := i + 1
			rowY := s.slotRowY[i]
			info := s.slotInfo(slot)
			deleteX := s.slotPanelX + s.slotPanelW - s.deleteBtnW

			if info.Occupied &&
				mx >= deleteX && mx < s.slotPanelX+s.slotPanelW &&
				my >= rowY && my < rowY+s.slotRowH {
				if s.deleteConfirmSlot == slot {
					audio.Get().PlaySFX("sfx/ui_confirm.wav")
					if err := g.DeleteSaveSlot(slot); err != nil {
						return err
					}
					s.slotInfos = g.ListSaveSlots()
					s.deleteConfirmSlot = 0
					if !g.HasSaveFile() {
						s.closePanel()
					}
				} else {
					audio.Get().PlaySFX("sfx/ui_hover.wav")
					s.deleteConfirmSlot = slot
				}
				return nil
			}

			rowClickRight := deleteX
			if !info.Occupied {
				rowClickRight = s.slotPanelX + s.slotPanelW
			}
			if mx >= s.slotPanelX && mx < rowClickRight &&
				my >= rowY && my < rowY+s.slotRowH {
				audio.Get().PlaySFX("sfx/ui_confirm.wav")
				if info.Occupied {
					return g.LoadSaveSlot(slot)
				}
				s.startFresh(g, slot)
				return nil
			}
		}

		if mx >= s.seedX && mx < s.seedX+s.seedW && my >= s.seedY && my < s.seedY+s.seedH {
			s.seedFocused = true
			audio.Get().PlaySFX("sfx/ui_hover.wav")
		} else {
			s.seedFocused = false
			s.deleteConfirmSlot = 0
		}
	}

	if s.seedFocused {
		s.runesBuffer = inp.AppendInputChars(s.runesBuffer[:0])
		for _, r := range s.runesBuffer {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				if len(s.seedText) < 20 {
					s.seedText += string(r)
					audio.Get().PlaySFX("sfx/pda_typewriter_tick.wav")
				}
			}
		}
		if inp.IsKeyJustPressed(ebiten.KeyBackspace) {
			if len(s.seedText) > 0 {
				s.seedText = s.seedText[:len(s.seedText)-1]
				audio.Get().PlaySFX("sfx/pda_typewriter_tick.wav")
			}
		}
	}

	return nil
}

func parseSeed(text string) int64 {
	if text == "" {
		return 12345
	}
	val, err := strconv.ParseInt(text, 10, 64)
	if err == nil {
		return val
	}
	var hash int64
	for _, char := range text {
		hash = hash*31 + int64(char)
	}
	return hash
}

func formatSlotTimestamp(ts int64) string {
	if ts == 0 {
		return "unknown date"
	}
	return time.Unix(ts, 0).Format("Jan 2, 15:04")
}

func (s *TitleScene) Draw(g GameContext, screen *ebiten.Image) {
	s.draw(g, screen)
}

func (s *TitleScene) draw(g TitleContext, screen *ebiten.Image) {
	if s.backgroundImage != nil {
		bounds := s.backgroundImage.Bounds()
		op := &ebiten.DrawImageOptions{}
		scaleX := float64(config.ScreenWidth) / float64(bounds.Dx())
		scaleY := float64(config.ScreenHeight) / float64(bounds.Dy())
		op.GeoM.Scale(scaleX, scaleY)
		screen.DrawImage(s.backgroundImage, op)
	} else if s.fallbackBackground != nil {
		screen.DrawImage(s.fallbackBackground, nil)
	}

	vector.FillRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{R: 0, G: 4, B: 12, A: 160}, false)

	op := &ebiten.DrawImageOptions{}
	scale := 5.0
	titleW := 200.0 * scale
	titleH := 20.0 * scale
	tx := (float64(config.ScreenWidth) - titleW) / 2.0
	ty := 150.0
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(tx, ty)
	if s.titleImg != nil {
		screen.DrawImage(s.titleImg, op)
	}

	inp := g.GetInput()
	cursor := inp.Cursor()
	mx, my := cursor.X, cursor.Y

	if s.panelMode == panelNone {
		subText := "D E E P   O C E A N   S U R V I V A L   A D V E N T U R E"
		subX := (config.ScreenWidth - len(subText)*6) / 2
		ebitenutil.DebugPrintAt(screen, subText, subX, int(ty+titleH+20))
		s.drawStartButton(screen, mx, my)
	} else {
		s.drawSlotPanel(screen, mx, my)
	}

	s.drawSeedInput(screen, g, mx, my)

	instText := "Press ENTER or click START to begin"
	instY := int(s.seedY + s.seedH + 18)
	if s.panelMode == panelSlots {
		instText = "Select a slot to continue, start fresh, or delete"
		instY = int(s.backBtnY + s.btnH + 16)
	} else if g.HasSaveFile() {
		instText = "Press ENTER or click START to manage save slots"
	}
	instX := (config.ScreenWidth - len(instText)*6) / 2
	ebitenutil.DebugPrintAt(screen, instText, instX, instY)
}

func (s *TitleScene) drawStartButton(screen *ebiten.Image, mx, my float64) {
	isHovered := mx >= s.btnX && mx < s.btnX+s.btnW && my >= s.btnY && my < s.btnY+s.btnH
	btnBgColor := color.RGBA{R: 12, G: 28, B: 48, A: 200}
	btnBorderColor := color.RGBA{R: 45, G: 130, B: 200, A: 255}

	if isHovered {
		btnBgColor = color.RGBA{R: 20, G: 45, B: 80, A: 240}
		btnBorderColor = color.RGBA{R: 60, G: 210, B: 240, A: 255}
		vector.StrokeRect(screen, float32(s.btnX-2), float32(s.btnY-2), float32(s.btnW+4), float32(s.btnH+4), 1.0, color.RGBA{R: 60, G: 210, B: 240, A: 100}, false)
	}

	vector.FillRect(screen, float32(s.btnX), float32(s.btnY), float32(s.btnW), float32(s.btnH), btnBgColor, false)
	vector.StrokeRect(screen, float32(s.btnX), float32(s.btnY), float32(s.btnW), float32(s.btnH), 2.0, btnBorderColor, false)

	btnTextOp := &ebiten.DrawImageOptions{}
	btnTextScale := 2.0
	btnTextW := 80.0 * btnTextScale
	btnTextH := 16.0 * btnTextScale
	btnTextX := s.btnX + (s.btnW-btnTextW)/2.0
	btnTextY := s.btnY + (s.btnH-btnTextH)/2.0

	btnTextOp.GeoM.Scale(btnTextScale, btnTextScale)
	btnTextOp.GeoM.Translate(btnTextX, btnTextY)
	if s.btnTextImg != nil {
		screen.DrawImage(s.btnTextImg, btnTextOp)
	}
}

func (s *TitleScene) drawSlotPanel(screen *ebiten.Image, mx, my float64) {
	panelTitle := "SELECT SAVE SLOT"
	titleX := (config.ScreenWidth - len(panelTitle)*6) / 2
	ebitenutil.DebugPrintAt(screen, panelTitle, titleX, 270)

	for i := 0; i < save.NumSlots; i++ {
		slot := i + 1
		rowY := s.slotRowY[i]
		info := s.slotInfo(slot)
		deleteX := s.slotPanelX + s.slotPanelW - s.deleteBtnW

		rowHoverRight := deleteX
		if !info.Occupied {
			rowHoverRight = s.slotPanelX + s.slotPanelW
		}
		isHovered := mx >= s.slotPanelX && mx < rowHoverRight &&
			my >= rowY && my < rowY+s.slotRowH

		bg := color.RGBA{R: 8, G: 18, B: 32, A: 200}
		border := color.RGBA{R: 45, G: 130, B: 200, A: 150}
		if isHovered {
			bg = color.RGBA{R: 20, G: 45, B: 80, A: 240}
			border = color.RGBA{R: 60, G: 210, B: 240, A: 255}
		}

		vector.FillRect(screen, float32(s.slotPanelX), float32(rowY), float32(s.slotPanelW), float32(s.slotRowH), bg, false)
		vector.StrokeRect(screen, float32(s.slotPanelX), float32(rowY), float32(s.slotPanelW), float32(s.slotRowH), 1.5, border, false)

		label := fmt.Sprintf("SLOT %d  EMPTY", slot)
		if info.Occupied {
			label = fmt.Sprintf("SLOT %d  Seed %d  %s", slot, info.WorldSeed, formatSlotTimestamp(info.Timestamp))
		}
		ebitenutil.DebugPrintAt(screen, label, int(s.slotPanelX+14), int(rowY+(s.slotRowH-8)/2))

		if info.Occupied {
			delHovered := mx >= deleteX && mx < s.slotPanelX+s.slotPanelW &&
				my >= rowY && my < rowY+s.slotRowH
			delBg := color.RGBA{R: 48, G: 16, B: 16, A: 220}
			delBorder := color.RGBA{R: 180, G: 60, B: 60, A: 255}
			if delHovered {
				delBg = color.RGBA{R: 80, G: 24, B: 24, A: 255}
				delBorder = color.RGBA{R: 240, G: 90, B: 90, A: 255}
			}
			vector.FillRect(screen, float32(deleteX), float32(rowY+4), float32(s.deleteBtnW-4), float32(s.slotRowH-8), delBg, false)
			vector.StrokeRect(screen, float32(deleteX), float32(rowY+4), float32(s.deleteBtnW-4), float32(s.slotRowH-8), 1.5, delBorder, false)

			delLabel := "DELETE"
			if s.deleteConfirmSlot == slot {
				delLabel = "CONFIRM?"
			}
			delLabelX := int(deleteX + (s.deleteBtnW-float64(len(delLabel)*6))/2.0)
			ebitenutil.DebugPrintAt(screen, delLabel, delLabelX, int(rowY+(s.slotRowH-8)/2))
		}
	}

	isBackHovered := mx >= s.backBtnX && mx < s.backBtnX+s.btnW && my >= s.backBtnY && my < s.backBtnY+s.btnH
	backBg := color.RGBA{R: 12, G: 28, B: 48, A: 200}
	backBorder := color.RGBA{R: 45, G: 130, B: 200, A: 255}
	if isBackHovered {
		backBg = color.RGBA{R: 20, G: 45, B: 80, A: 240}
		backBorder = color.RGBA{R: 60, G: 210, B: 240, A: 255}
	}
	vector.FillRect(screen, float32(s.backBtnX), float32(s.backBtnY), float32(s.btnW), float32(s.btnH), backBg, false)
	vector.StrokeRect(screen, float32(s.backBtnX), float32(s.backBtnY), float32(s.btnW), float32(s.btnH), 2.0, backBorder, false)
	backLabelX := int(s.backBtnX + (s.btnW-36)/2.0)
	ebitenutil.DebugPrintAt(screen, "BACK", backLabelX, int(s.backBtnY+(s.btnH-8)/2))
}

func (s *TitleScene) drawSeedInput(screen *ebiten.Image, g TitleContext, mx, my float64) {
	isSeedHovered := mx >= s.seedX && mx < s.seedX+s.seedW && my >= s.seedY && my < s.seedY+s.seedH

	seedBgColor := color.RGBA{R: 8, G: 18, B: 32, A: 200}
	seedBorderColor := color.RGBA{R: 45, G: 130, B: 200, A: 150}

	if s.seedFocused {
		seedBgColor = color.RGBA{R: 12, G: 28, B: 48, A: 220}
		seedBorderColor = color.RGBA{R: 60, G: 210, B: 240, A: 255}
		vector.StrokeRect(screen, float32(s.seedX-2), float32(s.seedY-2), float32(s.seedW+4), float32(s.seedH+4), 1.0, color.RGBA{R: 60, G: 210, B: 240, A: 80}, false)
	} else if isSeedHovered {
		seedBgColor = color.RGBA{R: 10, G: 22, B: 40, A: 210}
		seedBorderColor = color.RGBA{R: 60, G: 210, B: 240, A: 180}
	}

	vector.FillRect(screen, float32(s.seedX), float32(s.seedY), float32(s.seedW), float32(s.seedH), seedBgColor, false)
	vector.StrokeRect(screen, float32(s.seedX), float32(s.seedY), float32(s.seedW), float32(s.seedH), 1.5, seedBorderColor, false)

	displayText := "Seed: " + s.seedText
	if s.seedText == "" {
		displayText = "Seed: (random)"
	}

	if s.seedFocused && (int(g.GetTicks())/30)%2 == 0 {
		displayText += "|"
	}

	if s.seedTextImg != nil {
		if displayText != s.lastDisplayText {
			s.lastDisplayText = displayText
			s.seedTextImg.Fill(color.RGBA{0, 0, 0, 0})
			ebitenutil.DebugPrintAt(s.seedTextImg, displayText, 12, 2)
		}

		seedTextOp := &ebiten.DrawImageOptions{}
		if s.seedText == "" {
			seedTextOp.ColorScale.Scale(0.6, 0.7, 0.8, 1.0)
		} else if !s.seedFocused {
			seedTextOp.ColorScale.Scale(0.9, 0.9, 0.9, 1.0)
		}
		seedTextOp.GeoM.Translate(s.seedX, s.seedY+(s.seedH-20)/2)
		screen.DrawImage(s.seedTextImg, seedTextOp)
	}
}
