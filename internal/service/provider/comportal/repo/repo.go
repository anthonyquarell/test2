package repo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mechta-market/e-product/internal/errs"
	repoModel "github.com/mechta-market/e-product/internal/service/provider/comportal/repo/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type Repo struct {
	uri      string
	username string
	password string

	client *http.Client
}

func New(uri, username, password string) *Repo {
	return &Repo{
		uri:      strings.TrimRight(uri, "/") + "/",
		username: username,
		password: password,

		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 50,
			},
		},
	}
}

func (r *Repo) getProduct(ctx context.Context, sku string) (*repoModel.CatalogProduct, error) {
	catalogRep := &repoModel.CatalogRep{}

	_, err := r.sendRequest(
		ctx,
		http.MethodGet,
		"api/Catalog/Products",
		8*time.Second,
		nil,
		catalogRep,
		map[string]string{
			"imagesDisable": "true",
		})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	for _, product := range catalogRep.Data {
		if product.Sku == sku {
			return product, nil
		}
	}

	return nil, fmt.Errorf("product with SKU '%s' not found in catalog", sku)
}

func (r *Repo) CreateOrder(ctx context.Context, obj *providerModel.OrderRequest) (*providerModel.OrderResponse, error) {

	if obj == nil {
		return nil, fmt.Errorf("OrderReq cannot be nil")
	}

	if obj.ProviderProductID == "" {
		return nil, fmt.Errorf("ProductSKU cannot be nil")
	}

	catalogProduct, err := r.getProduct(ctx, obj.ProviderProductID)
	if err != nil {
		return nil, fmt.Errorf("repo.getProduct: %w", err)
	}

	if catalogProduct == nil {
		return nil, fmt.Errorf("catalogProduct not found")
	}

	apiReq := repoModel.EncodeOrderRequest(obj, catalogProduct)

	apiResp := &repoModel.OrderRep{}

	_, err = r.sendRequest(
		ctx,
		http.MethodPost,
		"api/Order",
		8*time.Second,
		apiReq,
		apiResp,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	decodedResponse := repoModel.DecodeOrderResponse(*apiResp)

	return decodedResponse, nil
}

func (r *Repo) GetCatalog(ctx context.Context, providerID string) ([]*providerModel.CatalogResponse, error) {
	catalogRep := &repoModel.CatalogRep{}

	_, err := r.sendRequest(
		ctx,
		http.MethodGet,
		"api/Catalog/Products",
		8*time.Second,
		nil,
		catalogRep,
		map[string]string{
			"imagesDisable": "true",
		})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	var result []*providerModel.CatalogResponse

	for _, product := range catalogRep.Data {
		result = append(result, repoModel.DecodeCatalogRep(product))
	}

	return result, nil
}

func (r *Repo) sendRequest(ctx context.Context, method, path string, timeout time.Duration, reqObj, repObj any, queryParams map[string]string) ([]byte, error) {
	path = strings.TrimLeft(path, "/")

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

	if r.username != "" {
		req.SetBasicAuth(r.username, r.password)
	}

	// query params
	if queryParams != nil {
		qPars := url.Values{}
		for k, v := range queryParams {
			qPars.Set(k, v)
		}
		req.URL.RawQuery = qPars.Encode()
	}

	slog.Info("Sending HTTP request", "method", req.Method, "url", req.URL.String())

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
