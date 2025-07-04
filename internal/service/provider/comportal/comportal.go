package comportal

import (
	"context"
	"fmt"
	"strings"

	"github.com/mechta-market/e-product/internal/errs"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type Service struct {
	repo RepoI
}

func New(repo RepoI) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateOrder(ctx context.Context, req *providerModel.OrderRequest) (*providerModel.OrderResponse, error) {
	comportalRep, err := s.repo.CreateOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("repo.CreateOrder: %w", err)
	}

	return comportalRep, nil
}

func (s *Service) CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	return nil, errs.ErrFull{
		Err:  errs.MethodNotSupported,
		Desc: "Данный провайдер не поддерживает услугу аннулирования заказа",
	}
}

func (s *Service) ListCatalog(ctx context.Context, providerID string) ([]*providerModel.CatalogResponse, error) {
	err := s.validateCatalog(ctx, &providerID)
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	result, err := s.repo.GetCatalog(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("repo.GetCatalog: %w", err)
	}

	return result, nil
}

func (s *Service) SupportsPool() bool {
	return true
}

func (s *Service) validateCatalog(_ context.Context, providerID *string) error {
	*providerID = strings.TrimSpace(*providerID)

	if *providerID == "" {
		return errs.ProviderIDRequired
	}

	return nil
}
