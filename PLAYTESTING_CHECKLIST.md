# SubGame: Master Play-Testing & QA Verification Checklist

This comprehensive play-testing guide provides structured, end-to-end verification procedures for **SubGame** (a 2D sci-fi underwater survival exploration game built with Ebitengine and Go).

Use this document to systematically test gameplay loops, player progression, vehicle physics, procedural cave biomes, predator AI behaviors, modular base crafting, DSP audio cues, and save/load persistence.

---

## 1. Play-Testing Overview & Test Personas

### 1.1 Test Personas & Play Styles
When executing a play-test session, adopt one of the following testing mindsets:

| Persona | Focus / Objectives | Key Questions to Evaluate |
|---|---|---|
| **The First-Time Diver** | Natural onboarding, intuitive controls, tutorial clarity, readability of UI & HUD bars. | Do I know where to go? Are oxygen warnings alarming enough? Is the Base Terminal self-explanatory? |
| **The Deep-Sea Explorer** | Biome discovery, map charting, finding trenches, scanning flora/fauna, reading PDA lore. | Does clearing fog feel rewarding? Do landmarks stand out? Is exploring dark caves tense yet fair? |
| **The Power Scavenger / Speedrunner** | Crafting rush, min-maxing vehicle kits, drilling Abyssal Ore, building the Escape Rocket ASAP. | Are recipe costs balanced? Does tech progression gate depth properly? Can the rocket be completed reliably? |
| **The Chaos / Edge-Case Breaker** | Rapid clicking, running out of O2 in vehicles, dying repeatedly, testing collision boundaries. | Does the game crash? Does lost cargo persist? Are there inventory dupes or clipping bugs? |

---

## 2. Test Session Log Template

| Field | Details |
|---|---|
| **Tester Name / ID** | |
| **Date & Time** | |
| **Commit / Build Hash** | |
| **OS & Screen Resolution** | |
| **Session Length (Minutes)** | |
| **Test Focus / Persona** | |
| **Total Deaths / Respawns** | |
| **Max Depth Reached (Tiles)** | |
| **Final Game State** | [ ] Won (Rocket Launched) &nbsp; [ ] In Progress &nbsp; [ ] Blocked / Bug |

---

## 3. Environment Setup & Quick QA Cheatsheet

### 3.1 Launching the Game
```bash
# Verify procedural DSP audio assets are built
make audio

# Run the test suite before play-testing
make test

# Launch the game
make run
# OR
go run ./cmd/game
```

### 3.2 Save File Management
- **Clean Slate Play-Test**: Delete or move `save_1.json`, `save_2.json`, and `save_3.json` in the root folder before starting (legacy `save.json` is migrated into slot 1 on first launch).
- **Save File Inspection**: Open `save_1.json` (or the active slot file) to verify serialization of player stats, inventory items, installed base modules, vehicles, and explored map chunks.

### 3.3 Developer / QA Shortcut Keys
These hotkeys can be used during testing to accelerate state verification:

| Key Binding | Function | Expected Outcome |
|---|---|---|
| `Ctrl` + `1` | Debug Spawn Scout Sub | Spawns a Scout Sub directly at player location and mounts it. |
| `Ctrl` + `2` | Debug Spawn Heavy Mech | Spawns a Heavy Mech directly at player location and mounts it. |
| `Ctrl` + `3` | Debug Spawn Skiff | Spawns The Skiff directly at player location (Overworld). |
| `Ctrl` + `4` | Fill Materials Inventory | Adds +10 Titanium, Copper, Quartz, Nickel, and Abyssal Ore to player inventory. |
| `Ctrl` + `5` | Full Stat Restore | Restores Health, Oxygen, and Stamina to 100% maximum capacity. |
| `Ctrl` + `M` | Reveal All Map POIs | Uncovers every trench, wreck, thermal vent, and shock kelp cave on the ocean chart. |
| `Y` | Toggle Lighting Mask | Enables / disables dynamic light cone & darkness shader (useful to inspect full cave layout). |
| `U` | Toggle Water Shader | Enables / disables full-screen water displacement & ripple shader. |
| `C` | Enter Debug Cave | Instantly transports the player into a test cave coordinate (50, 50). |
| `G` | Trigger Game Over | Immediately triggers death / Game Over screen to test death loop and cargo drop. |

