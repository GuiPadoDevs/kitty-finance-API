package handlers

import (
	"net/http"

	"financas-sah-api/internal/middleware"
	"financas-sah-api/internal/services"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetSummary retorna o resumo analítico completo do dashboard
func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	cycleID := r.URL.Query().Get("cycle_id")
	summary, err := h.dashboardService.GetSummary(r.Context(), userID, cycleID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, summary)
}
