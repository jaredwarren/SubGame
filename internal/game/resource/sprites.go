package resource

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/assets"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

var (
	TitaniumSprite *ebiten.Image
	CopperSprite   *ebiten.Image
	QuartzSprite   *ebiten.Image
	AbyssalSprite  *ebiten.Image
	spritesLoaded  bool
)

// LoadAssets preloads and chroma-keys all resource crystal sprites.
func LoadAssets() {
	sheet, err := assets.LoadChromaKeyedImage("ore_sheet")
	if err != nil {
		log.Printf("Error: Failed to load ore sheet: %v", err)
		return
	}

	bounds := sheet.Bounds()

	if bounds.Dx() == 3584 && bounds.Dy() == 1184 {
		// Specific coordinate slice for the user's high-res generated ore sheet
		TitaniumSprite = sheet.SubImage(image.Rect(832, 330, 1312, 856)).(*ebiten.Image)
		CopperSprite = sheet.SubImage(image.Rect(1327, 330, 1809, 856)).(*ebiten.Image)
		QuartzSprite = sheet.SubImage(image.Rect(1827, 330, 2308, 856)).(*ebiten.Image)
		AbyssalSprite = sheet.SubImage(image.Rect(2328, 330, 2790, 856)).(*ebiten.Image)
	} else {
		// General fallback: slice into 4 equal columns
		w := bounds.Dx() / 4
		h := bounds.Dy()
		TitaniumSprite = sheet.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
		CopperSprite = sheet.SubImage(image.Rect(w, 0, w*2, h)).(*ebiten.Image)
		QuartzSprite = sheet.SubImage(image.Rect(w*2, 0, w*3, h)).(*ebiten.Image)
		AbyssalSprite = sheet.SubImage(image.Rect(w*3, 0, w*4, h)).(*ebiten.Image)
	}
	spritesLoaded = true
}

// drawNodeBase renders the shared backing block behind all resource nodes.
func drawNodeBase(screen *ebiten.Image, tx, ty int, camX, camY float64) (float32, float32) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))
	vector.FillRect(screen, sx, sy, TileSize, TileSize, color.RGBA{25, 22, 30, 255}, false)
	vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, color.RGBA{45, 40, 52, 255}, false)
	return sx, sy
}

// drawMineral renders the mineral based on its type and attachment direction.
func drawMineral(screen *ebiten.Image, tx, ty int, camX, camY float64, hitsToMine int, mineralColor, coreColor color.Color, attachDir AttachDirection, mineralName string) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))

	var cx, cy float32
	var dirVec, perpVec [2]float32

	switch attachDir {
	case AttachTop:
		cx = sx + float32(TileSize)/2.0
		cy = sy
		dirVec = [2]float32{0, 1}
		perpVec = [2]float32{-1, 0}
	case AttachLeft:
		cx = sx
		cy = sy + float32(TileSize)/2.0
		dirVec = [2]float32{1, 0}
		perpVec = [2]float32{0, -1}
	case AttachRight:
		cx = sx + float32(TileSize)
		cy = sy + float32(TileSize)/2.0
		dirVec = [2]float32{-1, 0}
		perpVec = [2]float32{0, 1}
	case AttachBottom:
		cx = sx + float32(TileSize)/2.0
		cy = sy + float32(TileSize)
		dirVec = [2]float32{0, -1}
		perpVec = [2]float32{1, 0}
	default: // AttachNone (fallback / icon center)
		cx = sx + float32(TileSize)/2.0
		cy = sy + float32(TileSize)/2.0 + 8.0 // offset slightly down so growing up centers it
		dirVec = [2]float32{0, -1}
		perpVec = [2]float32{1, 0}
	}

	scale := float32(hitsToMine) / 3.0
	if scale < 0.35 {
		scale = 0.35
	}

	item.DrawMineralShape(screen, cx, cy, dirVec, perpVec, scale, mineralColor, coreColor, mineralName)
}

func drawMineralIcon(screen *ebiten.Image, cx, cy, size float32, mineralColor, coreColor color.Color, mineralName string) {
	item.DrawMineralIcon(screen, cx, cy, size, mineralColor, coreColor, mineralName)
}

func drawCracks(screen *ebiten.Image, sx, sy float32, hitsToMine int) {
	if hitsToMine >= 3 {
		// No cracks at full health (3 hits remaining)
		return
	}

	// Crack color: dark charcoal/black to represent fracture lines overlaying the mineral
	crackColor := color.RGBA{15, 15, 20, 235}

	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0

	// Stage 1 (1 hit taken, 2 remaining): draw primary cracks radiating from center
	if hitsToMine <= 2 {
		vector.StrokeLine(screen, cx-14, cy-12, cx-3, cy-2, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx-3, cy-2, cx+12, cy-14, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx-3, cy-2, cx-6, cy+14, 3.0, crackColor, false)
	}

	// Stage 2 (2 hits taken, 1 remaining): add detailed, longer branching fractures
	if hitsToMine <= 1 {
		// Branches from Stage 1 cracks
		vector.StrokeLine(screen, cx-14, cy-12, cx-22, cy-8, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx+12, cy-14, cx+20, cy-20, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx-6, cy+14, cx-14, cy+22, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx-6, cy+14, cx+8, cy+12, 2.0, crackColor, false)

		// Second independent fracture cluster
		vector.StrokeLine(screen, cx+14, cy+4, cx+3, cy-8, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx+3, cy-8, cx-8, cy-5, 3.0, crackColor, false)
	}
}

func drawNodeSprite(screen *ebiten.Image, tx, ty int, camX, camY float64, sprite *ebiten.Image, hitsToMine int) bool {
	if sprite == nil {
		return false
	}
	sx := float64(tx*TileSize - int(camX))
	sy := float64(ty*TileSize - int(camY))

	// Draw the node background block under the sprite
	drawNodeBase(screen, tx, ty, camX, camY)

	op := &ebiten.DrawImageOptions{}

	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())

	// Scale the sprite to match the full TileSize (64x64) without shrinking
	baseScaleX := float64(TileSize) / spriteW
	baseScaleY := float64(TileSize) / spriteH

	// Center the sprite on the origin (0,0) before scaling
	op.GeoM.Translate(-spriteW/2.0, -spriteH/2.0)
	// Scale to full tile size
	op.GeoM.Scale(baseScaleX, baseScaleY)
	// Translate to screen tile coordinates + center offset
	op.GeoM.Translate(sx+float64(TileSize)/2.0, sy+float64(TileSize)/2.0)

	screen.DrawImage(sprite, op)

	// Draw overlay cracks representing node damage state
	drawCracks(screen, float32(sx), float32(sy), hitsToMine)

	return true
}

func drawNodeIconSprite(screen *ebiten.Image, cx, cy, size float32, sprite *ebiten.Image) bool {
	if sprite == nil {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())

	op.GeoM.Translate(-spriteW/2.0, -spriteH/2.0)
	op.GeoM.Scale(float64(size)/spriteW, float64(size)/spriteH)
	op.GeoM.Translate(float64(cx), float64(cy))

	screen.DrawImage(sprite, op)
	return true
}
