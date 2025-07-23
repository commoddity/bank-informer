package log

import (
	"fmt"
	"math"
	"slices"
	"sync/atomic"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/cheggaaa/pb/v3"
	"github.com/commoddity/bank-informer/persistence"
)

type Logger struct {
	cryptoFiatConversion string
	cryptoValues         []string
	convertCurrencies    []string
	poktExchangeAmount   int64
	persistence          *persistence.Persistence
	progressChan         chan string
	chanLength           int
}

type Config struct {
	CryptoFiatConversion string
	CryptoValues         []string
	ConvertCurrencies    []string
	PoktExchangeAmount   int64
}

// Modified New function to include Persistence and POKT exchange amount
func New(config Config, persistence *persistence.Persistence, progressChan chan string, chanLength int) *Logger {
	return &Logger{
		cryptoFiatConversion: config.CryptoFiatConversion,
		cryptoValues:         config.CryptoValues,
		convertCurrencies:    config.ConvertCurrencies,
		poktExchangeAmount:   config.PoktExchangeAmount,
		persistence:          persistence,
		progressChan:         progressChan,
		chanLength:           chanLength,
	}
}

const (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorBlue  = "\033[34m"
	colorReset = "\033[0m"
)

var (
	fiatSymbols = map[string]string{
		"EUR": "€",  // Euro symbol
		"CAD": "$",  // Canadian Dollar symbol
		"USD": "$",  // US Dollar symbol
		"JPY": "¥",  // Japanese Yen symbol
		"GBP": "£",  // British Pound symbol
		"AUD": "$",  // Australian Dollar symbol
		"CHF": "₣",  // Swiss Franc symbol
		"CNY": "¥",  // Chinese Yuan symbol
		"SEK": "kr", // Swedish Krona symbol
		"NZD": "$",  // New Zealand Dollar symbol
		"ZAR": "R",  // South African Rand symbol
		"INR": "₹",  // Indian Rupee symbol
		"RUB": "₽",  // Russian Ruble symbol
		"BRL": "R$", // Brazilian Real symbol
		"KRW": "₩",  // South Korean Won symbol
		"IDR": "Rp", // Indonesian Rupiah symbol
		"MXN": "$",  // Mexican Peso symbol
		"ARS": "$",  // Argentine Peso symbol
		"MYR": "RM", // Malaysian Ringgit symbol
		"PHP": "₱",  // Philippine Peso symbol
		"PLN": "zł", // Polish Złoty symbol
		"THB": "฿",  // Thai Baht symbol
		"TRY": "₺",  // Turkish Lira symbol
		"VND": "₫",  // Vietnamese đồng symbol
	}

	fiatEmojis = map[string]string{
		"EUR": "🐗", // Wild Boar symbol for Europe
		"CAD": "🦆", // Duck symbol for Canada
		"USD": "🦅", // Eagle symbol for United States
		"JPY": "🦊", // Fox symbol for Japan
		"GBP": "🦡", // Badger symbol for Great Britain
		"AUD": "🦘", // Kangaroo symbol for Australia
		"CHF": "🐄", // Cow symbol for Switzerland
		"CNY": "🐼", // Panda symbol for China
		"SEK": "🦌", // Deer symbol for Sweden
		"NZD": "🥝", // Kiwi bird symbol for New Zealand
		"ZAR": "🦓", // Zebra symbol for South Africa
		"INR": "🐘", // Elephant symbol for India
		"RUB": "🐻", // Bear symbol for Russia
		"BRL": "🦜", // Parrot symbol for Brazil
		"KRW": "🐕", // Dog symbol for South Korea
		"IDR": "🦎", // Lizard symbol for Indonesia
		"MXN": "🦂", // Scorpion symbol for Mexico
		"ARS": "🦙", // Llama symbol for Argentina
		"MYR": "🦋", // Butterfly symbol for Malaysia
		"PHP": "🦈", // Shark symbol for the Philippines
		"PLN": "🦉", // Owl symbol for Poland
		"THB": "🐅", // Tiger symbol for Thailand
		"TRY": "🐐", // Goat symbol for Turkey
		"VND": "🐊", // Crocodile symbol for Vietnam
	}

	cryptoRoundValues = map[string]int{
		"POKT":  0,
		"WPOKT": 0,
		"USDC":  2,
		"USDT":  2,
		"ETH":   6,
		"WBTC":  6,
	}

	fiatRoundValues = map[string]int{
		"POKT": 4,
	}
)

func ValidateCurrencySymbol(currency, envVar string) error {
	if _, ok := fiatSymbols[currency]; !ok {
		return fmt.Errorf("invalid currency symbol %s found in env var %s", currency, envVar)
	}
	return nil
}

