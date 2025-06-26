package megogo

import (
	"context"
	"fmt"

	"github.com/mechta-market/e-product/internal/errs"
	"github.com/mechta-market/e-product/internal/service/provider"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type Service struct {
	repo RepoI
	key  provider.KeyI
}

func New(repo RepoI, key provider.KeyI) *Service {
	return &Service{
		repo: repo,
		key:  key,
	}
}

func (s *Service) CreateOrder(ctx context.Context, req *providerModel.OrderRequest) (*providerModel.OrderResponse, error) {
	megogoRep, err := s.repo.CreateOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("repo.CreateOrder: %w", err)
	}

	err = s.key.ActivateWithProvider(ctx, req, megogoRep)
	if err != nil {
		return nil, fmt.Errorf("key.ActivateWithProvider: %w", err)
	}

	return megogoRep, nil
}

func (s *Service) CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	megogoRep, err := s.repo.CancelOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("repo.CancelOrder: %w", err)
	}

	if !megogoRep.Success {
		return nil, fmt.Errorf("MEGOGO cancellation failed")
	}

	return megogoRep, nil
}

func (s *Service) ListCatalog(ctx context.Context, providerID string) ([]*providerModel.CatalogResponse, error) {
	return nil, errs.ErrFull{
		Err:  errs.MethodNotSupported,
		Desc: "Провайдер не поддерживает данную услугу",
	}
}
