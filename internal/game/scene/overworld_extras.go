package scene

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// InitializeExtras populates the overworld with fish, crates, and vents if not already initialized.
func (o *OverworldScene) InitializeExtras(g OverworldContext) {
	if o.initialized {
		return
	}
	o.initialized = true

	r := rand.New(rand.NewSource(o.World.Seed + 12345))

	// Scan the overworld map to place extras
	for tx := 0; tx < o.World.Width; tx++ {
		for ty := 0; ty < o.World.Height; ty++ {
			tile := o.World.OverworldMap[tx][ty]

			switch tile {
			case world.TileWreckage:
				// Spawn 2-3 cargo crates around each wreckage tile (within a radius of 2-3 tiles)
				numCrates := r.Intn(2) + 2
				for i := 0; i < numCrates; i++ {
					offsetX := (r.Float64() - 0.5) * 192.0
					offsetY := (r.Float64() - 0.5) * 192.0

					px := float64(tx*config.TileSize) + float64(config.TileSize)/2.0 + offsetX
					py := float64(ty*config.TileSize) + float64(config.TileSize)/2.0 + offsetY

					gtx := int(px) / config.TileSize
					gty := int(py) / config.TileSize
					if gtx >= 0 && gtx < o.World.Width && gty >= 0 && gty < o.World.Height {
						if o.World.OverworldMap[gtx][gty] != world.TileLand {
							o.crates = append(o.crates, &oe.FloatingCrate{
								Pos:        gvec.Vec2{X: px, Y: py},
								InitialPos: gvec.Vec2{X: px, Y: py},
								BobOffset:  r.Float64() * 100.0,
							})
						}
					}
				}

			case world.TileWater:
				// If this tile is close to land, spawn a school of fish
				dist := o.World.LandDist[tx][ty]
				if dist >= 1 && dist <= 3 {
					// Sparse shoreline schools — denser spawn tanks the overworld Update loop.
					if r.Float64() < 0.012 {
						schoolSize := r.Intn(3) + 3
						schoolBaseX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
						schoolBaseY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0

						for i := 0; i < schoolSize; i++ {
							fx := schoolBaseX + (r.Float64()-0.5)*32.0
							fy := schoolBaseY + (r.Float64()-0.5)*32.0

							o.fish = append(o.fish, &oe.CosmeticFish{
								Pos:       gvec.Vec2{X: fx, Y: fy},
								BasePos:   gvec.Vec2{X: schoolBaseX, Y: schoolBaseY},
								WobbleVal: r.Float64() * 100.0,
								WobbleSpd: r.Float64()*0.05 + 0.02,
							})
						}
					}
				}
			}
		}
	}

	// Attach ThermalVent entities directly to all TileThermoCave tiles generated in the world.
	ventIndex := 0
	for tx := 0; tx < o.World.Width; tx++ {
		for ty := 0; ty < o.World.Height; ty++ {
			if o.World.OverworldMap[tx][ty] == world.TileThermoCave {
				tileCenterX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
				tileCenterY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0
				o.vents = append(o.vents, oe.NewThermalVent(gvec.Vec2{X: tileCenterX, Y: tileCenterY}, int64(ventIndex*12345)))
				ventIndex++
			}
		}
	}
}

type fishContext struct {
	targetCenter gvec.Vec2
	scene        *OverworldScene
}

func (fc fishContext) TargetCenter() gvec.Vec2 {
	return fc.targetCenter
}

func (fc fishContext) IsSolid(x, y float64) bool {
	return fc.scene.IsSolid(x-2, y-2, 4, 4)
}

type crateContext struct {
	g OverworldContext
}

func (cc crateContext) GetTargetCenter() gvec.Vec2 {
	v := cc.g.GetActiveVehicle()
	if v != nil {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		return gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
	}
	p := cc.g.GetPlayer()
	return gvec.Vec2{X: p.Pos.X + p.Width/2.0, Y: p.Pos.Y + p.Height/2.0}
}

func (cc crateContext) GetTargetDimensions() gvec.Vec2 {
	v := cc.g.GetActiveVehicle()
	if v != nil {
		return v.GetDimensions()
	}
	p := cc.g.GetPlayer()
	return gvec.Vec2{X: p.Width, Y: p.Height}
}

