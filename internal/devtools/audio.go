package devtools

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jaredwarren/SubGame/internal/synth"
)

type soundInfo struct {
	Name string
}

type soundGroup struct {
	Name   string
	Sounds []soundInfo
}

type AudioReport struct {
	Groups []soundGroup
}

func catalogReport() *AudioReport {
	names := make([]string, 0, len(synth.SoundCatalog))
	for name := range synth.SoundCatalog {
		names = append(names, name)
	}
	sort.Strings(names)

	groups := map[string][]soundInfo{}
	var order []string
	for _, name := range names {
		prefix := "other"
		if i := strings.IndexByte(name, '/'); i >= 0 {
			prefix = name[:i]
		}
		if _, ok := groups[prefix]; !ok {
			order = append(order, prefix)
		}
		groups[prefix] = append(groups[prefix], soundInfo{Name: name})
	}

	out := make([]soundGroup, 0, len(order))
	for _, g := range order {
		out = append(out, soundGroup{Name: g, Sounds: groups[g]})
	}
	return &AudioReport{Groups: out}
}

type wavKey struct {
	name string
	seed int64
}

var (
	wavMu    sync.Mutex
	wavCache = map[wavKey][]byte{}
)

func synthesizeWAV(name string, seed int64) ([]byte, error) {
	gen, ok := synth.SoundCatalog[name]
	if !ok {
		return nil, fmt.Errorf("unknown sound %q", name)
	}
	key := wavKey{name: name, seed: seed}
	wavMu.Lock()
	if b, hit := wavCache[key]; hit {
		wavMu.Unlock()
		return b, nil
	}
	wavMu.Unlock()

	buf := gen(seed)
	if buf == nil {
		return nil, fmt.Errorf("generator returned nil for %s", name)
	}
	wav, err := buf.EncodeWAV()
	if err != nil {
		return nil, err
	}

	wavMu.Lock()
	if len(wavCache) > 48 {
		wavCache = map[wavKey][]byte{}
	}
	wavCache[key] = wav
	wavMu.Unlock()
	return wav, nil
}
