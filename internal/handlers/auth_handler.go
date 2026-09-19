package handlers

import (
	"encoding/json"
	"net/http"

	"financas-sah-api/internal/middleware"
	"financas-sah-api/internal/models"
	"financas-sah-api/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register lida com o cadastro de novas contas
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Dados inválidos enviados no corpo da requisição")
		return
	}

	res, err := h.authService.Register(r.Context(), req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondSuccess(w, http.StatusCreated, "Conta criada com sucesso! Bem-vinda ao Finanças Sah 🎀", res)
}

// Login realiza a autenticação do usuário
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Dados de login inválidos")
		return
	}

	res, err := h.authService.Login(r.Context(), req)
	if err != nil {
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	RespondSuccess(w, http.StatusOK, "Login realizado com sucesso! 🌸", res)
}

// GetMe retorna os dados do usuário autenticado pela sessão JWT
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	res, err := h.authService.GetMe(r.Context(), userID)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, res)
}

// UpdateSettings atualiza as configurações financeiras da usuária
func (h *AuthHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, "Não autorizado")
		return
	}

	var req models.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Parâmetros de configuração inválidos")
		return
	}

	updated, err := h.authService.UpdateSettings(r.Context(), userID, req)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, http.StatusOK, "Configurações atualizadas com sucesso! 🎀", updated)
}
