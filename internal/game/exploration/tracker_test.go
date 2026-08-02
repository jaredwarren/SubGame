package exploration

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/world"
)

func TestRevealCircleAndBounds(t *testing.T) {
	tr := NewTracker(20, 20)
	tr.Reveal(0, 0, 3)

	if !tr.IsExplored(0, 0) {
		t.Fatal("center should be explored")
	}
	if !tr.IsExplored(2, 2) {
		t.Fatal("(2,2) is within radius 3")
	}
	if tr.IsExplored(3, 3) {
		t.Fatal("(3,3) is outside radius 3 (dist^2=18 > 9)")
	}
	// Out of bounds never explored
	if tr.IsExplored(-1, 0) || tr.IsExplored(0, -1) || tr.IsExplored(20, 0) {
		t.Fatal("out-of-bounds tiles must not be explored")
	}

	// Reveal near far edge should clamp, not panic
	tr.Reveal(19, 19, 5)
	if !tr.IsExplored(19, 19) {
		t.Fatal("edge center should be explored")
	}
}

func TestRevealSkipsSameTile(t *testing.T) {
	tr := NewTracker(50, 50)
	tr.Reveal(10, 10, 2)
	first := len(tr.Drain())
	if first == 0 {
		t.Fatal("expected newly revealed tiles")
	}
	tr.Reveal(10, 10, 2)
	if len(tr.Drain()) != 0 {
		t.Fatal("same-tile Reveal should not re-append")
	}
	tr.Reveal(11, 10, 2)
	if len(tr.Drain()) == 0 {
		t.Fatal("moving to a new tile should reveal the leading edge")
	}
}

func TestVisitedAndSerializeRoundTrip(t *testing.T) {
	tr := NewTracker(30, 30)
	tr.Reveal(5, 5, 4)
	tr.MarkVisited(5, 5, world.TileTrench)
	tr.MarkVisited(-1, 0, world.TileWreckage) // void dive — ignored

	if !tr.IsVisited(5, 5) {
		t.Fatal("expected visited trench")
	}
	if tr.IsVisited(6, 6) {
		t.Fatal("unvisited site")
	}

	saved := tr.SerializeState()
	tr2 := NewTracker(30, 30)
	tr2.DeserializeState(saved)

	if !tr2.IsExplored(5, 5) {
		t.Fatal("deserialize lost explored tile")
	}
	if !tr2.IsVisited(5, 5) {
		t.Fatal("deserialize lost visited site")
	}
	if tt, ok := tr2.VisitedTile(5, 5); !ok || tt != world.TileTrench {
		t.Fatalf("visited tile type = %v, want TileTrench", tt)
	}
	if tr2.ExploredFraction() != tr.ExploredFraction() {
		t.Fatalf("explored fraction mismatch: got %v want %v", tr2.ExploredFraction(), tr.ExploredFraction())
	}
}
