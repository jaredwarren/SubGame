# SubGame Required Assets Checklist

This document provides a comprehensive list of all required assets (graphics, animations, UI textures, fonts, sound effects, and music) needed to fully replace the game's debug placeholders and make the title production-ready.

---

## 1. Typography & Fonts
To replace Ebitengine's basic debug printer with high-quality, scalable interfaces, the following fonts should be placed in `assets/fonts/`:

- [ ] **Primary HUD & UI Font** (`assets/fonts/primary_hud.ttf`)
  - **Type:** TrueType/OpenType (.ttf/.otf)
  - **Style:** Clean, futuristic, high-legibility sans-serif (e.g., *Outfit*, *Roboto*, or *Orbitron*).
  - **Purpose:** Used for all HUD meters, oxygen counts, battery levels, crafting item names, base schematic lists, and tooltips.
  - **Loading Code:** [internal/game/hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/hud.go), [internal/game/menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/menu.go)

- [ ] **Title & Cinematic Font** (`assets/fonts/title.ttf`)
  - **Type:** TrueType/OpenType (.ttf/.otf)
  - **Style:** Wide, spacing-optimized, stylized sci-fi font.
  - **Purpose:** Game logo on main title screen, win/loss scenes, biome transition texts ("ENTERING ABYSSAL ZONE"), and mission prompts.
  - **Loading Code:** [internal/game/title_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/title_scene.go), [internal/game/win_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/win_scene.go), [internal/game/gameover_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/gameover_scene.go)

---

## 2. Sprites & Graphic Atlases
Sprites should be placed in `assets/textures/` and loaded as `ebiten.Image` spritesheets or individual PNGs.

### Overworld (Top-Down Surface)
- [ ] **Water Tiles Sheet** (`assets/textures/overworld_water.png`)
  - **Details:** 64x64px repeating textures for Coastal Water (light teal), Deep Water (dark navy), and Trench transition edges.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art texture, tilesheet. Six repeating 64x64px seamless tile textures of ocean water: light turquoise coastal water, medium blue water, deep navy blue water, and matching transition borders. Flat shading, classic retro RPG style, seamless looping texture on all sides. No grids, clear details.`
  - **Render Logic:** [internal/game/overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/overworld.go)

- [ ] **Land / Reef Tiles Sheet** (`assets/textures/overworld_land.png`)
  - **Details:** 64x64px sandy shore borders, green grassy reef tiles, and rocky blocks representing islands.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art tilesheet, game asset. Seamless 64x64px tiles: sandy beach coastlines, vibrant green sea grass land, and sharp volcanic reef stone blocks. Includes matching transitions from grass to sand and sand to water. Orthographic top-down perspective, retro 90s RPG style, tileable repeating texture.`

- [ ] **Base Life Pod Sprite** (`assets/textures/lifepod_surface.png`)
  - **Details:** A 128x128px detailed capsule sprite floating in water with a solar array visible on top.
  - **AI Image Generation Prompt:**
    > `2D orthographic game sprite, 16-bit pixel art style. Industrial floating base capsule pod, metallic white and grey plating with high-contrast orange stripes, round blue dome glass hatches on top, solar array panel grids mounted on the hull. Floating on a clean, solid bright green chroma-key background, isolated game asset.`
  - **Render Logic:** [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go) (Base Station rendering)

- [ ] **The Skiff Surface Boat Sprite** (`assets/textures/skiff.png`)
  - **Details:** A 56x24px top-down motorboat sprite with an orange trim and solar recharging cells on the back.
  - **AI Image Generation Prompt:**
    > `2D top-down game sprite, 16-bit retro pixel art. A small motorized exploration boat (skiff), sharp bow, flat deck with a blue solar panel on the back, industrial white hull, bright orange safety stripe trim. Centered on a flat, solid green chroma-key background.`
  - **Render Logic:** [internal/game/vehicle/vehicle.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/vehicle.go)


