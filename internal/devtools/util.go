package devtools

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type nameCount struct {
	Name  string
	Count int
}

func sortedCounts(m map[string]int) []nameCount {
	out := make([]nameCount, 0, len(m))
	for name, n := range m {
		out = append(out, nameCount{Name: name, Count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func parseSeed(r *http.Request, def int64) int64 {
	raw := strings.TrimSpace(r.URL.Query().Get("seed"))
	if raw == "" {
		return def
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func parseQuery(r *http.Request, key, def string) string {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	return v
}

func bump(m map[string]int, name string) {
	m[name]++
}
