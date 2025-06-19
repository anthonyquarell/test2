package key

import (
	"context"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	mdmModel "github.com/mechta-market/e-product/internal/service/mdm/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type KeyServiceI interface {
	List(ctx context.Context, pars *model.ListReq) ([]*model.Main, int64, error)
	Load(ctx context.Context, keys []*model.Key) ([]*model.Main, error)
	Get(ctx context.Context, ordId string, errNE bool) (*model.Main, bool, error)
	GetByOrderID(ctx context.Context, orderID string, errNE bool) (*model.Main, bool, error)
	Delete(ctx context.Context, keyID string) error
	Activate(ctx context.Context, obj *model.Key) (*model.Key, error)
	AddWithProvider(ctx context.Context, key *model.Key) (*model.Main, error)
}

type MdmServiceI interface {
	FindProduct(ctx context.Context, productID *string) (*mdmModel.Product, bool, error)
	ListCatalog(ctx context.Context, providerID string) ([]*mdmModel.CatalogProduct, bool, error)
}

type ProviderServiceI interface {
	CreateOrder(ctx context.Context, obj *providerModel.OrderRequest) (*providerModel.OrderResponse, error)
	CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error)
}