func (cc crateContext) IsPiloting() bool {
	return cc.g.GetActiveVehicle() != nil
}

func (cc crateContext) AddLoot(loot item.Item) bool {
	v := cc.g.GetActiveVehicle()
	if v != nil {
		added := v.GetCargo().AddItem(loot, 1)
		if added {
			cc.g.AddItemToast(loot, 1)
		}
		return added
	}
	p := cc.g.GetPlayer()
	added := p.Inventory.AddItem(loot, 1)
	if added {
		p.RecalculateUpgrades()
		cc.g.NotifyQuestInventoryChanged(loot.GetID())
		cc.g.AddItemToast(loot, 1)
	}
	return added
}

func (cc crateContext) SpawnDebris(x, y float64, clr color.RGBA) {
	cc.g.SpawnDebris(x, y, clr)
}

func (cc crateContext) TriggerScreenShake(duration int, intensity float64) {
	cc.g.TriggerScreenShake(duration, intensity)
}

func (cc crateContext) SetMineWarning(msg string, duration, level int) {
	cc.g.SetMineWarning(msg, duration, level)
}

type ventContext struct {
	g OverworldContext
}

func (vc ventContext) GetTicks() float64 {
	return vc.g.GetTicks()
}

func (vc ventContext) SpawnBubble(x, y float64) {
	vc.g.SpawnBubble(x, y)
}

func (vc ventContext) TriggerScreenShake(duration int, intensity float64) {
	vc.g.TriggerScreenShake(duration, intensity)
}

func (vc ventContext) GetTargetCenter() gvec.Vec2 {
	v := vc.g.GetActiveVehicle()
	if v != nil {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		return gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
	}
	p := vc.g.GetPlayer()
	return gvec.Vec2{X: p.Pos.X + p.Width/2.0, Y: p.Pos.Y + p.Height/2.0}
}

func (vc ventContext) GetTargetDimensions() gvec.Vec2 {
	v := vc.g.GetActiveVehicle()
	if v != nil {
		return v.GetDimensions()
	}
	p := vc.g.GetPlayer()
	return gvec.Vec2{X: p.Width, Y: p.Height}
}

func (vc ventContext) IsPiloting() bool {
	return vc.g.GetActiveVehicle() != nil
}

func (vc ventContext) ApplyTargetForce(force gvec.Vec2) {
	v := vc.g.GetActiveVehicle()
	if v != nil {
		v.ApplyForce(force)
	} else {
		p := vc.g.GetPlayer()
		p.Vel = p.Vel.Add(force)
	}
}

func (vc ventContext) DamageTarget(damage float64) {
	v := vc.g.GetActiveVehicle()
	if v != nil {
		v.TakeDamage(damage)
	} else {
		p := vc.g.GetPlayer()
		p.CurrentHealth -= damage
	}
}

type whirlpoolContext struct {
	scene *OverworldScene
	g     OverworldContext
}

func (wc whirlpoolContext) BaseStationPos() gvec.Vec2 {
	return wc.g.GetBaseStation().Pos
}

func (wc whirlpoolContext) FindSafeSpawnPos(baseStationPos gvec.Vec2) gvec.Vec2 {
	rng := rand.New(rand.NewSource(int64(wc.g.GetTicks()) + wc.scene.World.Seed))
	return wc.scene.FindSafeWhirlpoolSpawnPos(baseStationPos, rng)
}

func (o *OverworldScene) FindSafeWhirlpoolSpawnPos(baseStationPos gvec.Vec2, rng *rand.Rand) gvec.Vec2 {
	w := o.World
	var candidates []gvec.Vec2
	for tx := 0; tx < w.Width; tx++ {
		for ty := 0; ty < w.Height; ty++ {
			if w.OverworldMap[tx][ty] == world.TileWater {
				// Land distance check: must be far from land
				if w.LandDist[tx][ty] >= 3 {
					tileX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
					tileY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0

					// Distance check to player spawn / base station (at least 15 tiles = 960 pixels)
					dist := math.Hypot(tileX-baseStationPos.X, tileY-baseStationPos.Y)
					if dist >= 960.0 {
						candidates = append(candidates, gvec.Vec2{X: tileX, Y: tileY})
					}
				}
			}
		}
	}

	if len(candidates) > 0 {
		idx := rng.Intn(len(candidates))
		return candidates[idx]
	}

	// Fallback: search for any water tile if no deep/far water tile is available
	for tx := 0; tx < w.Width; tx++ {
		for ty := 0; ty < w.Height; ty++ {
			if w.OverworldMap[tx][ty] == world.TileWater {
				tileX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
				tileY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0
				candidates = append(candidates, gvec.Vec2{X: tileX, Y: tileY})
			}
		}
	}
	if len(candidates) > 0 {
		idx := rng.Intn(len(candidates))
		return candidates[idx]
	}

	// Absolute fallback to map center
	return gvec.Vec2{
		X: float64(w.Width*config.TileSize) / 2.0,
		Y: float64(w.Height*config.TileSize) / 2.0,
	}
}

