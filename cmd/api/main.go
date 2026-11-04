package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go-prod-project/internal/config"
	"go-prod-project/internal/handlers"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println("Starting this server...")
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handlers.Health)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
