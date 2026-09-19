package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"financas-sah-api/internal/database"
	"financas-sah-api/internal/models"

	"github.com/jackc/pgx/v5"
)

type TransactionRepository struct {
	db *database.DB
}

func NewTransactionRepository(db *database.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create insere uma nova transação
func (r *TransactionRepository) Create(ctx context.Context, t *models.Transaction) error {
	query := `
		INSERT INTO transactions (user_id, cycle_id, category_id, title, amount, type, date, payment_method, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at;
	`
	return r.db.Pool.QueryRow(
		ctx,
		query,
		t.UserID,
		t.CycleID,
		t.CategoryID,
		t.Title,
		t.Amount,
		t.Type,
		t.Date,
		t.PaymentMethod,
		t.Notes,
	).Scan(&t.ID, &t.CreatedAt)
}

// List busca transações com filtros dinâmicos e dados da categoria com JOIN
func (r *TransactionRepository) List(ctx context.Context, userID string, filter models.TransactionFilter) ([]models.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.cycle_id, t.category_id,
			c.name as category_name, c.icon as category_icon, c.color as category_color,
			t.title, t.amount, t.type, t.date::text, t.payment_method, t.notes, t.created_at
		FROM transactions t
		INNER JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1
	`
	args := []interface{}{userID}
	paramIdx := 2

	if filter.CycleID != "" {
		query += " AND t.cycle_id = $" + strconv.Itoa(paramIdx)
		args = append(args, filter.CycleID)
		paramIdx++
	}

	if filter.CategoryID != "" {
		query += " AND t.category_id = $" + strconv.Itoa(paramIdx)
		args = append(args, filter.CategoryID)
		paramIdx++
	}

	if filter.Type != "" {
		query += " AND t.type = $" + strconv.Itoa(paramIdx)
		args = append(args, filter.Type)
		paramIdx++
	}

	query += " ORDER BY t.date DESC, t.created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT $" + strconv.Itoa(paramIdx)
		args = append(args, filter.Limit)
		paramIdx++
	} else {
		query += " LIMIT 100"
	}

	if filter.Offset > 0 {
		query += " OFFSET $" + strconv.Itoa(paramIdx)
		args = append(args, filter.Offset)
		paramIdx++
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar transações: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.CycleID,
			&t.CategoryID,
			&t.CategoryName,
			&t.CategoryIcon,
			&t.CategoryColor,
			&t.Title,
			&t.Amount,
			&t.Type,
			&t.Date,
			&t.PaymentMethod,
			&t.Notes,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler transação: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetByID busca uma transação específica com dados da categoria
func (r *TransactionRepository) GetByID(ctx context.Context, userID, id string) (*models.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.cycle_id, t.category_id,
			c.name as category_name, c.icon as category_icon, c.color as category_color,
			t.title, t.amount, t.type, t.date::text, t.payment_method, t.notes, t.created_at
		FROM transactions t
		INNER JOIN categories c ON t.category_id = c.id
		WHERE t.id = $1 AND t.user_id = $2;
	`
	var t models.Transaction
	err := r.db.Pool.QueryRow(ctx, query, id, userID).Scan(
		&t.ID,
		&t.UserID,
		&t.CycleID,
		&t.CategoryID,
		&t.CategoryName,
		&t.CategoryIcon,
		&t.CategoryColor,
		&t.Title,
		&t.Amount,
		&t.Type,
		&t.Date,
		&t.PaymentMethod,
		&t.Notes,
		&t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}

	return &t, nil
}

// Update atualiza uma transação existente
func (r *TransactionRepository) Update(ctx context.Context, t *models.Transaction) error {
	query := `
		UPDATE transactions
		SET category_id = $1, title = $2, amount = $3, type = $4, date = $5, payment_method = $6, notes = $7
		WHERE id = $8 AND user_id = $9;
	`
	tag, err := r.db.Pool.Exec(
		ctx,
		query,
		t.CategoryID,
		t.Title,
		t.Amount,
		t.Type,
		t.Date,
		t.PaymentMethod,
		t.Notes,
		t.ID,
		t.UserID,
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar transação: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("transação não encontrada ou você não tem permissão")
	}

	return nil
}

// Delete remove uma transação
func (r *TransactionRepository) Delete(ctx context.Context, userID, id string) error {
	query := `
		DELETE FROM transactions
		WHERE id = $1 AND user_id = $2;
	`
	tag, err := r.db.Pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("erro ao excluir transação: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("transação não encontrada")
	}

	return nil
}
