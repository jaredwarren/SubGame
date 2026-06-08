# Entity Isolation and Decoupling Guidelines

This document outlines the pattern for isolating entities and interactive objects in the codebase. The goal of this pattern is to prevent cyclic package imports (import loops), make the codebase easier to reason about, and simplify writing mocks and unit tests.

---

## The Core Problem

When entities (like `FloatingCrate` or `ThermalVent`) depend on the concrete `Game` struct or a monolithic `GameContext` interface defined in a parent scene package:
1. **Import loops occur**: The scene package must import the entity package to spawn and manage entities. If the entity package imports the scene package (for `GameContext`), the compiler rejects the cycle.
2. **Interface Bloat**: Entities are forced to depend on methods they don't care about (e.g., scene transitions, rendering shader configuration, etc.), violating the **Interface Segregation Principle (ISP)**.
3. **Harder Testing**: Mocking a monolithic context with 30+ methods requires a lot of boilerplate or leads to fragile tests.

---

## The Solution: Consumer-Defined Context Pattern

In Go, **interfaces belong to the package that consumes the values**, not the package that provides them. By defining a narrow, focused interface inside the entity's package:
- The entity package remains a "leaf" node in the dependency graph, importing only base packages (like `player` or `vehicle`).
- The parent scene package automatically satisfies the interface since Go's interface satisfaction is implicit.
- Unit tests only need to mock the few methods the entity actually uses.

### Rule 1: Define a Localized Interface

Define the interface near the struct that needs it. Name it `<TypeName>Context`.

```go
// package overworld

type CrateContext interface {
	GetPlayer() *player.Player
	GetActiveVehicle() vehicle.Vehicle
	SpawnDebris(x, y float64, clr color.RGBA)
	TriggerScreenShake(duration int, intensity float64)
	SetMineWarning(msg string, duration, level int)
}
```

> [!NOTE]
> Since package `overworld` only imports standard types (`player`, `vehicle`, `gvec`), it does not cause a cycle with `scene`. `scene` imports `overworld` to draw/update crates, which compiles cleanly.

---

### Rule 2: Pass Only the Context

Avoid passing transient parameters (like target centers, piloting state, or spatial grids) to the `Update` method. Instead, query the context and compute them internally.

#### ❌ Anti-Pattern (Coupled & Bloated Parameters)
```go
func (v *ThermalVent) Update(g GameContext, targetCenter gvec.Vec2, targetDims gvec.Vec2, isPiloting bool) {
    // ...
}
```

####  Pattern (Decoupled & Encapsulated)
```go
func (v *ThermalVent) Update(g VentContext) {
	p := g.GetPlayer()
	isPiloting := false
	var targetCenter gvec.Vec2
	var targetDims gvec.Vec2

	activeVeh := g.GetActiveVehicle()
	if activeVeh != nil {
		isPiloting = true
		vPos := activeVeh.GetPos()
		targetDims = activeVeh.GetDimensions()
		targetCenter = gvec.Vec2{X: vPos.X + targetDims.X/2.0, Y: vPos.Y + targetDims.Y/2.0}
	} else {
		targetCenter = gvec.Vec2{X: p.Pos.X + p.Width/2.0, Y: p.Pos.Y + p.Height/2.0}
		targetDims = gvec.Vec2{X: p.Width, Y: p.Height}
	}
    
    // ... rest of the logic remains unchanged ...
}
```

---

### Rule 3: Hide Layout / Grid Structures Behind Context Methods

If an entity performs collision or solidity checks against a tile grid:
- Do **NOT** pass `grid [][]bool` to the update method.
- Add an `IsSolid(x, y, w, h float64) bool` method to the context interface instead.

#### ❌ Anti-Pattern (Exposing Grid Layout)
```go
func (f *PassiveFish) Update(gr Runtime, CaveGrid [][]bool) {
    if !isSolid(CaveGrid, nextX, nextY, f.Dimensions.X, f.Dimensions.Y) { ... }
}
```

####  Pattern (Hiding Grid Layout)
```go
func (f *PassiveFish) Update(gr Runtime) {
    if !gr.IsSolid(nextX, nextY, f.Dimensions.X, f.Dimensions.Y) { ... }
}
```

> [!TIP]
> The concrete runtime adapter (`entityRuntimeAdapter`) implemented in the `game` package has access to the current scene's grid and handles the actual calculation. The entity remains completely oblivious to how the grid is stored.

---

### Rule 4: Use Self-Contained Constructors

Constructors (like `NewPassiveFish` or `NewThermalVent`) should accept their starting parameters and initialize local random sources or configurations rather than referencing external package loops.

```go
func NewPassiveFish(x, y float64, facingRight bool, swimPhase float64) *PassiveFish {
	return &PassiveFish{
		BaseEntity: BaseEntity{
			Type:       EntPassiveFish,
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 20, Y: 12},
			Active:     true,
		},
		FacingRight: facingRight,
		SwimPhase:   swimPhase,
	}
}
```
