package repository

import (
	"context"
	"errors"
	"fmt"

	"financas-sah-api/internal/database"
	"financas-sah-api/internal/models"
)

type DashboardRepository struct {
	db        *database.DB
	cycleRepo *CycleRepository
	txRepo    *TransactionRepository
}

func NewDashboardRepository(db *database.DB, cycleRepo *CycleRepository, txRepo *TransactionRepository) *DashboardRepository {
	return &DashboardRepository{
		db:        db,
		cycleRepo: cycleRepo,
		txRepo:    txRepo,
	}
}

// GetSummary monta o resumo financeiro completo para um ciclo específico ou o ciclo atual aberto
func (r *DashboardRepository) GetSummary(ctx context.Context, userID, cycleID string) (*models.DashboardSummaryResponse, error) {
	var cycle *models.FinancialCycle
	var err error

	// 1. Identifica o ciclo correto
	if cycleID != "" {
		cycle, err = r.cycleRepo.GetCycleByID(ctx, userID, cycleID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar ciclo especificado: %w", err)
		}
	} else {
		cycleSummary, err := r.cycleRepo.GetCurrentOpenCycle(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar ciclo atual: %w", err)
		}
		if cycleSummary != nil {
			cycle = &cycleSummary.Cycle
		}
	}

	if cycle == nil {
		return nil, errors.New("nenhum ciclo financeiro encontrado para gerar o relatório")
	}

	// 2. Calcula as somas totais do ciclo
	queryTotals := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expense
		FROM transactions
		WHERE user_id = $1 AND cycle_id = $2;
	`
	var totalIncome, totalExpense float64
	err = r.db.Pool.QueryRow(ctx, queryTotals, userID, cycle.ID).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular totais do dashboard: %w", err)
	}

	openingBalance := cycle.OpeningBalance
	currentBalance := openingBalance + totalIncome - totalExpense

	var savingsRate float64
	totalAvailable := openingBalance + totalIncome
	if totalAvailable > 0 {
		savingsRate = (currentBalance / totalAvailable) * 100
		if savingsRate < 0 {
			savingsRate = 0
		}
	}

	// 3. Agrupamento detalhado por categorias (Breakdown)
	queryBreakdown := `
		SELECT 
			c.id as category_id,
			c.name,
			c.type,
			c.icon,
			c.color,
			c.budget_limit,
			COALESCE(SUM(t.amount), 0) as total_amount,
			COUNT(t.id) as transaction_count
		FROM categories c
		LEFT JOIN transactions t ON t.category_id = c.id AND t.cycle_id = $2 AND t.user_id = $1
		WHERE c.user_id = $1
		GROUP BY c.id, c.name, c.type, c.icon, c.color, c.budget_limit
		ORDER BY total_amount DESC, c.name ASC;
	`

	rows, err := r.db.Pool.Query(ctx, queryBreakdown, userID, cycle.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar divisão de categorias: %w", err)
	}
	defer rows.Close()

	var incomeBreakdown []models.CategoryBreakdown
	var expenseBreakdown []models.CategoryBreakdown

	for rows.Next() {
		var b models.CategoryBreakdown
		err := rows.Scan(
			&b.CategoryID,
			&b.Name,
			&b.Type,
			&b.Icon,
			&b.Color,
			&b.BudgetLimit,
			&b.TotalAmount,
			&b.TransactionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler breakdown de categoria: %w", err)
		}

		// Calcula porcentagem do grupo
		if b.Type == "expense" {
			if totalExpense > 0 {
				b.PercentageOfTotal = (b.TotalAmount / totalExpense) * 100
			}
			if b.BudgetLimit > 0 {
				b.BudgetUsagePercent = (b.TotalAmount / b.BudgetLimit) * 100
				b.BudgetExceeded = b.TotalAmount > b.BudgetLimit
			}
			expenseBreakdown = append(expenseBreakdown, b)
		} else if b.Type == "income" {
			if totalIncome > 0 {
				b.PercentageOfTotal = (b.TotalAmount / totalIncome) * 100
			}
			incomeBreakdown = append(incomeBreakdown, b)
		}
	}

	// 4. Últimas 10 transações
	recentTx, err := r.txRepo.List(ctx, userID, models.TransactionFilter{
		CycleID: cycle.ID,
		Limit:   10,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transações recentes: %w", err)
	}

	return &models.DashboardSummaryResponse{
		Cycle:              *cycle,
		OpeningBalance:     openingBalance,
		TotalIncome:        totalIncome,
		TotalExpense:       totalExpense,
		CurrentBalance:     currentBalance,
		SavingsRate:        savingsRate,
		IncomeBreakdown:    incomeBreakdown,
		ExpenseBreakdown:   expenseBreakdown,
		RecentTransactions: recentTx,
	}, nil
}
