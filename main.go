package main

import (
	"log"
	"net/http"
)

const (
	address   = ":8080"
	staticDir = "static"
)

func main() {
	log.Printf("serving %s at http://localhost%s", staticDir, address)
	if err := http.ListenAndServe(address, newHandler(staticDir)); err != nil {
		log.Fatal(err)
	}
}

func newHandler(dir string) http.Handler {
	return http.FileServer(http.Dir(dir))
}
