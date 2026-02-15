package bybit

import (
	"encoding/json"
	"fmt"
)

const (
	userAgentName    = "bybit-go"
	userAgentVersion = "0.0.1"

	recvWindow = "5000"

	headerTimestamp  = "X-BAPI-TIMESTAMP"
	headerSignature  = "X-BAPI-SIGN"
	headerApiKey     = "X-BAPI-API-KEY"
	headerRecvWindow = "X-BAPI-RECV-WINDOW"
	headerSignType   = "X-BAPI-SIGN-TYPE"

	MAINNET        = "https://api.bybit.com"
	MAINNET_BACKT  = "https://api.bytick.com"
	NETHERLAND_ENV = "https://api.bybit.nl"
	HONGKONG_ENV   = "https://api.byhkbit.com"
	TURKEY_ENV     = "https://api.bybit-tr.com"
	KAZAKHSTAN_ENV = "https://api.bybit.kz"

	TESTNET  = "https://api-testnet.bybit.com"
	DEMO_ENV = "https://api-demo.bybit.com"
)

// APIError represents an error response from Bybit API.
type APIError struct {
	// Code is the status code of the error response.
	Code int `json:"retCode"`

	// Message contains a human-readable description of the error.
	Message string `json:"retMsg"`

	// Path contains the URL where the request was made.
	Path string `json:"-"`

	// Method contains the HTTP method used in the request.
	Method string `json:"-"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("[bybit error] %s: %s -> [%d] %s", e.Method, e.Path, e.Code, e.Message)
}

type ApiResponse[T any] struct {
	RetCode    int             `json:"retCode"`
	RetMsg     string          `json:"retMsg"`
	Result     T               `json:"result"`
	RetExtInfo json.RawMessage `json:"retExtInfo"`
	Time       int64           `json:"time"`
}
