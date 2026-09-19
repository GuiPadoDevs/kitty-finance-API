package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"financas-sah-api/internal/auth"
	"financas-sah-api/internal/models"
	"financas-sah-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	cycleRepo  *repository.CycleRepository
	jwtManager *auth.JWTManager
}

func NewAuthService(
	userRepo *repository.UserRepository,
	cycleRepo *repository.CycleRepository,
	jwtManager *auth.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		cycleRepo:  cycleRepo,
		jwtManager: jwtManager,
	}
}

// Register cria nova conta com categorias de boas-vindas e primeiro ciclo aberto
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthResponse, error) {
	// Validações
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || req.Email == "" || len(req.Password) < 6 {
		return nil, errors.New("nome, email válido e senha com no mínimo 6 caracteres são obrigatórios")
	}

	// Verifica se email já existe
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar email: %w", err)
	}
	if existing != nil {
		return nil, errors.New("já existe uma conta cadastrada com este e-mail")
	}

	// Gera hash bcrypt da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("falha ao criptografar senha: %w", err)
	}

	// Inicia transação no banco
	tx, err := s.userRepo.DB().Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir transação: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Cria usuário
	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}
	if err := s.userRepo.CreateUser(ctx, tx, &user); err != nil {
		return nil, fmt.Errorf("falha ao criar usuário: %w", err)
	}

	// 2. Cria configurações financeiras padrão
	salaryDay := 5
	if req.SalaryDay >= 1 && req.SalaryDay <= 31 {
		salaryDay = req.SalaryDay
	}

	settings := models.UserSettings{
		UserID:         user.ID,
		SalaryDay:      salaryDay,
		AutoCarryOver:  true,
		CurrencySymbol: "R$",
	}
	if err := s.userRepo.CreateSettings(ctx, tx, &settings); err != nil {
		return nil, fmt.Errorf("falha ao salvar configurações: %w", err)
	}

	// 3. Cria categorias fofas padrão da Hello Kitty
	if err := s.userRepo.CreateDefaultCategories(ctx, tx, user.ID); err != nil {
		return nil, fmt.Errorf("falha ao inicializar categorias: %w", err)
	}

	// 4. Cria primeiro ciclo financeiro aberto
	cycle, err := s.userRepo.CreateInitialCycle(ctx, tx, user.ID, salaryDay)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar ciclo inicial: %w", err)
	}

	// Confirma transação
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("falha ao confirmar registro: %w", err)
	}

	// 5. Gera token JWT
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar token de autenticação: %w", err)
	}

	return &models.AuthResponse{
		Token:        token,
		User:         user,
		Settings:     settings,
		CurrentCycle: cycle,
	}, nil
}

// Login autentica a usuária e retorna JWT
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("e-mail e senha são obrigatórios")
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	if user == nil {
		return nil, errors.New("e-mail ou senha incorretos")
	}

	// Valida senha
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("e-mail ou senha incorretos")
	}

	// Busca configurações
	settings, err := s.userRepo.GetSettings(ctx, user.ID)
	if err != nil || settings == nil {
		settings = &models.UserSettings{
			UserID:         user.ID,
			SalaryDay:      5,
			AutoCarryOver:  true,
			CurrencySymbol: "R$",
		}
	}

	// Busca ciclo aberto atual
	cycleSummary, _ := s.cycleRepo.GetCurrentOpenCycle(ctx, user.ID)
	var currentCycle *models.FinancialCycle
	if cycleSummary != nil {
		currentCycle = &cycleSummary.Cycle
	}

	// Gera token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar token: %w", err)
	}

	return &models.AuthResponse{
		Token:        token,
		User:         *user,
		Settings:     *settings,
		CurrentCycle: currentCycle,
	}, nil
}

// GetMe retorna dados do perfil, configurações e ciclo atual
func (s *AuthService) GetMe(ctx context.Context, userID string) (*models.AuthResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("usuário não encontrado")
	}

	settings, err := s.userRepo.GetSettings(ctx, userID)
	if err != nil || settings == nil {
		settings = &models.UserSettings{
			UserID:         user.ID,
			SalaryDay:      5,
			AutoCarryOver:  true,
			CurrencySymbol: "R$",
		}
	}

	cycleSummary, _ := s.cycleRepo.GetCurrentOpenCycle(ctx, userID)
	var currentCycle *models.FinancialCycle
	if cycleSummary != nil {
		currentCycle = &cycleSummary.Cycle
	}

	return &models.AuthResponse{
		User:         *user,
		Settings:     *settings,
		CurrentCycle: currentCycle,
	}, nil
}

// UpdateSettings atualiza as configurações do usuário
func (s *AuthService) UpdateSettings(ctx context.Context, userID string, req models.UpdateSettingsRequest) (*models.UserSettings, error) {
	settings, err := s.userRepo.GetSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar configurações: %w", err)
	}
	if settings == nil {
		settings = &models.UserSettings{
			UserID:         userID,
			SalaryDay:      5,
			AutoCarryOver:  true,
			CurrencySymbol: "R$",
		}
	}

	if req.SalaryDay != nil && *req.SalaryDay >= 1 && *req.SalaryDay <= 31 {
		settings.SalaryDay = *req.SalaryDay
	}
	if req.AutoCarryOver != nil {
		settings.AutoCarryOver = *req.AutoCarryOver
	}
	if req.CurrencySymbol != nil && strings.TrimSpace(*req.CurrencySymbol) != "" {
		settings.CurrencySymbol = strings.TrimSpace(*req.CurrencySymbol)
	}

	if err := s.userRepo.UpdateSettings(ctx, settings); err != nil {
		return nil, fmt.Errorf("falha ao salvar preferências: %w", err)
	}

	return settings, nil
}
