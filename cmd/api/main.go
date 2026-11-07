package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go-prod-project/internal/config"
	"go-prod-project/internal/db"
	"go-prod-project/internal/handlers"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.MustLoad()
	db, err := db.Connect(cfg.DatabaseUrl)

	if err != nil {
		log.Fatalf("main.db.connect: %v", err)
	}

	fmt.Println("database connected")
	fmt.Println("Starting this server...")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handlers.Health)
	mux.HandleFunc("GET /listings", handlers.List(db))
	mux.HandleFunc("DELETE /listings/{id}", handlers.DeleteListing(db))

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
