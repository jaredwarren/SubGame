package scene

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// update handles player input, side-scroller swimming physics, and checks exit transitions.
func (c *CaveScene) update(g CaveContext) error {
	if c.scrollActive {
		return c.updateScrollTransition(g)
	}

	g.SetWeaverTrackingTimer(0.0)

	c.spawnPlankton(g)

	entityRuntime := g.NewEntityRuntime()
	c.updateEntities(g, entityRuntime)

	activeVehicle := g.GetActiveVehicle()
	inp := g.GetInput()

	if activeVehicle != nil {
		c.updateVehicle(g, inp, activeVehicle)
	} else {
		p := g.GetPlayer()
		if p.Pos.Y < -8 {
			g.ExitCave()
			return nil
		}
		c.updatePlayer(g, inp, p, entityRuntime)
	}

	c.updateBoundaryTransitions(g)

	return nil
}

func (c *CaveScene) updateScrollTransition(g CaveContext) error {
	c.scrollTimer++
	if c.scrollTimer >= 45 {
		c.scrollActive = false
		g.HorizontalTransition(c.newTrenchX, c.newTrenchY, c.newTrenchKey, c.newCave, c.newCaveGrid, c.newNodes, c.newEntities)

		p := g.GetPlayer()
		caveW := len(c.CaveGrid)

		var width float64 = p.Width
		var height float64 = p.Height
		if v := g.GetActiveVehicle(); v != nil {
			dims := v.GetDimensions()
			width = dims.X
			height = dims.Y
		}

		var targetX float64
		var targetY float64

		switch c.scrollDir {
		case 1: // Right
			targetX = 20.0
			targetY = p.Pos.Y
			if v := g.GetActiveVehicle(); v != nil {
				targetY = v.GetPos().Y
			}
			for targetY > 0 {
				if !c.IsSolid(g, targetX, targetY, width, height) {
					break
				}
				targetY -= float64(config.TileSize)
			}
		case -1: // Left
			targetX = float64(caveW*config.TileSize) - width - 20.0
			targetY = p.Pos.Y
			if v := g.GetActiveVehicle(); v != nil {
				targetY = v.GetPos().Y
			}
			for targetY > 0 {
				if !c.IsSolid(g, targetX, targetY, width, height) {
					break
				}
				targetY -= float64(config.TileSize)
			}
		case 2: // Down (into deep Shock Kelp cave)
			targetX = float64(caveW/2*config.TileSize) + (float64(config.TileSize)-width)/2.0
			targetY = float64(config.TileSize * 2)
			for targetY < float64(len(c.CaveGrid[0])*config.TileSize) {
				if !c.IsSolid(g, targetX, targetY, width, height) {
					break
				}
				targetY += float64(config.TileSize)
			}
		case -2: // Up (back into shallow seabed chasm)
			chasmMinX, chasmMaxX, chasmTriggerY := float64(0), float64(0), float64(0)
			if chasm, ok := c.newCave.(cave.ChasmProvider); ok && chasm.HasFloorChasm() {
				chasmMinX, chasmMaxX, chasmTriggerY = chasm.GetChasmBounds()
			}
			if chasmMaxX > chasmMinX {
				targetX = (chasmMinX+chasmMaxX)/2.0 - width/2.0
			} else {
				targetX = float64(caveW/2*config.TileSize) - width/2.0
			}
			targetY = chasmTriggerY - height - 8.0
			placed := false
			for step := 0.0; step <= 256.0; step += 4.0 {
				testY := chasmTriggerY - height - 8.0 - step
				if testY >= 0 && !c.IsSolid(g, targetX, testY, width, height) {
					targetY = testY
					placed = true
					break
				}
			}
			if !placed {
				for targetY > 0 {
					if !c.IsSolid(g, targetX, targetY, width, height) {
						break
					}
					targetY -= 4.0
				}
			}
		}

		if targetY < 0 {
			targetY = 0
		}

		p.Vel = gvec.Vec2{}
		if v := g.GetActiveVehicle(); v != nil {
			v.SetPos(gvec.Vec2{X: targetX, Y: targetY})
			p.Pos.X = targetX + (width-p.Width)/2.0
			p.Pos.Y = targetY + (height-p.Height)/2.0
		} else {
			p.Pos.X = targetX
			p.Pos.Y = targetY
		}

		cam := g.GetCamera()
		cam.CenterOn(p.Pos.X, p.Pos.Y, p.Width, p.Height)
	}
	return nil
}

