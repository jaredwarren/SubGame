package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestCombinedInput_MergesUIScrollWheel(t *testing.T) {
	touch := NewTouchControls()
	touch.active = true
	touch.uiScrollWheelY = 2.0 // finger moved +30px → 30/15

	ci := NewCombinedInput(NewEbitenInput(), touch)
	_, wy := ci.Wheel()
	if wy != 2.0 {
		t.Fatalf("expected synthesized wheel Y 2.0, got %f", wy)
	}
}

func TestUIDragWheelScale_MatchesMenuScroll(t *testing.T) {
	// Menu applies ScrollY -= wy * 15. Drag converts pixels with the same scale
	// so a 30px finger move changes ScrollY by exactly 30.
	const fingerDY = 30.0
	wy := fingerDY / uiDragWheelScale
	scrollDelta := -wy * 15
	if scrollDelta != -fingerDY {
		t.Fatalf("expected 1:1 pixel scroll, got delta %f for finger %f", scrollDelta, fingerDY)
	}
}

func TestTouchControls_CaveHasSprintButton(t *testing.T) {
	touch := NewTouchControls()
	var sprint *touchButton
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyShift && b.visibleIn(TouchContextCave) {
			sprint = b
			break
		}
	}
	if sprint == nil {
		t.Fatal("expected a Shift (sprint) button visible in cave touch context")
	}
	if sprint.visibleIn(TouchContextOnFoot) {
		t.Fatal("cave sprint button should not appear on overworld on-foot controls")
	}
}

func TestTouchControls_QuestButtonInCaveOnly(t *testing.T) {
	touch := NewTouchControls()

	hasQuestOnFoot := false
	hasQuestCave := false
	hasQuestDriving := false
	hasQuestCaveDriving := false

	for _, b := range touch.buttons {
		if b.key == ebiten.KeyJ {
			if b.visibleIn(TouchContextOnFoot) {
				hasQuestOnFoot = true
			}
			if b.visibleIn(TouchContextCave) {
				hasQuestCave = true
			}
			if b.visibleIn(TouchContextDriving) {
				hasQuestDriving = true
			}
			if b.visibleIn(TouchContextCaveDriving) {
				hasQuestCaveDriving = true
			}
		}
	}

	if hasQuestOnFoot {
		t.Fatal("Quest button (KeyJ) should not be directly visible in TouchContextOnFoot (accessed via Map)")
	}
	if !hasQuestCave {
		t.Fatal("expected Quest/PDA button (KeyJ) visible in TouchContextCave")
	}
	if hasQuestDriving {
		t.Fatal("Quest button (KeyJ) should not be directly visible in TouchContextDriving")
	}
	if !hasQuestCaveDriving {
		t.Fatal("expected Quest/PDA button (KeyJ) visible in TouchContextCaveDriving")
	}
}

func TestTouchControls_EnterVehicleButton(t *testing.T) {
	touch := NewTouchControls()

	// When canEnterVehicle is false, no KeyF button should be visible in onFoot or cave.
	touch.SetCanEnterVehicle(false)
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyF {
			if b.visibleIn(TouchContextOnFoot) {
				t.Fatal("Enter Vehicle button should not be visible in TouchContextOnFoot when not near vehicle")
			}
			if b.visibleIn(TouchContextCave) {
				t.Fatal("Enter Vehicle button should not be visible in TouchContextCave when not near vehicle")
			}
		}
	}

	// When canEnterVehicle is true, an Enter Vehicle button should be visible in onFoot and cave.
	touch.SetCanEnterVehicle(true)
	var onFootEnter, caveEnter *touchButton
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyF && b.cx == 1108 && b.cy == 478 {
			if b.visibleIn(TouchContextOnFoot) {
				onFootEnter = b
			}
			if b.visibleIn(TouchContextCave) {
				caveEnter = b
			}
		}
	}
	if onFootEnter == nil {
		t.Fatal("expected Enter Vehicle button visible in TouchContextOnFoot when near vehicle")
	}
	if caveEnter == nil {
		t.Fatal("expected Enter Vehicle button visible in TouchContextCave when near vehicle")
	}
}

