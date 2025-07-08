package model

import (
	"time"

	commonModel "github.com/mechta-market/e-product/internal/domain/common/model"
)

type Main struct {
	ID                        string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	ProviderID                string
	ProductID                 string
	Value                     string
	Status                    string
	CustomerPhone             string
	OrderID                   string
	ProviderProductID         string
	ProviderExternalProductID string
	PromotionKey              string
	ProviderOrderID           string
	ProviderTransactionID     string
}

type ListReq struct {
	commonModel.ListParams

	ProviderID *string
	Status     *string
	OrderID    *string
	ProductID  *string
}

type Edit struct {
	ID                        *string
	UpdatedAt                 *time.Time
	ProviderID                *string
	ProductID                 *string
	Value                     *string
	Status                    *string
	CustomerPhone             *string
	OrderID                   *string
	ProviderProductID         *string
	ProviderExternalProductID *string
	ProviderOrderID           *string
	ProviderTransactionID     *string
}
