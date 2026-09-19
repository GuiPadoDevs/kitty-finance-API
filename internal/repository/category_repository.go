package repository

import (
	"context"
	"errors"
	"fmt"

	"financas-sah-api/internal/database"
	"financas-sah-api/internal/models"

	"github.com/jackc/pgx/v5"
)

type CategoryRepository struct {
	db *database.DB
}

func NewCategoryRepository(db *database.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// List retorna as categorias do usuário, com filtro opcional por tipo ('income'/'expense')
func (r *CategoryRepository) List(ctx context.Context, userID, catType string) ([]models.Category, error) {
	query := `
		SELECT id, user_id, name, type, icon, color, budget_limit, created_at
		FROM categories
		WHERE user_id = $1
	`
	args := []interface{}{userID}

	if catType != "" {
		query += " AND type = $2"
		args = append(args, catType)
	}

	query += " ORDER BY name ASC;"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Name,
			&c.Type,
			&c.Icon,
			&c.Color,
			&c.BudgetLimit,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha de categoria: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, nil
}

// GetByID busca uma categoria pelo ID e garante que pertence ao usuário
func (r *CategoryRepository) GetByID(ctx context.Context, userID, id string) (*models.Category, error) {
	query := `
		SELECT id, user_id, name, type, icon, color, budget_limit, created_at
		FROM categories
		WHERE id = $1 AND user_id = $2;
	`
	var c models.Category
	err := r.db.Pool.QueryRow(ctx, query, id, userID).Scan(
		&c.ID,
		&c.UserID,
		&c.Name,
		&c.Type,
		&c.Icon,
		&c.Color,
		&c.BudgetLimit,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar categoria: %w", err)
	}

	return &c, nil
}

// Create insere uma nova categoria
func (r *CategoryRepository) Create(ctx context.Context, c *models.Category) error {
	query := `
		INSERT INTO categories (user_id, name, type, icon, color, budget_limit)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at;
	`
	return r.db.Pool.QueryRow(
		ctx,
		query,
		c.UserID,
		c.Name,
		c.Type,
		c.Icon,
		c.Color,
		c.BudgetLimit,
	).Scan(&c.ID, &c.CreatedAt)
}

// Update atualiza uma categoria existente
func (r *CategoryRepository) Update(ctx context.Context, c *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, icon = $2, color = $3, budget_limit = $4
		WHERE id = $5 AND user_id = $6;
	`
	tag, err := r.db.Pool.Exec(ctx, query, c.Name, c.Icon, c.Color, c.BudgetLimit, c.ID, c.UserID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar categoria: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("categoria não encontrada ou você não tem permissão")
	}

	return nil
}

// Delete remove uma categoria após verificar se não há transações associadas
func (r *CategoryRepository) Delete(ctx context.Context, userID, id string) error {
	// 1. Verifica se existem transações vinculadas
	queryCount := `
		SELECT COUNT(*)
		FROM transactions
		WHERE category_id = $1 AND user_id = $2;
	`
	var count int
	if err := r.db.Pool.QueryRow(ctx, queryCount, id, userID).Scan(&count); err != nil {
		return fmt.Errorf("erro ao verificar vínculos da categoria: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("esta categoria não pode ser excluída porque possui %d lançamento(s) associado(s)", count)
	}

	// 2. Deleta
	queryDelete := `
		DELETE FROM categories
		WHERE id = $1 AND user_id = $2;
	`
	tag, err := r.db.Pool.Exec(ctx, queryDelete, id, userID)
	if err != nil {
		return fmt.Errorf("erro ao excluir categoria: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("categoria não encontrada")
	}

	return nil
}
