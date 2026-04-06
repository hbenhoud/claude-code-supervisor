package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hbenhoud/claude-code-supervisor/internal/api"
	"github.com/hbenhoud/claude-code-supervisor/internal/store"
)

func main() {
	port := flag.Int("port", 3000, "Frontend port")
	apiPort := flag.Int("api-port", 3001, "API/WebSocket port")
	flag.Parse()

	db, err := store.Open("")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	router := api.NewRouter(db)

	go func() {
		addr := fmt.Sprintf(":%d", *apiPort)
		log.Printf("API server listening on %s", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("API server error: %v", err)
			os.Exit(1)
		}
	}()

	// TODO: Serve embedded frontend in production, proxy to Vite in dev
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Frontend server listening on %s", addr)
	log.Printf("Open http://localhost:%d to view the dashboard", *port)
	if err := http.ListenAndServe(addr, http.FileServer(http.Dir("web/dist"))); err != nil {
		log.Fatalf("Frontend server error: %v", err)
	}
}
