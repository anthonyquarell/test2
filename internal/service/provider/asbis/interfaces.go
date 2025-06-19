package asbis

import (
	"context"
	"github.com/mechta-market/e-product/internal/service/provider/asbis/model"
)

type RepoI interface {
	CreateOrder(ctx context.Context, req *model.OrderReq) (*model.OrderRep, error)
	CancelOrder(ctx context.Context, obj *model.CancelReq) (*model.CancelRep, error)
}
