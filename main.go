package main

import (
	"sync"

	"github.com/commoddity/bank-informer/client"
	"github.com/commoddity/bank-informer/cmc"
	"github.com/commoddity/bank-informer/csv"
	"github.com/commoddity/bank-informer/env"
	"github.com/commoddity/bank-informer/eth"
	"github.com/commoddity/bank-informer/log"
	"github.com/commoddity/bank-informer/persistence"
	"github.com/commoddity/bank-informer/pokt"
	"github.com/commoddity/bank-informer/setup"
)

const (
	// Required env vars
	grovePortalAppID     = "GROVE_PORTAL_APP_ID"
	groveSecretKey       = "GROVE_SECRET_KEY"
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
	// BadgerDB path
	dbPath = "./db"
)

type options struct {
	ethConfig  eth.Config
	poktConfig pokt.Config
	cmcConfig  cmc.Config

	cryptoFiatConversion string
	convertCurrencies    []string
	cryptoValues         []string

	persistenceDBPath string
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
	// Validate that Grove Secret Key is valid
	groveSecretKey := env.GetString(groveSecretKey, "")
	if groveSecretKey != "" {
		if err := pokt.ValidateSecretKey(groveSecretKey); err != nil {
			panic(err)
		}
	}

	// Validate that ETH wallet address is valid
	ethWalletAddress := env.MustGetString(ethWalletAddressEnv)
	if err := eth.ValidateETHWalletAddress(ethWalletAddress); err != nil {
		panic(err)
	}

	return options{
		ethConfig: eth.Config{
			PortalAppID:      grovePortalAppID,
			SecretKey:        groveSecretKey,
			ETHWalletAddress: ethWalletAddress,
		},
		poktConfig: pokt.Config{
			PortalAppID:       grovePortalAppID,
			SecretKey:         groveSecretKey,
			POKTWalletAddress: env.MustGetString(poktWalletAddressEnv),
		},
		cmcConfig: cmc.Config{
			CmcAPIKey: env.MustGetString(cmcAPIKeyEnv),
		},

		cryptoFiatConversion: cryptoFiatConversion,
		convertCurrencies:    convertCurrencies,
		cryptoValues:         env.GetStringSlice(cryptoValuesEnv, defaultCryptoValues),

		persistenceDBPath: dbPath,
	}
}

// This program retrieves and logs the balances of ETH and POKT wallets.
// It also fetches the exchange rates for a list of currencies and calculates
// the fiat values for each balance. The balances, fiat values, and exchange rates
// are then logged for further use.
func main() {
	// Setup .env file if it doesn't exist
	setup.Start()

	var mu sync.Mutex
	var wg sync.WaitGroup

	// Gather options from env vars
	opts := gatherOptions()

	persistence := persistence.NewPersistence(opts.persistenceDBPath)
	defer persistence.Close()

	// Add 1 to chanLength to account for the call to get exchange rates
	chanLength := len(opts.cryptoValues) + len(opts.convertCurrencies)
	progressChan := make(chan string, chanLength)

	// Initialize logger
	logger := log.New(log.Config{
		CryptoFiatConversion: opts.cryptoFiatConversion,
		ConvertCurrencies:    opts.convertCurrencies,
		CryptoValues:         opts.cryptoValues,
	}, persistence, progressChan, chanLength)

	go logger.RunProgressBar()

	// Create a map to store balances
	balances := make(map[string]float64)
	for _, crypto := range opts.cryptoValues {
		balances[crypto] = 0
	}

	// Create clients
	httpClient := client.New()
	ethClient := eth.NewClient(opts.ethConfig, httpClient, progressChan, &mu, &wg)
	poktClient := pokt.NewClient(opts.poktConfig, httpClient, progressChan, &mu, &wg)
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
	errorChan := make(chan error, len(opts.convertCurrencies))

	// For each currency in the list of currencies to convert
	for _, convertCurrency := range opts.convertCurrencies {
		wg.Add(1)
		go func(currency string) {
			defer wg.Done()

			// Retrieve and store the exchange rates for the current currency
			currencyExchangeRates, err := cmcClient.GetExchangeRates(balances, currency)
			if err != nil {
				errorChan <- err
				return
			}

			mu.Lock()
			// Add the retrieved exchange rates to the map of exchange rates
			exchangeRates[currency] = currencyExchangeRates
			mu.Unlock()

			progressChan <- currency
		}(convertCurrency)
	}

	wg.Wait()
	close(errorChan)
	close(progressChan)

	// Check if there were any errors
	if len(errorChan) > 0 {
		panic(<-errorChan)
	}

	// Calculate the fiat values for each balance
	fiatValues := cmcClient.GetFiatValues(balances, exchangeRates)

	// Log the balances, fiat values, and exchange rates
	logger.LogBalances(balances, fiatValues, exchangeRates)

	// Write the balances, fiat values, and exchange rates to a CSV file
	err = csv.WriteCryptoValuesToCSV(persistence, opts.cryptoValues)
	if err != nil {
		panic(err)
	}

	// Clear BadgerDB of old entries (older than 72 hours)
	err = persistence.ClearOldEntries()
	if err != nil {
		panic(err)
	}
}
