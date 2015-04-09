// A definition of the interface to Tickers (fetchers of price data)

package main

type TickerData struct {
	High    float64
	Low     float64
	Avg     float64
	Last    float64
	Buy     float64
	Sell    float64
	Updated int64
	Live    bool
}

type Ticker interface {
	GetPrice() (*TickerData, error)
}
