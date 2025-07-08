package repo

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	repoModel "github.com/mechta-market/e-product/internal/service/provider/megogo/repo/model"
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

func (r *Repo) CreateOrder(ctx context.Context, obj *providerModel.OrderRequest) (*providerModel.OrderResponse, error) {
	path := "/subscription/subscribe"
	sign := r.createSignature(path, obj.CustomerPhone, obj.ProviderProductID)
	params := map[string]string{
		"phone":     obj.CustomerPhone,
		"serviceId": obj.ProviderProductID,
		"sign":      sign,
	}

	apiResp := &repoModel.MegogoResponse{}
	repBody, err := r.sendRequest(
		ctx,
		http.MethodGet,
		path,
		8*time.Second,
		nil,
		apiResp,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	// decode
	result := &providerModel.OrderResponse{}

	if !apiResp.Successful {
		return nil, fmt.Errorf("send request: %s", repBody)
	}

	return result, nil
}

func (r *Repo) CancelOrder(ctx context.Context, obj *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	path := "/subscription/unsubscribe"
	sign := r.createSignature(path, *obj.CustomerPhone, *obj.ProviderProductID)
	params := map[string]string{
		"phone":     *obj.CustomerPhone,
		"serviceId": *obj.ProviderProductID,
		"sign":      sign,
	}

	apiResp := &repoModel.MegogoResponse{}
	repBody, err := r.sendRequest(
		ctx,
		http.MethodGet,
		path,
		8*time.Second,
		nil,
		apiResp,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if !apiResp.Successful {
		return nil, fmt.Errorf("send request: %s", repBody)
	}

	// decode
	result := &providerModel.CancelResponse{
		Success: apiResp.Successful,
	}

	return result, nil
}

func (r *Repo) createSignature(path, phone, serviceId string) string {
	hashString := fmt.Sprintf("%sGET/terminals/%s%sphone=%sserviceId=%s",
		r.password, r.username, path, phone, serviceId)

	hash := sha256.Sum256([]byte(hashString))
	hashHex := hex.EncodeToString(hash[:])

	hashBase64 := base64.StdEncoding.EncodeToString([]byte(hashHex))
	sign := strings.ReplaceAll(strings.ReplaceAll(hashBase64, " ", ""), "=", "")

	return sign + "_" + r.username
}

func (r *Repo) sendRequest(ctx context.Context, method, path string, timeout time.Duration, reqObj, repObj any, queryParams map[string]string) ([]byte, error) {
	// /terminals/{partnerId}{path}
	uri := r.uri + "terminals/" + r.username + path

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	if reqObj != nil {
		req.Header.Set("Content-Type", "application/json")
	}

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
		return repBody, fmt.Errorf("bad response status: %s, uri: %s, respBody: %q", resp.Status, uri, string(repBody))
	}

	if repObj != nil {
		if err = json.Unmarshal(repBody, repObj); err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w, body: '%s'", err, string(repBody))
		}
	}

	return repBody, nil
}
