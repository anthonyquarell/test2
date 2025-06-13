package model

import (
	"github.com/mechta-market/e-product/internal/service/mdm/model"
	"github.com/samber/lo"
	"strconv"
)

type MDMProduct struct {
	ID             string            `json:"id"`
	Code           string            `json:"code"`
	ExternalNumber string            `json:"external_number"` // product_id
	ProviderID     string            `json:"provider_id"`     // provider_id
	ExternalID     string            `json:"external_id"`     // provider_product_id
	ServiceType    int               `json:"service_type"`
	Type           int               `json:"type"`
	Published      bool              `json:"published"`
	Name           map[string]string `json:"name_i18n"`
}

type SearchRep struct {
	Took int  `json:"took"`
	Hits Hits `json:"hits"`
}

type Hits struct {
	Total Total       `json:"total"`
	Hits  []HitRecord `json:"hits"`
}

type Total struct {
	Value int `json:"value"`
}

type HitRecord struct {
	Source ProductSource `json:"_source"`
}

type ProductSource struct {
	Provider Provider `json:"provider"`
}

type Provider struct {
	ExternalNumber string `json:"external_number"`
	ProviderID     string `json:"provider_id"`
	ExternalID     int    `json:"external_id"`
	PromotionKey   string `json:"promotion_key"`
}

type SearchReq struct {
	Query SearchQuery `json:"query"`
}

type SearchQuery struct {
	Bool BoolQuery `json:"bool"`
}

type BoolQuery struct {
	Must []TermQuery `json:"must"`
}

type TermQuery struct {
	Term map[string]interface{} `json:"term"`
}

func DecodeSearchRep(provider Provider) *model.Product {
	return &model.Product{
		ProviderID:        &provider.ProviderID,
		ProductID:         &provider.ExternalNumber,
		ProviderProductID: convertIntToStringPtr(provider.ExternalID),
		PromotionKey:      &provider.PromotionKey,
	}
}

func convertIntToStringPtr(value int) *string {
	if value == 0 {
		return nil
	}
	str := strconv.Itoa(value)
	return lo.ToPtr(str)
}
