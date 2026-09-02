package scene

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
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

func TestCombinedInput_AimOriginAndStickFacing(t *testing.T) {
	base := NewEbitenInput()
	touch := NewTouchControls()
	ci := NewCombinedInput(base, touch)

	// Simulate touch engagement.
	touch.active = true
	touch.stickActive = true
	touch.stickVec = gvec.Vec2{X: -1, Y: 0}
	touch.lastStickDir = gvec.Vec2{X: -1, Y: 0}

	dir, ok := ci.StickFacing()
	if !ok || dir.X != -1 || dir.Y != 0 {
		t.Fatalf("expected StickFacing to return (-1, 0), got: %+v, %v", dir, ok)
	}

	if !ci.TouchActive() {
		t.Fatal("expected TouchActive to return true")
	}

	// AimOrigin near left edge of screen (e.g. player at x=50, y=300).
	playerScreen := gvec.Vec2{X: 50, Y: 300}
	ci.SetAimOrigin(playerScreen)

	cur := ci.Cursor()
	// Cursor should sit ahead of the player (50 - 120 = -70), NOT ahead of screen center (640 - 120 = 520).
	expectedX := 50.0 - 120.0
	expectedY := 300.0
	if cur.X != expectedX || cur.Y != expectedY {
		t.Fatalf("expected cursor at (%v, %v), got (%v, %v)", expectedX, expectedY, cur.X, cur.Y)
	}

	// Vector from player to cursor is pointing LEFT (negative X), NOT toward screen center.
	dx := cur.X - playerScreen.X
	if dx >= 0 {
		t.Fatalf("expected dx to be negative (pointing left), got %v", dx)
	}
}

type stubCaveContext struct {
	CaveContext
	cam camera.Camera
	p   *player.Player
}

func (s *stubCaveContext) GetCamera() *camera.Camera { return &s.cam }
func (s *stubCaveContext) GetPlayer() *player.Player { return s.p }
func (s *stubCaveContext) IsPlayerSlowed() bool      { return false }
func (s *stubCaveContext) SpawnBubble(x, y float64)  {}

func TestCavePlayerMovement_FacingNearEdges(t *testing.T) {
	cave := NewCaveScene()
	p := player.NewPlayer(50, 200)

	base := NewEbitenInput()
	touch := NewTouchControls()
	ci := NewCombinedInput(base, touch)

	// Simulate touch engagement with stick pushed LEFT (-1, 0).
	touch.active = true
	touch.stickActive = true
	touch.stickVec = gvec.Vec2{X: -1, Y: 0}
	touch.lastStickDir = gvec.Vec2{X: -1, Y: 0}

	ctx := &stubCaveContext{cam: camera.Camera{Pos: gvec.Vec2{X: 0, Y: 0}}}
	ci.SetAimOrigin(gvec.Vec2{X: p.Pos.X + p.Width/2.0, Y: p.Pos.Y + p.Height/2.0})

	cave.handlePlayerMovement(ctx, ci, p)

	// math.Cos(p.Facing) must be negative (facing LEFT)!
	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected player to face LEFT (cos < 0) when moving left near cave edge, but got facing angle %v (cos = %v)",
			p.Facing, math.Cos(p.Facing))
	}

	// Release stick: facing should be retained, not flipped toward screen center.
	touch.stickActive = false
	cave.handlePlayerMovement(ctx, ci, p)

	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected player to RETAIN left facing when releasing stick, but got facing angle %v (cos = %v)",
			p.Facing, math.Cos(p.Facing))
	}

	// Push stick RIGHT: facing should flip to face RIGHT.
	touch.stickActive = true
	touch.stickVec = gvec.Vec2{X: 1, Y: 0}
	touch.lastStickDir = gvec.Vec2{X: 1, Y: 0}
	cave.handlePlayerMovement(ctx, ci, p)

	if math.Cos(p.Facing) <= 0 {
		t.Fatalf("expected player to face RIGHT (cos > 0) when moving right near cave edge, but got facing angle %v (cos = %v)",
			p.Facing, math.Cos(p.Facing))
	}
}

