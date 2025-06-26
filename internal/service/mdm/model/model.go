package model

type Product struct {
	ProductID          string
	ProviderID         string
	ProviderProductID  string
	PromotionKey       *string
	ProviderExternalID *string
}

type CatalogProduct struct {
	ProductID         *string
	ProviderProductID *string
	Slug              *string
	Name              *string
}