// UpdateExtras runs update logic for cosmetic fish, crates, and vents.
func (o *OverworldScene) UpdateExtras(g OverworldContext) {
	o.InitializeExtras(g)

	p := g.GetPlayer()
	pPos := p.Pos
	pDims := gvec.Vec2{X: p.Width, Y: p.Height}
	pCenter := gvec.Vec2{X: pPos.X + pDims.X/2.0, Y: pPos.Y + pDims.Y/2.0}

	var targetCenter gvec.Vec2

	activeVeh := g.GetActiveVehicle()
	if activeVeh != nil {
		vPos := activeVeh.GetPos()
		vDims := activeVeh.GetDimensions()
		targetCenter = gvec.Vec2{X: vPos.X + vDims.X/2.0, Y: vPos.Y + vDims.Y/2.0}
	} else {
		targetCenter = pCenter
	}

	fc := fishContext{
		targetCenter: targetCenter,
		scene:        o,
	}

	// Only simulate fish near the camera. Distant schools stay parked at their last pose.
	cam := g.GetCamera()
	fishMinX := cam.Pos.X - 200
	fishMaxX := cam.Pos.X + float64(config.ScreenWidth) + 200
	fishMinY := cam.Pos.Y - 200
	fishMaxY := cam.Pos.Y + float64(config.ScreenHeight) + 200
	for _, f := range o.fish {
		if f.Pos.X < fishMinX || f.Pos.X > fishMaxX || f.Pos.Y < fishMinY || f.Pos.Y > fishMaxY {
			continue
		}
		f.Update(fc)
	}

	// Update vents (push force, heat damage, bubble particle spawning)
	vc := ventContext{g: g}
	for _, v := range o.vents {
		v.Update(vc)
	}

	// Update crates (collisions, looting, respawn timer)
	cc := crateContext{g: g}
	for _, c := range o.crates {
		c.Update(cc)
	}
}

// DrawExtras draws cosmetic fish, crates, and vents.
func (o *OverworldScene) DrawExtras(g OverworldContext, screen *ebiten.Image) {
	cam := g.GetCamera()
	camX, camY := cam.Pos.X, cam.Pos.Y
	ticks := g.GetTicks()
	mult := GetOverworldLightMultiplier(g.GetTimeOfDay())

	// Draw vents (on the bottom layer)
	for _, v := range o.vents {
		if v.Pos.X >= camX-100 && v.Pos.X <= camX+float64(config.ScreenWidth)+100 &&
			v.Pos.Y >= camY-100 && v.Pos.Y <= camY+float64(config.ScreenHeight)+100 {
			v.Draw(screen, camX, camY, ticks, mult)
		}
	}

	// Draw crates
	for _, c := range o.crates {
		if !c.Collected && c.Pos.X >= camX-50 && c.Pos.X <= camX+float64(config.ScreenWidth)+50 &&
			c.Pos.Y >= camY-50 && c.Pos.Y <= camY+float64(config.ScreenHeight)+50 {
			c.Draw(screen, camX, camY, ticks, mult)
		}
	}

	// Draw fish
	for _, f := range o.fish {
		if f.Pos.X >= camX-50 && f.Pos.X <= camX+float64(config.ScreenWidth)+50 &&
			f.Pos.Y >= camY-50 && f.Pos.Y <= camY+float64(config.ScreenHeight)+50 {
			f.Draw(screen, camX, camY, ticks, mult)
		}
	}
}
