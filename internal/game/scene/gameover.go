package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

// GameOverScene represents the screen shown when the player dies.
type GameOverScene struct {
	panelX float64
	panelY float64
	panelW float64
	panelH float64

	respawnBtnX float64
	respawnBtnY float64
	respawnBtnW float64
	respawnBtnH float64

	titleBtnX float64
	titleBtnY float64
	titleBtnW float64
	titleBtnH float64

	touchIDs []ebiten.TouchID
}

// NewGameOverScene creates a new GameOverScene.
func NewGameOverScene() *GameOverScene {
	s := &GameOverScene{}
	s.layout()
	return s
}

// GameOverContext defines the narrow context interface required by GameOverScene.
type GameOverContext interface {
	GetInput() InputSource
	GetDeathReason() string
	Respawn()
	TransitionToTitle()
	SetCurrentState(s State)
}

func (s *GameOverScene) layout() {
	s.panelW = 680.0
	s.panelH = 430.0
	s.panelX = float64(config.ScreenWidth-int(s.panelW)) / 2.0
	s.panelY = float64(config.ScreenHeight-int(s.panelH)) / 2.0

	s.respawnBtnW = 310.0
	s.respawnBtnH = 52.0
	s.respawnBtnX = s.panelX + 22.0
	s.respawnBtnY = s.panelY + s.panelH - s.respawnBtnH - 20.0

	s.titleBtnW = 290.0
	s.titleBtnH = 52.0
	s.titleBtnX = s.panelX + s.panelW - s.titleBtnW - 22.0
	s.titleBtnY = s.respawnBtnY
}

func (s *GameOverScene) inRespawnBtn(x, y float64) bool {
	return x >= s.respawnBtnX && x < s.respawnBtnX+s.respawnBtnW &&
		y >= s.respawnBtnY && y < s.respawnBtnY+s.respawnBtnH
}

func (s *GameOverScene) inTitleBtn(x, y float64) bool {
	return x >= s.titleBtnX && x < s.titleBtnX+s.titleBtnW &&
		y >= s.titleBtnY && y < s.titleBtnY+s.titleBtnH
}

func (s *GameOverScene) OnEnter(g GameContext) {
	s.onEnter(g)
}

func (s *GameOverScene) onEnter(g GameOverContext) {
	s.layout()
}

func (s *GameOverScene) OnExit(g GameContext) {}

func (s *GameOverScene) Update(g GameContext) error {
	return s.update(g)
}

func (s *GameOverScene) update(g GameOverContext) error {
	s.layout()
	inp := g.GetInput()

	// 1. Keyboard shortcuts: ENTER/SPACE to respawn, ESC to return to title
	if inp.IsKeyJustPressed(ebiten.KeyEnter) || inp.IsKeyJustPressed(ebiten.KeySpace) {
		if audioMgr := audio.Get(); audioMgr != nil {
			audioMgr.PlaySFX("sfx/ui_confirm.wav")
		}
		g.Respawn()
		return nil
	}
	if inp.IsKeyJustPressed(ebiten.KeyEscape) {
		if audioMgr := audio.Get(); audioMgr != nil {
			audioMgr.PlaySFX("sfx/ui_cancel.wav")
		}
		g.TransitionToTitle()
		return nil
	}

	// 2. Mouse / desktop cursor clicks
	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := inp.Cursor()
		mx, my := float64(cursor.X), float64(cursor.Y)
		if s.inRespawnBtn(mx, my) {
			if audioMgr := audio.Get(); audioMgr != nil {
				audioMgr.PlaySFX("sfx/ui_confirm.wav")
			}
			g.Respawn()
			return nil
		}
		if s.inTitleBtn(mx, my) {
			if audioMgr := audio.Get(); audioMgr != nil {
				audioMgr.PlaySFX("sfx/ui_cancel.wav")
			}
			g.TransitionToTitle()
			return nil
		}
	}

	// 3. Mobile touch taps (support direct touch events)
	s.touchIDs = inpututil.AppendJustPressedTouchIDs(s.touchIDs[:0])
	for _, id := range s.touchIDs {
		tx, ty := ebiten.TouchPosition(id)
		fx, fy := float64(tx), float64(ty)
		if s.inRespawnBtn(fx, fy) {
			if audioMgr := audio.Get(); audioMgr != nil {
				audioMgr.PlaySFX("sfx/ui_confirm.wav")
			}
			g.Respawn()
			return nil
		}
		if s.inTitleBtn(fx, fy) {
			if audioMgr := audio.Get(); audioMgr != nil {
				audioMgr.PlaySFX("sfx/ui_cancel.wav")
			}
			g.TransitionToTitle()
			return nil
		}
	}

	return nil
}

