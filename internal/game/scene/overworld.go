package scene

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/assets"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// OverworldContext defines the narrow context interface required by OverworldScene.
type OverworldContext interface {
	GetInput() InputSource
	GetPlayer() *player.Player
	GetCamera() *camera.Camera
	GetWorld() *world.World
	GetExploration() *exploration.Tracker
	GetBaseStation() *base.BaseStation
	GetTimeOfDay() float64
	GetTicks() float64
	GetActiveVehicle() vehicle.Vehicle
	GetAllCaveVehicles() map[string][]vehicle.Vehicle
	GetActiveTrenchKey() string
	GetActiveTrenchCoords() (x, y int)
	IsInventoryOpen() bool
	SetInventoryOpen(v bool)
	EnterCave(tx, ty int)
	TransitionToPDA()
	SetCurrentState(s State)
	SpawnPlankton(x, y float64)
	SpawnDebris(x, y float64, clr color.RGBA)
	SpawnBubble(x, y float64)
	TriggerScreenShake(duration int, intensity float64)
	SetMineWarning(msg string, duration, level int)
	AddItemToast(it item.Item, qty int)
	GetDeathReason() string
	SetDeathReason(reason string)
	DestroyOverworldVehicle(v vehicle.Vehicle)
	ActivatePlayerItem(it item.Item)
	UseRepairTool()
	NotifyQuestInventoryChanged(id item.ItemID)
}

// OverworldScene manages the top-down surface sailing view.
type OverworldScene struct {
	World        *world.World
	whirlpool    *oe.Whirlpool
	crates       []*oe.FloatingCrate
	vents        []*oe.ThermalVent
	fish         []*oe.CosmeticFish
	tileTextures map[world.TileType]*ebiten.Image
	initialized  bool

	// Cached offscreen static image details
	cachedStaticImage *ebiten.Image
	cachedChunkX      int
	cachedChunkY      int
	hasCache          bool

	topdownSwimFrames []*ebiten.Image

	divePromptTimer    int
	hasShownDivePrompt bool
}

func (o *OverworldScene) loadTopdownSwimFrames() {
	if len(o.topdownSwimFrames) > 0 {
		return
	}
	sheet, err := assets.LoadChromaKeyedImage("diver_topdown_sheet")
	if err != nil {
		log.Printf("Warning: Failed to load diver topdown sheet: %v", err)
		return
	}
	w := sheet.Bounds().Dx() / 4
	h := sheet.Bounds().Dy()
	for i := 0; i < 4; i++ {
		rect := image.Rect(i*w, 0, (i+1)*w, h)
		frame := ebiten.NewImageFromImage(sheet.SubImage(rect))
		o.topdownSwimFrames = append(o.topdownSwimFrames, frame)
	}
}

// NewOverworldScene creates a new OverworldScene.
func NewOverworldScene(w *world.World) *OverworldScene {
	return &OverworldScene{World: w, divePromptTimer: 180}
}

func (o *OverworldScene) getTileTexture(tileType world.TileType) *ebiten.Image {
	if o.tileTextures == nil {
		o.tileTextures = map[world.TileType]*ebiten.Image{
			world.TileTrench:   trenchTexture,
			world.TileWreckage: wreckageTexture,
		}
	}
	if tileType == world.TileShockKelpCave && o.tileTextures[world.TileShockKelpCave] == nil {
		o.tileTextures[world.TileShockKelpCave] = cave.GenerateShockKelpReefTexture()
	}
	return o.tileTextures[tileType]
}

func (o *OverworldScene) OnEnter(g GameContext) {
	o.onEnter(g)
}

func (o *OverworldScene) onEnter(g OverworldContext) {
	if o.whirlpool == nil {
		o.whirlpool = oe.NewWhirlpool(g.GetWorld().Seed)
		rng := rand.New(rand.NewSource(g.GetWorld().Seed + 997))
		pos := o.FindSafeWhirlpoolSpawnPos(g.GetBaseStation().Pos, rng)
		o.whirlpool.Relocate(pos)
	}
}

func (o *OverworldScene) OnExit(g GameContext) {
	o.onExit(g)
}

func (o *OverworldScene) onExit(g OverworldContext) {}

func (o *OverworldScene) Update(g GameContext) error {
	return o.update(g)
}

func (o *OverworldScene) Draw(g GameContext, screen *ebiten.Image) {
	o.draw(g, screen)
}

var (
	trenchTexture         *ebiten.Image
	trenchTextureLoaded   bool
	wreckageTexture       *ebiten.Image
	wreckageTextureLoaded bool
	shockKelpTexture      *ebiten.Image
)


// LoadAssets preloads and chroma-keys all overworld tile textures.
func LoadAssets() {
	// 1. Trench Texture
	{
		img, err := assets.LoadChromaKeyedImage("trench_surface")
		if err != nil {
			log.Printf("Error: Failed to load trench surface: %v", err)
		} else {
			trenchTexture = img
			trenchTextureLoaded = true
		}
	}

	// 2. Wreckage Texture
	{
		img, err := assets.LoadChromaKeyedImage("wreckage_surface")
		if err != nil {
			log.Printf("Error: Failed to load wreckage surface: %v", err)
		} else {
			wreckageTexture = img
			wreckageTextureLoaded = true
		}
	}
}

// IsSolid checks if the proposed bounding box overlaps with solid land.
func (o *OverworldScene) IsSolid(x, y, w, h float64) bool {
	x1, x2, y1, y2 := tileRange(x, y, w, h, config.TileSize)

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
				continue
			}
			if o.World.OverworldMap[tx][ty] == world.TileLand {
				return true
			}
		}
	}
	return false
}

