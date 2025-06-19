package comportal

import (
	"context"
	"fmt"
	"github.com/mechta-market/e-product/internal/service/provider/comportal/model"
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
	comportalReq := &model.OrderReq{
		ProductSKU:   req.ProviderProductID,         // external_number
		Code:         req.ProviderExternalProductID, // external_id
		PromotionKey: req.PromotionKey,
	}

	comportalRep, err := s.repo.CreateOrder(ctx, comportalReq)
	if err != nil {
		return nil, fmt.Errorf("repo.CreateOrder: %w", err)
	}

	return &providerModel.OrderResponse{
		ID:      comportalRep.ID,
		OrderID: comportalRep.OrderID,
		Value:   comportalRep.Value,
		Link:    comportalRep.Link,
	}, nil
}

func (s *Service) CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	return nil, fmt.Errorf("comportal does not support cancellation")
}
