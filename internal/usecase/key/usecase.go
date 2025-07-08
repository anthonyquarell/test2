package key

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"log/slog"
	"strings"

	"github.com/mechta-market/e-product/internal/constant"
	commonModel "github.com/mechta-market/e-product/internal/domain/common/model"
	"github.com/mechta-market/e-product/internal/domain/common/util"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"github.com/mechta-market/e-product/internal/errs"
	mdmModel "github.com/mechta-market/e-product/internal/service/mdm/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type Usecase struct {
	service    KeyServiceI
	mdmService MdmServiceI
	providers  map[string]ProviderServiceI
}

func New(service KeyServiceI, mdmService MdmServiceI, providers map[string]ProviderServiceI) *Usecase {
	return &Usecase{
		service:    service,
		mdmService: mdmService,
		providers:  providers,
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

	return items, tCount, nil
}

func (u *Usecase) Load(ctx context.Context, objs []*model.Edit) error {
	if len(objs) == 0 {
		return errs.ErrFull{
			Err:  errs.EmptyData,
			Desc: "keys list cannot be empty",
		}
	}

	for _, obj := range objs {
		err := u.load(ctx, obj)
		if err != nil {
			return fmt.Errorf("load: %w", err)
		}
	}

	return nil
}

func (u *Usecase) load(ctx context.Context, obj *model.Edit) error {
	err := u.validateLoad(ctx, obj)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	existingKey, err := u.service.GetByValue(ctx, *obj.Value)
	if err != nil {
		return fmt.Errorf("service.GetByValue: %w", err)
	}

	// пропускаем ключ, если уже имеется запись в БД
	if existingKey != nil && existingKey.ProductID == *obj.ProductID {
		slog.Warn("attempt to add duplicate key", "value", obj.Value, "product_id", obj.ProductID)
		return nil
	}

	_, err = u.service.Create(ctx, obj)
	if err != nil {
		return fmt.Errorf("service.Create: %w", err)
	}

	return nil
}

func (u *Usecase) Get(ctx context.Context,
	id string,
) (*model.Main, error) {
	key, _, err := u.service.Get(ctx, id, true)
	if err != nil {
		return nil, fmt.Errorf("service.Get: %w", err)
	}

	return key, nil
}

func (u *Usecase) GetCatalog(ctx context.Context, providerID string) ([]*providerModel.CatalogResponse, error) {
	providerService, err := u.getProvider(providerID)
	if err != nil {
		return nil, fmt.Errorf("providerService.GetProvider: %w", err)
	}

	items, err := providerService.ListCatalog(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("providerService.ListCatalog: %w", err)
	}

	return items, nil
}

func (u *Usecase) Activate(ctx context.Context, productID, orderID, customerPhone string) (*string, error) {
	if err := u.validateActivate(ctx, orderID, productID, customerPhone); err != nil {
		return nil, err
	}

	product, _, err := u.mdmService.FindProduct(ctx, &productID)
	if err != nil {
		return nil, fmt.Errorf("mdmService.FindProduct: %w", err)
	}

	providerService, err := u.getProvider(product.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("getProvider: %w", err)
	}

	// Обращение к провайдеру
	id, err := u.createOrder(ctx, providerService, product, customerPhone) // customerPhone для megogo
	if err != nil {
		slog.Error("createOrder", "error", err)

		if !providerService.SupportsPool() {
			return nil, errs.ServiceNA
		}
	}

	key, err := u.activate(ctx, id, orderID, customerPhone, productID)
	if err != nil {
		return nil, fmt.Errorf("activate: %w", err)
	}

	return lo.ToPtr(key.Value), nil
}

func (u *Usecase) createOrder(ctx context.Context, providerService ProviderServiceI, product *mdmModel.Product, customerPhone string) (string, error) {
	orderReq := &providerModel.OrderRequest{
		ProviderID:                product.ProviderID,
		ProductID:                 product.ProductID,
		ProviderProductID:         product.ProviderProductID,
		ProviderExternalProductID: product.ProviderExternalID,
		PromotionKey:              product.PromotionKey,
		CustomerPhone:             customerPhone,
	}

	orderRep, err := providerService.CreateOrder(ctx, orderReq)
	if err != nil {
		return "", fmt.Errorf("providerService.CreateOrder: %w", err)
	}

	obj := &model.Edit{
		Value:                 lo.ToPtr(orderRep.Value),
		ProductID:             lo.ToPtr(product.ProductID),
		ProviderID:            lo.ToPtr(product.ProviderID),
		ProviderProductID:     lo.ToPtr(product.ProviderProductID),
		ProviderTransactionID: lo.ToPtr(orderRep.TransactionID),
		ProviderOrderID:       orderRep.OrderID,
	}

	id, err := u.service.Create(ctx, obj)
	if err != nil {
		return "", fmt.Errorf("service.Create: %w", err)
	}

	return id, nil
}

func (u *Usecase) activate(ctx context.Context, id, orderID, customerPhone, productID string) (*model.Main, error) {
	var item *model.Main
	var err error
	// при ошибке обращения к провайдеру
	if id == "" {
		listReq := &model.ListReq{
			ListParams: commonModel.ListParams{
				Page:     0,
				PageSize: 1,
			},
			ProductID: lo.ToPtr(productID),
			Status:    lo.ToPtr(constant.KeyStatusNew),
		}

		items, _, err := u.service.List(ctx, listReq)
		if err != nil {
			return nil, fmt.Errorf("serivce.List: %w", err)
		}

		if len(items) == 0 {
			return nil, errs.ErrFull{
				Err:  errs.ObjectNotFound,
				Desc: "Услуга провайдера в данный момент недоступна. Не найдено доступных ключей для данного продукта",
			}
		}

		item = items[0]

		item, _, err = u.service.Get(ctx, item.ID, true)
		if err != nil {
			return nil, fmt.Errorf("service.Get: %w", err)
		}
	} else {
		item, _, err = u.service.Get(ctx, id, true)
		if err != nil {
			return nil, fmt.Errorf("service.Get: %w", err)
		}
	}

	if item.Status != constant.KeyStatusNew {
		return nil, errs.AlreadyActivated
	}

	key := &model.Edit{
		ID:            lo.ToPtr(item.ID),
		OrderID:       lo.ToPtr(orderID),
		CustomerPhone: lo.ToPtr(customerPhone),
		Status:        lo.ToPtr(constant.KeyStatusActivated),
	}

	err = u.service.Update(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	return item, nil
}

func (u *Usecase) Cancel(ctx context.Context, orderID string) (*string, error) {
	err := u.validateCancel(ctx, &orderID)
	if err != nil {
		return nil, err
	}

	product, _, err := u.service.GetByOrderID(ctx, orderID, true)
	if err != nil {
		return nil, fmt.Errorf("service.GetByOrderID: %w", err)
	}

	if product.Status == constant.KeyStatusCancelled {
		return nil, errs.AlreadyCancelled
	}

	providerService, err := u.getProvider(product.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("providerService.GetProvider: %w", err)
	}

	cancelReq := &providerModel.CancelRequest{
		CancelID:          &product.ProviderTransactionID,
		ProductID:         &product.ProductID,
		ProviderProductID: &product.ProviderProductID,
		CustomerPhone:     &product.CustomerPhone,
	}

	_, err = providerService.CancelOrder(ctx, cancelReq)
	if err != nil {
		return nil, fmt.Errorf("providerService.CancelOrder: %w", err)
	}

	err = u.service.Update(ctx, &model.Edit{
		ID:     &product.ID,
		Status: lo.ToPtr(constant.KeyStatusCancelled),
	})
	if err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	return lo.ToPtr(product.ID), nil
}

func (u *Usecase) getProvider(providerID string) (ProviderServiceI, error) {
	provider, exists := u.providers[providerID]
	if !exists {
		return nil, errs.ErrFull{
			Err:  errs.ObjectNotFound,
			Desc: "Услуги провайдера не подключены",
			Fields: map[string]string{
				"providerID": providerID,
			},
		}
	}

	return provider, nil
}

func (u *Usecase) validateLoad(_ context.Context, obj *model.Edit) error {
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

func (u *Usecase) validateActivate(ctx context.Context, orderID, productID, customerPhone string) error {
	if customerPhone != "" {
		customerPhone = strings.TrimSpace(customerPhone)
		if !util.NormalizeAndValidatePhone(&customerPhone) {
			return errs.ErrFull{
				Err:  errs.InvalidPhone,
				Desc: "Номер телефона клиента не прошел валидацию",
			}
		}
	}

	orderID = strings.TrimSpace(orderID)
	productID = strings.TrimSpace(productID)

	if orderID == "" {
		return errs.OrderIDRequired
	}

	if productID == "" {
		return errs.ProductIDRequired
	}

	if customerPhone == "" {
		return errs.CustomerPhoneRequired
	}

	return nil
}

func (u *Usecase) validateCancel(_ context.Context, orderID *string) error {
	*orderID = strings.TrimSpace(*orderID)

	if *orderID == "" {
		return errs.OrderIDRequired
	}

	return nil
}