---

## 4. Master Play-Testing Checklist

### Phase 1: Title Screen, Onboarding & Settings
- [ ] **Title Screen UI**: Verify a single `START` button. With no saves, Start begins a new game in slot 1. With existing saves, Start opens the 3-slot picker (load occupied, start in empty, delete with confirm).
- [ ] **Title Music**: Ambient electronic synth track plays cleanly without audio popping or distortion.
- [ ] **Intro Cinematic**: Starting a new game displays introductory sequence / pod splashdown.
- [ ] **Tutorial & Training Quest Log**:
  - [ ] Opening new game initializes the `TRAINING` questline in the PDA (`[J]`).
  - [ ] No obtrusive persistent top-screen banner; interactive HUD pop-up notifications appear upon task progress or completion.
  - [ ] Opening the PDA displays the collapsible `[▼] TRAINING` section with step-by-step checklist tasks and progress bars.
- [ ] **Pause Overlay (`Esc`)**:
  - [ ] Opens responsive dark overlay pausing world simulation.
  - [ ] `RESUME`: Returns cleanly to previous state without physics glitches.
  - [ ] `SAVE GAME`: Displays yellow/green `GAME SAVED` notification and writes the active slot file (`save_1.json`, `save_2.json`, or `save_3.json`).
  - [ ] `RETURN TO TITLE`: Returns cleanly to the Title Screen.

---

### Phase 2: Core Diver Controls & Movement Physics
- [ ] **Swimming (`W, A, S, D` / Arrow Keys)**:
  - [ ] Overworld top-down movement feels responsive with gentle surface friction.
  - [ ] Cave side-view swimming exhibits realistic buoyancy and slight inertial coasting.
- [ ] **Sprinting (`Shift`)**:
  - [ ] Increases swim speed significantly.
  - [ ] Steadily drains Stamina bar.
  - [ ] Releasing `Shift` regenerates Stamina smoothly.
- [ ] **Diver Animation & Visor Direction**:
  - [ ] Diver sprite rotates / faces movement and aiming cursor direction.
  - [ ] Bubble particles emit behind fins during sprint-swimming.
- [ ] **Hotbar & Tool Swapping (`1 – 5`)**:
  - [ ] Keys `1` through `5` select active slots with clear HUD highlight.
  - [ ] Selecting the Flashlight slot automatically powers on the light.
- [ ] **Flashlight Toggle (`T`)**:
  - [ ] Toggles light on/off with audio click `sfx/flashlight_toggle.wav`.
  - [ ] Illuminates a directional cone following mouse cursor in dark caves.
- [ ] **Inventory Screen (`Tab` or `I`)**:
  - [ ] Opens inventory grid with sound `sfx/inventory_open.wav`.
  - [ ] Mouse clicks are consumed by inventory without triggering tool attacks in world.
  - [ ] Pressing `Tab`, `I`, or `Esc` closes inventory cleanly with `sfx/inventory_close.wav`.

---

### Phase 3: Surface Overworld & Navigation
- [ ] **Life Pod Anchor Interaction (`E`)**:
  - [ ] Approaching within 100 units of the Life Pod displays `[E] Enter Base Terminal` prompt.
  - [ ] Pressing `E` plays airlock sound `sfx/airlock_cycle.wav` and opens Base Terminal.
- [ ] **Day/Night Cycle & Solar Charging**:
  - [ ] Surface lighting shifts smoothly through dawn, noon, dusk, and night (4-minute full cycle).
  - [ ] Solar arrays on Base/Skiff recharge power during daylight hours and pause at night.
- [ ] **Surface Biomes & Landmarks**:
  - [ ] Safe Shallows, Coral Reefs, Thermal Barrens, and Open Oceans render distinctive color palettes.
  - [ ] Sandbanks and Islands block surface passage correctly.
  - [ ] Surface whirlpools create localized pull and caution warnings.
