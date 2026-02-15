package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_bybitClient_sign(t *testing.T) {
	tests := []struct {
		name      string
		timeStamp int64
		params    string
		apiKey    string
		apiSecret string
		signature string
	}{
		{
			"doc example",
			1658384314791,
			"category=option&symbol=BTC-29JUL22-25000-C",
			"XXXXXXXXXX",
			"XXXXXXXXXX",
			"XXXXXXXXXX",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hmac256 := hmac.New(sha256.New, []byte(tt.apiSecret))
			hmac256.Write([]byte(strconv.FormatInt(tt.timeStamp, 10) + tt.apiKey + recvWindow + tt.params))

			assert.Equal(t, tt.signature, hex.EncodeToString(hmac256.Sum(nil)))

		})
	}
}
