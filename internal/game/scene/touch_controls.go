package scene

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// TouchContext selects which virtual controls are visible and hit-testable.
type TouchContext int

const (
	TouchContextHidden TouchContext = iota
	TouchContextOnFoot // overworld on foot — Dive/Interact
	TouchContextCave   // cave on foot — Use (mine / tools)
	TouchContextDriving
	TouchContextCaveDriving
	TouchContextInventory
	TouchContextMenu
)

const (
	stickZoneMaxX    = 500.0
	stickZoneMinY    = 320.0
	stickMaxRadius   = 70.0
	stickDeadzone    = 0.25
	stickSprintMag   = 0.92
	buttonHitPadding = 10.0

	// uiDragThreshold is how far a finger must move in the menu before the
	// gesture becomes a scroll instead of a tap. Menu lists only listen to
	// mouse-wheel deltas, so drag motion is converted to wheel units
	// (pixels / uiDragWheelScale) to match ScrollY -= wy * 15.
	uiDragThreshold  = 10.0
	uiDragWheelScale = 15.0
)

var (
	touchIconColor  = color.RGBA{215, 235, 245, 255}
	touchFillColor  = color.RGBA{18, 24, 38, 170}
	touchFillHeld   = color.RGBA{60, 95, 130, 210}
	touchStrokeCol  = color.RGBA{90, 130, 165, 220}
	touchStickBase  = color.RGBA{18, 24, 38, 120}
	touchStickKnob  = color.RGBA{120, 165, 200, 190}
	touchStickStrok = color.RGBA{90, 130, 165, 160}
)

// touchButton is one on-screen virtual button. It may inject a keyboard key
// and/or a left-click (Use). nearestUse marks Use presses that should mine
// the nearest in-range target instead of aiming at the cursor.
type touchButton struct {
	cx, cy, r  float64
	key        ebiten.Key
	mouseLeft  bool
	nearestUse bool
	icon       *ebiten.Image
	contexts   []TouchContext
	touchID    ebiten.TouchID
	held       bool
	condition  func() bool
}

func (b *touchButton) visibleIn(ctx TouchContext) bool {
	if b.condition != nil && !b.condition() {
		return false
	}
	for _, c := range b.contexts {
		if c == ctx {
			return true
		}
	}
	return false
}

func (b *touchButton) hit(x, y float64) bool {
	return math.Hypot(x-b.cx, y-b.cy) <= b.r+buttonHitPadding
}

// TouchControls owns the virtual thumbstick, on-screen buttons, and raw touch
// tracking. It exposes virtual key/tap state which CombinedInput merges with
// the physical keyboard/mouse.
type TouchControls struct {
	active  bool // touch mode: shown after any touch, hidden on keyboard/mouse input
	context TouchContext

	canEnterVehicle        bool
	canEnterLifePod        bool
	hasVehicleSonar        bool
	hasVehicleSpecial      bool
	hasFlashlightAvailable bool
	flashlightOn           bool

	// Thumbstick state. The stick anchors where the touch begins inside the
	// bottom-left zone and follows that touch until release.
	stickActive  bool
	stickTouch   ebiten.TouchID
	stickOrigin  gvec.Vec2
	stickVec     gvec.Vec2 // clamped to unit length
	lastStickDir gvec.Vec2 // last non-deadzone direction, unit length

	buttons []*touchButton

	virtualHeld map[ebiten.Key]bool
	virtualJust map[ebiten.Key]bool

	virtualLeftClick bool
	preferNearestUse bool

	// Aim touch state: in cave mode, touching/dragging outside buttons and stick
	// aims the flashlight toward the touch position.
	aimActive bool
	aimTouch  ebiten.TouchID
	aimPos    gvec.Vec2

	// A tap is a touch press not captured by the stick or a button. It acts
	// as a left-click at the touch position unless consumed by world-tap
	// handling (boarding vehicles, opening the lifepod).
	tapPending  bool
	tapConsumed bool
	tapPos      gvec.Vec2

	// Menu UI drag: press is held until release so a finger swipe can scroll
	// list tabs (fabricator / quests / logs) without firing a click. If the
	// finger never exceeds uiDragThreshold, release synthesizes a tap.
	uiDragActive   bool
	uiDragTouch    ebiten.TouchID
	uiDragStart    gvec.Vec2
	uiDragLast     gvec.Vec2
	uiDragMoved    bool
	uiScrollWheelY float64 // synthesized wheel Y for this frame (matches ebiten.Wheel)

	hotbarTouched bool
	hotbarSlot    int
	hotbarPending int

	touchIDs    []ebiten.TouchID
	justTouched []ebiten.TouchID
	keyScratch  []ebiten.Key
}

