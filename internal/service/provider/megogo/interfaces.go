package megogo

import (
	"context"

	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type RepoI interface {
	CreateOrder(ctx context.Context, req *providerModel.OrderRequest) (*providerModel.OrderResponse, error)
	CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error)
}
