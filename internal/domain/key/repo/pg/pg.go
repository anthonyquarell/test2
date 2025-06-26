package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mechta-market/mobone/v2"
	moboneTools "github.com/mechta-market/mobone/v2/tools"
	"github.com/opentracing/opentracing-go"
	"github.com/samber/lo"

	commonRepoPg "github.com/mechta-market/e-product/internal/domain/common/repo/pg"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	repoModel "github.com/mechta-market/e-product/internal/domain/key/repo/pg/model"
)

type Repo struct {
	*commonRepoPg.Base
	ModelStore *mobone.ModelStore
}

func New(con *pgxpool.Pool) *Repo {
	base := commonRepoPg.NewBase(con)
	return &Repo{
		Base: base,
		ModelStore: &mobone.ModelStore{
			Con:       base.Con,
			QB:        base.QB,
			TableName: "key",
		},
	}
}

func (r *Repo) List(ctx context.Context, pars *model.ListReq) (_ []*model.Main, _ int64, finalError error) {
	tracingSpan, ctx := opentracing.StartSpanFromContext(ctx, "key.repo.PG.List")
	defer tracingSpan.Finish()
	defer func() {
		if finalError != nil {
			tracingSpan.SetTag("error", true)
			tracingSpan.LogKV("error", finalError.Error())
		}
	}()

	conditions, conditionExps := r.getConditions(pars)
	sort := moboneTools.ConstructSortColumns(allowedSortFields, pars.Sort)

	items := make([]*repoModel.Select, 0)

	totalCount, err := r.ModelStore.List(ctx, mobone.ListParams{
		Conditions:           conditions,
		ConditionExpressions: conditionExps,
		Page:                 pars.Page,
		PageSize:             pars.PageSize,
		WithTotalCount:       pars.WithTotalCount,
		OnlyCount:            pars.OnlyCount,
		Sort:                 sort,
	}, func(add bool) mobone.ListModelI {
		item := &repoModel.Select{}

		if add {
			items = append(items, item)
		}
		return item
	})

	if err != nil {
		return nil, 0, fmt.Errorf("ModelStore.List: %w", err)
	}

	return lo.Map(items, repoModel.DecodeMain), totalCount, nil
}

func (r *Repo) Get(ctx context.Context, id string) (_ *model.Main, _ bool, finalError error) {
	tracingSpan, ctx := opentracing.StartSpanFromContext(ctx, "key.repo.PG.Get")
	defer tracingSpan.Finish()
	defer func() {
		if finalError != nil {
			tracingSpan.SetTag("error", true)
			tracingSpan.LogKV("error", finalError.Error())
		}
	}()

	m := &repoModel.Select{
		ID: id,
	}

	found, err := r.ModelStore.Get(ctx, m)
	if err != nil {
		return nil, false, fmt.Errorf("ModelStore.Get: %w", err)
	}
	if !found {
		return nil, false, nil
	}

	return repoModel.DecodeMain(m, 0), true, nil
}

func (r *Repo) GetByValue(ctx context.Context, value string) (_ *model.Main, finalError error) {
	tracingSpan, ctx := opentracing.StartSpanFromContext(ctx, "key.repo.PG.Get")
	defer tracingSpan.Finish()
	defer func() {
		if finalError != nil {
			tracingSpan.SetTag("error", true)
			tracingSpan.LogKV("error", finalError.Error())
		}
	}()

	m := &repoModel.SelectByValue{}
	m.Value = value

	found, err := r.ModelStore.Get(ctx, m)

	if err != nil {
		return nil, fmt.Errorf("ModelStore.Get: %w", err)
	}
	if !found {
		return nil, nil
	}

	return repoModel.DecodeMain(&m.Select, 0), nil
}

func (r *Repo) GetByOrderID(ctx context.Context, orderID string) (*model.Main, bool, error) {
	var m repoModel.Select
	query := `SELECT id, product_id, provider_id, provider_product_id, customer_phone, status, value
			  FROM key WHERE order_id = $1`

	err := r.Con.QueryRow(ctx, query, orderID).Scan(
		&m.ID,
		&m.ProductID,
		&m.ProviderID,
		&m.ProviderProductID,
		&m.CustomerPhone,
		&m.Status,
		&m.Value,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("key.repo.PG.GetByOrderID: %w", err)
	}

	return repoModel.DecodeMain(&m, 0), true, nil
}

func (r *Repo) Update(ctx context.Context, obj *model.Edit) (finalError error) {
	tracingSpan, ctx := opentracing.StartSpanFromContext(ctx, "key.repo.PG.Update")
	defer tracingSpan.Finish()
	defer func() {
		if finalError != nil {
			tracingSpan.SetTag("error", true)
			tracingSpan.LogKV("error", finalError.Error())
		}
	}()

	err := r.ModelStore.Update(ctx, repoModel.EncodeEdit(obj))
	if err != nil {
		return fmt.Errorf("ModelStore.Update: %w", err)
	}

	return nil
}

func (r *Repo) Set(ctx context.Context, obj *model.Edit) (finalError error) {
	tracingSpan, ctx := opentracing.StartSpanFromContext(ctx, "key.repo.PG.UpdateOrCreate")
	defer tracingSpan.Finish()
	defer func() {
		if finalError != nil {
			tracingSpan.SetTag("error", true)
			tracingSpan.LogKV("error", finalError.Error())
		}
	}()

	upsertObj := repoModel.EncodeEdit(obj)

	err := r.ModelStore.UpdateOrCreate(ctx, upsertObj)
	if err != nil {
		return fmt.Errorf("ModelStore.UpdateOrCreate: %w", err)
	}

	obj.ID = &upsertObj.ID

	return nil
}

// insert data from provider

func (r *Repo) CreateWithProvider(ctx context.Context, obj *model.Edit) (finalError error) {
	tracingSpan, ctx := opentracing.StartSpanFromContext(ctx, "ord.repo.PG.CreateWithProvider")
	defer tracingSpan.Finish()
	defer func() {
		if finalError != nil {
			tracingSpan.SetTag("error", true)
			tracingSpan.LogKV("error", finalError.Error())
		}
	}()

	var query string
	var args []any
	var insertedID string

	if obj.ID != nil && *obj.ID != "" {
		query = `
       INSERT INTO key (id, provider_id, product_id, value, status, customer_phone, order_id, provider_product_id, provider_order_id)
       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
       RETURNING id;
       `
		args = []any{
			*obj.ID,
			obj.ProviderID,
			obj.ProductID,
			obj.Value,
			obj.Status,
			obj.CustomerPhone,
			obj.OrderID,
			obj.ProviderProductID,
			obj.ProviderOrderID,
		}
	} else {
		query = `
       INSERT INTO key (provider_id, product_id, value, status, customer_phone, order_id, provider_product_id, provider_order_id)
       VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
       RETURNING id;
       `
		args = []any{
			obj.ProviderID,
			obj.ProductID,
			obj.Value,
			obj.Status,
			obj.CustomerPhone,
			obj.OrderID,
			obj.ProviderProductID,
			obj.ProviderOrderID,
		}
	}

	err := r.Con.QueryRow(ctx, query, args...).Scan(&insertedID)
	if err != nil {
		return fmt.Errorf("repo.CreateWithProvider: %w", err)
	}

	obj.ID = &insertedID

	return nil
}
