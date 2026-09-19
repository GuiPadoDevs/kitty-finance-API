package models

import (
	"time"
)

// User representa a conta de acesso da usuária
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserSettings armazena preferências de cálculo financeiro
type UserSettings struct {
	UserID         string    `json:"user_id"`
	SalaryDay      int       `json:"salary_day"`
	AutoCarryOver  bool      `json:"auto_carry_over"`
	CurrencySymbol string    `json:"currency_symbol"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// FinancialCycle representa um mês financeiro customizado ou fechamento manual
type FinancialCycle struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Name           string     `json:"name"`
	StartDate      string     `json:"start_date"` // YYYY-MM-DD
	EndDate        *string    `json:"end_date,omitempty"`
	Status         string     `json:"status"` // OPEN, CLOSED
	OpeningBalance float64    `json:"opening_balance"`
	TotalIncome    float64    `json:"total_income"`
	TotalExpense   float64    `json:"total_expense"`
	FinalBalance   float64    `json:"final_balance"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Category representa um tópico/categoria customizável (ex: Maquiagem, Salário, Cartão)
type Category struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // income, expense
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	BudgetLimit float64   `json:"budget_limit"`
	CreatedAt   time.Time `json:"created_at"`
}

// Transaction representa uma entrada ou saída de dinheiro
type Transaction struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	CycleID       string    `json:"cycle_id"`
	CategoryID    string    `json:"category_id"`
	CategoryName  string    `json:"category_name,omitempty"`
	CategoryIcon  string    `json:"category_icon,omitempty"`
	CategoryColor string    `json:"category_color,omitempty"`
	Title         string    `json:"title"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"` // income, expense
	Date          string    `json:"date"` // YYYY-MM-DD
	PaymentMethod string    `json:"payment_method"`
	Notes         *string   `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// HealthResponse estrutura da resposta de healthcheck
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Database  string            `json:"database"`
	Services  map[string]string `json:"services"`
}
