package main

import (
	"github.com/commoddity/bank-informer/client"
	"github.com/commoddity/bank-informer/cmc"
	"github.com/commoddity/bank-informer/covalent"
	"github.com/commoddity/bank-informer/env"
	"github.com/commoddity/bank-informer/log"
	"github.com/commoddity/bank-informer/pokt"
)

const (
	// Required env vars
	covalentAPIKeyEnv    = "COVALENT_API_KEY"
	ethWalletAddressEnv  = "ETH_WALLET_ADDRESS"
	poktPortalAppID      = "POKT_PORTAL_APP_ID"
	poktWalletAddressEnv = "POKT_WALLET_ADDRESS"
	cmcAPIKeyEnv         = "CMC_API_KEY"
	// Optional env vars
	cryptoFiatConversionEnv = "CRYPTO_FIAT_CONVERSION"
	convertCurrenciesEnv    = "CONVERT_CURRENCIES"
	cryptoValuesEnv         = "CRYPTO_VALUES"
)

type options struct {
	covalentConfig covalent.Config
	poktConfig     pokt.Config
	cmcConfig      cmc.Config

	cryptoFiatConversion string
	convertCurrencies    []string
	cryptoValues         []string
}

func gatherOptions() options {
	// Validate that all converted currencies are valid
	convertCurrencies := env.GetStringSlice(convertCurrenciesEnv, "CAD")
	for _, currency := range convertCurrencies {
		if err := log.ValidateCurrencySymbol(currency, convertCurrenciesEnv); err != nil {
			panic(err)
		}
	}

	return options{
		covalentConfig: covalent.Config{
			APIKey:           env.MustGetString(covalentAPIKeyEnv),
			EthWalletAddress: env.MustGetString(ethWalletAddressEnv),
		},
		poktConfig: pokt.Config{
			PortalAppID:       env.MustGetString(poktPortalAppID),
			POKTWalletAddress: env.MustGetString(poktWalletAddressEnv),
		},
		cmcConfig: cmc.Config{
			CmcAPIKey: env.MustGetString(cmcAPIKeyEnv),
		},
		convertCurrencies:    convertCurrencies,
		cryptoFiatConversion: env.GetString(cryptoFiatConversionEnv, "CAD"),
		cryptoValues:         env.GetStringSlice(cryptoValuesEnv, "USDC,ETH,POKT"),
	}
}

// This program retrieves and logs the balances of ETH and POKT wallets.
// It also fetches the exchange rates for a list of currencies and calculates
// the fiat values for each balance. The balances, fiat values, and exchange rates
// are then logged for further use.
func main() {
	// Initialize HTTP client
	httpClient := client.New()

	// Gather options from env vars
	opts := gatherOptions()

	// Initialize logger
	logger := log.New(log.Config{
		CryptoFiatConversion: opts.cryptoFiatConversion,
		CryptoValues:         opts.cryptoValues,
		ConvertCurrencies:    opts.convertCurrencies,
	})

	// Create a channel to signal when DisplayLoadingBar is done
	done := make(chan bool)

	// Start a goroutine to display a 4 second loading bar while fetching financial information
	go logger.DisplayLoadingBar(done)

	// Create a Covalent client
	covalentClient := covalent.NewClient(opts.covalentConfig, httpClient)

	// Create a POKT client
	poktClient := pokt.NewClient(opts.poktConfig, httpClient)

	// Create a CMC client
	cmcClient := cmc.NewClient(opts.cmcConfig, httpClient)

	// Create a map to store balances
	balances := make(map[string]float64)

	// Retrieve and store ETH wallet balance from Covalent
	err := covalentClient.GetEthWalletBalance(balances)
	if err != nil {
		panic(err)
	}

	// Retrieve and store POKT wallet balance from POKT Provider
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
	<-done

	// Log the balances, fiat values, and exchange rates
	logger.LogBalances(balances, fiatValues, exchangeRates)
}
