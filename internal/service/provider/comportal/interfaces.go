package comportal

import (
	"context"

	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type RepoI interface {
	CreateOrder(ctx context.Context, req *providerModel.OrderRequest) (*providerModel.OrderResponse, error)
	GetCatalog(ctx context.Context, providerID string) ([]*providerModel.CatalogResponse, error)
}
