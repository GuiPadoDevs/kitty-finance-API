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

type TransactionHandler struct {
	txService *services.TransactionService
}

func NewTransactionHandler(txService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{txService: txService}
}

// List lista os lançamentos financeiros com filtros
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	filter := models.TransactionFilter{
		CycleID:    r.URL.Query().Get("cycle_id"),
		CategoryID: r.URL.Query().Get("category_id"),
		Type:       r.URL.Query().Get("type"),
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			filter.Limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			filter.Offset = parsed
		}
	}

	transactions, err := h.txService.List(r.Context(), userID, filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, transactions)
}

// Create registra uma nova receita ou despesa
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Dados inválidos para criação do lançamento")
		return
	}

	created, err := h.txService.Create(r.Context(), userID, req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	msg := "Despesa registrada com sucesso! 🎀"
	if created.Type == "income" {
		msg = "Entrada registrada com sucesso! 💚"
	}

	RespondSuccess(w, http.StatusCreated, msg, created)
}

// GetByID busca um lançamento pelo ID
func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	if userID == "" || id == "" {
		RespondError(w, http.StatusBadRequest, "ID do lançamento inválido")
		return
	}

	tx, err := h.txService.GetByID(r.Context(), userID, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, tx)
}

// Update altera um lançamento existente
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	if userID == "" || id == "" {
		RespondError(w, http.StatusBadRequest, "ID do lançamento inválido")
		return
	}

	var req models.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Dados inválidos para edição do lançamento")
		return
	}

	updated, err := h.txService.Update(r.Context(), userID, id, req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondSuccess(w, http.StatusOK, "Lançamento atualizado com sucesso! 🌸", updated)
}

// Delete remove um lançamento
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	if userID == "" || id == "" {
		RespondError(w, http.StatusBadRequest, "ID do lançamento inválido")
		return
	}

	if err := h.txService.Delete(r.Context(), userID, id); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondSuccess(w, http.StatusOK, "Lançamento removido com sucesso! ✨", nil)
}
