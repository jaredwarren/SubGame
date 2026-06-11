package scene

import (
	"image/color"
	_ "image/jpeg"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
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
		btnH:      60,
		seedText:  "12345",
		seedW:     240,
		seedH:     40,
	}
	s.btnX = (float64(config.ScreenWidth) - s.btnW) / 2.0
	s.btnY = 460.0

	s.seedX = (float64(config.ScreenWidth) - s.seedW) / 2.0
	s.seedY = 535.0

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
	ebitenutil.DebugPrintAt(s.btnTextImg, "D I V E", 20, 0)

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

func (s *TitleScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateTitle)
}

func (s *TitleScene) OnExit(g GameContext) {}

func (s *TitleScene) Update(g GameContext) error {
	inp := g.GetInput()

	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := inp.Cursor()
		mx, my := cursor.X, cursor.Y
		if mx >= s.seedX && mx < s.seedX+s.seedW && my >= s.seedY && my < s.seedY+s.seedH {
			s.seedFocused = true
		} else {
			s.seedFocused = false
		}
	}

	if s.seedFocused {
		s.runesBuffer = inp.AppendInputChars(s.runesBuffer[:0])
		for _, r := range s.runesBuffer {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				if len(s.seedText) < 20 {
					s.seedText += string(r)
				}
			}
		}
		if inp.IsKeyJustPressed(ebiten.KeyBackspace) {
			if len(s.seedText) > 0 {
				s.seedText = s.seedText[:len(s.seedText)-1]
			}
		}
	}

	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := inp.Cursor()
		mx, my := cursor.X, cursor.Y
		if mx >= s.btnX && mx < s.btnX+s.btnW && my >= s.btnY && my < s.btnY+s.btnH {
			g.TransitionToIntro(parseSeed(s.seedText))
			return nil
		}
	}

	if inp.IsKeyJustPressed(ebiten.KeyEnter) {
		g.TransitionToIntro(parseSeed(s.seedText))
		return nil
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

func (s *TitleScene) Draw(g GameContext, screen *ebiten.Image) {
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

	// Bounding box checking for hovering the seed input
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

	instText := "Press ENTER or Click DIVE to begin your descent"
	instX := (config.ScreenWidth - len(instText)*6) / 2
	ebitenutil.DebugPrintAt(screen, instText, instX, int(s.seedY+s.seedH+25))
}
