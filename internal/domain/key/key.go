package key

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"strings"
	"time"

	"github.com/mechta-market/e-product/internal/constant"
	commonModel "github.com/mechta-market/e-product/internal/domain/common/model"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"github.com/mechta-market/e-product/internal/errs"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
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

func (s *Service) Load(ctx context.Context, keys []*model.Edit) error {
	if len(keys) == 0 {
		return errs.ErrFull{
			Err:  errs.EmptyData,
			Desc: "keys list cannot be empty",
		}
	}

	for _, keyEdit := range keys {
		keyEdit.Status = lo.ToPtr(constant.KeyStatusNew)
		keyEdit.UpdatedAt = lo.ToPtr(time.Now())
		keyEdit.OrderID = lo.ToPtr("")
		keyEdit.CustomerPhone = lo.ToPtr("")
		keyEdit.ProviderOrderID = lo.ToPtr("")
		err := s.set(ctx, keyEdit)
		if err != nil {
			return fmt.Errorf("repoDb.Set: %w", err)
		}
	}

	return nil
}

func (s *Service) set(ctx context.Context, obj *model.Edit) error {
	var err error

	err = s.validateLoad(ctx, obj)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	existingKey, err := s.repoDb.GetByValue(ctx, *obj.Value)
	if err != nil {
		return fmt.Errorf("repoDb.GetByValue: %w", err)
	}

	if existingKey != nil {
		obj.ID = &existingKey.ID
	}

	err = s.repoDb.Set(ctx, obj)
	if err != nil {
		return fmt.Errorf("repoDb.Set: %w", err)
	}

	if obj.ID == nil {
		return fmt.Errorf("repoDb.Set")
	}

	return nil
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
		return nil, false, fmt.Errorf("repoDb.Get: %w", err)
	}
	if !found {
		if errNE {
			return nil, false, errs.ErrFull{
				Err:  errs.ObjectNotFound,
				Desc: fmt.Sprintf("Заказ с номером %s не найден", orderID),
			}
		}
		return nil, false, nil
	}

	return result, true, nil
}

func (s *Service) ActivateWithPool(ctx context.Context, obj *providerModel.OrderRequest) (*model.Edit, error) {
	key := &model.Edit{
		ProductID:     lo.ToPtr(obj.ProductID),
		OrderID:       lo.ToPtr(obj.OrderID),
		CustomerPhone: lo.ToPtr(obj.CustomerPhone),
	}

	listReq := &model.ListReq{
		ListParams: commonModel.ListParams{
			Page:           0,
			PageSize:       1,
			WithTotalCount: true,
		},
		ProductID: key.ProductID,
		Status:    lo.ToPtr(constant.KeyStatusNew),
	}

	items, tCount, err := s.repoDb.List(ctx, listReq)
	if err != nil {
		return nil, fmt.Errorf("repoDb.List: %w", err)
	}

	if tCount == 0 {
		return nil, errs.ObjectNotFound
	}

	key.UpdatedAt = lo.ToPtr(time.Now())

	keyToUpdate := &model.Edit{
		ID:            &items[0].ID,
		Value:         &items[0].Value,
		UpdatedAt:     key.UpdatedAt,
		ProductID:     key.ProductID,
		CustomerPhone: key.CustomerPhone,
		OrderID:       key.OrderID,
		Status:        lo.ToPtr(constant.KeyStatusActivated),
	}

	err = s.repoDb.Update(ctx, keyToUpdate)
	if err != nil {
		return nil, fmt.Errorf("repoDb.Update: %w", err)
	}

	return &model.Edit{
		Value: &items[0].Value,
	}, nil
}

func (s *Service) ActivateWithProvider(ctx context.Context, req *providerModel.OrderRequest, rep *providerModel.OrderResponse) error {
	key := &model.Edit{
		ID:                rep.ID,
		ProviderOrderID:   rep.OrderID,
		OrderID:           lo.ToPtr(req.OrderID),
		ProviderID:        lo.ToPtr(req.ProviderID),
		ProductID:         lo.ToPtr(req.ProductID),
		ProviderProductID: lo.ToPtr(req.ProviderProductID),
		Value:             lo.ToPtr(rep.Value),
		CustomerPhone:     lo.ToPtr(req.CustomerPhone),
		Status:            lo.ToPtr(constant.KeyStatusActivated),
	}

	existingKey, err := s.repoDb.GetByValue(ctx, *key.Value)
	if err != nil {
		return fmt.Errorf("key.repo.PG.GetByValue %w", err)
	}

	if existingKey != nil {
		rep.ID = &existingKey.ID
		return nil
	}

	// if provider doesnt have order ID
	if key.ProviderOrderID == nil {
		key.ProviderOrderID = lo.ToPtr("")
	}

	err = s.repoDb.CreateWithProvider(ctx, key)
	if err != nil {
		return fmt.Errorf("repoDb.CreateWithProvider: %w", err)
	}

	if key.ID == nil {
		return fmt.Errorf("repoDb.CreateWithProvider")
	}

	return nil
}

func (s *Service) Update(ctx context.Context, obj *model.Edit) error {
	err := s.repoDb.Update(ctx, obj)
	if err != nil {
		return fmt.Errorf("repoDb.Update: %w", err)
	}

	return nil
}

func (s *Service) validateLoad(_ context.Context, obj *model.Edit) error {
	*obj.ProductID = strings.TrimSpace(*obj.ProductID)
	*obj.Value = strings.TrimSpace(*obj.Value)

	if *obj.ProductID == "" {
		return errs.ProductIDRequired
	}

	if *obj.Value == "" {
		return errs.ValueRequired
	}

	return nil
}