- [ ] **Ocean Chart & Fog of War (`M` or `J` -> Tab 5)**:
  - [ ] Swimming over the surface clears the fog of war in real-time.
  - [ ] Explored percentage stat updates accurately in the legend panel.
  - [ ] Life Pod icon, player position, and visited trench icons display in correct coordinates.
  - [ ] `Ctrl` + `M` reveals all POIs with correct color-coded map markers.
- [ ] **Death & Lost Cargo Recovery Loop**:
  - [ ] Drowning or losing all health triggers `sfx/player_drown.wav` and Game Over screen.
  - [ ] Player respawns at Life Pod with equipped gear preserved.
  - [ ] Dropped inventory creates a **Lost Cargo Beacon** at death coordinates.
  - [ ] Lost Cargo beacon emits an acoustic locator ping and displays a pulsating map icon.
  - [ ] Approaching the beacon in overworld/cave allows recovering all dropped loot.

---

### Phase 4: Subterranean Cave Exploration & 2D Diving
- [ ] **Trench Entry & Exit Transitions**:
  - [ ] Hovering over a Trench tile (`TileTrench`, `TileShockKelpCave`, `TileThermoCave`, `TileWreckage`) prompts `[E] Dive into Trench`.
  - [ ] Pressing `E` plays dive transition sound and smoothly switches to 2D side-view cave scene.
  - [ ] Swimming to the top surface threshold in cave exits back to the Overworld at the exact trench coordinates.
- [ ] **Lighting & Atmospheric Darkness**:
  - [ ] Ambient light fades with depth, plunging deep caves into pitch blackness.
  - [ ] Flashlight beam pierces the dark in a cone toward cursor.
  - [ ] Bioluminescent flora/fauna emit localized glowing halos.
- [ ] **Cave Biome Generation Verification**:
  - [ ] **Shallow Seabed Cave**: Cellular automata layout, lush coral pockets, passive fish/crabs, abundant Titanium & Copper veins.
  - [ ] **Organic Trench Cave**: Deep vertical Drunkard's Walk tunnels, glowing Shatter-Bulbs, False-Bulb Snare ambush predators.
  - [ ] **Thermo Cave / Silicate Smoker Trenches**: Heat distortion shaders, rhythmic steaming Brimstone Siphons, Nickel & Abyssal Ore deposits, Thermocline Rammer predators.
  - [ ] **Shock Kelp Cave**: Electric blue lighting, pulsing Shock Kelp coils, Voltaic Lurker predators.
  - [ ] **Wreckage Corridor Cave**: Metallic hull corridors, breakable vents, Scrap Metal, Electronic Waste, decryptable data terminals.
  - [ ] **The Void / Abyssal Zone**: Crushing dark, slowdown Pallid Nerve-Mats, high depth pressure, Electro-Weaver apex predator.

---

### Phase 5: Resource Harvesting, Tools & Fauna
- [ ] **Mining Resource Nodes (Left Click)**:
  - [ ] Left-clicking within range strikes node, emitting mineral impact sounds `sfx/mining_hit.wav`.
  - [ ] Node durability decreases; breaking node yields item with crumble audio `sfx/mining_break.wav`.
  - [ ] Item drops into player inventory or floats in water if inventory is full.
  - [ ] Mined node types verified:
    - [ ] Titanium Chunk
    - [ ] Copper Vein
    - [ ] Quartz Crystal
    - [ ] Nickel Deposit
    - [ ] Abyssal Ore Block (requires Heavy Mech Drill Arm)
    - [ ] Scrap Metal (yields Titanium)
    - [ ] Electronic Waste (yields Copper)
- [ ] **Catching Passive Fauna**:
  - [ ] Left-clicking on swimming Cave Fish captures `Raw Fish`.
  - [ ] Left-clicking on crawling Cave Crab captures `Raw Crab`.
- [ ] **Scanner Tool Mechanics**:
  - [ ] Equipping Scanner and holding Right Click on flora, fauna, or wreckage displays 2-second scan progress circle.
  - [ ] Completing scan unlocks corresponding PDA Database entry with audio fanfare `sfx/pda_unlock_fanfare.wav`.
