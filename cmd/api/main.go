package main

import (
	"log"
	"net/http"
	"os"

	"youtube-music-video-api/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9898"
	}
	
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", handlers.HealthHandler)
	
	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}