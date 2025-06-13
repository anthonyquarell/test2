package model

import (
	"github.com/mechta-market/e-product/internal/constant"
	commonModel "github.com/mechta-market/e-product/internal/domain/common/model"
	"time"
)

type Main struct {
	ID                string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ProviderID        string
	ProductID         string
	Value             string
	Status            string
	CustomerPhone     string
	OrderID           string
	ProviderProductID string
	PromotionKey      string

	Computed Computed
}

type ListReq struct {
	commonModel.ListParams

	ProviderID *string
	Status     *string
	OrderID    *string
}

type Key struct {
	ID                *string
	UpdatedAt         *time.Time
	ProviderID        *string
	ProductID         *string
	Value             *string
	Status            *string
	CustomerPhone     *string
	OrderID           *string
	ProviderProductID *string
}

type Computed struct {
	IsStatusValid bool
}

func (o *Main) Compute() {
	// value должен отсутствовать
	o.Value = ""

	o.Computed.IsStatusValid = o.ComputeIsStatusValid()
}

func (o *Main) ComputeIsStatusValid() bool {
	return o.Status == constant.KeyStatusActivated ||
		o.Status == constant.KeyStatusNew
}

//func (o *Main) MaskValue() string {
//	return o.Value == ""
//}