- [ ] **Tactical Deployables**:
  - [ ] **Sonic Decoy**: Using item launches decoy toward cursor; emits acoustic ring waves that attract nearby predators.
  - [ ] **Chemical Deterrent**: Using item disperses purple deterrent cloud that blinds/halts chasing predators.
- [ ] **Repair Tool**:
  - [ ] Holding Left Click on damaged vehicle or base module restores hull/integrity points with welding spark FX.
- [ ] **Food Consumption**:
  - [ ] Consuming `Cooked Fish` restores Health (+25) and Stamina (+15).
  - [ ] Consuming `Cooked Crab` restores Health (+20) and Stamina (+20).

---

### Phase 6: Predator AI & Hazard Counter-Play
- [ ] **False-Bulb Snare (Ambush Predator)**:
  - [ ] Hangs from ceiling disguised as a glowing Shatter-Bulb.
  - [ ] **Light Aversion**: Freezes motionless when flashlight beam is pointed directly at it.
  - [ ] **Strike**: Lunges aggressively when player turns flashlight away into the dark.
  - [ ] Counter-test: Backing away while keeping flashlight trained on it prevents attack.
- [ ] **Thermocline Rammer (Pursuit Predator)**:
  - [ ] Eyeless; ignores flashlight.
  - [ ] Aggros on thruster vibration, fast sprinting, and Shatter-Bulb popping sound waves.
  - [ ] Charges in straight horizontal/vertical lines.
  - [ ] Counter-test: Sidestepping charge causes Rammer to crash into cave wall and become stunned for 3 seconds.
- [ ] **Voltaic Lurker & Shock Kelp (Electric Hazards)**:
  - [ ] Shock Kelp emits periodic electrical arcs (tax damage on touch).
  - [ ] Voltaic Lurker discharges electrical bursts when player approaches too closely.
- [ ] **Electro-Weaver (Abyssal Apex Stalker)**:
  - [ ] Stalks silently when active electronics (Flashlight / Sonar) are running.
  - [ ] Causes screen UI glitching and flashlight flickering before striking.
  - [ ] Deals massive damage (45 HP) on strike.
  - [ ] Counter-test: Turning off flashlight and cutting engine allows player to drift safely past in stealth.
- [ ] **Environmental Hazards**:
  - [ ] **Brimstone Siphons**: Vent scalding water in timed intervals; burns player if touched during vent pulse.
  - [ ] **Pallid Nerve-Mats**: Cling to diver suit and reduce swim speed by 50%.
  - [ ] **Shatter-Bulbs**: Popping restores +20 O2 but emits sound wave that alerts predators within 150 tiles.

---

### Phase 7: Vehicles — Skiff, Scout Sub & Heavy Mech
- [ ] **The Skiff (Surface Cargo Vessel)**:
  - [ ] Crafted from `Skiff Kit` (10 Titanium) and deployed on Overworld water.
  - [ ] Boarded / exited using `E` key.
  - [ ] High movement speed across surface waters.
  - [ ] Massive storage hold for long-distance scavenging runs.
  - [ ] Solar charging module recharges battery in sunlight.
- [ ] **The Scout Sub (Agile Cave Explorer)**:
  - [ ] Crafted from `Scout Sub Kit` (6 Titanium, 4 Copper, 2 Quartz) and deployed in caves.
  - [ ] High 2D swimming agility; fits through tight 2-tile wide cave bottlenecks.
  - [ ] **Sonar Ping (`Q`)**: Emits full-screen sonar pulse that illuminates walls and reveals hidden predators.
  - [ ] **Crush Depth Warning**: Exceeding depth limit triggers audible creaking and rapid hull integrity damage.
  - [ ] Battery drains during movement/sonar; rechargeable via `Power Cell`.
- [ ] **The Heavy Mech (Deep Abyssal Walker)**:
  - [ ] Crafted from `Heavy Mech Kit` (8 Titanium, 6 Copper, 4 Quartz).
  - [ ] Ignores buoyancy; drops straight to the ocean floor.
  - [ ] Seafloor walking physics with dust puff particle effects on landing.
  - [ ] **Thruster Jump (`Space`)**: Launches mech upward over obstacles and chasms.
  - [ ] **Drill Arm (Left Click)**: Heavy drill punch required to shatter `Abyssal Ore Blocks`.
  - [ ] Heavy armor provides immunity to minor creature bites.