### Cave Environment (Side-Scroller Grid)
- [x] **Shallow Cave Tiles** (Procedurally drawn with sand speckles in code)
  - **Details:** Sandy, coral-overgrown rock textures with border tiles for slopes.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art tilesheet, side-view platformer style. Organic cave wall tiles of sandy yellow reef rock, overgrown with tiny colorful corals and seaweed. Includes inner fills, corner blocks, and slopes. Repeating tileable texture, clean grid lines.`
  - **Render Logic:** [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go)

- [ ] **Mid-Depth Cave Tiles** (`assets/textures/cave_mid.png`)
  - **Details:** Dark teal, slime-covered basalt rock textures with bioluminescent moss streaks.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art tilesheet, side-view cave walls. Basalt volcanic stone blocks in deep teal and blue, overgrown with bioluminescent neon cyan moss and tiny glowing spores. 2D platformer grid, seamless tileable textures.`

- [ ] **Deep Cave (Volcanic) Tiles** (`assets/textures/cave_deep.png`)
  - **Details:** Basalt/obsidian rock textures with glowing orange volcanic core lines.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art tilesheet, side-view cave walls. Jagged obsidian and basalt volcanic rock blocks with pulsing streams of glowing orange and yellow lava cracks. Side-scroller grid style, tileable.`

- [ ] **Abyssal Cave Tiles** (`assets/textures/cave_abyssal.png`)
  - **Details:** Pitch-black porous rock tiles highlighted with chalky white fossil deposits.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art tilesheet, side-view abyssal crevices. Pitch-black stone tiles detailed with ash-white fossil shapes, pale calcified bones, and grey organic mats. Seamless tileable textures, dark atmospheric style.`

