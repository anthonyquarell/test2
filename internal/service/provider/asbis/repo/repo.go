package repo

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"golang.org/x/crypto/pkcs12"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	repoModel "github.com/mechta-market/e-product/internal/service/provider/asbis/repo/model"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
)

type Repo struct {
	uri              string
	username         string
	password         string
	p12CertPath      string
	p12Password      string
	serverCACertPath string

	client *http.Client
}

func New(uri, username, password, p12CertPath, p12Password, serverCACertPath string) (*Repo, error) {
	tlsConfig, err := tlsConnection(p12CertPath, p12Password, serverCACertPath)
	if err != nil {
		return nil, fmt.Errorf("tlsConnection: %w", err)
	}
	return &Repo{
		uri:              strings.TrimRight(uri, "/") + "/",
		username:         username,
		password:         password,
		p12CertPath:      p12CertPath,
		p12Password:      p12Password,
		serverCACertPath: serverCACertPath,

		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 50,
				TLSClientConfig:     tlsConfig,
			},
		},
	}, nil
}

func (r *Repo) CreateOrder(ctx context.Context, obj *providerModel.OrderRequest) (*providerModel.OrderResponse, error) {
	apiReq := repoModel.EncodeActivateRequest(obj)
	apiResp := &repoModel.OrderRep{}

	repBody, err := r.sendRequest(
		ctx,
		http.MethodPost,
		"api/esd/sb/req",
		8*time.Second,
		apiReq,
		apiResp,
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	result := repoModel.DecodeActivateResponse(*apiResp)

	if apiResp.ErrorCode == "79004" {
		return nil, fmt.Errorf("send request: %s", repBody)
	}

	return result, nil
}

func (r *Repo) CancelOrder(ctx context.Context, obj *providerModel.CancelRequest) (*providerModel.CancelResponse, error) {
	apiReq := repoModel.EncodeCancelRequest(obj, providerModel.GenerateUUID()) // генерируется новый CancelID аннулирования
	apiResp := &repoModel.OrderRep{}

	_, err := r.sendRequest(
		ctx,
		http.MethodPost,
		"api/esd/sb/req",
		8*time.Second,
		apiReq,
		apiResp,
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	result := repoModel.DecodeCancelResponse(*apiResp)

	return result, nil
}

func (r *Repo) sendRequest(ctx context.Context, method, path string, timeout time.Duration, reqObj, repObj any) ([]byte, error) {
	path = strings.TrimLeft(path, "/")

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var reqStream io.Reader
	if reqObj != nil {
		xmlData, err := xml.Marshal(reqObj)
		if err != nil {
			return nil, fmt.Errorf("fail to marshal reqObj: %w", err)
		}
		xmlWithHeader := []byte(xml.Header + string(xmlData))
		reqStream = bytes.NewBuffer(xmlWithHeader)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, r.uri+path, reqStream)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Accept", "application/xml")

	if r.username != "" {
		req.SetBasicAuth(r.username, r.password)
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
		return repBody, fmt.Errorf("bad response status: %s, uri: %s, respBody: %q", resp.Status, r.uri+path, string(repBody))
	}

	if repObj != nil {
		if err = xml.Unmarshal(repBody, repObj); err != nil {
			return nil, fmt.Errorf("xml.Unmarshal: %w, body: '%s'", err, string(repBody))
		}
	}

	return repBody, nil
}

func tlsConnection(p12CertPath, p12Password, serverCACertPath string) (*tls.Config, error) {
	p12Data, err := os.ReadFile(p12CertPath)
	if err != nil {
		return nil, err
	}

	blocks, err := pkcs12.ToPEM(p12Data, p12Password)
	if err != nil {
		return nil, err
	}

	var clientCert tls.Certificate

	for _, b := range blocks {
		switch b.Type {
		case "PRIVATE KEY":
			privateKey, err := x509.ParsePKCS8PrivateKey(b.Bytes)
			if err != nil {
				privateKey, err = x509.ParseECPrivateKey(b.Bytes)
			}
			if err != nil {
				privateKey, err = x509.ParsePKCS1PrivateKey(b.Bytes)
			}
			if err != nil {
				return nil, err
			}
			clientCert.PrivateKey = privateKey

		case "CERTIFICATE":
			clientCert.Certificate = append(clientCert.Certificate, b.Bytes)
		}
	}

	if clientCert.PrivateKey == nil || len(clientCert.Certificate) == 0 {
		return nil, err
	}

	systemRoots, _ := x509.SystemCertPool()
	if systemRoots == nil {
		return nil, err
	}

	serverCACert, err := os.ReadFile(serverCACertPath)
	if err != nil {
		return nil, err
	}
	if !systemRoots.AppendCertsFromPEM(serverCACert) {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            systemRoots,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	return tlsConfig, nil
}
