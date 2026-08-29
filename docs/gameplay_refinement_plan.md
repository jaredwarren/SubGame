# Gameplay Refinement & Balance Plan

Goal: make the game **fun, replayable, and challenging — with exploration as the core motivator**. Difficulty should create *tension* that makes exploration feel rewarding, not walls that stop the player. Every change below should be judged against one question: **"Does this make the player want to see what's over the next ridge / down the next trench?"**

---

## 1. Design Pillars

1. **Curiosity is the reward loop.** The player should always have 2–3 visible "I wonder what's there" hooks: a silhouette in the fog, a sonar blip, an unexplored biome tint, a lore gap in the PDA.
2. **Danger is atmosphere, not attrition.** Enemies and hazards exist to make places feel alive and to make the player plan — not to grind them down. Death should be rare, dramatic, and mostly the player's fault.
3. **Depth = commitment.** Going deeper should feel like an expedition (prep, risk, payoff), not a stat check.
4. **Losses sting but never erase progress.** Knowledge (blueprints, lore, map) is permanent; cargo is at risk.

---

## 2. Fix the Loose Ends First (cheap wins)

These are already in the codebase but dormant. Finishing them adds content for nearly free:

| Item | Current state | Plan |
|---|---|---|
| **Scanner upgrade** | Equippable, no effect | Make it the exploration keystone: scanning fauna/flora/geology unlocks lore entries and reveals resource nodes through walls in a short radius. See §6. |
| **Player Energy stat** | Initialized, never drained | Either delete it or repurpose it as the power pool for handheld tools (scanner pulses, flashlight, future tools). Recommend repurposing — it gives food/power cells a second use. |
| **Thermo Caves** | Fully implemented, never placed | Scatter 3–4 Thermo Cave entrances near Thermal Barrens biome / thermal vents. Instant new destination type with existing Rammer + Siphon spawns. |
| **Unused lore triggers** (`scan`, `depth`, `salvage`) | Defined, unfired | Wire them up: `scan` → Scanner, `depth` → first time crossing depth bands, `salvage` → wreckage loot. |
| **Save/load** | Serialize exists for lore only | Ship a basic save (player, inventory, world seed, unlocks, lore). Replayability and session length both depend on this — **highest priority item in the whole plan.** |

---

## 3. Balance Tuning (existing numbers)

### 3.1 Death penalty — currently too harsh for an exploration game
Full inventory wipe on death punishes the exact behavior you want to encourage (venturing out). Replace with:

- **Cargo drop, not cargo wipe:** on death, inventory drops in a marked "lost cargo" beacon at the death site. Recover it by returning — which turns death into a *new expedition prompt* instead of a rage-quit.
- Equipped upgrades (O2 tanks, fins, scanner) stay with the player.
- Optional: cargo beacon despawns after 2–3 in-game days for stakes.

### 3.2 Oxygen — the core tension dial
- Current: 100 base, drain 1.0/s, +60/+140 tanks. The numbers are fine; the *feel* needs work.
- Add escalating feedback: heartbeat/vignette below 30%, HUD warning at 15 s remaining. Panic should come from presentation, not faster drain.
- **Air pockets / ShatterBulbs as route planning:** ShatterBulb already gives +20 O2 — lean into it. Guarantee 1–2 O2 sources per cave beyond ~40 tiles deep so skilled routing extends dives. Consider a craftable **O2 Drop Canister** the player can place to build a personal supply line into a deep cave (very Subnautica pipe-like, strongly pro-exploration).

### 3.3 Combat/hazard damage
- **ElectroWeaver 45 dmg** is nearly half the health bar from one ambient predator — fine as the *scariest* thing in the abyss, but it should telegraph (audio crackle + light flicker 1–2 s before strike) so avoidance is a skill.
- Add brief **invulnerability frames (~1 s)** after any hit so overlapping hazards (Siphon + ShockKelp) can't shred the player instantly.
- SandViper (10) and VoltaicLurker (15) are good "tax" damage — keep.
- **Vehicle crush**: 0.08 hull/tick over depth limit is a good slow warning; keep, but add loud creaking + HUD flash so the player understands *why* they're dying.

