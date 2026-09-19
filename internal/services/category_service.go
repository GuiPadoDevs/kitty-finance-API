package services

import (
	"context"
	"errors"
	"strings"

	"financas-sah-api/internal/models"
	"financas-sah-api/internal/repository"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// List lista categorias filtradas por tipo se fornecido
func (s *CategoryService) List(ctx context.Context, userID, catType string) ([]models.Category, error) {
	if userID == "" {
		return nil, errors.New("não autorizado")
	}
	catType = strings.ToLower(strings.TrimSpace(catType))
	return s.categoryRepo.List(ctx, userID, catType)
}

// Create valida e cria um novo tópico
func (s *CategoryService) Create(ctx context.Context, userID string, req models.CreateCategoryRequest) (*models.Category, error) {
	if userID == "" {
		return nil, errors.New("não autorizado")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, errors.New("o nome da categoria é obrigatório")
	}

	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type != "income" && req.Type != "expense" {
		return nil, errors.New("o tipo deve ser 'income' (entrada) ou 'expense' (saída)")
	}

	if req.Icon == "" {
		req.Icon = "sparkles"
	}
	if req.Color == "" {
		req.Color = "#FF85A2"
	}

	category := models.Category{
		UserID:      userID,
		Name:        req.Name,
		Type:        req.Type,
		Icon:        req.Icon,
		Color:       req.Color,
		BudgetLimit: req.BudgetLimit,
	}

	if err := s.categoryRepo.Create(ctx, &category); err != nil {
		return nil, err
	}

	return &category, nil
}

// Update valida e altera os dados da categoria
func (s *CategoryService) Update(ctx context.Context, userID, id string, req models.UpdateCategoryRequest) (*models.Category, error) {
	if userID == "" || id == "" {
		return nil, errors.New("parâmetros inválidos")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, errors.New("o nome da categoria é obrigatório")
	}

	existing, err := s.categoryRepo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("categoria não encontrada")
	}

	existing.Name = req.Name
	if req.Icon != "" {
		existing.Icon = req.Icon
	}
	if req.Color != "" {
		existing.Color = req.Color
	}
	existing.BudgetLimit = req.BudgetLimit

	if err := s.categoryRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete exclui uma categoria com validações
func (s *CategoryService) Delete(ctx context.Context, userID, id string) error {
	if userID == "" || id == "" {
		return errors.New("parâmetros inválidos")
	}
	return s.categoryRepo.Delete(ctx, userID, id)
}
