package model

import (
	"time"

	"github.com/mechta-market/e-product/internal/domain/key/model"
)

type Upsert struct {
	ID                    string
	UpdatedAt             *time.Time
	ProviderID            *string
	ProductID             *string
	Value                 *string
	Status                *string
	CustomerPhone         *string
	OrderID               *string
	ProviderOrderID       *string
	ProviderProductID     *string
	ProviderTransactionID *string
}

func (m *Upsert) UpdateColumnMap() map[string]any {
	res := m.CreateColumnMap()

	pkMap := m.PKColumnMap()
	for k := range pkMap {
		delete(res, k)
	}

	return res
}

// PKColumnMap возвращает первичный ключ для ON CONFLICT
func (m *Upsert) PKColumnMap() map[string]any {
	return map[string]any{
		"id": m.ID,
	}
}

func (m *Upsert) CreateColumnMap() map[string]any {
	result := make(map[string]any, 10)

	if m.UpdatedAt != nil {
		result["updated_at"] = *m.UpdatedAt
	}

	if m.ProviderID != nil {
		result["provider_id"] = *m.ProviderID
	}

	if m.ProductID != nil {
		result["product_id"] = *m.ProductID
	}

	if m.Value != nil {
		result["value"] = *m.Value
	}

	if m.Status != nil {
		result["status"] = *m.Status
	}

	if m.CustomerPhone != nil {
		result["customer_phone"] = *m.CustomerPhone
	}

	if m.OrderID != nil {
		result["order_id"] = *m.OrderID
	}

	if m.ProviderOrderID != nil {
		result["provider_order_id"] = *m.ProviderOrderID
	}

	if m.ProviderProductID != nil {
		result["provider_product_id"] = *m.ProviderProductID
	}

	if m.ProviderTransactionID != nil {
		result["provider_transaction_id"] = *m.ProviderTransactionID
	}

	return result
}

func (m *Upsert) ReturningColumnMap() map[string]any {
	return map[string]any{
		"id": &m.ID,
	}
}

func EncodeEdit(m *model.Edit) *Upsert {
	result := &Upsert{}

	if m.ID != nil && *m.ID != "" {
		result.ID = *m.ID
	}

	result.UpdatedAt = m.UpdatedAt
	result.ProviderID = m.ProviderID
	result.ProductID = m.ProductID
	result.Value = m.Value
	result.Status = m.Status
	result.CustomerPhone = m.CustomerPhone
	result.OrderID = m.OrderID
	result.ProviderOrderID = m.ProviderOrderID
	result.ProviderProductID = m.ProviderProductID
	result.ProviderTransactionID = m.ProviderTransactionID

	return result
}
