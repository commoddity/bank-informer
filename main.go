package main

import (
	"sync"

	"github.com/commoddity/bank-informer/client"
	"github.com/commoddity/bank-informer/cmc"
	"github.com/commoddity/bank-informer/config"
	"github.com/commoddity/bank-informer/csv"
	"github.com/commoddity/bank-informer/eth"
	"github.com/commoddity/bank-informer/log"
	"github.com/commoddity/bank-informer/persistence"
	"github.com/commoddity/bank-informer/pokt"
	"github.com/commoddity/bank-informer/setup"
)

// This program retrieves and logs the balances of ETH and POKT wallets.
// It also fetches the exchange rates for a list of currencies and calculates
// the fiat values for each balance. The balances, fiat values, and exchange rates
// are then logged for further use.
func main() {
	// Setup .env file if it doesn't exist
	setup.Start()

	// Gather options from env vars
	config, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize persistence module
	persistence := persistence.NewPersistence()
	defer persistence.Close()

	// Add 1 to chanLength to account for the call to get exchange rates
	chanLength := len(config.CryptoValues) + len(config.ConvertCurrencies)
	progressChan := make(chan string, chanLength)

	// Initialize logger
	logger := log.New(log.Config{
		CryptoFiatConversion: config.CryptoFiatConversion,
		ConvertCurrencies:    config.ConvertCurrencies,
		CryptoValues:         config.CryptoValues,
	}, persistence, progressChan, chanLength)

	// Start the progress bar in a goroutine
	go logger.RunProgressBar()

	// Create a map to store balances
	balances := make(map[string]float64)
	for _, crypto := range config.CryptoValues {
		balances[crypto] = 0
	}

	// Create mutex and wait group
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create ETH client
	httpClient := client.New()
	ethConfig := eth.Config{
		PathApiUrl:       config.PathApiUrl,
		PathApiKey:       config.PathApiKey,
		HttpClient:       httpClient,
		ETHWalletAddress: config.EthWalletAddress,
	}
	ethClient := eth.NewClient(ethConfig, progressChan, &mu, &wg)

	// Create POKT client
	poktConfig := pokt.Config{
		PathApiUrl:         config.PathApiUrl,
		PathApiKey:         config.PathApiKey,
		POKTWalletAddress:  config.PoktWalletAddress,
		HttpClient:         httpClient,
		PoktExchangeAmount: config.PoktExchangeAmount,
	}
	poktClient := pokt.NewClient(poktConfig, progressChan, &mu, &wg)

	// Create CMC client
	cmcConfig := cmc.Config{
		CMCAPIKey:         config.CMCAPIKey,
		ConvertCurrencies: config.ConvertCurrencies,
		HttpClient:        httpClient,
	}
	cmcClient := cmc.NewClient(cmcConfig, progressChan, &mu, &wg)

	// Retrieve and store ERC20 wallet balances through Grove Portal
	err = ethClient.GetETHWalletBalances(balances)
	if err != nil {
		panic(err)
	}

	// Retrieve and store POKT wallet balance through Grove Portal
	err = poktClient.GetWalletBalance(balances)
	if err != nil {
		panic(err)
	}

	// Retrieve and store the exchange rates for the current currency
	exchangeRates, err := cmcClient.GetAllExchangeRates(balances)
	if err != nil {
		panic(err)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the progress bar channel
	close(progressChan)

	// Calculate the fiat values for each balance
	fiatValues := cmcClient.GetFiatValues(balances, exchangeRates)

	// Log the balances, fiat values, and exchange rates
	logger.LogBalances(balances, fiatValues, exchangeRates)

	// Write the balances, fiat values, and exchange rates to a CSV file
	err = csv.WriteCryptoValuesToCSV(persistence, config.CryptoValues)
	if err != nil {
		panic(err)
	}

	// Clear BadgerDB of old entries (older than 72 hours)
	err = persistence.ClearOldEntries()
	if err != nil {
		panic(err)
	}
}