- [ ] **Cave Flora Sprites**
  - **Details:** Swaying Kelp / Sea Grass (animated), glowing bulb plants, volcanic chimneys, and purple nerve mats.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art game asset spritesheet. A set of side-scrolling underwater environmental details: 4 frames of swaying green kelp stalks, 3 varieties of glowing blue sea mushrooms, a hydrothermal vent volcano chimney venting orange dust, and a flat purple roots carpet. Clean bright green chroma-key background.`
  - **Render Logic:** [internal/game/biome_entity.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/biome_entity.go)


### Characters & Vehicles
- [x] **Diver Spritesheet** (`assets/textures/diver_sheet.png`)
  - **Details:** Animated sheet containing:
    - 4-frame Idle float cycle (diver body bobs gently).
    - 8-frame Swim cycle (kicking fins, turning visor).
    - 4-frame Mining strike strike (swinging a pickaxe tool).
    - 1-frame Damage/Stun impact state.
  - **Style Matching Colors:** Industrial orange wetsuit, yellow oxygen cylinder tank on the back, cyan/light-blue glass visor helmet.
  - **AI Image Generation Prompt:**
    > `2D side-scrolling video game asset, sprite sheet, 16-bit retro pixel art style. A deep-sea diver character wearing an industrial orange wetsuit with a yellow oxygen cylinder tank strapped to their back and a large round cyan glass visor helmet. Side-view profile perspective. The sheet must contain a clean grid sequence of animation frames: Row 1 has 4 frames of idle floating bobbing cycle; Row 2 has 8 frames of swimming/kicking cycle; Row 3 has 4 frames of swinging a handheld pickaxe/scanner tool forward; Row 4 has 1 frame of damage recoil. Crisp pixel outlines, flat clean shading, presented on a solid bright green background for easy transparency removal. No shadows on background, no water overlays.`
  - **Render Logic:** [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go)

- [x] **Scout Sub Sprite** (`assets/textures/scout_sub.png`)
  - **Details:** A 48x32px mini-sub sprite with a glass bubble cockpit, a back propeller (2-frame rotation), and a front headlight lens.
  - **AI Image Generation Prompt:**
    > `2D side-view mini-submarine game sprite, 16-bit retro pixel art. A small exploration sub with a circular glass cockpit displaying a cyan glow, a yellow and industrial teal hull, front glass searchlight lens, and back copper propellers. Side profile view, solid green chroma-key background.`
  - **Render Logic:** [internal/game/vehicle/vehicle.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/vehicle.go)

- [x] **Heavy Mech Spritesheet** (`assets/textures/heavy_mech.png`)
  - **Details:** A walker robot containing:
    - Left/Right walker legs (4-frame walk cycle).
    - Mech torso cockpit with metallic hatch.
    - Right arm claws.
    - Left arm Drill (4-frame spinning loop).
    - Thruster flame bursts (spark animation).
  - **AI Image Generation Prompt:**
    > `2D side-scrolling video game asset, sprite sheet, 16-bit retro pixel art style. An industrial walker mech suit, dark grey and orange iron plating. The sheet must contain a clean grid sequence of animations: Row 1 has 4 frames of leg walking cycle; Row 2 has 4 frames of drill arm rotation loop; Row 3 has 1 frame of thruster ignition sparks. Solid bright green background, clean pixels, no shadows. Centered on a flat, solid green chroma-key background.`
  - **Render Logic:** [internal/game/vehicle/vehicle.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/vehicle.go)


### Biome Creatures & Flora Entities
- [ ] **Shatter-Bulb Sprite** (`assets/textures/shatter_bulb.png`)
  - **Details:** 24x24px bulb attached to a dark stem with a bright glowing blue/cyan gas bladder.
  - **AI Image Generation Prompt:**
    > `2D game sprite, 16-bit retro pixel art. A small bioluminescent cave bulb plant, a glowing cyan gas bladder bulb sitting on a dark green stem. Clean solid green background, isolated asset.`
  - **Render Logic:** [internal/game/biome_entity.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/biome_entity.go)

- [ ] **False-Bulb Snare Spritesheet** (`assets/textures/snare_sheet.png`)
  - **Details:** Mimic state (resembles Shatter-Bulb), and Aggro state (bulb splits open to reveal a circular mouth and slithering tentacles).
  - **AI Image Generation Prompt:**
    > `2D side-view creature spritesheet, 16-bit retro pixel art. The False-Bulb Snare predator: 4 frames of it hanging from a ceiling mimicking a glowing blue Shatter-bulb plant; 4 frames of it waking up, opening a circular mouth with sharp teeth, showing a red eye pupil, and extending lunging tentacles. Flat green background.`

- [ ] **Thermocline Rammer Spritesheet** (`assets/textures/rammer_sheet.png`)
  - **Details:** Armored dark orange predator fish. Includes a tail fin swim cycle and a charging pose.
  - **AI Image Generation Prompt:**
    > `2D side-scrolling creature spritesheet, 16-bit retro pixel art. The Thermocline Rammer: an armored predator fish with a dark orange body, thick shovel-like grey iron head plate, and thrashing tail fin. Sheet contains: Row 1 with 4 frames of swimming tail-fin wiggle; Row 2 with 1 frame of charging sprint. Solid green background.`

- [ ] **Electro-Weaver Spritesheet** (`assets/textures/weaver_sheet.png`)
  - **Details:** Serpentine ribbon monster. Transparent head sprite with glowing yellow nerve fibers, and segmented tail body parts that slither behind.
  - **AI Image Generation Prompt:**
    > `2D side-view serpent creature spritesheet, 16-bit retro pixel art. The Electro-Weaver: a long serpentine deep-sea ribbon monster with a transparent glowing head, glowing yellow eye dots, and multiple segment joint rings flowing behind. Includes 4 frames of slithering body waves and 2 frames of blue electric discharge sparks. Solid green background.`


### Mineable Minerals & Ore Nodes
- [ ] **Ore Node Spritesheet** (`assets/textures/ore_sheet.png`)
  - **Details:** A 256x64px horizontal spritesheet containing four 64x64px tile frames side-by-side (matching the game's `TileSize = 64` layout):
    - Frame 0 (X: 0-63): **Titanium Node** (metallic silver-grey crystals embedded in dark stone block).
    - Frame 1 (X: 64-127): **Copper Node** (reddish-orange raw metal veins running through dark stone block).
    - Frame 2 (X: 128-191): **Quartz Node** (translucent glowing cyan crystal cluster embedded in dark volcanic rock).
    - Frame 3 (X: 192-255): **Abyssal Ore Node** (glowing radioactive violet-purple crystal shards embedded in dark black deep-sea stone).
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art horizontal spritesheet, 256x64px total size, containing four 64x64px tile frames side-by-side. Solid pure green background (#00FF00) for chroma-keying. From left to right: 1. Titanium node (raw metallic silver-grey crystals embedded in dark stone block), 2. Copper node (reddish-orange raw metal veins branching through dark stone block), 3. Quartz node (translucent glowing cyan crystal cluster in dark rock block), 4. Abyssal ore node (glowing radioactive violet-purple crystal shards embedded in dark black deep-sea rock block). Clear pixel-perfect edges, game asset sheet style. Centered on a flat, solid green chroma-key background.`
  - **Render Logic:** [internal/game/resource/resource.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/resource/resource.go)


