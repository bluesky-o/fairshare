package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bluesky-o/fairshare/pkg/firebase"
	"github.com/go-chi/chi/v5"

	"github.com/bluesky-o/fairshare/internal/config"
	"github.com/bluesky-o/fairshare/internal/database"
	"github.com/bluesky-o/fairshare/internal/handlers"
	authmiddleware "github.com/bluesky-o/fairshare/internal/middleware"
	"github.com/bluesky-o/fairshare/internal/repository"
	"github.com/bluesky-o/fairshare/internal/services"
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

	firebaseClient, err := firebase.NewClient(cfg.FirebaseServiceAccountPath)

	if err != nil {
		log.Fatalf("failed to initilize firebase: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	groupRepo := repository.NewGroupRepository(db)
	groupService := services.NewGroupService(groupRepo, userRepo)
	groupHandler := handlers.NewGroupHandler(groupService)
	expenseRepo := repository.NewExpenseRepository(db) 
	expenseService := services.NewExpenseService(expenseRepo, groupRepo)
	expenseHandler := handlers.NewExpenseHandler(expenseService) 

	r := chi.NewRouter()

	r.Get("/health", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{ "status": "ok", "message": "FairShare API is running" }` + "\n"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authmiddleware.Authenticate(firebaseClient))
		r.Get("/user/test", userHandler.Test)
		r.Post("/auth/register", userHandler.Register)
		r.Get("/users/me", userHandler.GetMe)
		r.Put("/users/me", userHandler.UpdateMe)
		r.Get("/users/search", userHandler.FindByEmail)
		r.Post("/groups", groupHandler.CreateGroup)
		r.Get("/groups", groupHandler.GetMyGroups)
		r.Get("/groups/{id}", groupHandler.GetGroup)
		r.Put("/groups/{id}", groupHandler.UpdateGroup)
		r.Delete("/groups/{id}", groupHandler.DeleteGroup)
		r.Post("/groups/{id}/members", groupHandler.AddMember)
		r.Delete("/groups/{id}/members/{uid}", groupHandler.RemoveMember)
		r.Post("/groups/{id}/expenses", expenseHandler.CreateExpense)
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