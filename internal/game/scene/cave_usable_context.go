package scene

import (
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Ensure caveUsableContext implements item.UsableContext.
var _ item.UsableContext = (*caveUsableContext)(nil)

type caveUsableContext struct {
	scene *CaveScene
	g     CaveContext
	p     *player.Player
	inp   InputSource
}

func (c *caveUsableContext) PlayerPos() gvec.Vec2 {
	return c.p.Pos
}

func (c *caveUsableContext) PlayerDims() gvec.Vec2 {
	return gvec.Vec2{X: c.p.Width, Y: c.p.Height}
}

func (c *caveUsableContext) CursorWorldPos() gvec.Vec2 {
	cursor := c.inp.Cursor()
	cam := c.g.GetCamera()
	return gvec.Vec2{X: cam.Pos.X + cursor.X, Y: cam.Pos.Y + cursor.Y}
}

func (c *caveUsableContext) SpawnSonicDecoy(pos gvec.Vec2, vel gvec.Vec2) {
	decoy := entity.NewSonicDecoy(pos.X, pos.Y, vel)
	c.scene.Entities = append(c.scene.Entities, decoy)
	c.g.SetCaveEntities(c.g.GetActiveTrenchKey(), c.scene.Entities)
}

func (c *caveUsableContext) SpawnDeterrentCloud(pos gvec.Vec2) {
	cloud := entity.NewDeterrentCloud(pos.X, pos.Y)
	c.scene.Entities = append(c.scene.Entities, cloud)
	c.g.SetCaveEntities(c.g.GetActiveTrenchKey(), c.scene.Entities)
}

func (c *caveUsableContext) SetMineWarning(msg string, duration, level int) {
	c.g.SetMineWarning(msg, duration, level)
}
