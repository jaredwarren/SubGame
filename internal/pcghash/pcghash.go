// Package pcghash is the SubGame adapter for github.com/jaredwarren/pcg/spatialhash.
// Domain salts and feature flags live here; game types stay out of the core library.
package pcghash

// UseSpatialHash switches discrete resource/spawn rolls to spatialhash.
// Default false preserves legacy seed + ordered rand.Float64() behavior.
var UseSpatialHash bool

// Domain salts — independent streams at the same coordinates.
const (
	DomainDebris     uint64 = 0xdeb71501 // scrap / electronic waste presence
	DomainDebrisKind uint64 = 0xdeb71502 // weighted kind after a debris hit
	DomainOre        uint64 = 0xa11ce001
	DomainFauna      uint64 = 0xfa0fa001
	DomainDeco       uint64 = 0xdec04001
)
