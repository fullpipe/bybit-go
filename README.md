# bybit-api-go

[![test](https://github.com/fullpipe/bybit-go/actions/workflows/test.yml/badge.svg)](https://github.com/fullpipe/bybit-go/actions/workflows/test.yml)
[![lint](https://github.com/fullpipe/bybit-go/actions/workflows/lint.yml/badge.svg)](https://github.com/fullpipe/bybit-go/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/fullpipe/bybit-go.svg)](https://pkg.go.dev/github.com/fullpipe/bybit-go)

Go client for the Bybit REST APIs. Typed as strongly as possible.

Domains:
  - market
  - order
  - position
  - pre-upgrade
  - account
  - asset
  - user
  - spread
  - rfq
  - spot-margin-uta
  - new-crypto-loan
  - crypto-loan
  - otc
  - broker
  - earn

## Usage

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fullpipe/bybit-go"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("BYBIT_API_KEY")
	apiSecret := os.Getenv("BYBIT_API_SECRET")

	client, err := bybit.NewBybitClient(apiKey, apiSecret, bybit.WithBaseURL(bybit.MAINNET))
	if err != nil {
		panic(err)
	}

	// Set current Leverage
	err = client.SpotMarginTradeSetLeverage(ctx, bybit.SpotMarginTradeSetLeverageParams{Leverage: "5"})
	if err != nil {
		panic(err)
	}

	// Get MarkPriceKline
	mk, err := client.GetMarketMarkPriceKline(ctx, bybit.MarketMarkPriceKlineQuery{
		Category: bybit.CategoryLinear,
		Symbol:   bybit.SymbolBtcusdt,
		Interval: bybit.Interval1,
		Limit:    5,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(mk)
}

```

## TODO

- websocket client
- parse types with `... (ms)` in comments as int64, even its is string originaly
- not all enums detected, like `AssetSettlementRecordListItem.Symbol`
- remove `github.com/google/go-querystring/query` dep?
- as I understand "price" field are strings because possible overflow of int64, parse them to `big.Int` or add custom string type wrapper

## Contributing

Contributions are welcome! 

- but if you found problem within `*_gen.go` files please open issue first.
- if some domain is missing also open issue first
