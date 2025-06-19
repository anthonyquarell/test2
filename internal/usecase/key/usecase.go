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
	"log/slog"
	"strings"
	"sync"
)

type Usecase struct {
	service     KeyServiceI
	mdmService  MdmServiceI
	providers   map[string]ProviderServiceI
	providersMu sync.RWMutex
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

func (u *Usecase) GetCatalog(ctx context.Context, providerID string) ([]*CatalogProductT, error) {
	items, found, err := u.mdmService.ListCatalog(ctx, providerID)
	if err != nil {
		slog.Error("mdmService.ListCatalog", "err", err)
		return nil, errs.ErrFull{
			Err:  errs.MdmNotAvailable,
			Desc: "Произошла ошибка. Повторите попытку позднее",
		}
	}

	if !found {
		return nil, nil
	}

	return DecodeMdmCatalog(items), nil
}

func (u *Usecase) ActivateByProvider(ctx context.Context, productID, orderID, customerPhone string) (*Key, error) {
	err := u.validateOrd(ctx, orderID, productID, customerPhone)
	if err != nil {
		return nil, err
	}
	// get product_id , provider_id, product_provider_id, promotion_key
	product, found, err := u.mdmService.FindProduct(ctx, &productID)
	if err != nil {
		slog.Error("mdmService.FindProduct", "err", err)
		return nil, errs.ErrFull{
			Err:  errs.MdmNotAvailable,
			Desc: "Произошла ошибка. Повторите попытку позднее",
		}
	}

	if !found {
		return nil, nil
	}

	providerService, err := u.GetProvider(*product.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("providerMngService.GetProvider: %w", err)
	}

	orderReq := &providerModel.OrderRequest{
		ProductID:                 product.ProductID,
		ProviderProductID:         product.ProviderProductID,
		ProviderExternalProductID: product.ProviderExternalID, // для компортала
		PromotionKey:              product.PromotionKey,       // для компортала
		CustomerPhone:             &customerPhone,
	}

	order, err := providerService.CreateOrder(ctx, orderReq)
	if err != nil {
		slog.Error("providerService.CreateOrder", "err", err)
		return nil, errs.ErrFull{
			Err:  errs.ProviderNotAvailable,
			Desc: "Услуга провайдера в данный момент недоступна",
		}
	}

	newKey := &model.Key{
		ID:                order.ID,
		OrderID:           &orderID,
		ProviderID:        ptrOrDefault(product.ProviderID, ""),
		ProductID:         ptrOrDefault(product.ProductID, ""),
		ProviderProductID: ptrOrDefault(product.ProviderProductID, ""),
		ProviderOrderID:   ptrOrDefault(order.OrderID, ""),
		Value:             order.Value,
		CustomerPhone:     &customerPhone,
		Status:            lo.ToPtr(constant.KeyStatusActivated),
	}

	key, err := u.service.AddWithProvider(ctx, newKey)
	if err != nil {
		return nil, fmt.Errorf("service.AddWithProvider: %w", err)
	}

	return &Key{
		Main: *key,
	}, nil
}

func (u *Usecase) CancelByProvider(ctx context.Context, orderID string) (*string, error) {
	err := u.validateCancel(ctx, &orderID)
	if err != nil {
		return nil, err
	}

	product, found, err := u.service.GetByOrderID(ctx, orderID, true)
	if err != nil {
		return nil, fmt.Errorf("service.GetByOrderID: %w", err)
	}

	if !found {
		return nil, nil
	}

	providerService, err := u.GetProvider(product.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("providerService.GetProvider: %w", err)
	}

	cancelReq := &providerModel.CancelRequest{
		ID:                &product.ID,
		ProductID:         &product.ProductID,
		ProviderProductID: &product.ProviderProductID,
		CustomerPhone:     &product.CustomerPhone, // для megogo
	}

	cancel, err := providerService.CancelOrder(ctx, cancelReq)
	if err != nil {
		return nil, fmt.Errorf("providerService.CancelOrder: %w", err)
	}

	if !cancel.Success {
		return nil, fmt.Errorf("providerService.CancelOrder: %w", err)
	}

	err = u.service.Delete(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("service.Delete: %w", err)
	}

	return &product.ID, nil
}

func (u *Usecase) GetProvider(providerID string) (ProviderServiceI, error) {
	u.providersMu.RLock()
	defer u.providersMu.RUnlock()

	provider, exists := u.providers[providerID]
	if !exists {
		return nil, fmt.Errorf("provider with ID %s not found", providerID)
	}

	return provider, nil
}

func (u *Usecase) validateOrd(_ context.Context, orderID, productID, customerPhone string) error {
	if customerPhone != "" {
		customerPhone = strings.TrimSpace(customerPhone)
		if !util.NormalizeAndValidatePhone(&customerPhone) {
			return errs.ErrFull{
				Err:  errs.InvalidPhone,
				Desc: "Номер телефона клиента не прошел валидацию",
			}
		}
	}

	orderID = strings.TrimSpace(productID)
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
