# SubGame Required Assets & RetroDiffusion Generation Guide

This document provides a comprehensive checklist of all required graphics, animations, tilesets, UI textures, fonts, and audio assets needed for SubGame, along with **RetroDiffusion setup instructions, prompt templates, and style reference anchors** for generating pixel-art assets with AI.

---

## 1. RetroDiffusion Workflow & Reference Image Guide

### Using RetroDiffusion for SubGame
[RetroDiffusion](https://www.retrodiffusion.com/) (available as a standalone generator or Aseprite plugin) generates high-quality retro pixel art. To ensure consistent visual style, color palettes, and proportions across all game assets—and to make adding new animation states (e.g. digging, dying) effortless—follow this workflow:

1. **Upload Style & Sprite Reference Anchors**:
   Sprite reference images saved at `assets/retrodiffusion_refs/`:
   - [diver_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/diver_ref.png) — **Character Anchor (Side-View)** (Scuba diver in industrial orange wetsuit, cyan visor, yellow oxygen tank). Upload when generating side-view diver animations (swimming, digging, dying, mining).
   - [diver_topdown_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/diver_topdown_ref.png) — **Character Anchor (Top-Down)** (Scuba diver seen from above with yellow oxygen tank, cyan helmet visor, and black flippers). Upload when generating overworld top-down swimming frames.
   - [scout_sub_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/scout_sub_ref.png) — **Vehicle Anchor** (Mini-sub & mechs with cyan dome cockpits, yellow/teal hulls).
   - [cave_tileset_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/cave_tileset_ref.png) — **Environment Anchor** (Sandy reef, basalt cave walls, bioluminescent moss, volcanic stone).
   - [sea_creatures_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/sea_creatures_ref.png) — **Fauna & Entity Anchor** (Deep-sea alien predators, ribbon monsters, shatter-bulbs, crabs).

2. **256x256 Pixel Concept Art Anchors**:
   Exact 256x256px environment concept art references saved at `assets/retrodiffusion_refs/`:
   - [concept_overworld_256.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/concept_overworld_256.png) — **Overworld Surface Scene (256x256)** (Top-down ocean surface, lifepod base capsule, skiff motorboat, tropical reef island).
   - [concept_cave_shallow_256.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/concept_cave_shallow_256.png) — **Shallow Cave Biome Scene (256x256)** (Sandy reef rock walls, colorful coral reef ledges, cyan shatter-bulb plants, titanium ore nodes).
   - [concept_cave_volcanic_256.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/concept_cave_volcanic_256.png) — **Volcanic Cave Biome Scene (256x256)** (Jagged black basalt stone, glowing orange lava cracks, hydrothermal vents, copper ore, heavy walker mech).
   - [concept_cave_abyssal_256.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/concept_cave_abyssal_256.png) — **Abyssal Trench Scene (256x256)** (Pitch-black porous rock, calcified fossil bones, glowing cyan bioluminescent moss, electro-weaver monster).

3. **RetroDiffusion Settings Configuration**:
   - **Model Preset**: `16-bit SNES / Genesis` or `General Pixel Art v2`
   - **Chroma-Key Background**: Set background to `#00FF00` (Pure Green) or use the **Transparent Background** feature for easy auto-keying in `internal/assets`.
   - **Palette Lock**: Select **16-color** or **32-color** palette restriction. Key hex codes for SubGame:
     - Suit Orange: `#FF6600`
     - Visor Cyan: `#00E5FF`
     - Oxygen Tank Yellow: `#FFCC00`
     - Base Metal Slate: `#333333` / `#555555`
   - **Reference Weight (Image-to-Image)**: Set reference strength to **0.60 – 0.75** when generating new animation states for existing characters/vehicles.

3. **Adding New Animation States (e.g. Digging, Dying)**:
   - Instead of regenerating the full sprite sheet, generate single frame sequences (e.g., 4 frames for digging) using [diver_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/diver_ref.png) as reference.
   - Use the prompt templates below for exact pose and motion control.

---

## 2. Typography & Fonts

Place TrueType/OpenType font files in `assets/fonts/`:

- [ ] **Primary HUD & UI Font** (`assets/fonts/primary_hud.ttf`)
  - **Style:** Clean, futuristic, high-legibility sans-serif (*Outfit*, *Roboto*, or *Orbitron*).
  - **Purpose:** Used for all HUD meters, oxygen counts, battery levels, crafting item names, base schematic lists, and tooltips.
  - **Loading Code:** [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go), [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go)

- [ ] **Title & Cinematic Font** (`assets/fonts/title.ttf`)
  - **Style:** Wide, spacing-optimized, stylized sci-fi font.
  - **Purpose:** Game logo on main title screen, win/loss scenes, biome transition texts ("ENTERING ABYSSAL ZONE").
  - **Loading Code:** [title.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/title.go), [win.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/win.go), [gameover.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/gameover.go)

---

## 3. Sprites & Graphic Atlases

Place sprites in `assets/textures/` loaded as `ebiten.Image` spritesheets or individual frame PNGs.

### Diver Character (Side-Scroller Cave Player)
- [x] **Diver Spritesheet** (`assets/textures/diver_sheet.png`)
  - **Reference Image**: [diver_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/diver_ref.png)
  - **Animation States**:
    - **Idle Float**: 4-frame bobbing cycle (gently floating in water).
    - **Swim Cycle**: 8-frame fin kicking cycle.
    - **Mining Strike**: 4-frame swing cycle with pickaxe scanner.
    - **[NEW] Digging State**: 4-frame digging animation (striking downward into sand/rock with shovel/drill tool).
    - **[NEW] Dying / Knockout State**: 5-frame depressurization/collapse sequence (diver floats upward limp with suit seal failing).
    - **Damage / Stun**: 1-frame recoil impact pose.
  - **RetroDiffusion Prompts**:
    - **Idle Float**:
      > `16-bit retro pixel art sprite sheet, 4 frames horizontal. Deep-sea scuba diver in industrial orange wetsuit with yellow oxygen tank and cyan glass visor floating gently in water, side-view profile. Solid green chroma-key background #00FF00.`
    - **Swim Cycle**:
      > `16-bit retro pixel art sprite sheet, 8 frames horizontal. Deep-sea scuba diver swimming forward kicking flippers, industrial orange suit, cyan glass visor helmet. Side-view profile, clean pixel outlines. Solid green chroma-key background #00FF00.`
    - **Digging State (New)**:
      > `16-bit retro pixel art sprite sheet, 4 frames horizontal sequence. Deep-sea scuba diver in industrial orange suit swinging a hand drill tool downwards into the seabed rock, kicking up small sediment dust. Side-view profile, solid green chroma-key background #00FF00.`
    - **Dying State (New)**:
      > `16-bit retro pixel art sprite sheet, 5 frames horizontal sequence. Deep-sea scuba diver collapsing, losing oxygen pressure, floating limp and drifting upwards with cracked cyan visor helmet. Side-view profile, solid green chroma-key background #00FF00.`
  - **Render Logic**: [cave_draw.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_draw.go), [cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave.go)


### Vehicles & Submersibles
- [x] **Scout Sub Sprite** (`assets/textures/scout_sub.png`)
  - **Reference Image**: [scout_sub_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/scout_sub_ref.png)
  - **Details**: Mini-sub with glowing cyan bubble cockpit, yellow hull with dark teal trim, front headlight, copper rear propeller (2-frame rotation).
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art side-view game sprite. Small exploration mini-submarine with a glowing cyan glass dome cockpit, yellow industrial hull plating, dark teal frame accents, front searchlight, copper propeller at back. Side profile, solid green background #00FF00.`
  - **Render Logic**: [scoutsub.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/scoutsub.go)

- [x] **Heavy Mech Spritesheet** (`assets/textures/heavy_mech.png`)
  - **Reference Image**: [scout_sub_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/scout_sub_ref.png) (Use for color match)
  - **Details**: Walker mech with 4-frame leg walking cycle, 4-frame rotary drill arm loop, thruster flame spark frame.
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art sprite sheet. Industrial deep-sea walker mech suit, dark grey and orange iron armor plating. Row 1: 4 frames leg walking cycle. Row 2: 4 frames spinning drill arm. Side-view profile, solid green background #00FF00.`
  - **Render Logic**: [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go)

- [x] **The Skiff Surface Boat** (`assets/textures/skiff.png`)
  - **Details**: Top-down motorboat sprite with orange trim and rear solar charging grid.
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art 2D top-down game sprite. Small motorized ocean exploration skiff boat, sharp bow, solar panel on flat deck, industrial white hull with safety orange trim. Centered on solid green background #00FF00.`
  - **Render Logic**: [skiff.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/skiff.go)

- [x] **Base Life Pod** (`assets/textures/lifepod_surface.png`)
  - **Details**: 128x128px floating capsule base with solar array mounted on top.
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art 2D orthographic game sprite. Industrial floating base capsule pod, metallic white hull with orange stripes, solar array panel grid on top dome. Solid green background #00FF00.`
  - **Render Logic**: [game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go)


### Cave Environment & Biome Tilesets
- **Reference Image**: [cave_tileset_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/cave_tileset_ref.png)

- [ ] **Shallow Cave Tiles** (`assets/textures/cave_shallow.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art tilesheet, side-scroller cave walls. Organic cave tiles of sandy yellow reef rock overgrown with tiny coral spores and sea grass. Includes solid fill blocks, corner caps, and slopes. Tileable seamless texture, 64x64px grid.`

- [ ] **Mid-Depth Cave Tiles** (`assets/textures/cave_mid.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art tilesheet, side-view cave walls. Basalt volcanic stone blocks in deep teal and blue, overgrown with bioluminescent neon cyan moss and glowing spores. Tileable seamless texture, 64x64px grid.`

- [ ] **Deep Cave (Volcanic) Tiles** (`assets/textures/cave_deep.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art tilesheet, side-view cave walls. Jagged obsidian volcanic rock blocks with glowing streams of pulsing orange lava cracks. Tileable seamless texture, 64x64px grid.`

- [ ] **Abyssal Cave Tiles** (`assets/textures/cave_abyssal.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art tilesheet, side-view abyssal crevice walls. Pitch-black stone tiles detailed with ash-white fossil shapes and calcified bone fragments. Tileable seamless texture, 64x64px grid.`


### Biome Creatures, Flora & Entities
- **Reference Image**: [sea_creatures_ref.png](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/retrodiffusion_refs/sea_creatures_ref.png)

- [ ] **Thermocline Rammer Spritesheet** (`assets/textures/rammer_sheet.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art creature sprite sheet, 4 frames horizontal. Armored dark orange predator fish with a heavy shovel-shaped grey iron head plate and swimming tail fin wiggle. Side-view, solid green background #00FF00.`
  - **Render Logic**: [thermocline_rammer.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/thermocline_rammer.go)

- [ ] **Electro-Weaver Spritesheet** (`assets/textures/weaver_sheet.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art creature sprite sheet, 4 frames horizontal. Serpentine ribbon monster with a translucent glowing yellow head, yellow eye dots, and slithering segmented tail. Side-view, solid green background #00FF00.`
  - **Render Logic**: [electro_weaver.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/electro_weaver.go)

- [ ] **False-Bulb Snare Spritesheet** (`assets/textures/snare_sheet.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art creature sprite sheet. 4 frames of glowing blue bulb plant mimicry; 4 frames of opening circular mouth with sharp teeth and extending tentacles. Side-view, solid green background #00FF00.`
  - **Render Logic**: [false_bulb_snare.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/false_bulb_snare.go)

- [ ] **Sand Viper Spritesheet** (`assets/textures/sand_viper.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art creature sprite sheet, 6 frames slithering movement. Sandy-gold scale snake predator with glowing yellow eyes. Side-view, solid green background #00FF00.`
  - **Render Logic**: [sand_viper.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/sand_viper.go)

- [ ] **Shatter-Bulb Plant** (`assets/textures/shatter_bulb.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art game sprite. Glowing cyan gas bulb plant sitting on a dark green stem. Clean solid green background #00FF00.`
  - **Render Logic**: [shatter_bulb.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/shatter_bulb.go)

- [ ] **Passive Crab & Passive Fish** (`assets/textures/passive_crab.png`, `assets/textures/passive_fish.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art creature sprite. Small bright red seabed crab with tiny black eye stalks and walking legs. Side-view, solid green background #00FF00.`


### Mineable Ore Nodes & Item Icons
- [x] **Ore Node Sheet** (`assets/textures/ore_sheet.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art horizontal sheet of 4 ore blocks, 64x64px each. Solid green background #00FF00. 1. Titanium (silver crystals in stone), 2. Copper (red veins in stone), 3. Quartz (glowing cyan crystals), 4. Abyssal ore (radioactive purple shards in black rock).`
  - **Render Logic**: [sprites.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/resource/sprites.go)

- [x] **Inventory Icons Sheet** (`assets/textures/item_icons.png`)
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art inventory icon grid set, 48x48px square tiles on dark grey panel backings. Icons for metal ores, oxygen tanks, flippers, scanner tool, solar panels, storage vaults, sub modules, batteries, thermal generator, cooked fish, cooked crab. Crisp pixel art.`
  - **Render Logic**: [sprites.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/sprites.go)

---

## 4. Audio Assets Checklist

Audio assets should be placed in `assets/audio/sfx/` (.wav) and `assets/audio/music/` (.mp3/.ogg).

### Sound Effects (SFX)
- [ ] **Diver Splash** (`assets/audio/sfx/splash.wav`) — Water entry transition.
- [ ] **Swim stroke bubbles** (`assets/audio/sfx/swim.wav`) — Swooshing water.
- [ ] **Mining strike hit** (`assets/audio/sfx/mining_hit.wav`) — Metallic clink.
- [ ] **Digging crunch** (`assets/audio/sfx/dig_crunch.wav`) — Soft sand/rock digging crunch.
- [ ] **Resource shattered** (`assets/audio/sfx/ore_break.wav`) — Rock shatter sound.
- [ ] **Scanner scanning loop** (`assets/audio/sfx/scanner_scan.wav`) — High pitch digital sweep.
- [ ] **Oxygen refill hiss** (`assets/audio/sfx/o2_refill.wav`) — Air gasp.
- [ ] **Scout Sub Engine loop** (`assets/audio/sfx/sub_engine.wav`) — Sub electric hum.
- [ ] **Mech Walker Step & Drill** (`assets/audio/sfx/mech_step.wav`, `assets/audio/sfx/mech_drill.wav`) — Mechanical thuds & gear grinding.

### Music & Ambient Soundscapes
- [ ] **Main Title Theme** (`assets/audio/music/main_title.mp3`) — Deep synth chords.
- [ ] **Overworld Ocean Theme** (`assets/audio/music/overworld.mp3`) — Breezy electronic ocean beats.
- [ ] **Shallow Cave Ambient** (`assets/audio/music/cave_shallow.mp3`) — Low synth pads and distant water echoes.
- [ ] **Abyssal Zone Ambient** (`assets/audio/music/cave_abyssal.mp3`) — Spooky, minimal horror drones.
