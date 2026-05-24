# **2D Subnautica-Inspired Game: AI-Parseable Implementation Plan**

This document outlines a structured, step-by-step roadmap for implementing the 2D Subnautica-inspired game in **Go** using the **Ebitengine** library. It is designed to be easily read, parsed, and executed by an AI coding assistant.

---

## **Project Architecture & Layout**

To maintain a clean and scalable codebase, we will organize the project using standard Go layout practices:

```text
/SubGame
├── cmd/
│   └── game/
│       └── main.go                 # Entry point: Ebitengine setup and main execution loop
├── internal/
│   ├── game/
│   │   ├── state.go                # State Machine (Overworld, Cave, BaseMenu, GameOver)
│   │   ├── game.go                 # Implement ebiten.Game interface
│   │   ├── player.go               # Player structs (health, oxygen, stamina, inventory, physics)
│   │   ├── overworld.go            # Top-down overworld logic, rendering, sailing physics
│   │   ├── cave.go                 # Side-scrolling cave logic, tiles, swimming physics
│   │   ├── camera.go               # Camera control and smooth tracking
│   │   └── config.go               # Constants and configuration settings
│   ├── world/
│   │   ├── generator.go            # Orchestrates surface and cave map generation
│   │   ├── noise.go                # Perlin/Simplex noise implementation for Overworld
│   │   ├── cellular.go             # Cellular Automata algorithm for shallow caves
│   │   └── drunkard.go             # Drunkard's Walk algorithm for deep abyssal crevasse caves
│   ├── ui/
│   │   ├── hud.go                  # HUD rendering (oxygen, stamina, health bars)
│   │   ├── menu.go                 # Base building / Modular Management Menu (Fabricator, Storage, Medical)
│   │   └── font.go                 # Font loading and text rendering helpers
│   ├── vehicle/
│   │   ├── vehicle.go              # General vehicle interface and inventory
│   │   ├── skiff.go                # Overworld surface skiff implementation
│   │   ├── scout_sub.go            # Cave mini-sub implementation
│   │   └── heavy_mech.go           # Cave heavy mech suit implementation
│   ├── entity/
│   │   ├── flora.go                # Interactive flora (Shatter-bulbs, Brimstone siphons, etc.)
│   │   ├── predator.go             # Predator AI & behaviors (False-bulb snare, Thermocline rammer, Weaver)
│   │   └── resource.go             # Mineable resource node entities (Titanium, Copper, Quartz, Abyssal Ore)
│   └── shader/
│       └── light_cone.kage         # Kage fragment shader for flashlight line-of-sight cone
├── assets/
│   └── (Sprites, fonts, sounds, etc. will go here)
├── go.mod
└── game_implementation_plan.md     # This file
```

---

## **Phase 1: Project Setup & Core State Machine**
**Goal:** Initialize the project, setup dependencies, configure the window, and establish a functional game state machine that allows transitioning between states.

### **Tasks**
- [x] **1.1. Go Module Initialization**
  - Run `go mod init github.com/jaredwarren/SubGame`
  - Get Ebitengine dependency: `go get github.com/hajimehoshi/ebiten/v2`
- [x] **1.2. Basic Entry Point (`cmd/game/main.go`)**
  - Initialize the main function to configure `ebiten.RunGame` with a window width/height of 1280x720 (resizable).
  - Implement a basic game structure that runs a 60 FPS update/draw cycle.
- [x] **1.3. Define Game States (`internal/game/state.go`)**
  - Define an enum for game states: `StateOverworld`, `StateCave`, `StateBaseMenu`, `StateGameOver`.
  - Create a state coordinator interface or struct to manage transitioning and sharing player data across states.
- [x] **1.4. State Transition Triggers**
  - Implement transition key bindings (e.g., press `O` to switch to Overworld, `C` to Cave, `M` to Base Menu) to manually verify the state machine.
  - Draw temporary text placeholders on the screen representing each state.

### **Verification**
- Run `go run ./cmd/game/main.go`.
- Ensure the window starts, shows 60 FPS, and pressing key binds switches the screen text between "Overworld State", "Cave State", and "Base Menu State".

---

## **Phase 2: Player Metrics & Movement Physics**
**Goal:** Implement player physics for both sailing in the Overworld (top-down) and swimming in Caves (side-view), along with resource metrics (oxygen, health, stamina).

