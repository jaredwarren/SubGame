package data

import "github.com/jaredwarren/SubGame/internal/game/item"

// Material identity/visuals live in package item (shared by inventory + resource nodes).
// Re-exported here for catalog browsing.

type MaterialDef = item.MaterialDef

var (
	MaterialTitanium        = item.MaterialTitanium
	MaterialCopper          = item.MaterialCopper
	MaterialQuartz          = item.MaterialQuartz
	MaterialAbyssalOre      = item.MaterialAbyssalOre
	MaterialNickel          = item.MaterialNickel
	MaterialTungsten        = item.MaterialTungsten
	MaterialScrapMetal      = item.MaterialScrapMetal
	MaterialElectronicWaste = item.MaterialElectronicWaste
)
