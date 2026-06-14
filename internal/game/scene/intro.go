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
	seed       int64
	background *ebiten.Image
	textLines  []introTextLine
}

type introTextLine struct {
	img  *ebiten.Image
	x    float64
	y    float64
	clr  color.Color
	text string
}

// NewIntroScene creates a new IntroScene.
func NewIntroScene() *IntroScene {
	s := &IntroScene{}
	s.initialize()
	return s
}

func (s *IntroScene) initialize() {
	// Pre-render background gradient
	s.background = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	for y := 0; y < config.ScreenHeight; y++ {
		ratio := float64(y) / float64(config.ScreenHeight)
		r := uint8(2 - 2*ratio)
		gr := uint8(8 - 6*ratio)
		b := uint8(22 - 12*ratio)
		vector.StrokeLine(s.background, 0, float32(y), float32(config.ScreenWidth), float32(y), 1.0, color.RGBA{R: r, G: gr, B: b, A: 255}, false)
	}

	// Pre-render text lines
	panelW := 640.0
	panelX := (float64(config.ScreenWidth) - panelW) / 2.0
	panelY := (float64(config.ScreenHeight) - 360.0) / 2.0
	textStartY := panelY + 30.0

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

	s.textLines = make([]introTextLine, 0, len(lines))
	for i, line := range lines {
		if line == "" {
			continue
		}
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

		lx := panelX + float64(offsetX)
		ly := textStartY + float64(i*22)

		w := len(line) * 6
		textImg := ebiten.NewImage(w, 16)
		ebitenutil.DebugPrintAt(textImg, line, 0, 0)

		s.textLines = append(s.textLines, introTextLine{
			img:  textImg,
			x:    lx,
			y:    ly,
			clr:  textColor,
			text: line,
		})
	}
}

// SetSeed stores the world seed to be used when starting the game.
func (s *IntroScene) SetSeed(seed int64) {
	s.seed = seed
}

// IntroContext defines the narrow context interface required by IntroScene.
type IntroContext interface {
	GetInput() InputSource
	StartGame(seed int64)
	SetCurrentState(s State)
}

func (s *IntroScene) OnEnter(g GameContext) {
	s.onEnter(g)
}

func (s *IntroScene) onEnter(g IntroContext) {
	g.SetCurrentState(StateIntro)
}

func (s *IntroScene) OnExit(g GameContext) {}

func (s *IntroScene) Update(g GameContext) error {
	return s.update(g)
}

func (s *IntroScene) update(g IntroContext) error {
	inp := g.GetInput()
	if inp.IsKeyJustPressed(ebiten.KeyEnter) || inp.IsKeyJustPressed(ebiten.KeySpace) || inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.StartGame(s.seed)
		return nil
	}
	return nil
}

func (s *IntroScene) Draw(g GameContext, screen *ebiten.Image) {
	s.draw(g, screen)
}

func (s *IntroScene) draw(g IntroContext, screen *ebiten.Image) {
	if s.background != nil {
		screen.DrawImage(s.background, nil)
	}

	panelW := float32(640)
	panelH := float32(360)
	panelX := float32(config.ScreenWidth-int(panelW)) / 2.0
	panelY := float32(config.ScreenHeight-int(panelH)) / 2.0

	// Draw terminal window
	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{4, 10, 18, 230}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 1.5, color.RGBA{45, 130, 200, 255}, false)

	for _, line := range s.textLines {
		// Draw text drop shadow
		ebitenutil.DebugPrintAt(screen, line.text, int(line.x)+1, int(line.y)+1)

		// Draw main text
		if line.img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(line.x, line.y)
			op.ColorScale.ScaleWithColor(line.clr)
			screen.DrawImage(line.img, op)
		}
	}
}
