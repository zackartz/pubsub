package main

import (
	"log"
	"net/http"

	"github.com/zackartz/pubsub/internal"
)

var addr = ":8080"

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./public")))
	mux.Handle("/ws", internal.GetRoom())

	log.Printf("listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