// FindNearestWater locates the closest valid non-solid water position for an entity of dimensions (w, h).
func (o *OverworldScene) FindNearestWater(pos gvec.Vec2, w, h float64, baseStation *base.BaseStation) (gvec.Vec2, bool) {
	hasBase := baseStation != nil && baseStation.Size.X > 0 && baseStation.Size.Y > 0
	isSolid := func(p gvec.Vec2) bool {
		if p.X < 0 || p.X+w >= float64(o.World.Width*config.TileSize) ||
			p.Y < 0 || p.Y+h >= float64(o.World.Height*config.TileSize) {
			return true
		}
		if o.IsSolid(p.X, p.Y, w, h) {
			return true
		}
		if hasBase {
			bPos, bSize := baseStation.Pos, baseStation.Size
			if p.X < bPos.X+bSize.X && p.X+w > bPos.X &&
				p.Y < bPos.Y+bSize.Y && p.Y+h > bPos.Y {
				return true
			}
		}
		return false
	}

	if !isSolid(pos) {
		return pos, true
	}

	// Search outwards in expanding rings (from 8px up to 240px)
	const maxRadius = 240.0
	const step = 8.0
	for r := step; r <= maxRadius; r += step {
		for a := 0; a < 16; a++ {
			angle := float64(a) * (math.Pi / 8.0)
			candidate := gvec.Vec2{
				X: pos.X + math.Cos(angle)*r,
				Y: pos.Y + math.Sin(angle)*r,
			}
			if !isSolid(candidate) {
				return candidate, true
			}
		}
	}
	return pos, false
}

// FindSafeExitPosition determines a safe, non-solid water position to place the player
// when exiting a vehicle, avoiding land tiles, base station, and world boundaries.
func (o *OverworldScene) FindSafeExitPosition(vPos, vDims gvec.Vec2, facing float64, pW, pH float64, baseStation *base.BaseStation) gvec.Vec2 {
	hasBase := baseStation != nil && baseStation.Size.X > 0 && baseStation.Size.Y > 0
	isSolid := func(p gvec.Vec2) bool {
		if p.X < 0 || p.X+pW >= float64(o.World.Width*config.TileSize) ||
			p.Y < 0 || p.Y+pH >= float64(o.World.Height*config.TileSize) {
			return true
		}
		if o.IsSolid(p.X, p.Y, pW, pH) {
			return true
		}
		if hasBase {
			bPos, bSize := baseStation.Pos, baseStation.Size
			if p.X < bPos.X+bSize.X && p.X+pW > bPos.X &&
				p.Y < bPos.Y+bSize.Y && p.Y+pH > bPos.Y {
				return true
			}
		}
		return false
	}

	vCenter := gvec.Vec2{
		X: vPos.X + vDims.X/2.0,
		Y: vPos.Y + vDims.Y/2.0,
	}

	// 1. Primary candidates: Port, Starboard, Stern, and Bow relative to vehicle heading
	cosF := math.Cos(facing)
	sinF := math.Sin(facing)
	perpX := -sinF
	perpY := cosF

	sideDist := vDims.Y/2.0 + pH/2.0 + 8.0
	rearDist := vDims.X/2.0 + pW/2.0 + 8.0
	frontDist := vDims.X/2.0 + pW/2.0 + 8.0

	candidates := []gvec.Vec2{
		// Port (left side of boat)
		{X: vCenter.X + perpX*sideDist - pW/2.0, Y: vCenter.Y + perpY*sideDist - pH/2.0},
		// Starboard (right side of boat)
		{X: vCenter.X - perpX*sideDist - pW/2.0, Y: vCenter.Y - perpY*sideDist - pH/2.0},
		// Stern (behind boat)
		{X: vCenter.X - cosF*rearDist - pW/2.0, Y: vCenter.Y - sinF*rearDist - pH/2.0},
		// Bow (in front of boat)
		{X: vCenter.X + cosF*frontDist - pW/2.0, Y: vCenter.Y + sinF*frontDist - pH/2.0},
		// Cardinal box boundaries
		{X: vPos.X - pW - 8.0, Y: vCenter.Y - pH/2.0},
		{X: vPos.X + vDims.X + 8.0, Y: vCenter.Y - pH/2.0},
		{X: vCenter.X - pW/2.0, Y: vPos.Y - pH - 8.0},
		{X: vCenter.X - pW/2.0, Y: vPos.Y + vDims.Y + 8.0},
		// Corner offsets
		{X: vPos.X - pW - 6.0, Y: vPos.Y - pH - 6.0},
		{X: vPos.X + vDims.X + 6.0, Y: vPos.Y - pH - 6.0},
		{X: vPos.X - pW - 6.0, Y: vPos.Y + vDims.Y + 6.0},
		{X: vPos.X + vDims.X + 6.0, Y: vPos.Y + vDims.Y + 6.0},
	}

	for _, c := range candidates {
		if !isSolid(c) {
			return c
		}
	}

	// 2. Secondary candidates: Expanding rings around vehicle center
	for r := 32.0; r <= 160.0; r += 12.0 {
		for a := 0; a < 16; a++ {
			ang := float64(a) * (math.Pi / 8.0)
			c := gvec.Vec2{
				X: vCenter.X + math.Cos(ang)*r - pW/2.0,
				Y: vCenter.Y + math.Sin(ang)*r - pH/2.0,
			}
			if !isSolid(c) {
				return c
			}
		}
	}

	// 3. Fallback: closest water to vehicle center
	if safePos, ok := o.FindNearestWater(vCenter, pW, pH, baseStation); ok {
		return safePos
	}

	return vPos
}