### **Tasks**
- [x] **2.1. Player Struct and Stats (`internal/game/player.go`)**
  - Create `Player` struct with attributes: `X, Y float64`, `Vx, Vy float64`, `Width, Height float64`.
  - Add status bars: `MaxHealth, CurrentHealth`, `MaxOxygen, CurrentOxygen`, `MaxStamina, CurrentStamina (Max / Current with regeneration)`, `MaxEnergy, CurrentEnergy`.
- [x] **2.2. Overworld Top-Down Sailing Physics (`internal/game/overworld.go`)**
  - Implement WASD steering for the player's boat/pod on the water surface.
  - Apply top-down velocity, friction, and steering inertia (turning circles).
- [x] **2.3. Cave Side-Scrolling Swimming Physics (`internal/game/cave.go`)**
  - Implement 2D side-view physics.
  - **Buoyancy**: Add vertical upward drift when player stops moving downward (simulating early-game buoyancy).
  - **Drag**: High fluid resistance limiting horizontal and vertical velocities.
  - **Inertia**: Smooth acceleration and deceleration.
- [x] **2.4. Oxygen and Stamina Depletion Loops**
  - Oxygen drains constantly inside the cave state.
  - Oxygen recovers instantly when reaching the surface or entering the Overworld state.
  - Stamina drains during fast swimming (pressing Shift) and regenerates when moving normally or resting. Max stamina slowly decreases over long periods until resting/eating.
- [x] **2.5. HUD Rendering (`internal/game/hud.go`)**
  - Render progress bars for Health (Red), Oxygen (Blue), and Stamina (Green / Yellow) in the corner of the screen.

### **Verification**
- Run game and switch to "Overworld". Verify top-down boat controls (smooth sliding and turning).
- Switch to "Cave". Verify side-view swimming. Check buoyancy (floating up) and high drag.
- Verify HUD gauges drain and refill appropriately based on activities.

---

## **Phase 3: Procedural Generation & Cave Grid Transition**
**Goal:** Generate a procedural overworld ocean/island grid, procedural caves beneath sinkholes, and manage dynamic grid-to-grid loading.

### **Tasks**
- [x] **3.1. Surface Generation (`internal/world/noise.go`)**
  - Use 2D Simplex/Perlin noise to generate an overworld ocean map.
  - Define passable ocean tiles, impassable island/reef tiles, and specific "Trench/Sinkhole" tiles.
- [x] **3.2. Cave Transition Logic (`internal/game/game.go`)**
  - When the player boat overlaps with a Trench tile in the Overworld, present an interaction prompt (e.g., "Press [E] to Dive").
  - On transition, record the player's Overworld coordinates and instantiate a 2D side-view Cave Grid corresponding to that trench.
- [x] **3.3. Cellular Automata for Shallow Caves (`internal/world/cellular.go`)**
  - Implement Cellular Automata (typically 4-5 simulation steps of birth/survival rules) to generate organic, bubble-like shallow cave pockets.
- [x] **3.4. Drunkard's Walk for Deep Crevices (`internal/world/drunkard.go`)**
  - Implement a Drunkard's Walk algorithm heavily weighted to dig downwards.
  - Generate narrow, winding vertical shafts connecting shallow chambers to deep biomes.
- [x] **3.5. Solid-Tile Collisions**
  - Build basic AABB (Axis-Aligned Bounding Box) collision detection between the player's box and solid cave tiles.
  - Prevent player from swimming through cave walls.

### **Verification**
- Explore the Overworld, find a Trench, and press [E] to load into the Caves.
- Verify that the cave loads correctly, is populated by solid blocks, and the player cannot pass through blocks.
- Verify transitioning back to the surface places the player next to the corresponding Trench.

---

## **Phase 4: Dynamic Lighting, Shaders & Camera Tracking**
**Goal:** Set up a smooth tracking camera and implement the line-of-sight flashlight using Ebitengine's Kage shader language.

### **Tasks**
- [x] **4.1. Smooth Camera Controller (`internal/game/camera.go`)**
  - Create a Camera struct with X, Y coordinates.
  - Implement linear interpolation (Lerp) tracking so the camera smoothly pans to center on the player's position.
- [x] **4.2. Kage Flashlight Shader Setup (`internal/shader/light_cone.kage`)**
  - Write a Kage fragment shader.
  - Pass uniforms: player coordinate on screen, flashlight facing angle (from mouse position), and light cone angle (e.g., 45 degrees).
  - Darken screen pixels that fall outside the light cone, creating a pitch-black vignette.
- [x] **4.3. Light Source Rendering**
  - Integrate the shader into Ebitengine's rendering pipeline.
  - Draw light overlay on top of the rendered cave map.