- [ ] **Vehicle Inventory & Power Management**:
  - [ ] Pressing `Tab` inside a vehicle opens dual Player/Vehicle cargo transfer UI.
  - [ ] Inserting `Power Cell` recharges vehicle battery to 100%.

---

### Phase 8: Base Management Terminal & PDA Quest Log
- [ ] **Tab 0: Base Overview & Upgrades**:
  - [ ] Displays base schematic and installed module slots.
  - [ ] Installing `Solar Array Module` (+0.08 power/tick) and `Solar Array MKII` (+0.20 power/tick).
  - [ ] Installing `Storage Vault Module` (+24 slots) and `Storage Vault MKII` (+48 slots).
  - [ ] Uninstalling modules returns item to player inventory.
  - [ ] **Safety Check**: Uninstalling storage with full items triggers `Vault has too many items to uninstall!` warning and blocks removal.
- [ ] **Tab 1: Fabricator Crafting**:
  - [ ] Scrolling mouse wheel navigates crafting list smoothly.
  - [ ] Recipe ingredients and unlock status display accurately.
  - [ ] Crafting costs 10 base power; fails with audio error `sfx/ui_error.wav` if power is insufficient.
  - [ ] Successful craft adds item and plays `sfx/fabricator_success.wav`.
  - [ ] Key recipes verified:
    - [ ] High Capacity O2 Tank (+60s capacity)
    - [ ] Ultra High Capacity O2 Tank (+140s capacity)
    - [ ] Propulsion Fins (Increases cave swim speed from 3.5 to 4.3)
    - [ ] Scanner Tool, Flashlight, Repair Tool
    - [ ] Sonic Decoy, Chemical Deterrent
    - [ ] Power Cell, Thermal Generator
    - [ ] Cooked Fish, Cooked Crab
    - [ ] Escape Rocket (Victory Item)
- [ ] **Tab 2: Storage Vault**:
  - [ ] Available only when Storage Vault module is installed.
  - [ ] Left-clicking inventory items transfers them to base vault; left-clicking vault items transfers back.
- [ ] **Tab 3: Medical Bay**:
  - [ ] Available only when Medical module is installed.
  - [ ] Consumes 15 base power to heal +40 player HP.
- [ ] **Tab 4: Quests & Objectives**:
  - [ ] Access directly via Base Terminal or detached PDA shortcut `[J]`.
  - [ ] **Category Headers (Collapsible Accordion)**:
    - [ ] `[▼] / [►] TRAINING`: Survival onboarding tutorial tasks.
    - [ ] `[▼] / [►] SURVIVAL & UPGRADES`: Midgame crafting, exploration, and vehicle milestones.
    - [ ] `[▼] / [►] PROJECT ESCAPE`: Rocket construction and endgame objectives.
  - [ ] Clicking a category header toggles collapsed/expanded state without resetting scroll position.
  - [ ] Clicking a quest in the left panel selects it with illuminated cyan border and displays detailed multi-step checklist on the right.
  - [ ] Multi-step tasks display status icons (`[✓]` / `[ ]`) and numeric progress meters (`(x/y)`).
  - [ ] Completed quests display green badges `[COMPLETED]` and dim slightly.
- [ ] **Tab 5: PDA Lore Database**:
  - [ ] Lists all unlocked narrative entries by Category (Fauna, Flora, Geology, Wreckage).
  - [ ] Scrolling and selecting entries renders formatted dual logs (AetherCorp Telemetry vs Triton Journal).
- [ ] **Tab 6: Ocean Chart / Map**:
  - [ ] Displays live ocean chart with fog-of-war, POI markers, legend, and coordinates.
  - [ ] Accessible anywhere via `[M]`.

---

### Phase 9: Procedural DSP Audio System Verification
- [ ] **Pure Go Synthesis Verification**:
  - [ ] No external audio dependencies or audio dropouts during intensive scenes.
