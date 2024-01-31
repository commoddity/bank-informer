package cmc

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"

	"github.com/commoddity/bank-informer/client"
)

const cmcURL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=%s&convert=%s"

type Config struct {
	CmcAPIKey         string
	ConvertCurrencies []string
}

type Client struct {
	Config            Config
	HTTPClient        *http.Client
	convertCurrencies []string
	progressChan      chan string
	mutex             *sync.Mutex
	waitGroup         *sync.WaitGroup
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

func NewClient(config Config, httpClient *http.Client, progressChan chan string, mutex *sync.Mutex, waitGroup *sync.WaitGroup) *Client {
	return &Client{
		Config:            config,
		HTTPClient:        httpClient,
		convertCurrencies: config.ConvertCurrencies,
		progressChan:      progressChan,
		mutex:             mutex,
		waitGroup:         waitGroup,
	}
}

func (c *Client) GetAllExchangeRates(balances map[string]float64) (map[string]map[string]float64, error) {
	exchangeRates := make(map[string]map[string]float64)
	errorChan := make(chan error, len(c.convertCurrencies))

	// For each currency in the list of currencies to convert
	for _, convertCurrency := range c.convertCurrencies {
		c.waitGroup.Add(1)
		go func(currency string) {
			defer c.waitGroup.Done()

			// Retrieve and store the exchange rates for the current currency
			currencyExchangeRates, err := c.getExchangeRates(balances, currency)
			if err != nil {
				errorChan <- err
				return
			}

			c.mutex.Lock()
			// Add the retrieved exchange rates to the map of exchange rates
			exchangeRates[currency] = currencyExchangeRates
			c.mutex.Unlock()

			c.progressChan <- currency
		}(convertCurrency)
	}

	c.waitGroup.Wait()
	close(errorChan)

	// Check if there were any errors
	if len(errorChan) > 0 {
		return nil, <-errorChan
	}

	return exchangeRates, nil
}

func (c *Client) getExchangeRates(balances map[string]float64, convertCurrency string) (map[string]float64, error) {
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
