package megogo

import (
	"context"
	"github.com/mechta-market/e-product/internal/service/provider/megogo/model"
)

type RepoI interface {
	CreateOrder(ctx context.Context, req *model.OrderReq) (*model.OrderRep, error)
	CancelOrder(ctx context.Context, req *model.CancelReq) (*model.CancelRep, error)
}
