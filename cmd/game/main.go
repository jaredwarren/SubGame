package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

func main() {
	// Configure Ebitengine window options
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(config.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Create and start the game loop
	g := game.NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
