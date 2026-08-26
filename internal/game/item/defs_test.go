package item

import "testing"

func TestCatalogHasCoreMinerals(t *testing.T) {
	for _, id := range []ItemID{IDTitanium, IDCopper, IDQuartz, IDNickel, IDTungsten, IDAbyssalOre} {
		def := Def(id)
		if def == nil {
			t.Fatalf("missing catalog entry for %q", id)
		}
		if def.ID != id {
			t.Errorf("def.ID = %q, want %q", def.ID, id)
		}
		if def.Name == "" || def.MaxStack <= 0 {
			t.Errorf("%q has empty name or non-positive MaxStack: %+v", id, def)
		}
	}
}

func TestCatalogO2Upgrade(t *testing.T) {
	def := Def(IDO2TankHC)
	if def == nil || !def.IsO2Upgrade || def.MaxO2Capacity <= 0 {
		t.Fatalf("expected O2 tank HC to be an O2 upgrade, got %+v", def)
	}
}
