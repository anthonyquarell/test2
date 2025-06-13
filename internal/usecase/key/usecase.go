package key

import (
	"context"
	"fmt"
	"github.com/mechta-market/e-product/internal/constant"
	"github.com/mechta-market/e-product/internal/domain/common/util"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"github.com/mechta-market/e-product/internal/errs"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
	"github.com/samber/lo"
)

type Usecase struct {
	service            KeyServiceI
	mdmService         MdmServiceI
	providerService    ProviderServiceI
	providerMngService ProviderServiceMngI
}

func New(service KeyServiceI, mdmService MdmServiceI) *Usecase {
	return &Usecase{
		service:    service,
		mdmService: mdmService,
	}
}

func (u *Usecase) List(ctx context.Context,
	pars *model.ListReq,
) ([]*model.Main, int64, error) {
	if err := util.RequirePageSize(pars.ListParams, constant.MaxPageSize); err != nil {
		return nil, 0, errs.IncorrectPageSize
	}

	items, tCount, err := u.service.List(ctx, pars)
	if err != nil {
		return nil, 0, fmt.Errorf("service.List: %w", err)
	}

	for _, item := range items {
		item.Compute()
	}

	return items, tCount, nil
}

func (u *Usecase) Load(ctx context.Context, obj []*model.Key) ([]*Key, error) {
	createdKeys, err := u.service.Load(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("service.Load: %w", err)
	}

	result := lo.Map(createdKeys, func(key *model.Main, _ int) *Key {
		// в ответе не должно быть ключа
		key.Value = ""
		return &Key{Main: *key}
	})

	return result, nil
}

func (u *Usecase) Get(ctx context.Context,
	id string,
) (*Key, error) {
	key, found, err := u.service.Get(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("service.Get: %w", err)
	}
	if !found {
		return nil, nil
	}

	key.Compute()

	return &Key{
		Main: *key,
	}, nil
}

func (u *Usecase) Activate(ctx context.Context,
	obj *model.Key,
) (*model.Key, error) {
	result, err := u.service.Activate(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("service.Set: %w", err)
	}

	return result, nil
}

func (u *Usecase) GetMdmProduct(ctx context.Context, productID string) (*Key, error) {
	key, found, err := u.mdmService.FindProduct(ctx, &productID)
	if err != nil {
		return nil, fmt.Errorf("mdmService.FindProduct: %w", err)
	}
	if !found {
		return nil, nil
	}

	return &Key{
		Main: DecodeMdmRep(key),
	}, nil
}

func (u *Usecase) ActivateByProvider(ctx context.Context, productID, customerPhone string) (*Key, error) {

	// get product_id , provider_id, product_provider_id, promotion_key
	product, found, err := u.mdmService.FindProduct(ctx, &productID)
	if err != nil {
		return nil, fmt.Errorf("mdmService.FindProduct: %w", err)
	}

	if !found {
		return nil, nil
	}

	providerService, err := u.providerMngService.GetProvider(*product.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("providerMngService.GetProvider: %w", err)
	}

	orderReq := &providerModel.OrderRequest{
		ProductID:         product.ProductID,
		ProviderProductID: product.ProviderProductID,
		PromotionKey:      product.PromotionKey, // для компортала
	}

	order, err := providerService.CreateOrder(ctx, orderReq)
	if err != nil {
		return nil, fmt.Errorf("providerService.CreateOrder: %w", err)
	}

	newKey := &model.Key{
		ID:            order.ID,
		ProviderID:    order.ProviderID,
		ProductID:     order.ProductID,
		Value:         order.Value,
		OrderID:       order.OrderID,
		CustomerPhone: &customerPhone,
		Status:        lo.ToPtr(constant.KeyStatusActivated),
	}

	key, err := u.service.AddWithProvider(ctx, newKey)
	if err != nil {
		return nil, fmt.Errorf("service.AddWithProvider: %w", err)
	}

	return &Key{
		Main: *key,
	}, nil
}
