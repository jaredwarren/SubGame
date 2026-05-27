# SubGame Polish & Production-Readiness Checklist

This document details the checklist of items required to elevate **SubGame** from a prototype with placeholder vector graphics to a polished, atmospheric, production-ready 2D deep-sea survival game. 

It focuses purely on **gameplay, aesthetics, UI/UX, audio, and gameplay mechanics**, leaving out distribution and delivery.

---

## 1. Visual Polish & Aesthetics
The game currently relies on primitive vector shapes (`vector.FillCircle`, `vector.FillRect`) for character, vehicle, creature, and environment rendering. A polished game needs actual artwork, animations, and fluid particle/shader effects.

- [ ] **Sprite Art & Tile Texturing**
  - **Goal:** Replace flat color blocks and circles with actual sprite textures.
  - **Details:**
    - Use stylized 2D pixel-art or hand-drawn textures for water, islands, and cave blocks.
    - Create distinctive sprites for each resource node (Titanium chunk, Copper vein, Quartz crystal, glowing Abyssal Ore) instead of solid colored circles.
    - Add environmental decorations: swaying sea kelp, background bubble vents, stalactites/stalagmites, and coral reefs.
  - **Files involved:** 
    - [internal/game/overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/overworld.go) (tile rendering)
    - [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go) (cave block rendering)
    - [internal/game/resource/resource.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/resource/resource.go) (ore rendering)

- [ ] **Character & Creature Animations**
  - **Goal:** Render animations for swimming, walking, and creature behaviors instead of static shapes.
  - **Details:**
    - **Diver:** Implement a multi-frame swimming spritesheet. Rotate the sprites smoothly based on `player.Facing` (using `ebiten.DrawImageOptions.GeoM.Rotate`). Add a sprint-kick animation.
    - **Scout Sub:** Add propeller rotations and thruster ignition flashes.
    - **Heavy Mech:** Add walking leg movement cycle, thruster nozzle fires, and a drilling arm punch animation.
    - **Creatures:** Add body-bend tail wiggles for the *Thermocline Rammer* and *Electro-Weaver* to convey natural serpentine movement.
  - **Files involved:**
    - [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go) (player drawing)
    - [internal/game/vehicle/vehicle.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/vehicle.go) (vehicle drawing)
    - [internal/game/biome_entity.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/biome_entity.go) (creature updating & drawing)

- [ ] **Ambient & Underwater Effects**
  - **Goal:** Establish a deep-sea atmosphere using water filters and micro-particles.
  - **Details:**
    - **Floating Plankton/Detritus:** Spawn passive, slow-drifting dust particles ("marine snow") in the background layer of the caves.
    - **Screen Water Overlay:** Implement a full-screen water displacement shader (using a Kage uniform offset by game ticks) to create a gentle aquatic shimmer.
    - **Heat Distortion:** Apply a localized heat wave distortion shader around active volcanic chimneys (*Brimstone Siphons*).
  - **Files involved:**
    - [internal/game/particles.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/particles.go)
    - [internal/game/shader.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/shader.go) (Kage shaders)
    - [internal/game/biome_entity.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/biome_entity.go) (Brimstone Siphon rendering)

- [ ] **Flashlight & Bioluminescent Lighting Improvements**
  - **Goal:** Improve contrast and immersion in dark zones.
  - **Details:**
    - Add a soft bloom or gradient falloff to the edges of the flashlight cone (currently a sharp angle cut).
    - Give bioluminescent plants/spores a pulsing glow effect by modifying their draw radius using a cosine wave over time.
    - Introduce visual light beam decay: the beam should become narrower and dimmer as depth increases due to ocean turbidity, requiring O2/utility light upgrades.
  - **Files involved:**
    - [internal/game/shader.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/shader.go) (LightConeShaderCode)
    - [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go) (LightShader uniform setup and bioluminescent drawings)

---

## 2. UI/UX Polish
A modern, polished game needs high-fidelity user interfaces. The current HUD and menus are rendered using monochrome borders, solid boxes, and unscalable system fonts via the basic debug printer.

- [ ] **TrueType/OpenType Font Integration**
  - **Goal:** Replace `ebitenutil.DebugPrintAt` with a custom pixel-art font or modern sans-serif typography.
  - **Details:**
    - Load a `.ttf` or `.otf` font (e.g., from Google Fonts like Outfit or Orbitron) using Ebitengine's `text/v2` package.
    - Implement drop shadows, glowing text outlines for warning messages, and consistent typography hierarchy.
  - **Files involved:**
    - [internal/game/hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/hud.go)
    - [internal/game/menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/menu.go)
    - [internal/game/title_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/title_scene.go)

