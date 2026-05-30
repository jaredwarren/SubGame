package scene

import (
	"image/color"
	_ "image/jpeg"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

// TitleScene manages the title screen.
type TitleScene struct {
	backgroundImage        *ebiten.Image
	backgroundLoadErr      error
	titleText              string
	btnX, btnY, btnW, btnH float64
}

// NewTitleScene creates a new TitleScene.
func NewTitleScene() *TitleScene {
	s := &TitleScene{
		titleText: "S U B G A M E",
		btnW:      240,
		btnH:      60,
	}
	s.btnX = (float64(config.ScreenWidth) - s.btnW) / 2.0
	s.btnY = 460.0

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

func (s *TitleScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateTitle)
}

func (s *TitleScene) OnExit(g GameContext) {}

func (s *TitleScene) Update(g GameContext) error {
	inp := g.GetInput()

	if inp.IsKeyJustPressed(ebiten.KeyEnter) {
		g.TransitionToOverworld()
		return nil
	}

	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := inp.Cursor()
		mx, my := cursor.X, cursor.Y
		if mx >= s.btnX && mx < s.btnX+s.btnW && my >= s.btnY && my < s.btnY+s.btnH {
			g.TransitionToOverworld()
			return nil
		}
	}

	return nil
}

func (s *TitleScene) Draw(g GameContext, screen *ebiten.Image) {
	if s.backgroundImage != nil {
		bounds := s.backgroundImage.Bounds()
		op := &ebiten.DrawImageOptions{}
		scaleX := float64(config.ScreenWidth) / float64(bounds.Dx())
		scaleY := float64(config.ScreenHeight) / float64(bounds.Dy())
		op.GeoM.Scale(scaleX, scaleY)
		screen.DrawImage(s.backgroundImage, op)
	} else {
		for y := 0; y < config.ScreenHeight; y++ {
			ratio := float64(y) / float64(config.ScreenHeight)
			r := uint8(5 - 5*ratio)
			gr := uint8(20 - 15*ratio)
			b := uint8(45 - 25*ratio)
			vector.StrokeLine(screen, 0, float32(y), float32(config.ScreenWidth), float32(y), 1.0, color.RGBA{R: r, G: gr, B: b, A: 255}, false)
		}
	}

	vector.FillRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{R: 0, G: 4, B: 12, A: 160}, false)

	titleImg := ebiten.NewImage(200, 20)
	ebitenutil.DebugPrintAt(titleImg, s.titleText, 40, 2)

	op := &ebiten.DrawImageOptions{}
	scale := 5.0
	titleW := 200.0 * scale
	titleH := 20.0 * scale
	tx := (float64(config.ScreenWidth) - titleW) / 2.0
	ty := 150.0
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(tx, ty)
	screen.DrawImage(titleImg, op)

	subText := "D E E P   O C E A N   S U R V I V A L   A D V E N T U R E"
	subX := (config.ScreenWidth - len(subText)*6) / 2
	ebitenutil.DebugPrintAt(screen, subText, subX, int(ty+titleH+20))

	inp := g.GetInput()
	cursor := inp.Cursor()
	mx, my := cursor.X, cursor.Y
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

	btnText := "D I V E"
	btnTextImg := ebiten.NewImage(80, 16)
	ebitenutil.DebugPrintAt(btnTextImg, btnText, 20, 0)

	btnTextOp := &ebiten.DrawImageOptions{}
	btnTextScale := 2.0
	btnTextW := 80.0 * btnTextScale
	btnTextH := 16.0 * btnTextScale
	btnTextX := s.btnX + (s.btnW-btnTextW)/2.0
	btnTextY := s.btnY + (s.btnH-btnTextH)/2.0

	btnTextOp.GeoM.Scale(btnTextScale, btnTextScale)
	btnTextOp.GeoM.Translate(btnTextX, btnTextY)
	screen.DrawImage(btnTextImg, btnTextOp)

	instText := "Press ENTER or Click DIVE to begin your descent"
	instX := (config.ScreenWidth - len(instText)*6) / 2
	ebitenutil.DebugPrintAt(screen, instText, instX, int(s.btnY+s.btnH+40))
}
