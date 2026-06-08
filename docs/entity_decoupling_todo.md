# Decoupling & Entity Isolation Todo List

This document lists the remaining entities and hazards in the codebase that should be refactored or moved to align with the **Consumer-Defined Interface Pattern**.

---

## 🌊 Overworld Entities

### 1. `Whirlpool` (Move & Decouple)
- [x] **Location**: Move `/Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/whirlpool.go` to the overworld entity package at `internal/game/entity/overworld/whirlpool.go`.
- [x] **Interface**: Define `WhirlpoolContext` in the new file:
  ```go
  type WhirlpoolContext interface {
  	GetWorld() *world.World
  	GetBaseStation() *base.BaseStation
  }
  ```
- [x] **Signature**: Change `Whirlpool.Update(w *world.World, baseStationPos gvec.Vec2)` to `Whirlpool.Update(g WhirlpoolContext)`.
- [x] **Caller**: Update [overworld.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/scene/overworld.go) to use the new package type `oe.Whirlpool` and updated methods.

### 2. `CosmeticFish` (Refactor Signature)
- [ ] **Location**: [cosmetic_fish.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/overworld/cosmetic_fish.go)
- [ ] **Interface**: Define `CosmeticFishContext`:
  ```go
  type CosmeticFishContext interface {
      PlayerCenter() gvec.Vec2
      IsSolid(x, y float64) bool
  }
  ```
- [ ] **Signature**: Change `Update(playerCenter gvec.Vec2, isSolid func(x, y float64) bool)` to `Update(g CosmeticFishContext)`.

---

## 🪨 Cave Entities

Currently, all cave entities implement `entity.CaveEntity` and accept the large `entity.Runtime` interface (19 methods). To satisfy the **Interface Segregation Principle (ISP)**, they should declare their own narrow interfaces.

### 3. `PassiveCrab`
- [ ] **Location**: [creatures.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/creatures.go)
- [ ] **Interface**: Define `CrabContext`:
  ```go
  type CrabContext interface {
      PlayerPos() gvec.Vec2
      PlayerDims() gvec.Vec2
      FlashlightOn() bool
      PlayerFacing() float64
      PlayerFacingActiveVehicle() float64
      IsSolid(x, y, w, h float64) bool
  }
  ```

### 4. `FalseBulbSnare`
- [ ] **Location**: [hazards.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/hazards.go)
- [ ] **Interface**: Define `SnareContext`:
  ```go
  type SnareContext interface {
      PlayerPos() gvec.Vec2
      PlayerDims() gvec.Vec2
      FlashlightOn() bool
      PlayerFacing() float64
      HasActiveVehicle() bool
      ActiveVehicleFacing() float64
      SoundWaveTimer() int
      SoundWaveX() float64
      SoundWaveY() float64
      Emit(cmd GameCommand)
  }
  ```

### 5. `ThermoclineRammer`
- [ ] **Location**: [hazards.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/hazards.go)
- [ ] **Interface**: Define `RammerContext`:
  ```go
  type RammerContext interface {
      PlayerPos() gvec.Vec2
      PlayerDims() gvec.Vec2
      PlayerVel() gvec.Vec2
      IsPlayerSprinting() bool
      HasActiveVehicle() bool
      ActiveVehicleMoving() bool
      ActiveVehiclePos() gvec.Vec2
      ActiveVehicleDims() gvec.Vec2
      SoundWaveTimer() int
      SoundWaveX() float64
      SoundWaveY() float64
      IsSolid(x, y, w, h float64) bool
      Emit(cmd GameCommand)
  }
  ```

### 6. `ElectroWeaver`
- [ ] **Location**: [hazards.go](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/game/entity/hazards.go)
- [ ] **Interface**: Define `WeaverContext`:
  ```go
  type WeaverContext interface {
      PlayerPos() gvec.Vec2
      PlayerDims() gvec.Vec2
      FlashlightOn() bool
      SonarActive() bool
      HasActiveVehicle() bool
      TimeOfDay() float64
      IsSolid(x, y, w, h float64) bool
      Emit(cmd GameCommand)
  }
  ```

> [!NOTE]
> For cave entities to satisfy `entity.CaveEntity` while using narrow interfaces, their method signature in the interface can remain generic (accepting `gr Runtime`), but inside each individual file, you can define helper functions or assign type assertions to validate compliance with the segregated interfaces.
