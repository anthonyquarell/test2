package model

import (
	"github.com/mechta-market/e-product/internal/service/provider/comportal/constant"
	"github.com/mechta-market/e-product/internal/service/provider/comportal/model"
	"github.com/samber/lo"
	"strconv"
)

type CatalogProduct struct {
	Code        int    `json:"code"` // ProviderProductID
	Sku         string `json:"sku"`  // ProductID
	LicenseType string `json:"license_type"`
	Name        string `json:"name"` // Vendor
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

func DecodeSearchRep(product CatalogProduct) *model.Product {
	return &model.Product{
		ProductID:         &product.Sku,
		ProviderProductID: convertIntToStringPtr(product.Code),
		LicenseType:       &product.LicenseType,
	}
}

func EncodeOrderRequest(req *model.OrderReq, catalogProduct CatalogProduct, ptid string) *OrderReq {
	return &OrderReq{
		SKU:                  *req.ProductSKU,            // mdm data
		PromotionKey:         *req.PromotionKey,          // mdm data
		ProductCode:          *req.Code,                  // mdm data
		Vendor:               catalogProduct.Name,        // "Kaspersky" - из каталога
		LicenseType:          catalogProduct.LicenseType, // из каталога
		Count:                constant.Count,
		WayOfGettingDocument: constant.WayOfGettingDocument,
		PTID:                 ptid,
	}
}

func DecodeOrderResponse(rep OrderRep) *model.OrderRep {
	var value, link *string

	if len(rep.Data.Keys) > 0 {
		keyData := rep.Data.Keys[0]

		if len(keyData.Tokens) > 0 {
			value = &keyData.Tokens[0]
		}

		if len(keyData.Links) > 0 && keyData.Links[0] != "" {
			link = &keyData.Links[0]
		}
	}

	return &model.OrderRep{
		OrderID:       &rep.Data.OrderID,
		TransactionID: &rep.Data.TransactionID,
		Value:         value,
		Link:          link,
	}
}

func convertIntToStringPtr(value int) *string {
	if value == 0 {
		return nil
	}
	str := strconv.Itoa(value)
	return lo.ToPtr(str)
}
