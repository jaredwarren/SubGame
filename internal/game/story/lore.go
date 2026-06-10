package story

// Paragraph represents a single paragraph of narrative text, with an optional subheader.
type Paragraph struct {
	Header string `json:"header"`
	Text   string `json:"text"`
}

// LoreEntry represents a single database log or encyclopedia entry that can be unlocked.
type LoreEntry struct {
	ID            string      `json:"id"`
	Category      string      `json:"category"`      // e.g. "Fauna", "Flora", "Wreckage", "Geology"
	Title         string      `json:"title"`         // The title of the entry displayed in the list
	TriggerType   string      `json:"trigger_type"`   // e.g. "scan", "mine", "catch", "depth", "salvage"
	TriggerTarget string      `json:"trigger_target"` // e.g. "electro_weaver", "Titanium", "Void"
	Paragraphs    []Paragraph `json:"paragraphs"`
	Unlocked      bool        `json:"-"`
}
