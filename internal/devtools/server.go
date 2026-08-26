package devtools

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
)

//go:embed templates/*.html
var templateFS embed.FS

func mustPage(files ...string) *template.Template {
	return template.Must(template.New("page").ParseFS(templateFS, files...))
}

var (
	worldPage = mustPage("templates/layout.html", "templates/world.html")
	cavePage  = mustPage("templates/layout.html", "templates/cave.html")
	audioPage = mustPage("templates/layout.html", "templates/audio.html")
)

type page struct {
	Title     string
	Active    string
	Seed      int64
	Error     string
	World     *WorldReport
	Cave      *CaveReport
	CaveType  string
	Biome     string
	CaveKinds []namedOpt
	Biomes    []namedOpt
	Audio     *AudioReport
}

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/world", http.StatusFound)
	})
	mux.HandleFunc("/world", handleWorld)
	mux.HandleFunc("/world/map.png", handleWorldPNG)
	mux.HandleFunc("/caves", handleCaves)
	mux.HandleFunc("/caves/map.png", handleCavePNG)
	mux.HandleFunc("/audio", handleAudio)
	mux.HandleFunc("/audio/wav", handleAudioWAV)
	return mux
}

func render(w http.ResponseWriter, t *template.Template, data page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleWorld(w http.ResponseWriter, r *http.Request) {
	seed := parseSeed(r, 12345)
	render(w, worldPage, page{
		Title:  "World",
		Active: "world",
		Seed:   seed,
		World:  inspectWorld(seed),
	})
}

func handleWorldPNG(w http.ResponseWriter, r *http.Request) {
	seed := parseSeed(r, 12345)
	report := inspectWorld(seed)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(report.png)
}

func handleCaves(w http.ResponseWriter, r *http.Request) {
	seed := parseSeed(r, 12345)
	kind := parseQuery(r, "type", "shallow")
	biome := parseQuery(r, "biome", "shallow_reef")
	p := page{
		Title:     "Caves",
		Active:    "caves",
		Seed:      seed,
		CaveType:  kind,
		Biome:     biome,
		CaveKinds: caveKindOptions(),
		Biomes:    caveBiomes,
	}
	report, err := inspectCave(kind, biome, seed)
	if err != nil {
		p.Error = err.Error()
		render(w, cavePage, p)
		return
	}
	p.Cave = report
	render(w, cavePage, p)
}

func handleCavePNG(w http.ResponseWriter, r *http.Request) {
	seed := parseSeed(r, 12345)
	kind := parseQuery(r, "type", "shallow")
	biome := parseQuery(r, "biome", "shallow_reef")
	report, err := inspectCave(kind, biome, seed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(report.png) == 0 {
		http.Error(w, "no cave grid", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(report.png)
}

func handleAudio(w http.ResponseWriter, r *http.Request) {
	seed := parseSeed(r, 12345)
	render(w, audioPage, page{
		Title:  "Audio",
		Active: "audio",
		Seed:   seed,
		Audio:  catalogReport(),
	})
}

func handleAudioWAV(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	seed := parseSeed(r, 12345)
	wav, err := synthesizeWAV(name, seed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Cache-Control", "max-age=120")
	_, _ = w.Write(wav)
}
