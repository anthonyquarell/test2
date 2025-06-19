package asbis

import (
	"context"
	"fmt"
	"github.com/mechta-market/e-product/internal/service/provider/asbis/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type Service struct {
	repo RepoI
}

func New(repo RepoI) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrder(ctx context.Context, req *providerModel.OrderRequest) (*providerModel.OrderResponse, error) {
	asbisReq := &model.OrderReq{
		ProductNumber: req.ProviderProductID, // ProductNumber = Номенклатура.мчт_НомерПродуктаЭлектроннойУслуги
		TermNumber:    req.ProductID,         // TermNumber = Номенклатура.мчт_ПровайдерЭлектроннойУслуги.ПартнерскийИД
	}

	asbisRep, err := s.repo.CreateOrder(ctx, asbisReq)
	if err != nil {
		return nil, fmt.Errorf("repo.CreateOrder: %w", err)
	}

	if !asbisRep.Success {
		return nil, fmt.Errorf("ASBIS activation failed: %s", asbisRep.ErrorMsg)
	}

	result := &providerModel.OrderResponse{
		ID:        asbisRep.ID,
		ProductID: req.ProductID,
		OrderID:   asbisRep.ID,
		Value:     asbisRep.Value,
		Link:      asbisRep.Link,
	}

	return result, nil
}

func (s *Service) CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	asbisReq := &model.CancelReq{
		ProductNumber:         req.ProviderProductID,
		TermNumber:            req.ProductID,
		OriginalTransactionID: req.ID,
	}

	asbisRep, err := s.repo.CancelOrder(ctx, asbisReq)
	if err != nil {
		return nil, fmt.Errorf("repo.CancelOrder: %w", err)
	}

	if !asbisRep.Success {
		return nil, fmt.Errorf("ASBIS cancellation failed")
	}

	result := &providerModel.CancelResponse{
		ID:           &asbisRep.OriginalTransactionID,
		ErrorMessage: &asbisRep.ErrorMessage,
		Success:      asbisRep.Success,
	}

	return result, nil
}