### Inventory Items & Upgrade Icons
- [ ] **Inventory Icons Sheet** (`assets/textures/item_icons.png`)
  - **Details:** 48x48px grid cells containing icons for all craftable items.
  - **AI Image Generation Prompt:**
    > `16-bit retro pixel art icon set, 48x48px square tiles. Grid of inventory icons including raw metal chunks, gas tanks, propulsion swim fins, handheld scanners, sub capsules, walker kits, solar panel modules, power cells, and escape rocket nose cones. High legibility, dark dark grey panel backing for each icon, pixel-perfect clean lines.`
  - **Render Logic:** [internal/game/item/item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go)


### UI & Overlay Textures
- [ ] **Title Screen Graphics** (`assets/textures/ui_title.png`)
  - **Details:** Futuristic logo card with stylized spacing and deep-sea silhouette banner.
  - **AI Image Generation Prompt:**
    > `Futuristic game title logo card, 16-bit pixel art style, sci-fi lettering for 'SUBGAME', deep-sea backdrop of a submarine shadow silhouette against a navy blue ocean gradient, retro game UI banner. Transparent background.`

- [ ] **UI Panel Borders** (`assets/textures/ui_panels.png`)
  - **Details:** Corner slice templates for HUD boxes, inventory grids, and base management tabs.
  - **AI Image Generation Prompt:**
    > `Sleek high-tech UI panel frames and corner buttons, 16-bit pixel art style, glowing cyan borders, semi-transparent dark slate center backings, futuristic dashboard menu widgets. Set of buttons, tabs, and window frames. Flat green background.`

- [ ] **Scanner UI Overlay** (`assets/textures/scanner_hud.png`)
  - **Details:** Digital targeting reticles and scanner completion meters.
  - **AI Image Generation Prompt:**
    > `2D sci-fi targeting HUD overlay, clean pixel lines, neon green reticle scope rings, scanning telemetry details, retro radar compass layout. Isolated on black transparent background.`

- [ ] **Vehicle Cockpit Overlays**
  - **Details:** Sub cockpit window frame (semi-circular grid) and Mech walker interior panel view.
  - **AI Image Generation Prompt:**
    > `2D spaceship cockpit border overlay, 16-bit pixel art style, circular metal window frame, small glowing dash dials, radar screen screens, industrial safety orange gauges. Center window area is hollow and transparent. Side-view game asset.`

---

## 3. Atmospheric Sound Effects (SFX)
Audio assets should be placed in `assets/audio/sfx/` in `.wav` or `.ogg` format.

