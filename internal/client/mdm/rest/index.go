package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/goccy/go-json"
	localConstant "github.com/mechta-market/e-product/internal/client/constant"
	"github.com/mechta-market/e-product/internal/client/mdm"
	"github.com/mechta-market/e-product/internal/errs"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type St struct {
	routes map[string]routeSt
}

type routeSt struct {
	dbName string
	uri    string
	token  string // bearer
	client *http.Client
}

func New() *St {
	return &St{
		routes: make(map[string]routeSt, 100),
	}
}

func (o *St) AddRoute(dbName, uri, token string) {
	dbName = strings.Trim(dbName, "/")
	o.routes[dbName] = routeSt{
		dbName: dbName,
		uri:    strings.TrimRight(uri, "/") + "/",
		token:  token,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (o *St) Send(ctx context.Context, obj *mdm.SendReq) (_ []byte, finalError error) {
	route, ok := o.routes[obj.DbName]
	if !ok {
		return nil, fmt.Errorf("route not found: %s", obj.DbName)
	}

	var reqStream io.Reader
	if obj.ReqObj != nil {
		jsonRaw, err := json.Marshal(obj.ReqObj)
		if err != nil {
			return nil, fmt.Errorf("fail to marshal reqObj: %w", err)
		}
		reqStream = bytes.NewReader(jsonRaw)
	}

	uri := route.uri + "/" + strings.TrimLeft(obj.Path, "/")

	req, err := http.NewRequestWithContext(ctx, obj.Method, uri, reqStream)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Source", localConstant.Source)

	// auth
	if route.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", route.token))
	}

	// query params
	if obj.Params != nil {
		qPars := url.Values{}
		for k, v := range obj.Params {
			qPars.Set(k, v)
		}
		req.URL.RawQuery = qPars.Encode()
	}

	// send request
	resp, err := route.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fail to send request: %w, uri: %s", err, uri)
	}
	defer resp.Body.Close()

	// read body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read response respBody: %w, uri: %s", err, uri)
	}

	// check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound {
			return respBody, fmt.Errorf("bad response status: %w %s, uri: %s, respBody: %q", errs.ObjectNotFound, resp.Status, uri, string(respBody))
		}
		return respBody, fmt.Errorf("bad response status: %s, uri: %s, respBody: %q", resp.Status, uri, string(respBody))
	}

	// parse body
	if obj.RepObj != nil {
		err = json.Unmarshal(respBody, obj.RepObj)
		if err != nil {
			return respBody, fmt.Errorf("fail to unmarshal response: %w, uri: %s, respBody: %q", err, uri, string(respBody))
		}
	}

	return respBody, nil
}