### 3.4 Resource economy
- Early game (Titanium/Copper) pacing is decent. The risk: mid-game turns into "grind Quartz/Nickel." Counter it with **found loot > mined loot** for variety: more FloatingCrates, wreckage caches, and one-off "mineral geodes" (single tile, big yield, visually distinct) that make wandering pay off.
- **Abyssal Ore needing the Mech** is a strong gate. Make sure at least 12–15 ore are reliably reachable per world so the 10-ore rocket cost never dead-ends a run.
- Reduce Fabricator crafting cost for food (10 power to cook a fish is steep relative to Solar's 0.08/tick) — or make cooking free at a small **Galley** module.

### 3.5 Movement & stamina
- Cave top speed 3.5 → 6.5 with Fins is a great power spike; keep.
- Sprint stamina (1.5/s drain, 1.0/s regen) barely matters. Either make sprint meaningfully faster in caves (escape tool vs. predators) or fold stamina into the Energy rework.

---

## 4. Exploration Features (the heart of the plan)

### 4.1 Make the map worth reading
- **Player-drawn map / auto-chart:** overworld fog-of-war that clears as you travel; dive sites get icons once visited. The PDA gets a Map tab. This alone multiplies the exploration payoff of every trip.
- **Signal beacons:** occasional radio pings ("distress signal, bearing NW") that mark a temporary point of interest — a crashed pod with loot, a rare creature sighting, a mini-wreck. Gives every session a destination.
- **Landmarks:** 5–10 hand-authored set pieces scattered by the generator — a giant skeleton, a sunken statue, an AetherCorp buoy graveyard. No mechanical reward needed; being *seen* is the reward (plus a lore entry).
- **Map & Navigation Extensions (Post-v1):**
  - **Zoom / Pan Controls:** 2× / 4× zoom levels centered on player with arrow keys or click-and-drag panning across the 500×500 chart.
  - **Player-Placed Map Pins:** Ability to place and color-code custom pins directly on the map screen (pairs with the Deployable Beacon tool).
  - **Cave Mini-Maps:** Per-cave localized fog-of-war tracking using the same bitset `Tracker` pattern on individual cave grids.
  - **HUD Compass Strip:** Top-of-screen orientation ribbon showing bearings to home, vehicles, lost cargo, and active pins.

### 4.2 New destination types (reuse cave infrastructure)
| Destination | Hook | Contents |
|---|---|---|
| **Sinkhole / Blue Hole** | Visible dark circle in shallow water | Vertical shaft cave, tight O2 planning, rich mid-tier ore |
| **Kelp Labyrinth** | Overgrown maze in Kelp Forest biome | Passive fauna, hidden crates, NerveMats as soft walls |
| **Thermal Fields** | The unused Thermo Caves (see §2) | Heat hazard + Thermal Gen synergy, Quartz-rich |
| **The Rift** | One per world, deepest point, visible on map from start | Endgame Abyssal Ore motherlode + final lore cache; effectively the "destination the whole game points at" |
| **Derelict Life Pods** | Other divers didn't make it | Small loot + a journal lore entry each — breadcrumb storytelling |

### 4.3 Environmental storytelling > quest markers
You already have the dual-voice lore system (AetherCorp vs Triton). Expand from 5 entries to **25–40**, tied to *places* more than actions: entering a biome first time, reaching a landmark, scanning a species. The PDA becomes a collection meta-game (see §7 replayability).

### 4.4 Tiered Wreck Progression & Gated Blueprint Acquisition

To prevent sequence breaking and ensure that **each vehicle unlocks the next tier naturally without skipping steps**, the 3 world shipwrecks are structured as dedicated progression gates:

```
[On-Foot / Skiff in Shallow Reefs]
       │  (Swim down with basic O2)
       ▼
[Ship 0: Research Tender (20–40 tiles, Est. Depth 30m)]
       │  • Guaranteed Scout Sub Kit in a random room above 40m depth
       │  • O2 Bubbles available above 40m; zero bubbles below 40m
       │  • Secondary: Ultra O2 Tank, Sonar Amp, Storage Vault MKII
       ▼
[Craft Scout Sub (Mobile O2 & Hull to Depth 60m)]
       │  (Pilot Sub down central elevator shaft into oxygen dead zone)
       ▼
[Ship 1: Submersible Transport (60–100 tiles, Est. Depth 60m)]
       │  • Guaranteed Scout Sub Depth Module MK1 in upper deck (< 40m)
       │  • Craft Depth Module MK1 (5 Titanium, 3 Nickel, 2 Quartz) -> increases Sub depth to 120m
       │  • Guaranteed Heavy Mech Kit in bottom Engineering Bay (Deck 3, ty ~ 100)
       │  • O2 Cutoff: zero bubbles below 40m depth — on-foot divers drown!
       │  • Secondary: Thermal Generator, Solar Array MKII, Surface Sonar
       ▼
[Craft Heavy Mech (High-Torque Drill & Hull to Depth 120m)]
       │  (Drill through Reinforced Blast Bulkheads; Scout Sub lacks mining drill!)
       ▼
[Ship 2: AetherCorp Flagship (120+ tiles, Est. Depth 100m+)]
       │  • Deep Vault blocked by Reinforced Blast Bulkhead (Mining Pick sparks harmlessly)
       │  • Heavy Mech Drill breaks the bulkhead
       │  • Inside Vault: Guaranteed Escape Rocket Blueprint + Triton Comm Core
       ▼
[Endgame Climax: Craft Escape Rocket OR Signal Deep Relay in The Rift]
```

#### Detailed Wreck Specifications

| Feature | Ship 0: Research Tender | Ship 1: Submersible Transport | Ship 2: AetherCorp Flagship |
|---|---|---|---|
| **World Distance Band** | 20–40 tiles from Life Pod | 60–100 tiles from Life Pod | 120+ tiles in abyssal ocean |
| **Biome Location** | Shallow Reef / Kelp edge | Mid-ocean / Thermal Barrens border | Abyssal Blue deep ocean |
| **Est. Dive Depth** | 30m | 60m (upper) / 100m+ (lower) | 100m+ |
| **O2 Bubble Distribution** | Above 40m only; zero bubbles below 40m | Above 40m only; zero bubbles below 40m | Above 40m only; zero bubbles below 40m |
| **Progression Blueprint** | **Scout Sub Kit** (Guaranteed in room < 40m) | **Depth Module MK1** (< 40m) & **Heavy Mech Kit** (Eng. Bay > 80m) | **Escape Rocket Blueprint** (Guaranteed in Vault) |
| **Key Quest/Lore Items** | Research Log #1 | Triton Engineering Manifest | **Triton Comm Core** (Vault) |
| **Secondary Blueprints** | Ultra O2 Tank, Sonar Amp, Storage Vault MKII | Thermal Gen, Solar MKII, Surface Sonar | Surface Sonar, Abyssal caches |
| **Hard Gating Mechanism** | Basic navigation & standard O2 | Zero-O2 dead zone requires sub; lower deck requires Depth Module MK1 | Reinforced Bulkhead requires Heavy Mech Drill |
| **Sequence Break Prevention** | None (Introductory wreck) | On-foot diver drowns; Scout Sub without MK1 crushed below 60m | Hand pick cannot dent blast door; only Mech Drill cuts through |

#### Mechanics to Implement

1. **Deterministic Wreck Distance Bands in World Gen:**
   - Replace uniform random scattering of `TileWreckage` with distance-banded placement relative to `FindLifepodSpawn()`:
     - Band 0: Radius 20–40 tiles, water depth shallow/mid.
     - Band 1: Radius 60–100 tiles, water depth mid/deep.
     - Band 2: Radius 120–180 tiles, water depth abyssal.
   - Update `DivePrompt` and map tooltip with distinct wreck names:
     - *"Research Tender Wreckage (Est. Depth: 30m)"*
     - *"Submersible Transport Wreckage (Est. Depth: 60m)"*
     - *"AetherCorp Flagship Wreckage (Est. Depth: 100m+)"*

2. **Per-Ship O2 Bubble Cutoff in Wreck Caves:**
   - In `GenerateDeepEntitiesFromSpec` for `CaveWreckage`:
     - If `c.ShipIndex == 0`: ShatterBulbs spawn normally across all decks (`MaxTY: 9999`).
     - If `c.ShipIndex == 1`: ShatterBulbs spawn ONLY on the top deck (`MaxTY: 28`). Below Deck 1, zero ShatterBulbs spawn.
     - If `c.ShipIndex == 2`: ShatterBulbs spawn sparsely on upper decks; completely absent in the Deep Vault.

3. **Reinforced Blast Bulkhead Entity:**
   - A 2-tile wide structural barrier block placed at the entrance of Ship 2's lowest vault (Deck 3, `ty ~ 100`).
   - Mining with hand tools (Survival Pick): deals 0 damage, emits rich metallic sparks, and triggers a HUD warning: *"Bulkhead integrity 100%. Heavy Mech Drill required!"*.
   - Drilling with Heavy Mech (`heavymech.go`): drill grinding audio, heavy debris particles, takes ~4–5 seconds of sustained drilling, then shatters, granting 4x Scrap Metal and opening the doorway.

4. **Guaranteed Blueprint Node Generation:**
   - Rewrite `GenerateWreckageResources` in `internal/game/resource/generation.go`:
     - Remove the random shuffle that could omit the Scout Sub in Ship 0.
     - Ship 0: Always place `Scout Sub Kit` in the Bridge/Deck 1 room + 2 random T1 utility blueprints.
     - Ship 1: Always place `Heavy Mech Kit` in the bottom Engineering Bay (Deck 3) + 2 industrial blueprints.
     - Ship 2: Always place `Escape Rocket` + `Triton Comm Core` inside the sealed Deep Vault behind the Reinforced Bulkhead.

---

## 5. New Creatures & Hazards

Keep the ratio roughly **1 hostile : 1 ambient : 1 puzzle-hazard** — the ocean should feel alive, not hostile.

### Ambient life (cheap, huge atmosphere value)
- **Fish schools** that flee from the player and predators — doubles as a threat radar ("why did the fish just scatter?").
- **Glow Jellies:** drifting light sources in deep caves; harmless, mesmerizing, and a natural lantern to route by.
- **Reef Grazer** (large passive): slow whale-like silhouette in the overworld; scanning it is a lore milestone. Never attacks. Sells the scale of the ocean.

### New hostiles (each with a counter-play, not just damage)
| Creature | Zone | Behavior | Counter |
|---|---|---|---|
| **Lantern Angler** | Deep organic trench | Dangles a light that mimics a Glow Jelly; lunges when approached | Scanner reveals it; teaches "verify before you approach" |
| **Husk Swarm** | Wreckage | Small bots, individually weak, aggro on flashlight | Flashlight off = stealth route; Decoy Launcher trivializes them |
| **Tidecaller** | Overworld deep water | Creates a local current pulling toward it | Skiff can outrun; on foot, must swim perpendicular (classic riptide lesson) |
| **Mimic Crate** | Near wrecks | FalseBulb pattern applied to FloatingCrates | Scanner check; keeps free-loot pickups tense |

### Hazards-as-puzzles
- **Current tunnels:** one-way flows in caves — free fast travel one way, planning problem the other.
- **Collapsing ceilings:** mined-out sections shed debris after a warning rumble; makes mining locations a choice.
- **Ink/silt clouds:** disturbed silt kills visibility temporarily instead of HP — punishes recklessness without damage.

**Explicitly avoid:** a "hunter that follows you everywhere" (Subnautica Reaper-style stalkers work there because of scale; in a 4-min day loop it just adds anxiety), and any enemy that camps the Life Pod.

---

## 6. Power-ups, Tools & Upgrades

Current upgrade list (O2 ×2, Fins, Scanner) is thin at 4 slots. Additions, ordered by exploration value:

### Tools (use Energy pool from §2)
1. **Scanner (finish it):** pulse reveals ore/creatures/lore-scannables through walls in ~8-tile radius; costs Energy. This is the single highest-value feature in this plan after save/load.
2. **Grapple Line:** short tether pull — vertical mobility in caves and an escape option. Skill expression without combat.
3. **Repair Tool:** fix vehicle hull in the field (uses Scrap Metal). Makes deep vehicle expeditions viable instead of one-way bets.
4. **Deployable Beacon:** droppable light/marker that shows on HUD compass. Breadcrumbs for deep caves — pure exploration QoL.

### Upgrades (fill the 4 slots meaningfully)
- **Depth Suit MKI/MKII:** unlocks personal diving below vehicle-crush zones for short windows — high-risk on-foot access to Abyssal areas *before* the Mech, as an alternate path.
- **Rebreather:** −30% O2 drain (multiplicative alternative to bigger tanks — a build choice).
- **Insulated Plating:** halves thermal/shock damage; the "I'm going to Thermo Caves" pick.
- **Cargo Harness:** +8 inventory at the cost of −10% speed — tradeoff, not straight upgrade.

### Consumable power-ups (found > crafted, to reward looting)
- **Stim Kelp:** brief speed burst (native alternative to fins).
- **O2 Candle:** instant +50 O2, wreck loot only.
- **Sonar Flare:** one-shot full-cave reveal — the "wow" moment consumable.

---

## 7. Replayability

1. **Seeded worlds + world summary screen.** Show the seed at world-gen; let players enter one. Cheap, and speedrunners/friends will trade seeds.
2. **Randomize the progression skeleton:** shuffle which wreck holds which blueprint tier, Rift location, biome layout. The *route* to the rocket should differ per run even though the beats are fixed.
3. **PDA completion meta:** lore % and creature-scan % per category, displayed on the win screen. "Escaped in 3 days, 41% archive complete" invites another run.
4. **Alternate Endings (Three Endings ≈ Three Playthroughs):**

   Each ending represents a distinct gameplay archetype, narrative resolution, and philosophical choice regarding the ocean world. Reaching any ending concludes the run with a custom illustrated outro vignette, unique synth music fanfare, narrative debrief transcript, and a comprehensive **Expedition Stats Dashboard** (Mission Time, Days Elapsed, Total Deaths, Deepest Depth, Archive Completion %, and an Ending Completion Trophy Badge). Upon completion, the save slot records the completion badge and permanently unlocks **New Game+**, while preserving a pre-climax checkpoint reload option for sandbox play.

   ---

   ### Ending 1: Escape (The Technician / Speedrun Route)
   *The fast, tech-focused route: obey corporate protocol, extract the anomaly, and break orbit.*
   - **Playstyle Identity:** Tech rushing, min-maxing vehicle upgrades, drilling Abyssal Ore with the Heavy Mech, and building the extraction vessel as quickly as possible.
   - **Gameplay Prerequisites & Tech Gate:**
     - Discover and fabricate the **Heavy Mech Kit**.
     - Descend into the Abyssal Void (depth > 100) and mine **10 Abyssal Ore** nodes with the Heavy Mech Drill.
     - Fabricate the **Escape Rocket** item at the Base Fabricator (10 Abyssal Ore, 10 Titanium, 5 Copper, 5 Quartz).
   - **Trigger & Presentation (Surface Launch Sequence):**
     - Crafting the rocket deploys a physical **Surface Launch Platform & Rocket** adjacent to Life Pod 5 on the overworld surface.
     - The player must swim to the surface, climb the boarding ladder, and enter the capsule (`[E] Board Escape Rocket`).
     - Inside the capsule, the player pulls the master ignition lever (`[Space] Initiate Launch Sequence`).
     - A 5-second staging countdown begins with sirens and warning klaxons, followed by screen shake, particle exhaust, and ignition SFX (`rocket_ignition.wav`, `rocket_liftoff_roar.wav`).
     - An animated cutscene shows the rocket ascending, piercing the storm clouds, breaking through the ocean planet's atmosphere, and entering orbit.
   - **Narrative Payoff & Outro Vignette:**
     - **Illustrated Vignette:** The streamlined escape rocket docking with an imposing, sterile AetherCorp deep-space freighter in orbit above the azure planet.
     - **AetherCorp Corporate Debrief Transcript:**
       > *"Telemetry verified. Specimen 'Abyssal Ore' secured in class-4 containment. Congratulations on surviving contract deployment, Diver 724. Please report to Quarantine Bay C for decontamination and mandatory memory debriefing. Note: Transportation, fuel surcharge, and life pod fabrication costs (Total: 4,812,000 Credits) have been automatically deducted from your corporate account."*
     - **Audio Theme:** Triumphant, militaristic synth/orchestral fanfare (`escape_outro.mp3`).
   - **Implementation Requirements:**
     - `internal/game/item/rocket.go`: Spawn/place surface rocket entity on craft.
     - `internal/game/entity/surface_rocket.go`: Overworld entity with boarding interaction.
     - `internal/game/scene/rocket_launch.go`: Dedicated launch sequence scene with countdown, screen shake, and ascent animation.
     - `internal/game/scene/win.go`: Integration of Escape vignette, AetherCorp transcript, and stats card.

   ---

   ### Ending 2: Signal Triton (The Deep Diver / Investigator Route)
   *The exploration-heavy route: scavenge deep hazardous sites, repair the ancient Deep Relay in The Rift, expose AetherCorp's crimes, and ally with the breakaway Triton fleet.*
   - **Playstyle Identity:** Deep expedition planning, multi-biome navigation, hazard mitigation, and uncovering the truth behind the doomed Triton expedition.
   - **Gameplay Prerequisites & Tech Gate (Multi-Biome Scavenger Hunt):**
     - Discover **The Rift** (the deepest abyssal canyon in the world, depth > 120) and locate the submerged, offline **Triton Deep Relay Station**.
     - Scavenge **3 unique, non-craftable salvaged components** hidden across the world's most dangerous locations:
       1. **Triton Quantum Comm Core:** Salvaged from a locked, high-pressure secure vault deep inside **Shipwreck 2** (requires Scout Sub navigation and cutting tool/mech).
       2. **Geothermal Power Coupling:** Retrieved from an active magma chamber inside the **Thermo Caves** (requires thermal hazard protection / Insulated Plating).
       3. **Abyssal Resonance Prism:** Mined from crystalline clusters on the floor of **The Rift** using the Heavy Mech Drill.
     - Ensure the base grid has surplus energy (or bring **2 charged Power Cells**) to jump-start the Relay's subsea transmission capacitors.
   - **Trigger & Presentation (Deep Relay Activation):**
     - The player swims/drives to the Deep Relay terminal at the floor of The Rift.
     - Interact with the console to seat all 3 salvaged components and connect power conduits.
     - Throw the master broadcast breaker (`[E] Broadcast Subspace Distress Ping`).
     - The Deep Relay's massive coils spin up, illuminating the pitch-black abyss with teal bioluminescence. It fires a massive acoustic shockwave through the trench (`sfx/sonar_ping_deep.wav`), pulsing water rings and echoing into the deep.
   - **Narrative Payoff & Outro Vignette:**
     - **Illustrated Vignette:** In the pitch black of the Rift, illuminated only by the glowing relay coils, a sleek, silent Triton stealth dropship descends through the trench, extending mechanical docking clamps and warm floodlights toward the player's submersible.
     - **Triton Collective Response Transcript:**
       > *"Relay 9-Omega online... Transmitting on secure Triton frequency. Diver, this is Flagship Hyperion of the Free Triton Coalition. We hear you. We thought everyone on the surface was lost when AetherCorp cut the orbital comms and triggered the facility scuttle to bury their research. We are sending an atmospheric stealth shuttle directly into the trench to bring you home. You don't owe AetherCorp another second of your life. Stand by for extraction."*
     - **Audio Theme:** Atmospheric, mysterious, and uplifting synth progression featuring deep-sub choral harmonies.
   - **Implementation Requirements:**
     - `internal/game/world/rift.go` & `internal/game/cave/rift_cave.go`: The Rift biome and Deep Relay structure generation.
     - `internal/game/item/quest_items.go`: Definitions for `Triton Comm Core`, `Geothermal Coupling`, and `Abyssal Prism`.
     - `internal/game/entity/deep_relay.go`: Interactive station entity handling component slots, power checks, and activation VFX.
     - `internal/game/scene/win.go`: Integration of Triton dropship vignette, coalition audio theme, and stats card.

   ---

   ### Ending 3: Stay (The Naturalist / Completionist Route)
   *The mastery route: catalog 100% of the planetary ecosystem in a single playthrough, reject corporate exploitation, and settle permanently as the ocean's solitary caretaker.*
   - **Playstyle Identity:** Methodical exploration, PDA codex completion, scanning every living organism, discovering every geological secret, and uncovering all environmental lore logs.
   - **Gameplay Prerequisites & Single-Run 100% Archive Requirement:**
     - Must achieve **100% Planetary Archive completion** in a single playthrough:
       - **100% Fauna Scans:** Every shallow, mid-depth, and abyssal species scanned (Cave Fish, Crab, SandViper, VoltaicLurker, ElectroWeaver, Glow Jellies, Reef Grazer, etc.).
       - **100% Flora Scans:** Every plant and fungal species cataloged (ShatterBulb, ShockKelp, NerveMat, etc.).
       - **100% Geology Scans:** Every mineral deposit type scanned and logged.
       - **100% Lore Logs:** All 25–40 narrative logs discovered (Life Pod journals, Triton research logs, and AetherCorp executive memos).
     - Single-run gating ensures this ending remains the ultimate achievement badge of complete mastery.
   - **Trigger & Presentation (Voluntary Terminal Broadcast):**
     - Upon reaching 100% Archive, a distinctive golden notification pings the PDA: *"Planetary Archive 100% Complete: Open Ledger Protocol Available."*
     - A new option unlocks at the Life Pod 5 / Habitat Base terminal: `[E] Archive Complete: Broadcast Public Ledger & Settle Here`.
     - An explicit confirmation dialog prevents accidental triggers (*"Are you ready to conclude your expedition and make this ocean your permanent home?"*).
     - Upon confirmation, the diver uploads the complete scientific database simultaneously to an open, uncensorable public interstellar frequency, destroying AetherCorp's proprietary monopoly before severing their corporate biometric tracking beacon.
   - **Narrative Payoff & Outro Vignette:**
     - **Illustrated Vignette:** The diver sitting peacefully on the observation deck of their habitat at dusk. In the gentle twilight above the sea, alien auroras swirl across the sky, while beneath the glass, bioluminescent fish schools and the majestic silhouette of a Reef Grazer swim gently past.
     - **The Naturalist's Personal Epilogue Transcript:**
       > *"I could have built the rocket. I could have run to Triton. But watching the light drift through the kelp canopy at dusk, I realized there is nothing out among the stars worth trading this for. AetherCorp wanted to strip-mine this world until the trenches went silent. Now, the open ledger has broadcast every genome, every reef coordinate, and every anomaly to the civilian network. They can never hide what is here, and they can never own it. My oxygen tank is full, the reef is quiet, and for the first time in my life... I am home."*
     - **Audio Theme:** Gentle, contemplative piano with ambient ocean swells, shifting into an emotional, resonant synth crescendo.
   - **Implementation Requirements:**
     - `internal/game/story/archive.go`: Complete single-run archive aggregator calculating percentage across Fauna, Flora, Geology, and Lore.
     - `internal/game/scene/menu.go`: Base station / Life Pod terminal action and confirmation prompt.
     - `internal/game/scene/win.go`: Integration of Sunset Habitat vignette, Naturalist epilogue text, and stats card.

   ---

   ### Ending Comparison & Implementation Matrix

   | Feature | Ending 1: Escape | Ending 2: Signal Triton | Ending 3: Stay |
   |---|---|---|---|
   | **Primary Motivator** | Speed, survival, progression | Deep exploration, lore investigation | 100% completionism, mastery |
   | **Primary Gate** | Heavy Mech + 10 Abyssal Ore | Scavenge 3 deep hazard sites + Rift | Single-run 100% PDA Archive |
   | **Target Playtime** | 2–3 hours | 3–5 hours | 5–8 hours |
   | **World Location** | Life Pod 5 (Surface) | The Rift floor (Deepest abyss) | Habitat Base / Life Pod Terminal |
   | **Cinematic Style** | High-energy rocket launch | Atmospheric abyssal rendezvous | Contemplative sunset twilight |
   | **Narrative Tone** | Cold corporate satire / debt | Rebellious coalition rescue | Poetic peaceful naturalist |
   | **Audio Vibe** | Triumphant industrial fanfare | Deep mysterious space synth | Warm piano & ocean ambient |
   | **Post-Game Reward** | 'AetherCorp Survivor' Badge + NG+ | 'Triton Pioneer' Badge + NG+ | 'Planetary Custodian' Badge + NG+ |
5. **Post-win New Game+:** keep lore/blueprints, new seed, slightly denser hostiles. Low effort once save/load exists.
6. **Difficulty presets** (three toggles, not a slider): *Explorer* (no cargo loss, −50% damage), *Standard*, *Expedition* (O2 drain +25%, crates rarer). Explorer mode is your "focus on exploration" promise made literal.

---

## 8. Pacing & Difficulty Curve (target)

| Phase | Time | Player state | Tension source |
|---|---|---|---|
| Tutorial | 0–10 min | On foot + Skiff, shallow reefs | Oxygen management, first SandViper |
| Early | 10–40 min | Fins + O2 tank, Solar base | First trench dives, FalseBulb lesson |
| Mid | 40–90 min | Scout Sub, Scanner, 2 wrecks found | Depth-60 wall, ElectroWeaver fear, blueprint hunting |
| Late | 90–150 min | Heavy Mech, Thermo/Rift access | Abyssal expeditions, hull management |
| End | 150+ | Rocket parts assembling | The one last deep run; choice of ending |

Checks against this curve:
- Nothing in the first 30 minutes should be able to kill a careful player in under ~4 seconds of sustained mistakes.
- Every phase should introduce **one new place** and **one new creature**, minimum.
- The day/night cycle (4 min) is short for expedition pacing — consider 8–10 min days once solar charging matters more.

---

## 9. Prioritized Roadmap

**Phase 1 — Foundation (do first)**
1. Save/load system
2. Death → cargo beacon (kill the inventory wipe)
3. Finish Scanner + wire lore triggers
4. Place Thermo Caves on the map
5. Hit invulnerability frames + ElectroWeaver telegraph


---

**Phase 2 — Exploration payoff**
6. Fog-of-war map + PDA Map tab
7. Landmarks (5–10 set pieces) + derelict life pods
8. Lore expansion to 25+ entries
9. Ambient life: fish schools, glow jellies
10. Signal beacon events

---

**Phase 3 — Systems depth**
11. New tools: Grapple, Repair Tool, Beacon
12. New upgrades: Rebreather, Depth Suit, Insulated Plating
13. 3–4 new creatures (Lantern Angler, Husk Swarm first)
14. Sinkhole + Kelp Labyrinth destinations
15. The Rift

**Phase 4 — Replayability & Alternate Endings**
16. Seed display/entry + blueprint shuffle
17. Difficulty presets
18. Polish Ending 1: Escape (Surface rocket platform + boarding launch sequence)
19. Implement Ending 2: Signal Triton (The Rift Deep Relay + 3 salvage items)
20. Implement Ending 3: Stay (Single-run 100% archive aggregator + base terminal trigger)
21. Unified Victory Scene refactor (Illustrated vignettes, distinct music, expedition stats dashboard)
22. New Game+ & ending completion badges


---

## 10. Balance Change Summary (quick reference)

| Value | Current | Proposed | Why |
|---|---|---|---|
| Death penalty | Full inventory wipe | Drop-at-site beacon | Encourage risk-taking |
| ElectroWeaver damage | 45, no telegraph | 45 with 1–2 s telegraph | Fear via avoidance skill |
| Post-hit i-frames | None | ~1 s | Prevent hazard-stack instakills |
| Food craft cost | 10 power | 0–2 power (Galley) | Food shouldn't compete with fabrication |
| Day length | 14400 ticks (~4 min) | ~2× longer | Expedition pacing, solar relevance |
| Sprint | Marginal | Faster in caves or fold into Energy | Give stamina a job |
| Player Energy | Unused | Tool power pool | Powers scanner/grapple; food synergy |
| Abyssal Ore supply | Unverified | Guarantee ≥12–15/world | Never dead-end the rocket |
