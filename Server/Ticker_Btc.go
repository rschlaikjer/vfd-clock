// Grabs price info from BTC-e API

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

const btc_ticker_url = "https://btc-e.com/api/2/btc_usd/ticker"
const btc_trader_api = "http://xvjpf.org:4322/api"

type TickerBTC struct {
	price_points      []TickerData
	volatility_points []float64
	preseed_points    []*TickerData
	preseed_index     int
	preseed_length    int
}

type btcTickerResponse struct {
	Ticker TickerData
}

func (t *TickerBTC) GetPrice() (*TickerData, error) {
	response, err := http.Get(btc_ticker_url)
	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		td := new(btcTickerResponse)
		jserr := json.Unmarshal(contents, td)
		if jserr != nil {
			return nil, err
		}

		if td == nil || &td.Ticker == nil {
			return nil, errors.New("Ticker BTC: td.Ticker was nil! Contents: " + string(contents))
		}
		return &td.Ticker, nil
	}
}

func NewTickerBTC() *TickerBTC {
	t := new(TickerBTC)
	t.price_points = make([]TickerData, 0)
	t.volatility_points = make([]float64, 0)
	t.preseed_points = nil
	return t
}

type TraderApiJson struct {
	Profit float64
}

func getTraderProfit() (float64, error) {
	resp, err := http.Get(btc_trader_api)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var trader_data TraderApiJson
	err = json.Unmarshal(contents, &trader_data)
	if err != nil {
		return 0, err
	}

	return trader_data.Profit, nil
}
