package assets

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	rawassets "github.com/jaredwarren/SubGame/assets"
)

type Option func(*loadOptions)

type loadOptions struct {
	trim bool
}

func WithTrim() Option {
	return func(o *loadOptions) {
		o.trim = true
	}
}

// LoadChromaKeyedImage decodes the image, chroma-keys it,
// and optionally trims empty space.
func LoadChromaKeyedImage(name string, opts ...Option) (*ebiten.Image, error) {
	var data []byte
	switch name {
	case "heavy_mech":
		data = rawassets.HeavyMechPNG
	case "scout_sub":
		data = rawassets.ScoutSubPNG
	case "skiff":
		data = rawassets.SkiffPNG
	case "diver_sheet":
		data = rawassets.DiverSheetPNG
	case "item_icons":
		data = rawassets.ItemIconsPNG
	case "lifepod_surface":
		data = rawassets.LifepodSurfacePNG
	case "ore_sheet":
		data = rawassets.OreSheetPNG
	case "overworld_water":
		data = rawassets.OverworldWaterPNG
	case "trench_surface":
		data = rawassets.TrenchSurfacePNG
	case "wreckage_surface":
		data = rawassets.WreckageSurfacePNG
	default:
		return nil, fmt.Errorf("unknown asset: %s", name)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	var config loadOptions
	for _, opt := range opts {
		opt(&config)
	}

	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for i := 0; i < len(rgba.Pix); i += 4 {
		ru := rgba.Pix[i]
		gu := rgba.Pix[i+1]
		bu := rgba.Pix[i+2]

		if gu > 140 && ru < 100 && bu < 100 {
			rgba.Pix[i] = 0
			rgba.Pix[i+1] = 0
			rgba.Pix[i+2] = 0
			rgba.Pix[i+3] = 0
		} else if config.trim {
			pixelIndex := i / 4
			x := pixelIndex % bounds.Dx()
			y := pixelIndex / bounds.Dx()
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	var finalImg *ebiten.Image
	if config.trim && maxX >= minX && maxY >= minY {
		subRect := image.Rect(minX, minY, maxX+1, maxY+1)
		finalImg = ebiten.NewImageFromImage(rgba.SubImage(subRect))
	} else {
		finalImg = ebiten.NewImageFromImage(rgba)
	}

	return finalImg, nil
}