func (c *CaveScene) spawnPlankton(g CaveContext) {
	cam := g.GetCamera()
	particles := g.GetParticles()
	planktonCount := 0
	for _, p := range particles {
		if p.Type == particle.ParticlePlankton {
			planktonCount++
		}
	}
	if planktonCount < 120 {
		for i := 0; i < 2; i++ {
			rx := cam.Pos.X + rand.Float64()*float64(config.ScreenWidth)
			ry := cam.Pos.Y + rand.Float64()*float64(config.ScreenHeight)
			if ry >= 0 {
				g.SpawnPlankton(rx, ry)
			}
		}
	}
}

func (c *CaveScene) updateEntities(g CaveContext, entityRuntime entity.Runtime) {
	var gridW, gridH int
	if len(c.CaveGrid) > 0 {
		gridW = len(c.CaveGrid)
		gridH = len(c.CaveGrid[0])
	}

	for _, ent := range c.Entities {
		ent.Update(entityRuntime)

		if gridW > 0 && gridH > 0 {
			maxX := float64(gridW * config.TileSize)
			maxY := float64(gridH * config.TileSize)
			pos := ent.GetPos()
			dims := ent.GetDimensions()

			if pos.X < 0 {
				pos.X = 0
			} else if pos.X > maxX-dims.X {
				pos.X = maxX - dims.X
			}

			if pos.Y < -32.0 {
				pos.Y = -32.0
			} else if pos.Y > maxY-dims.Y {
				pos.Y = maxY - dims.Y
			}
			ent.SetPos(pos)
		}
	}

	activeCount := 0
	for _, ent := range c.Entities {
		if ent.IsActive() {
			c.Entities[activeCount] = ent
			activeCount++
		}
	}
	c.Entities = c.Entities[:activeCount]
	g.SetCaveEntities(g.GetActiveTrenchKey(), c.Entities)

	g.DrainEntityCommands(entityRuntime)
}

func (c *CaveScene) updateVehicle(g CaveContext, inp InputSource, activeVehicle vehicle.Vehicle) {
	if mech, ok := activeVehicle.(*vehicle.HeavyMech); ok && !mech.IsDrilling {
		if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			cursor := inp.Cursor()
			cam := g.GetCamera()
			worldX := cam.Pos.X + cursor.X
			worldY := cam.Pos.Y + cursor.Y

			mtx := int(worldX) / config.TileSize
			mty := int(worldY) / config.TileSize

			for i := 0; i < len(c.Nodes); i++ {
				node := c.Nodes[i]
				nodeTx, nodeTy := node.GetTilePos()
				if nodeTx == mtx && nodeTy == mty && node.GetHitsToMine() > 0 {
					px := mech.Pos.X + mech.Dimensions.X/2
					py := mech.Pos.Y + mech.Dimensions.Y/2
					nx := float64(nodeTx*config.TileSize + config.TileSize/2)
					ny := float64(nodeTy*config.TileSize + config.TileSize/2)
					if math.Hypot(px-nx, py-ny) <= 120.0 {
						mech.DrillStrike(node)
						audio.Get().PlaySFXVaried("sfx/mech_drill_loop.wav", 0.7, 0.05)
						break
					}
				}
			}
		}
	}
}

func (c *CaveScene) updatePlayer(g CaveContext, inp InputSource, p *player.Player, entityRuntime entity.Runtime) {
	c.handlePlayerMining(g, inp, p, entityRuntime)
	c.handlePlayerMovement(g, inp, p)
}

