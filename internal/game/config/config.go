package config

// Configuration constants for the game
const (
	ScreenWidth  = 1280
	ScreenHeight = 720
	Title        = "2D Subnautica-Inspired Game"
	TileSize     = 64 // Size of each tile in pixels

	// Debugging settings
	LightCaveForDebug = false // Set to true to reveal the caves for debugging; false for darkness
)

var (
	// FlashlightFollowsMouse controls the cave flashlight aiming mode on desktop:
	// true  -> Flashlight follows the mouse cursor.
	// false -> Flashlight only follows player movement direction / key inputs.
	FlashlightFollowsMouse = false
)