func TestTouchControls_FlashlightButton(t *testing.T) {
	touch := NewTouchControls()

	// When hasFlashlightAvailable is false, the button should NOT be visible anywhere.
	touch.SetHasFlashlightAvailable(false)
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyT {
			if b.visibleIn(TouchContextCave) {
				t.Fatal("Flashlight button should NOT be visible when hasFlashlightAvailable is false")
			}
			if b.visibleIn(TouchContextCaveDriving) {
				t.Fatal("Flashlight button should NOT be visible when hasFlashlightAvailable is false")
			}
		}
	}

	// When hasFlashlightAvailable is true:
	touch.SetHasFlashlightAvailable(true)
	hasFlashlightCave := false
	hasFlashlightCaveDriving := false
	hasFlashlightOnFoot := false
	hasFlashlightDriving := false

	for _, b := range touch.buttons {
		if b.key == ebiten.KeyT {
			if b.visibleIn(TouchContextCave) {
				hasFlashlightCave = true
			}
			if b.visibleIn(TouchContextCaveDriving) {
				hasFlashlightCaveDriving = true
			}
			if b.visibleIn(TouchContextOnFoot) {
				hasFlashlightOnFoot = true
			}
			if b.visibleIn(TouchContextDriving) {
				hasFlashlightDriving = true
			}
		}
	}

	if !hasFlashlightCave {
		t.Fatal("expected Flashlight toggle button (KeyT) visible in TouchContextCave when available")
	}
	if !hasFlashlightCaveDriving {
		t.Fatal("expected Flashlight toggle button (KeyT) visible in TouchContextCaveDriving when available")
	}
	if hasFlashlightOnFoot {
		t.Fatal("Flashlight toggle button should NOT be visible in TouchContextOnFoot")
	}
	if hasFlashlightDriving {
		t.Fatal("Flashlight toggle button should NOT be visible in TouchContextDriving")
	}
}

func TestCavePlayerMovement_AimTouchOverridesStick(t *testing.T) {
	cave := NewCaveScene()
	p := player.NewPlayer(200, 200)

	base := NewEbitenInput()
	touch := NewTouchControls()
	ci := NewCombinedInput(base, touch)

	// Simulate movement stick pointing LEFT (-1, 0).
	touch.active = true
	touch.stickActive = true
	touch.stickVec = gvec.Vec2{X: -1, Y: 0}
	touch.lastStickDir = gvec.Vec2{X: -1, Y: 0}

	// But user touches/drags with second finger to the RIGHT (aiming at x=500, y=200).
	touch.aimActive = true
	touch.aimPos = gvec.Vec2{X: 500, Y: 200}

	ctx := &stubCaveContext{cam: camera.Camera{Pos: gvec.Vec2{X: 0, Y: 0}}}
	cave.handlePlayerMovement(ctx, ci, p)

	// Player body should face LEFT with the movement stick (cos < 0).
	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected player body to face stick direction (left, cos < 0), got facing angle %v (cos = %v)",
			p.Facing, math.Cos(p.Facing))
	}

	// Flashlight should aim toward aim touch point (to the RIGHT, cos > 0).
	if math.Cos(p.FlashlightAngle) <= 0 {
		t.Fatalf("expected flashlight to aim toward aim touch (to the right, cos > 0), got flashlight angle %v (cos = %v)",
			p.FlashlightAngle, math.Cos(p.FlashlightAngle))
	}

	// Release aim touch: flashlight should fall back to stick facing (LEFT, cos < 0) over frames.
	touch.aimActive = false
	for i := 0; i < 25; i++ {
		cave.handlePlayerMovement(ctx, ci, p)
	}

	if math.Cos(p.FlashlightAngle) >= 0 {
		t.Fatalf("expected flashlight to return to stick direction (left, cos < 0) once aim touch is released, got angle %v (cos = %v)",
			p.FlashlightAngle, math.Cos(p.FlashlightAngle))
	}
}

type stubInput struct {
	keysPressed map[ebiten.Key]bool
	cursor      gvec.Vec2
}

func (s *stubInput) Update()                                            {}
func (s *stubInput) IsKeyPressed(k ebiten.Key) bool                     { return s.keysPressed[k] }
func (s *stubInput) IsKeyJustPressed(k ebiten.Key) bool                 { return false }
func (s *stubInput) IsMouseButtonJustPressed(b ebiten.MouseButton) bool { return false }
func (s *stubInput) Cursor() gvec.Vec2                                  { return s.cursor }
func (s *stubInput) Wheel() (float64, float64)                          { return 0, 0 }
func (s *stubInput) AppendInputChars(runes []rune) []rune               { return runes }

