// Command webserve hosts the web/ directory for browser and phone testing.
// Run `make serve`, then open http://localhost:8069 (localhost required for PWA install).
package main

import (
	"log"
	"net/http"
	"path"
	"strings"
)

func main() {
	addr := ":8069"
	root := http.Dir("web")
	fileServer := http.FileServer(root)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ensure directory URLs end with / so relative PWA paths resolve correctly.
		if r.URL.Path != "/" && !strings.Contains(path.Base(r.URL.Path), ".") && !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
			return
		}
		setPWAHeaders(w, r.URL.Path)
		fileServer.ServeHTTP(w, r)
	})

	log.Printf("Serving web/ on http://localhost%s", addr)
	log.Printf("PWA install needs a secure context: use http://localhost%s (not a LAN IP)", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func setPWAHeaders(w http.ResponseWriter, urlPath string) {
	switch {
	case strings.HasSuffix(urlPath, ".webmanifest"), strings.HasSuffix(urlPath, "manifest.json"):
		w.Header().Set("Content-Type", "application/manifest+json")
	case strings.HasSuffix(urlPath, "sw.js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Service-Worker-Allowed", "/")
	case strings.HasSuffix(urlPath, ".wasm"):
		w.Header().Set("Content-Type", "application/wasm")
	}
}
