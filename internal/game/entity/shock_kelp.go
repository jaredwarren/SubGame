package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// ShockKelp is a purple kelp variant that shocks the player on contact.
type ShockKelp struct {
	BaseEntity
	SwayPhase     float64
	ShockCooldown int
}

func (k *ShockKelp) Update(gr Runtime) {
	k.SwayPhase += 0.035 // Sway slightly faster/more erratically than regular kelp

	if k.ShockCooldown > 0 {
		k.ShockCooldown--
	}

	if k.ShockCooldown == 0 {
		vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
		targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
		hasVehicle := gr.HasActiveVehicle()
		if hasVehicle {
			vPos := gr.ActiveVehiclePos()
			targetX, targetY = vPos.X, vPos.Y
			vDims := gr.ActiveVehicleDims()
			vWidth, vHeight = vDims.X, vDims.Y
		}

		if rectsOverlap(k.Pos.X, k.Pos.Y, k.Dimensions.X, k.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
			k.ShockCooldown = 80 // Cooldown in frames

			// Push away from the kelp stalk horizontally
			kelpCenterX := k.Pos.X + k.Dimensions.X/2.0
			targetCenterX := targetX + vWidth/2.0
			dirX := 1.0
			if targetCenterX < kelpCenterX {
				dirX = -1.0
			}

			forceVec := gvec.Vec2{X: dirX * 4.5, Y: -2.5}

			if hasVehicle {
				gr.Emit(DamageActiveVehicleCmd{Amount: 12.0})
				gr.Emit(KnockbackActiveVehicleCmd{Force: forceVec})
				gr.Emit(SetMineWarningCmd{
					Message:  "VEHICLE SHOCKED BY PURPLE KELP!",
					Duration: 90,
					Level:    2,
				})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: 8.0})
				gr.Emit(KnockbackPlayerCmd{Force: forceVec})
				gr.Emit(SetMineWarningCmd{
					Message:  "SHOCKED BY PURPLE KELP!",
					Duration: 90,
					Level:    2,
				})
			}
		}
	}
}

func (k *ShockKelp) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
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

		// Draw deep purple stalk
		vector.StrokeLine(screen, lastX, lastY, nextX, nextY, 2.5-float32(factor)*1.0, color.RGBA{110, 30, 160, 255}, false)

		// Draw violet/magenta leaves
		leafSize := 5.0 - float32(factor)*2.0
		if leafSize < 2.0 {
			leafSize = 2.0
		}
		vector.FillCircle(screen, nextX-4.0, nextY, leafSize, color.RGBA{170, 50, 220, 220}, false)
		vector.FillCircle(screen, nextX+4.0, nextY, leafSize, color.RGBA{170, 50, 220, 220}, false)

		// Draw glowing electric cyan/yellow tip at the very top segment
		if i == numSegments-1 {
			pulse := float32(math.Sin(timeOfDay*0.1+k.SwayPhase)) * 1.5
			vector.FillCircle(screen, nextX, nextY, 6.0+pulse, color.RGBA{0, 230, 255, 100}, false)
			vector.FillCircle(screen, nextX, nextY, 3.5, color.RGBA{255, 255, 200, 255}, false)
		}

		lastX = nextX
		lastY = nextY
	}

	// Draw animated electrical sparks discharging from the tip
	if k.ShockCooldown > 50 {
		for s := 0; s < 3; s++ {
			rx := lastX + float32(rand.Intn(24)-12)
			ry := lastY + float32(rand.Intn(24)-12)
			vector.StrokeLine(screen, lastX, lastY, rx, ry, 1.0, color.RGBA{160, 220, 255, 255}, false)
		}
	} else if rand.Float64() < 0.08 {
		rx := lastX + float32(rand.Intn(16)-8)
		ry := lastY + float32(rand.Intn(16)-8)
		vector.StrokeLine(screen, lastX, lastY, rx, ry, 0.8, color.RGBA{160, 220, 255, 180}, false)
	}
}