// NewTouchControls creates the virtual control overlay with all buttons laid out.
func NewTouchControls() *TouchControls {
	onFoot := []TouchContext{TouchContextOnFoot}
	cave := []TouchContext{TouchContextCave}
	driving := []TouchContext{TouchContextDriving, TouchContextCaveDriving}
	overworldDriving := []TouchContext{TouchContextDriving}
	caveOrCaveDriving := []TouchContext{TouchContextCave, TouchContextCaveDriving}
	onFootOrCave := []TouchContext{TouchContextOnFoot, TouchContextCave}
	gameplay := []TouchContext{TouchContextOnFoot, TouchContextCave, TouchContextDriving, TouchContextCaveDriving}

	tc := &TouchControls{
		context:       TouchContextOnFoot,
		hotbarPending: -1,
		virtualHeld:   make(map[ebiten.Key]bool),
		virtualJust:   make(map[ebiten.Key]bool),
	}

	tc.buttons = []*touchButton{
		// Primary action: dive on overworld, use/mine in caves, exit while driving.
		{cx: 1185, cy: 615, r: 48, key: ebiten.KeyE, icon: makeTouchIcon(drawIconDive), contexts: onFoot},
		{cx: 1185, cy: 615, r: 48, mouseLeft: true, nearestUse: true, icon: makeTouchIcon(drawIconUse), contexts: cave},
		{cx: 1185, cy: 615, r: 48, key: ebiten.KeyF, icon: makeTouchIcon(drawIconExit), contexts: driving},
		// Cave sprint (hold Shift). Stick no longer auto-sprints in caves.
		{cx: 1185, cy: 505, r: 36, key: ebiten.KeyShift, icon: makeTouchIcon(drawIconSprint), contexts: cave},
		// Secondary arc.
		{cx: 1078, cy: 655, r: 34, key: ebiten.KeyTab, icon: makeTouchIcon(drawIconInventory), contexts: gameplay},
		{cx: 1053, cy: 560, r: 34, key: ebiten.KeyM, icon: makeTouchIcon(drawIconMap), contexts: onFoot},
		{cx: 1053, cy: 560, r: 34, key: ebiten.KeyJ, icon: makeTouchIcon(drawIconPDA), contexts: caveOrCaveDriving},
		{cx: 1108, cy: 478, r: 34, key: ebiten.KeyF, icon: makeTouchIcon(drawIconEnterVehicle), contexts: onFootOrCave, condition: func() bool { return tc.canEnterVehicle }},
		{cx: 1000, cy: 460, r: 34, key: ebiten.KeyE, icon: makeTouchIcon(drawIconFabricator), contexts: onFoot, condition: func() bool { return tc.canEnterLifePod }},
		{cx: 1000, cy: 460, r: 34, key: ebiten.KeyT, icon: makeTouchIcon(drawIconFlashlight), contexts: caveOrCaveDriving, condition: func() bool { return tc.hasFlashlightAvailable }},
		{cx: 1053, cy: 560, r: 34, key: ebiten.KeyQ, icon: makeTouchIcon(drawIconSonar), contexts: driving, condition: func() bool { return tc.hasVehicleSonar }},
		{cx: 1108, cy: 478, r: 34, key: ebiten.KeySpace, icon: makeTouchIcon(drawIconAction), contexts: driving, condition: func() bool { return tc.hasVehicleSpecial }},
		{cx: 1000, cy: 460, r: 34, key: ebiten.KeyM, icon: makeTouchIcon(drawIconMap), contexts: overworldDriving},
		// Corner buttons.
		{cx: 1240, cy: 40, r: 26, key: ebiten.KeyEscape, icon: makeTouchIcon(drawIconPause), contexts: gameplay},
		{cx: 1240, cy: 40, r: 26, key: ebiten.KeyEscape, icon: makeTouchIcon(drawIconClose), contexts: []TouchContext{TouchContextInventory}},
		{cx: 1240, cy: 40, r: 26, key: ebiten.KeyE, icon: makeTouchIcon(drawIconClose), contexts: []TouchContext{TouchContextMenu}},
	}

	return tc
}

// SetContext selects the button set for the current game state. Call before Update.
func (t *TouchControls) SetContext(ctx TouchContext) { t.context = ctx }

// SetCanEnterVehicle updates whether the player is close enough to enter a vehicle.
func (t *TouchControls) SetCanEnterVehicle(can bool) { t.canEnterVehicle = can }

// SetCanEnterLifePod updates whether the player is close enough to enter the Life Pod.
func (t *TouchControls) SetCanEnterLifePod(can bool) { t.canEnterLifePod = can }

// SetVehicleCapabilities updates whether the active vehicle has sonar and/or special upgrades.
func (t *TouchControls) SetVehicleCapabilities(hasSonar, hasSpecial bool) {
	t.hasVehicleSonar = hasSonar
	t.hasVehicleSpecial = hasSpecial
}

