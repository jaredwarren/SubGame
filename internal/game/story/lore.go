package story

// Paragraph represents a single paragraph of narrative text, with an optional subheader.
type Paragraph struct {
	Header string
	Text   string
}

// LoreEntry represents a single database log or encyclopedia entry that can be unlocked.
type LoreEntry struct {
	ID            string
	Category      string // e.g. "Fauna", "Flora", "Wreckage", "Geology"
	Title         string // The title of the entry displayed in the list
	TriggerType   string // e.g. "scan", "mine", "catch", "depth", "salvage"
	TriggerTarget string // e.g. "electro_weaver", "Titanium", "Void"
	Paragraphs    []Paragraph
	Unlocked      bool
}
