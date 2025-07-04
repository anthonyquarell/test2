package key

import (
	"context"

	"github.com/mechta-market/e-product/internal/domain/key/model"
)

type RepoDbI interface {
	List(ctx context.Context, pars *model.ListReq) (_ []*model.Main, _ int64, finalError error)
	Get(ctx context.Context, id string) (_ *model.Main, _ bool, finalError error)
	GetByOrderID(ctx context.Context, orderID string) (_ *model.Main, _ bool, finalError error)
	GetByValue(ctx context.Context, value string) (_ *model.Main, finalError error)
	Update(ctx context.Context, obj *model.Edit) (finalError error)
	Create(ctx context.Context, obj *model.Edit) (_ string, finalError error)
}
