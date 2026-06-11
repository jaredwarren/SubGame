package vehicle

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/assets"
)

// LoadAssets loads and processes all vehicle sheets at game startup.
func LoadAssets() {
	loadHeavyMechSheet()
	loadScoutSubSheet()
	loadSkiffSheet()
}

func loadHeavyMechSheet() {
	img, _, err := image.Decode(bytes.NewReader(assets.HeavyMechPNG))
	if err != nil {
		log.Printf("Error: Failed to decode heavy mech sheet: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Chroma-key green pixels using fast direct byte slice manipulation
	for i := 0; i < len(rgba.Pix); i += 4 {
		ru := rgba.Pix[i]
		gu := rgba.Pix[i+1]
		bu := rgba.Pix[i+2]

		if gu > 140 && ru < 100 && bu < 100 {
			rgba.Pix[i] = 0
			rgba.Pix[i+1] = 0
			rgba.Pix[i+2] = 0
			rgba.Pix[i+3] = 0
		}
	}

	heavyMechSheet = ebiten.NewImageFromImage(rgba)
}

func loadScoutSubSheet() {
	img, _, err := image.Decode(bytes.NewReader(assets.ScoutSubPNG))
	if err != nil {
		log.Printf("Error: Failed to decode scout sub sheet: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Chroma-key green pixels using fast direct byte slice manipulation
	for i := 0; i < len(rgba.Pix); i += 4 {
		ru := rgba.Pix[i]
		gu := rgba.Pix[i+1]
		bu := rgba.Pix[i+2]

		if gu > 140 && ru < 100 && bu < 100 {
			rgba.Pix[i] = 0
			rgba.Pix[i+1] = 0
			rgba.Pix[i+2] = 0
			rgba.Pix[i+3] = 0
		}
	}

	scoutSubSheet = ebiten.NewImageFromImage(rgba)
}

func loadSkiffSheet() {
	img, _, err := image.Decode(bytes.NewReader(assets.SkiffPNG))
	if err != nil {
		log.Printf("Error: Failed to decode skiff sheet: %v", err)
		return
	}

	skiffSheet = ebiten.NewImageFromImage(img)
}
