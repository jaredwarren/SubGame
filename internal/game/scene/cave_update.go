package scene

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Update handles player input, side-scroller swimming physics, and checks exit transitions.
func (c *CaveScene) Update(g GameContext) error {
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

func (c *CaveScene) updateScrollTransition(g GameContext) error {
	c.scrollTimer++
	if c.scrollTimer >= 45 {
		c.scrollActive = false
		g.HorizontalTransition(c.newTrenchX, c.newTrenchY, c.newTrenchKey, c.newCave, c.newCaveGrid, c.newNodes, c.newEntities)

		p := g.GetPlayer()
		caveW := len(c.CaveGrid)

		var width float64 = p.Width
		if v := g.GetActiveVehicle(); v != nil {
			width = v.GetDimensions().X
		}

		var targetX float64
		if c.scrollDir == 1 {
			targetX = 20.0
		} else {
			targetX = float64(caveW*config.TileSize) - width - 20.0
		}

		// Find a safe Y coordinate (not solid) by sliding upwards
		safeY := p.Pos.Y
		var checkW, checkH float64 = p.Width, p.Height
		if v := g.GetActiveVehicle(); v != nil {
			vDims := v.GetDimensions()
			checkW, checkH = vDims.X, vDims.Y
			safeY = v.GetPos().Y
		}

		for safeY > 0 {
			if !c.IsSolid(g, targetX, safeY, checkW, checkH) {
				break
			}
			safeY -= float64(config.TileSize)
		}
		if safeY < 0 {
			safeY = 0
		}

		p.Vel = gvec.Vec2{}
		if v := g.GetActiveVehicle(); v != nil {
			v.SetPos(gvec.Vec2{X: targetX, Y: safeY})
			p.Pos.X = targetX + (checkW-p.Width)/2.0
			p.Pos.Y = safeY + (checkH-p.Height)/2.0
		} else {
			p.Pos.X = targetX
			p.Pos.Y = safeY
		}

		cam := g.GetCamera()
		cam.CenterOn(p.Pos.X, p.Pos.Y, p.Width, p.Height)
	}
	return nil
}

func (c *CaveScene) spawnPlankton(g GameContext) {
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

func (c *CaveScene) updateEntities(g GameContext, entityRuntime entity.Runtime) {
	for _, ent := range c.Entities {
		ent.Update(entityRuntime)
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

func (c *CaveScene) updateVehicle(g GameContext, inp InputSource, activeVehicle vehicle.Vehicle) {
	if mech, ok := activeVehicle.(*vehicle.HeavyMech); ok && !mech.IsDrilling {
		if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			cursor := inp.Cursor()
			camX := mech.Pos.X - config.ScreenWidth/2 + mech.Dimensions.X/2
			camY := mech.Pos.Y - config.ScreenHeight/2 + mech.Dimensions.Y/2
			worldX := camX + cursor.X
			worldY := camY + cursor.Y

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
						break
					}
				}
			}
		}
	}
}

func (c *CaveScene) updatePlayer(g GameContext, inp InputSource, p *player.Player, entityRuntime entity.Runtime) {
	c.handlePlayerMining(g, inp, p, entityRuntime)
	c.handlePlayerMovement(g, inp, p)
}

func (c *CaveScene) handlePlayerMining(g GameContext, inp InputSource, p *player.Player, entityRuntime entity.Runtime) {
	if !inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	p.IsMining = true
	p.MiningAnimTimer = 24

	cursor := inp.Cursor()
	camX := p.Pos.X - config.ScreenWidth/2 + p.Width/2
	camY := p.Pos.Y - config.ScreenHeight/2 + p.Height/2
	worldX := camX + cursor.X
	worldY := camY + cursor.Y

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
						unlocked := g.GetStoryManager().TriggerEvent("pop", "shatter-bulb")
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
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
						unlocked := g.GetStoryManager().TriggerEvent("catch", harvestedItem.GetName())
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
						}
					} else {
						g.SetMineWarning("Inventory full!", 90, 1)
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
					continue
				}
				node.SetHitsToMine(node.GetHitsToMine() - 1)

				nodeColor := color.RGBA{150, 150, 150, 255}
				if cRgba, ok := node.GetColor().(color.RGBA); ok {
					nodeColor = cRgba
				}
				g.SpawnDebris(nx, ny, nodeColor)

				if node.GetHitsToMine() <= 0 {
					if bpNode, ok := node.(*resource.BlueprintNode); ok {
						for idx := range CraftingRecipes {
							if CraftingRecipes[idx].NewResult().GetName() == bpNode.RecipeResultName {
								CraftingRecipes[idx].Unlocked = true
								g.SetMineWarning("Unlocked: "+bpNode.RecipeResultName+"!", 120, 1)
								break
							}
						}
						unlocked := g.GetStoryManager().TriggerEvent("mine", bpNode.GetName())
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
						}
					} else {
						p.Inventory.AddItem(node, 1)
						unlocked := g.GetStoryManager().TriggerEvent("mine", node.GetName())
						if unlocked != nil {
							g.SetMineWarning("Decrypted PDA Log: "+unlocked.Title, 120, 1)
						}
					}
					c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
				}
				break
			}
		}
	}
}

