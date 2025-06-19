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

func (s *Service) ListCatalog(ctx context.Context, providerID string) ([]*model.CatalogProduct, bool, error) {
	err := s.validateCatalog(ctx, &providerID)
	if err != nil {
		return nil, false, fmt.Errorf("validate: %w", err)
	}

	result, err := s.repo.GetCatalog(ctx, providerID)
	if err != nil {
		return nil, false, fmt.Errorf("repo.GetCatalog: %w", err)
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

func (s *Service) validateCatalog(_ context.Context, providerID *string) error {
	*providerID = strings.TrimSpace(*providerID)

	if *providerID == "" {
		return errs.ProviderIDRequired
	}

	return nil
}