- [ ] **Biome Soundscapes**:
  - [ ] Overworld: Gentle melodic ocean surface synth.
  - [ ] Shallow Caves: Submerged low-frequency hums with droplet echoes.
  - [ ] Volcanic Trenches: Deep thermal rumbles and sizzling steam.
  - [ ] Abyssal Zone: Dark, eerie, dissonant low drones.
- [ ] **Acoustic Low-Pass Filtering**:
  - [ ] Submerging underwater applies a low-pass filter to sound effects.
- [ ] **Survival Audio Alerts**:
  - [ ] Oxygen < 30%: Heavy breathing audio loop `sfx/heavy_breathing_loop.wav` starts.
  - [ ] Oxygen < 30%: Robotic voice alert *"Oxygen low"* (`sfx/voice_o2_low.wav`) plays once.
  - [ ] Oxygen < 10%: Robotic voice alert *"Oxygen critical"* (`sfx/voice_o2_critical.wav`) plays.
  - [ ] Health < 25%: Throbbing heartbeat loop `sfx/heartbeat_loop.wav` plays.
  - [ ] Refilling oxygen stops breathing loop and resets voice alert triggers.
- [ ] **Interactive SFX Pitch Variation**:
  - [ ] Repetitive pickaxe hits and fin kicks vary in pitch by ±5–10% to prevent audio fatigue.

---

### Phase 10: Endgame, Victory & Run Statistics
- [ ] **Escape Rocket Construction**:
  - [ ] Gather required materials: 10 Abyssal Ore, 10 Titanium, 5 Copper, 5 Quartz.
  - [ ] Craft `Escape Rocket` at the Base Fabricator.
  - [ ] Crafting triggers victory fanfare `sfx/pda_unlock_fanfare.wav` and transitions to **Game Won** scene.
- [ ] **Win Scene Presentation**:
  - [ ] Rocket launch sequence displays cleanly.
  - [ ] Comprehensive run statistics are summarized:
    - [ ] Total Play Time
    - [ ] Total Resources Mined (Titanium, Copper, Quartz, Nickel, Abyssal Ore)
    - [ ] Total Items Crafted
    - [ ] Max Depth Explored
    - [ ] Marine Life Bio-Scanned (%)
    - [ ] Total Deaths / Respawns
  - [ ] Option to return to Title Screen works cleanly.

---

## 5. Bug Report & Issue Logging Template

When identifying an issue or balance concern during testing, log it using the format below:

```markdown
### [BUG / BALANCE / UX] Brief Title of Issue

- **Severity**: [ Critical (Crash/Blocker) | High (Broken Mechanic) | Medium (Balance/Visual) | Low (Polish/Typo) ]
- **Game State / Scene**: [ Title | Overworld | Cave | Base Menu | Pause | Game Over | Win ]
- **Location / Coords**: (e.g. Overworld (142, 88) or Shallow Seabed Cave Depth 65)
- **Vehicle / Equipment**: (e.g. In Scout Sub with Ultra O2 Tank)

#### Steps to Reproduce:
1. ...
2. ...
3. ...

#### Expected Result:
What should happen according to game design.

#### Actual Result:
What actually happened (e.g., player clipped through wall, sound failed to loop, item duplicated).

#### Logs / Screenshots:
Paste console output, error trace, or description of visual glitches.
```

---

## 6. Qualitative Play-Tester Feedback Questionnaire

After completing a play-test session, score and summarize your impressions:

1. **Immersion & Atmosphere (1–10)**: Did the darkness, lighting shaders, and DSP ambient audio create a tense deep-sea feel?
2. **Exploration & Curiosity (1–10)**: Did the ocean chart, fog of war, and trench discoveries make you eager to dive deeper?
3. **Control Feel & Physics (1–10)**: How did swimming, sub agility, and mech walking feel? Was there any sluggishness or frustration?
4. **Survival Tension Balance (1–10)**: Was oxygen drain fair? Did Shatter-Bulbs provide good emergency routes?
5. **Predator Encounter Quality (1–10)**: Did predators feel distinct with clear counter-play (e.g. freezing the Snare with light, dodging the Rammer)?
6. **Most Frustrating Moment**:
7. **Most Rewarding / Memorable Moment**:
8. **Top 3 Recommended Fixes or Improvements**:
