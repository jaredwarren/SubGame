package game

// Camera manages viewport translation and centers/tracks a target with linear interpolation.
type Camera struct {
	X, Y float64
}

// NewCamera creates and returns an initialized Camera.
func NewCamera(x, y float64) *Camera {
	return &Camera{
		X: x,
		Y: y,
	}
}

// CenterOn centers the camera viewport directly on the target coordinates instantly.
func (c *Camera) CenterOn(targetX, targetY, targetWidth, targetHeight float64) {
	c.X = targetX + targetWidth/2.0 - float64(ScreenWidth)/2.0
	c.Y = targetY + targetHeight/2.0 - float64(ScreenHeight)/2.0
}

// Track interpolates the camera position towards the target coordinates smoothly.
// lerpRate controls camera damping speed (typically between 0.05 and 0.15).
func (c *Camera) Track(targetX, targetY, targetWidth, targetHeight, lerpRate float64) {
	destX := targetX + targetWidth/2.0 - float64(ScreenWidth)/2.0
	destY := targetY + targetHeight/2.0 - float64(ScreenHeight)/2.0

	c.X += (destX - c.X) * lerpRate
	c.Y += (destY - c.Y) * lerpRate
}
