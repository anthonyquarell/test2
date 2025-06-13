package repo

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	clientComportal "github.com/mechta-market/e-product/internal/client/comportal"
	"github.com/mechta-market/e-product/internal/service/provider/comportal/model"
	repoModel "github.com/mechta-market/e-product/internal/service/provider/comportal/repo/model"
)

type Repo struct {
	client clientComportal.Client
}

func New(client clientComportal.Client) *Repo {
	return &Repo{client: client}
}

//func (r *Repo) getProduct(ctx context.Context, sku string) (*model.Product, error) {
//	catalogRep := &repoModel.CatalogRep{}
//
//	_, err := r.client.Send(ctx, &clientComportal.SendReq{
//		Method: "GET",
//		Path:   "api/Catalog/Products",
//		RepObj: catalogRep,
//		Params: map[string]string{
//			"imagesDisable": "true",
//		},
//	})
//	if err != nil {
//		return nil, fmt.Errorf("failed to get catalog: %w", err)
//	}
//
//	// match with mdm by sku (product_id)
//	for _, product := range catalogRep.Data {
//		if product.Sku == sku {
//			return repoModel.DecodeSearchRep(*product), nil
//		}
//	}
//
//	return nil, fmt.Errorf("product with SKU '%s' not found in Comportal catalog", sku)
//}

func (r *Repo) getProduct(ctx context.Context, sku string) (*repoModel.CatalogProduct, error) {
	catalogRep := &repoModel.CatalogRep{}

	_, err := r.client.Send(ctx, &clientComportal.SendReq{
		Method: "GET",
		Path:   "api/Catalog/Products",
		RepObj: catalogRep,
		Params: map[string]string{
			"imagesDisable": "true",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %w", err)
	}

	for _, product := range catalogRep.Data {
		if product.Sku == sku {
			return product, nil
		}
	}

	return nil, fmt.Errorf("product with SKU '%s' not found in catalog", sku)
}

func (r *Repo) CreateOrder(ctx context.Context, obj *model.OrderReq) (*model.OrderRep, error) {
	catalogProduct, err := r.getProduct(ctx, *obj.ProductSKU)
	if err != nil {
		return nil, fmt.Errorf("repo.getProduct: %w", err)
	}

	apiReq := repoModel.EncodeOrderRequest(obj, *catalogProduct, uuid.New().String())

	apiResp := &repoModel.OrderRep{}

	_, err = r.client.Send(ctx, &clientComportal.SendReq{
		Method: "POST",
		Path:   "api/Order",
		ReqObj: apiReq,
		RepObj: apiResp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return repoModel.DecodeOrderResponse(*apiResp), nil
}
