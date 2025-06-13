package repo

import (
	"context"
	"fmt"
	mdmClient "github.com/mechta-market/e-product/internal/client/mdm"
	"github.com/mechta-market/e-product/internal/service/mdm/constant"
	"github.com/mechta-market/e-product/internal/service/mdm/model"
	repoModel "github.com/mechta-market/e-product/internal/service/mdm/repo/model"
)

type Repo struct {
	client mdmClient.Client
}

func New(client mdmClient.Client) *Repo {
	return &Repo{client: client}
}

// TODO: поменять местами return result & error

func (r *Repo) GetByProductID(ctx context.Context, productID string) (*model.Product, error) {
	searchReq := repoModel.SearchReq{
		Query: repoModel.SearchQuery{
			Bool: repoModel.BoolQuery{
				Must: []repoModel.TermQuery{
					{
						Term: map[string]any{
							"service_type": constant.ServiceType,
						},
					},
					{
						Term: map[string]any{
							"type": constant.ProductType,
						},
					},
				},
			},
		},
	}

	searchRepObj := &repoModel.SearchRep{}

	_, err := r.client.Send(ctx, &mdmClient.SendReq{
		DbName: "mdm",
		Method: "POST",
		Path:   "product/_search",
		ReqObj: searchReq,
		RepObj: searchRepObj,
	})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	for _, hit := range searchRepObj.Hits.Hits {
		provider := hit.Source.Provider

		if provider.ExternalNumber == productID {
			return repoModel.DecodeSearchRep(provider), nil
		}
	}

	return nil, fmt.Errorf("product with external_number '%s' not found in MDM", productID)
}
