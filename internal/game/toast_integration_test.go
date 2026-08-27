package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

func TestGame_AddItemToast(t *testing.T) {
	g := NewGame()

	if len(g.toasts.GetToasts()) != 0 {
		t.Fatalf("expected 0 toasts initially, got %d", len(g.toasts.GetToasts()))
	}

	tit := &item.Titanium{}
	g.AddItemToast(tit, 2)

	toasts := g.toasts.GetToasts()
	if len(toasts) != 1 {
		t.Fatalf("expected 1 toast after AddItemToast, got %d", len(toasts))
	}
	if toasts[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", toasts[0].Quantity)
	}
	if toasts[0].Item.GetID() != item.IDTitanium {
		t.Errorf("expected Titanium, got %v", toasts[0].Item.GetID())
	}
}

func TestGame_VehiclePickupToast(t *testing.T) {
	g := NewGame()

	// Deploy a skiff
	sub := vehicle.NewScoutSub(100, 100)
	g.ActiveVehicle = sub
	g.OverworldVehicles = []vehicle.Vehicle{sub}

	// Pick up vehicle
	g.PickUpActiveVehicle()

	toasts := g.toasts.GetToasts()
	if len(toasts) == 0 {
		t.Fatalf("expected toast after picking up vehicle")
	}

	kitID := toasts[0].Item.GetID()
	if kitID != item.IDScoutSubKit {
		t.Errorf("expected ScoutSubKit toast, got %v", kitID)
	}
}

func TestGame_VehicleRuntimeAddItemToastCmd(t *testing.T) {
	g := NewGame()

	rt := &vehicleRuntimeAdapter{
		g: g,
		cmds: []vehicle.GameCommand{
			vehicle.AddItemToastCmd{
				Item:     &item.Nickel{},
				Quantity: 1,
			},
		},
	}

	g.drainVehicleCommands(rt)

	toasts := g.toasts.GetToasts()
	if len(toasts) != 1 {
		t.Fatalf("expected 1 toast after drainVehicleCommands, got %d", len(toasts))
	}
	if toasts[0].Item.GetID() != item.IDNickel {
		t.Errorf("expected Nickel toast, got %v", toasts[0].Item.GetID())
	}
}

func TestGame_AdvanceTimersUpdatesToasts(t *testing.T) {
	g := NewGame()
	g.AddItemToast(&item.Quartz{}, 1)

	toasts := g.toasts.GetToasts()
	if len(toasts) != 1 {
		t.Fatalf("expected 1 toast")
	}

	initTimer := toasts[0].Timer
	g.advanceTimers()

	if toasts[0].Timer != initTimer-1 {
		t.Errorf("expected timer to decrement to %d, got %d", initTimer-1, toasts[0].Timer)
	}
}

func TestGame_DrawToastsNoPanic(t *testing.T) {
	g := NewGame()
	g.currentState = StateOverworld
	g.AddItemToast(&item.Titanium{}, 1)
	g.AddItemToast(&item.Copper{}, 2)

	screen := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	g.Draw(screen)
}
