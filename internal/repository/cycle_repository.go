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

type CycleRepository struct {
	db *database.DB
}

func NewCycleRepository(db *database.DB) *CycleRepository {
	return &CycleRepository{db: db}
}

// GetCurrentOpenCycle busca o ciclo atualmente ativo e calcula os totais em tempo real
func (r *CycleRepository) GetCurrentOpenCycle(ctx context.Context, userID string) (*models.CycleSummaryResponse, error) {
	queryCycle := `
		SELECT id, user_id, name, start_date::text, end_date::text, status, opening_balance,
		       total_income, total_expense, final_balance, closed_at, notes, created_at
		FROM financial_cycles
		WHERE user_id = $1 AND status = 'OPEN'
		ORDER BY created_at DESC
		LIMIT 1;
	`

	var cycle models.FinancialCycle
	err := r.db.Pool.QueryRow(ctx, queryCycle, userID).Scan(
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
		&cycle.ClosedAt,
		&cycle.Notes,
		&cycle.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar ciclo atual: %w", err)
	}

	// Calcula métricas ao vivo a partir das transações cadastradas neste ciclo
	queryTotals := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expense,
			COUNT(*) as total_transactions
		FROM transactions
		WHERE user_id = $1 AND cycle_id = $2;
	`

	var income, expense float64
	var totalTx int
	err = r.db.Pool.QueryRow(ctx, queryTotals, userID, cycle.ID).Scan(&income, &expense, &totalTx)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular totais do ciclo: %w", err)
	}

	currentBalance := cycle.OpeningBalance + income - expense

	var savingsRate float64
	totalAvailable := cycle.OpeningBalance + income
	if totalAvailable > 0 {
		savingsRate = (currentBalance / totalAvailable) * 100
		if savingsRate < 0 {
			savingsRate = 0
		}
	}

	return &models.CycleSummaryResponse{
		Cycle:             cycle,
		OpeningBalance:    cycle.OpeningBalance,
		TotalIncome:       income,
		TotalExpense:      expense,
		CurrentBalance:    currentBalance,
		SavingsRate:       savingsRate,
		TotalTransactions: totalTx,
	}, nil
}

// ListCycles lista todos os ciclos ordenados decrescentemente
func (r *CycleRepository) ListCycles(ctx context.Context, userID string, limit, offset int) ([]models.FinancialCycle, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, user_id, name, start_date::text, end_date::text, status, opening_balance,
		       total_income, total_expense, final_balance, closed_at, notes, created_at
		FROM financial_cycles
		WHERE user_id = $1
		ORDER BY start_date DESC, created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar ciclos: %w", err)
	}
	defer rows.Close()

	var cycles []models.FinancialCycle
	for rows.Next() {
		var c models.FinancialCycle
		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Name,
			&c.StartDate,
			&c.EndDate,
			&c.Status,
			&c.OpeningBalance,
			&c.TotalIncome,
			&c.TotalExpense,
			&c.FinalBalance,
			&c.ClosedAt,
			&c.Notes,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha do ciclo: %w", err)
		}
		cycles = append(cycles, c)
	}

	return cycles, nil
}

// GetCycleByID busca um ciclo específico
func (r *CycleRepository) GetCycleByID(ctx context.Context, userID, cycleID string) (*models.FinancialCycle, error) {
	query := `
		SELECT id, user_id, name, start_date::text, end_date::text, status, opening_balance,
		       total_income, total_expense, final_balance, closed_at, notes, created_at
		FROM financial_cycles
		WHERE id = $1 AND user_id = $2;
	`

	var c models.FinancialCycle
	err := r.db.Pool.QueryRow(ctx, query, cycleID, userID).Scan(
		&c.ID,
		&c.UserID,
		&c.Name,
		&c.StartDate,
		&c.EndDate,
		&c.Status,
		&c.OpeningBalance,
		&c.TotalIncome,
		&c.TotalExpense,
		&c.FinalBalance,
		&c.ClosedAt,
		&c.Notes,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar ciclo por ID: %w", err)
	}

	return &c, nil
}

// CloseCycleAndOpenNew realiza o fechamento manual do mês e abre um novo ciclo
func (r *CycleRepository) CloseCycleAndOpenNew(
	ctx context.Context,
	userID string,
	req models.CloseCycleRequest,
) (*models.FinancialCycle, *models.FinancialCycle, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Busca o ciclo aberto atual
	queryCurrent := `
		SELECT id, opening_balance
		FROM financial_cycles
		WHERE user_id = $1 AND status = 'OPEN'
		FOR UPDATE;
	`
	var currentID string
	var openingBalance float64
	err = tx.QueryRow(ctx, queryCurrent, userID).Scan(&currentID, &openingBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, errors.New("nenhum ciclo financeiro aberto encontrado para fechar")
		}
		return nil, nil, fmt.Errorf("erro ao buscar ciclo aberto: %w", err)
	}

	// 2. Calcula os totais finais de transações
	queryTotals := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1 AND cycle_id = $2;
	`
	var totalIncome, totalExpense float64
	err = tx.QueryRow(ctx, queryTotals, userID, currentID).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao calcular totais para fechamento: %w", err)
	}

	finalBalance := openingBalance + totalIncome - totalExpense
	endDate := time.Now().Format("2006-01-02")

	// 3. Atualiza e congela o ciclo atual como CLOSED
	queryClose := `
		UPDATE financial_cycles
		SET status = 'CLOSED',
		    end_date = $1,
		    total_income = $2,
		    total_expense = $3,
		    final_balance = $4,
		    closed_at = NOW(),
		    notes = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, name, start_date::text, end_date::text, status, opening_balance,
		          total_income, total_expense, final_balance, closed_at, notes, created_at;
	`

	var closedCycle models.FinancialCycle
	err = tx.QueryRow(ctx, queryClose, endDate, totalIncome, totalExpense, finalBalance, req.Notes, currentID, userID).Scan(
		&closedCycle.ID,
		&closedCycle.UserID,
		&closedCycle.Name,
		&closedCycle.StartDate,
		&closedCycle.EndDate,
		&closedCycle.Status,
		&closedCycle.OpeningBalance,
		&closedCycle.TotalIncome,
		&closedCycle.TotalExpense,
		&closedCycle.FinalBalance,
		&closedCycle.ClosedAt,
		&closedCycle.Notes,
		&closedCycle.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao fechar ciclo atual: %w", err)
	}

	// 4. Determina o saldo de abertura do novo ciclo
	newOpeningBalance := 0.00
	if req.CarryOver != nil && *req.CarryOver {
		newOpeningBalance = finalBalance
	}

	newStartDate := req.NewStartDate
	if newStartDate == "" {
		newStartDate = time.Now().Format("2006-01-02")
	}

	newCycleName := req.NewCycleName
	if newCycleName == "" {
		newCycleName = fmt.Sprintf("Ciclo %s 🎀", time.Now().Format("Jan/2006"))
	}

	// 5. Cria o novo ciclo com status 'OPEN'
	queryNew := `
		INSERT INTO financial_cycles (user_id, name, start_date, status, opening_balance)
		VALUES ($1, $2, $3, 'OPEN', $4)
		RETURNING id, user_id, name, start_date::text, end_date::text, status, opening_balance,
		          total_income, total_expense, final_balance, closed_at, notes, created_at;
	`

	var newCycle models.FinancialCycle
	err = tx.QueryRow(ctx, queryNew, userID, newCycleName, newStartDate, newOpeningBalance).Scan(
		&newCycle.ID,
		&newCycle.UserID,
		&newCycle.Name,
		&newCycle.StartDate,
		&newCycle.EndDate,
		&newCycle.Status,
		&newCycle.OpeningBalance,
		&newCycle.TotalIncome,
		&newCycle.TotalExpense,
		&newCycle.FinalBalance,
		&newCycle.ClosedAt,
		&newCycle.Notes,
		&newCycle.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao abrir novo ciclo: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("erro ao confirmar fechamento e abertura de ciclo: %w", err)
	}

	return &closedCycle, &newCycle, nil
}