func TestTouchControls_MapButtonContexts(t *testing.T) {
	touch := NewTouchControls()

	hasMapOnFoot := false
	hasMapCave := false
	hasMapDriving := false
	hasMapCaveDriving := false

	for _, b := range touch.buttons {
		if b.key == ebiten.KeyM {
			if b.visibleIn(TouchContextOnFoot) {
				hasMapOnFoot = true
			}
			if b.visibleIn(TouchContextCave) {
				hasMapCave = true
			}
			if b.visibleIn(TouchContextDriving) {
				hasMapDriving = true
			}
			if b.visibleIn(TouchContextCaveDriving) {
				hasMapCaveDriving = true
			}
		}
	}

	if !hasMapOnFoot {
		t.Fatal("expected Map button visible in TouchContextOnFoot")
	}
	if hasMapCave {
		t.Fatal("Map button should NOT be visible in TouchContextCave")
	}
	if !hasMapDriving {
		t.Fatal("expected Map button visible in TouchContextDriving (overworld)")
	}
	if hasMapCaveDriving {
		t.Fatal("Map button should NOT be visible in TouchContextCaveDriving")
	}
}

func TestTouchControls_VehicleCapabilitiesButtons(t *testing.T) {
	touch := NewTouchControls()

	// Default / disabled: neither sonar (KeyQ) nor special (KeySpace) should be visible.
	touch.SetVehicleCapabilities(false, false)
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyQ && b.visibleIn(TouchContextDriving) {
			t.Fatal("Sonar button should NOT be visible when vehicle lacks sonar upgrade")
		}
		if b.key == ebiten.KeySpace && b.visibleIn(TouchContextDriving) {
			t.Fatal("Special button should NOT be visible when vehicle lacks special upgrade")
		}
	}

	// Sonar only enabled.
	touch.SetVehicleCapabilities(true, false)
	foundSonar := false
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyQ && b.visibleIn(TouchContextDriving) {
			foundSonar = true
		}
		if b.key == ebiten.KeySpace && b.visibleIn(TouchContextDriving) {
			t.Fatal("Special button should NOT be visible when only sonar is enabled")
		}
	}
	if !foundSonar {
		t.Fatal("expected Sonar button visible when sonar capability is true")
	}

	// Special only enabled.
	touch.SetVehicleCapabilities(false, true)
	foundSpecial := false
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyQ && b.visibleIn(TouchContextDriving) {
			t.Fatal("Sonar button should NOT be visible when only special is enabled")
		}
		if b.key == ebiten.KeySpace && b.visibleIn(TouchContextDriving) {
			foundSpecial = true
		}
	}
	if !foundSpecial {
		t.Fatal("expected Special button visible when special capability is true")
	}

	// Both enabled.
	touch.SetVehicleCapabilities(true, true)
	foundSonar = false
	foundSpecial = false
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyQ && b.visibleIn(TouchContextDriving) {
			foundSonar = true
		}
		if b.key == ebiten.KeySpace && b.visibleIn(TouchContextDriving) {
			foundSpecial = true
		}
	}
	if !foundSonar || !foundSpecial {
		t.Fatal("expected both Sonar and Special buttons visible when both capabilities are true")
	}
}

func TestTouchControls_LifePodFabricatorButton(t *testing.T) {
	touch := NewTouchControls()

	// When canEnterLifePod is false, no Life Pod button should be visible.
	touch.SetCanEnterLifePod(false)
	for _, b := range touch.buttons {
		if b.cx == 1000 && b.cy == 460 && b.key == ebiten.KeyE {
			if b.visibleIn(TouchContextOnFoot) {
				t.Fatal("Life Pod button should NOT be visible in TouchContextOnFoot when not near Life Pod")
			}
		}
	}

	// When canEnterLifePod is true, Life Pod button should be visible in TouchContextOnFoot only.
	touch.SetCanEnterLifePod(true)
	var podBtn *touchButton
	for _, b := range touch.buttons {
		if b.cx == 1000 && b.cy == 460 && b.key == ebiten.KeyE {
			if b.visibleIn(TouchContextOnFoot) {
				podBtn = b
			}
			if b.visibleIn(TouchContextCave) {
				t.Fatal("Life Pod button should NOT be visible in TouchContextCave")
			}
			if b.visibleIn(TouchContextDriving) {
				t.Fatal("Life Pod button should NOT be visible in TouchContextDriving")
			}
		}
	}
	if podBtn == nil {
		t.Fatal("expected Life Pod button visible in TouchContextOnFoot when near Life Pod")
	}
}



