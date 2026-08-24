package entity

import (
	"math"

	"github.com/jaredwarren/SubGame/internal/gvec"
)

// TargetQuerier is the Runtime subset needed for chase-target acquisition.
type TargetQuerier interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
	PlayerFacing() float64
	HasActiveVehicle() bool
	ActiveVehiclePos() gvec.Vec2
	ActiveVehicleDims() gvec.Vec2
	ActiveVehicleFacing() float64
	FlashlightOn() bool
	FindClosestDecoy(pos gvec.Vec2, maxDist float64) (gvec.Vec2, bool)
}

// TargetInfo describes an acquired chase/attack target.
type TargetInfo struct {
	CenterX, CenterY   float64
	Width, Height      float64
	TopLeftX, TopLeftY float64
	IsDecoy            bool
}

// AcquireTarget picks decoy (within decoyRange) → active vehicle → player.
// decoyTargetSize is used as W/H when the target is a decoy.
func AcquireTarget(gr TargetQuerier, from gvec.Vec2, decoyRange, decoyTargetSize float64) TargetInfo {
	if decoyRange > 0 {
		if decoyPos, ok := gr.FindClosestDecoy(from, decoyRange); ok {
			return TargetInfo{
				CenterX:  decoyPos.X,
				CenterY:  decoyPos.Y,
				Width:    decoyTargetSize,
				Height:   decoyTargetSize,
				TopLeftX: decoyPos.X - decoyTargetSize/2,
				TopLeftY: decoyPos.Y - decoyTargetSize/2,
				IsDecoy:  true,
			}
		}
	}
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		vDims := gr.ActiveVehicleDims()
		return TargetInfo{
			CenterX:  vPos.X + vDims.X/2,
			CenterY:  vPos.Y + vDims.Y/2,
			Width:    vDims.X,
			Height:   vDims.Y,
			TopLeftX: vPos.X,
			TopLeftY: vPos.Y,
		}
	}
	pPos := gr.PlayerPos()
	pDims := gr.PlayerDims()
	return TargetInfo{
		CenterX:  pPos.X + pDims.X/2,
		CenterY:  pPos.Y + pDims.Y/2,
		Width:    pDims.X,
		Height:   pDims.Y,
		TopLeftX: pPos.X,
		TopLeftY: pPos.Y,
	}
}

// InFlashlightCone reports whether entity center is inside the player's (or
// active vehicle's) flashlight cone of halfAngle radians.
func InFlashlightCone(gr TargetQuerier, entityCenter gvec.Vec2, lightOrigin gvec.Vec2, halfAngle float64) bool {
	if !gr.FlashlightOn() {
		return false
	}
	facing := gr.PlayerFacing()
	if gr.HasActiveVehicle() {
		facing = gr.ActiveVehicleFacing()
	}
	dx := entityCenter.X - lightOrigin.X
	dy := entityCenter.Y - lightOrigin.Y
	angleToEnt := math.Atan2(dy, dx)
	diff := angleToEnt - facing
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}
	return math.Abs(diff) < halfAngle
}
