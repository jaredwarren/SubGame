# SubGame Expansion Plan: Deep-Sea Caves & Ecosystems

This document details the design and architectural plan for implementing the next phase of **SubGame** features. These additions aim to increase map variety, extend progression to infinite depth levels, balance resource harvesting, and bring the shallow biomes to life.

---

## 1. Overworld Additions & New Cave Types (Wreckage Caverns)

### Concept
Currently, the Overworld consists of islands, open water, and sinkhole Trenches. When the player enters a Trench, an organic cave is generated. To increase variety, we will add sunken ship **Wreckage Sites** to the Overworld. Diving into these sites loads a new type of cavern structured like a sunken vessel.

### Mechanics & Design
* **Wreckage Overworld Tile (`TileWreckage`)**: 
  - Procedurally scattered in deep ocean coordinates.
  - Rendered on the Overworld map as a dark iron-gray block with rust-brown borders.
  - Overlapping it prompts: `[E] to Salvage Wreckage`.
* **Wreckage Cave Generation**:
  - Instead of using Cellular Automata (bubble caves) or Drunkard's Walk (vertical crevasses), Wreckage Caves will generate a grid-aligned structure representing a ship's decks.
  - Features a central vertical elevator shaft (for player navigation and vehicle ascent/descent) connected to horizontal corridors (decks).
  - Branching off the corridors are rectangular rooms/compartments (cabins, cargo bays) separated by narrow door openings.
  - Walls are rendered in iron-gray with warning hazard stripe borders.
* **Salvage Loot Nodes**:
  - Instead of raw mineral crystals, wreckage rooms will spawn breakable `Scrap Metal` and `Electronic Waste` crates.
  - These can be harvested using the player's pickaxe or the Heavy Mech's drill, and then refined into Titanium and Copper at the Base Station terminal.

### Cave Interface & Extensibility
To support adding more cave types later (e.g. Ice Caves, Coral Caverns, Magma Vents) without bloating the transition and rendering states, all cave variations will be structured under a unified `Cave` interface.

* **`Cave` Interface Definition**:
  ```go
  type Cave interface {
      GetCaveType() CaveType
      GetGrid() [][]bool
      DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64)
      DrawTiles(screen *ebiten.Image, camX, camY float64, startTileX, startTileY, endTileX, endTileY int)
      GenerateEntities(seed int64) []CaveEntity
      GenerateResources(seed int64) []Resource
  }
  ```
* **Concrete Implementations**:
  - `ShallowSeabedCave`: Implements the open-topped sin/cos seabed layout.
  - `OrganicTrenchCave`: Implements the dual-layer cellular automata/drunkard's walk crevice.
  - `WreckageCorridorCave`: Implements the deck-room layout with hazard stripes and scrap loot.

### Files Involved
* [internal/world/generator.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/world/generator.go): Define the `Cave` interface, `CaveType` enum, wreckage generator, and factory methods.
* [internal/game/overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/overworld.go): Add wreckage rendering and prompt.
* [internal/game/cave.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/cave.go): Refactor scene logic to delegate drawing, background filling, and initialization hooks to the active `Cave` interface.
* [internal/game/item/item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go): Add `ScrapMetal` and `ElectronicWaste` items.
* [internal/game/menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/menu.go): Add refinement recipes (e.g. `1x Scrap Metal -> 2x Titanium`).

---

## 2. Infinite Overworld Edges & The Void (Ecological Dead Zone)

### Concept
Instead of constraining the player with a strict border around the generated 100x100 Overworld map, the player can sail outward infinitely in any direction. The coordinates outside the active play area transition into the "Ecological Void"—an eerie, empty expanse of dark deep-sea water. Diving here enters an empty, pitch-black cavern that drops down forever.

### Mechanics & Design
* **Removing Borders**:
  - Update boat collision checking in the Overworld to treat coordinates outside the map `[0, World.Width)` or `[0, World.Height)` as open, navigable water.
* **Void Map Rendering**:
  - Render out-of-bounds tiles as dark water squares (very dark navy/black) going on to infinity.
  - Draw a subtle vignette fade when the player transitions into the Void to build atmosphere.
* **HUD Telemetry Obfuscation**:
  - When the player is out of bounds on the Overworld:
    - Position display: `Pos: X: ??? Y: ???` (or coordinates fade out).
    - Current Zone: `Zone: Ecological Void`.
    - Dive Depth estimate: `Est. Dive Depth: ???`.
