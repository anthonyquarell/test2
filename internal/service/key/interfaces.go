package key

import (
	"context"
	"github.com/mechta-market/e-product/internal/domain/key/model"
)

type KeyServiceI interface {
	CreateIfNotExist(ctx context.Context, obj *model.Key) error
}
