package scene

import "github.com/jaredwarren/SubGame/internal/game/data"

// Crafting recipe tables are owned by package data. Aliases preserve existing call sites.

type (
	Ingredient = data.Ingredient
	Recipe     = data.Recipe
)

// CraftingRecipes is the global list of craftable item upgrades.
var CraftingRecipes = data.CraftingRecipes

// DefaultCraftingRecipes returns a fresh copy of the default CraftingRecipes slice.
func DefaultCraftingRecipes() []Recipe {
	return data.DefaultCraftingRecipes()
}