* **Void Cavern (`VoidCave` implementation)**:
  - If the player initiates a dive in the out-of-bounds void:
    - Generate a `VoidCave` instance satisfying the `Cave` interface.
    - **No walls**: Returns a `nil` cave grid, meaning collision is disabled and the player can swim horizontally and sink vertically infinitely.
    - **Pitch black**: `DrawBackground` fills the screen with solid pitch black, showing only the player's flashlight cone slicing into empty water.
    - **No life or minerals**: Spawns 0 creatures, hazards, or resource nodes.
    - HUD Depth telemetry continues to tick up as the player sinks, but they will never hit a bottom.
    - Swimming up past `Y < -8` returns the player to their out-of-bounds position on the Overworld.

### Files Involved
* [internal/game/overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/overworld.go): Update `isSolid` to allow moving out of bounds, and render empty void tiles outside the generated map size.
* [internal/game/hud.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/hud.go): Obfuscate position, zone name, and dive depth telemetry with `"???"` when out of bounds.
* [internal/world/generator.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/world/generator.go): Implement `VoidCave` struct and return it when `GetCave` is called with out-of-bounds coordinates.
* [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go): Handle transitions into `VoidCave`.

---

## 3. Resource Frequency & Density Scaling

### Concept
Reward players for taking the risk of diving deeper. Resource abundance, vein sizes, node health, and rarity will scale dynamically with the effective depth of the player.

### Mechanics & Design
* **Vein Density**:
  - Base spawn chance of mineral nodes will increase slightly per depth tier (e.g., from 4% in shallow waters to 8.5% at abyssal depths).
* **Rarity Distribution**:
  - **Shallow (<30m)**: Abundant Titanium (70%) and Copper (30%). No Quartz or Abyssal Ore.
  - **Mid (30m - 60m)**: Common Copper (40%) and Titanium (40%), introducing Quartz (20%).
  - **Abyssal (60m - 90m)**: Abundant Quartz (30%), Titanium (30%), Copper (30%), introducing Abyssal Ore (10%).
  - **Super Deep (>90m)**: Rarity shifts heavily to Quartz (35%) and Abyssal Ore (25%), with Titanium and Copper becoming scarce.
* **Node Durability**:
  - Deep resource blocks will require more hits to mine, scaling up pickaxe/drill usage and making vehicle drilling modules more crucial.

### Files Involved
* [internal/game/resource/resource.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/resource/resource.go): Modify `GenerateResourceNodes` to accept the `caveType` parameter and evaluate the vertical tile coordinate depth to weight density and types.

---

## 4. Passive Creatures in Shallow Caves

### Concept
Shallow caves currently feel empty and static. Introducing harmless, passive wildlife adds atmosphere, movement, and a source of food.

### Mechanics & Design
* **Cave Fish (Swimming)**:
  - Small, colorful fish that swim organically back and forth using a soft horizontal sine wave.
  - If the player swims near them, they dart away in the opposite direction.
* **Cave Crab (Crawling)**:
  - Small crabs that walk along horizontal solid blocks.
  - Subject to gravity: falls down shafts if they walk off edges.
  - If the player approaches or shines a flashlight on them, they withdraw into a shell, freezing in place.
* **Catching & Consuming**:
  - Players can interact with passive creatures when close to add them to their inventory as `Raw Fish` or `Raw Crab`.
  - At the Base Terminal Fabricator, players can cook them (`Cooked Fish`, `Cooked Crab`).
  - Eating cooked items from the inventory restores a chunk of Health and Stamina. Raw items can be consumed in an emergency for a small stamina boost but no health.

### Passive Creature Interface & Extensibility
To make it easy to introduce new passive wildlife in the future (e.g. Sea Snails, Jellyfish, Shrimp), all passive creatures will implement a dedicated `PassiveCreature` interface.
* **`PassiveCreature` Interface Definition**:
  ```go
  type PassiveCreature interface {
      CaveEntity
      GetHarvestedItem() item.Item
      CanCatch(playerPos gvec.Vec2) bool
  }
  ```
* **Concrete Implementations**:
  - `PassiveFish`: Swimming AI, returns `RawFish`.
  - `PassiveCrab`: Ground crawling/shell retreating AI, returns `RawCrab`.
* **Interaction Hook**:
  - Clicking on a `CaveEntity` in `game.go` will dynamically check if it satisfies the `PassiveCreature` interface. If it does, it queries `CanCatch()` and adds `GetHarvestedItem()` to player's inventory on success.

### Files Involved
* [internal/game/biome_entity.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/biome_entity.go): Define `PassiveCreature` interface, implement `PassiveFish` and `PassiveCrab`, and spawn them in `GenerateCaveEntities()`.
* [internal/game/item/item.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/item/item.go): Add raw and cooked food item structs and icons.
* [internal/game/menu.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/menu.go): Add cooking recipes to the Fabricator.
* [internal/game/game.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/game.go): Handle click detection on passive creatures for harvesting, and add food consumption hooks to inventory clicks.