- [ ] **Aesthetic HUD & Alarms**
  - **Goal:** Redesign stat bars to feel integrated into the diver's futuristic helmet HUD.
  - **Details:**
    - Style bars with metallic borders, slanted angles, and subtle glass reflection layers.
    - **O2 Alarm:** When Oxygen drops below 25%, display a blinking red vignette warning mask on the screen edges and flicker the flashlight.
    - **Low HP Warning:** Pulsate the screen vignette with a heavy red border during low health states.
    - Add a depth progression gauge on the side of the screen showing the diver's current vertical depth compared to maximum crush depth limits.
  - **Files involved:**
    - [internal/game/hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/hud.go)

- [ ] **Premium Menu Interfaces & Interactive HUD**
  - **Goal:** Redesign the Base Anchor Terminal and Player Inventory tabs with fluid animations and detail.
  - **Details:**
    - **Glassmorphic panels:** Implement blur shaders or semi-transparent glowing frames for overlay menus.
    - **Crafting Icons:** Replace colored circle icons with custom weapon, vehicle kit, module, and item sprites.
    - **Hover Tooltips:** Display rich, contextual popup boxes showing item descriptions, stats (e.g., "Increases max O2 to 160s"), and sell/craft ratios.
    - **Grid Slide-In:** Animate the inventory opening and closing with slide-in or fade transitions.
  - **Files involved:**
    - [internal/game/menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/menu.go)
    - [internal/game/hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/hud.go) (DrawInventory / DrawVehicleInventory)

- [ ] **World-Space Interaction Prompts**
  - **Goal:** Make "Pilot" and "Interaction" indicators feel like part of the world rather than drawing static black panels.
  - **Details:**
    - Draw floating buttons (e.g., an animated `[E]` key icon) directly above vehicles, trenches, and life pods that bob gently up and down.
  - **Files involved:**
    - [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go) (Draw overrides for prompts)
    - [internal/game/overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/overworld.go)
    - [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go)

---

## 3. Gameplay Mechanics & Feature Gaps
While the core movement and collection loops are built, there are major feature gaps that limit player progression and interaction.

- [ ] **Scanner Tool Mechanics**
  - **Goal:** Implement scanning mechanics to explore lore, gather information, and map environments.
  - **Details:**
    - **Action:** Allow players to equip the `Scanner Tool` (press a hotkey, e.g., `3`) inside caves. Right-clicking and holding on an entity (creature, plant, unique geography) triggers a 2-second scan progress circle.
    - **Mechanic:** Scanning reveals structural details:
      - Scan walls to detect hidden ore veins behind breakable rock.
      - Scan plants (*Shatter-bulbs*, *Nerve-mats*) to log their oxygen yield/hazard behavior.
      - Scan predators (*Thermocline Rammer*, *Electro-Weaver*) to log their detection trigger and weaknesses.
  - **Files involved:**
    - [internal/game/player.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/player.go)
    - [internal/game/item/item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go)
    - [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go) (Input detection for scanning)

- [ ] **PDA Log & Lore Database Menu**
  - **Goal:** Add a digital logbook containing lore entries unlocked by the Scanner.
  - **Details:**
    - Add a new "5. PDA DATABASE" tab inside the Base Terminal menu (or allow opening it anywhere with the `J` key).
    - Unlocking entries displays stylized text cards detailing flora, fauna, and geological phenomena with custom sketches.
    - Add a mission log tracking steps toward building the *Escape Rocket*.
  - **Files involved:**
    - [internal/game/menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/menu.go)
    - [internal/game/item/item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go)

- [ ] **Environmental Hazards & Interactive Physics**
  - **Goal:** Make swimming through caves feel more dynamic and hazardous.
  - **Details:**
    - **Currents:** Add water currents (represented by drifting bubbles or blue vectors) in narrow tunnels that push the player, requiring propulsion fins to navigate.
    - **Breakable Geodes:** Add breakable crystal chunks that yield rare quartz or copper when struck.
    - **Oxygen Geysers:** Introduce geothermal vents that release oxygen bubbles instead of steam.
  - **Files involved:**
    - [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go)
    - [internal/game/biome_entity.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/biome_entity.go)

