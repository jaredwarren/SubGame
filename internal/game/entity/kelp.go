package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
)

// Kelp is a decorative swaying sea plant.
type Kelp struct {
	BaseEntity
	SwayPhase float64
}

func (k *Kelp) Update(gr Runtime) {
	k.SwayPhase += 0.03
}

func (k *Kelp) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(k.Pos.X - camera.Pos.X)
	sy := float32(k.Pos.Y - camera.Pos.Y)
	sw := float32(k.Dimensions.X)
	sh := float32(k.Dimensions.Y)
	cx := sx + sw/2.0
	bottomY := sy + sh

	numSegments := int(sh / 8.0)
	if numSegments < 3 {
		numSegments = 3
	}
	segmentHeight := sh / float32(numSegments)

	lastX := cx
	lastY := bottomY

	for i := 0; i < numSegments; i++ {
		factor := float64(i+1) / float64(numSegments)
		swayOffset := float32(math.Sin(k.SwayPhase+float64(i)*0.4)) * 8.0 * float32(factor)
		nextX := cx + swayOffset
		nextY := bottomY - float32(i+1)*segmentHeight

		vector.StrokeLine(screen, lastX, lastY, nextX, nextY, 2.5-float32(factor)*1.0, color.RGBA{34, 139, 34, 255}, false)

		leafSize := 5.0 - float32(factor)*2.0
		if leafSize < 2.0 {
			leafSize = 2.0
		}
		vector.FillCircle(screen, nextX-4.0, nextY, leafSize, color.RGBA{46, 150, 60, 220}, false)
		vector.FillCircle(screen, nextX+4.0, nextY, leafSize, color.RGBA{46, 150, 60, 220}, false)

		lastX = nextX
		lastY = nextY
	}
}
