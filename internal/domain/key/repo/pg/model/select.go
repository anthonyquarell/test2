package model

import (
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"time"
)

type Select struct {
	ID            string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ProviderID    string
	ProductID     string
	Value         string
	Status        string
	CustomerPhone string
	OrderID       string
}

func (m *Select) ListColumnMap() map[string]any {

	return map[string]any{
		"id":             &m.ID,
		"created_at":     &m.CreatedAt,
		"updated_at":     &m.UpdatedAt,
		"provider_id":    &m.ProviderID,
		"product_id":     &m.ProductID,
		"value":          &m.Value,
		"status":         &m.Status,
		"customer_phone": &m.CustomerPhone,
		"order_id":       &m.OrderID,
	}
}

func (m *Select) PKColumnMap() map[string]any {
	return map[string]any{
		"id": m.ID,
	}
}

func (m *Select) DefaultSortColumns() []string {
	return []string{
		"created_at desc",
	}
}

func DecodeMain(m *Select, _ int) *model.Main {
	return &model.Main{
		ID:            m.ID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		ProviderID:    m.ProviderID,
		ProductID:     m.ProductID,
		Value:         m.Value,
		Status:        m.Status,
		CustomerPhone: m.CustomerPhone,
		OrderID:       m.OrderID,
	}
}
