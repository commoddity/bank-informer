package main

import (
	"github.com/commoddity/bank-informer/client"
	"github.com/commoddity/bank-informer/cmc"
	"github.com/commoddity/bank-informer/env"
	"github.com/commoddity/bank-informer/eth"
	"github.com/commoddity/bank-informer/log"
	"github.com/commoddity/bank-informer/pokt"
	"github.com/commoddity/bank-informer/setup"
)

const (
	// Required env vars
	grovePortalAppID     = "GROVE_PORTAL_APP_ID"
	ethWalletAddressEnv  = "ETH_WALLET_ADDRESS"
	poktWalletAddressEnv = "POKT_WALLET_ADDRESS"
	cmcAPIKeyEnv         = "CMC_API_KEY"
	// Optional env vars
	cryptoFiatConversionEnv = "CRYPTO_FIAT_CONVERSION"
	convertCurrenciesEnv    = "CONVERT_CURRENCIES"
	cryptoValuesEnv         = "CRYPTO_VALUES"
	// Default currency values
	defaultConvertCurrencies    = "USD"
	defaultCryptoFiatConversion = "USD"
	defaultCryptoValues         = "USDC,ETH,POKT"
)

type options struct {
	ethConfig  eth.Config
	poktConfig pokt.Config
	cmcConfig  cmc.Config

	cryptoFiatConversion string
	convertCurrencies    []string
	cryptoValues         []string
}

func gatherOptions() options {
	// Validate that all converted currencies are valid
	convertCurrencies := env.GetStringSlice(convertCurrenciesEnv, defaultConvertCurrencies)
	for _, currency := range convertCurrencies {
		if err := log.ValidateCurrencySymbol(currency, convertCurrenciesEnv); err != nil {
			panic(err)
		}
	}
	// Validate that cryptoFiatConversion is valid
	cryptoFiatConversion := env.GetString(cryptoFiatConversionEnv, defaultCryptoFiatConversion)
	if err := log.ValidateCurrencySymbol(cryptoFiatConversion, cryptoFiatConversionEnv); err != nil {
		panic(err)
	}
	// Validate that Grove Portal App ID is valid
	grovePortalAppID := env.MustGetString(grovePortalAppID)
	if err := pokt.ValidatePortalAppID(grovePortalAppID); err != nil {
		panic(err)
	}
	// Validate that ETH wallet address is valid
	ethWalletAddress := env.MustGetString(ethWalletAddressEnv)
	if err := eth.ValidateETHWalletAddress(ethWalletAddress); err != nil {
		panic(err)
	}

	return options{
		ethConfig: eth.Config{
			PortalAppID:      grovePortalAppID,
			ETHWalletAddress: ethWalletAddress,
		},
		poktConfig: pokt.Config{
			PortalAppID:       grovePortalAppID,
			POKTWalletAddress: env.MustGetString(poktWalletAddressEnv),
		},
		cmcConfig: cmc.Config{
			CmcAPIKey: env.MustGetString(cmcAPIKeyEnv),
		},

		cryptoFiatConversion: cryptoFiatConversion,
		convertCurrencies:    convertCurrencies,
		cryptoValues:         env.GetStringSlice(cryptoValuesEnv, defaultCryptoValues),
	}
}

// This program retrieves and logs the balances of ETH and POKT wallets.
// It also fetches the exchange rates for a list of currencies and calculates
// the fiat values for each balance. The balances, fiat values, and exchange rates
// are then logged for further use.
func main() {
	// Setup .env file if it doesn't exist
	setup.Start()

	// Gather options from env vars
	opts := gatherOptions()

	// Initialize logger
	logger := log.New(log.Config{
		CryptoFiatConversion: opts.cryptoFiatConversion,
		ConvertCurrencies:    opts.convertCurrencies,
		CryptoValues:         opts.cryptoValues,
	})

	// Start a goroutine to display a 4 second loading bar while fetching financial information
	loadingBarDone := make(chan bool)
	go logger.DisplayLoadingBar(loadingBarDone)

	// Create a map to store balances
	balances := make(map[string]float64)
	for _, crypto := range opts.cryptoValues {
		balances[crypto] = 0
	}

	// Create clients
	httpClient := client.New()
	ethClient := eth.NewClient(opts.ethConfig, httpClient)
	poktClient := pokt.NewClient(opts.poktConfig, httpClient)
	cmcClient := cmc.NewClient(opts.cmcConfig, httpClient)

	// Retrieve and store ERC20 wallet balances through Grove Portal
	err := ethClient.GetETHWalletBalances(balances)
	if err != nil {
		panic(err)
	}

	// Retrieve and store POKT wallet balance through Grove Portal
	err = poktClient.GetWalletBalance(balances)
	if err != nil {
		panic(err)
	}

	// Create a map to store exchange rates
	exchangeRates := make(map[string]map[string]float64)

	// For each currency in the list of currencies to convert
	for _, convertCurrency := range opts.convertCurrencies {
		// Retrieve and store the exchange rates for the current currency
		currencyExchangeRates, err := cmcClient.GetExchangeRates(balances, convertCurrency)
		if err != nil {
			panic(err)
		}

		// Add the retrieved exchange rates to the map of exchange rates
		exchangeRates[convertCurrency] = currencyExchangeRates
	}

	// Calculate the fiat values for each balance
	fiatValues := cmcClient.GetFiatValues(balances, exchangeRates)

	// Wait for DisplayLoadingBar to finish
	<-loadingBarDone

	// Log the balances, fiat values, and exchange rates
	logger.LogBalances(balances, fiatValues, exchangeRates)
}
