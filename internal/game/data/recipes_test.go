package data

import "testing"

func TestLegacyRecipeIndexesUnchanged(t *testing.T) {
	// Older saves store unlocks by index. New recipes must be appended.
	want := []string{
		"High Capacity O2 Tank",
		"Ultra High Capacity O2 Tank",
		"Propulsion Fins",
		"Scanner Tool",
		"Skiff Kit",
	}
	if len(CraftingRecipes) < len(want) {
		t.Fatalf("expected at least %d recipes, got %d", len(want), len(CraftingRecipes))
	}
	for i, name := range want {
		got := CraftingRecipes[i].ResultName()
		if got != name {
			t.Errorf("recipe[%d] = %q, want %q", i, got, name)
		}
	}
}

func TestFlashlightIsEarlyGame(t *testing.T) {
	var found bool
	for _, rcp := range CraftingRecipes {
		if rcp.ResultName() != "Flashlight" {
			continue
		}
		found = true
		if rcp.Tier != 0 || !rcp.Unlocked {
			t.Errorf("Flashlight should be Tier 0 unlocked, got tier=%d unlocked=%v", rcp.Tier, rcp.Unlocked)
		}
		for _, ing := range rcp.Ingredients {
			if ing.NewItem().GetName() == "Power Cell" {
				t.Error("Flashlight should not require a Power Cell")
			}
		}
	}
	if !found {
		t.Fatal("Flashlight recipe missing")
	}
}

func TestSkiffLightRecipe(t *testing.T) {
	var found bool
	for _, rcp := range CraftingRecipes {
		if rcp.ResultName() != "Skiff Light Module" {
			continue
		}
		found = true
		if rcp.Tier != 0 || !rcp.Unlocked {
			t.Errorf("Skiff Light Module should be Tier 0 unlocked, got tier=%d unlocked=%v", rcp.Tier, rcp.Unlocked)
		}
		ingMap := make(map[string]int)
		for _, ing := range rcp.Ingredients {
			ingMap[ing.NewItem().GetName()] = ing.Quantity
		}
		if ingMap["Titanium"] != 3 {
			t.Errorf("expected 3 Titanium, got %d", ingMap["Titanium"])
		}
		if ingMap["Copper"] != 2 {
			t.Errorf("expected 2 Copper, got %d", ingMap["Copper"])
		}
		if ingMap["Quartz"] != 2 {
			t.Errorf("expected 2 Quartz, got %d", ingMap["Quartz"])
		}
	}
	if !found {
		t.Fatal("Skiff Light Module recipe missing")
	}
}
