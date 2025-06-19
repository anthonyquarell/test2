package repo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/mechta-market/e-product/internal/errs"
	"github.com/mechta-market/e-product/internal/service/constant"
	"github.com/mechta-market/e-product/internal/service/provider/comportal/model"
	repoModel "github.com/mechta-market/e-product/internal/service/provider/comportal/repo/model"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
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
		15*time.Second,
		nil,
		catalogRep,
		map[string]string{
			"imagesDisable": "true",
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %w", err)
	}

	for _, product := range catalogRep.Data {
		if product.Sku == sku {
			return product, nil
		}
	}

	return nil, fmt.Errorf("product with SKU '%s' not found in catalog", sku)
}

func (r *Repo) CreateOrder(ctx context.Context, obj *model.OrderReq) (*model.OrderRep, error) {

	if obj == nil {
		return nil, fmt.Errorf("OrderReq cannot be nil")
	}

	if obj.ProductSKU == nil {
		return nil, fmt.Errorf("ProductSKU cannot be nil")
	}

	catalogProduct, err := r.getProduct(ctx, *obj.ProductSKU)
	if err != nil {
		return nil, fmt.Errorf("repo.getProduct: %w", err)
	}

	if catalogProduct == nil {
		return nil, fmt.Errorf("catalogProduct not found")
	}

	apiReq := repoModel.EncodeOrderRequest(obj, *catalogProduct, "43c55461-49c6-4f9e-8cfd-b3bda7e0427c") // TODO: здесь генерировать ID uuid.New().String()

	apiResp := &repoModel.OrderRep{}

	repBody, err := r.sendRequest(
		ctx,
		http.MethodPost,
		"api/Order",
		15*time.Second,
		apiReq,
		apiResp,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	slog.Info("Raw response from comportal", "response_body", string(repBody))

	slog.Info("Parsed response from comportal", "parsed_response", apiResp)

	decodedResponse := repoModel.DecodeOrderResponse(*apiResp)

	slog.Info("Final decoded response", "decoded", decodedResponse)

	return decodedResponse, nil
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
	req.Header.Set("Source", constant.Source)

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
