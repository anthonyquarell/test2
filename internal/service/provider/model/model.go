package model

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
	Value   string
	Success bool
	Link    *string
	OrderID *string // номер заказа провайдера
	ID      *string // номер транзакции
}

type CancelRequest struct {
	ID                *string
	ProductID         *string
	ProviderProductID *string
	CustomerPhone     *string
}

type CancelResponse struct {
	Success bool
	ID      *string
	//ErrorMessage *string
}

type CatalogResponse struct {
	Name                      *string
	Desc                      *string
	ProviderProductID         *string
	ProviderExternalProductID *string
}
