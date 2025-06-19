package model

type OrderReq struct {
	ProductNumber *string // ProductID
	TermNumber    *string // ProviderProductID
}

type OrderRep struct {
	Success  bool
	Value    *string
	Link     *string
	Receipt  *string
	ErrorMsg string
	ID       *string
}

type CancelReq struct {
	ProductNumber         *string // ProductID
	TermNumber            *string // ProviderProductID
	OriginalTransactionID *string
}

type CancelRep struct {
	Success               bool
	ErrorMessage          string
	OriginalTransactionID string
}
