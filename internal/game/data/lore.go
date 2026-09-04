package data

// DefaultLoreEntries is the compile-time lore database.
// Unlock state is stored on each entry's Unlocked field at runtime.
var DefaultLoreEntries = []*LoreEntry{
	{
		ID:            "raw_fish_caught",
		Category:      "Fauna",
		Title:         "Cave Fish Bio-Scan",
		TriggerType:   "catch",
		TriggerTarget: "raw fish",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP CLASSIFIED TELEMETRY",
				Text:   "Specimen: Cave Fish. Yields standard biomass. Consuming after cooking restores moderate stamina. Do not engage in unnecessary hunting; focus on resource extraction quotas.",
			},
			{
				Header: "TRITON BIOLOGIST JOURNAL - DR. ARIS",
				Text:   "These small fish navigate the shallow caves with incredible precision using lateral line vibration sensors. They are harmless, but their sudden darting paths have startled me more than once in the dark.",
			},
		},
	},
	{
		ID:            "raw_crab_caught",
		Category:      "Fauna",
		Title:         "Cave Crab Bio-Scan",
		TriggerType:   "catch",
		TriggerTarget: "raw crab",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP CLASSIFIED TELEMETRY",
				Text:   "Specimen: Cave Crab. Subject withdraws into a calcified shell when exposed to light or proximity. Biomass is edible but low-yield.",
			},
			{
				Header: "TRITON EXPEDITION - OBS-44",
				Text:   "We noticed the crabs anchor themselves firmly to horizontal rock ledges. When we flashed our heavy submersible lights on them, they froze instantly. They seem to use the darkness as their primary defense.",
			},
		},
	},
	{
		ID:            "shatter_bulb_popped",
		Category:      "Flora",
		Title:         "Shatter-Bulb Analysis",
		TriggerType:   "pop",
		TriggerTarget: "shatter-bulb",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP HAZARD ASSESSMENT",
				Text:   "Shatter-Bulb gas pockets are pressurized. Popping them releases breathable O2 but emits high-frequency acoustic waves. Predators within 150 meters will detect this sound profile.",
			},
			{
				Header: "TRITON EXPEDITION NOTES",
				Text:   "We've been using the Shatter-Bulbs to refill our tanks in emergencies. It feels like breathing glass, but it keeps the lungs going. Just be careful not to pop them when the larger things are swimming nearby.",
			},
		},
	},
	{
		ID:            "mined_copper",
		Category:      "Geology",
		Title:         "Copper Vein Salvage",
		TriggerType:   "mine",
		TriggerTarget: "copper",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP RESOURCE LOG",
				Text:   "Copper Node: High purity detected. Primary use: wiring harnesses, circuit integration, scanner tools, and solar assemblies.",
			},
			{
				Header: "TRITON ENGI NOTE",
				Text:   "We found copper deposits embedded in the cave walls. It's stable enough to mine with a standard pick, but the deeper we go, the more the local electrical anomalies seem to corrode the copper tools.",
			},
		},
	},
	{
		ID:            "mined_abyssal",
		Category:      "Geology",
		Title:         "Abyssal Shard Discovery",
		TriggerType:   "mine",
		TriggerTarget: "abyssal ore",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP URGENT MEMO",
				Text:   "Classified Asset: Abyssal Ore. Highly dense radioactive isotope. Warning: Retrieve all samples. Under no circumstances should field personnel discuss the isotope's energetic output on unencrypted lines.",
			},
			{
				Header: "TRITON LAST LOG - COMMANDER STERLING",
				Text:   "We found it. The Abyssal Ore is glowing in the pitch black. The reactor is overloaded, and the hull integrity is down to 30%. Aethercorp told us to hold our position, but the storm is coming from the deep. We need to build the rocket now. If anyone finds this... don't look for us.",
			},
		},
	},
	{
		ID:            "wreck_research_log",
		Category:      "Wreckage",
		Title:         "Triton-01 Science Telemetry",
		TriggerType:   "read",
		TriggerTarget: "wreck_research_log",
		Paragraphs: []Paragraph{
			{
				Header: "RESEARCH TENDER SURVEY LOG",
				Text:   "Preliminary bathymetric surveys indicate abnormal geothermal venting and bioluminescent mega-flora. Scout submersible schematics verified for shallow exploration. Automated life-support remains nominal above 40 meters.",
			},
			{
				Header: "CHIEF SURVEYOR CHEN",
				Text:   "The seabed here is alive. Small crustacean specimens have already begun tearing into our discarded conduit covers, using alloy scraps as makeshift carapaces. Fascinating adaptability.",
			},
		},
	},
	{
		ID:            "wreck_transport_manifest",
		Category:      "Wreckage",
		Title:         "Cargo Hauler Heavy Manifest",
		TriggerType:   "read",
		TriggerTarget: "wreck_transport_manifest",
		Paragraphs: []Paragraph{
			{
				Header: "SUBMERSIBLE TRANSPORT LOGISTICS",
				Text:   "Manifest: Heavy Mech Kit assembly, high-output depth module prototypes, and auxiliary reinforced plating. Vessel encountered severe magnetic shear near the mid-trench boundary.",
			},
			{
				Header: "DAMAGE ASSESSMENT",
				Text:   "Multiple bulkheads ruptured on impact. Lower deck floor grating collapsed, creating hazardous vertical drops into flooded machinery bays. Electrical conduits arcing intermittently across engineering corridors.",
			},
		},
	},
	{
		ID:            "wreck_flagship_blackbox",
		Category:      "Wreckage",
		Title:         "AetherCorp Flagship Black Box",
		TriggerType:   "read",
		TriggerTarget: "wreck_flagship_blackbox",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP COMMAND TELEMETRY",
				Text:   "CRITICAL LOCKDOWN: Catastrophic hull failure at abyssal depths. Deep Vault sealed behind Reinforced Blast Bulkheads. Automated defense systems engaged. Escape Rocket staging schematics preserved within secure vault.",
			},
			{
				Header: "FINAL TRANSMISSION - ADMIRAL VANCE",
				Text:   "The pressure is crushing the outer armor like tin foil. Deep-sea organic tendrils are breaching through the cracks in the blast shields. To whoever penetrates this bulkhead: the escape rocket is your only way off this planet. Drill through.",
			},
		},
	},
	{
		ID:            "scrap_crab_bio",
		Category:      "Fauna",
		Title:         "Scrap Hermit Crab Bio-Scan",
		TriggerType:   "catch",
		TriggerTarget: "ScrapHermitCrab",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP BIOMETRIC ANALYSIS",
				Text:   "Specimen: Paguroidea Derelictus (Scrap Hermit Crab). Opportunistic scavengers that utilize artificial industrial debris—such as discarded rations tins, pipe joints, and gears—as protective shells. Highly resilient to physical impact when withdrawn.",
			},
			{
				Header: "SALVAGE PROTOCOL",
				Text:   "Harvesters may recover both edible protein and salvageable structural alloys or electronic components from collected specimens.",
			},
		},
	},
}
