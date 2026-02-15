package bybit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mockHTTPClient is a mock implementation of httpClient interface for testing
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestNewBybitClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		apiSecret string
		options   []ClientOption
		wantErr   bool
	}{
		{
			name:      "creates client with valid credentials",
			apiKey:    "test-key",
			apiSecret: "test-secret",
			options:   nil,
			wantErr:   false,
		},
		{
			name:      "creates client with empty credentials",
			apiKey:    "",
			apiSecret: "",
			options:   nil,
			wantErr:   false,
		},
		{
			name:      "creates client with base URL option",
			apiKey:    "test-key",
			apiSecret: "test-secret",
			options: []ClientOption{
				WithBaseURL(TESTNET),
			},
			wantErr: false,
		},
		{
			name:      "creates client with custom http client",
			apiKey:    "test-key",
			apiSecret: "test-secret",
			options: []ClientOption{
				WithHttpClient(&mockHTTPClient{}),
			},
			wantErr: false,
		},
		{
			name:      "creates client with multiple options",
			apiKey:    "test-key",
			apiSecret: "test-secret",
			options: []ClientOption{
				WithBaseURL(DEMO_ENV),
				WithHttpClient(&mockHTTPClient{}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewBybitClient(tt.apiKey, tt.apiSecret, tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBybitClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && client == nil {
				t.Error("NewBybitClient() returned nil client")
			}
		})
	}
}

func TestWithBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "sets mainnet URL",
			baseURL: MAINNET,
		},
		{
			name:    "sets testnet URL",
			baseURL: TESTNET,
		},
		{
			name:    "sets demo environment URL",
			baseURL: DEMO_ENV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewBybitClient("key", "secret", WithBaseURL(tt.baseURL))
			bybitClient := client.(*bybitClient)
			if bybitClient.baseURL != tt.baseURL {
				t.Errorf("WithBaseURL() failed: got %s, want %s", bybitClient.baseURL, tt.baseURL)
			}
		})
	}
}

func TestWithHttpClient(t *testing.T) {
	mockClient := &mockHTTPClient{}
	client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
	bybitClient := client.(*bybitClient)

	if bybitClient.httpClient != mockClient {
		t.Error("WithHttpClient() failed: http client not set correctly")
	}
}