- [x] **4.4. Bioluminescent Highlights**
  - Ensure items marked as "bioluminescent" (flora, predator eyes) are rendered on a layer *above* the shader light overlay so they glow in the dark.

### **Verification**
- Test camera scrolling in both Overworld and Cave views.
- Move mouse cursor around player in the Caves; verify the lit cone rotates towards the mouse, while the rest of the screen remains dark.
- Check that simulated "bioluminescent" test points remain visible even in pitch black.

---

## **Phase 5: Resource Gathering & Inventory System**
**Goal:** Implement inventory grids, item stack management, mineable resource nodes, and player mining mechanics.

### **Tasks**
- [ ] **5.1. Inventory Struct (`internal/game/player.go`)**
  - Create an `Inventory` struct containing an array of item slots (e.g., 24 slots).
  - Implement logic for item definitions (ID, Name, Icon/Color, MaxStackSize).
  - Add helper functions: `AddItem()`, `RemoveItem()`, `CanFit()`.
- [ ] **5.2. Resource Node Entities (`internal/entity/resource.go`)**
  - Define `ResourceNode` struct: Type (Titanium, Copper, Quartz), Position, HitsToMine (e.g., 3 hits).
  - Place these nodes on random solid cave tiles during generation.
- [ ] **5.3. Mining Mechanic**
  - When the player is within range of a resource node and presses the left mouse button, trigger a mining strike.
  - Reduce HitsToMine, play a visual particle effect, and add the resource to the player's inventory on destruction.
- [ ] **5.4. Inventory UI Overlay**
  - Render an inventory screen showing grid slots, icons, and stack numbers when the player presses `Tab`.

### **Verification**
- Swim through a cave, find a resource node (colored block), and click it to mine.
- Verify resource node collapses and drops items.
- Press `Tab` and verify items are displayed in the inventory slots.

---

## **Phase 6: Modular Management Menu & Base Building**
**Goal:** Implement the menu-based base upgrade system (No tile-building, pure menu interaction for fabricator, storage, and upgrades).

### **Tasks**
- [ ] **6.1. Anchor Terminals & Interactivity**
  - Create anchor points in the world (e.g., the starting Life Pod on the surface, or deployable Base Modules).
  - Interact with an anchor to switch the game state to `StateBaseMenu`.
- [ ] **6.2. Base Management Menu UI Layout (`internal/ui/menu.go`)**
  - Create a tabbed UI interface containing:
    - **Overview Tab:** Shows current module schematic slots (e.g., Slot 1: Fabricator, Slot 2: Storage, Slot 3: Infirmary).
    - **Fabricator Tab:** Lists craftable items, raw materials required, and button to craft.
    - **Storage Tab:** Grid transfer UI to move items from player inventory to base inventory.
    - **Medical Tab:** Spend base power to heal the player or cure statuses.
- [ ] **6.3. Upgrades and Modules Crafting Engine**
  - Implement recipes verification: check player inventory for requirements, consume ingredients, and spawn the upgrade item.
  - Support O2 Tank upgrades (which directly increase player `MaxOxygen`), Fins upgrade (increases speed), and Scanner tool.
- [ ] **6.4. Base Power Loop**
  - Base has a power supply that slowly regenerates if surface-mounted solar modules exist, and drains when using fabricator or medical stations.

### **Verification**
- Park boat next to the Life Pod, press `E` to open Base Menu.
- Navigate between Fabricator, Storage, and Medical tabs.
- Gather raw titanium, craft a "High Capacity O2 Tank", and verify that player's max oxygen capacity is increased.

---

## **Phase 7: Vehicle Hierarchy**
**Goal:** Build vehicle entities, handle vehicle piloting states, depth limits, and vehicle modules.

### **Tasks**
- [ ] **7.1. Vehicle Interface (`internal/vehicle/vehicle.go`)**
  - Define `Vehicle` interface containing methods: `Update()`, `Draw()`, `GetPos()`, `GetOxygen()`, `TakeDamage()`, `GetPerspective()`.
- [ ] **7.2. The Skiff (Surface Boat)**
  - Implement Skiff rendering in Overworld.
  - Allow player to enter/exit. When piloting, player inventory is extended with Skiff's large storage.
  - Implement Solar Charging module: charges internal battery in daytime.
- [ ] **7.3. The Scout Sub (Cave Mini-Sub)**
  - Small, agile 2-tile wide physics entity in the Caves.
  - Prevents player oxygen depletion while inside.
  - **Sonar module**: Pressing `Q` emits a radial wave that temporarily reveals cave outlines on the map (bypassing flashlight vignette for 3 seconds).
