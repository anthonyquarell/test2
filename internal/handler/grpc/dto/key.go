package dto

import (
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mechta-market/e-product/internal/constant"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
	e_product_v1 "github.com/mechta-market/e-product/pkg/proto/e_product"
)

func DecodeKeyListReq(v *e_product_v1.KeyListReq) *model.ListReq {
	result := &model.ListReq{
		ListParams: DecodeListParams(v.ListParams),
		ProviderID: v.ProviderId,
		OrderID:    v.OrderId,
		ProductID:  v.ProductId,
	}

	if v.Status != nil {
		result.Status = mapProtoEnumToStatus(*v.Status)
	}

	return result
}

func DecodeLoadKeyReq(v *e_product_v1.LoadKeyReq) []*model.Edit {
	return lo.Map(v.Keys, func(item *e_product_v1.KeyItem, _ int) *model.Edit {
		return &model.Edit{
			ProductID: &item.ProductId,
			Value:     &item.Value,
		}
	})
}

func EncodeKeyMain(v *model.Main, _ int) *e_product_v1.KeyResponseItem {
	if v == nil {
		return nil
	}

	return &e_product_v1.KeyResponseItem{
		Id:                v.ID,
		ProviderId:        v.ProviderID,
		ProductId:         v.ProductID,
		CreatedAt:         timestamppb.New(v.CreatedAt),
		UpdatedAt:         timestamppb.New(v.UpdatedAt),
		CustomerPhone:     v.CustomerPhone,
		Status:            mapStatusToProtoEnum(v.Status),
		OrderId:           v.OrderID,
		ProviderProductId: v.ProviderProductID,
		ProviderOrderId:   v.ProviderOrderID,
	}
}

func EncodeCatalogRep(v *providerModel.CatalogResponse, _ int) *e_product_v1.CatalogItem {
	if v == nil {
		return nil
	}

	return &e_product_v1.CatalogItem{
		ProviderProductId:         lo.FromPtrOr(v.ProviderProductID, ""),
		ProviderExternalProductId: lo.FromPtrOr(v.ProviderExternalProductID, ""),
		Name:                      lo.FromPtrOr(v.Name, ""),
		Desc:                      lo.FromPtrOr(v.Desc, ""),
	}
}

func EncodeActivateRep(v *string) *e_product_v1.KeyActivateRep {
	if v == nil {
		return nil
	}

	return &e_product_v1.KeyActivateRep{
		Value: *v,
	}
}

func EncodeCancelRep(v *string) *e_product_v1.KeyCancelRep {
	if v == nil {
		return nil
	}

	return &e_product_v1.KeyCancelRep{
		Id: *v,
	}
}

//

func mapStatusToProtoEnum(status string) e_product_v1.KeyStatus {
	switch status {
	case constant.KeyStatusNew:
		return e_product_v1.KeyStatus_new
	case constant.KeyStatusActivated:
		return e_product_v1.KeyStatus_activated
	case constant.KeyStatusCancelled:
		return e_product_v1.KeyStatus_cancelled
	default:
		return e_product_v1.KeyStatus_new
	}
}

func mapProtoEnumToStatus(status e_product_v1.KeyStatus) *string {
	var s string

	switch status {
	case e_product_v1.KeyStatus_new:
		s = constant.KeyStatusNew
	case e_product_v1.KeyStatus_activated:
		s = constant.KeyStatusActivated
	case e_product_v1.KeyStatus_cancelled:
		s = constant.KeyStatusCancelled
	default:
		return nil
	}

	return &s
}
