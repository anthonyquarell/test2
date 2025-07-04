package grpc

import (
	"context"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/mechta-market/e-product/internal/handler/grpc/dto"
	keyUsecase "github.com/mechta-market/e-product/internal/usecase/key"
	"github.com/mechta-market/e-product/pkg/proto/common"
	e_product_v1 "github.com/mechta-market/e-product/pkg/proto/e_product"
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

func (h *Key) Load(ctx context.Context, req *e_product_v1.LoadKeyReq) (*emptypb.Empty, error) {
	loadReq := dto.DecodeLoadKeyReq(req)

	err := h.keyUsecase.Load(ctx, loadReq)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
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

func (h *Key) Get(ctx context.Context, req *e_product_v1.KeyGetReq) (*e_product_v1.KeyResponseItem, error) {
	result, err := h.keyUsecase.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return dto.EncodeKeyMain(result, 0), nil
}

func (h *Key) Activate(ctx context.Context, req *e_product_v1.KeyActivateReq) (*e_product_v1.KeyActivateRep, error) {
	result, err := h.keyUsecase.Activate(ctx, req.ProductId, req.OrderId, req.CustomerPhone)
	if err != nil {
		return nil, err
	}

	return dto.EncodeActivateRep(result), nil
}

func (h *Key) Cancel(ctx context.Context, req *e_product_v1.KeyCancelReq) (*e_product_v1.KeyCancelRep, error) {
	result, err := h.keyUsecase.Cancel(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return dto.EncodeCancelRep(result), nil
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
