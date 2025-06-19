package dto

import (
	"github.com/mechta-market/e-product/internal/constant"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	usecase "github.com/mechta-market/e-product/internal/usecase/key"
	e_product_v1 "github.com/mechta-market/e-product/pkg/proto/e_product"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DecodeKeyListReq(v *e_product_v1.KeyListReq) *model.ListReq {
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

func DecodeLoadKeyReq(v *e_product_v1.LoadKeyReq) []*model.Key {
	return lo.Map(v.Keys, func(item *e_product_v1.KeyItem, _ int) *model.Key {
		return &model.Key{
			ProductID: &item.ProductId,
			Value:     &item.Value,
		}
	})
}

// DecodeKeyActivateReq1 при ручной активации через /key/activate
func DecodeKeyActivateReq1(v *e_product_v1.KeyActivateReq) *model.Key {
	return &model.Key{
		ProductID:     &v.ProductId,
		OrderID:       &v.OrderId,
		CustomerPhone: &v.CustomerPhone,
	}
}

// DecodeKeyActivateReq2 при активации если активация по провайдеру недоступна
func DecodeKeyActivateReq2(v *e_product_v1.CreateOrderReq) *model.Key {
	return &model.Key{
		ProductID:     &v.ProductId,
		OrderID:       &v.OrderId,
		CustomerPhone: &v.CustomerPhone,
	}
}

func EncodeKeyActivateResp(resp *model.Key) *e_product_v1.KeyActivateRep {
	if resp == nil {
		return nil
	}

	return &e_product_v1.KeyActivateRep{
		Value: *resp.Value,
	}
}

func EncodeLoadKeyRep(v []*usecase.Key) *e_product_v1.LoadKeyRep {
	return &e_product_v1.LoadKeyRep{
		Keys: lo.Map(v, func(key *usecase.Key, _ int) *e_product_v1.KeyResponseItem {
			return EncodeRepItem(key)
		}),
	}
}

func EncodeKeyGetResponse(v *usecase.Key) *e_product_v1.KeyGetRep {
	response := &e_product_v1.KeyGetRep{}

	if v != nil {
		response.Key = EncodeUsecaseKey(v, 0)
	}

	return response
}

func EncodeUsecaseKey(v *usecase.Key, _ int) *e_product_v1.KeyResponseItem {
	return EncodeKeyMain(&v.Main, 0)
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

func EncodeRepItem(v *usecase.Key) *e_product_v1.KeyResponseItem {
	if v == nil {
		return nil
	}

	return &e_product_v1.KeyResponseItem{
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

func DecodeGetByProductIDReq(v *e_product_v1.GetByProductIDReq) *model.Key {
	return &model.Key{
		ProductID: &v.ProductId,
	}
}

func EncodeGetByProductIDRep(v *usecase.Key) *e_product_v1.GetByProductIDRep {
	if v == nil {
		return nil
	}

	return &e_product_v1.GetByProductIDRep{
		ProductId:                 v.Main.ProductID,
		ProviderId:                v.Main.ProviderID,
		ProviderProductId:         v.Main.ProviderProductID,
		PromotionKey:              v.Main.PromotionKey,
		ProviderExternalProductId: v.Main.ProviderExternalProductID,
	}
}

func EncodeCatalogRep(v *usecase.CatalogProductT, _ int) *e_product_v1.CatalogItem {
	if v == nil {
		return nil
	}

	return &e_product_v1.CatalogItem{
		ProductId:         v.ProductID,
		ProviderProductId: v.ProviderProductID,
		Slug:              v.Slug,
		Name:              v.Name,
	}
}

func EncodeCreateOrderRep(v *usecase.Key) *e_product_v1.CreateOrderRep {
	if v == nil {
		return nil
	}

	return &e_product_v1.CreateOrderRep{
		Id:                &v.ID,
		ProviderId:        &v.ProviderID,
		ProductId:         &v.ProductID,
		ProviderProductId: &v.ProviderProductID,
		OrderId:           &v.OrderID,
		ProviderOrderId:   &v.ProviderOrderID,
		CustomerPhone:     &v.CustomerPhone,
		Value:             v.Value,
		Status:            lo.ToPtr(mapStatusToProtoEnum(v.Status)),
	}
}

// EncodeCreateOrderRepPool возвращает ключ, если не удалось получить его из провайдера
func EncodeCreateOrderRepPool(v *model.Key) *e_product_v1.CreateOrderRep {
	if v == nil {
		return nil
	}

	return &e_product_v1.CreateOrderRep{
		Value: *v.Value,
	}
}

func EncodeCancelOrderRep(v *string) *e_product_v1.CancelOrderRep {
	if v == nil {
		return nil
	}

	return &e_product_v1.CancelOrderRep{
		OrderId: *v,
	}
}

//

func mapStatusToProtoEnum(status string) e_product_v1.KeyStatus {
	switch status {
	case constant.KeyStatusNew:
		return e_product_v1.KeyStatus_new
	case constant.KeyStatusActivated:
		return e_product_v1.KeyStatus_activated
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
	default:
		return nil
	}

	return &s
}
