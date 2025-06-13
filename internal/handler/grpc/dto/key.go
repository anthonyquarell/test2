package dto

import (
	"github.com/mechta-market/e-product/internal/constant"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	usecase "github.com/mechta-market/e-product/internal/usecase/key"
	electronic_product_v1 "github.com/mechta-market/e-product/pkg/proto/electronic_product"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DecodeKeyListReq(v *electronic_product_v1.KeyListReq) *model.ListReq {
	result := &model.ListReq{
		ListParams: DecodeListParams(v.ListParams),
		ProviderID: v.ProviderId,
		OrderID:    v.OrderId,
	}

	if v.Status != nil {
		result.Status = mapProtoEnumToStatus(*v.Status)
	}

	return result
}

func DecodeLoadKeyReq(v *electronic_product_v1.LoadKeyReq) []*model.Key {
	return lo.Map(v.Keys, func(item *electronic_product_v1.KeyItem, _ int) *model.Key {
		return &model.Key{
			ProductID: &item.ProductId,
			Value:     &item.Value,
		}
	})
}

func DecodeKeyActivateReq(v *electronic_product_v1.KeyActivateReq) *model.Key {
	return &model.Key{
		ProductID:     &v.ProductId,
		OrderID:       &v.OrderId,
		CustomerPhone: &v.CustomerPhone,
	}
}

func EncodeKeyActivateResp(resp *model.Key) *electronic_product_v1.KeyActivateRep {
	if resp == nil {
		return nil
	}

	return &electronic_product_v1.KeyActivateRep{
		Value: *resp.Value,
	}
}

func EncodeLoadKeyRep(v []*usecase.Key) *electronic_product_v1.LoadKeyRep {
	return &electronic_product_v1.LoadKeyRep{
		Keys: lo.Map(v, func(key *usecase.Key, _ int) *electronic_product_v1.KeyResponseItem {
			return EncodeRepItem(key)
		}),
	}
}

func EncodeKeyGetResponse(v *usecase.Key) *electronic_product_v1.KeyGetRep {
	response := &electronic_product_v1.KeyGetRep{}

	if v != nil {
		response.Key = EncodeUsecaseKey(v, 0)
	}

	return response
}

func EncodeUsecaseKey(v *usecase.Key, _ int) *electronic_product_v1.KeyResponseItem {
	return EncodeKeyMain(&v.Main, 0)
}

func EncodeKeyMain(v *model.Main, _ int) *electronic_product_v1.KeyResponseItem {
	if v == nil {
		return nil
	}

	return &electronic_product_v1.KeyResponseItem{
		Id:            v.ID,
		ProviderId:    v.ProviderID,
		ProductId:     v.ProductID,
		CreatedAt:     timestamppb.New(v.CreatedAt),
		UpdatedAt:     timestamppb.New(v.UpdatedAt),
		CustomerPhone: v.CustomerPhone,
		Status:        mapStatusToProtoEnum(v.Status),
		OrderId:       v.OrderID,
	}
}

func EncodeRepItem(v *usecase.Key) *electronic_product_v1.KeyResponseItem {
	if v == nil {
		return nil
	}

	return &electronic_product_v1.KeyResponseItem{
		Id:            v.ID,
		ProviderId:    v.ProviderID,
		ProductId:     v.ProductID,
		CreatedAt:     timestamppb.New(v.CreatedAt),
		UpdatedAt:     timestamppb.New(v.UpdatedAt),
		CustomerPhone: v.CustomerPhone,
		Status:        mapStatusToProtoEnum(v.Status),
		OrderId:       v.OrderID,
	}
}

// mdm methods

func DecodeGetByProductIDReq(v *electronic_product_v1.GetByProductIDReq) *model.Key {
	return &model.Key{
		ProductID: &v.ProductId,
	}
}

func EncodeGetByProductIDRep(v *usecase.Key) *electronic_product_v1.GetByProductIDRep {
	if v == nil {
		return nil
	}

	return &electronic_product_v1.GetByProductIDRep{
		ProductId:         v.Main.ProductID,
		ProviderId:        v.Main.ProviderID,
		ProviderProductId: v.Main.ProviderProductID,
		PromotionKey:      v.Main.PromotionKey,
	}
}

//

func mapStatusToProtoEnum(status string) electronic_product_v1.KeyStatus {
	switch status {
	case constant.KeyStatusNew:
		return electronic_product_v1.KeyStatus_new
	case constant.KeyStatusActivated:
		return electronic_product_v1.KeyStatus_activated
	default:
		return electronic_product_v1.KeyStatus_new
	}
}

func mapProtoEnumToStatus(status electronic_product_v1.KeyStatus) *string {
	var s string

	switch status {
	case electronic_product_v1.KeyStatus_new:
		s = constant.KeyStatusNew
	case electronic_product_v1.KeyStatus_activated:
		s = constant.KeyStatusActivated
	default:
		return nil
	}

	return &s
}
