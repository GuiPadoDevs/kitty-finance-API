package models

// --- CATEGORY DTOs ---

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"` // income, expense
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	BudgetLimit float64 `json:"budget_limit"`
}

type UpdateCategoryRequest struct {
	Name        string  `json:"name"`
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	BudgetLimit float64 `json:"budget_limit"`
}

// --- TRANSACTION DTOs ---

type CreateTransactionRequest struct {
	CycleID       *string `json:"cycle_id,omitempty"` // Se omitido, vincula ao ciclo atual aberto
	CategoryID    string  `json:"category_id"`
	Title         string  `json:"title"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"` // income, expense
	Date          string  `json:"date"` // YYYY-MM-DD
	PaymentMethod string  `json:"payment_method"`
	Notes         *string `json:"notes,omitempty"`
}

type UpdateTransactionRequest struct {
	CategoryID    string  `json:"category_id"`
	Title         string  `json:"title"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	Date          string  `json:"date"`
	PaymentMethod string  `json:"payment_method"`
	Notes         *string `json:"notes,omitempty"`
}

type TransactionFilter struct {
	CycleID    string
	CategoryID string
	Type       string
	Limit      int
	Offset     int
}

// --- DASHBOARD & ANALYTICS DTOs ---

type CategoryBreakdown struct {
	CategoryID         string  `json:"category_id"`
	Name               string  `json:"name"`
	Type               string  `json:"type"`
	Icon               string  `json:"icon"`
	Color              string  `json:"color"`
	BudgetLimit        float64 `json:"budget_limit"`
	TotalAmount        float64 `json:"total_amount"`
	PercentageOfTotal  float64 `json:"percentage_of_total"`  // % do total de despesas (ou receitas)
	BudgetExceeded     bool    `json:"budget_exceeded"`      // true se ultrapassou o limite
	BudgetUsagePercent float64 `json:"budget_usage_percent"` // % do limite utilizado
	TransactionCount   int     `json:"transaction_count"`
}

type DashboardSummaryResponse struct {
	Cycle              FinancialCycle      `json:"cycle"`
	OpeningBalance     float64             `json:"opening_balance"`
	TotalIncome        float64             `json:"total_income"`
	TotalExpense       float64             `json:"total_expense"`
	CurrentBalance     float64             `json:"current_balance"`
	SavingsRate        float64             `json:"savings_rate"`
	IncomeBreakdown    []CategoryBreakdown `json:"income_breakdown"`
	ExpenseBreakdown   []CategoryBreakdown `json:"expense_breakdown"`
	RecentTransactions []Transaction       `json:"recent_transactions"`
}
