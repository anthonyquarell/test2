package repo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/mechta-market/e-product/internal/errs"
	serviceConstant "github.com/mechta-market/e-product/internal/service/constant"
	"github.com/mechta-market/e-product/internal/service/mdm/constant"
	"github.com/mechta-market/e-product/internal/service/mdm/model"
	repoModel "github.com/mechta-market/e-product/internal/service/mdm/repo/model"
	"io"
	"net/http"
	"strings"
	"time"
)

type Repo struct {
	uri   string
	token string // bearer

	client *http.Client
}

func New(uri, token string) *Repo {
	return &Repo{
		uri:   strings.TrimRight(uri, "/") + "/",
		token: token,

		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 50,
			},
		},
	}
}

// TODO: поменять местами return result & error

func (r *Repo) GetByProductID(ctx context.Context, productID string) (*model.Product, error) {
	searchRepObj := &repoModel.HitRecord{}

	_, err := r.sendRequest(
		ctx,
		http.MethodGet,
		"product/_doc/"+productID,
		15*time.Second,
		nil,
		searchRepObj,
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	id := searchRepObj.ID
	provider := searchRepObj.Source.Provider

	if id == productID {
		return repoModel.DecodeSearchRep(provider, *searchRepObj), nil
	}

	return nil, errs.ObjectNotFound
}

func (r *Repo) GetCatalog(ctx context.Context, providerID string) ([]*model.CatalogProduct, error) {
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

		Size: constant.Size,
	}

	searchRepObj := &repoModel.SearchRep{}

	_, err := r.sendRequest(
		ctx,
		http.MethodPost,
		"product/_search",
		15*time.Second,
		searchReq,
		searchRepObj,
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	var result []*model.CatalogProduct

	for _, hit := range searchRepObj.Hits.Hits {
		provider := hit.Source.Provider

		if provider.ProviderID != providerID {
			continue
		}

		result = append(result, repoModel.DecodeCatalogRep(provider, hit))
	}

	return result, nil
}

func (r *Repo) sendRequest(ctx context.Context, method, path string, timeout time.Duration, reqObj, repObj any) ([]byte, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var reqStream io.Reader
	if reqObj != nil {
		jsonData, err := json.Marshal(reqObj)
		if err != nil {
			return nil, fmt.Errorf("fail to marshal reqObj: %w", err)
		}
		reqStream = bytes.NewBuffer(jsonData)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, r.uri+path, reqStream)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source", serviceConstant.Source)

	if r.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.token))
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	repBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound {
			return repBody, fmt.Errorf("bad response status: %w %s, uri: %s, respBody: %q", errs.ObjectNotFound, resp.Status, r.uri+path, string(repBody))
		}
		return repBody, fmt.Errorf("bad response status: %s, uri: %s, respBody: %q", resp.Status, r.uri+path, string(repBody))
	}

	if repObj != nil {
		if err = json.Unmarshal(repBody, repObj); err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w, body: '%s'", err, string(repBody))
		}
	}

	return repBody, nil
}
