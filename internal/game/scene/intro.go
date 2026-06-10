package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

// IntroScene displays introductory narrative lore before the game starts.
type IntroScene struct {
	seed int64
}

// NewIntroScene creates a new IntroScene.
func NewIntroScene() *IntroScene {
	return &IntroScene{}
}

// SetSeed stores the world seed to be used when starting the game.
func (s *IntroScene) SetSeed(seed int64) {
	s.seed = seed
}

func (s *IntroScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateIntro)
}

func (s *IntroScene) OnExit(g GameContext) {}

func (s *IntroScene) Update(g GameContext) error {
	inp := g.GetInput()
	if inp.IsKeyJustPressed(ebiten.KeyEnter) || inp.IsKeyJustPressed(ebiten.KeySpace) || inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.StartGame(s.seed)
		return nil
	}
	return nil
}

func (s *IntroScene) Draw(g GameContext, screen *ebiten.Image) {
	// Draw a dark space/ocean gradient background
	for y := 0; y < config.ScreenHeight; y++ {
		ratio := float64(y) / float64(config.ScreenHeight)
		r := uint8(2 - 2*ratio)
		gr := uint8(8 - 6*ratio)
		b := uint8(22 - 12*ratio)
		vector.StrokeLine(screen, 0, float32(y), float32(config.ScreenWidth), float32(y), 1.0, color.RGBA{R: r, G: gr, B: b, A: 255}, false)
	}

	panelW := float32(640)
	panelH := float32(360)
	panelX := float32(config.ScreenWidth-int(panelW)) / 2.0
	panelY := float32(config.ScreenHeight-int(panelH)) / 2.0

	// Draw terminal window
	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{4, 10, 18, 230}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{45, 130, 200, 255}, false)

	lines := []string{
		"★ AETHERCORP INBOUND VESSEL LOG - MISSION #44-B ★",
		"=================================================",
		"CONTRACT OBJ : Salvage missing research vessel 'Triton'.",
		"CARGO FOCUS  : Recover Abyssal Ore nodes and core data logs.",
		"STATUS       : Atmospheric descent established.",
		"",
		"WARNING: Severe electromagnetic anomaly detected from seabed.",
		"Warning: Critical flight system disruption... reactor overload.",
		"Warning: Auxiliary power failing. Manual ejection initialized.",
		"",
		"System: Escape Pod 5 ejected. Sinking into ocean coordinates...",
		"",
		"[ Press SPACE, ENTER, or CLICK to splash down and begin. ]",
	}

	textStartY := int(panelY) + 30
	for i, line := range lines {
		// Highlight warnings in red/orange, titles in gold, prompt in green
		textColor := color.RGBA{180, 210, 240, 255}
		if i == 0 {
			textColor = color.RGBA{220, 180, 50, 255} // Gold
		} else if i == 1 {
			textColor = color.RGBA{45, 130, 200, 255} // Blue line
		} else if i >= 6 && i <= 8 {
			textColor = color.RGBA{240, 80, 80, 255} // Red warning
		} else if i == 12 {
			textColor = color.RGBA{60, 210, 110, 255} // Green prompt
		}

		offsetX := (int(panelW) - len(line)*6) / 2
		if offsetX < 20 {
			offsetX = 20
		}

		// Draw text drop shadow
		ebitenutil.DebugPrintAt(screen, line, int(panelX)+offsetX+1, textStartY+i*22+1)
		
		// Draw main text
		drawColoredText(screen, line, int(panelX)+offsetX, textStartY+i*22, textColor)
	}
}

func drawColoredText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	w := len(str) * 6
	if w <= 0 {
		return
	}
	textImg := ebiten.NewImage(w, 16)
	ebitenutil.DebugPrintAt(textImg, str, 0, 0)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)

	screen.DrawImage(textImg, op)
}