func (c *CaveScene) handlePlayerMining(g CaveContext, inp InputSource, p *player.Player, entityRuntime entity.Runtime) {
	if !inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	activeItem := p.GetActiveItem()
	if activeItem != nil {
		if _, isDeployable := activeItem.(vehicle.Deployable); isDeployable {
			g.ActivatePlayerItem(activeItem)
			return
		}
		if _, ok := activeItem.(*item.RepairTool); ok {
			g.UseRepairTool()
			audio.Get().PlaySFX("sfx/repair_tool_loop.wav")
			return
		}
		if usable, ok := activeItem.(item.UsableItem); ok {
			ctx := &caveUsableContext{
				scene: c,
				g:     g,
				p:     p,
				inp:   inp,
			}
			if usable.Use(ctx) {
				if p.Hotbar != nil && p.ActiveSlot >= 0 && p.ActiveSlot < len(p.Hotbar.Slots) {
					slot := &p.Hotbar.Slots[p.ActiveSlot]
					if slot.Item != nil {
						slot.Quantity--
						if slot.Quantity <= 0 {
							slot.Item = nil
						}
					}
				}
				p.RecalculateUpgrades()
				return
			}
		}
	}

	p.IsMining = true
	p.MiningAnimTimer = 24

	cursor := inp.Cursor()
	cam := g.GetCamera()
	worldX := cam.Pos.X + cursor.X
	worldY := cam.Pos.Y + cursor.Y

	mtx := int(worldX) / config.TileSize
	mty := int(worldY) / config.TileSize

	for _, ent := range c.Entities {
		sb, ok := ent.(*entity.ShatterBulb)
		if ok && sb.IsActive() {
			pos := sb.GetPos()
			dims := sb.GetDimensions()
			if worldX >= pos.X && worldX < pos.X+dims.X && worldY >= pos.Y && worldY < pos.Y+dims.Y {
				px := p.Pos.X + p.Width/2
				py := p.Pos.Y + p.Height/2
				if math.Hypot(px-(pos.X+dims.X/2), py-(pos.Y+dims.Y/2)) <= 96.0 {
					if bulb, ok := ent.(*entity.ShatterBulb); ok {
						bulb.Pop(entityRuntime)
						audio.Get().PlaySFX("sfx/shatter_bulb_pop.wav")
						unlocked := g.GetStoryManager().TriggerEvent("pop", "shatter-bulb")
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
							audio.Get().PlaySFX("sfx/scanner_complete.wav")
						}
					}
					break
				}
			}
		}
	}

	for i, ent := range c.Entities {
		if !ent.IsActive() {
			continue
		}
		if creature, ok := ent.(entity.PassiveCreature); ok {
			pos := ent.GetPos()
			dims := ent.GetDimensions()
			if worldX >= pos.X && worldX < pos.X+dims.X && worldY >= pos.Y && worldY < pos.Y+dims.Y {
				playerCenter := gvec.Vec2{X: p.Pos.X + p.Width/2, Y: p.Pos.Y + p.Height/2}
				if creature.CanCatch(playerCenter) {
					harvestedItem := creature.GetHarvestedItem()
					if p.Inventory.AddItem(harvestedItem, 1) {
						ent.SetActive(false)
						c.Entities = append(c.Entities[:i], c.Entities[i+1:]...)
						g.SetMineWarning("Caught "+harvestedItem.GetName()+"!", 90, 1)
						audio.Get().PlaySFX("sfx/item_pickup.wav")
						unlocked := g.GetStoryManager().TriggerEvent("catch", harvestedItem.GetName())
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
							audio.Get().PlaySFX("sfx/scanner_complete.wav")
						}
					} else {
						g.SetMineWarning("Inventory full!", 90, 1)
						audio.Get().PlaySFX("sfx/ui_error.wav")
					}
					break
				}
			}
		}
	}

	for i := 0; i < len(c.Nodes); i++ {
		node := c.Nodes[i]
		nodeTx, nodeTy := node.GetTilePos()
		if nodeTx == mtx && nodeTy == mty && node.GetHitsToMine() > 0 {
			px := p.Pos.X + p.Width/2
			py := p.Pos.Y + p.Height/2
			nx := float64(nodeTx*config.TileSize + config.TileSize/2)
			ny := float64(nodeTy*config.TileSize + config.TileSize/2)

			if math.Hypot(px-nx, py-ny) <= 96.0 {
				if node.RequiresMech() {
					g.SetMineWarning("Requires Heavy Mech Drill Arm to harvest", 120, 1)
					audio.Get().PlaySFX("sfx/ui_error.wav")
					continue
				}
				node.SetHitsToMine(node.GetHitsToMine() - 1)
				audio.Get().PlaySFXVaried("sfx/mining_hit.wav", 0.75, 0.05)

				nodeColor := color.RGBA{150, 150, 150, 255}
				if cRgba, ok := node.GetColor().(color.RGBA); ok {
					nodeColor = cRgba
				}
				g.SpawnDebris(nx, ny, nodeColor)

				if node.GetHitsToMine() <= 0 {
					audio.Get().PlaySFX("sfx/ore_break.wav")
					if resName := node.GetRecipeResultName(); resName != "" {
						recipes := g.GetCraftingRecipes()
						for idx := range recipes {
							if recipes[idx].NewResult().GetName() == resName {
								recipes[idx].Unlocked = true
								g.SetMineWarning("Unlocked: "+resName+"!", 120, 1)
								audio.Get().PlaySFX("sfx/pda_unlock_fanfare.wav")
								break
							}
						}
						unlocked := g.GetStoryManager().TriggerEvent("mine", node.GetName())
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
							audio.Get().PlaySFX("sfx/scanner_complete.wav")
						}
					} else {
						p.Inventory.AddItem(node, 1)
						audio.Get().PlaySFX("sfx/item_pickup.wav")
						unlocked := g.GetStoryManager().TriggerEvent("mine", node.GetName())
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
							audio.Get().PlaySFX("sfx/scanner_complete.wav")
						}
					}
					c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
				}
				break
			}
		}
	}
}