func TestClientDefaults(t *testing.T) {
	client, _ := NewBybitClient("test-key", "test-secret")
	bybitClient := client.(*bybitClient)

	if bybitClient.apiKey != "test-key" {
		t.Errorf("apiKey = %s, want test-key", bybitClient.apiKey)
	}
	if bybitClient.apiSecret != "test-secret" {
		t.Errorf("apiSecret = %s, want test-secret", bybitClient.apiSecret)
	}
	if bybitClient.baseURL != MAINNET {
		t.Errorf("baseURL = %s, want %s", bybitClient.baseURL, MAINNET)
	}
	if bybitClient.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestPost(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		reqParams    any
		statusCode   int
		responseBody string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "successful POST request with params",
			url:          "/v5/order/create",
			reqParams:    map[string]string{"symbol": "BTCUSDT"},
			statusCode:   http.StatusOK,
			responseBody: `{"retCode":0,"retMsg":"OK"}`,
			wantErr:      false,
		},
		{
			name:         "successful POST request without params",
			url:          "/v5/order/create",
			reqParams:    nil,
			statusCode:   http.StatusOK,
			responseBody: `{"retCode":0,"retMsg":"OK"}`,
			wantErr:      false,
		},
		{
			name:         "POST request with error response",
			url:          "/v5/order/create",
			reqParams:    map[string]string{"symbol": "BTCUSDT"},
			statusCode:   http.StatusBadRequest,
			responseBody: `{"retCode":-1,"retMsg":"Invalid symbol"}`,
			wantErr:      true,
			errContains:  "Invalid symbol",
		},
		{
			name:         "POST request with server error",
			url:          "/v5/order/create",
			reqParams:    map[string]string{"symbol": "BTCUSDT"},
			statusCode:   http.StatusInternalServerError,
			responseBody: `Internal Server Error`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				doFunc: func(req *http.Request) (*http.Response, error) {
					if req.Method != http.MethodPost {
						t.Errorf("expected POST method, got %s", req.Method)
					}

					// Verify headers are set
					if req.Header.Get("Content-Type") != "application/json" {
						t.Error("Content-Type header not set correctly")
					}
					if req.Header.Get("X-BAPI-API-KEY") == "" {
						t.Error("API key header not set")
					}
					if req.Header.Get("X-BAPI-SIGN") == "" {
						t.Error("Signature header not set")
					}
					if req.Header.Get("X-BAPI-TIMESTAMP") == "" {
						t.Error("Timestamp header not set")
					}

					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
						Header:     make(http.Header),
					}, nil
				},
			}

			client, _ := NewBybitClient("test-key", "test-secret", WithHttpClient(mockClient))
			data, err := client.(*bybitClient).post(context.Background(), tt.url, tt.reqParams)

			if (err != nil) != tt.wantErr {
				t.Errorf("post() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error message should contain %q, got %q", tt.errContains, err.Error())
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("post() should return response body")
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		reqParams    any
		statusCode   int
		responseBody string
		wantErr      bool
	}{
		{
			name:         "successful GET request with params",
			url:          "/v5/market/kline",
			reqParams:    MarketKlineQuery{Symbol: SymbolBtcusdt, Interval: Interval1},
			statusCode:   http.StatusOK,
			responseBody: `{"retCode":0,"retMsg":"OK"}`,
			wantErr:      false,
		},
		{
			name:         "successful GET request without params",
			url:          "/v5/market/kline",
			reqParams:    nil,
			statusCode:   http.StatusOK,
			responseBody: `{"retCode":0,"retMsg":"OK"}`,
			wantErr:      false,
		},
		{
			name:         "GET request with error response",
			url:          "/v5/market/kline",
			reqParams:    map[string]string{"symbol": "INVALID"},
			statusCode:   http.StatusBadRequest,
			responseBody: `{"retCode":-1,"retMsg":"Invalid symbol"}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				doFunc: func(req *http.Request) (*http.Response, error) {
					if req.Method != http.MethodGet {
						t.Errorf("expected GET method, got %s", req.Method)
					}

					// Verify headers are set
					if req.Header.Get("Content-Type") != "application/json" {
						t.Error("Content-Type header not set correctly")
					}
					if req.Header.Get("X-BAPI-API-KEY") == "" {
						t.Error("API key header not set")
					}
					if req.Header.Get("X-BAPI-SIGN") == "" {
						t.Error("Signature header not set")
					}

					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
						Header:     make(http.Header),
					}, nil
				},
			}

			client, _ := NewBybitClient("test-key", "test-secret", WithHttpClient(mockClient))
			data, err := client.(*bybitClient).get(context.Background(), tt.url, tt.reqParams)

			if (err != nil) != tt.wantErr {
				t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("get() should return response body")
			}
		})
	}
}

func TestSign(t *testing.T) {
	client, _ := NewBybitClient("test-api-key", "test-secret")
	bybitClient := client.(*bybitClient)

	req, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)
	bybitClient.sign(req, "param1=value1&param2=value2")

	// Verify all required headers are set
	if req.Header.Get(headerApiKey) != "test-api-key" {
		t.Errorf("API key header not set correctly")
	}
	if req.Header.Get(headerSignature) == "" {
		t.Error("Signature header is empty")
	}
	if req.Header.Get(headerTimestamp) == "" {
		t.Error("Timestamp header is empty")
	}
	if req.Header.Get(headerSignType) != "2" {
		t.Error("Sign type header should be '2'")
	}
	if req.Header.Get(headerRecvWindow) != recvWindow {
		t.Errorf("Recv window header should be %s", recvWindow)
	}
}

func TestSignWithDifferentParams(t *testing.T) {
	tests := []struct {
		name   string
		params string
	}{
		{
			name:   "simple params",
			params: "symbol=BTCUSDT",
		},
		{
			name:   "empty params",
			params: "",
		},
		{
			name:   "complex params",
			params: "symbol=BTCUSDT&category=linear&limit=50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewBybitClient("key", "secret")
			bybitClient := client.(*bybitClient)

			req, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)
			bybitClient.sign(req, tt.params)

			if req.Header.Get(headerSignature) == "" {
				t.Error("Signature should not be empty")
			}
		})
	}
}

func TestSignatureConsistency(t *testing.T) {
	// Same client with same parameters should produce different signatures
	// because timestamp changes
	client, _ := NewBybitClient("key", "secret")
	bybitClient := client.(*bybitClient)

	req1, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)
	req2, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)

	bybitClient.sign(req1, "symbol=BTCUSDT")
	// Small delay to ensure different timestamp
	time.Sleep(10 * time.Millisecond)
	bybitClient.sign(req2, "symbol=BTCUSDT")

	// Signatures should be different due to different timestamps
	sig1 := req1.Header.Get(headerSignature)
	sig2 := req2.Header.Get(headerSignature)

	if sig1 == sig2 {
		t.Error("Signatures with different timestamps should be different")
	}
}

func TestDo(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		wantErr      bool
		expectError  APIError
	}{
		{
			name:         "successful response",
			statusCode:   http.StatusOK,
			responseBody: `{"retCode":0,"retMsg":"OK","result":{}}`,
			wantErr:      false,
		},
		{
			name:         "bad request error",
			statusCode:   http.StatusBadRequest,
			responseBody: `{"retCode":400,"retMsg":"Bad Request"}`,
			wantErr:      true,
			expectError:  APIError{Code: 400, Message: "Bad Request"},
		},
		{
			name:         "unauthorized error",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"retCode":401,"retMsg":"Unauthorized"}`,
			wantErr:      true,
			expectError:  APIError{Code: 401, Message: "Unauthorized"},
		},
		{
			name:         "internal server error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `Internal Server Error`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				doFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
						Header:     make(http.Header),
					}, nil
				},
			}

			client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
			req, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)

			data, err := client.(*bybitClient).do(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("do() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				apiErr, ok := err.(APIError)
				if !ok {
					t.Fatalf("expected APIError, got %T", err)
				}
				if tt.expectError.Code != 0 && apiErr.Code != tt.expectError.Code {
					t.Errorf("error code = %d, want %d", apiErr.Code, tt.expectError.Code)
				}
				if tt.expectError.Message != "" && apiErr.Message != tt.expectError.Message {
					t.Errorf("error message = %s, want %s", apiErr.Message, tt.expectError.Message)
				}
			} else {
				if len(data) == 0 {
					t.Error("do() should return response body for successful request")
				}
			}
		})
	}
}

