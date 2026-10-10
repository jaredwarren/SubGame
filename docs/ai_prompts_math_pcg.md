# Prompting AI Agents for Math-Inspired PCG Libraries (Go)

This guide turns three research-inspired ideas from [openai/math](https://github.com/openai/math) into **copy-paste agent prompts** that produce **standalone, reusable Go libraries**—usable in SubGame, other games, tools, or simulations.

**Important framing for agents:** Treat the math families as *design inspiration and quality criteria*, not as a mandate to port research proofs. Deliver **practical approximations** with stable public APIs, tests, and zero coupling to any one game engine or content schema.

| Topic | openai/math family | Library property |
|---|---|---|
| Spatial hashing / flawless seeding | [007 – Ordinary two-point correlations of multiplicative functions](https://github.com/openai/math/blob/main/reasoning_traces/ordinary-two-point-correlations.pdf) | Coordinate → value hashes without lattice streaks |
| Constraint solvers under time budget | [102 – Ordinary NP-hardness at the basic semidefinite threshold](https://github.com/openai/math/blob/main/reasoning_traces/basic-semidefinite-threshold-np-hardness.pdf) | WFC / tile CSP with timed relaxation |
| Coupled discrete fields | [221 – Mézard–Parisi diluted spin glasses](https://github.com/openai/math/blob/main/reasoning_traces/mezard-parisi-formula.pdf), [271 – Spontaneous magnetization / Heisenberg](https://github.com/openai/math/blob/main/reasoning_traces/spontaneous-magnetization-quantum-heisenberg-ferromagnet.pdf) | Neighbor energy minimization → contiguous “territories” |

---

## Library-first vs game integration

| Layer | What the agent builds | Allowed dependencies |
|---|---|---|
| **Core libraries** (default) | Separate Go module(s) or packages under something like `pcg/{spatialhash,relaxwfc,spinglass}` | Stdlib only (`math`, `math/bits`, `math/rand`, `time`, `testing`) |
| **Optional adapters** | Thin wrappers for a specific game (e.g. SubGame) | May import the game; must not be required to use the core |

**Already partially modular?** Earlier drafts suggested `internal/pcg/` inside SubGame. That is swappable *within one repo*, but it is **not** a generic library: paths, types (`TileType`, `World`), and integration notes were SubGame-specific. This document asks for **importable, game-agnostic packages** first.

### Suggested module layout

```text
# Option A — one module with three packages
github.com/<you>/pcg/
  spatialhash/
  relaxwfc/
  spinglass/
  go.mod
  README.md

# Option B — three tiny modules (max reuse / independent versioning)
github.com/<you>/pcg-spatialhash
github.com/<you>/pcg-relaxwfc
github.com/<you>/pcg-spinglass
```

Place the new module **outside** `internal/` (or in a sibling repo). Games depend on it via `go.mod`; they never fork the algorithms into scene code.

---

## How to use these prompts

1. **One library per agent turn.**
2. Paste the **shared library preamble**, then one algorithm prompt.
3. Do **not** attach game source unless you are on an optional adapter follow-up.
4. Require acceptance tests (determinism, budget, no lattice artifacts) before any game wiring.
5. Only after the library merges: run a short **adapter** prompt for SubGame or another project.

### Shared library preamble (paste first)

```text
You are authoring a reusable Go library for procedural content generation.

Goals:
- Importable from any Go project (games, tools, servers). No Ebitengine, SDL, Unity,
  or project-specific types in the public API.
- Pure Go, no CGO, no BLAS/numpy ports, no GPU requirement.
- Deterministic given (seed + inputs). No time.Now, no map-iteration nondeterminism,
  no global rand.
- Allocation-light hot paths; document complexity and recommended budgets.
- openai/math family IDs are INSPIRATION. Deliver practical approximations with
  comments citing the family number and the intended property (low spatial
  correlation, hardness-aware relaxation, energy minimization)—not formal proofs.

Deliverables for this turn:
1. go.mod module path (propose github.com/<you>/pcg/... unless the user specifies)
2. Exported API + implementation
3. Table-driven tests + benchmarks where useful
4. README section: when to use, non-goals, example snippet
5. Explicit list of what must stay OUT of this package (game tiles, entities, rendering)

Do not wire into any game unless this prompt says "adapter".
```

---

## 1. Multiplicative spatial hashing (Family 007)

### Library goal

A **coordinate hash** for discrete rolls and seeds: ore, loot, spawn eligibility, decoration accents—anywhere you need `f(seed, x, y[, z])` without axis striping or order-dependent `rand` streams.

### Design brief

- Avalanche: flipping one bit of seed or a coordinate changes many output bits
- No obvious axis-aligned or diagonal stripes in scatter plots
- Stable `uint64` / `float64 ∈ [0,1)` API
- Domain/salt parameter so independent streams do not correlate
- Multiplicative mixer stage (odd constants, rotate-xor)—inspired by correlation-control ideas around multiplicative functions, **not** a number-theory library

### Agent prompt

```text
[Include shared library preamble]

Task: Implement Family-007-inspired spatial hashing as a standalone Go package.

Background (openai/math Family 007): ordinary two-point correlations of multiplicative
functions concern avoiding structured pairwise correlations. Translate that *property*
into a game/tool hash: seed + integer coordinates → values without noticeable lattice
correlations over large 2D/3D integer regions.

Package: spatialhash (module e.g. github.com/<you>/pcg/spatialhash)

```go
package spatialhash

// Hash2 returns a deterministic uint64 for (seed, x, y).
func Hash2(seed uint64, x, y int64) uint64

// Unit2 maps Hash2 to [0,1).
func Unit2(seed uint64, x, y int64) float64

// Hash3 / Unit3 for (x, y, z) — z may be depth, layer, or realm id.
func Hash3(seed uint64, x, y, z int64) uint64
func Unit3(seed uint64, x, y, z int64) float64

// Mix is Hash2/Hash3 with an explicit domain salt (ore vs fauna vs debris).
func Mix(seed, domain uint64, coords ...int64) uint64
func Unit(seed, domain uint64, coords ...int64) float64

// Chance returns Unit(...) < p (p in [0,1]).
func Chance(seed, domain uint64, p float64, coords ...int64) bool

// Stream returns an independent *rand.Rand for a chunk/cell (math/rand or math/rand/v2
// — pick one, document it, and keep it deterministic).
func Stream(seed uint64, chunkX, chunkY int64) *rand.Rand
```

Requirements:
1. Integer mixing only (large odd multipliers, xor-shifts, rotations). Document why
   this beats naive `x*A + y*B` for axis correlation.
2. Endian/platform-stable bit patterns for the same inputs.
3. Tests:
   - identical inputs → identical outputs
   - avalanche spot checks
   - 512×512 Unit2 histogram (16 bins): soft chi-square bounds; row/column means
     must not show obvious striping (document thresholds; keep test deterministic)
4. README example: spawning with Chance; seeding a chunk RNG with Stream.
5. Non-goals: not a Perlin/Simplex replacement (optional thin gradient-index helper
   is OK only if it calls Hash2 and stays in a subfile).

No game types, no file I/O, no rendering.
```

### Follow-up prompts

```text
Add a debug helper that returns [][]float64 for a window and an ASCII dump for
tests—still no graphics API.
```

```text
Add Hash4 and document recommended domain salt constants (user-defined; library
only provides examples in README).
```

### Acceptance checklist

- [ ] Importable with stdlib-only deps
- [ ] Deterministic across runs / GOARCH
- [ ] No visible stripes in a large scatter dump
- [ ] Domain salts produce uncorrelated streams for the same coordinates

---

## 2. Timed CSP / WFC with relaxation (Family 102)

### Library goal

A **budgeted tile/module constraint solver**: attempt hard rules, detect a hard regime, relax soft rules / fill fallbacks, always return a complete grid. Consumers supply their own tile IDs and adjacency rules (dungeons, cities, ship interiors, UI layouts, etc.).

### Design brief

Family 102 is about **NP-hardness near semidefinite thresholds**. Library translation:

1. Measure hardness online (contradictions, stalled entropy, time/steps).
2. Relax: drop lowest-weight soft constraints or use a soft-fail tile.
3. Bound work—never spin forever.

Do **not** implement a real SDP solver. Soft weights + thresholded dropping are enough.

### Agent prompt

```text
[Include shared library preamble]

Task: Implement Family-102-inspired constraint generation with timed relaxation
as a standalone Go package.

Background (openai/math Family 102): hardness near basic semidefinite thresholds
motivates switching to approximations when exact search becomes intractable.
Build a WFC-like (or greedy CSP) placer that relaxes before blowing the budget.

Package: relaxwfc (module e.g. github.com/<you>/pcg/relaxwfc)

```go
package relaxwfc

type TileID int16

type Direction uint8 // North, East, South, West — export consts

type Rule struct {
    A, B   TileID
    Dir    Direction
    Hard   bool    // if false, eligible to drop during relaxation
    Weight float32 // drop lowest weight first
}

type Config struct {
    Width, Height int
    Tiles         []TileID // palette (optional; may infer from Rules)
    Rules         []Rule
    Seed          uint64
    Budget        time.Duration // 0 = MaxSteps only
    MaxSteps      int
    SoftFailTile  TileID
}

type Result struct {
    Grid           [][]TileID
    Relaxations    int
    Contradictions int
    TimedOut       bool
    Steps          int
}

func Solve(cfg Config) (Result, error)
```

Algorithm sketch (refine as needed):
1. Per-cell domains; propagate hard rules.
2. On contradiction or budget pressure: drop lowest-Weight soft rules touching the
   conflict region; restore domains in a small radius; continue.
3. On Budget / MaxSteps: fill remaining cells with best compatible tile else SoftFailTile.
4. Expose counters on Result.

Requirements:
1. Deterministic for fixed Seed + Config.
2. Zero knowledge of any game: no biome names, no “wreckage”, no engine imports.
3. Tests with a tiny abstract tileset (A/B/C/D or Water/Wall/Door/Corridor as *examples*):
   - easy maps → Relaxations==0
   - over-constrained maps → full grid, no panics / no nil holes
   - Budget/MaxSteps path sets TimedOut and finishes quickly
4. README: how to encode adjacency; recommended budgets for offline vs per-frame.
5. Forbid: interior-point SDP, BLAS, CGO.

Optional progressive API in the same package (or follow-up):
Begin / Step / Finalize for spreading work across ticks—still game-agnostic.
```

### Follow-up prompts

```text
Add entropy-based cell selection and BenchmarkSolve_64x64. Document p95 guidance;
do not hard-fail CI on wall clock unless using a step-count assertion.
```

```text
Add Validate(grid, rules hard-only) helper so callers can assert invariants after Solve.
```

### Acceptance checklist

- [ ] Always returns a complete grid (or a documented error only for invalid Config)
- [ ] Soft rules drop in weight order
- [ ] Budget / MaxSteps honored
- [ ] Public API has no game/engine types

---

## 3. Spin-glass-style coupled fields (Families 221 & 271)

### Library goal

A **grid (or graph) energy minimizer** that settles interacting labels/scalars into contiguous regions—factions, biomes, wealth, danger, control maps, anything with neighbor coupling + external bias. Callers pass priors and biases; the library does not know “trench” or “faction lore.”

### Design brief

Ask for a discrete Ising/Potts-like anneal or coordinate descent—not a research replica of Mézard–Parisi / Heisenberg models.

Keep cell state generic:

- `Label int8` (Potts / category)
- optional continuous channels `[]float32` (e.g. danger, prosperity) *or* fixed fields with clear names but **no game semantics required**

External fields and neighbor weights come from `Config` / callbacks.

### Agent prompt

```text
[Include shared library preamble]

Task: Implement Family-221/271-inspired coupled field relaxation as a standalone
Go package.

Background:
- Family 221 (diluted spin glasses / Mézard–Parisi): conflicting couplings settle
  into stable low-energy states.
- Family 271 (Heisenberg / magnetization): local interactions form macroscopic
  ordered regions.

Deliver a lightweight annealer—not a scientific simulator.

Package: spinglass (module e.g. github.com/<you>/pcg/spinglass)

```go
package spinglass

type Cell struct {
    Label int8 // Potts spin; use Locked to pin
    // Channels are optional continuous fields (length decided by Config.NumChannels).
    // Keep them in a parallel []float32 buffer if that stays clearer—document one style.
    Channels []float32
    Locked   bool // if true, Label never changes
}

type Config struct {
    Width, Height int
    Seed          uint64
    NumLabels     int
    NumChannels   int
    Cells         []Cell // len = Width*Height
    // Bias[label][i] external field favoring label at cell i (nil = none)
    // Or provide:
    LabelBias func(cellIndex int, label int8) float64
    // ChannelBias(cellIndex, channel) prior pull for continuous fields
    ChannelBias func(cellIndex int, channel int) float64
    NeighborWeight float64 // J
    BorderPenalty  float64 // optional cost for disagreeing neighbors
    ChannelSmooth  float64 // neighbor pull for continuous channels
    Iterations     int
    Temperature0   float64
    Temperature1   float64
    Connectivity   int // 4 or 8
}

type Result struct {
    Cells  []Cell
    Energy float64
}

func Minimize(cfg Config) (Result, error)
func Energy(cfg Config, cells []Cell) float64
```

Algorithm:
1. Initialize unlocked labels from a deterministic hash of (Seed, index) if unset.
2. For each iteration: Metropolis or greedy label updates; smooth channels toward
   neighbors + ChannelBias (Gauss–Seidel). Deterministic scan order.
3. Anneal Temperature0 → Temperature1.
4. Document energy terms. Complexity O(iterations * width * height * neighbors).

Requirements:
1. Deterministic for fixed Config.
2. No imports of game packages; no terrain enums inside the library.
3. Tests:
   - two Locked opposing labels → contiguous territories under strong NeighborWeight
   - pure-greedy (T=0) run: energy non-increasing
   - 128×128 × 50 iters completes quickly in tests
4. README examples: abstract “region paint”, optional note that games map labels to
   factions/biomes *outside* this package.
5. Forbid heavy linear algebra; 4/8 stencil (optional diluted long-range edges later).

Optional: accept an adjacency list graph mode for non-grid maps (hex, navmesh)—
only if it stays small; otherwise leave a TODO interface.
```

### Follow-up prompts

```text
Add diluted long-range couplings: each cell links to K seeded partners (sparse
spin-glass style) without O(n²) cost.
```

```text
Add BorderPolylines(labels) returning edge segments between differing labels—
pure geometry, no rendering.
```

### Acceptance checklist

- [ ] Locked cells respected
- [ ] Contiguous blobs under default couplings (not salt-and-pepper)
- [ ] Same seed/config ⇒ identical result
- [ ] Callers inject all game meaning via bias callbacks / post-mapping

---

## Optional: SubGame (or any game) adapter prompts

Run **after** libraries exist. Keep adapters thin.

### Shared adapter preamble

```text
You are integrating existing PCG libraries into a specific game. Do not fork or
reimplement algorithms. Import the modules and translate game data ↔ library Config.
Preserve world seeds and add feature flags where behavior changes.
```

### SubGame adapter examples

```text
[Include shared adapter preamble]

Wire github.com/<you>/pcg/spatialhash into SubGame resource/fauna rolls:
- Call sites: internal/game/resource/generation.go, cave spawn helpers
- Keep world.Noise2D FBM for continuous landmasses
- Feature flag UseSpatialHash for A/B
- Tests: same seed ⇒ same rolls; changing map size must not reshuffle far cells
```

```text
[Include shared adapter preamble]

Use relaxwfc for one wreckage-corridor kit in
internal/game/cave/wreckage_corridor_cave.go. Map game tiles → TileID in the
adapter only. CA caves remain default for organic biomes. Budget ~2–4ms runtime
or longer for offline bake.
```

```text
[Include shared adapter preamble]

Run spinglass.Minimize on overworld using LandDist/WaterDist/tile maps as
LabelBias/ChannelBias. Store results on a sidecar Influence map (optional save
field). Drive fauna weights / sonar tint from channels—do not put SubGame types
inside the spinglass module.
```

### Combined bake (game tool, not core lib)

```text
Compose an offline tool (cmd/ or internal/devtools) that:
1. Uses spatialhash for all discrete RNG
2. Existing FBM + CA for geometry
3. relaxwfc for structural kits in masks
4. spinglass for influence fields
5. resource gen consuming Unit3 + influence

This tool may import SubGame packages; the three PCG libraries must remain
game-agnostic.
```

---

## Prompting patterns that work well

| Do | Don’t |
|---|---|
| Ask for a module path + stdlib-only core | Bury algorithms under `internal/game/...` |
| Cite family id + one *property* | Paste papers / demand Lean |
| Specify signatures, budgets, tests | “Make it mathematically correct” with no API |
| Keep game types in adapters | Put `TileWreckage` inside `relaxwfc` |
| Claim inspiration in README | Claim “implements openai/math result X” |
| One library per agent turn | One mega-prompt for lib + full game rewrite |

### Critic / review prompt (after code lands)

```text
Review the new pcg libraries:
1. Any game/engine imports in core packages?
2. Nondeterminism (map iteration, time.Now, global rand)?
3. Unbounded loops in relaxwfc?
4. Hidden O(n²) in spinglass?
5. Comments/README overclaiming openai/math equivalence? Soften to inspiration.
6. Smallest vertical slice for one consumer app (hash-only is fine).
```

---

## Suggested ship order

1. **spatialhash** module + tests + README  
2. **relaxwfc** module + abstract tileset fixtures  
3. **spinglass** module + locked-capital fixture  
4. Optional: SubGame (or other game) adapters, one call site each  

---

## References

- Collection root: [github.com/openai/math](https://github.com/openai/math)
- Family 007: [ordinary-two-point-correlations.pdf](https://github.com/openai/math/blob/main/reasoning_traces/ordinary-two-point-correlations.pdf)
- Family 102: [basic-semidefinite-threshold-np-hardness.pdf](https://github.com/openai/math/blob/main/reasoning_traces/basic-semidefinite-threshold-np-hardness.pdf)
- Family 221: [mezard-parisi-formula.pdf](https://github.com/openai/math/blob/main/reasoning_traces/mezard-parisi-formula.pdf)
- Family 271: [spontaneous-magnetization-quantum-heisenberg-ferromagnet.pdf](https://github.com/openai/math/blob/main/reasoning_traces/spontaneous-magnetization-quantum-heisenberg-ferromagnet.pdf)
- Catalogue / overview: [CONTENTS.md](https://github.com/openai/math/blob/main/CONTENTS.md), [overview.pdf](https://github.com/openai/math/blob/main/overview.pdf)

---

## Disclaimer

The openai/math manuscripts are research artifacts at varying verification stages. Libraries should claim **inspiration** (correlation control, hardness-aware approximation, interacting energy models), not equivalence to those proofs. Prefer tested, importable Go over academic fidelity.
