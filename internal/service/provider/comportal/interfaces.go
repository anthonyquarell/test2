package comportal

import (
	"context"
	"github.com/mechta-market/e-product/internal/service/provider/comportal/model"
)

type RepoI interface {
	CreateOrder(ctx context.Context, req *model.OrderReq) (*model.OrderRep, error)
}
