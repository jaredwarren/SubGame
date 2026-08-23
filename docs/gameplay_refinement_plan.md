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
4. **Alternate endings:**
   - **Escape** (current rocket ending) — the fast route.
   - **Signal Triton** — repair the deep relay in the Rift instead: longer, exploration-heavy route with a different final lore payoff.
   - Optional: **Stay** — complete the archive to 100% for a quiet epilogue. Three endings ≈ three playthroughs.
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

**Phase 4 — Replayability**
16. Seed display/entry + blueprint shuffle
17. Difficulty presets
18. Second ending (Signal Triton)
19. PDA completion stats on win screen
20. New Game+

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
