package services

import (
	"context"
	"errors"

	"financas-sah-api/internal/models"
	"financas-sah-api/internal/repository"
)

type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
}

func NewDashboardService(dashboardRepo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

// GetSummary retorna o resumo com métricas, porcentagens e agrupamentos
func (s *DashboardService) GetSummary(ctx context.Context, userID, cycleID string) (*models.DashboardSummaryResponse, error) {
	if userID == "" {
		return nil, errors.New("não autorizado")
	}
	return s.dashboardRepo.GetSummary(ctx, userID, cycleID)
}
