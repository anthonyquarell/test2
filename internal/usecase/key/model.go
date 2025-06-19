package key

import (
	"github.com/mechta-market/e-product/internal/domain/key/model"
	mdmModel "github.com/mechta-market/e-product/internal/service/mdm/model"
	"github.com/samber/lo"
	"time"
)

type Key struct {
	model.Main
}

type CatalogProductT struct {
	ProductID         string
	ProviderProductID string
	Slug              string
	Name              string
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
		ProviderID:                deref(mdm.ProviderID),
		ProductID:                 deref(mdm.ProductID),
		ProviderProductID:         deref(mdm.ProviderProductID),
		ProviderExternalProductID: deref(mdm.ProviderExternalID),
	}
}

func DecodeMdmCatalog(mdmProducts []*mdmModel.CatalogProduct) []*CatalogProductT {
	return lo.Map(mdmProducts, func(p *mdmModel.CatalogProduct, _ int) *CatalogProductT {
		return &CatalogProductT{
			ProductID:         deref(p.ProductID),
			ProviderProductID: deref(p.ProviderProductID),
			Slug:              deref(p.Slug),
			Name:              deref(p.Name),
		}
	})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrOrDefault(ptr *string, defaultVal string) *string {
	if ptr != nil {
		return ptr
	}
	return &defaultVal
}
