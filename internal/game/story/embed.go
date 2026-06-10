package story

import _ "embed"

// LoreJSONBytes embeds the raw JSON content of the lore database.
//go:embed lore.json
var LoreJSONBytes []byte
