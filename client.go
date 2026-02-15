package bybit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"
)

type bybitClient struct {
	apiKey     string
	apiSecret  string
	baseURL    string
	httpClient httpClient
}

type ClientOption func(*bybitClient)

// WithBaseURL is a client option to set the base URL of the Bybit HTTP client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *bybitClient) {
		c.baseURL = baseURL
	}
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func WithHttpClient(httpClient httpClient) ClientOption {
	return func(c *bybitClient) {
		c.httpClient = httpClient
	}
}

func NewBybitClient(apiKey string, APISecret string, options ...ClientOption) (BybitApi, error) {
	c := &bybitClient{
		apiKey:     apiKey,
		apiSecret:  APISecret,
		baseURL:    MAINNET,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	for _, opt := range options {
		opt(c)
	}

	return c, nil
}

func (c *bybitClient) post(ctx context.Context, url string, reqParams any) (data []byte, err error) {
	var jsonData []byte
	if reqParams != nil {
		jsonData, err = json.Marshal(reqParams)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	c.sign(req, string(jsonData[:]))

	return c.do(ctx, req)
}

func (c *bybitClient) get(ctx context.Context, url string, reqParams any) (data []byte, err error) {
	params := ""
	if reqParams != nil {
		urlValues, err := query.Values(reqParams)
		if err != nil {
			return nil, err
		}

		params = urlValues.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+url+"?"+params, nil)
	if err != nil {
		return nil, err
	}

	c.sign(req, params)

	return c.do(ctx, req)
}

func (c *bybitClient) sign(req *http.Request, params string) {
	timeStamp := time.Now().UnixMilli()

	hmac256 := hmac.New(sha256.New, []byte(c.apiSecret))
	hmac256.Write([]byte(strconv.FormatInt(timeStamp, 10) + c.apiKey + recvWindow + params))

	signature := hex.EncodeToString(hmac256.Sum(nil))

	req.Header.Set(headerApiKey, c.apiKey)
	req.Header.Set(headerSignature, signature)
	req.Header.Set(headerTimestamp, strconv.FormatInt(timeStamp, 10))
	req.Header.Set(headerSignType, "2")
	req.Header.Set(headerRecvWindow, recvWindow)
}

func (c *bybitClient) do(ctx context.Context, req *http.Request) (data []byte, err error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgentName+"/"+userAgentVersion)

	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr APIError
		err := json.Unmarshal(body, &apiErr)
		if err != nil {
			apiErr.Code = resp.StatusCode
			apiErr.Message = string(body)
		}

		apiErr.Method = req.Method
		apiErr.Path = req.URL.String()

		return nil, apiErr
	}

	return body, nil
}

var _ BybitApi = (*bybitClient)(nil)
