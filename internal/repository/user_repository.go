package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"financas-sah-api/internal/database"
	"financas-sah-api/internal/models"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser cria um novo usuário
func (r *UserRepository) CreateUser(ctx context.Context, tx pgx.Tx, user *models.User) error {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`
	return tx.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
}

// GetByEmail busca o usuário pelo email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = $1;
	`
	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar usuário por email: %w", err)
	}
	return &user, nil
}

// GetByID busca o usuário pelo ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE id = $1;
	`
	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar usuário por ID: %w", err)
	}
	return &user, nil
}

// CreateSettings cria as configurações iniciais do usuário
func (r *UserRepository) CreateSettings(ctx context.Context, tx pgx.Tx, settings *models.UserSettings) error {
	query := `
		INSERT INTO user_settings (user_id, salary_day, auto_carry_over, currency_symbol)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at;
	`
	return tx.QueryRow(
		ctx,
		query,
		settings.UserID,
		settings.SalaryDay,
		settings.AutoCarryOver,
		settings.CurrencySymbol,
	).Scan(&settings.CreatedAt, &settings.UpdatedAt)
}

// GetSettings busca as configurações do usuário
func (r *UserRepository) GetSettings(ctx context.Context, userID string) (*models.UserSettings, error) {
	query := `
		SELECT user_id, salary_day, auto_carry_over, currency_symbol, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1;
	`
	var s models.UserSettings
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&s.UserID,
		&s.SalaryDay,
		&s.AutoCarryOver,
		&s.CurrencySymbol,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar configurações: %w", err)
	}
	return &s, nil
}

// UpdateSettings atualiza as preferências do usuário
func (r *UserRepository) UpdateSettings(ctx context.Context, s *models.UserSettings) error {
	query := `
		UPDATE user_settings
		SET salary_day = $1, auto_carry_over = $2, currency_symbol = $3, updated_at = NOW()
		WHERE user_id = $4
		RETURNING updated_at;
	`
	return r.db.Pool.QueryRow(ctx, query, s.SalaryDay, s.AutoCarryOver, s.CurrencySymbol, s.UserID).Scan(&s.UpdatedAt)
}

// CreateDefaultCategories cria tópicos/categorias padrões carinhosos da Hello Kitty
func (r *UserRepository) CreateDefaultCategories(ctx context.Context, tx pgx.Tx, userID string) error {
	defaultCategories := []struct {
		Name        string
		Type        string
		Icon        string
		Color       string
		BudgetLimit float64
	}{
		// Entradas
		{"Salário 💼", "income", "briefcase", "#34D399", 0.00},
		{"Renda Extra / Mimos 🎁", "income", "sparkles", "#6EE7B7", 0.00},

		// Saídas
		{"Cartão de Crédito 💳", "expense", "credit-card", "#FF1E40", 0.00},
		{"Maquiagem & Skincare 💄", "expense", "sparkles", "#FF85A2", 0.00},
		{"Mercado & Comidinhas 🍓", "expense", "shopping-cart", "#F472B6", 0.00},
		{"Lazer & Rolês 🎀", "expense", "heart", "#C084FC", 0.00},
		{"Mimos & Comprinhas 🛍️", "expense", "gift", "#FB7185", 0.00},
		{"Contas Fixas 🏠", "expense", "home", "#FB923C", 0.00},
	}

	query := `
		INSERT INTO categories (user_id, name, type, icon, color, budget_limit)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	for _, cat := range defaultCategories {
		if _, err := tx.Exec(ctx, query, userID, cat.Name, cat.Type, cat.Icon, cat.Color, cat.BudgetLimit); err != nil {
			return fmt.Errorf("erro ao criar categoria padrão %s: %w", cat.Name, err)
		}
	}

	return nil
}

// CreateInitialCycle cria o primeiro ciclo financeiro aberto para a usuária
func (r *UserRepository) CreateInitialCycle(ctx context.Context, tx pgx.Tx, userID string, salaryDay int) (*models.FinancialCycle, error) {
	now := time.Now()
	startDate := now.Format("2006-01-02")
	cycleName := fmt.Sprintf("Ciclo Inicial 🎀 (%s)", now.Format("Jan/2006"))

	query := `
		INSERT INTO financial_cycles (user_id, name, start_date, status, opening_balance)
		VALUES ($1, $2, $3, 'OPEN', 0.00)
		RETURNING id, user_id, name, start_date::text, end_date::text, status, opening_balance, total_income, total_expense, final_balance, created_at;
	`

	var cycle models.FinancialCycle
	err := tx.QueryRow(ctx, query, userID, cycleName, startDate).Scan(
		&cycle.ID,
		&cycle.UserID,
		&cycle.Name,
		&cycle.StartDate,
		&cycle.EndDate,
		&cycle.Status,
		&cycle.OpeningBalance,
		&cycle.TotalIncome,
		&cycle.TotalExpense,
		&cycle.FinalBalance,
		&cycle.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar ciclo inicial: %w", err)
	}

	return &cycle, nil
}

// DB retorna a instância do banco para transações
func (r *UserRepository) DB() *database.DB {
	return r.db
}
