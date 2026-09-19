package models

// RegisterRequest payload de cadastro
type RegisterRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	SalaryDay int    `json:"salary_day,omitempty"` // Opcional, padrão 5
}

// LoginRequest payload de login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse resposta de autenticação com token
type AuthResponse struct {
	Token        string        `json:"token"`
	User         User          `json:"user"`
	Settings     UserSettings  `json:"settings"`
	CurrentCycle *FinancialCycle `json:"current_cycle,omitempty"`
}

// UpdateSettingsRequest payload para atualizar configurações do usuário
type UpdateSettingsRequest struct {
	SalaryDay      *int    `json:"salary_day,omitempty"`
	AutoCarryOver  *bool   `json:"auto_carry_over,omitempty"`
	CurrencySymbol *string `json:"currency_symbol,omitempty"`
}

// CloseCycleRequest payload para fechar o mês/ciclo atual
type CloseCycleRequest struct {
	NewCycleName string  `json:"new_cycle_name"`          // Ex: "05/Nov a 04/Dez" ou "Novembro/2024"
	NewStartDate string  `json:"new_start_date"`          // YYYY-MM-DD
	CarryOver    *bool   `json:"carry_over,omitempty"`    // Se true, transfere o saldo restante
	Notes        *string `json:"notes,omitempty"`         // Observações do mês que encerrou
}

// CycleSummaryResponse resumo detalhado do ciclo
type CycleSummaryResponse struct {
	Cycle              FinancialCycle `json:"cycle"`
	OpeningBalance     float64        `json:"opening_balance"`
	TotalIncome        float64        `json:"total_income"`
	TotalExpense       float64        `json:"total_expense"`
	CurrentBalance     float64        `json:"current_balance"`
	SavingsRate        float64        `json:"savings_rate"` // % do total ganho que sobrou
	TotalTransactions  int            `json:"total_transactions"`
}
