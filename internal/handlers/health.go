package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"financas-sah-api/internal/database"
	"financas-sah-api/internal/models"
)

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "healthy 🌸"
	if err := h.db.Pool.Ping(ctx); err != nil {
		dbStatus = "unreachable ⚠️: " + err.Error()
	}

	response := models.HealthResponse{
		Status:    "online 🎀",
		Timestamp: time.Now().Format(time.RFC3339),
		Database:  dbStatus,
		Services: map[string]string{
			"api":      "running",
			"database": dbStatus,
			"theme":    "Hello Kitty Edition 🎀",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if dbStatus != "healthy 🌸" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(response)
}
