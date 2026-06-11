package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bluesky-o/fairshare/internal/middleware"
	"github.com/bluesky-o/fairshare/internal/models"
	"github.com/bluesky-o/fairshare/internal/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Test (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "status": "ok", "message": "This is test api" }` + "\n"))
}

func (h *UserHandler) Register (w http.ResponseWriter, r *http.Request) {
	firebaseUID := middleware.GetUserID(r)

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		
	}
}