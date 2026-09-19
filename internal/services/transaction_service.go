package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"financas-sah-api/internal/models"
	"financas-sah-api/internal/repository"
)

type TransactionService struct {
	txRepo       *repository.TransactionRepository
	categoryRepo *repository.CategoryRepository
	cycleRepo    *repository.CycleRepository
}

func NewTransactionService(
	txRepo *repository.TransactionRepository,
	categoryRepo *repository.CategoryRepository,
	cycleRepo *repository.CycleRepository,
) *TransactionService {
	return &TransactionService{
		txRepo:       txRepo,
		categoryRepo: categoryRepo,
		cycleRepo:    cycleRepo,
	}
}

// Create registra um novo lançamento financeiro
func (s *TransactionService) Create(ctx context.Context, userID string, req models.CreateTransactionRequest) (*models.Transaction, error) {
	if userID == "" {
		return nil, errors.New("não autorizado")
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return nil, errors.New("o título/descrição do lançamento é obrigatório")
	}

	if req.Amount <= 0 {
		return nil, errors.New("o valor do lançamento deve ser maior que zero")
	}

	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type != "income" && req.Type != "expense" {
		return nil, errors.New("o tipo deve ser 'income' (entrada) ou 'expense' (saída)")
	}

	if req.Date == "" {
		req.Date = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", req.Date); err != nil {
			return nil, errors.New("formato de data inválido. Use AAAA-MM-DD")
		}
	}

	// Valida se a categoria existe e pertence ao usuário
	category, err := s.categoryRepo.GetByID(ctx, userID, req.CategoryID)
	if err != nil || category == nil {
		return nil, errors.New("categoria selecionada não encontrada")
	}

	// Se cycle_id não foi informado, descobre o ciclo aberto atual
	var targetCycleID string
	if req.CycleID != nil && *req.CycleID != "" {
		cycle, err := s.cycleRepo.GetCycleByID(ctx, userID, *req.CycleID)
		if err != nil || cycle == nil {
			return nil, errors.New("ciclo financeiro especificado não encontrado")
		}
		targetCycleID = cycle.ID
	} else {
		currentCycle, err := s.cycleRepo.GetCurrentOpenCycle(ctx, userID)
		if err != nil || currentCycle == nil {
			return nil, errors.New("nenhum ciclo financeiro aberto no momento para registrar lançamentos")
		}
		targetCycleID = currentCycle.Cycle.ID
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = "Cartão"
	}

	transaction := models.Transaction{
		UserID:        userID,
		CycleID:       targetCycleID,
		CategoryID:    category.ID,
		CategoryName:  category.Name,
		CategoryIcon:  category.Icon,
		CategoryColor: category.Color,
		Title:         req.Title,
		Amount:        req.Amount,
		Type:          req.Type,
		Date:          req.Date,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
	}

	if err := s.txRepo.Create(ctx, &transaction); err != nil {
		return nil, fmt.Errorf("erro ao registrar lançamento: %w", err)
	}

	return &transaction, nil
}

// List retorna a lista de transações com filtros
func (s *TransactionService) List(ctx context.Context, userID string, filter models.TransactionFilter) ([]models.Transaction, error) {
	if userID == "" {
		return nil, errors.New("não autorizado")
	}
	return s.txRepo.List(ctx, userID, filter)
}

// GetByID busca um lançamento pelo ID
func (s *TransactionService) GetByID(ctx context.Context, userID, id string) (*models.Transaction, error) {
	if userID == "" || id == "" {
		return nil, errors.New("parâmetros inválidos")
	}
	tx, err := s.txRepo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, errors.New("lançamento não encontrado")
	}
	return tx, nil
}

// Update altera uma transação existente
func (s *TransactionService) Update(ctx context.Context, userID, id string, req models.UpdateTransactionRequest) (*models.Transaction, error) {
	if userID == "" || id == "" {
		return nil, errors.New("parâmetros inválidos")
	}

	existing, err := s.txRepo.GetByID(ctx, userID, id)
	if err != nil || existing == nil {
		return nil, errors.New("lançamento não encontrado")
	}

	if req.Title != "" {
		existing.Title = strings.TrimSpace(req.Title)
	}
	if req.Amount > 0 {
		existing.Amount = req.Amount
	}
	if req.Type == "income" || req.Type == "expense" {
		existing.Type = req.Type
	}
	if req.Date != "" {
		if _, err := time.Parse("2006-01-02", req.Date); err == nil {
			existing.Date = req.Date
		}
	}
	if req.PaymentMethod != "" {
		existing.PaymentMethod = req.PaymentMethod
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	// Se alterou a categoria, valida
	if req.CategoryID != "" && req.CategoryID != existing.CategoryID {
		category, err := s.categoryRepo.GetByID(ctx, userID, req.CategoryID)
		if err != nil || category == nil {
			return nil, errors.New("nova categoria não encontrada")
		}
		existing.CategoryID = category.ID
		existing.CategoryName = category.Name
		existing.CategoryIcon = category.Icon
		existing.CategoryColor = category.Color
	}

	if err := s.txRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete remove um lançamento
func (s *TransactionService) Delete(ctx context.Context, userID, id string) error {
	if userID == "" || id == "" {
		return errors.New("parâmetros inválidos")
	}
	return s.txRepo.Delete(ctx, userID, id)
}