func TestCavePlayerMovement_KeyboardFacingAndMouseFlashlight(t *testing.T) {
	cave := NewCaveScene()
	p := player.NewPlayer(200, 200)
	ctx := &stubCaveContext{cam: camera.Camera{Pos: gvec.Vec2{X: 0, Y: 0}}, p: p}

	inp := &stubInput{
		keysPressed: make(map[ebiten.Key]bool),
		cursor:      gvec.Vec2{X: 200, Y: 200},
	}

	// 1. In FlashlightFollowsMouse = true mode:
	config.FlashlightFollowsMouse = true

	// Move Left with A: Body faces LEFT.
	inp.keysPressed[ebiten.KeyA] = true
	cave.handlePlayerMovement(ctx, inp, p)
	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected player body to face LEFT (cos < 0) on KeyA, got %v (cos = %v)", p.Facing, math.Cos(p.Facing))
	}

	// Release A (idle): Body retains LEFT facing.
	inp.keysPressed[ebiten.KeyA] = false
	cave.handlePlayerMovement(ctx, inp, p)
	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected player body to RETAIN LEFT facing when idle, got %v (cos = %v)", p.Facing, math.Cos(p.Facing))
	}

	// Move Right with D: Body faces RIGHT.
	inp.keysPressed[ebiten.KeyD] = true
	cave.handlePlayerMovement(ctx, inp, p)
	if math.Cos(p.Facing) <= 0 {
		t.Fatalf("expected player body to face RIGHT (cos > 0) on KeyD, got %v (cos = %v)", p.Facing, math.Cos(p.Facing))
	}
	inp.keysPressed[ebiten.KeyD] = false

	// Position mouse to the LEFT while player body faces RIGHT.
	// Flashlight should point toward mouse (LEFT, cos < 0), while body stays facing RIGHT (cos > 0).
	inp.cursor = gvec.Vec2{X: 50, Y: 200}
	cave.handlePlayerMovement(ctx, inp, p)

	if math.Cos(p.Facing) <= 0 {
		t.Fatalf("expected player body to stay facing RIGHT while aiming mouse left, got %v", p.Facing)
	}
	if math.Cos(p.FlashlightAngle) >= 0 {
		t.Fatalf("expected flashlight to aim towards mouse (LEFT, cos < 0), got %v (cos = %v)", p.FlashlightAngle, math.Cos(p.FlashlightAngle))
	}

	// 2. In FlashlightFollowsMouse = false mode (Keyboard-only mode):
	config.FlashlightFollowsMouse = false

	// When moving Left (KeyA), flashlight aligns with body facing LEFT
	inp.keysPressed = make(map[ebiten.Key]bool)
	inp.keysPressed[ebiten.KeyA] = true
	for i := 0; i < 20; i++ {
		cave.handlePlayerMovement(ctx, inp, p)
	}
	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected body to face LEFT on KeyA in keyboard mode, got %v", p.Facing)
	}
	if math.Cos(p.FlashlightAngle) >= 0 {
		t.Fatalf("expected flashlight to face LEFT with body in keyboard mode, got %v", p.FlashlightAngle)
	}

	// When moving UP (KeyW), flashlight points UP (sin < 0)
	inp.keysPressed = make(map[ebiten.Key]bool)
	inp.keysPressed[ebiten.KeyW] = true
	for i := 0; i < 20; i++ {
		cave.handlePlayerMovement(ctx, inp, p)
	}
	if math.Sin(p.Facing) >= -0.8 {
		t.Fatalf("expected body to face UP on KeyW, got %v (sin = %v)", p.Facing, math.Sin(p.Facing))
	}
	if math.Sin(p.FlashlightAngle) >= -0.8 {
		t.Fatalf("expected flashlight to point UP on KeyW in keyboard mode, got %v (sin = %v)", p.FlashlightAngle, math.Sin(p.FlashlightAngle))
	}

	// When moving DOWN (KeyS), flashlight points DOWN (sin > 0)
	inp.keysPressed = make(map[ebiten.Key]bool)
	inp.keysPressed[ebiten.KeyS] = true
	for i := 0; i < 20; i++ {
		cave.handlePlayerMovement(ctx, inp, p)
	}
	if math.Sin(p.Facing) <= 0.8 {
		t.Fatalf("expected body to face DOWN on KeyS, got %v (sin = %v)", p.Facing, math.Sin(p.Facing))
	}
	if math.Sin(p.FlashlightAngle) <= 0.8 {
		t.Fatalf("expected flashlight to point DOWN on KeyS in keyboard mode, got %v (sin = %v)", p.FlashlightAngle, math.Sin(p.FlashlightAngle))
	}

	// When moving UP-LEFT (KeyW + KeyA), flashlight points UP-LEFT (sin < 0 && cos < 0)
	inp.keysPressed = make(map[ebiten.Key]bool)
	inp.keysPressed[ebiten.KeyW] = true
	inp.keysPressed[ebiten.KeyA] = true
	for i := 0; i < 20; i++ {
		cave.handlePlayerMovement(ctx, inp, p)
	}
	if math.Sin(p.Facing) >= 0 || math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected body to face UP-LEFT on KeyW+KeyA, got %v", p.Facing)
	}
	if math.Sin(p.FlashlightAngle) >= 0 || math.Cos(p.FlashlightAngle) >= 0 {
		t.Fatalf("expected flashlight to point UP-LEFT on KeyW+KeyA in keyboard mode, got %v", p.FlashlightAngle)
	}

	// In keyboard mode, moving the mouse to the right does NOT move the flashlight
	inp.cursor = gvec.Vec2{X: 500, Y: 200}
	cave.handlePlayerMovement(ctx, inp, p)
	if math.Cos(p.FlashlightAngle) >= 0 {
		t.Fatalf("expected flashlight to remain facing UP-LEFT despite mouse pointing right in keyboard mode, got %v", p.FlashlightAngle)
	}

	// Restore default
	config.FlashlightFollowsMouse = false
}

