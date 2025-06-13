package mdm

import (
	"context"
	"fmt"
	"github.com/mechta-market/e-product/internal/errs"
	"github.com/mechta-market/e-product/internal/service/mdm/model"
	"strings"
)

type Service struct {
	repo RepoI
}

func New(repo RepoI) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) FindProduct(ctx context.Context, productID *string) (*model.Product, bool, error) {
	err := s.validate(ctx, productID)
	if err != nil {
		return nil, false, fmt.Errorf("validate: %w", err)
	}

	result, err := s.repo.GetByProductID(ctx, *productID)
	if err != nil {
		return nil, false, fmt.Errorf("repo.GetByProductID: %w", err)
	}

	if result.ProviderID == nil || *result.ProviderID == "" {
		return result, false, fmt.Errorf("provider_id is empty for product %s", productID)
	}

	return result, true, nil
}

func (s *Service) validate(_ context.Context, productID *string) error {
	*productID = strings.TrimSpace(*productID)

	if *productID == "" {
		return errs.ProductIDRequired
	}

	return nil
}