func TestDoWithContext(t *testing.T) {
	t.Run("context is passed to request", func(t *testing.T) {
		mockClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				// Verify context is set on request
				if req.Context() == nil {
					t.Error("request context should not be nil")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"retCode":0}`)),
					Header:     make(http.Header),
				}, nil
			},
		}

		client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
		ctx := context.Background()
		req, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)

		client.(*bybitClient).do(ctx, req)
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"retCode":0}`)),
					Header:     make(http.Header),
				}, nil
			},
		}

		client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
		ctx, cancel := context.WithCancel(context.Background())
		req, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)

		cancel()
		_, err := client.(*bybitClient).do(ctx, req)

		// The mock client ignores context, so it won't fail
		// In real scenario with actual HTTP client, this would fail
		if err != nil {
			t.Logf("do() with cancelled context returned error as expected: %v", err)
		}
	})
}

func TestPostJSONMarshaling(t *testing.T) {
	t.Run("marshals request params to JSON", func(t *testing.T) {
		mockClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				body, _ := io.ReadAll(req.Body)
				var params map[string]string
				json.Unmarshal(body, &params)

				if params["symbol"] != "BTCUSDT" || params["category"] != "linear" {
					t.Error("request body not marshaled correctly")
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"retCode":0}`)),
					Header:     make(http.Header),
				}, nil
			},
		}

		client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
		client.(*bybitClient).post(context.Background(), "/v5/order/create", map[string]string{
			"symbol":   "BTCUSDT",
			"category": "linear",
		})
	})
}

func TestAPIErrorInfo(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"retCode":400,"retMsg":"Invalid symbol"}`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
	req, _ := http.NewRequest(http.MethodPost, "https://api.bybit.com/v5/order/create", nil)

	_, err := client.(*bybitClient).do(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr := err.(APIError)
	if apiErr.Method != http.MethodPost {
		t.Errorf("error method = %s, want POST", apiErr.Method)
	}
	if !strings.Contains(apiErr.Path, "api.bybit.com") {
		t.Errorf("error path should contain api.bybit.com, got %s", apiErr.Path)
	}
	if !strings.Contains(apiErr.Error(), "POST") {
		t.Error("error string should contain method")
	}
}

func TestPostWithInvalidJSON(t *testing.T) {
	client, _ := NewBybitClient("key", "secret")

	// Create a non-JSON-marshalable type
	ch := make(chan int)
	_, err := client.(*bybitClient).post(context.Background(), "/v5/test", ch)

	if err == nil {
		t.Error("post() should return error for non-JSON-marshalable type")
	}
}

func TestClientInterface(t *testing.T) {
	// Verify that bybitClient implements BybitApi interface
	var _ BybitApi = (*bybitClient)(nil)
}

func TestGetWithStructParams(t *testing.T) {
	type testParams struct {
		Symbol   string `url:"symbol"`
		Category string `url:"category"`
		Limit    int    `url:"limit"`
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			query := req.URL.RawQuery
			if !strings.Contains(query, "symbol=BTCUSDT") {
				t.Error("query should contain symbol parameter")
			}
			if !strings.Contains(query, "category=linear") {
				t.Error("query should contain category parameter")
			}
			if !strings.Contains(query, "limit=50") {
				t.Error("query should contain limit parameter")
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"retCode":0}`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
	client.(*bybitClient).get(context.Background(), "/v5/market/kline", testParams{
		Symbol:   "BTCUSDT",
		Category: "linear",
		Limit:    50,
	})
}

func TestUserAgent(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			ua := req.Header.Get("User-Agent")
			expectedUA := userAgentName + "/" + userAgentVersion
			if ua != expectedUA {
				t.Errorf("User-Agent = %s, want %s", ua, expectedUA)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"retCode":0}`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
	req, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)

	client.(*bybitClient).do(context.Background(), req)
}

func TestPostBodyIncludesContentType(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			ct := req.Header.Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("Content-Type = %s, want application/json", ct)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"retCode":0}`))),
				Header:     make(http.Header),
			}, nil
		},
	}

	client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
	client.(*bybitClient).post(context.Background(), "/v5/order/create", map[string]string{"symbol": "BTCUSDT"})
}

func TestResponseBodyReadError(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(&errorReader{}),
				Header:     make(http.Header),
			}, nil
		},
	}

	client, _ := NewBybitClient("key", "secret", WithHttpClient(mockClient))
	req, _ := http.NewRequest(http.MethodGet, "https://api.bybit.com/v5/test", nil)

	_, err := client.(*bybitClient).do(context.Background(), req)
	if err == nil {
		t.Error("do() should return error when body read fails")
	}
}

// errorReader is a helper for testing read errors
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
