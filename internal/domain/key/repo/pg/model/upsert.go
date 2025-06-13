package model

import (
	"github.com/google/uuid"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"time"
)

type Upsert struct {
	ID            string
	UpdatedAt     *time.Time
	ProviderID    *string
	ProductID     *string
	Value         *string
	Status        *string
	CustomerPhone *string
	OrderID       *string
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
	result := make(map[string]any, 8)

	result["id"] = m.ID

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

	return result
}

func (m *Upsert) ReturningColumnMap() map[string]any {
	return map[string]any{
		"id":             &m.ID,
		"provider_id":    m.ProviderID,
		"product_id":     &m.ProductID,
		"updated_at":     &m.UpdatedAt,
		"customer_phone": &m.CustomerPhone,
		"status":         &m.Status,
		"order_id":       &m.OrderID,
		// НЕ возвращаем value согласно ТЗ!
	}
}

func EncodeItem(m *model.Key) *Upsert {
	result := &Upsert{}

	if m.ID == nil || *m.ID == "" {
		result.ID = uuid.New().String()
	} else {
		result.ID = *m.ID
	}

	result.UpdatedAt = m.UpdatedAt
	result.ProviderID = m.ProviderID
	result.ProductID = m.ProductID
	result.Value = m.Value
	result.Status = m.Status
	result.CustomerPhone = m.CustomerPhone
	result.OrderID = m.OrderID

	return result
}
