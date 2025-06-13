package key

import (
	"context"
	"fmt"
	"github.com/mechta-market/e-product/internal/constant"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"github.com/mechta-market/e-product/internal/errs"
	"github.com/samber/lo"
	"strings"
	"time"
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

func (s *Service) Load(ctx context.Context, keys []*model.Key) ([]*model.Main, error) {
	if len(keys) == 0 {
		return nil, errs.ErrFull{
			Err:  errs.EmptyData,
			Desc: "keys list cannot be empty",
		}
	}

	var createdKeys []*model.Main

	for _, keyEdit := range keys {
		id, err := s.createIfNotExist(ctx, keyEdit)
		if err != nil {
			return nil, fmt.Errorf("create key: %w", err)
		}

		createdKey, found, err := s.repoDb.Get(ctx, id)
		if err != nil || !found {
			return nil, errs.ErrFull{
				Err:  err,
				Desc: fmt.Sprintf("failed to get created key ID %s", id),
				Fields: map[string]string{
					"key_id": id,
				},
			}
		}

		createdKeys = append(createdKeys, createdKey)
	}

	return createdKeys, nil
}

func (s *Service) Get(ctx context.Context, ordId string, errNE bool) (*model.Main, bool, error) {
	result, found, err := s.repoDb.Get(ctx, ordId)
	if err != nil {
		return nil, false, fmt.Errorf("repoDb.Get: %w", err)
	}
	if !found {
		if errNE {
			return nil, false, nil // errs.ObjectNotFound ?
		}
		return nil, false, nil
	}

	return result, true, nil
}

func (s *Service) createIfNotExist(ctx context.Context, obj *model.Key) (string, error) {
	var err error

	err = s.validateLoad(ctx, obj)
	if err != nil {
		return "", fmt.Errorf("validate: %w", err)
	}

	err = s.repoDb.CreateIfNotExist(ctx, obj)
	if err != nil {
		return "", fmt.Errorf("repoDb.CreateIfNotExist: %w", err)
	}

	if obj.ID == nil {
		return "", fmt.Errorf("repository failed to set ID after CreateIfNotExist")
	}

	return *obj.ID, nil
}

func (s *Service) Activate(ctx context.Context, obj *model.Key) (*model.Key, error) {
	err := s.validateActivate(ctx, obj)
	if err != nil {
		return nil, err
	}

	availableKey, err := s.repoDb.GetForActivate(ctx, *obj.OrderID, *obj.ProductID, *obj.CustomerPhone)
	if err != nil {
		return nil, fmt.Errorf("repoDb.GetForActivate: %w", err)
	}

	if availableKey == nil {
		return nil, errs.ErrFull{
			Err:  errs.ObjectNotFound,
			Desc: "no available keys found for this order",
			Fields: map[string]string{
				"order_id":    *obj.OrderID,
				"product_id":  *obj.ProductID,
				"customer_id": *obj.CustomerPhone,
			},
		}
	}

	if availableKey.Status == constant.KeyStatusActivated {
		return nil, errs.ErrFull{
			Err:  errs.AlreadyActivated,
			Desc: "key is already activated",
			Fields: map[string]string{
				"key_id": availableKey.ID,
			},
		}
	}

	keyValue := availableKey.Value
	obj.UpdatedAt = lo.ToPtr(time.Now())

	keyToUpdate := &model.Key{
		ID:            &availableKey.ID,
		UpdatedAt:     obj.UpdatedAt,
		ProductID:     obj.ProductID,
		Status:        lo.ToPtr(constant.KeyStatusActivated),
		CustomerPhone: obj.CustomerPhone,
		OrderID:       obj.OrderID,
	}

	err = s.repoDb.Update(ctx, keyToUpdate)
	if err != nil {
		return nil, fmt.Errorf("repoDb.Update: %w", err)
	}

	return &model.Key{
		Value: &keyValue,
	}, nil
}

func (s *Service) validateLoad(_ context.Context, obj *model.Key) error {
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

func (s *Service) validateActivate(_ context.Context, obj *model.Key) error {

	if *obj.ProductID == "" {
		return errs.ProductIDRequired
	}

	if *obj.OrderID == "" {
		return errs.OrderIDRequired
	}

	if *obj.CustomerPhone == "" {
		return errs.CustomerPhoneRequired
	}

	if obj.CustomerPhone != nil {
		if !obj.NormalizeAndValidatePhone() {
			return errs.ErrFull{
				Err:  errs.InvalidPhone,
				Desc: "Номер телефона клиента не прошел валидацию: " + obj.PhonesStr(),
			}
		}
	}

	return nil
}

func (s *Service) AddWithProvider(ctx context.Context, key *model.Key) (*model.Main, error) {
	err := s.repoDb.CreateWithProvider(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("repoDb.CreateWithProvider: %w", err)
	}

	if key.ID == nil {
		return nil, fmt.Errorf("repository did not set ID after CreateWithProvider")
	}

	createdKey, found, err := s.repoDb.Get(ctx, *key.ID)
	if err != nil {
		return nil, fmt.Errorf("repoDb.Get: %w", err)
	}
	if !found {
		return nil, errs.ErrFull{
			Err:  errs.ObjectNotFound,
			Desc: fmt.Sprintf("failed to get created key ID %s", *key.ID),
			Fields: map[string]string{
				"key_id": *key.ID,
			},
		}
	}

	return createdKey, nil
}