func (c *CaveScene) handlePlayerMovement(g CaveContext, inp InputSource, p *player.Player) {
	if p.StunTimer > 0 {
		p.Vel = gvec.Vec2{}
		return
	}

	cam := g.GetCamera()
	cursor := inp.Cursor()
	pScreenX := p.Pos.X + p.Width/2.0 - cam.Pos.X
	pScreenY := p.Pos.Y + p.Height/2.0 - cam.Pos.Y
	p.Facing = math.Atan2(cursor.Y-pScreenY, cursor.X-pScreenX)

	speedProps := p.Speed["cave"]
	swimForce := speedProps.Acceleration
	maxSpeed := speedProps.TopSpeed
	buoyancy := p.Buoyancy
	drag := speedProps.Drag

	isSprinting := inp.IsKeyPressed(ebiten.KeyShift)
	if isSprinting && p.CurrentStamina > 0 {
		swimForce *= 1.5
		maxSpeed *= 1.6
	}
	if g.IsPlayerSlowed() {
		swimForce *= 0.5
		maxSpeed *= 0.5
	}

	swimming := false
	if inp.IsKeyPressed(ebiten.KeyW) || inp.IsKeyPressed(ebiten.KeyArrowUp) {
		p.Vel.Y -= swimForce
		swimming = true
	}
	if inp.IsKeyPressed(ebiten.KeyS) || inp.IsKeyPressed(ebiten.KeyArrowDown) {
		p.Vel.Y += swimForce
		swimming = true
	}
	if inp.IsKeyPressed(ebiten.KeyA) || inp.IsKeyPressed(ebiten.KeyArrowLeft) {
		p.Vel.X -= swimForce
		swimming = true
	}
	if inp.IsKeyPressed(ebiten.KeyD) || inp.IsKeyPressed(ebiten.KeyArrowRight) {
		p.Vel.X += swimForce
		swimming = true
	}

	p.Vel.Y += buoyancy
	p.Vel = p.Vel.Scale(drag)

	speed := p.Vel.Length()
	if speed > maxSpeed {
		p.Vel = p.Vel.Scale(maxSpeed / speed)
	}

	if swimming && speed > 2.0 && rand.Float64() < 0.12 {
		g.SpawnBubble(p.Pos.X+p.Width/2.0, p.Pos.Y+p.Height/2.0)
	}

	c.checkCollisions(g, p)

	isMoving := speed > 0.1
	p.UpdateStats(true, isSprinting && isMoving && swimming)
}

