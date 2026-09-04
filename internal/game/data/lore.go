package data

// DefaultLoreEntries is the compile-time lore database.
// Unlock state is stored on each entry's Unlocked field at runtime.
var DefaultLoreEntries = []*LoreEntry{
	{
		ID:            "raw_fish_caught",
		Category:      "Fauna",
		Title:         "Cave Fish Bio-Scan",
		TriggerType:   "catch,scan",
		TriggerTarget: "raw fish,cave fish",
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
		TriggerType:   "catch,scan",
		TriggerTarget: "raw crab,cave crab",
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
		TriggerType:   "pop,scan",
		TriggerTarget: "shatter-bulb,shatterbulb",
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
		TriggerType:   "mine,scan",
		TriggerTarget: "copper,copper vein",
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
		TriggerType:   "mine,scan",
		TriggerTarget: "abyssal ore,abyssal shard",
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
		TriggerType:   "read,scan",
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
		TriggerType:   "read,scan",
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
		TriggerType:   "read,scan",
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
		TriggerType:   "catch,scan",
		TriggerTarget: "scraphermitcrab,scrap hermit crab",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP BIOMETRIC ANALYSIS",
				Text:   "Specimen: Paguroidea Derelictus (Scrap Hermit Crab). Opportunistic scavengers that utilize artificial industrial debris—such as discarded rations tins, pipe joints, and gears—as protective shells. Highly resilient to physical impact when withdrawn.",
			},
			{
				Header: "SALVAGE PROTOCOL & DEFENSE TIPS",
				Text:   "DEFENSE & HARVEST: Shell protects against direct impacts. Wait until the crab emerges to swim, or approach from behind. Harvesters may recover both edible protein and salvageable structural alloys or electronic scrap.",
			},
		},
	},
	{
		ID:            "thermocline_rammer_bio",
		Category:      "Fauna",
		Title:         "Thermocline Rammer Telemetry",
		TriggerType:   "scan",
		TriggerTarget: "thermocline_rammer,thermocline rammer",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP THREAT ASSESSMENT - APEX CHARGER",
				Text:   "Specimen: Bathyal Perciformes (Thermocline Rammer). Heavily armored cranial plate built for hyper-velocity hydrodynamic ramming. Possesses no optical eyes, navigating entirely via lateral vibration sensors that track engine cavitation and fast swimming.",
			},
			{
				Header: "TRITON SURVIVAL PROTOCOL & DEFENSE TIPS",
				Text:   "COUNTERPLAY: Kill thrusters or swim gently when Rammer enters hunting posture. If charged, execute a sharp 90-degree dodge toward solid rock—its momentum will cause it to crash into the cave wall, stunning it for 3-4 seconds. Unlocks Decoy Launcher schematics.",
			},
		},
	},
	{
		ID:            "electro_weaver_bio",
		Category:      "Fauna",
		Title:         "Electro-Weaver Telemetry",
		TriggerType:   "scan",
		TriggerTarget: "electro_weaver,electro weaver,electro-weaver",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP THREAT ASSESSMENT - ELECTRICAL PREDATOR",
				Text:   "Specimen: Anguilliformes Galvanicus (Electro-Weaver). Bioluminescent serpentine stalker with electro-receptive dorsal crests. Highly attracted to artificial electromagnetic signatures, including high-intensity sub lights and active vehicle power generators.",
			},
			{
				Header: "TRITON SURVIVAL PROTOCOL & DEFENSE TIPS",
				Text:   "COUNTERPLAY: When electro-static interference is detected (visual HUD jittering), immediately extinguish all headlamps and cut submersible engines. The stalker will break lock after 5 seconds of total darkness. Sonar pings provoke immediate hostility. Unlocks Sonar Amplifier schematics.",
			},
		},
	},
	{
		ID:            "voltaic_lurker_bio",
		Category:      "Fauna",
		Title:         "Voltaic Lurker Telemetry",
		TriggerType:   "scan",
		TriggerTarget: "voltaic_lurker,voltaic lurker",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP BIOMETRIC ANALYSIS",
				Text:   "Specimen: Voltaic Lurker. Ambient ambush organism capable of discharging high-voltage electromagnetic shocks within a 4-meter perimeter when disturbed.",
			},
			{
				Header: "TRITON SURVIVAL PROTOCOL & DEFENSE TIPS",
				Text:   "COUNTERPLAY: Maintain minimum 5-meter standoff distance. The creature's bio-electric organs are vulnerable to chemical irritants. Discharging Chemical Deterrents neutralizes its arc coils temporarily, enabling safe passage. Unlocks Chemical Discharger schematics.",
			},
		},
	},
	{
		ID:            "false_bulb_snare_bio",
		Category:      "Fauna",
		Title:         "False-Bulb Snare Telemetry",
		TriggerType:   "scan",
		TriggerTarget: "false_bulb_snare,false-bulb snare,false bulb snare",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP HAZARD ASSESSMENT",
				Text:   "Specimen: Mimetic Siphonophore (False-Bulb Snare). Ceiling-suspended ambush organism that mimics the luminous gas sac of an oxygen-producing Shatter-Bulb.",
			},
			{
				Header: "TRITON SURVIVAL PROTOCOL & DEFENSE TIPS",
				Text:   "COUNTERPLAY: Keep flashlight beam focused directly on the bulb. The creature is photophobic and freezes completely under direct illumination. The moment the light vector leaves its coordinates, it lunges downward with barbed tentacles.",
			},
		},
	},
	{
		ID:            "ink_squid_bio",
		Category:      "Fauna",
		Title:         "Ink Squid Telemetry",
		TriggerType:   "scan",
		TriggerTarget: "ink_squid,ink squid,glow_squid,glow squid",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP BIOMETRIC ANALYSIS",
				Text:   "Specimen: Teuthida Melanos (Ink Squid). Highly agile cephalopod inhabiting mid-depth trenches. Emits dense neuro-toxic ink clouds that disrupt diver navigation and blind vehicle optical sensors.",
			},
			{
				Header: "TRITON SURVIVAL PROTOCOL & DEFENSE TIPS",
				Text:   "COUNTERPLAY: Do not corner ink squids in narrow fissures. Dispersed ink clouds can be neutralized with vehicle air filtration or chemical neutralizers. Harvesting yield includes concentrated ink pigments for defensive chemical munitions. Unlocks Chemical Deterrent.",
			},
		},
	},
	{
		ID:            "sand_viper_bio",
		Category:      "Fauna",
		Title:         "Sand Viper Telemetry",
		TriggerType:   "scan",
		TriggerTarget: "sand_viper,sand viper",
		Paragraphs: []Paragraph{
			{
				Header: "AETHERCORP HAZARD ASSESSMENT",
				Text:   "Specimen: Psammophis Benthos (Sand Viper). Ambush predator buried beneath loose silicate floor sediment. Strikes with venomous jaws when divers swim near the cave bed.",
			},
			{
				Header: "TRITON SURVIVAL PROTOCOL & DEFENSE TIPS",
				Text:   "COUNTERPLAY: Watch for shifting seabed particle ripples and faint dorsal spines. Maintain positive buoyancy 2-3 meters above the cave floor to stay outside strike range. Hand Scanner pulses reveal burrowed vipers through sediment.",
			},
		},
	},
}
