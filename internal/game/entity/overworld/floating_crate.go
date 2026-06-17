package overworld

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/light"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// CrateContext defines the interface needed by FloatingCrate to interact with the game scene.
type CrateContext interface {
	GetTargetCenter() gvec.Vec2
	GetTargetDimensions() gvec.Vec2
	IsPiloting() bool
	AddLoot(loot item.Item) bool
	SpawnDebris(x, y float64, clr color.RGBA)
	TriggerScreenShake(duration int, intensity float64)
	SetMineWarning(msg string, duration, level int)
}

// FloatingCrate represents a lootable cargo crate near shipwrecks.
type FloatingCrate struct {
	Pos          gvec.Vec2
	InitialPos   gvec.Vec2
	BobOffset    float64
	Collected    bool
	RespawnTimer int
}

// Draw renders a wooden cargo crate with a cross board pattern.
func (c *FloatingCrate) Draw(screen *ebiten.Image, camX, camY float64, ticks float64, mult float64) {
	if c.Collected {
		return
	}

	bob := math.Sin(ticks*0.045+c.BobOffset) * 2.0
	sx := float32(c.Pos.X - camX)
	sy := float32(c.Pos.Y - camY + bob)

	const size = 15.0
	const half = size / 2.0

	// Dark wood background
	fillClr := light.ApplyLight(color.RGBA{135, 85, 40, 255}, mult)
	vector.FillRect(screen, sx-half, sy-half, size, size, fillClr, false)

	// Lighter wood border & planks cross lines
	strokeClr := light.ApplyLight(color.RGBA{200, 130, 60, 255}, mult)
	vector.StrokeRect(screen, sx-half, sy-half, size, size, 1.2, strokeClr, false)
	vector.StrokeLine(screen, sx-half+2, sy-half+2, sx+half-2, sy+half-2, 1.0, strokeClr, false)
	vector.StrokeLine(screen, sx+half-2, sy-half+2, sx-half+2, sy+half-2, 1.0, strokeClr, false)

	// Specular highlight pixel on the top edge
	highlightClr := light.ApplyLight(color.RGBA{255, 255, 255, 120}, mult)
	vector.FillRect(screen, sx-half+2, sy-half+2, 1.5, 1.5, highlightClr, false)
}

func (c *FloatingCrate) Update(g CrateContext) {
	targetCenter := g.GetTargetCenter()
	targetDims := g.GetTargetDimensions()

	if c.Collected {
		c.RespawnTimer--
		if c.RespawnTimer <= 0 {
			distToPlayer := math.Hypot(targetCenter.X-c.InitialPos.X, targetCenter.Y-c.InitialPos.Y)
			if distToPlayer > 400.0 {
				c.Collected = false
				c.Pos = c.InitialPos
			} else {
				c.RespawnTimer = 60
			}
		}
		return
	}

	// Proximity check for collection
	dist := math.Hypot(targetCenter.X-c.Pos.X, targetCenter.Y-c.Pos.Y)
	colRadius := 16.0 + max(targetDims.X, targetDims.Y)/2.0
	if dist < colRadius {
		// Loot generation
		loot := c.GenerateLoot()

		added := g.AddLoot(loot)

		if added {
			c.Collected = true
			c.RespawnTimer = 7200 // 2 minutes (120 seconds * 60 ticks)

			// FX
			g.SpawnDebris(c.Pos.X, c.Pos.Y, color.RGBA{139, 90, 43, 255})
			g.TriggerScreenShake(10, 1.5)

			g.SetMineWarning("Salvaged: "+loot.GetName()+"!", 120, 1)
		} else {
			g.SetMineWarning("Inventory full! Cannot salvage crate.", 60, 2)
		}
	}
}

// LootItem defines a cargo entry with its cumulative spawn threshold and constructor.
type LootItem struct {
	Threshold float64
	Factory   func() item.Item
}

var crateLootTable = []LootItem{
	{Threshold: 0.50, Factory: func() item.Item { return &item.ScrapMetal{} }},
	{Threshold: 0.75, Factory: func() item.Item { return &item.Titanium{} }},
	{Threshold: 0.90, Factory: func() item.Item { return &item.Copper{} }},
	{Threshold: 0.96, Factory: func() item.Item { return &item.ElectronicWaste{} }},
	{Threshold: 1.00, Factory: func() item.Item { return &item.PowerCell{} }},
}

// GenerateLoot determines the randomized item inside the crate based on weights.
func (c *FloatingCrate) GenerateLoot() item.Item {
	r := rand.Float64()
	for _, entry := range crateLootTable {
		if r < entry.Threshold {
			return entry.Factory()
		}
	}
	return &item.ScrapMetal{} // fallback
}
