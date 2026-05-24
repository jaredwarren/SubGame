package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game"
)

func main() {
	// Configure Ebitengine window options
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle(game.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Create and start the game loop
	g := game.NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
