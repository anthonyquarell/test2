package grpc

import (
	"context"
	"github.com/mechta-market/e-product/internal/handler/grpc/dto"
	keyUsecase "github.com/mechta-market/e-product/internal/usecase/key"
	"github.com/mechta-market/e-product/pkg/proto/common"
	electronic_product_v1 "github.com/mechta-market/e-product/pkg/proto/electronic_product"
	"github.com/samber/lo"
)

type Key struct {
	electronic_product_v1.UnsafeKeyServer
	keyUsecase *keyUsecase.Usecase
}

func NewKey(keyUsecase *keyUsecase.Usecase) *Key {
	return &Key{
		keyUsecase: keyUsecase,
	}
}

func (h *Key) Load(ctx context.Context, req *electronic_product_v1.LoadKeyReq) (*electronic_product_v1.LoadKeyRep, error) {
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

func (h *Key) List(ctx context.Context, req *electronic_product_v1.KeyListReq) (*electronic_product_v1.KeyListRep, error) {
	if req.ListParams == nil {
		req.ListParams = &common.ListParamsSt{}
	}

	items, tCount, err := h.keyUsecase.List(ctx, dto.DecodeKeyListReq(req))
	if err != nil {
		return nil, err
	}

	return &electronic_product_v1.KeyListRep{
		PaginationInfo: &common.PaginationInfoSt{
			Page:       req.ListParams.Page,
			PageSize:   req.ListParams.PageSize,
			TotalCount: tCount,
		},
		Keys: lo.Map(items, dto.EncodeKeyMain),
	}, nil
}

func (h *Key) Get(ctx context.Context, req *electronic_product_v1.KeyGetReq) (*electronic_product_v1.KeyGetRep, error) {
	result, err := h.keyUsecase.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return dto.EncodeKeyGetResponse(result), nil
}

func (h *Key) Activate(ctx context.Context, req *electronic_product_v1.KeyActivateReq) (*electronic_product_v1.KeyActivateRep, error) {
	result, err := h.keyUsecase.Activate(ctx, dto.DecodeKeyActivateReq(req))
	if err != nil {
		return nil, err
	}

	return dto.EncodeKeyActivateResp(result), nil
}

func (h *Key) GetByProductID(ctx context.Context, req *electronic_product_v1.GetByProductIDReq) (*electronic_product_v1.GetByProductIDRep, error) {
	result, err := h.keyUsecase.GetMdmProduct(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}

	return dto.EncodeGetByProductIDRep(result), nil
}
