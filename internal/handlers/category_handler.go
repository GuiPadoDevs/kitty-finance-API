package handlers

import (
	"encoding/json"
	"net/http"

	"financas-sah-api/internal/middleware"
	"financas-sah-api/internal/models"
	"financas-sah-api/internal/services"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List lista as categorias do usuário
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	catType := r.URL.Query().Get("type")
	categories, err := h.categoryService.List(r.Context(), userID, catType)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, categories)
}

// Create cadastra uma nova categoria
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Dados inválidos para criação da categoria")
		return
	}

	created, err := h.categoryService.Create(r.Context(), userID, req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondSuccess(w, http.StatusCreated, "Tópico cadastrado com sucesso! 🎀", created)
}

// Update altera uma categoria existente
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	if userID == "" || id == "" {
		RespondError(w, http.StatusBadRequest, "ID da categoria inválido")
		return
	}

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Dados inválidos para edição da categoria")
		return
	}

	updated, err := h.categoryService.Update(r.Context(), userID, id, req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondSuccess(w, http.StatusOK, "Tópico atualizado com sucesso! 🌸", updated)
}

// Delete remove uma categoria
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	if userID == "" || id == "" {
		RespondError(w, http.StatusBadRequest, "ID da categoria inválido")
		return
	}

	if err := h.categoryService.Delete(r.Context(), userID, id); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondSuccess(w, http.StatusOK, "Categoria removida com sucesso! ✨", nil)
}
