package key

import (
	"github.com/mechta-market/e-product/internal/domain/key/model"
	mdmModel "github.com/mechta-market/e-product/internal/service/mdm/model"
	"time"
)

type Key struct {
	model.Main
}

type RepItemT struct {
	ID            string
	ProviderID    string
	ProductID     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CustomerPhone string
	Status        string
	OrderID       string
}

func DecodeMdmRep(mdm *mdmModel.Product) model.Main {
	return model.Main{
		ProviderID:        deref(mdm.ProviderID),
		ProductID:         deref(mdm.ProductID),
		ProviderProductID: deref(mdm.ProviderProductID),
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
