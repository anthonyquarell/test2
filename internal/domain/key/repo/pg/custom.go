package pg

import "github.com/mechta-market/e-product/internal/domain/key/model"

var (
	allowedSortFields = map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
)

func (r *Repo) getConditions(pars *model.ListReq) (map[string]any, map[string][]any) {
	conditions := make(map[string]any)
	conditionExps := make(map[string][]any)

	if pars.ProviderID != nil {
		conditions["provider_id"] = *pars.ProviderID
	}

	if pars.Status != nil {
		conditions["status"] = *pars.Status
	}

	if pars.OrderID != nil {
		conditions["order_id"] = *pars.OrderID
	}

	if pars.ProductID != nil {
		conditions["product_id"] = *pars.ProductID
	}

	return conditions, conditionExps
}
