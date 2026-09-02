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

- [x] **The Skiff Surface Boat - Side View** (`assets/textures/skiff_surface.png`)
  - **Details**: 16-bit retro pixel art 2D side-view game sprite of the skiff boat floating at the surface waterline inside shallow caves.
  - **RetroDiffusion Prompt**:
    > `16-bit retro pixel art 2D side-view game sprite of the small electric ocean exploration skiff boat shown in the top-down reference image (Image 1), rendered in full side profile facing left in the exact pixel art style and color palette of the reference images. Industrial off-white hull with a sharp angled bow, prominent vibrant safety orange stripe along the gunwale, low-profile cockpit console with dark cyan tinted glass windshield, flat rear deck structure mounting a blue solar panel array, and an electric outboard motor with propeller at the stern. Clean crisp pixel art line art and shading, isolated on a solid bright chroma key green background #00FF00.`
  - **Render Logic**: [cave_draw.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_draw.go)

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

## 4. Audio Assets Checklist

Audio assets should be placed in `assets/audio/sfx/` (.wav format, 44.1kHz 16-bit) and `assets/audio/music/` (.mp3 or .ogg format, seamlessly loopable where applicable).

---

### 4.1 Player Movement & Survival SFX
- [x] **Diver Splash (Water Entry / Exit)** (`assets/audio/sfx/splash.wav`, `assets/audio/sfx/splash_exit.wav`)
  - **Sound Profile:** Satisfying hydraulic splash and bubbling displacement when leaping into/out of cave entrances or diving from the Skiff.
  - **Logic / Triggers:** [overworld_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/overworld_update.go), [cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave.go)
- [x] **Swim Stroke & Fin Kicking** (`assets/audio/sfx/swim_stroke.wav`)
  - **Sound Profile:** Soft aquatic swoosh and fluid fin kick rhythmically played during steady underwater movement.
  - **Logic / Triggers:** [player.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/player/player.go)