func TestCavePlayerMovement_DesktopCombinedInput(t *testing.T) {
	cave := NewCaveScene()
	p := player.NewPlayer(200, 200)
	p.Facing = math.Pi / 2.0 // simulate entering cave with vertical overworld transition angle
	ctx := &stubCaveContext{cam: camera.Camera{Pos: gvec.Vec2{X: 0, Y: 0}}, p: p}

	base := NewEbitenInput()
	base.cursor = gvec.Vec2{X: 200, Y: 200}
	touch := NewTouchControls() // touch.active is false by default (desktop)
	ci := NewCombinedInput(base, touch)

	// Entering cave initializes / normalizes vertical facing
	cave.onEnter(ctx)
	if math.Abs(math.Cos(p.Facing)) < 0.2 {
		t.Fatalf("expected vertical facing to normalize, got %v", p.Facing)
	}

	// 1. Press KeyA (swimming left) on Desktop CombinedInput
	base.pressedKeys[ebiten.KeyA] = true
	cave.handlePlayerMovement(ctx, ci, p)
	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected desktop player to face LEFT (cos < 0) on KeyA with CombinedInput, got %v", p.Facing)
	}

	// 2. Release KeyA (idle): retains LEFT facing
	base.pressedKeys[ebiten.KeyA] = false
	cave.handlePlayerMovement(ctx, ci, p)
	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected desktop player to RETAIN left facing with CombinedInput, got %v", p.Facing)
	}

	// 3. Move mouse to right (x=400, y=200): In mouse mode, flashlight aims RIGHT, player body stays facing LEFT
	config.FlashlightFollowsMouse = true
	base.cursor = gvec.Vec2{X: 400, Y: 200}
	cave.handlePlayerMovement(ctx, ci, p)

	if math.Cos(p.Facing) >= 0 {
		t.Fatalf("expected player body to stay facing LEFT, got %v", p.Facing)
	}
	if math.Cos(p.FlashlightAngle) <= 0 {
		t.Fatalf("expected flashlight to aim RIGHT toward mouse on desktop with CombinedInput, got %v", p.FlashlightAngle)
	}
}

func TestTouchControls_FlashlightToggleVisual(t *testing.T) {
	touch := NewTouchControls()
	touch.active = true
	touch.SetContext(TouchContextCave)
	touch.SetHasFlashlightAvailable(true)

	// Verify button exists
	var flashBtn *touchButton
	for _, b := range touch.buttons {
		if b.key == ebiten.KeyT && b.visibleIn(TouchContextCave) {
			flashBtn = b
			break
		}
	}
	if flashBtn == nil {
		t.Fatal("expected flashlight button visible in cave when available")
	}

	// Toggle state off
	touch.SetFlashlightState(false)
	if touch.flashlightOn {
		t.Fatal("expected flashlightOn to be false")
	}

	// Toggle state on
	touch.SetFlashlightState(true)
	if !touch.flashlightOn {
		t.Fatal("expected flashlightOn to be true")
	}
}

