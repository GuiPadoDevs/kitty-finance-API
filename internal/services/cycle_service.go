package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"financas-sah-api/internal/models"
	"financas-sah-api/internal/repository"
)

type CycleService struct {
	cycleRepo *repository.CycleRepository
}

func NewCycleService(cycleRepo *repository.CycleRepository) *CycleService {
	return &CycleService{cycleRepo: cycleRepo}
}

// GetCurrentCycle retorna o resumo do ciclo aberto atual
func (s *CycleService) GetCurrentCycle(ctx context.Context, userID string) (*models.CycleSummaryResponse, error) {
	if userID == "" {
		return nil, errors.New("identificação de usuário ausente")
	}

	summary, err := s.cycleRepo.GetCurrentOpenCycle(ctx, userID)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, errors.New("nenhum ciclo financeiro ativo no momento")
	}

	return summary, nil
}

// ListCycles lista todos os ciclos do usuário (histórico)
func (s *CycleService) ListCycles(ctx context.Context, userID string, limit, offset int) ([]models.FinancialCycle, error) {
	if userID == "" {
		return nil, errors.New("identificação de usuário ausente")
	}
	return s.cycleRepo.ListCycles(ctx, userID, limit, offset)
}

// GetCycleByID retorna os dados de um ciclo por ID
func (s *CycleService) GetCycleByID(ctx context.Context, userID, cycleID string) (*models.FinancialCycle, error) {
	if userID == "" || cycleID == "" {
		return nil, errors.New("usuário e ID do ciclo são obrigatórios")
	}

	cycle, err := s.cycleRepo.GetCycleByID(ctx, userID, cycleID)
	if err != nil {
		return nil, err
	}
	if cycle == nil {
		return nil, errors.New("ciclo financeiro não encontrado")
	}

	return cycle, nil
}

// CloseCycle finaliza o ciclo atual e inicializa o próximo
func (s *CycleService) CloseCycle(ctx context.Context, userID string, req models.CloseCycleRequest) (map[string]interface{}, error) {
	if userID == "" {
		return nil, errors.New("identificação de usuário ausente")
	}

	req.NewCycleName = strings.TrimSpace(req.NewCycleName)
	if req.NewCycleName == "" {
		return nil, errors.New("por favor, dê um nome para o novo ciclo (ex: '05/Nov a 04/Dez' ou 'Novembro')")
	}

	closed, next, err := s.cycleRepo.CloseCycleAndOpenNew(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("falha ao fechar ciclo: %w", err)
	}

	return map[string]interface{}{
		"message":      "Ciclo finalizado com sucesso! Novo ciclo iniciado 🎀",
		"closed_cycle": closed,
		"new_cycle":    next,
	}, nil
}
