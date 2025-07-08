package key

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"time"

	"github.com/mechta-market/e-product/internal/domain/key/model"
	"github.com/mechta-market/e-product/internal/errs"
)

type Service struct {
	repoDb RepoDbI
}

func New(repoDb RepoDbI) *Service {
	return &Service{repoDb: repoDb}
}

func (s *Service) List(ctx context.Context, pars *model.ListReq) ([]*model.Main, int64, error) {
	items, tCount, err := s.repoDb.List(ctx, pars)
	if err != nil {
		return nil, 0, fmt.Errorf("repoDb.List: %w", err)
	}

	return items, tCount, nil
}

func (s *Service) Get(ctx context.Context, id string, errNE bool) (*model.Main, bool, error) {
	result, found, err := s.repoDb.Get(ctx, id)
	if err != nil {
		return nil, false, fmt.Errorf("repoDb.Get: %w", err)
	}
	if !found {
		if errNE {
			return nil, false, errs.ErrFull{
				Err:  errs.ObjectNotFound,
				Desc: "Ключ не найден",
			}
		}
		return nil, false, nil
	}

	return result, true, nil
}

func (s *Service) GetByOrderID(ctx context.Context, orderID string, errNE bool) (*model.Main, bool, error) {
	result, found, err := s.repoDb.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, false, fmt.Errorf("repoDb.GetByOrderID: %w", err)
	}
	if !found {
		if errNE {
			return nil, false, errs.ErrFull{
				Err:  errs.ObjectNotFound,
				Desc: fmt.Sprintf("Ключ с номером заказа %s не найден", orderID),
			}
		}
		return nil, false, nil
	}

	return result, true, nil
}

func (s *Service) GetByValue(ctx context.Context, value string) (*model.Main, error) {
	result, err := s.repoDb.GetByValue(ctx, value)
	if err != nil {
		return nil, fmt.Errorf("repoDb.GetByValue: %w", err)
	}

	return result, nil
}

func (s *Service) Update(ctx context.Context, obj *model.Edit) error {
	obj.UpdatedAt = lo.ToPtr(time.Now())

	err := s.repoDb.Update(ctx, obj)
	if err != nil {
		return fmt.Errorf("repoDb.Update: %w", err)
	}

	return nil
}

func (s *Service) Create(ctx context.Context, obj *model.Edit) (string, error) {
	id, err := s.repoDb.Create(ctx, obj)
	if err != nil {
		return "", fmt.Errorf("repoDb.Create: %w", err)
	}

	return id, nil
}
