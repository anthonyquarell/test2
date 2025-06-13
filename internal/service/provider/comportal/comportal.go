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
		ProductSKU:   req.ProductID, // external_number
		Code:         req.ProviderProductID,
		PromotionKey: req.PromotionKey,
	}

	comportalRep, err := s.repo.CreateOrder(ctx, comportalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order in comportal: %w", err)
	}

	return &providerModel.OrderResponse{
		OrderID: comportalRep.OrderID,
		Value:   comportalRep.Value,
		Link:    comportalRep.Link,
	}, nil
}
