package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

func TestToastManager_Add(t *testing.T) {
	tm := NewToastManager()

	tit := &item.Titanium{}
	cop := &item.Copper{}

	// Add single
	tm.Add(tit, 1)
	if len(tm.GetToasts()) != 1 {
		t.Fatalf("expected 1 toast, got %d", len(tm.GetToasts()))
	}
	if tm.GetToasts()[0].Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", tm.GetToasts()[0].Quantity)
	}
	if tm.GetToasts()[0].Timer != ToastMaxDuration {
		t.Errorf("expected timer %d, got %d", ToastMaxDuration, tm.GetToasts()[0].Timer)
	}

	// Add same item again -> combines and resets timer
	tm.GetToasts()[0].Timer = 50 // artificially advance timer
	tm.Add(tit, 2)
	if len(tm.GetToasts()) != 1 {
		t.Fatalf("expected still 1 toast after duplicate add, got %d", len(tm.GetToasts()))
	}
	if tm.GetToasts()[0].Quantity != 3 {
		t.Errorf("expected combined quantity 3, got %d", tm.GetToasts()[0].Quantity)
	}
	if tm.GetToasts()[0].Timer != ToastMaxDuration {
		t.Errorf("expected refreshed timer %d, got %d", ToastMaxDuration, tm.GetToasts()[0].Timer)
	}

	// Add second item -> stacks
	tm.Add(cop, 1)
	if len(tm.GetToasts()) != 2 {
		t.Fatalf("expected 2 toasts, got %d", len(tm.GetToasts()))
	}
	// Newer item at top
	if tm.GetToasts()[0].Item.GetID() != cop.GetID() {
		t.Errorf("expected copper at index 0, got %v", tm.GetToasts()[0].Item.GetID())
	}
}

func TestToastManager_MaxCapacity(t *testing.T) {
	tm := NewToastManager()

	items := []item.Item{
		&item.Titanium{},
		&item.Copper{},
		&item.Quartz{},
		&item.Nickel{},
		&item.Tungsten{},
		&item.AbyssalOre{},
	}

	for _, it := range items {
		tm.Add(it, 1)
	}

	if len(tm.GetToasts()) != ToastMaxActive {
		t.Fatalf("expected %d toasts max, got %d", ToastMaxActive, len(tm.GetToasts()))
	}

	// Newest item should be at the front
	if tm.GetToasts()[0].Item.GetID() != item.IDAbyssalOre {
		t.Errorf("expected newest item (AbyssalOre) at index 0, got %v", tm.GetToasts()[0].Item.GetID())
	}
}

func TestToastManager_UpdateExpiration(t *testing.T) {
	tm := NewToastManager()
	tm.Add(&item.Titanium{}, 1)

	// Artificially set timer to 2
	tm.GetToasts()[0].Timer = 2

	tm.Update()
	if len(tm.GetToasts()) != 1 || tm.GetToasts()[0].Timer != 1 {
		t.Fatalf("expected 1 toast with timer 1 after update")
	}

	tm.Update()
	if len(tm.GetToasts()) != 0 {
		t.Fatalf("expected toast to expire and be removed, got %d toasts", len(tm.GetToasts()))
	}
}

func TestToastManager_Draw(t *testing.T) {
	tm := NewToastManager()
	screen := ebiten.NewImage(640, 480)

	// Draw empty
	tm.Draw(screen)

	// Draw with items
	tm.Add(&item.Titanium{}, 1)
	tm.Add(&item.Copper{}, 3)
	tm.Draw(screen)

	// Draw in fade-out state
	tm.GetToasts()[0].Timer = 10
	tm.Draw(screen)
}
