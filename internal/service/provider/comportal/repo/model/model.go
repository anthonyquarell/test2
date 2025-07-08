package model

import (
	"github.com/samber/lo"
	"strconv"

	"github.com/mechta-market/e-product/internal/service/provider/comportal/constant"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type CatalogProduct struct {
	Code        int    `json:"code"` // ExternalProviderProductID
	Sku         string `json:"sku"`  // ProviderProductID
	LicenseType string `json:"license_type"`
	Name        string `json:"name"` // Vendor
	Description string `json:"description"`
}

type CatalogRep struct {
	Data []*CatalogProduct `json:"data"`
}

type OrderReq struct {
	SKU                  string `json:"sku"`                    // "KL10410DAFS"
	Vendor               string `json:"vendor"`                 // "Kaspersky"
	ProductCode          string `json:"productCode"`            // "232113"
	LicenseType          string `json:"licenseType"`            // "Base"
	Count                string `json:"count"`                  // "1"
	WayOfGettingDocument string `json:"wayOfGettingDocument"`   // "2"
	PTID                 string `json:"ptid"`                   // UUID
	PromotionKey         string `json:"promotionKey,omitempty"` // optional
}

type OrderRep struct {
	Data OrderData `json:"data"`
}

type OrderData struct {
	Keys          []OrderKey `json:"keys"`
	OrderID       string     `json:"orderNumber"`
	TransactionID string     `json:"ptid"`
}

type OrderKey struct {
	Links  []string `json:"links"`
	Tokens []string `json:"tokens"`
}

func DecodeCatalogRep(product *CatalogProduct) *providerModel.CatalogResponse {
	return &providerModel.CatalogResponse{
		ProviderProductID:         &product.Sku,
		ProviderExternalProductID: convertIntToStringPtr(product.Code),
		Name:                      &product.Name,
		Desc:                      &product.Description,
	}
}

func EncodeOrderRequest(req *providerModel.OrderRequest, catalogProduct *CatalogProduct) *OrderReq {
	return &OrderReq{
		SKU:                  req.ProviderProductID,             // mdm data
		PromotionKey:         *req.PromotionKey,                 // mdm data
		ProductCode:          strconv.Itoa(catalogProduct.Code), // mdm data
		Vendor:               catalogProduct.Name,               // "Kaspersky" - из каталога
		LicenseType:          catalogProduct.LicenseType,        // из каталога
		Count:                constant.CountOfKeys,
		WayOfGettingDocument: constant.WayOfGettingDocument,
		PTID:                 "43c55461-49c6-4f9e-8cfd-b3bda7e0427c", // TODO: здесь генерировать providerModel.GenerateUUID()
	}
}

func DecodeOrderResponse(rep OrderRep) *providerModel.OrderResponse {
	result := &providerModel.OrderResponse{
		OrderID:       &rep.Data.OrderID,
		TransactionID: rep.Data.TransactionID,
	}

	if len(rep.Data.Keys) > 0 {
		keyData := rep.Data.Keys[0]

		if len(keyData.Tokens) > 0 {
			result.Value = keyData.Tokens[0]
		}

		if len(keyData.Links) > 0 && keyData.Links[0] != "" {
			result.Link = &keyData.Links[0]
		}
	}

	return result
}

func convertIntToStringPtr(value int) *string {
	if value == 0 {
		return nil
	}
	str := strconv.Itoa(value)
	return lo.ToPtr(str)
}