// SetHasFlashlightAvailable updates whether the player is currently holding a Flashlight or driving a vehicle with headlights.
func (t *TouchControls) SetHasFlashlightAvailable(avail bool) {
	t.hasFlashlightAvailable = avail
}

// SetFlashlightState updates whether the flashlight / headlights are actively toggled ON.
func (t *TouchControls) SetFlashlightState(on bool) {
	t.flashlightOn = on
}

// Active reports whether touch mode is engaged (controls visible, synthetic cursor in use).
func (t *TouchControls) Active() bool { return t.active }

// VirtualKeyPressed reports a key held via the stick or a button this frame.
func (t *TouchControls) VirtualKeyPressed(k ebiten.Key) bool { return t.virtualHeld[k] }

// VirtualKeyJustPressed reports a key press synthesized this frame.
func (t *TouchControls) VirtualKeyJustPressed(k ebiten.Key) bool { return t.virtualJust[k] }

// InjectJustPressed synthesizes a one-frame key press (used by world-tap handling).
func (t *TouchControls) InjectJustPressed(k ebiten.Key) {
	t.virtualJust[k] = true
	t.virtualHeld[k] = true
}

// VirtualLeftClickJustPressed reports a Use-button press that should act as left-click.
func (t *TouchControls) VirtualLeftClickJustPressed() bool { return t.virtualLeftClick }

// PreferNearestUse reports that this frame's Use press should target the nearest
// interactable (ore, bulb, catchable) instead of the cursor tile.
func (t *TouchControls) PreferNearestUse() bool { return t.preferNearestUse }

// TapCursor returns the position of this frame's unconsumed tap, if any.
func (t *TouchControls) TapCursor() (gvec.Vec2, bool) {
	if t.tapPending && !t.tapConsumed {
		return t.tapPos, true
	}
	return gvec.Vec2{}, false
}

// UIPointer returns the active menu finger position while a UI drag/tap is held.
func (t *TouchControls) UIPointer() (gvec.Vec2, bool) {
	if t.uiDragActive {
		return t.uiDragLast, true
	}
	return gvec.Vec2{}, false
}

// UIScrollWheelY returns wheel-Y synthesized from a menu finger drag this frame.
func (t *TouchControls) UIScrollWheelY() float64 { return t.uiScrollWheelY }

// ConsumeTap suppresses this frame's tap so it does not become a left-click.
func (t *TouchControls) ConsumeTap() { t.tapConsumed = true }

// ConsumeHotbarTouch returns the hotbar slot index touched this frame, if any.
func (t *TouchControls) ConsumeHotbarTouch() (int, bool) {
	if t.hotbarTouched {
		t.hotbarTouched = false
		return t.hotbarSlot, true
	}
	return -1, false
}

// TriggerHotbarSlot simulates a touch on a hotbar slot for testing.
func (t *TouchControls) TriggerHotbarSlot(slot int) {
	t.active = true
	t.hotbarPending = slot
}

// StickFacing returns the stick direction (unit length) while it is engaged
// past the deadzone, for cursor-facing movement.
func (t *TouchControls) StickFacing() (gvec.Vec2, bool) {
	if t.stickActive && math.Hypot(t.stickVec.X, t.stickVec.Y) > stickDeadzone {
		return t.lastStickDir, true
	}
	return gvec.Vec2{}, false
}

// AimTouch returns the active screen touch position used to aim the flashlight in caves.
func (t *TouchControls) AimTouch() (gvec.Vec2, bool) {
	if t.aimActive {
		return t.aimPos, true
	}
	return gvec.Vec2{}, false
}

// StickAxes returns the analog stick vector in [-1, 1] while a stick touch is held.
func (t *TouchControls) StickAxes() (gvec.Vec2, bool) {
	if !t.stickActive {
		return gvec.Vec2{}, false
	}
	return t.stickVec, true
}

