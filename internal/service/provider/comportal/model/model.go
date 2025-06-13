package model

type Product struct {
	ProviderID        *string
	ProductID         *string
	ProviderProductID *string
	LicenseType       *string
	Count             *int
}

type CatalogReq struct {
	ProviderID  *string
	ProductType *string
}

type OrderReq struct {
	ProductSKU   *string
	Code         *string
	Vendor       *string
	PromotionKey *string
}

type OrderRep struct {
	Value         *string
	Link          *string
	OrderID       *string
	TransactionID *string
}