- [ ] **7.4. The Heavy Mech**
  - Heavy physics entity: ignores buoyancy and sinks to the bottom.
  - Immune to minor predator damage.
  - **Drill Arm module**: Allows breaking deep "Abyssal Ore Blocks".
- [ ] **7.5. Vehicle Depth Limits**
  - Define depth limits for each vehicle. Exceeding the depth limit inflicts damage over time.

### **Verification**
- Build and enter the Scout Sub. Verify movement controls inside caves.
- Press `Q` to trigger Sonar Ping and confirm it briefly illuminates the surrounding cave grid.
- Take Scout Sub below its depth threshold and verify it takes hull damage.

---

## **Phase 8: Biomes & Predator AI**
**Goal:** Implement biological behaviors for specific biomes, including visual palettes and interactive predator entities.

### **Tasks**
- [ ] **8.1. Biome Mapping during Generation**
  - Divide the cave grid height map into three bands:
    - **Mid-Depth (Cyan/Teal):** Luminous Pneumatophore Grotto.
    - **Deep (Dark Grey/Orange):** Silicate Smoker Trenches.
    - **Abyssal (Vantablack/White):** Benthic Brine-Falls.
- [ ] **8.2. Biome 1 Entities: Shatter-Bulbs & False-Bulb Snare**
  - **Shatter-bulbs**: Breakable entities that release 20 units of oxygen but trigger a "sound radius" alert on pop.
  - **False-Bulb Snare AI**: Suspended ceiling predator. Mimics Shatter-bulb.
    - *AI Behavior*: Stays frozen if player's flashlight coordinates cross its position. If the player's flashlight points away (light vector doesn't intersect), it charges the player.
- [ ] **8.3. Biome 2 Entities: Brimstone Siphons & Thermocline Rammer**
  - **Brimstone Siphons**: Shoots fire/steam particles vertically or horizontally on a timer, hurting player on contact.
  - **Thermocline Rammer AI**: Eyeless fish.
    - *AI Behavior*: Ignores light. Aggros if player uses vehicle thrusters or swims fast within detection radius. Charges in straight horizontal or vertical lines. If it hits a cave wall, it enters a "stunned" state for 3 seconds.
- [ ] **8.4. Biome 3 Entities: Pallid Nerve-Mats & Electro-Weaver**
  - **Nerve-mats**: Ground plants that apply a slow debuff if player collides with them.
  - **Electro-Weaver AI**: Serpentine stalker.
    - *AI Behavior*: Tracks electrical output (flashlight on or sonar ping used). When tracking, screen UI begins to jitter and flickering lights occur. After 5 seconds of tracking, it strikes. Turning off flashlight and cutting vehicle power causes it to lose tracking.

### **Verification**
- Enter Biome 1. Pop a Shatter-bulb, confirm O2 is restored. Point light away from False-Bulb and verify it lunges.
- Enter Biome 2. Dodge heat vents. Trigger fast swim near Rammer, verify it charges and stuns itself against walls.
- Enter Biome 3. Turn on sonar, watch UI flicker, turn off lights to evade Weaver.

---

## **Phase 9: Progression, Polish, & Integration**
**Goal:** Connect all systems together to form a playable loop, add game-over conditions, and refine visual/audio feedback.

### **Tasks**
- [ ] **9.1. Game Loop Integration**
  - Define the win condition: Building a specific surface Escape Rocket using resources collected from the deepest abyssal trenches.
  - Define lose conditions: Player running out of oxygen (drown) or health (predator attack) with respawning at the Life Pod (with inventory drop).
- [ ] **9.2. Day/Night Cycle in Overworld**
  - Track time of day. Overworld changes lighting from bright blue to dark navy.
  - Solar modules stop charging vehicles at night.
- [ ] **9.3. Visual Polish & Particle Systems**
  - Add bubbles particle effects rising from the player and vehicles while underwater.
  - Add mining debris particles when hitting rocks.
  - Add screen shake when vehicles take collision damage or get rammed.
- [ ] **9.4. Audio System (Optional / Placeholders)**
  - Integrate Go audio libraries (`github.com/hajimehoshi/ebiten/v2/audio`).
  - Add basic sound effects for: diving, mining, taking damage, and sonar pings.

### **Verification**
- Test the complete game loop from starting at the Life Pod, diving for raw materials, upgrading gear, building vehicles, going deeper, and gathering Abyssal Ore to construct the final escape project.