// Update polls raw touches and rebuilds virtual key/tap state for this frame.
func (t *TouchControls) Update() {
	for k := range t.virtualJust {
		delete(t.virtualJust, k)
	}
	for k := range t.virtualHeld {
		delete(t.virtualHeld, k)
	}
	t.tapPending = false
	t.tapConsumed = false
	t.virtualLeftClick = false
	t.preferNearestUse = false
	t.uiScrollWheelY = 0
	t.hotbarTouched = false
	if t.hotbarPending >= 0 {
		t.hotbarTouched = true
		t.hotbarSlot = t.hotbarPending
		t.hotbarPending = -1
	}

	// Physical keyboard/mouse input hides the touch overlay.
	t.keyScratch = inpututil.AppendJustPressedKeys(t.keyScratch[:0])
	if len(t.keyScratch) > 0 ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		t.active = false
	}

	t.justTouched = inpututil.AppendJustPressedTouchIDs(t.justTouched[:0])
	if len(t.justTouched) > 0 {
		t.active = true
	}
	t.touchIDs = ebiten.AppendTouchIDs(t.touchIDs[:0])

	gameplay := t.context == TouchContextOnFoot || t.context == TouchContextCave || t.context == TouchContextDriving || t.context == TouchContextCaveDriving
	menuUI := t.context == TouchContextMenu

	// Leaving the menu cancels an in-progress UI drag without synthesizing a tap.
	if t.uiDragActive && !menuUI {
		t.uiDragActive = false
		t.uiDragMoved = false
	}

	// Route new touches: stick zone → stick, button hit → button, else tap
	// (or deferred menu drag/tap).
	for _, id := range t.justTouched {
		xi, yi := ebiten.TouchPosition(id)
		x, y := float64(xi), float64(yi)

		if gameplay && !t.stickActive && x < stickZoneMaxX && y > stickZoneMinY {
			t.stickActive = true
			t.stickTouch = id
			t.stickOrigin = gvec.Vec2{X: x, Y: y}
			t.stickVec = gvec.Vec2{}
			continue
		}
		if btn := t.buttonAt(x, y); btn != nil {
			btn.held = true
			btn.touchID = id
			if btn.key != 0 {
				t.virtualJust[btn.key] = true
			}
			if btn.mouseLeft {
				t.virtualLeftClick = true
			}
			if btn.nearestUse {
				t.preferNearestUse = true
			}
			continue
		}
		if gameplay && !menuUI {
			if slot := HUDHotbarSlotAt(x, y); slot >= 0 {
				t.hotbarTouched = true
				t.hotbarSlot = slot
				continue
			}
		}
		if menuUI && !t.uiDragActive {
			t.uiDragActive = true
			t.uiDragTouch = id
			t.uiDragStart = gvec.Vec2{X: x, Y: y}
			t.uiDragLast = t.uiDragStart
			t.uiDragMoved = false
			continue
		}
		if (t.context == TouchContextCave || t.context == TouchContextCaveDriving) && !t.aimActive {
			t.aimActive = true
			t.aimTouch = id
			t.aimPos = gvec.Vec2{X: x, Y: y}
		}
		t.tapPending = true
		t.tapPos = gvec.Vec2{X: x, Y: y}
	}

	// Track menu finger: drag past threshold → scroll; release without move → tap.
	if t.uiDragActive {
		if !t.touchAlive(t.uiDragTouch) {
			if !t.uiDragMoved {
				t.tapPending = true
				t.tapPos = t.uiDragStart
			}
			t.uiDragActive = false
			t.uiDragMoved = false
		} else {
			xi, yi := ebiten.TouchPosition(t.uiDragTouch)
			pos := gvec.Vec2{X: float64(xi), Y: float64(yi)}
			dy := pos.Y - t.uiDragLast.Y
			t.uiDragLast = pos
			if !t.uiDragMoved {
				if math.Hypot(pos.X-t.uiDragStart.X, pos.Y-t.uiDragStart.Y) >= uiDragThreshold {
					t.uiDragMoved = true
				}
			}
			if t.uiDragMoved && dy != 0 {
				// Finger down → content follows → ScrollY decreases → positive wheel.
				t.uiScrollWheelY = dy / uiDragWheelScale
			}
		}
	}

	// Track or release the stick touch.
	if t.stickActive {
		if !gameplay || !t.touchAlive(t.stickTouch) {
			t.stickActive = false
			t.stickVec = gvec.Vec2{}
		} else {
			xi, yi := ebiten.TouchPosition(t.stickTouch)
			dx := float64(xi) - t.stickOrigin.X
			dy := float64(yi) - t.stickOrigin.Y
			mag := math.Hypot(dx, dy)
			if mag > stickMaxRadius {
				dx *= stickMaxRadius / mag
				dy *= stickMaxRadius / mag
			}
			t.stickVec = gvec.Vec2{X: dx / stickMaxRadius, Y: dy / stickMaxRadius}
		}
	}

	// Track or release the aim touch (for cave flashlight aiming).
	if t.aimActive {
		if (t.context != TouchContextCave && t.context != TouchContextCaveDriving) || !t.touchAlive(t.aimTouch) {
			t.aimActive = false
		} else {
			xi, yi := ebiten.TouchPosition(t.aimTouch)
			t.aimPos = gvec.Vec2{X: float64(xi), Y: float64(yi)}
		}
	}

	// Release buttons whose touch ended or whose context changed.
	for _, b := range t.buttons {
		if b.held && (!b.visibleIn(t.context) || !t.touchAlive(b.touchID)) {
			b.held = false
		}
		if b.held && b.key != 0 {
			t.virtualHeld[b.key] = true
		}
	}

	// Map stick deflection to WASD. Full deflection sprints on overworld/driving;
	// caves use a dedicated Sprint button so normal swimming does not drain stamina.
	// The Skiff reads StickAxes directly for aim-steering; other craft still use WASD.
	if t.stickActive {
		mag := math.Hypot(t.stickVec.X, t.stickVec.Y)
		if mag > stickDeadzone {
			nx, ny := t.stickVec.X/mag, t.stickVec.Y/mag
			t.lastStickDir = gvec.Vec2{X: nx, Y: ny}
			if nx < -0.4 {
				t.virtualHeld[ebiten.KeyA] = true
			}
			if nx > 0.4 {
				t.virtualHeld[ebiten.KeyD] = true
			}
			if ny < -0.4 {
				t.virtualHeld[ebiten.KeyW] = true
			}
			if ny > 0.4 {
				t.virtualHeld[ebiten.KeyS] = true
			}
			if mag > stickSprintMag && t.context != TouchContextCave && t.context != TouchContextCaveDriving {
				t.virtualHeld[ebiten.KeyShift] = true
			}
		}
	}
}