func TestHUDHotbarSlotAt(t *testing.T) {
	// Center of slot 0
	minX0, minY0, maxX0, maxY0 := HUDHotbarSlotRect(0)
	midX0 := (minX0 + maxX0) / 2.0
	midY0 := (minY0 + maxY0) / 2.0
	if slot := HUDHotbarSlotAt(midX0, midY0); slot != 0 {
		t.Fatalf("expected slot 0 at (%f, %f), got %d", midX0, midY0, slot)
	}

	// Center of slot 2
	minX2, minY2, maxX2, maxY2 := HUDHotbarSlotRect(2)
	midX2 := (minX2 + maxX2) / 2.0
	midY2 := (minY2 + maxY2) / 2.0
	if slot := HUDHotbarSlotAt(midX2, midY2); slot != 2 {
		t.Fatalf("expected slot 2 at (%f, %f), got %d", midX2, midY2, slot)
	}

	// Center of slot 4
	minX4, minY4, maxX4, maxY4 := HUDHotbarSlotRect(4)
	midX4 := (minX4 + maxX4) / 2.0
	midY4 := (minY4 + maxY4) / 2.0
	if slot := HUDHotbarSlotAt(midX4, midY4); slot != 4 {
		t.Fatalf("expected slot 4 at (%f, %f), got %d", midX4, midY4, slot)
	}

	// Outside hotbar
	if slot := HUDHotbarSlotAt(100, 100); slot != -1 {
		t.Fatalf("expected -1 outside hotbar, got %d", slot)
	}
}

func TestTouchControls_HotbarTouch(t *testing.T) {
	touch := NewTouchControls()
	touch.active = true
	touch.SetContext(TouchContextOnFoot)

	// Simulate touch routing on slot 3
	touch.hotbarTouched = true
	touch.hotbarSlot = 3

	slot, ok := touch.ConsumeHotbarTouch()
	if !ok || slot != 3 {
		t.Fatalf("expected to consume hotbar touch for slot 3, got (%d, %v)", slot, ok)
	}

	// Second call should return false
	if _, ok := touch.ConsumeHotbarTouch(); ok {
		t.Fatal("expected ConsumeHotbarTouch to return false after consumption")
	}
}

func TestHUDVehicleLightButtonHit(t *testing.T) {
	minX, minY, maxX, maxY := HUDVehicleLightButtonRect()
	midX := (minX + maxX) / 2.0
	midY := (minY + maxY) / 2.0

	if !HUDVehicleLightButtonHit(midX, midY) {
		t.Fatalf("expected hit at center (%f, %f)", midX, midY)
	}
	if !HUDVehicleLightButtonHit(minX, minY) {
		t.Fatalf("expected hit at top-left (%f, %f)", minX, minY)
	}
	if !HUDVehicleLightButtonHit(maxX, maxY) {
		t.Fatalf("expected hit at bottom-right (%f, %f)", maxX, maxY)
	}

	// Outside button
	if HUDVehicleLightButtonHit(minX-10, midY) {
		t.Fatal("expected miss to the left of button")
	}
	if HUDVehicleLightButtonHit(maxX+10, midY) {
		t.Fatal("expected miss to the right of button")
	}
	if HUDVehicleLightButtonHit(midX, minY-10) {
		t.Fatal("expected miss above button")
	}
	if HUDVehicleLightButtonHit(midX, maxY+10) {
		t.Fatal("expected miss below button")
	}
}

func TestTouchControls_TouchActiveState(t *testing.T) {
	touch := NewTouchControls()
	if touch.Active() {
		t.Fatal("expected touch to start inactive")
	}
	touch.active = true
	if !touch.Active() {
		t.Fatal("expected touch to report active when active flag is set")
	}

	ci := NewCombinedInput(NewEbitenInput(), touch)
	if !ci.TouchActive() {
		t.Fatal("expected CombinedInput.TouchActive() to report true when touch is active")
	}
}
