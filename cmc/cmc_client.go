package cmc

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/commoddity/bank-informer/client"
)

const cmcURL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=%s&convert=%s"

type Config struct {
	CmcAPIKey string
}

type Client struct {
	Config     Config
	HTTPClient *http.Client
}

type cmcResult struct {
	Data map[string]struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
		Quote  map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

func NewClient(config Config, httpClient *http.Client) *Client {
	return &Client{
		Config:     config,
		HTTPClient: httpClient,
	}
}

func (c *Client) GetExchangeRates(balances map[string]float64, convertCurrency string) (map[string]float64, error) {
	url := fmt.Sprintf(cmcURL, getCurrencyKeys(balances), convertCurrency)

	header := http.Header{}
	header.Set("Accepts", "application/json")
	header.Add("X-CMC_PRO_API_KEY", c.Config.CmcAPIKey)

	cmcRes, err := client.Get[cmcResult](url, header, c.HTTPClient)
	if err != nil {
		return nil, err
	}

	prices := make(map[string]float64)
	for currency, data := range cmcRes.Data {
		prices[currency] = data.Quote[convertCurrency].Price
	}

	return prices, nil
}

func (c *Client) GetFiatValues(balances map[string]float64, fiatExchangeRates map[string]map[string]float64) map[string]float64 {
	fiatValues := make(map[string]float64)

	for fiat, exchangeRates := range fiatExchangeRates {
		for currency, balance := range balances {
			if exchangeRate, ok := exchangeRates[currency]; ok {
				fiatValue := balance * exchangeRate
				roundedFiatValue := math.Round(fiatValue*100) / 100
				fiatValues[fiat] += roundedFiatValue
			}
		}
	}

	return fiatValues
}

func getCurrencyKeys(balances map[string]float64) string {
	currencyKeys := make([]string, 0, len(balances))
	for key := range balances {
		currencyKeys = append(currencyKeys, key)
	}
	return strings.Join(currencyKeys, ",")
}