func (t *TouchControls) buttonAt(x, y float64) *touchButton {
	for _, b := range t.buttons {
		if b.visibleIn(t.context) && b.hit(x, y) {
			return b
		}
	}
	return nil
}

func (t *TouchControls) touchAlive(id ebiten.TouchID) bool {
	for _, cur := range t.touchIDs {
		if cur == id {
			return true
		}
	}
	return false
}

// Draw renders the thumbstick and visible buttons. Call last, over the HUD.
func (t *TouchControls) Draw(screen *ebiten.Image) {
	if !t.active || t.context == TouchContextHidden {
		return
	}

	if t.context == TouchContextOnFoot || t.context == TouchContextCave || t.context == TouchContextDriving || t.context == TouchContextCaveDriving {
		origin := gvec.Vec2{X: 170, Y: 565}
		if t.stickActive {
			origin = t.stickOrigin
		}
		vector.FillCircle(screen, float32(origin.X), float32(origin.Y), stickMaxRadius, touchStickBase, true)
		vector.StrokeCircle(screen, float32(origin.X), float32(origin.Y), stickMaxRadius, 2, touchStickStrok, true)
		knobX := origin.X + t.stickVec.X*stickMaxRadius
		knobY := origin.Y + t.stickVec.Y*stickMaxRadius
		vector.FillCircle(screen, float32(knobX), float32(knobY), 28, touchStickKnob, true)
	}

	for _, b := range t.buttons {
		if !b.visibleIn(t.context) {
			continue
		}
		fill := touchFillColor
		stroke := touchStrokeCol
		if b.key == ebiten.KeyT && t.flashlightOn {
			fill = color.RGBA{140, 110, 20, 210}
			stroke = color.RGBA{255, 230, 90, 255}
		}
		if b.held {
			fill = touchFillHeld
		}
		vector.FillCircle(screen, float32(b.cx), float32(b.cy), float32(b.r), fill, true)
		vector.StrokeCircle(screen, float32(b.cx), float32(b.cy), float32(b.r), 2, stroke, true)
		if b.icon != nil {
			op := &ebiten.DrawImageOptions{}
			w, h := b.icon.Bounds().Dx(), b.icon.Bounds().Dy()
			op.GeoM.Translate(b.cx-float64(w)/2.0, b.cy-float64(h)/2.0)
			if b.key == ebiten.KeyT && t.flashlightOn {
				op.ColorScale.Scale(1.2, 1.1, 0.6, 1.0)
			}
			screen.DrawImage(b.icon, op)
		}
	}
}

// --- vector icons -----------------------------------------------------------

const touchIconSize = 48

// makeTouchIcon renders one icon painter into a cached image.
func makeTouchIcon(paint func(dst *ebiten.Image)) *ebiten.Image {
	img := ebiten.NewImage(touchIconSize, touchIconSize)
	paint(img)
	return img
}

// strokeArc approximates a circular arc with line segments (the vector package
// in use here has no direct arc stroke helper).
func strokeArc(dst *ebiten.Image, cx, cy, r, a0, a1 float64, width float32, clr color.Color) {
	const steps = 14
	px := cx + math.Cos(a0)*r
	py := cy + math.Sin(a0)*r
	for i := 1; i <= steps; i++ {
		a := a0 + (a1-a0)*float64(i)/steps
		x := cx + math.Cos(a)*r
		y := cy + math.Sin(a)*r
		vector.StrokeLine(dst, float32(px), float32(py), float32(x), float32(y), width, clr, true)
		px, py = x, y
	}
}

