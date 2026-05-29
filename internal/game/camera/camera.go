package camera

import "github.com/jaredwarren/SubGame/internal/gvec"

// Camera manages viewport translation and centers/tracks a target with linear interpolation.
type Camera struct {
	Pos          gvec.Vec2
	ScreenWidth  int
	ScreenHeight int
}

// NewCamera creates and returns an initialized Camera.
func NewCamera(x, y float64, screenWidth, screenHeight int) *Camera {
	return &Camera{
		Pos:          gvec.Vec2{X: x, Y: y},
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// CenterOn centers the camera viewport directly on the target coordinates instantly.
func (c *Camera) CenterOn(targetX, targetY, targetWidth, targetHeight float64) {
	c.Pos.X = targetX + targetWidth/2.0 - float64(c.ScreenWidth)/2.0
	c.Pos.Y = targetY + targetHeight/2.0 - float64(c.ScreenHeight)/2.0
}

// Track interpolates the camera position towards the target coordinates smoothly.
// lerpRate controls camera damping speed (typically between 0.05 and 0.15).
func (c *Camera) Track(targetX, targetY, targetWidth, targetHeight, lerpRate float64) {
	destX := targetX + targetWidth/2.0 - float64(c.ScreenWidth)/2.0
	destY := targetY + targetHeight/2.0 - float64(c.ScreenHeight)/2.0

	c.Pos.X += (destX - c.Pos.X) * lerpRate
	c.Pos.Y += (destY - c.Pos.Y) * lerpRate
}
