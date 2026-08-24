# Code Quality & Architecture Refactor Plan

A survey of the recurring problems in this codebase — inconsistency, copy-paste
duplication, non-idiomatic Go, and per-frame performance hazards — with a
proposed pattern for fixing each one. The unifying theme: **move content out of
code and into data-driven registries**, so adding a creature, item, ore, cave,
or vehicle means adding a definition row instead of editing 5–15 files.

Line numbers are approximate (as of this writing) but the files are exact.

---

## Table of contents

1. [The cost of adding content today](#1-the-cost-of-adding-content-today)
2. [Content is code, not data](#2-content-is-code-not-data)
3. [The `Game` god object and fat `GameContext`](#3-the-game-god-object-and-fat-gamecontext)
4. [Stringly-typed identity everywhere](#4-stringly-typed-identity-everywhere)
5. [Copy-paste duplication hot list](#5-copy-paste-duplication-hot-list)
6. [Per-frame performance hazards](#6-per-frame-performance-hazards)
7. [Save system fragility](#7-save-system-fragility)
8. [Smaller idiomatic-Go cleanups](#8-smaller-idiomatic-go-cleanups)
9. [Suggested migration order](#9-suggested-migration-order)

---

## 1. The cost of adding content today

Measured by counting the files/places that must be edited right now:

| Add one… | Touch points today | Should be |
|---|---|---|
| Simple inventory item | ~5 files (type stub, `item.go` registry, `sprites.go` coords, `data/recipes.go`, `data/items.go`) | 1 definition row |
| Mineable ore | ~7–8 files (above + `materials.go`, `resource/nodes.go` enum + registry + wrapper, `generation.go` tier fields, **two** mineral draw switches) | 1 definition row |
| Creature (biome-table fauna) | ~5–7 files (new `entity/*.go`, `archetypes.go`, `cave/spawn_ids.go`, `cave/spawn.go`, `cave/biome_config.go`, `data/entities.go`) | 1 def + optionally 1 behavior file |
| Creature (cave-scripted predator) | ~4–6 files (new type + hand-written spawn calls inside each `cave/*_cave.go`) | 1 def + spawn table row |
| Cave type (simple) | ~8–12 places (enum, cave impl, `TileType`, `tile_info.go` registry, `biome.go`, **two** music switches, map UI, devtools) | 1 spec + 1 generator func |
| Cave type (two-tier subterranean) | ~12–15+ places (above + `SubterraneanSpec`, chasm branches in `shallow_seabed_cave.go`, hardcoded fallback in `cave_update.go`) | 1 spec + generator funcs |
| Vehicle | new ~500-line file + save-name aliases + kit item + recipes | 1 def + shared controller |

Every section below works toward shrinking this table.

---

## 2. Content is code, not data

### 2.1 The `data/` package is a catalog of re-exports, not a source of truth

`internal/game/data/doc.go` declares the right intent, but `entities.go`,
`items.go`, and `biomes.go` are almost entirely type aliases and re-exported
pointers (`SandViperArchetype = entity.SandViperArchetype`). The runtime never
reads through `data/`; new content routinely skips it, so the "browseable
catalog" silently drifts out of date.

**Fix:** invert the ownership. Definitions (stats, colors, spawn weights,
stack sizes, music tracks) live in `data/` as plain structs with no Ebiten
imports; domain packages *consume* them. Delete the re-export shims. If an
import cycle blocks this, that is a signal the domain package is holding data
it shouldn't.

### 2.2 One concrete Go type per item, instantiated via reflection

`internal/game/item/item.go` (~100–151) keeps a
`map[reflect.Type]*ItemMetadata` registry plus `reflect.New` factories keyed by
**display name**. Every item is an empty struct type whose methods just read
the registry. Capabilities are modeled as ~10 assertion interfaces
(`O2UpgradeItem`, `SpeedUpgradeItem`, `Consumable`, …), and
`GetSpeedUpgrade() map[string]Speed` is stringly-typed on top.

**Fix — ID + definition table:**

```go
type ItemID string // stable, never shown to players: "titanium", "sonic_decoy"

type ItemDef struct {
    ID          ItemID
    Name        string        // display only; renaming never breaks saves
    MaxStack    int
    Color       color.Color
    Icon        IconSpec      // atlas coords or procedural spec
    Consumable  *ConsumableDef  // nil if not consumable
    Upgrade     *UpgradeDef     // O2 / speed / vehicle / base module in one struct
}

var Items = map[ItemID]*ItemDef{ ... } // or loaded from embedded JSON

type Stack struct { ID ItemID; Qty int; State *ItemState } // one runtime type
```

One runtime `Stack` type replaces ~40 empty structs, kills the reflection,
and makes saves key on stable IDs. Items with real state (vehicle kits) keep a
small optional `State` field instead of implementing `StatefulItem`.

### 2.3 Resource nodes: a registry undermined by wrapper types

`internal/game/resource/nodes.go` already has the right idea
(`nodeRegistry map[NodeType]*NodeTypeInfo`), then defeats it at ~197–288 with
eight wrapper structs kept only "for type-assertions compatibility":

```201:210:internal/game/resource/nodes.go
type TitaniumNode struct{ ResourceNode }

func (n *TitaniumNode) GetBaseItem() item.Item { return &item.Titanium{} }

func NewTitaniumNode(tx, ty int) *TitaniumNode {
	return &TitaniumNode{ResourceNode{
		BaseResourceNode: BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3},
		Type:             NodeTitanium,
	}}
}
```

Meanwhile `drawMineral`/`drawMineralIcon` switch on mineral **names**
(`"Copper"`, `"Quartz"`) in *both* `resource/sprites.go` (~326–367) and
`item/sprites.go` (~323–335), and `resource/generation.go` (~37–46) hardcodes
one weight field per ore (`TitaniumWeight, CopperWeight, …`) instead of a
weight list.

**Fix:** delete the wrappers — callers switch on `node.Type` (or better, use
registry lookups) instead of type-asserting. Replace per-ore weight fields
with the `[]ResourceSpawnEntry` pattern the biome path already uses. Fold
mineral drawing into one function parameterized by `MaterialDef` (add the
shape variant to the def) so a new ore is: one `MaterialDef`, one `NodeType`
row, one spawn-table entry.

### 2.4 Creatures: parallel `*Def` types, hand-wired spawns, duplicated AI

- `internal/game/entity/archetypes.go` (~5–279): ten near-identical `*Def`
  structs exposed as exported **mutable** pointers.
- `internal/game/cave/spawn_ids.go`: `FaunaID` covers only PassiveFish,
  PassiveCrab, SandViper. ElectroWeaver, ThermoclineRammer, FalseBulbSnare,
  VoltaicLurker, BrimstoneSiphon are spawned by hand inside each
  `cave/*_cave.go` generator.
- Targeting logic (decoy → vehicle → player, flashlight-cone checks) is
  copy-pasted across `false_bulb_snare.go` (~69–93),
  `thermocline_rammer.go` (~72–97), and `electro_weaver.go`.
- Some balance never made it into archetypes at all: hardcoded dims in
  `false_bulb_snare.go` (~33) and `thermocline_rammer.go` (~37), the magic
  `RestoreOxygenCmd{Amount: 20}` in `shatter_bulb.go` (~59).

**Fix — split "stats" from "behavior":**

```go
type FaunaDef struct {
    ID        FaunaID
    Dims      gvec.Vec2
    Health    float64
    Damage    float64
    Speed     SpeedDef
    Senses    SenseDef        // aggro range, decoy range, light aversion
    Behavior  BehaviorID      // "ambush", "patrol_ram", "stationary_trap"…
    Drops     []DropEntry
}
```

Behaviors stay as Go (a small number of `Behavior` implementations picked by
`BehaviorID`); everything numeric lives in the def. Extend `FaunaID` +
`cave/spawn.go` to cover *all* creatures so caves spawn from weighted tables,
not bespoke loops. Extract shared helpers on `entity.Runtime` for target
acquisition and flashlight-cone math.

### 2.5 Caves: half-migrated to `CaveBiomeSpec`

`internal/game/cave/biome_config.go` defines `CaveBiomeSpec` with
`FloraSpawns` / `FaunaSpawns` / `MineralSpawns` / `Rules` — and only the
shallow seabed uses it. The four deep caves (`organic_trench_cave.go`,
`shock_kelp_cave.go`, `thermo_cave.go`, `wreckage_corridor_cave.go`) hardcode
their own chances (`0.60` kelp, `0.08` lurker…) and each contains a ~35-line
copy of the same coral-attachment spawn block (see §5). All biomes also share
one `DefaultSpawnRules`, so the "rules" knob is currently decorative.

**Fix:** finish the migration. A cave type becomes:

```go
type CaveSpec struct {
    Type         CaveType
    Biome        *CaveBiomeSpec   // palette + spawn tables + rules
    GenerateGrid func(r *rand.Rand, w, h int) [][]bool  // the only real code
    Music        string
    Ambient      [4]float32
    Subterranean *SubterraneanSpec // optional deep tier
}
```

Music/ambient move onto the spec so the duplicated `switch` in
`game.go` (~306–321) and `context_impl.go` (~145–156) collapse into
`spec.Music`. Fix the `ChasmTarget == 0` zero-value collision in
`shallow_seabed_cave.go` (~104–106) by making the zero value mean "unset"
(reorder the enum or use a pointer), and replace the chasm `if/else` on
concrete targets with data on the spec.

### 2.6 Vehicles: balance is data, behavior is triplicated

`vehicle/archetypes.go` extracts the numbers, but `skiff.go`, `scoutsub.go`,
and `heavymech.go` (~450–530 lines each) triplicate Update loops, stun gates,
battery drain, axis collision, `TakeDamage`/`Repair`/`RechargeBattery`, and
kit cloning.

**Fix:** one `vehicle.Controller` struct driven by a `VehicleDef` (movement
model, dims, battery, perspective, draw spec). Per-vehicle uniqueness — mech
mining arms, skiff surface handling — becomes small strategy hooks or flags on
the def, not whole files.

---

## 3. The `Game` god object and fat `GameContext`

- `internal/game/game.go` (~50–134): `Game` holds ~50 fields — scenes, player,
  world, camera, vehicles, cave caches, sonar, particles, story/quests,
  exploration, lost cargo, save slot, tutorial, debug flags — with 120+
  methods spread over `game.go`, `game_update.go`, `context_impl.go`.
- `internal/game/scene/scene.go` (~34–157): `GameContext` composes half a
  dozen sub-interfaces into a kitchen-sink facade; `context_impl.go` is ~422
  lines of pass-through getters/setters.
- `entity_runtime_adapter.go` / `vehicle_runtime_adapter.go`: import-cycle
  workarounds that buffer `Emit(cmd)` commands and drain them through large
  type-switches (`game_update.go` ~655–701). Every new interaction needs a new
  command struct *and* a new switch arm.
- Dual state: `currentScene` + `currentState` must be manually kept in sync
  (`TransitionTo` sets one, each scene's `OnEnter` sets the other). A missed
  `SetCurrentState` silently breaks input, HUD, save, and audio.
- Field visibility is arbitrary: `Input`, `ActiveVehicle`, `Particles`,
  `TimeOfDay` are exported and mutated freely while `player`, `world`,
  `camera` are private behind accessors.

**Fix — split into subsystems with narrow interfaces:**

```go
type Game struct {
    scenes    *SceneManager   // owns currentScene; state derived, not duplicated
    session   *Session        // player, world, camera, vehicles, base
    caves     *CaveCache      // caveNodes, caveEntities, per-trench vehicles
    fx        *Effects        // particles, shake, banners, sound waves
    progress  *Progress       // story, quests, recipes, exploration
    saves     *SaveManager
}
```

Scenes then depend on only the 2–3 subsystem interfaces they use instead of
`GameContext`. Derive the `State` enum from the active scene
(`scene.StateID()`) so there is one source of truth. Make export status
deliberate: subsystem structs export what scenes need, nothing else.

The world/session reset logic duplicated across `NewGame`, `StartGame`
(`context_impl.go` ~32–95), and `loadSaveFromPath` (~716–852) becomes one
`newSession(seed, fromSave *SaveData)` constructor so new-game / restart /
load can't diverge (they already have: tutorial titanium is granted in some
paths and not others).

---

## 4. Stringly-typed identity everywhere

The same disease in many organs — human-readable strings used as primary keys:

| Where | Example | Risk |
|---|---|---|
| Item saves (`item.go` ~140–151, `inventory.go` ~461–494) | `GetName()` → `NewItemByName("Raw Fish")` | rename breaks saves |
| Vehicle saves (`save.go` ~59–72, `vehicle/interface.go` ~59–70) | `"The Skiff"` aliases + `default:` returns a Skiff | unknown vehicle silently becomes a Skiff |
| Scene persistence (`game.go` ~581–584, ~833) | `"Overworld"` / `"Cave"` | typo = wrong scene on load |
| Vehicle location (`game.go` ~610, ~781) | `"overworld"` vs trench-key strings | two conventions for one concept |
| Quest tasks (`quest/quest.go` ~240–344) | giant `switch` on `"train_dive"`, `"gear_o2_hc"` | typos fail silently; every quest edits the switch |
| Entity attach dirs (`shock_kelp.go` ~21–36, `brimstone_siphon.go` ~49–53, `voltaic_lurker.go` ~63–69, `coral.go` ~59–63) | `"floor"`, `"left"`, `"up"` | `resource` package already has typed `AttachDirection`; entities don't use it |
| Speed upgrades (`item.go` ~64) | `map[string]Speed` keyed by mode name | unchecked keys |
| PDA/menu tabs (`context_impl.go` ~312–327, `game_update.go` ~218–224) | `ActiveTab = 4` magic ints | renumbering breaks navigation |
| Blueprint unlocks (`generation.go` ~165–176, `cave_update.go`) | match `recipe.NewResult().GetName() == resName` | order-dependent linear scans by display name |

**Fix:** introduce typed IDs (`ItemID`, `VehicleID`, `FaunaID` extended,
`QuestTaskID`, `SceneID`, `TabID`, shared `AttachDirection`) and use them for
persistence, lookups, and switches. Display names become a `Name` field on
the def, freely renameable. Quests specifically should move from a switch to
declarative predicates:

```go
type TaskDef struct {
    ID        QuestTaskID
    Condition Condition // e.g. HasItem{ID: "o2_tank_hc", N: 1}, ReachDepth{M: 300}
}
```

evaluated by a small interpreter — and ideally event-driven (see §6.5) rather
than polled every tick.

---

## 5. Copy-paste duplication hot list

Each of these is one extracted helper away from deleting 100+ lines:

1. **Coral spawn block ×5** — identical ~35-line attachment-scan/spawn loop in
   `shallow_seabed_cave.go` (~481–517), `organic_trench_cave.go` (~185–221),
   `shock_kelp_cave.go` (~192–228), `thermo_cave.go` (~160–196),
   `wreckage_corridor_cave.go` (~107–143). Extract
   `spawnCoral(grid, r, chance, biome)`.
2. **Cave trench transitions ×4** — `scene/cave_update.go` (~511–702): the
   down/left/right/up handlers each copy a ~40-line stash-old/build-key/
   factory/lazy-generate/recenter block, with already-drifted seed formulas.
   Extract `transitionCave(dir Direction)`.
3. **Cave music selection ×2** — `game.go` (~306–321) and `context_impl.go`
   (~145–156). Becomes `spec.Music` per §2.5.
4. **Vehicle save serialization ×2** — `game.go` `SaveGame` (~586–640)
   repeats the identical `SavedVehicle` construction for overworld and cave
   lists. Extract `serializeVehicle(v, location, active)`.
5. **Vehicle removal ×3** — `game.go` (~485–495), `game_update.go` (~558–564,
   ~795–811).
6. **Predator targeting AI ×3** — §2.4.
7. **Vehicle Update/collision/kit code ×3** — §2.6.
8. **Mineral drawing ×2 packages** — §2.3.
9. **Inventory slot drawing ×2** — `hud_inventory.go` repeats the slot loop
   for solo vs vehicle panels.
10. **Kelp stalk rendering ×2** — `kelp.go` (~18–56) vs `shock_kelp.go` draw;
    share one stalk renderer parameterized by palette (and give decorative
    kelp an archetype like everything else).

---

## 6. Per-frame performance hazards

These run at 60 fps; each is either steady GC pressure or avoidable CPU.

### 6.1 Allocations inside Draw/Update

| Site | Problem | Fix |
|---|---|---|
| `entity/coral.go` ~252, ~316, ~387 | `ebiten.NewImage(1,1)` + `Fill` **per coral per frame** for `DrawTriangles` | package-level white pixel (pattern already exists as `emptyImage` in `vehicle/rendering.go`) |
| `cave/shock_kelp_cave.go` ~31–54, `thermo_cave.go` ~32–61 | `rand.New(rand.NewSource(...))` 40× per `DrawBackground` frame | derive particle params arithmetically from `i` and ticks, or precompute |
| `cave_type.go` ~34 + impls | `GetAmbientColor() []float32` allocates a 4-float slice per frame | return `[4]float32` |
| `game_update.go` ~46, `context_impl.go` ~210–216 | fresh `vehicleRuntimeAdapter` / `entityRuntimeAdapter` per tick | store adapters as fields on `Game`, reset command buffers with `cmds[:0]` |
| `hud.go`, `game_draw.go` ~191 | `fmt.Sprintf` per frame for telemetry/markers | cache strings, re-format only when the value changes |
| `audio/manager.go` ~121, ~173, ~248 | `wav.DecodeWithSampleRate` on **every** SFX play; players never `Close()`d | cache decoded PCM per path, pool/close players |

### 6.2 O(n²) entity scans

`entity_runtime_adapter.go` (~141–254): `FindClosestDecoy`,
`CheckDeterrentOcclusion`, `CheckDeterrentSlowing` each walk all cave entities
with type assertions, called from every predator's Update → O(entities ×
predators) per tick, with the radius math copy-pasted twice.

**Fix:** build typed lists (`decoys []*SonicDecoy`, `clouds []*DeterrentCloud`)
once per tick before the entity update pass; predators query those.

### 6.3 Fog-of-war overlay

`scene/overworld_draw.go` (~174–219): every unexplored visible tile does up to
16 subcell fills, each calling a neighborhood-searching `fogDistToExplored` —
O(visible × 16 × search) every frame.

**Fix:** maintain a cached distance-to-explored field updated only when
`Reveal` changes tiles (the exploration tracker already knows exactly which
tiles changed), and render fog from the cache.

### 6.4 Per-tile smoothed water offsets

`world/generator.go` (~442–484) averages a 5×5 neighborhood per visible tile
per frame from `overworld_draw.go`. Precompute the smoothed map once at
world-gen.

### 6.5 Quest polling

`game_update.go` (~34, ~79–99) calls `questManager.CheckProgress(g)` every
tick; predicates like `CountInventoryItem` call `item.NewItemByName` and scan
inventories each time. **Fix:** make progress event-driven — pickup/craft/
depth-change events update counters; quests subscribe. Combines with §4's
declarative task conditions.

### 6.6 One-time but worth fixing

- `world/generator.go` ~235–244: BFS uses `queue = queue[1:]` on a 500×500
  map — use a head index or ring buffer.
- `tile_info.go` ~168–208: full-map scan + bubble sort per wreckage cave
  factory — precompute after scatter.
- `exploration/tracker.go` ~103–104: `newlyRevealed` grows unboundedly until
  the map UI is opened — drain or cap from the update loop.

---

## 7. Save system fragility

- **`Version` is written but never read** (`save/save.go` ~76–77): no
  migration path, so schema changes spawn ad-hoc dual representations — e.g.
  recipes persisted as **both** `UnlockedRecipes []int` (order-fragile) and
  `UnlockedRecipeNames []string` (rename-fragile).
- **Silent data loss on load:** vehicle `Facing` is saved but never restored;
  unknown vehicle names default to a Skiff; base station `Power` isn't in
  `SavedBaseStation` at all, so fabricator charge resets to 75 every load.
- **Autosave errors discarded:** `transition.go` (~95, ~125) does
  `_ = g.SaveGame()` on cave enter/exit — a failed write is invisible.
- Round-trip state (health/battery) is reconstructed through gameplay methods
  (`TakeDamage`, `RechargeBattery`) instead of set directly, which will
  misbehave once those methods gain side effects.

**Fix:**

1. Check `Version` on load; write migration funcs `v1→v2→…` so old saves
   upgrade instead of half-loading.
2. Persist stable IDs (§4) for items, vehicles, recipes, scenes. Keep the
   name-based path only inside a v1→v2 migration.
3. Add `Power` and any other missing fields; restore `Facing`; make unknown
   IDs a loud load error, not a free Skiff.
4. Surface autosave failures via the existing warning banner.
5. Give save structs direct setters (`v.SetHealth`) for hydration.

---

## 8. Smaller idiomatic-Go cleanups

- **Re-export shim files** left from a package split: `internal/game/scene.go`,
  `state.go`, `input.go` alias types from `internal/game/scene`. Finish the
  move and delete them.
- **Exported mutable globals** for balance: `entity.*Archetype` pointers,
  `resource.GenConfig`. Once defs live in `data/`, make them unexported with
  accessor functions, or at minimum document mutation ownership.
- **`math/rand` + `NewSource`** everywhere — prefer `math/rand/v2`
  (`rand.New(rand.NewPCG(...))`) for new code; it's faster and allocation-free
  for the common cases.
- **`internal/synth/presets.go`** (1,972 lines): a single
  `map[string]PresetGenerator` of inline closures. Split by category
  (`presets_ui.go`, `presets_creatures.go`, …) and register via small `init`
  or explicit assembly; not a runtime hot path, but a review/merge bottleneck.
- **UI layout constants:** `scene/layout.go` is the right pattern but only
  covers inventory grids. Menu chrome (`menu.go` draw ~376–801 mirrors its
  hit-testing constants in update ~126–351), HUD (~38–176), title/pause/intro
  all bake magic pixels. Define layout rects once and share them between draw
  and hit-test — this also makes `BaseMenuScene.draw` splittable into
  per-tab functions.
- **Magic gameplay numbers in glue code:** day length `14400` (twice), crush
  depth/damage, board/repair ranges, camera lerp `0.08` — move into
  `config`/`data` next to their peers.

---

## 9. Suggested migration order

Ordered for value ÷ risk; each step is independently shippable.

**Phase 1 — mechanical wins (no design changes)**
1. Shared white-pixel image for coral; kill per-frame `rand.New`; return
   `[4]float32` ambient; reuse runtime adapters. (§6.1)
2. Cache decoded SFX; close players. (§6.1)
3. Extract the duplicated helpers: coral spawn, cave transition, vehicle
   serialization, music lookup, mineral drawing. (§5)
4. Surface autosave errors; restore `Facing`; persist base `Power`. (§7)

**Phase 2 — identity and data foundations**
5. Introduce typed IDs (`ItemID`, `VehicleID`, `FaunaID`, `QuestTaskID`) and a
   versioned save with a v1→v2 migration off display names. (§4, §7)
6. Convert items to the definition-table model; delete per-item structs and
   reflection. (§2.2)
7. Delete resource-node wrapper types; spawn weights become tables. (§2.3)

**Phase 3 — data-driven content**
8. Extend `FaunaID`/spawn tables to all creatures; split behavior from stats;
   deep caves consume `CaveBiomeSpec`. (§2.4, §2.5)
9. `CaveSpec` registry with music/ambient/generator; fix `ChasmTarget` zero
   value. (§2.5)
10. Shared vehicle controller driven by `VehicleDef`. (§2.6)
11. Declarative, event-driven quest conditions. (§4, §6.5)

**Phase 4 — structural**
12. Split `Game` into subsystems; shrink `GameContext` to per-scene
    interfaces; derive `State` from the scene; unify the three session-reset
    paths. (§3)
13. Fog/water-offset caching; exploration drain; BFS queue. (§6.3, §6.4, §6.6)

**Phase 3 polish — deferred tail work**

Phase 3 shipped the core migrations (FaunaID/spawn tables, `CaveSpec`,
`VehicleDef`/`Controller`, declarative quest conditions). The following items
were intentionally left open; each is independently shippable but too large
for a single incidental commit:

14. **Unified `FaunaDef` registry** (§2.4): collapse the ten `*Def` structs in
    `entity/archetypes.go` into one `FaunaDef` table keyed by `FaunaID`, with
    `BehaviorID` picking a small set of Go behavior implementations. Creatures
    today still carry hardcoded dims in a few types (`false_bulb_snare.go`,
    `thermocline_rammer.go`) and magic numbers in glue (`shatter_bulb.go`).
    *Touch:* `entity/archetypes.go`, each predator file, `cave/spawn.go`,
    `data/entities.go`. Estimate: 1–2 focused PRs.

15. **Shallow chasm-rim data tables** (§2.5): `shallow_seabed_cave.go` still
    branches on `ChasmTarget` for draw palette, ambient tint, and rim entity
    spawns (~170 lines of `if target == OrganicTrench / else ShockKelp`).
    Move rim spawn weights, vein colors, and ambient overrides onto
    `CaveSpec.Subterranean` or a `ChasmRimSpec` keyed by target cave type.
    *Touch:* `cave/spec.go`, `shallow_seabed_cave.go`. Estimate: one PR.

16. **Full vehicle `Update` merge** (§2.6): `Controller` today covers
    damage/repair/recharge and stun gating only; `skiff.go`, `scoutsub.go`, and
    `heavymech.go` still triplicate movement, axis collision, battery drain,
    and kit cloning (~450 lines each). Finish extracting a shared update loop
    driven by `VehicleDef` with per-craft strategy hooks (mech mining arms,
    skiff surface handling).
    *Touch:* `vehicle/controller.go`, all three vehicle files. Estimate: 1–2 PRs.

17. **Event-driven quest progress** (§6.5): `CheckProgress` still runs every
    tick even though conditions are declarative. Emit progress events on
    inventory pickup, craft completion, depth milestones, vehicle deploy, and
    cave enter/exit; have `QuestManager` subscribe and only re-evaluate affected
    tasks. Combines with typed IDs (item 18 below — done).
    *Touch:* `quest/quest.go`, `game_update.go`, inventory/craft/transition
    hooks. Estimate: one PR.

18. ~~**Quest `ItemID` / `VehicleID` keys**~~ — **done**: `Condition` and
    `QuestContext` now use `item.ItemID` and `vehicle.VehicleID` instead of
    display names.

After phases 2–3, the table in §1 collapses: a new ore, creature, or cave is a
definition row plus (at most) one behavior or generator function — and the
compiler, not grep, tells you when something is missing.
