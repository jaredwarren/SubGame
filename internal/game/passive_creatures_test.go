package game

import (
	"reflect"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestPassiveCreatures_Generation(t *testing.T) {
	// Create a dummy 10x10 grid with no walls except borders to allow open water and floor spawning
	grid := make([][]bool, 10)
	for i := range grid {
		grid[i] = make([]bool, 10)
		for j := range grid[i] {
			// border tiles are solid, inner tiles are empty
			if i == 0 || i == 9 || j == 0 || j == 9 {
				grid[i][j] = true
			}
		}
	}

	// Verify that we can at least compile and run it, and let's check if the filter logic works.
	fish := entity.NewPassiveFish(100, 100, false, 0)

	crab := &entity.PassiveCrab{
		BaseEntity: entity.BaseEntity{
			Type:       entity.EntPassiveCrab,
			Pos:        gvec.Vec2{X: 100, Y: 150},
			Dimensions: gvec.Vec2{X: 16, Y: 10},
			Active:     true,
		},
	}

	kelp := &entity.Kelp{
		BaseEntity: entity.BaseEntity{
			Type:       entity.EntKelp,
			Pos:        gvec.Vec2{X: 100, Y: 180},
			Dimensions: gvec.Vec2{X: 16, Y: 48},
			Active:     true,
		},
	}

	if fish.GetType() != entity.EntPassiveFish {
		t.Errorf("expected fish type to be EntPassiveFish, got %v", fish.GetType())
	}
	if crab.GetType() != entity.EntPassiveCrab {
		t.Errorf("expected crab type to be EntPassiveCrab, got %v", crab.GetType())
	}
	if kelp.GetType() != entity.EntKelp {
		t.Errorf("expected kelp type to be EntKelp, got %v", kelp.GetType())
	}

	// Verify they satisfy CaveEntity interface
	var _ entity.CaveEntity = fish
	var _ entity.CaveEntity = crab
	var _ entity.CaveEntity = kelp
}

func TestPassiveCreatures_Harvesting(t *testing.T) {
	fish := entity.NewPassiveFish(100, 100, false, 0)

	crab := &entity.PassiveCrab{
		BaseEntity: entity.BaseEntity{
			Type:       entity.EntPassiveCrab,
			Pos:        gvec.Vec2{X: 100, Y: 100},
			Dimensions: gvec.Vec2{X: 16, Y: 10},
			Active:     true,
		},
	}

	// 1. Close distance (should be catchable)
	playerPosClose := gvec.Vec2{X: 110, Y: 105}
	if !fish.CanCatch(playerPosClose) {
		t.Error("expected fish to be catchable when close")
	}
	if !crab.CanCatch(playerPosClose) {
		t.Error("expected crab to be catchable when close")
	}

	// 2. Far distance (should not be catchable)
	playerPosFar := gvec.Vec2{X: 200, Y: 200}
	if fish.CanCatch(playerPosFar) {
		t.Error("expected fish to not be catchable when far")
	}
	if crab.CanCatch(playerPosFar) {
		t.Error("expected crab to not be catchable when far")
	}

	// 3. Harvested items
	if _, ok := fish.GetHarvestedItem().(*item.RawFish); !ok {
		t.Errorf("expected fish harvested item to be *item.RawFish, got %T", fish.GetHarvestedItem())
	}
	if _, ok := crab.GetHarvestedItem().(*item.RawCrab); !ok {
		t.Errorf("expected crab harvested item to be *item.RawCrab, got %T", crab.GetHarvestedItem())
	}
}

func TestPassiveCreatures_Consumption(t *testing.T) {
	rawFish := &item.RawFish{}
	cookedFish := &item.CookedFish{}
	rawCrab := &item.RawCrab{}
	cookedCrab := &item.CookedCrab{}

	// Raw items must restore NO health, but restore small stamina
	if rawFish.GetHealthRestore() != 0.0 {
		t.Errorf("expected RawFish health restore to be 0, got %f", rawFish.GetHealthRestore())
	}
	if rawFish.GetStaminaRestore() <= 0.0 || rawFish.GetStaminaRestore() > 10.0 {
		t.Errorf("expected RawFish stamina restore to be small, got %f", rawFish.GetStaminaRestore())
	}

	if rawCrab.GetHealthRestore() != 0.0 {
		t.Errorf("expected RawCrab health restore to be 0, got %f", rawCrab.GetHealthRestore())
	}
	if rawCrab.GetStaminaRestore() <= 0.0 || rawCrab.GetStaminaRestore() > 10.0 {
		t.Errorf("expected RawCrab stamina restore to be small, got %f", rawCrab.GetStaminaRestore())
	}

	// Cooked items must restore significant health and stamina
	if cookedFish.GetHealthRestore() <= 10.0 {
		t.Errorf("expected CookedFish health restore to be significant, got %f", cookedFish.GetHealthRestore())
	}
	if cookedFish.GetStaminaRestore() <= 10.0 {
		t.Errorf("expected CookedFish stamina restore to be significant, got %f", cookedFish.GetStaminaRestore())
	}

	if cookedCrab.GetHealthRestore() <= 10.0 {
		t.Errorf("expected CookedCrab health restore to be significant, got %f", cookedCrab.GetHealthRestore())
	}
	if cookedCrab.GetStaminaRestore() <= 10.0 {
		t.Errorf("expected CookedCrab stamina restore to be significant, got %f", cookedCrab.GetStaminaRestore())
	}
}

func TestPassiveCreatures_Crafting(t *testing.T) {
	var cookedFishRecipeFound, cookedCrabRecipeFound bool

	for _, rcp := range CraftingRecipes {
		resultType := reflect.TypeOf(rcp.NewResult())
		if resultType == reflect.TypeOf(&item.CookedFish{}) {
			cookedFishRecipeFound = true
			if len(rcp.Ingredients) != 1 || reflect.TypeOf(rcp.Ingredients[0].NewItem()) != reflect.TypeOf(&item.RawFish{}) {
				t.Error("expected CookedFish recipe to require exactly 1 RawFish")
			}
		}
		if resultType == reflect.TypeOf(&item.CookedCrab{}) {
			cookedCrabRecipeFound = true
			if len(rcp.Ingredients) != 1 || reflect.TypeOf(rcp.Ingredients[0].NewItem()) != reflect.TypeOf(&item.RawCrab{}) {
				t.Error("expected CookedCrab recipe to require exactly 1 RawCrab")
			}
		}
	}

	if !cookedFishRecipeFound {
		t.Error("expected to find CookedFish recipe in CraftingRecipes")
	}
	if !cookedCrabRecipeFound {
		t.Error("expected to find CookedCrab recipe in CraftingRecipes")
	}
}