// drawIconDive draws a wave line with an arrow plunging beneath it.
func drawIconDive(dst *ebiten.Image) {
	c := touchIconColor
	// Wave across the top.
	strokeArc(dst, 12, 14, 6, math.Pi, 2*math.Pi, 2.5, c)
	strokeArc(dst, 24, 14, 6, 0, math.Pi, 2.5, c)
	strokeArc(dst, 36, 14, 6, math.Pi, 2*math.Pi, 2.5, c)
	// Downward arrow.
	vector.StrokeLine(dst, 24, 20, 24, 40, 3, c, true)
	vector.StrokeLine(dst, 24, 40, 15, 31, 3, c, true)
	vector.StrokeLine(dst, 24, 40, 33, 31, 3, c, true)
}

// drawIconUse draws a pickaxe for cave mining / tool use.
func drawIconUse(dst *ebiten.Image) {
	c := touchIconColor
	// Handle.
	vector.StrokeLine(dst, 14, 36, 30, 14, 3.5, c, true)
	// Head.
	vector.StrokeLine(dst, 22, 10, 38, 18, 4, c, true)
	vector.StrokeLine(dst, 22, 10, 16, 18, 4, c, true)
	vector.StrokeLine(dst, 16, 18, 38, 18, 3, c, true)
}

// drawIconSprint draws a double chevron (>>) for hold-to-sprint.
func drawIconSprint(dst *ebiten.Image) {
	c := touchIconColor
	// Left chevron.
	vector.StrokeLine(dst, 10, 14, 22, 24, 3.5, c, true)
	vector.StrokeLine(dst, 22, 24, 10, 34, 3.5, c, true)
	// Right chevron.
	vector.StrokeLine(dst, 24, 14, 36, 24, 3.5, c, true)
	vector.StrokeLine(dst, 36, 24, 24, 34, 3.5, c, true)
}

// drawIconExit draws a doorway bracket with an arrow leaving it.
func drawIconExit(dst *ebiten.Image) {
	c := touchIconColor
	// Door frame open to the right.
	vector.StrokeLine(dst, 20, 8, 8, 8, 3, c, true)
	vector.StrokeLine(dst, 8, 8, 8, 40, 3, c, true)
	vector.StrokeLine(dst, 8, 40, 20, 40, 3, c, true)
	// Arrow out.
	vector.StrokeLine(dst, 16, 24, 40, 24, 3, c, true)
	vector.StrokeLine(dst, 40, 24, 31, 15, 3, c, true)
	vector.StrokeLine(dst, 40, 24, 31, 33, 3, c, true)
}

// drawIconEnterVehicle draws a doorway/hatch frame with an arrow entering it.
func drawIconEnterVehicle(dst *ebiten.Image) {
	c := touchIconColor
	// Door frame open to the left.
	vector.StrokeLine(dst, 28, 8, 40, 8, 3, c, true)
	vector.StrokeLine(dst, 40, 8, 40, 40, 3, c, true)
	vector.StrokeLine(dst, 40, 40, 28, 40, 3, c, true)
	// Arrow pointing in.
	vector.StrokeLine(dst, 8, 24, 32, 24, 3, c, true)
	vector.StrokeLine(dst, 23, 15, 32, 24, 3, c, true)
	vector.StrokeLine(dst, 23, 33, 32, 24, 3, c, true)
}

// drawIconFabricator draws a life-pod / fabricator unit with a crafting cube.
func drawIconFabricator(dst *ebiten.Image) {
	c := touchIconColor
	// Outer capsule / unit frame.
	vector.StrokeRect(dst, 11, 8, 26, 32, 2.5, c, true)
	// Top pod beacon / hatch cap.
	vector.StrokeLine(dst, 18, 8, 18, 5, 2, c, true)
	vector.StrokeLine(dst, 30, 8, 30, 5, 2, c, true)
	vector.StrokeLine(dst, 18, 5, 30, 5, 2, c, true)
	// Internal synthesis / crafting cube (isometric).
	// Top face.
	vector.StrokeLine(dst, 24, 15, 31, 19, 1.8, c, true)
	vector.StrokeLine(dst, 31, 19, 24, 23, 1.8, c, true)
	vector.StrokeLine(dst, 24, 23, 17, 19, 1.8, c, true)
	vector.StrokeLine(dst, 17, 19, 24, 15, 1.8, c, true)
	// Bottom edges.
	vector.StrokeLine(dst, 17, 19, 17, 28, 1.8, c, true)
	vector.StrokeLine(dst, 31, 19, 31, 28, 1.8, c, true)
	vector.StrokeLine(dst, 24, 23, 24, 32, 1.8, c, true)
	vector.StrokeLine(dst, 17, 28, 24, 32, 1.8, c, true)
	vector.StrokeLine(dst, 31, 28, 24, 32, 1.8, c, true)
}

