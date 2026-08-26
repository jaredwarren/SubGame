package cave

import "testing"

func TestCaveBuilderRegistry(t *testing.T) {
	kinds := AllKinds()
	if len(kinds) < 8 {
		t.Fatalf("expected at least 8 cave kinds, got %d", len(kinds))
	}
	for _, k := range kinds {
		if Kind(k.ID()) != k {
			t.Fatalf("Kind(%q) did not round-trip", k.ID())
		}
		c, err := BuildKind(k.ID(), DefaultShallowReefBiome, 42)
		if err != nil {
			t.Fatalf("BuildKind(%q): %v", k.ID(), err)
		}
		if c == nil {
			t.Fatalf("BuildKind(%q) returned nil cave", k.ID())
		}
		if k.ID() != "void" && len(c.GetGrid()) == 0 {
			t.Fatalf("BuildKind(%q): expected non-empty grid", k.ID())
		}
		if k.ID() == "void" && len(c.GetGrid()) != 0 {
			t.Fatalf("BuildKind(void): expected empty grid")
		}
	}
}

func TestBuildKindUnknown(t *testing.T) {
	_, err := BuildKind("nope", nil, 1)
	if err == nil {
		t.Fatal("expected error for unknown kind")
	}
}
