package story

import "github.com/jaredwarren/SubGame/internal/game/data"

// Lore types and the default entry table are owned by package data.

type (
	Paragraph = data.Paragraph
	LoreEntry = data.LoreEntry
)

// DefaultLoreEntries aliases the compile-time lore database in package data.
var DefaultLoreEntries = data.DefaultLoreEntries