// drawIconFlashlight draws a flashlight with forward-projecting light rays.
func drawIconFlashlight(dst *ebiten.Image) {
	c := touchIconColor
	// Flashlight body / handle.
	vector.StrokeRect(dst, 10, 20, 14, 8, 2, c, true)
	// Grip grooves.
	vector.StrokeLine(dst, 14, 21, 14, 27, 1.5, c, true)
	vector.StrokeLine(dst, 18, 21, 18, 27, 1.5, c, true)
	// Top switch button.
	vector.StrokeLine(dst, 15, 17, 19, 17, 2, c, true)
	vector.StrokeLine(dst, 15, 17, 15, 20, 1.5, c, true)
	vector.StrokeLine(dst, 19, 17, 19, 20, 1.5, c, true)
	// Angled bezel / head.
	vector.StrokeLine(dst, 24, 20, 31, 15, 2, c, true)
	vector.StrokeLine(dst, 31, 15, 31, 33, 2.5, c, true)
	vector.StrokeLine(dst, 31, 33, 24, 28, 2, c, true)
	// Light beam rays.
	vector.StrokeLine(dst, 35, 16, 43, 11, 2, c, true)
	vector.StrokeLine(dst, 36, 24, 45, 24, 2, c, true)
	vector.StrokeLine(dst, 35, 32, 43, 37, 2, c, true)
}

// drawIconInventory draws a 2x2 grid of slots.
func drawIconInventory(dst *ebiten.Image) {
	c := touchIconColor
	vector.StrokeRect(dst, 9, 9, 13, 13, 2.5, c, true)
	vector.StrokeRect(dst, 26, 9, 13, 13, 2.5, c, true)
	vector.StrokeRect(dst, 9, 26, 13, 13, 2.5, c, true)
	vector.StrokeRect(dst, 26, 26, 13, 13, 2.5, c, true)
}

// drawIconMap draws a folded chart with a location dot.
func drawIconMap(dst *ebiten.Image) {
	c := touchIconColor
	vector.StrokeRect(dst, 8, 11, 32, 26, 2.5, c, true)
	// Fold creases.
	vector.StrokeLine(dst, 19, 11, 17, 37, 1.5, c, true)
	vector.StrokeLine(dst, 29, 11, 31, 37, 1.5, c, true)
	// Location dot.
	vector.FillCircle(dst, 24, 22, 3, c, true)
}

// drawIconPDA draws a handheld device with text lines.
func drawIconPDA(dst *ebiten.Image) {
	c := touchIconColor
	vector.StrokeRect(dst, 13, 7, 22, 34, 2.5, c, true)
	vector.StrokeLine(dst, 18, 15, 30, 15, 2, c, true)
	vector.StrokeLine(dst, 18, 22, 30, 22, 2, c, true)
	vector.StrokeLine(dst, 18, 29, 26, 29, 2, c, true)
}

// drawIconPause draws two vertical bars.
func drawIconPause(dst *ebiten.Image) {
	c := touchIconColor
	vector.FillRect(dst, 15, 12, 6, 24, c, true)
	vector.FillRect(dst, 27, 12, 6, 24, c, true)
}

// drawIconClose draws an X.
func drawIconClose(dst *ebiten.Image) {
	c := touchIconColor
	vector.StrokeLine(dst, 13, 13, 35, 35, 3.5, c, true)
	vector.StrokeLine(dst, 35, 13, 13, 35, 3.5, c, true)
}

// drawIconSonar draws a ping source with expanding arcs.
func drawIconSonar(dst *ebiten.Image) {
	c := touchIconColor
	vector.FillCircle(dst, 13, 35, 3.5, c, true)
	strokeArc(dst, 13, 35, 11, -math.Pi/2, 0, 2.5, c)
	strokeArc(dst, 13, 35, 19, -math.Pi/2, 0, 2.5, c)
	strokeArc(dst, 13, 35, 27, -math.Pi/2, 0, 2.5, c)
}

// drawIconAction draws a lightning bolt (vehicle special: decoy/jump).
func drawIconAction(dst *ebiten.Image) {
	c := touchIconColor
	vector.StrokeLine(dst, 28, 7, 17, 26, 3, c, true)
	vector.StrokeLine(dst, 17, 26, 26, 26, 3, c, true)
	vector.StrokeLine(dst, 26, 26, 19, 41, 3, c, true)
}

// --- combined input ----------------------------------------------------------

// CombinedInput implements InputSource by merging physical keyboard/mouse input
// with virtual touch input. When touch mode is active it also synthesizes the
// cursor position: taps become clicks at the tap point, and while the stick is
// engaged the cursor sits ahead of the screen center in the movement direction
// so cursor-facing (cave swimming, scout sub) follows the stick.
type CombinedInput struct {
	base  *EbitenInput
	touch *TouchControls

	aimOrigin  gvec.Vec2
	lastCursor gvec.Vec2
	haveCursor bool
}

