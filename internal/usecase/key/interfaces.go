package key

import (
	"context"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	mdmModel "github.com/mechta-market/e-product/internal/service/mdm/model"
	"github.com/mechta-market/e-product/internal/service/provider"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type KeyServiceI interface {
	List(ctx context.Context, pars *model.ListReq) ([]*model.Main, int64, error)
	Load(ctx context.Context, keys []*model.Key) ([]*model.Main, error)
	Get(ctx context.Context, ordId string, errNE bool) (*model.Main, bool, error)
	Activate(ctx context.Context, obj *model.Key) (*model.Key, error)
	AddWithProvider(ctx context.Context, key *model.Key) (*model.Main, error)
}

type MdmServiceI interface {
	FindProduct(ctx context.Context, productID *string) (*mdmModel.Product, bool, error)
}

type ProviderServiceI interface {
	CreateOrder(ctx context.Context, obj *providerModel.OrderRequest) (*providerModel.OrderResponse, error)
}

type ProviderServiceMngI interface {
	GetProvider(providerID string) (provider.ServiceI, error)
	ListProviders() []provider.ServiceI
	RegisterProvider(providerID string, i provider.ServiceI) error
}
