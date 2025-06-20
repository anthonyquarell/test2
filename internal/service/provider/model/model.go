package model

type Product struct {
	ID                *string
	ProviderID        *string
	ProductID         *string
	ProviderProductID *string
	LicenseType       *string
	Count             *int32
}

type OrderRequest struct {
	ProductID                 *string
	ProviderProductID         *string
	ProviderExternalProductID *string
	PromotionKey              *string
	CustomerPhone             *string
}

type OrderResponse struct {
	ID                *string
	OrderID           *string
	ProviderID        *string
	ProductID         *string
	ProviderProductID *string
	Value             *string
	Link              *string
}

type CancelRequest struct {
	ID                *string
	ProductID         *string
	ProviderProductID *string
	CustomerPhone     *string
}

type CancelResponse struct {
	ID           *string
	Success      bool
	ErrorMessage *string
}
