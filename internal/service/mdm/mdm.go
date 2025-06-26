package mdm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mechta-market/e-product/internal/errs"
	"github.com/mechta-market/e-product/internal/service/mdm/model"
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
		switch {
		case errors.Is(err, errs.ObjectNotFound):
			return nil, false, errs.ErrFull{
				Err:  errs.ObjectNotFound,
				Desc: "Продукт не найден",
				Fields: map[string]string{
					"product_id": *productID,
				},
			}
		default:
			slog.Error("repo.GetByProductID", "err", err)
			return nil, false, errs.ErrFull{
				Err:  errs.MdmNotAvailable,
				Desc: "Произошла ошибка. Повторите попытку позднее",
			}
		}
	}

	if result.ProviderID == "" {
		return result, false, fmt.Errorf("provider_id is empty for product %d", productID)
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
