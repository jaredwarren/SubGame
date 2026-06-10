package game

import (
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Type aliases so existing game-package code (game.go, tests, adapters) compiles
// without modification while the canonical definitions live in the scene package.

type Scene = scene.Scene
type GameContext = scene.GameContext

type TitleScene = scene.TitleScene
type IntroScene = scene.IntroScene
type OverworldScene = scene.OverworldScene
type CaveScene = scene.CaveScene
type BaseMenuScene = scene.BaseMenuScene
type GameOverScene = scene.GameOverScene
type GameWonScene = scene.GameWonScene
type HUD = scene.HUD

// Crafting types re-exported for tests in package game.
type Recipe = scene.Recipe
type Ingredient = scene.Ingredient

// CraftingRecipes provides package-game access to the canonical list.
var CraftingRecipes = scene.CraftingRecipes

// Constructor wrappers so NewGame() and tests call unchanged function names.
func NewTitleScene() *TitleScene      { return scene.NewTitleScene() }
func NewIntroScene() *IntroScene      { return scene.NewIntroScene() }
func NewCaveScene() *CaveScene        { return scene.NewCaveScene() }
func NewBaseMenuScene() *BaseMenuScene { return scene.NewBaseMenuScene() }
func NewGameOverScene() *GameOverScene { return scene.NewGameOverScene() }
func NewGameWonScene() *GameWonScene   { return scene.NewGameWonScene() }
func NewHUD() *HUD                    { return scene.NewHUD() }

func NewOverworldScene(w *world.World) *OverworldScene {
	return scene.NewOverworldScene(w)
}

// GetOverworldLightMultiplier re-exported for any remaining references.
func GetOverworldLightMultiplier(timeOfDay float64) float64 {
	return scene.GetOverworldLightMultiplier(timeOfDay)
}