- [ ] **Advanced Vehicle Systems (Vehicle Juice)**
  - **Goal:** Elevate vehicle piloting mechanics.
  - **Details:**
    - **Cockpit HUD Overlay:** When piloting a sub or mech, draw a futuristic cockpit border overlay around the screen with dial indicators, battery percentages, and warning lights.
    - **Sonar Static Effect:** Emitting a sonar ping (*Scout Sub*) should display a digital scanline swipe and screen ripple distortion.
    - **Mech Drilling Feedback:** Drilling an ore block should trigger heavy screen shake, grinding spark particles, and slow the walker's movement.
  - **Files involved:**
    - [internal/game/vehicle/vehicle.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/vehicle.go)
    - [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go)

---

## 4. Atmospheric Sound & Audio System
Audio is essential to capture the isolated, tense feeling of deep-sea exploration. There is currently no audio implementation.

- [ ] **Ambient Soundscapes**
  - **Goal:** Play background soundtracks adjusted dynamically by player state and depth.
  - **Details:**
    - **Overworld:** A calm, melodic electronic synth theme representing the ocean surface.
    - **Shallow Caves:** Low underwater hums with echoing water droplet sound effects.
    - **Volcanic Trenches:** Low rumbling base pads with sizzling vent sounds.
    - **Abyssal Zone:** Eerie, quiet, dark ambient tracks to build dread.
  - **Technical:** Use `github.com/hajimehoshi/ebiten/v2/audio` to loop MP3/OGG files.
  - **Files involved:**
    - [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go) (audio manager setup)
    - [internal/game/scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene.go) (OnEnter triggers)

- [ ] **Interactive Sound Effects (SFX)**
  - **Goal:** Add audio feedback for all actions and collisions.
  - **Details:**
    - **Sailing/Swimming:** Gentle splash sounds when diving; muffled bubbling sounds when sprinting.
    - **Mining:** Metal pick hitting rock (high pitch clink) and rock crumbling (low pitch crunch) on node destruction.
    - **Pop:** A hollow popping sound when harvesting *Shatter-bulbs*.
    - **Pings:** A ringing metallic sonar ping when pressing `Q` in the sub.
    - **Collisions:** A heavy mechanical metal thud when the *Heavy Mech* lands hard or collides with walls.
  - **Files involved:**
    - [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go)
    - [internal/game/overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/overworld.go)
    - [internal/game/vehicle/vehicle.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/vehicle/vehicle.go)

- [ ] **Survival Audio Alerts**
  - **Goal:** Use audio to convey urgency when stats drop.
  - **Details:**
    - **Heavy Breathing:** Muffled heavy breathing sound loop that speeds up as current O2 falls below 30%.
    - **Heartbeat:** Low heartbeat thump that increases in pace when player health is critical.
    - **System Warning Voice:** Futuristic computer voice lines: *"Oxygen low"* or *"Warning: maximum depth limit exceeded"*.
  - **Files involved:**
    - [internal/game/hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/hud.go) (or player stats update logic)

---

## 5. Onboarding & Systems Polish
Providing a smooth transition between scenes, a solid save/load format, and tutorial guidance is key to making the game feel production-grade.

- [ ] **Tutorial & Controls Screen**
  - **Goal:** Teach players controls and mechanics during gameplay.
  - **Details:**
    - On starting a new game, display a sleek controls guide card.
    - Add context-sensitive tutorial popups during the first 10 minutes (e.g., *"Press Shift to swim faster; watch your Stamina."* or *"Scan the ocean surface to find trenches to dive into."*).
  - **Files involved:**
    - [internal/game/title_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/title_scene.go)
    - [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go)

- [ ] **Intro/Outro Cinematic Sequences**
  - **Goal:** Frame the game's start and end conditions with animations.
  - **Details:**
    - **Intro:** Instead of immediately spawning at the Life Pod, animate the pod falling from the sky and splashing into the ocean in a side-scrolling or cinematic cutscene.
    - **Outro:** When the *Escape Rocket* is crafted, play a sequence showing the rocket engines firing, launching upward, breaking through the clouds, and leaving orbit.
  - **Files involved:**
    - [internal/game/title_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/title_scene.go)
    - [internal/game/win_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/win_scene.go)

- [ ] **JSON Save & Load System**
  - **Goal:** Persist game states, positions, bases, and vehicles across sessions.
  - **Details:**
    - Serialize the player stats, player inventory, active vehicles, customized base upgrades, base storage, and explored coordinates into a local `save.json` file.
    - Add a "CONTINUE" option on the main Title Scene if a save file exists.
  - **Files involved:**
    - [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go)
    - [internal/game/title_scene.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/title_scene.go)
