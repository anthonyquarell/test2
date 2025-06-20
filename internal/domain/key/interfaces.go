package key

import (
	"context"
	"github.com/mechta-market/e-product/internal/domain/key/model"
)

type RepoDbI interface {
	CreateIfNotExist(ctx context.Context, obj *model.Key) (finalError error)
	List(ctx context.Context, pars *model.ListReq) (_ []*model.Main, _ int64, finalError error)
	Get(ctx context.Context, id string) (_ *model.Main, _ bool, finalError error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Main, bool, error)
	Delete(ctx context.Context, id string) (finalError error)
	ExistsByValue(ctx context.Context, value string) (bool, error)
	Update(ctx context.Context, obj *model.Key) (finalError error)
	GetForActivate(ctx context.Context, productID string) (*model.Main, error)
	CreateWithProvider(ctx context.Context, obj *model.Key) (finalError error)
}