func (c *CaveScene) handlePlayerMovement(g GameContext, inp InputSource, p *player.Player) {
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

func (c *CaveScene) updateBoundaryTransitions(g GameContext) {
	if !c.IsShallow {
		return
	}

	p := g.GetPlayer()
	wld := g.GetWorld()
	tx, ty := g.GetActiveTrenchCoords()
	caveW := len(c.CaveGrid)

	playerX := p.Pos.X
	playerW := p.Width
	if v := g.GetActiveVehicle(); v != nil {
		playerX = v.GetPos().X
		playerW = v.GetDimensions().X
	}

	if playerX <= 0 {
		newTx := tx - 1
		if newTx >= 0 && wld.OverworldMap[newTx][ty] == world.TileWater {
			// Trigger transition left
			c.scrollActive = true
			c.scrollTimer = 0
			c.scrollDir = -1

			c.oldCave = c.ActiveCave
			c.oldCaveGrid = c.CaveGrid
			c.oldNodes = c.Nodes
			c.oldEntities = c.Entities
			c.oldTrenchX, c.oldTrenchY = tx, ty
			c.oldTrenchKey = g.GetActiveTrenchKey()
			c.oldCamX = g.GetCamera().Pos.X
			c.oldCamY = g.GetCamera().Pos.Y

			c.newTrenchX, c.newTrenchY = newTx, ty
			c.newTrenchKey = fmt.Sprintf("%d_%d", newTx, ty)
			c.newCaveGrid = wld.GetCave(newTx, ty)
			c.newCave = cave.NewShallowSeabedCave(c.newCaveGrid)
			c.newNodes = g.GetCaveNodes(c.newTrenchKey)
			if c.newNodes == nil {
				c.newNodes = c.newCave.GenerateResources(int64(newTx*97 + ty*41))
				g.SetCaveNodes(c.newTrenchKey, c.newNodes)
			}
			c.newEntities = g.GetCaveEntities(c.newTrenchKey)
			if c.newEntities == nil {
				c.newEntities = c.newCave.GenerateEntities(int64(newTx*97 + ty*41))
				g.SetCaveEntities(c.newTrenchKey, c.newEntities)
			}
			c.newCamX = float64(caveW*config.TileSize - config.ScreenWidth)
			c.newCamY = c.oldCamY
		}
	} else if playerX+playerW >= float64(caveW*config.TileSize) {
		newTx := tx + 1
		if newTx < wld.Width && wld.OverworldMap[newTx][ty] == world.TileWater {
			// Trigger transition right
			c.scrollActive = true
			c.scrollTimer = 0
			c.scrollDir = 1

			c.oldCave = c.ActiveCave
			c.oldCaveGrid = c.CaveGrid
			c.oldNodes = c.Nodes
			c.oldEntities = c.Entities
			c.oldTrenchX, c.oldTrenchY = tx, ty
			c.oldTrenchKey = g.GetActiveTrenchKey()
			c.oldCamX = g.GetCamera().Pos.X
			c.oldCamY = g.GetCamera().Pos.Y

			c.newTrenchX, c.newTrenchY = newTx, ty
			c.newTrenchKey = fmt.Sprintf("%d_%d", newTx, ty)
			c.newCaveGrid = wld.GetCave(newTx, ty)
			c.newCave = cave.NewShallowSeabedCave(c.newCaveGrid)
			c.newNodes = g.GetCaveNodes(c.newTrenchKey)
			if c.newNodes == nil {
				c.newNodes = c.newCave.GenerateResources(int64(newTx*97 + ty*41))
				g.SetCaveNodes(c.newTrenchKey, c.newNodes)
			}
			c.newEntities = g.GetCaveEntities(c.newTrenchKey)
			if c.newEntities == nil {
				c.newEntities = c.newCave.GenerateEntities(int64(newTx*97 + ty*41))
				g.SetCaveEntities(c.newTrenchKey, c.newEntities)
			}
			c.newCamX = 0
			c.newCamY = c.oldCamY
		}
	}
}
