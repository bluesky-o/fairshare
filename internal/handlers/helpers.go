package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bluesky-o/fairshare/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.APIResponse{
		Success: false,
		Error: message,
	})
}

func writeSuccess(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.APIResponse{
		Success: true,
		Data: message,
	})
}