- [x] **Sprinting / Dash Bubbles** (`assets/audio/sfx/swim_sprint.wav`)
  - **Sound Profile:** Muffled, rapid rushing water and bubbling trail when holding Shift to boost swim speed.
  - **Logic / Triggers:** [player.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/player/player.go), [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Oxygen Depletion / Heavy Breathing Loop** (`assets/audio/sfx/heavy_breathing_loop.wav`)
  - **Sound Profile:** Muffled, claustrophobic breathing loop that escalates in tempo as oxygen drops below 30%.
  - **Logic / Triggers:** [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go), [player.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/player/player.go)
- [x] **Critical Health Heartbeat** (`assets/audio/sfx/heartbeat_loop.wav`)
  - **Sound Profile:** Heavy, muted bass heartbeat thumping in ears when player health falls below 25%.
  - **Logic / Triggers:** [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go)
- [x] **Player Hurt / Impact Stun** (`assets/audio/sfx/player_hurt.wav`)
  - **Sound Profile:** Heavy blunt impact groan and suit vibration rattle on predator bite or hazard contact.
  - **Logic / Triggers:** [player.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/player/player.go), [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Drowning / Suit Depressurization Collapse** (`assets/audio/sfx/player_drown.wav`, `assets/audio/sfx/suit_breach.wav`)
  - **Sound Profile:** Exhaling bubble burst, suit seal alarm beep, and slow audio fade-out upon zero oxygen blackout.
  - **Logic / Triggers:** [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go), [gameover.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/gameover.go)
- [x] **Oxygen Tank Refill Gasp** (`assets/audio/sfx/o2_refill.wav`)
  - **Sound Profile:** Pressurized air hiss and deep inhalation gasp upon surfacing, entering base, or breaking a Shatter-Bulb.
  - **Logic / Triggers:** [player.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/player/player.go), [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Item Pickup / Collection Pop** (`assets/audio/sfx/item_pickup.wav`)
  - **Sound Profile:** Crisp, buoyant aquatic pop and pocketing chime when collecting ore items, fish, or floating salvage.
  - **Logic / Triggers:** [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go), [inventory.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/inventory.go)

---

### 4.2 Tools, Equipment & Usables SFX
- [x] **Mining Tool Strike (Pickaxe / Hand Drill)** (`assets/audio/sfx/mining_hit.wav`)
  - **Sound Profile:** Heavy acoustic hammer strike thunk on solid rock with low-frequency punch and gritty stone fracture.
  - **Logic / Triggers:** [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Digging Tool Crunch** (`assets/audio/sfx/dig_crunch.wav`)
  - **Sound Profile:** Gritty, granular sand/silt shovel digging sound.
  - **Logic / Triggers:** [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Resource Node Shatter / Break** (`assets/audio/sfx/ore_break.wav`)
  - **Sound Profile:** Heavy stone fracturing and crystal crumble on resource node destruction.
  - **Logic / Triggers:** [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Scanner Tool - Scanning Loop** (`assets/audio/sfx/scanner_loop.wav`)
  - **Sound Profile:** Oscillating high-pitch digital telemetry sweep and targeting hum while holding right-click scan.
  - **Logic / Triggers:** [item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go), [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Scanner Tool - Scan Complete & PDA Unlock Chime** (`assets/audio/sfx/scanner_complete.wav`)
  - **Sound Profile:** Bright, melodic sci-fi chime signaling a successful scan and newly unlocked database entry.
  - **Logic / Triggers:** [story/manager.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/story/manager.go), [cave_update.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_update.go)
- [x] **Flashlight Click (Toggle On/Off)** (`assets/audio/sfx/flashlight_toggle.wav`)
  - **Sound Profile:** Heavy waterproof rubber-sealed switch click and bulb activation click.
  - **Logic / Triggers:** [item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go), [cave_usable_context.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_usable_context.go)
- [x] **Repair Tool - Welding Arc Loop** (`assets/audio/sfx/repair_tool_loop.wav`)
  - **Sound Profile:** Sizzling plasma arc buzz, electrical hiss, and molten spark popping while repairing vehicles or structures.
  - **Logic / Triggers:** [item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go), [cave_usable_context.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/cave_usable_context.go)
- [x] **Repair Tool - Repair Completed** (`assets/audio/sfx/repair_tool_complete.wav`)
  - **Sound Profile:** Positive dual-tone confirmation beep and metal latch lock-in.
  - **Logic / Triggers:** [item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go)
- [x] **Sonic Decoy - Launch & Acoustic Pulse** (`assets/audio/sfx/decoy_launch.wav`, `assets/audio/sfx/decoy_pulse.wav`)
  - **Sound Profile:** Pneumatic canister launch whoosh followed by loud rhythmic sonar decoy pulses luring away predators.
  - **Logic / Triggers:** [sonic_decoy.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/sonic_decoy.go), [item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go)
- [x] **Chemical Deterrent - Cloud Dispersion** (`assets/audio/sfx/deterrent_disperse.wav`)
  - **Sound Profile:** Pressurized aerosol hiss and bubbling chemical cloud expansion.
  - **Logic / Triggers:** [deterrent_cloud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/deterrent_cloud.go), [item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go)

---

### 4.3 Vehicles & Machinery SFX
- [x] **The Skiff - Surface Outboard Motor Loop** (`assets/audio/sfx/skiff_engine_loop.wav`)
  - **Sound Profile:** Throaty electric marine motor hum and water wake churning while navigating overworld waters.
  - **Logic / Triggers:** [skiff.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/skiff.go)
- [x] **The Skiff - Solar Array Charging Hum** (`assets/audio/sfx/skiff_solar_charge.wav`)
  - **Sound Profile:** Soft high-frequency photovoltaic inverter hum during daytime solar replenishment.
  - **Logic / Triggers:** [skiff.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/skiff.go)
- [x] **Scout Sub - Electric Propeller Loop** (`assets/audio/sfx/sub_engine_loop.wav`)
  - **Sound Profile:** Low, smooth electric submersible turbine whir with gentle propeller cavitation.
  - **Logic / Triggers:** [scoutsub.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/scoutsub.go)
- [x] **Scout Sub - Active Sonar Ping & Echo** (`assets/audio/sfx/sub_sonar_ping.wav`, `assets/audio/sfx/sub_sonar_echo.wav`)
  - **Sound Profile:** Reverberant metallic sonar ping (pressing `Q`) and fading return blips echoing off cave walls and fauna.
  - **Logic / Triggers:** [sonar.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/sonar/sonar.go), [scoutsub.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/scoutsub.go)
- [x] **Scout Sub - Hull Stress & Crush Depth Groan** (`assets/audio/sfx/sub_hull_creak.wav`)
  - **Sound Profile:** Ominous metal creaking, shearing stress groans, and popping rivets when diving beyond vehicle depth limits.
  - **Logic / Triggers:** [scoutsub.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/scoutsub.go)
- [x] **Heavy Mech - Hydraulic Step & Footstep Thud** (`assets/audio/sfx/mech_step.wav`)
  - **Sound Profile:** Massive mechanical servo whine and heavy seabed thud shaking sediment.
  - **Logic / Triggers:** [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go)
- [x] **Heavy Mech - Heavy Rotary Drill Loop** (`assets/audio/sfx/mech_drill_loop.wav`)
  - **Sound Profile:** Deep motorized gear grind and high-torque tungsten drill chewing through abyssal rock.
  - **Logic / Triggers:** [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go)
- [x] **Heavy Mech - Water-Jet Thruster Boost** (`assets/audio/sfx/mech_thruster_loop.wav`)
  - **Sound Profile:** Roaring high-pressure hydro-thruster stream providing vertical mech propulsion.
  - **Logic / Triggers:** [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go)
- [x] **Heavy Mech - Ground Slam Landing** (`assets/audio/sfx/mech_impact.wav`)
  - **Sound Profile:** Heavy metal clank and shockwave thud on touchdown.
  - **Logic / Triggers:** [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go)
- [x] **Vehicle Boarding, Docking & Egress** (`assets/audio/sfx/vehicle_enter.wav`, `assets/audio/sfx/vehicle_exit.wav`)
  - **Sound Profile:** Pneumatic canopy / hatch seal cycling whoosh when mounting or dismounting sub/mech/skiff.
  - **Logic / Triggers:** [scoutsub.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/scoutsub.go), [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go), [skiff.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/skiff.go)
- [x] **Vehicle Damage Alarm & Warning Klaxon** (`assets/audio/sfx/vehicle_alarm.wav`)
  - **Sound Profile:** Urgent dual-tone cabin warning siren when hull integrity is below 20%.
  - **Logic / Triggers:** [scoutsub.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/scoutsub.go), [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go)

---

### 4.4 Marine Fauna, Flora & Environmental Hazards SFX
- [x] **Thermocline Rammer - Charge Roar & Collision Smash** (`assets/audio/sfx/rammer_roar.wav`, `assets/audio/sfx/rammer_impact.wav`)
  - **Sound Profile:** Low guttural predator roar before charging; heavy bone-cracking armor plate smash on impact.
  - **Logic / Triggers:** [thermocline_rammer.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/thermocline_rammer.go)
- [x] **Electro-Weaver - Telegraph Crackle & High-Voltage Shock** (`assets/audio/sfx/weaver_charge.wav`, `assets/audio/sfx/weaver_shock.wav`)
  - **Sound Profile:** 1.5s rising electrical hum and sparkling ionization telegraphing attack, followed by an explosive lightning zap.
  - **Logic / Triggers:** [electro_weaver.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/electro_weaver.go)
- [x] **Voltaic Lurker - Ambush Surge & Electric Arc** (`assets/audio/sfx/lurker_ambush.wav`)
  - **Sound Profile:** Sudden cavitation rush and crackling electrical arc when bursting from ambush shadows.
  - **Logic / Triggers:** [voltaic_lurker.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/voltaic_lurker.go)
- [x] **False-Bulb Snare - Lure Bioluminescent Hum & Trap Snap** (`assets/audio/sfx/snare_lure_hum.wav`, `assets/audio/sfx/snare_snap.wav`)
  - **Sound Profile:** Hypnotic glassy chiming pulse from bulb, followed by violent fleshy jaw snap and tentacle constriction.
  - **Logic / Triggers:** [false_bulb_snare.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/false_bulb_snare.go)
- [x] **Sand Viper - Burrow Rustle & Ambush Strike Hiss** (`assets/audio/sfx/viper_burrow.wav`, `assets/audio/sfx/viper_strike.wav`)
  - **Sound Profile:** Sediment churning rustle in seabed rock followed by an aggressive aquatic venom strike hiss.
  - **Logic / Triggers:** [sand_viper.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/sand_viper.go)
- [x] **Oxygen Bubble Plant (Shatter-Bulb) - Deep Resonant Bubble Bloop & Liquid Cavity Pop** (`assets/audio/sfx/shatter_bulb_pop.wav`)
  - **Sound Profile:** Deep, hollow resonant Minnaert bubble "bloop" and liquid water glug releasing O2 into player tanks.
  - **Logic / Triggers:** [shatter_bulb.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/shatter_bulb.go)
- [x] **Shock Kelp - Ambient Buzz & Contact Zap** (`assets/audio/sfx/shock_kelp_hum.wav`, `assets/audio/sfx/shock_kelp_zap.wav`)
  - **Sound Profile:** Subtle ambient electrical hum when near fronds; sharp static shock zap on player contact.
  - **Logic / Triggers:** [shock_kelp.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/shock_kelp.go)
- [x] **Brimstone Siphon / Thermal Vent - Geothermal Eruption & Steam Hiss** (`assets/audio/sfx/thermal_vent_hiss.wav`, `assets/audio/sfx/vent_bubble_rumble.wav`)
  - **Sound Profile:** Sizzling superheated water hiss, bubbling mineral discharge, and low geothermal rumbling.
  - **Logic / Triggers:** [brimstone_siphon.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/brimstone_siphon.go), [thermal_vent.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/overworld/thermal_vent.go)
- [x] **Nerve Mat - Bio-Electric Sting** (`assets/audio/sfx/nerve_mat_sting.wav`)
  - **Sound Profile:** Wet squelching sting and painful burning sizzle when stepping on toxic mat flora.
  - **Logic / Triggers:** [nerve_mat.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/nerve_mat.go)
- [x] **Overworld Whirlpool - Rushing Vortex Roar** (`assets/audio/sfx/whirlpool_roar_loop.wav`)
  - **Sound Profile:** Roaring vortex rush, rushing whitewater, and deep suction rumble dragging down vessels.
  - **Logic / Triggers:** [whirlpool.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/overworld/whirlpool.go)
- [x] **Passive Fish & Cosmetic Schools - Fin Flutter** (`assets/audio/sfx/fish_swim_flutter.wav`)
  - **Sound Profile:** Gentle aquatic fin swish and darting school scatter sound.
  - **Logic / Triggers:** [passive_fish.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/passive_fish.go), [cosmetic_fish.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/overworld/cosmetic_fish.go)
- [x] **Passive Seabed Crab - Skittering Clicks** (`assets/audio/sfx/crab_skitter.wav`)
  - **Sound Profile:** Subtle chitinous leg tapping and sandy scurrying on seabed floor.
  - **Logic / Triggers:** [passive_crab.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/passive_crab.go)
- [x] **Floating Crate / Lost Cargo - Latch Pop & Salvage Jingle** (`assets/audio/sfx/cargo_unlatch.wav`)
  - **Sound Profile:** Rusted pressurized latch unsealing pop followed by a discovery chime.
  - **Logic / Triggers:** [floating_crate.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/overworld/floating_crate.go), [lost_cargo.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/lost_cargo.go)

---

### 4.5 Base Building, Fabricator Crafting & Energy Systems SFX
- [x] **Fabricator - 3D Printing / Molecular Synthesis Loop** (`assets/audio/sfx/fabricator_craft_loop.wav`)
  - **Sound Profile:** Rhythmic electronic printer hum, particle realignment sweeps, and laser welding buzz.
  - **Logic / Triggers:** [recipes.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/data/recipes.go), [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go)
- [x] **Fabricator - Synthesis Complete Chime** (`assets/audio/sfx/fabricator_success.wav`)
  - **Sound Profile:** Melodic three-note chime and pneumatic item tray ejection click.
  - **Logic / Triggers:** [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go)
- [x] **Base Module Construction & Anchoring** (`assets/audio/sfx/base_build.wav`)
  - **Sound Profile:** Heavy structural clamp lock, magnetic seal click, and power link chime.
  - **Logic / Triggers:** [base.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/base/base.go)
- [x] **Base Module Deconstruction** (`assets/audio/sfx/base_deconstruct.wav`)
  - **Sound Profile:** Reverse synthesis laser dissolve and structural uncoupling hiss.
  - **Logic / Triggers:** [base.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/base/base.go)
- [x] **Base Airlock Hatch - Cycle & Drain** (`assets/audio/sfx/airlock_cycle.wav`)
  - **Sound Profile:** Heavy hydraulic door slide, water drainage torrent, and chamber pressurization hiss.
  - **Logic / Triggers:** [base.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/base/base.go)
- [x] **Power Online / Generator Hum Loop** (`assets/audio/sfx/power_online.wav`, `assets/audio/sfx/generator_hum_loop.wav`)
  - **Sound Profile:** Rising power capacitor whine and clean low-frequency generator hum (Solar/Thermal).
  - **Logic / Triggers:** [base.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/base/base.go)
- [x] **Power Failure / Base Low Power Warning** (`assets/audio/sfx/base_power_down.wav`, `assets/audio/sfx/power_alarm.wav`)
  - **Sound Profile:** Winding-down power transformer groan, ambient light flicker hum, and low intermittent alarm beep.
  - **Logic / Triggers:** [base.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/base/base.go), [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go)
- [x] **Storage Vault Locker - Open & Close** (`assets/audio/sfx/storage_open.wav`, `assets/audio/sfx/storage_close.wav`)
  - **Sound Profile:** Magnetic locker latch unseal and smooth sliding storage drawer sound.
  - **Logic / Triggers:** [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go)
- [x] **Food Consumption & Medkit Injection** (`assets/audio/sfx/eat_crunch.wav`, `assets/audio/sfx/medkit_apply.wav`)
  - **Sound Profile:** Crisp chewing sound for cooked fish/crab; pneumatic hypo-spray hiss for medical healing.
  - **Logic / Triggers:** [item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go), [hud_inventory.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud_inventory.go)

---

### 4.6 User Interface, HUD, Map & PDA SFX
- [x] **UI Button Hover / Focus Tick** (`assets/audio/sfx/ui_hover.wav`)
  - **Sound Profile:** Crisp, subtle electronic click/tick when moving cursor over menu buttons or inventory slots.
  - **Logic / Triggers:** [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go), [title.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/title.go)
- [x] **UI Button Click / Confirmation** (`assets/audio/sfx/ui_confirm.wav`)
  - **Sound Profile:** Clean tactile sci-fi confirmation beep on button selection or crafting action.
  - **Logic / Triggers:** [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go), [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go)
- [x] **UI Cancel / Invalid Action Buzzer** (`assets/audio/sfx/ui_cancel.wav`, `assets/audio/sfx/ui_error.wav`)
  - **Sound Profile:** Soft back-out thud for closing menus; low muted double-buzz when missing ingredients or inventory full.
  - **Logic / Triggers:** [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go)
- [x] **Inventory & Crafting Drawer Open / Close** (`assets/audio/sfx/inventory_open.wav`, `assets/audio/sfx/inventory_close.wav`)
  - **Sound Profile:** Futuristic digital HUD boot chirp and holographic panel slide whoosh.
  - **Logic / Triggers:** [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go), [hud_inventory.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud_inventory.go)
- [x] **Hotbar Slot Selection Click** (`assets/audio/sfx/hotbar_switch.wav`)
  - **Sound Profile:** Quick tactile ratchet click when cycling hotbar items with keys 1-5 or mouse wheel.
  - **Logic / Triggers:** [hud_inventory.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud_inventory.go)
- [x] **Map Open & Waypoint Ping** (`assets/audio/sfx/map_open.wav`, `assets/audio/sfx/map_ping.wav`)
  - **Sound Profile:** Holographic chart initialization shimmer followed by a resonant sonar blip when placing custom map pins.
  - **Logic / Triggers:** [menu_map.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu_map.go)
- [x] **PDA Typewriter Scroll / Teletype Ticks** (`assets/audio/sfx/pda_typewriter_tick.wav`)
  - **Sound Profile:** Very quiet, crisp data-terminal static clicks as lore entries render letter-by-letter.
  - **Logic / Triggers:** [story/manager.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/story/manager.go), [menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/menu.go)
- [x] **PDA Milestone / Tech Unlock Fanfare** (`assets/audio/sfx/pda_unlock_fanfare.wav`)
  - **Sound Profile:** Inspiring dual-tone sci-fi fanfare upon unlocking a major blueprint tier or reaching depth milestone.
  - **Logic / Triggers:** [story/manager.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/story/manager.go)

---

### 4.7 Survival Voice Alerts (PDA Synthesized Voice Lines)
- [x] **Voice: "Oxygen Low"** (`assets/audio/sfx/voice_o2_low.wav`)
  - **Sound Profile:** Calm, filtered female AI voice: *"Oxygen low. 30 seconds remaining."*
  - **Logic / Triggers:** [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go), [player.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/player/player.go)
- [x] **Voice: "Oxygen Critical"** (`assets/audio/sfx/voice_o2_critical.wav`)
  - **Sound Profile:** Urgent female AI voice with soft alert tone: *"Warning: Oxygen critical."*
  - **Logic / Triggers:** [hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/hud.go)
- [x] **Voice: "Maximum Depth Limit Exceeded"** (`assets/audio/sfx/voice_depth_warning.wav`)
  - **Sound Profile:** Authoritative AI warning: *"Warning: Maximum depth limit exceeded. Vessel hull damage imminent."*
  - **Logic / Triggers:** [scoutsub.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/scoutsub.go), [heavymech.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/heavymech.go)
- [x] **Voice: "Power Reserves Depleted"** (`assets/audio/sfx/voice_power_low.wav`)
  - **Sound Profile:** Flat AI warning: *"Emergency: Habitat power reserves depleted."*
  - **Logic / Triggers:** [base.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/base/base.go)

---

### 4.8 Music, Biome Ambient Soundscapes & Cutscenes
- [x] **Main Title Theme** (`assets/audio/music/main_title.mp3`)
  - **Sound Profile:** Deep, ethereal synthesizer pads, slowly evolving sub-bass, and sparse crystalline melodic motifs setting an alien ocean atmosphere.
  - **Logic / Triggers:** [title.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/title.go)
- [x] **Intro Cinematic - Life Pod Atmospheric Entry & Splash** (`assets/audio/music/intro_cinematic.mp3`, `assets/audio/sfx/pod_atmospheric_entry.wav`, `assets/audio/sfx/pod_water_crash.wav`)
  - **Sound Profile:** Roaring reentry shockwaves, shuddering metal hulls, urgent alarms, violent ocean impact smash, and subsequent calming ocean surge.
  - **Logic / Triggers:** [intro.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/intro.go)
- [x] **Overworld Surface Ocean Theme** (`assets/audio/music/overworld_surface.mp3`)
  - **Sound Profile:** Breezy, bright electronic ocean ambient with gentle percussion, acoustic piano notes, and wind/wave soundscapes.
  - **Logic / Triggers:** [overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/overworld.go), [biome.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/world/biome.go)
- [x] **Shallow Coral Reef Cave Ambient** (`assets/audio/music/cave_shallow.mp3`)
  - **Sound Profile:** Warm underwater synth pads, echoing water droplets, and gentle aquatic resonance for sunlit reef caves.
  - **Logic / Triggers:** [shallow_seabed_cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/shallow_seabed_cave.go), [biome_config.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/biome_config.go)
- [x] **Kelp Forest & Shock Kelp Cave Ambient** (`assets/audio/music/cave_kelp.mp3`)
  - **Sound Profile:** Murky, swaying ambient textures with subtle bioluminescent pulses, resonant low drones, and distant aquatic fauna calls.
  - **Logic / Triggers:** [shock_kelp_cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/shock_kelp_cave.go), [biome_config.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/biome_config.go)
- [x] **Thermal Barrens & Volcanic Cave Ambient** (`assets/audio/music/cave_volcanic.mp3`)
  - **Sound Profile:** Deep tectonic sub-bass rumbles, bubbling hydrothermal vent hisses, and menacing low brass chords.
  - **Logic / Triggers:** [thermo_cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/thermo_cave.go), [biome_config.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/biome_config.go)
- [x] **Abyssal Trench & Void Ambient** (`assets/audio/music/cave_abyssal.mp3`)
  - **Sound Profile:** Deep, pitch-black minimal horror drone, sparse metallic groans, claustrophobic silence, and distant predator echoes.
  - **Logic / Triggers:** [organic_trench_cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/organic_trench_cave.go), [void_cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/void_cave.go)
- [x] **Wreckage Corridor / Derelict Ship Ambient** (`assets/audio/music/cave_wreckage.mp3`)
  - **Sound Profile:** Creaking twisted hull beams, hollow acoustic water reverberations, distant electrical short-circuit sparks, and mournful synth pads.
  - **Logic / Triggers:** [wreckage_corridor_cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave/wreckage_corridor_cave.go)
- [x] **Escape Rocket Launch & Victory Outro Theme** (`assets/audio/music/escape_outro.mp3`, `assets/audio/sfx/rocket_ignition.wav`, `assets/audio/sfx/rocket_liftoff_roar.wav`)
  - **Sound Profile:** Staging countdown chime, thunderous rocket engine ignition roar, and a triumphant, soaring synth/orchestral fanfare as the rocket breaks orbit.
  - **Logic / Triggers:** [win.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/win.go)
- [x] **Game Over / Desolation Screen Theme** (`assets/audio/music/game_over_theme.mp3`)
  - **Sound Profile:** Somber, fading ambient synth chord and slow heartbeat fading into oceanic abyss silence.
  - **Logic / Triggers:** [gameover.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/gameover.go)

