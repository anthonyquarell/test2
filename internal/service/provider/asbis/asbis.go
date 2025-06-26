package asbis

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"log/slog"

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
	asbisRep, err := s.repo.CreateOrder(ctx, req)
	if err != nil {
		slog.Error("repo.CreateOrder", "err", err)
		// обращение к бд при ошибке
		item, poolErr := s.key.ActivateWithPool(ctx, req)
		if poolErr != nil {
			return nil, poolErr
		}
		return &providerModel.OrderResponse{
			Value: lo.FromPtr(item.Value),
		}, nil
	}

	err = s.key.ActivateWithProvider(ctx, req, asbisRep)
	if err != nil {
		return nil, fmt.Errorf("key.ActivateWithProvider: %w", err)
	}

	return asbisRep, nil
}

func (s *Service) CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	asbisRep, err := s.repo.CancelOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("repo.CancelOrder: %w", err)
	}

	if !asbisRep.Success {
		return nil, fmt.Errorf("ASBIS cancellation failed")
	}

	result := &providerModel.CancelResponse{
		ID:      asbisRep.ID,
		Success: asbisRep.Success,
	}

	return result, nil
}

func (s *Service) ListCatalog(ctx context.Context, providerID string) ([]*providerModel.CatalogResponse, error) {
	return nil, errs.ErrFull{
		Err:  errs.MethodNotSupported,
		Desc: "Провайдер не поддерживает данную услугу",
	}
}
