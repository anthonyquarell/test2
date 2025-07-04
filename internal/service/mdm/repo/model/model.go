package model

import (
	"github.com/samber/lo"
	"strconv"

	"github.com/mechta-market/e-product/internal/service/mdm/model"
)

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
	ID     string        `json:"_id"`
	Source ProductSource `json:"_source"`
}

type ProductSource struct {
	Provider Provider `json:"provider"`
	NameI18N NameI18N `json:"name_i18n"`
	Slug     string   `json:"slug"`
}

type Provider struct {
	ExternalNumber string `json:"external_number"`
	ProviderID     string `json:"provider_id"`
	ExternalID     int    `json:"external_id"`
	PromotionKey   string `json:"promotion_key,omitempty"` // только для пакетов услуг
}

type NameI18N struct {
	Ru string `json:"ru"`
}

type SearchReq struct {
	Query SearchQuery `json:"query"`
	Size  int         `json:"size"`
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

func DecodeSearchRep(provider Provider, hit HitRecord) *model.Product {
	return &model.Product{
		ProviderID:         provider.ProviderID,
		ProductID:          hit.ID,
		ProviderProductID:  provider.ExternalNumber,
		ProviderExternalID: convertIntToStringPtr(provider.ExternalID),
		PromotionKey:       &provider.PromotionKey,
	}
}

func convertIntToStringPtr(value int) *string {
	if value == 0 {
		return nil
	}
	str := strconv.Itoa(value)
	return lo.ToPtr(str)
}
