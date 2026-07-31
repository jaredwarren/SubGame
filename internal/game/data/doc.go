// Package data is the browseable catalog of compile-time game balance and content tables.
//
// Ownership (chosen to avoid import cycles):
//   - Owned here: player defaults, lore entries, crafting recipes
//   - Re-exported from domain packages: entity/vehicle archetypes, materials, cave biomes, resource gen config
//
// Domain packages keep runtime logic (Update/Draw/factories). Use this package as the
// index when tuning balance; follow re-exports to the owning package when needed.
package data
