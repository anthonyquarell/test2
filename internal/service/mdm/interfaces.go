package mdm

import (
	"context"

	"github.com/mechta-market/e-product/internal/service/mdm/model"
)

type RepoI interface {
	GetByProductID(ctx context.Context, productID string) (*model.Product, error)
}
