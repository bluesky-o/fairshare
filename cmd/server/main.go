package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bluesky-o/fairshare/internal/config"
	"github.com/bluesky-o/fairshare/internal/database"
	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("hello world")

	cfg := config.Load()
	db, err := database.Connect(cfg.DatabasePath)

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close()

	err = db.RunMigrations("internal/database/migrations/001_initial_schema.sql")

	if err != nil {
		log.Fatalf("failed to run migration %v", err)
	}

	r := chi.NewRouter()

	r.Get("/health", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{ "status": "ok", "message": "FairShare API is running" }` + "\n"))
	})

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.AppPort),
		Handler: r,
		ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	log.Printf("Server started on port %s (env : %s)", cfg.AppPort, cfg.AppEnv)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start %v", err)
	}
}