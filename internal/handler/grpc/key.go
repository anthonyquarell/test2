package grpc

import (
	"context"
	"github.com/mechta-market/e-product/internal/handler/grpc/dto"
	keyUsecase "github.com/mechta-market/e-product/internal/usecase/key"
	"github.com/mechta-market/e-product/pkg/proto/common"
	e_product_v1 "github.com/mechta-market/e-product/pkg/proto/e_product"
	"github.com/samber/lo"
)

type Key struct {
	e_product_v1.UnsafeKeyServer
	keyUsecase *keyUsecase.Usecase
}

func NewKey(keyUsecase *keyUsecase.Usecase) *Key {
	return &Key{
		keyUsecase: keyUsecase,
	}
}

func (h *Key) Load(ctx context.Context, req *e_product_v1.LoadKeyReq) (*e_product_v1.LoadKeyRep, error) {
	loadReq := dto.DecodeLoadKeyReq(req)

	result, err := h.keyUsecase.Load(ctx, loadReq)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return dto.EncodeLoadKeyRep(result), nil
}

func (h *Key) List(ctx context.Context, req *e_product_v1.KeyListReq) (*e_product_v1.KeyListRep, error) {
	if req.ListParams == nil {
		req.ListParams = &common.ListParamsSt{}
	}

	items, tCount, err := h.keyUsecase.List(ctx, dto.DecodeKeyListReq(req))
	if err != nil {
		return nil, err
	}

	return &e_product_v1.KeyListRep{
		PaginationInfo: &common.PaginationInfoSt{
			Page:       req.ListParams.Page,
			PageSize:   req.ListParams.PageSize,
			TotalCount: tCount,
		},
		Keys: lo.Map(items, dto.EncodeKeyMain),
	}, nil
}

func (h *Key) Get(ctx context.Context, req *e_product_v1.KeyGetReq) (*e_product_v1.KeyGetRep, error) {
	result, err := h.keyUsecase.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return dto.EncodeKeyGetResponse(result), nil
}

func (h *Key) Activate(ctx context.Context, req *e_product_v1.KeyActivateReq) (*e_product_v1.KeyActivateRep, error) {
	result, err := h.keyUsecase.Activate(ctx, dto.DecodeKeyActivateReq1(req))
	if err != nil {
		return nil, err
	}

	return dto.EncodeKeyActivateResp(result), nil
}

func (h *Key) GetByProductID(ctx context.Context, req *e_product_v1.GetByProductIDReq) (*e_product_v1.GetByProductIDRep, error) {
	result, err := h.keyUsecase.GetMdmProduct(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}

	return dto.EncodeGetByProductIDRep(result), nil
}

func (h *Key) Catalog(ctx context.Context, req *e_product_v1.GetCatalogReq) (*e_product_v1.GetCatalogRep, error) {
	result, err := h.keyUsecase.GetCatalog(ctx, req.ProviderId)
	if err != nil {
		return nil, err
	}

	return &e_product_v1.GetCatalogRep{
		Items: lo.Map(result, dto.EncodeCatalogRep),
	}, nil
}

func (h *Key) CreateOrder(ctx context.Context, req *e_product_v1.CreateOrderReq) (*e_product_v1.CreateOrderRep, error) {
	result, err := h.keyUsecase.ActivateByProvider(ctx, req.ProductId, req.OrderId, req.CustomerPhone)
	if err != nil {
		// если провайдер недоступен, идет поиск свободного ключа из пула
		poolResult, _ := h.keyUsecase.Activate(ctx, dto.DecodeKeyActivateReq2(req))
		if poolResult == nil {
			return nil, err
		}

		return dto.EncodeCreateOrderRepPool(poolResult), nil
	}

	return dto.EncodeCreateOrderRep(result), nil
}

func (h *Key) CancelOrder(ctx context.Context, req *e_product_v1.CancelOrderReq) (*e_product_v1.CancelOrderRep, error) {
	result, err := h.keyUsecase.CancelByProvider(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return dto.EncodeCancelOrderRep(result), nil
}