func (s *GameOverScene) Draw(g GameContext, screen *ebiten.Image) {
	s.draw(g, screen)
}

func (s *GameOverScene) draw(g GameOverContext, screen *ebiten.Image) {
	s.layout()

	// Full-screen atmospheric death tint (deep abyssal crimson)
	vector.FillRect(screen, 0, 0, float32(config.ScreenWidth), float32(config.ScreenHeight), color.RGBA{R: 16, G: 6, B: 10, A: 245}, false)

	// Center panel outer frame & shadow glow
	vector.StrokeRect(screen, float32(s.panelX-3), float32(s.panelY-3), float32(s.panelW+6), float32(s.panelH+6), 1.0, color.RGBA{R: 220, G: 45, B: 55, A: 60}, false)
	vector.FillRect(screen, float32(s.panelX), float32(s.panelY), float32(s.panelW), float32(s.panelH), color.RGBA{R: 22, G: 10, B: 14, A: 240}, false)
	vector.StrokeRect(screen, float32(s.panelX), float32(s.panelY), float32(s.panelW), float32(s.panelH), 2.0, color.RGBA{R: 210, G: 45, B: 55, A: 220}, false)

	// Top emergency hazard stripe & header
	vector.FillRect(screen, float32(s.panelX), float32(s.panelY), float32(s.panelW), 56, color.RGBA{R: 45, G: 14, B: 18, A: 255}, false)
	vector.FillRect(screen, float32(s.panelX), float32(s.panelY), float32(s.panelW), 4, color.RGBA{R: 235, G: 55, B: 65, A: 255}, false)

	subHeader := "// CRITICAL ALERT: DIVER VITALS TERMINATED //"
	subX := int(s.panelX + (s.panelW-float64(len(subHeader)*6))/2.0)
	ebitenutil.DebugPrintAt(screen, subHeader, subX, int(s.panelY)+12)

	title := "G A M E   O V E R"
	titleX := int(s.panelX + (s.panelW-float64(len(title)*6))/2.0)
	ebitenutil.DebugPrintAt(screen, title, titleX, int(s.panelY)+32)

	vector.StrokeLine(screen, float32(s.panelX+20), float32(s.panelY+62), float32(s.panelX+s.panelW-20), float32(s.panelY+62), 1.5, color.RGBA{R: 190, G: 45, B: 55, A: 160}, false)

	// Incident cause box
	causeY := s.panelY + 74
	causeH := 58.0
	causeW := s.panelW - 44
	vector.FillRect(screen, float32(s.panelX+22), float32(causeY), float32(causeW), float32(causeH), color.RGBA{R: 36, G: 12, B: 18, A: 225}, false)
	vector.StrokeRect(screen, float32(s.panelX+22), float32(causeY), float32(causeW), float32(causeH), 1.5, color.RGBA{R: 190, G: 50, B: 60, A: 190}, false)
	vector.FillRect(screen, float32(s.panelX+22), float32(causeY), 4, float32(causeH), color.RGBA{R: 245, G: 65, B: 75, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "CAUSE OF DEATH:", int(s.panelX)+36, int(causeY)+10)
	msg := "Your hull cracked or you ran out of oxygen."
	if reason := g.GetDeathReason(); reason != "" {
		msg = reason
	}
	ebitenutil.DebugPrintAt(screen, msg, int(s.panelX)+36, int(causeY)+30)

	// Salvage & Recovery Telemetry
	telemY := s.panelY + 148
	ebitenutil.DebugPrintAt(screen, "SALVAGE & EMERGENCY PROTOCOL STATUS:", int(s.panelX)+26, int(telemY))
	vector.StrokeLine(screen, float32(s.panelX+26), float32(telemY+18), float32(s.panelX+s.panelW-26), float32(telemY+18), 1.0, color.RGBA{R: 180, G: 60, B: 70, A: 120}, false)

	bullets := []struct {
		tag  string
		text string
	}{
		{"[STATUS]", "Diver life support expired. Emergency telemetry broadcast active."},
		{"[CARGO] ", "Surface salvage crate deployed at incident site (upgrades kept)."},
		{"[RADAR] ", "Salvage crate coordinates are pinned to your PDA Navigation Map [M]."},
		{"[CLONE] ", "Emergency clone and equipment standing by at Life Pod base station."},
	}

	for i, b := range bullets {
		rowY := int(telemY) + 28 + i*26
		ebitenutil.DebugPrintAt(screen, b.tag+" "+b.text, int(s.panelX)+26, rowY)
	}

	// Divider line above buttons
	vector.StrokeLine(screen, float32(s.panelX+22), float32(s.respawnBtnY-16), float32(s.panelX+s.panelW-22), float32(s.respawnBtnY-16), 1.5, color.RGBA{R: 180, G: 45, B: 55, A: 140}, false)

	// Hover states
	inp := g.GetInput()
	cursor := inp.Cursor()
	mx, my := float64(cursor.X), float64(cursor.Y)
	hoverRespawn := s.inRespawnBtn(mx, my)
	hoverTitle := s.inTitleBtn(mx, my)

	// Respawn Button
	rBgColor := color.RGBA{R: 16, G: 48, B: 56, A: 240}
	rBorderColor := color.RGBA{R: 45, G: 190, B: 210, A: 255}
	if hoverRespawn {
		rBgColor = color.RGBA{R: 26, G: 80, B: 92, A: 255}
		rBorderColor = color.RGBA{R: 95, G: 240, B: 255, A: 255}
		vector.StrokeRect(screen, float32(s.respawnBtnX-2), float32(s.respawnBtnY-2), float32(s.respawnBtnW+4), float32(s.respawnBtnH+4), 1.0, color.RGBA{R: 95, G: 240, B: 255, A: 120}, false)
	}
	vector.FillRect(screen, float32(s.respawnBtnX), float32(s.respawnBtnY), float32(s.respawnBtnW), float32(s.respawnBtnH), rBgColor, false)
	vector.StrokeRect(screen, float32(s.respawnBtnX), float32(s.respawnBtnY), float32(s.respawnBtnW), float32(s.respawnBtnH), 2.0, rBorderColor, false)

	rLabel := "RESPAWN AT LIFE POD"
	rLabelX := int(s.respawnBtnX + (s.respawnBtnW-float64(len(rLabel)*6))/2.0)
	ebitenutil.DebugPrintAt(screen, rLabel, rLabelX, int(s.respawnBtnY)+12)

	rHint := "[ TAP / ENTER / SPACE ]"
	rHintX := int(s.respawnBtnX + (s.respawnBtnW-float64(len(rHint)*6))/2.0)
	ebitenutil.DebugPrintAt(screen, rHint, rHintX, int(s.respawnBtnY)+28)

	// Title / Main Menu Button
	tBgColor := color.RGBA{R: 42, G: 18, B: 24, A: 240}
	tBorderColor := color.RGBA{R: 190, G: 60, B: 70, A: 255}
	if hoverTitle {
		tBgColor = color.RGBA{R: 75, G: 28, B: 38, A: 255}
		tBorderColor = color.RGBA{R: 245, G: 95, B: 105, A: 255}
		vector.StrokeRect(screen, float32(s.titleBtnX-2), float32(s.titleBtnY-2), float32(s.titleBtnW+4), float32(s.titleBtnH+4), 1.0, color.RGBA{R: 245, G: 95, B: 105, A: 120}, false)
	}
	vector.FillRect(screen, float32(s.titleBtnX), float32(s.titleBtnY), float32(s.titleBtnW), float32(s.titleBtnH), tBgColor, false)
	vector.StrokeRect(screen, float32(s.titleBtnX), float32(s.titleBtnY), float32(s.titleBtnW), float32(s.titleBtnH), 2.0, tBorderColor, false)

	tLabel := "MAIN MENU"
	tLabelX := int(s.titleBtnX + (s.titleBtnW-float64(len(tLabel)*6))/2.0)
	ebitenutil.DebugPrintAt(screen, tLabel, tLabelX, int(s.titleBtnY)+12)

	tHint := "[ TAP / ESCAPE ]"
	tHintX := int(s.titleBtnX + (s.titleBtnW-float64(len(tHint)*6))/2.0)
	ebitenutil.DebugPrintAt(screen, tHint, tHintX, int(s.titleBtnY)+28)
}
