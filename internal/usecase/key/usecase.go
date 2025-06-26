package key

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"log/slog"
	"strings"
	"time"

	"github.com/mechta-market/e-product/internal/constant"
	"github.com/mechta-market/e-product/internal/domain/common/util"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"github.com/mechta-market/e-product/internal/errs"
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

func (u *Usecase) Load(ctx context.Context, obj []*model.Edit) error {
	err := u.service.Load(ctx, obj)
	if err != nil {
		return fmt.Errorf("service.Load: %w", err)
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
		slog.Error("providerService.ListCatalog", "err", err)
		return nil, err
	}

	return items, nil
}

func (u *Usecase) Activate(ctx context.Context, productID, orderID, customerPhone string) (*model.Main, error) {
	err := u.validateActivate(ctx, orderID, productID, customerPhone)
	if err != nil {
		return nil, err
	}
	// get product_id, provider_id, product_provider_id, promotion_key
	product, found, err := u.mdmService.FindProduct(ctx, &productID)
	if err != nil {
		return nil, fmt.Errorf("mdmService.FindProduct: %w", err)
	}

	if !found {
		return nil, nil
	}

	providerService, err := u.getProvider(product.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("providerService.GetProvider: %w", err)
	}

	orderReq := &providerModel.OrderRequest{
		ProviderID:                product.ProviderID,
		ProductID:                 product.ProductID,
		ProviderProductID:         product.ProviderProductID,
		ProviderExternalProductID: product.ProviderExternalID, // для компортала
		PromotionKey:              product.PromotionKey,       // для компортала
		OrderID:                   orderID,
		CustomerPhone:             customerPhone,
	}

	order, err := providerService.CreateOrder(ctx, orderReq)
	if err != nil {
		slog.Error("providerService.CreateOrder", "err", err)
		switch {
		case errors.Is(err, errs.ObjectNotFound):
			return nil, errs.ErrFull{
				Err:  errs.ObjectNotFound,
				Desc: "Услуга провайдера в данный момент недоступна. Не найдено доступных ключей для данного продукта",
			}
		default:
			return nil, errs.ErrFull{
				Err:  errs.ProviderNotAvailable,
				Desc: "Услуга провайдера в данный момент недоступна",
			}

		}
	}

	return &model.Main{
		Value: order.Value,
	}, nil
}

func (u *Usecase) Cancel(ctx context.Context, orderID string) (*string, error) {
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

	providerService, err := u.getProvider(product.ProviderID)
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
		slog.Error("providerService.CancelOrder", "err", err)
		return nil, errs.ErrFull{
			Err:  errs.ProviderNotAvailable,
			Desc: "Услуга провайдера в данный момент недоступна",
		}
	}

	if !cancel.Success {
		return nil, fmt.Errorf("providerService.CancelOrder: %w", err)
	}

	err = u.service.Update(ctx, &model.Edit{
		ID:        &product.ID,
		Value:     &product.Value,
		UpdatedAt: lo.ToPtr(time.Now()),
		Status:    lo.ToPtr(constant.KeyStatusCancelled),
	})
	if err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	return &product.ID, nil
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

	_, found, err := u.service.GetByOrderID(ctx, orderID, false)
	if err != nil {
		return fmt.Errorf("service.GetByOrderID: %w", err)
	}
	if found {
		return errs.ErrFull{
			Err:  errs.AlreadyExists,
			Desc: "Заказ с таким номером уже существует",
		}
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
