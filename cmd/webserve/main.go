// Command webserve hosts the web/ directory for browser and phone testing.
// Run `make serve`, then open http://<your-lan-ip>:8080 on a phone.
package main

import (
	"log"
	"net/http"
)

func main() {
	addr := ":8069"
	log.Printf("Serving web/ on http://localhost%s (use your LAN IP on a phone)", addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir("web"))))
}