### Diver Interactions
- [ ] **Diver Splash** (`assets/audio/sfx/splash.wav`) - Played when transition from Overworld to Cave occurs.
- [ ] **Swim stroke bubbles** (`assets/audio/sfx/swim.wav`) - Muffled water swoosh when moving.
- [ ] **Mining strike hit** (`assets/audio/sfx/mining_hit.wav`) - Hard metallic pickaxe clink.
- [ ] **Resource shattered** (`assets/audio/sfx/ore_break.wav`) - Low crunching rock collapse sound.
- [ ] **Shatter-bulb pop** (`assets/audio/sfx/bulb_pop.wav`) - A hollow, wet squishy burst.
- [ ] **Scanner Scanning loop** (`assets/audio/sfx/scanner_scan.wav`) - High pitch digital sweeping noise.
- [ ] **Scanner Scan complete** (`assets/audio/sfx/scanner_done.wav`) - Two quick futuristic beeps.
- [ ] **Oxygen refill hiss** (`assets/audio/sfx/o2_refill.wav`) - Quick pressurized air gasp when entering base/surface.

### Vehicles & Systems
- [ ] **Skiff Engine loop** (`assets/audio/sfx/skiff_engine.wav`) - Low rumbling motorboat sound.
- [ ] **Scout Sub Engine loop** (`assets/audio/sfx/sub_engine.wav`) - High-tech electric humming tone.
- [ ] **Mech Walker Step** (`assets/audio/sfx/mech_step.wav`) - Heavy mechanical metal thud.
- [ ] **Mech Thruster fire** (`assets/audio/sfx/mech_thruster.wav`) - Pressurized fire hiss.
- [ ] **Mech Drill arm loop** (`assets/audio/sfx/mech_drill.wav`) - Grinding gear/rotary drilling sound.
- [ ] **Sub Sonar Ping** (`assets/audio/sfx/sonar_ping.wav`) - Classic long echoing sonar ping.
- [ ] **Vehicle Collision damage** (`assets/audio/sfx/hull_scrape.wav`) - Grating steel crash and screeches.

### Hazards & Creature Alerts
- [ ] **Siphon lava jet** (`assets/audio/sfx/steam_jet.wav`) - Sudden hot water hiss.
- [ ] **False-Bulb Snare charge** (`assets/audio/sfx/snare_lunge.wav`) - Sudden screeching wet sound.
- [ ] **Rammer dash** (`assets/audio/sfx/rammer_dash.wav`) - Deep water rush/groan.
- [ ] **Electro-Weaver sparks** (`assets/audio/sfx/electricity.wav`) - Crackling high-voltage sparks.
- [ ] **Low O2 Heartbeat loop** (`assets/audio/sfx/heartbeat.wav`) - Low base thuds, speeding up on critical health.
- [ ] **Low O2 gasp loop** (`assets/audio/sfx/gasping.wav`) - Panicked gasping breaths.
- [ ] **Warning voice alerts** (`assets/audio/sfx/voice_o2_low.wav`, `assets/audio/sfx/voice_crush_depth.wav`, `assets/audio/sfx/voice_warning.wav`) - Synthetic robot voices.

---

## 4. Music & Ambient Soundscapes
Music tracks should be placed in `assets/audio/music/` in `.mp3` or `.ogg` format.

- [ ] **Main Title Theme** (`assets/audio/music/main_title.mp3`)
  - **Style:** Deep, sweeping synth chords; lonely sci-fi melody.
  - **Purpose:** Title scene loop.
- [ ] **Sailing Overworld Theme** (`assets/audio/music/overworld.mp3`)
  - **Style:** Calm, uplifting, breezy electronic ocean beats.
- [ ] **Shallow Cave Ambient** (`assets/audio/music/cave_shallow.mp3`)
  - **Style:** Muffled, distant echoes, low synth pads.
- [ ] **Volcanic smoker Trenches Ambient** (`assets/audio/music/cave_volcanic.mp3`)
  - **Style:** Heavy base drones, industrial hums, low rhythmic tension.
- [ ] **Abyssal Brine Falls Ambient** (`assets/audio/music/cave_abyssal.mp3`)
  - **Style:** Spooky, minimal, quiet horror chords with long pauses.
