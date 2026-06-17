package scene

import (
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

// Ingredient represents an item constructor and quantity required for a recipe.
type Ingredient struct {
	NewItem  func() item.Item
	Quantity int
}

// Recipe defines ingredients needed to craft a target item.
type Recipe struct {
	NewResult      func() item.Item
	ResultQuantity int
	Ingredients    []Ingredient
	Tier           int
	Unlocked       bool
}

// CraftingRecipes is the global list of craftable item upgrades.
var CraftingRecipes = []Recipe{
	{
		NewResult: func() item.Item { return &item.O2TankHC{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.O2TankUHC{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.O2TankHC{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.Fins{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.Scanner{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &vehicle.SkiffKit{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 10},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &vehicle.ScoutSubKit{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &vehicle.HeavyMechKit{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 8},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
		Tier:     2,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.UpgradeSolar{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.UpgradeSolarMKII{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.UpgradeSolar{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 6},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.UpgradeStorage{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.UpgradeStorageMKII{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.UpgradeStorage{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.UpgradeSolar{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 3},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.SonarAmplifier{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.PowerCell{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult: func() item.Item { return &item.ThermalGenerator{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 4},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult: func() item.Item { return &item.EscapeRocket{} },
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.AbyssalOre{} }, Quantity: 10},
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 10},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 5},
		},
		Tier:     2,
		Unlocked: false,
	},
	{
		NewResult:      func() item.Item { return &item.CookedFish{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.RawFish{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.CookedCrab{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.RawCrab{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.Titanium{} },
		ResultQuantity: 2,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ScrapMetal{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.Copper{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ElectronicWaste{} }, Quantity: 1},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.SonicDecoy{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.ElectronicWaste{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 2},
		},
		Tier:     0,
		Unlocked: true,
	},
	{
		NewResult:      func() item.Item { return &item.ChemicalDeterrent{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 1},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult:      func() item.Item { return &item.DecoyLauncher{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Copper{} }, Quantity: 3},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 2},
		},
		Tier:     1,
		Unlocked: false,
	},
	{
		NewResult:      func() item.Item { return &item.ChemicalDischarger{} },
		ResultQuantity: 1,
		Ingredients: []Ingredient{
			{NewItem: func() item.Item { return &item.Titanium{} }, Quantity: 5},
			{NewItem: func() item.Item { return &item.Quartz{} }, Quantity: 4},
		},
		Tier:     1,
		Unlocked: false,
	},
}

// DefaultCraftingRecipes returns a fresh copy of the default CraftingRecipes slice.
func DefaultCraftingRecipes() []Recipe {
	recipes := make([]Recipe, len(CraftingRecipes))
	for i, rcp := range CraftingRecipes {
		recipes[i] = rcp
	}
	return recipes
}
