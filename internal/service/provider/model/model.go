package model

import "github.com/google/uuid"

type OrderRequest struct {
	ProviderID        string
	ProductID         string
	ProviderProductID string
	CustomerPhone     string
	OrderID           string
	// для comportal
	ProviderExternalProductID *string
	PromotionKey              *string
}

type OrderResponse struct {
	Value         string
	Success       bool
	TransactionID string // номер транзакции
	Link          *string
	OrderID       *string // номер заказа провайдера
}

type CancelRequest struct {
	CancelID          *string
	ProductID         *string
	ProviderProductID *string
	CustomerPhone     *string
}

type CancelResponse struct {
	Success       bool
	TransactionID *string
	//ErrorMessage *string
}

type CatalogResponse struct {
	Name                      *string
	Desc                      *string
	ProviderProductID         *string
	ProviderExternalProductID *string
}

func GenerateUUID() string {
	return uuid.New().String()
}
