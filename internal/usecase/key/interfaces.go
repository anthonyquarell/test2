package key

import (
	"context"

	"github.com/mechta-market/e-product/internal/domain/key/model"
	mdmModel "github.com/mechta-market/e-product/internal/service/mdm/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type KeyServiceI interface {
	List(ctx context.Context, pars *model.ListReq) ([]*model.Main, int64, error)
	Get(ctx context.Context, ID string, errNE bool) (*model.Main, bool, error)
	GetByOrderID(ctx context.Context, orderID string, errNE bool) (*model.Main, bool, error)
	GetByValue(ctx context.Context, value string) (_ *model.Main, finalError error)
	Update(ctx context.Context, edit *model.Edit) error
	Create(ctx context.Context, obj *model.Edit) (string, error)
}

type MdmServiceI interface {
	FindProduct(ctx context.Context, productID *string) (*mdmModel.Product, bool, error)
}

type ProviderServiceI interface {
	CreateOrder(ctx context.Context, obj *providerModel.OrderRequest) (*providerModel.OrderResponse, error)
	CancelOrder(ctx context.Context, req *providerModel.CancelRequest) (*providerModel.CancelResponse, error)
	ListCatalog(ctx context.Context, providerID string) ([]*providerModel.CatalogResponse, error)
	SupportsPool() bool
}