func (c *CaveScene) updateBoundaryTransitions(g CaveContext) {
	p := g.GetPlayer()
	wld := g.GetWorld()
	tx, ty := g.GetActiveTrenchCoords()
	caveW := len(c.CaveGrid)

	playerX := p.Pos.X
	playerY := p.Pos.Y
	playerW := p.Width
	playerH := p.Height
	if v := g.GetActiveVehicle(); v != nil {
		pos := v.GetPos()
		dims := v.GetDimensions()
		playerX = pos.X
		playerY = pos.Y
		playerW = dims.X
		playerH = dims.Y
	}

	if c.IsShallow {
		var info *world.TileTypeInfo
		if tx >= 0 && tx < wld.Width && ty >= 0 && ty < wld.Height {
			tileType := wld.OverworldMap[tx][ty]
			info = world.GetTileInfo(tileType)
		}

		// 1. Check vertical downward transition through seabed floor chasm
		if info != nil && info.Subterranean != nil {
			if chasm, ok := c.ActiveCave.(cave.ChasmProvider); ok && chasm.HasFloorChasm() {
				minX, maxX, triggerY := chasm.GetChasmBounds()
				playerCenterX := playerX + playerW/2.0
				if playerCenterX >= minX && playerCenterX <= maxX && playerY+playerH >= triggerY-4.0 {
					grid := wld.GetSubterraneanCave(tx, ty)
					var deep cave.Cave
					if info.Subterranean.DeepFactory != nil {
						deep = info.Subterranean.DeepFactory(grid, wld, tx, ty)
					} else {
						deep = cave.NewOrganicTrenchCave(grid)
					}
					seed := int64(tx*97 + ty*41 + 5555)
					c.beginCaveTransition(g, tx, ty, caveTransitionDest{
						dir:      2, // ScrollDown
						newTX:    tx,
						newTY:    ty,
						newKey:   fmt.Sprintf("%d_%d%s", tx, ty, info.Subterranean.DeepKeySuffix),
						grid:     grid,
						cave:     deep,
						seed:     seed,
						newCamX:  float64(len(grid)/2*config.TileSize - config.ScreenWidth/2),
						newCamY:  0,
					})
					return
				}
			}
		}

		// 2. Check horizontal transitions (left / right)
		if playerX <= 0 {
			newTx := tx - 1
			if newTx >= 0 && wld.IsShallowTile(newTx, ty) {
				grid := wld.GetCave(newTx, ty)
				c.beginCaveTransition(g, tx, ty, caveTransitionDest{
					dir:     -1,
					newTX:   newTx,
					newTY:   ty,
					newKey:  fmt.Sprintf("%d_%d", newTx, ty),
					grid:    grid,
					cave:    shallowCaveForTile(wld, grid, newTx, ty),
					seed:    int64(newTx*97 + ty*41),
					newCamX: float64(caveW*config.TileSize - config.ScreenWidth),
					newCamY: g.GetCamera().Pos.Y,
				})
			}
		} else if playerX+playerW >= float64(caveW*config.TileSize) {
			newTx := tx + 1
			if newTx < wld.Width && wld.IsShallowTile(newTx, ty) {
				grid := wld.GetCave(newTx, ty)
				c.beginCaveTransition(g, tx, ty, caveTransitionDest{
					dir:     1,
					newTX:   newTx,
					newTY:   ty,
					newKey:  fmt.Sprintf("%d_%d", newTx, ty),
					grid:    grid,
					cave:    shallowCaveForTile(wld, grid, newTx, ty),
					seed:    int64(newTx*97 + ty*41),
					newCamX: 0,
					newCamY: g.GetCamera().Pos.Y,
				})
			}
		}
	} else if !c.IsShallow {
		var info *world.TileTypeInfo
		if tx >= 0 && tx < wld.Width && ty >= 0 && ty < wld.Height {
			tileType := wld.OverworldMap[tx][ty]
			info = world.GetTileInfo(tileType)
		}
		if info != nil && info.Subterranean != nil {
			// 3. Check vertical upward transition from subterranean deep cave back to shallow seabed
			if playerY <= 0 {
				grid := wld.GetCave(tx, ty)
				var shallow cave.Cave
				if info.Subterranean.ShallowFactory != nil {
					shallow = info.Subterranean.ShallowFactory(grid, wld, tx, ty)
				} else if info.CaveFactory != nil {
					shallow = info.CaveFactory(grid, wld, tx, ty)
				} else {
					shallow = cave.NewShallowSeabedCave(grid)
				}

				chasmMinX, chasmMaxX, chasmTriggerY := float64(0), float64(0), float64(0)
				if chasm, ok := shallow.(cave.ChasmProvider); ok && chasm.HasFloorChasm() {
					chasmMinX, chasmMaxX, chasmTriggerY = chasm.GetChasmBounds()
				}
				newCamX := float64(len(grid)/2*config.TileSize - config.ScreenWidth/2)
				if chasmMaxX > chasmMinX {
					newCamX = (chasmMinX+chasmMaxX)/2.0 - float64(config.ScreenWidth)/2
				}
				newCamY := chasmTriggerY - float64(config.ScreenHeight)/2
				if newCamY < 0 {
					newCamY = 0
				}

				c.beginCaveTransition(g, tx, ty, caveTransitionDest{
					dir:     -2, // ScrollUp
					newTX:   tx,
					newTY:   ty,
					newKey:  fmt.Sprintf("%d_%d", tx, ty),
					grid:    grid,
					cave:    shallow,
					seed:    int64(tx*97 + ty*41),
					newCamX: newCamX,
					newCamY: newCamY,
				})
			}
		}
	}
}

