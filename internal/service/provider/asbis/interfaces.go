package asbis

import (
	"context"

	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type RepoI interface {
	CreateOrder(ctx context.Context, obj *providerModel.OrderRequest) (*providerModel.OrderResponse, error)
	CancelOrder(ctx context.Context, obj *providerModel.CancelRequest) (*providerModel.CancelResponse, error)
}
