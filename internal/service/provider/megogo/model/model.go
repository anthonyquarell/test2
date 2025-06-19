package model

type OrderReq struct {
	ProviderProductID string
	CustomerPhone     string
}

type OrderRep struct {
	ProviderProductID string
	Success           bool
	Value             string
}

type CancelReq struct {
	ProviderProductID string
	CustomerPhone     string
}

type CancelRep struct {
	Success bool
}
