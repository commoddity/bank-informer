package log

import (
	"fmt"
	"math"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/commoddity/bank-informer/persistence"
)

type Logger struct {
	cryptoFiatConversion string
	cryptoValues         []string
	convertCurrencies    []string
	persistence          *persistence.Persistence
}

type Config struct {
	CryptoFiatConversion string
	CryptoValues         []string
	ConvertCurrencies    []string
}

// Modified New function to include Persistence
func New(config Config, persistence *persistence.Persistence) *Logger {
	return &Logger{
		cryptoFiatConversion: config.CryptoFiatConversion,
		cryptoValues:         config.CryptoValues,
		convertCurrencies:    config.ConvertCurrencies,
		persistence:          persistence,
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
		"ETH":   6,
		"WBTC":  6,
	}
)

func ValidateCurrencySymbol(currency, envVar string) error {
	if _, ok := fiatSymbols[currency]; !ok {
		return fmt.Errorf("invalid currency symbol %s found in env var %s", currency, envVar)
	}
	return nil
}

/* ------------ Log Funcs ------------ */

func (l *Logger) DisplayLoadingBar(done chan bool) {
	fmt.Println("🔎 Bank Informer script is starting at", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Print("🔄 Fetching exchange rates for the following currencies: ", l.convertCurrencies, "\n")
	fmt.Print("💹 Crypto totals will be displayed in both crypto and the following fiat currency: ", l.cryptoFiatConversion, "\n")
	fmt.Print("💻 Crypto values will be displayed for the following cryptocurrencies: ", l.cryptoValues, "\n")
	fmt.Print("🚀 Getting financial information ...\n")

	for i := 0; i <= 100; i++ {
		fmt.Printf("\r%d%% ", i) // This will print the percentage
		for j := 0; j < i; j++ {
			fmt.Print("▓")
		}
		time.Sleep(40 * time.Millisecond)
		if i == 100 {
			fmt.Print("\r\033[2K") // This will clear the entire line when loading reaches 100%
		}
	}

	fmt.Print("✅ Financial information fetched successfully!\n\n")

	done <- true
}

func (l *Logger) LogBalances(balances map[string]float64, fiatValues map[string]float64, exchangeRates map[string]map[string]float64) {
	currentDate := time.Now().Format("2006-01-02") // format: YYYY-MM-DD
	previousDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	fiatTotal := 0.0

	fmt.Println("<--------- 🔐 Crypto Balances 🔐 --------->")
	for _, crypto := range l.cryptoValues {
		if balance, ok := balances[crypto]; ok {
			fiatValue := exchangeRates[l.cryptoFiatConversion][crypto]
			fiatBalance := balance * fiatValue

			fmt.Printf("%s - %s @ %s%s = %s%s %s", crypto, formatFloat(crypto, balance), fiatSymbols[l.cryptoFiatConversion], formatFloat("", fiatValue), fiatSymbols[l.cryptoFiatConversion], formatFloat("", fiatBalance), l.cryptoFiatConversion)

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
					fmt.Printf(" %s%s%s\n", color, formatFloat("", difference), colorReset)
				}

				fiatTotal += avgValues.FiatBalance
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

	fmt.Println("\n<--------- 💰 Fiat Total Balances 💰 --------->")
	defaultFiatBalance := fiatValues[l.cryptoFiatConversion]
	differenceInDefaultFiat := defaultFiatBalance - fiatTotal

	for _, fiat := range l.convertCurrencies {
		if balance, ok := fiatValues[fiat]; ok {
			fmt.Printf("%s %s - %s%s", fiatEmojis[fiat], fiat, fiatSymbols[fiat], formatFloat("", balance))

			if fiat == l.cryptoFiatConversion {
				color := getColorForDifference(differenceInDefaultFiat)
				if differenceInDefaultFiat == 0 {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], "0.00", colorReset)
				} else {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], formatFloat("", differenceInDefaultFiat), colorReset)
				}
			} else {
				exchangeRate := balance / defaultFiatBalance
				difference := differenceInDefaultFiat * exchangeRate
				color := getColorForDifference(difference)
				if difference == 0 {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], "0.00", colorReset)
				} else {
					fmt.Printf(" %s%s%s%s\n", color, fiatSymbols[fiat], formatFloat("", difference), colorReset)
				}
			}
		}
	}
}

func formatFloat(crypto string, num float64) string {
	roundValue, ok := cryptoRoundValues[crypto]
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
