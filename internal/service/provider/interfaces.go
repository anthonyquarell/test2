package provider

import (
	"context"
	"github.com/mechta-market/e-product/internal/service/provider/model"
)

type ServiceI interface {
	CreateOrder(ctx context.Context, obj *model.OrderRequest) (*model.OrderResponse, error)

	//GetProduct(ctx context.Context, providerProductID string) (*model.Product, error)

	//GetOrderStatus(ctx context.Context, orderID string) (*model.OrderRep, error)
	//
	//GetProviderID() string
	//
	//GetProviderName() string
}

type ServiceManagerI interface {
	GetProvider(providerID string) (ServiceI, error)
	ListProviders() []ServiceI
	RegisterProvider(providerID string, provider ServiceI) error
}