// NewCombinedInput creates an InputSource merging physical input with touch controls.
func NewCombinedInput(base *EbitenInput, touch *TouchControls) *CombinedInput {
	return &CombinedInput{base: base, touch: touch}
}

// Update polls physical input then touch state. Call once per tick.
func (c *CombinedInput) Update() {
	c.base.Update()
	c.touch.Update()
}

// SetAimOrigin sets the screen coordinates of the player/vehicle from which synthetic cursor offsets originate.
func (c *CombinedInput) SetAimOrigin(origin gvec.Vec2) {
	c.aimOrigin = origin
}

// Cursor returns the physical cursor, or a synthetic one in touch mode.
func (c *CombinedInput) Cursor() gvec.Vec2 {
	if !c.touch.Active() {
		return c.base.Cursor()
	}
	if pos, ok := c.touch.AimTouch(); ok {
		c.lastCursor = pos
		c.haveCursor = true
		return pos
	}
	if pos, ok := c.touch.TapCursor(); ok {
		c.lastCursor = pos
		c.haveCursor = true
		return pos
	}
	if pos, ok := c.touch.UIPointer(); ok {
		c.lastCursor = pos
		c.haveCursor = true
		return pos
	}
	origin := c.aimOrigin
	if origin == (gvec.Vec2{}) {
		origin = gvec.Vec2{X: float64(config.ScreenWidth) / 2.0, Y: float64(config.ScreenHeight) / 2.0}
	}
	if dir, ok := c.touch.StickFacing(); ok {
		c.lastCursor = gvec.Vec2{
			X: origin.X + dir.X*120.0,
			Y: origin.Y + dir.Y*120.0,
		}
		c.haveCursor = true
	}
	if !c.haveCursor {
		c.lastCursor = origin
		c.haveCursor = true
	}
	return c.lastCursor
}

func (c *CombinedInput) IsKeyJustPressed(k ebiten.Key) bool {
	return c.base.IsKeyJustPressed(k) || c.touch.VirtualKeyJustPressed(k)
}

func (c *CombinedInput) IsKeyPressed(k ebiten.Key) bool {
	return c.base.IsKeyPressed(k) || c.touch.VirtualKeyPressed(k)
}

func (c *CombinedInput) IsMouseButtonJustPressed(b ebiten.MouseButton) bool {
	if c.base.IsMouseButtonJustPressed(b) {
		return true
	}
	if b == ebiten.MouseButtonLeft && c.touch.Active() {
		if c.touch.VirtualLeftClickJustPressed() {
			return true
		}
		_, ok := c.touch.TapCursor()
		return ok
	}
	return false
}

// PreferNearestUse reports whether cave mining should target the nearest
// interactable (Use button) rather than the cursor tile.
func (c *CombinedInput) PreferNearestUse() bool {
	return c.touch != nil && c.touch.PreferNearestUse()
}

// StickAxes returns the analog thumbstick vector while a stick touch is held.
func (c *CombinedInput) StickAxes() (gvec.Vec2, bool) {
	if c.touch == nil {
		return gvec.Vec2{}, false
	}
	return c.touch.StickAxes()
}

// StickFacing returns the thumbstick direction if engaged past the deadzone on touch.
func (c *CombinedInput) StickFacing() (gvec.Vec2, bool) {
	if c.touch == nil || !c.touch.Active() {
		return gvec.Vec2{}, false
	}
	return c.touch.StickFacing()
}

// TouchActive reports whether touch controls are active.
func (c *CombinedInput) TouchActive() bool {
	return c.touch != nil && c.touch.Active()
}

// TapCursor returns the screen position of a tap this frame, if any.
func (c *CombinedInput) TapCursor() (gvec.Vec2, bool) {
	if c.touch == nil || !c.touch.Active() {
		return gvec.Vec2{}, false
	}
	return c.touch.TapCursor()
}

// ConsumeTap suppresses this frame's tap on touch controls.
func (c *CombinedInput) ConsumeTap() {
	if c.touch != nil {
		c.touch.ConsumeTap()
	}
}

// AimTouch returns the active screen touch position used to aim the flashlight in caves.
func (c *CombinedInput) AimTouch() (gvec.Vec2, bool) {
	if c.touch == nil || !c.touch.Active() {
		return gvec.Vec2{}, false
	}
	return c.touch.AimTouch()
}

func (c *CombinedInput) Wheel() (float64, float64) {
	wx, wy := c.base.Wheel()
	if c.touch != nil {
		wy += c.touch.UIScrollWheelY()
	}
	return wx, wy
}

func (c *CombinedInput) AppendInputChars(runes []rune) []rune {
	return c.base.AppendInputChars(runes)
}