/* ------------ Log Funcs ------------ */

func (l *Logger) RunProgressBar() {
	fmt.Println("🔎 Bank Informer script is starting at", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Print("🔄 Fetching exchange rates for the following currencies: ", l.convertCurrencies, "\n")
	fmt.Print("💹 Crypto totals will be displayed in both crypto and the following fiat currency: ", l.cryptoFiatConversion, "\n")
	fmt.Print("💻 Crypto values will be displayed for the following cryptocurrencies: ", l.cryptoValues, "\n")

	bar := pb.StartNew(l.chanLength)

	bar.SetTemplateString(`{{string . "prefix"}} {{bar . "[" "=" ">" "_" "]"}} {{percent .}}`)
	bar.SetWidth(90)
	bar.SetMaxWidth(90)

	var currentRelay atomic.Int32

	// Increment progress bar each time a bool is received in the channel
	for val := range l.progressChan {
		currentRelay := currentRelay.Add(1)

		if currentRelay < int32(l.chanLength) {
			prefix := fmt.Sprintf("📡 Fetching data for %5s", val)
			bar.Set("prefix", prefix).Increment()
		} else {
			prefix := "🚀 Successfully fetched all data!"
			bar.Set("prefix", prefix).SetCurrent(int64(l.chanLength)).Finish()
		}
	}
}

func (l *Logger) LogBalances(balances map[string]float64, fiatValues map[string]float64, exchangeRates map[string]map[string]float64) {
	currentDate := time.Now().Format("2006-01-02") // format: YYYY-MM-DD
	previousDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	fiatTotal := 0.0

	<-time.After(100 * time.Millisecond)

	poktTotal := 0.0
	poktFiatTotal := 0.0

	fmt.Println("\n<--------- 🔐 Crypto Balances 🔐 --------->")
	for _, crypto := range l.cryptoValues {
		if balance, ok := balances[crypto]; ok {
			fiatValue := exchangeRates[l.cryptoFiatConversion][crypto]
			fiatBalance := balance * fiatValue

			fmt.Printf("%s - %s @ %s%s = %s%s %s", crypto, formatCryptoFloat(crypto, balance), fiatSymbols[l.cryptoFiatConversion], formatFiatFloat(crypto, fiatValue), fiatSymbols[l.cryptoFiatConversion], formatFiatFloat("", fiatBalance), l.cryptoFiatConversion)

			// Fetch average values from the previous day
			previousKey := fmt.Sprintf("%s-%s", crypto, previousDate)
			avgValues, err := l.persistence.GetAverageCryptoValues(previousKey)
			if err != nil {
				fmt.Printf(" %sNo data%s\n", colorBlue, colorReset)
			} else {
				difference := fiatBalance - avgValues.FiatBalance
				color := getColorForDifference(difference)

				if difference == 0 {
					fmt.Printf(" %s%s%s\n", color, "0.00", colorReset)
				} else {
					fmt.Printf(" %s%s%s\n", color, formatFiatFloat("", difference), colorReset)
				}

				fiatTotal += avgValues.FiatBalance
			}

			if crypto == "POKT" || crypto == "WPOKT" {
				poktTotal += balance
				poktFiatTotal += fiatBalance
			}

			key := fmt.Sprintf("%s-%s", crypto, currentDate)
			cryptoVal := persistence.CryptoValues{
				CryptoBalance: balance,
				FiatValue:     fiatValue,
				FiatBalance:   fiatBalance,
			}

			err = l.persistence.WriteCryptoValues(key, cryptoVal)
			if err != nil {
				fmt.Printf("Error writing crypto values to database: %s\n", err)
			}
		}
	}

	hasMultiplePokts := slices.Contains(l.cryptoValues, "WPOKT") && slices.Contains(l.cryptoValues, "POKT")
	if hasMultiplePokts && poktTotal > 0 {
		fiatValue := exchangeRates[l.cryptoFiatConversion]["POKT"]
		fmt.Printf("\n%s - %s @ %s%s = %s%s %s\n", "POKT Total", formatCryptoFloat("POKT", poktTotal), fiatSymbols[l.cryptoFiatConversion], formatFiatFloat("", fiatValue), fiatSymbols[l.cryptoFiatConversion], formatFiatFloat("", poktFiatTotal), l.cryptoFiatConversion)
	}

	// Display exchange balances section if POKT exchange amount is configured
	exchangeFiatValues := make(map[string]float64)
	if l.poktExchangeAmount > 0 {
		fmt.Println("\n<--------- 🌐 Exchange Balances 🌐 --------->")

		exchangeAmountFloat := float64(l.poktExchangeAmount)
		if fiatValue, ok := exchangeRates[l.cryptoFiatConversion]["POKT"]; ok {
			exchangeFiatBalance := exchangeAmountFloat * fiatValue

			fmt.Printf("POKT - %s @ %s%s = %s%s %s",
				formatCryptoFloat("POKT", exchangeAmountFloat),
				fiatSymbols[l.cryptoFiatConversion],
				formatFiatFloat("POKT", fiatValue),
				fiatSymbols[l.cryptoFiatConversion],
				formatFiatFloat("", exchangeFiatBalance),
				l.cryptoFiatConversion)

			// Fetch average values from the previous day for exchange amount
			previousKey := fmt.Sprintf("POKT-EXCHANGE-%s", previousDate)
			avgValues, err := l.persistence.GetAverageCryptoValues(previousKey)
			if err != nil {
				fmt.Printf(" %sNo data%s\n", colorBlue, colorReset)
			} else {
				difference := exchangeFiatBalance - avgValues.FiatBalance
				color := getColorForDifference(difference)

				if difference == 0 {
					fmt.Printf(" %s%s%s\n", color, "0.00", colorReset)
				} else {
					fmt.Printf(" %s%s%s\n", color, formatFiatFloat("", difference), colorReset)
				}

				fiatTotal += avgValues.FiatBalance
			}

			// Store exchange amount data
			key := fmt.Sprintf("POKT-EXCHANGE-%s", currentDate)
			cryptoVal := persistence.CryptoValues{
				CryptoBalance: exchangeAmountFloat,
				FiatValue:     fiatValue,
				FiatBalance:   exchangeFiatBalance,
			}

			err = l.persistence.WriteCryptoValues(key, cryptoVal)
			if err != nil {
				fmt.Printf("Error writing exchange crypto values to database: %s\n", err)
			}

			// Calculate exchange fiat values for all currencies
			for _, fiat := range l.convertCurrencies {
				if exchangeRate, ok := exchangeRates[fiat]["POKT"]; ok {
					exchangeFiatValues[fiat] = exchangeAmountFloat * exchangeRate
				}
			}
		}
	}

	fmt.Println("\n<--------- 💰 Fiat Total Balances 💰 --------->")
	defaultFiatBalance := fiatValues[l.cryptoFiatConversion] + exchangeFiatValues[l.cryptoFiatConversion]
	differenceInDefaultFiat := defaultFiatBalance - fiatTotal

	for _, fiat := range l.convertCurrencies {
		if balance, ok := fiatValues[fiat]; ok {
			// Add exchange amounts to total balance
			totalBalance := balance + exchangeFiatValues[fiat]
			fmt.Printf("%s %s - %s%s", fiatEmojis[fiat], fiat, fiatSymbols[fiat], formatFiatFloat("", totalBalance))

			if fiat == l.cryptoFiatConversion {
				color := getColorForDifference(differenceInDefaultFiat)
				if differenceInDefaultFiat == 0 {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], "0.00", colorReset)
				} else {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], formatFiatFloat("", differenceInDefaultFiat), colorReset)
				}
			} else {
				totalBalanceInDefaultFiat := fiatValues[l.cryptoFiatConversion] + exchangeFiatValues[l.cryptoFiatConversion]
				exchangeRate := totalBalance / totalBalanceInDefaultFiat
				difference := differenceInDefaultFiat * exchangeRate
				color := getColorForDifference(difference)
				if difference == 0 {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], "0.00", colorReset)
				} else {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], formatFiatFloat("", difference), colorReset)
				}
			}
		}
	}
}

func formatCryptoFloat(crypto string, num float64) string {
	roundValue, ok := cryptoRoundValues[crypto]
	if !ok {
		roundValue = 2 // default to 2 decimal places if crypto not found in map
	}
	p := message.NewPrinter(language.English)
	format := fmt.Sprintf("%%.%df", roundValue)
	return p.Sprintf(format, num)
}

func formatFiatFloat(crypto string, num float64) string {
	roundValue, ok := fiatRoundValues[crypto]
	if !ok {
		roundValue = 2 // default to 2 decimal places if crypto not found in map
	}
	p := message.NewPrinter(language.English)
	format := fmt.Sprintf("%%.%df", roundValue)
	return p.Sprintf(format, num)
}

func getColorForDifference(difference float64) string {
	const tolerance = 0.01

	if math.Abs(difference) < tolerance {
		difference = 0
	}

	var color string
	switch {
	case difference < 0:
		color = colorRed
	case difference > 0:
		color = colorGreen
	default:
		color = colorBlue
	}
	return color
}
