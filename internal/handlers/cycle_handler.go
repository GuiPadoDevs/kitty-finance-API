package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"financas-sah-api/internal/middleware"
	"financas-sah-api/internal/models"
	"financas-sah-api/internal/services"

	"github.com/go-chi/chi/v5"
)

type CycleHandler struct {
	cycleService *services.CycleService
}

func NewCycleHandler(cycleService *services.CycleService) *CycleHandler {
	return &CycleHandler{cycleService: cycleService}
}

// GetCurrentCycle retorna o resumo do ciclo aberto atual
func (h *CycleHandler) GetCurrentCycle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	summary, err := h.cycleService.GetCurrentCycle(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, summary)
}

// ListCycles retorna a lista de todos os ciclos (histórico)
func (h *CycleHandler) ListCycles(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	cycles, err := h.cycleService.ListCycles(r.Context(), userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, cycles)
}

// GetCycleByID retorna os detalhes de um ciclo específico
func (h *CycleHandler) GetCycleByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	cycleID := chi.URLParam(r, "id")

	if userID == "" || cycleID == "" {
		RespondError(w, http.StatusBadRequest, "ID do ciclo não informado")
		return
	}

	cycle, err := h.cycleService.GetCycleByID(r.Context(), userID, cycleID)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, cycle)
}

// CloseCycle fecha o mês/ciclo atual e abre o novo
func (h *CycleHandler) CloseCycle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	var req models.CloseCycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Parâmetros de fechamento de ciclo inválidos")
		return
	}

	res, err := h.cycleService.CloseCycle(r.Context(), userID, req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, res)
}
