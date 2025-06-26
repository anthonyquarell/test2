package provider

import (
	"context"

	domainModel "github.com/mechta-market/e-product/internal/domain/key/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type KeyI interface {
	ActivateWithPool(ctx context.Context, obj *providerModel.OrderRequest) (*domainModel.Edit, error)
	ActivateWithProvider(ctx context.Context, req *providerModel.OrderRequest, rep *providerModel.OrderResponse) error
}
