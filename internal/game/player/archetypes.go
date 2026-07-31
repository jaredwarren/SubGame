package player

import "github.com/jaredwarren/SubGame/internal/game/data"

// Player balance tables are owned by package data. Aliases preserve existing call sites.

type PlayerDef = data.PlayerDef

var (
	PlayerArchetype = data.PlayerArchetype
	DefaultSpeed    = data.DefaultSpeed
)
