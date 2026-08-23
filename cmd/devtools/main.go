package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jaredwarren/SubGame/internal/devtools"
)

func main() {
	addr := flag.String("addr", "localhost:8088", "HTTP listen address")
	flag.Parse()

	log.Printf("SubGame devtools hub: http://%s", *addr)
	if err := http.ListenAndServe(*addr, devtools.Handler()); err != nil {
		log.Fatal(err)
	}
}