// caveTransitionDest holds the destination side of a trench scroll transition.
type caveTransitionDest struct {
	dir                int
	newTX, newTY       int
	newKey             string
	grid               [][]bool
	cave               cave.Cave
	seed               int64
	newCamX, newCamY   float64
}

// beginCaveTransition stashes the current cave as the scroll "old" side and
// prepares the destination cave (lazy-generating nodes/entities when needed).
func (c *CaveScene) beginCaveTransition(g CaveContext, oldTX, oldTY int, dest caveTransitionDest) {
	c.scrollActive = true
	c.scrollTimer = 0
	c.scrollDir = dest.dir

	c.oldCave = c.ActiveCave
	c.oldCaveGrid = c.CaveGrid
	c.oldNodes = c.Nodes
	c.oldEntities = c.Entities
	c.oldTrenchX, c.oldTrenchY = oldTX, oldTY
	c.oldTrenchKey = g.GetActiveTrenchKey()
	c.oldCamX = g.GetCamera().Pos.X
	c.oldCamY = g.GetCamera().Pos.Y

	c.newTrenchX, c.newTrenchY = dest.newTX, dest.newTY
	c.newTrenchKey = dest.newKey
	c.newCaveGrid = dest.grid
	c.newCave = dest.cave

	c.newNodes = g.GetCaveNodes(c.newTrenchKey)
	if c.newNodes == nil {
		c.newNodes = c.newCave.GenerateResources(dest.seed)
		g.SetCaveNodes(c.newTrenchKey, c.newNodes)
	}
	c.newEntities = g.GetCaveEntities(c.newTrenchKey)
	if c.newEntities == nil {
		c.newEntities = c.newCave.GenerateEntities(dest.seed)
		g.SetCaveEntities(c.newTrenchKey, c.newEntities)
	}

	c.newCamX = dest.newCamX
	c.newCamY = dest.newCamY
}

func shallowCaveForTile(wld *world.World, grid [][]bool, tx, ty int) cave.Cave {
	tileType := wld.OverworldMap[tx][ty]
	info := world.GetTileInfo(tileType)
	if info != nil && info.CaveFactory != nil {
		return info.CaveFactory(grid, wld, tx, ty)
	}
	newSpec := world.GetBiomeInfo(wld.BiomeMap[tx][ty])
	var caveSpec *cave.CaveBiomeSpec
	if newSpec != nil {
		caveSpec = newSpec.CaveSpec
	}
	return cave.NewShallowSeabedCaveWithBiome(grid, caveSpec)
}
