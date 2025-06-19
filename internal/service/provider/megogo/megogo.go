package megogo

import (
	"context"
	"fmt"
	"github.com/mechta-market/e-product/internal/service/provider/megogo/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type Service struct {
	repo RepoI
}

func New(repo RepoI) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrder(ctx context.Context, req *providerModel.OrderRequest) (*providerModel.OrderResponse, error) {
	megogoReq := &model.OrderReq{
		ProviderProductID: *req.ProviderProductID,
		CustomerPhone:     *req.CustomerPhone,
	}

	megogoResp, err := s.repo.CreateOrder(ctx, megogoReq)
	if err != nil {
		return nil, fmt.Errorf("repo.CreateOrder: %w", err)
	}

	if !megogoResp.Success {
		return nil, fmt.Errorf("MEGOGO activation failed")
	}

	result := &providerModel.OrderResponse{
		ProviderProductID: &megogoResp.ProviderProductID,
		Value:             &megogoResp.Value,
	}

	return result, nil
}

func (s *Service) CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	megogoReq := &model.CancelReq{
		ProviderProductID: *req.ProviderProductID,
		CustomerPhone:     *req.CustomerPhone,
	}

	megogoResp, err := s.repo.CancelOrder(ctx, megogoReq)
	if err != nil {
		return nil, fmt.Errorf("repo.CancelOrder: %w", err)
	}

	if !megogoResp.Success {
		return nil, fmt.Errorf("MEGOGO cancellation failed")
	}

	result := &providerModel.CancelResponse{
		Success: megogoResp.Success,
	}

	return result, nil
}
