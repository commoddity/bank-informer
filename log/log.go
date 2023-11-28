package log

import (
	"fmt"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Logger struct {
	cryptoFiatConversion string
	cryptoValues         []string
	convertCurrencies    []string
}

type Config struct {
	CryptoFiatConversion string
	CryptoValues         []string
	ConvertCurrencies    []string
}

func New(config Config) *Logger {
	return &Logger{
		cryptoFiatConversion: config.CryptoFiatConversion,
		cryptoValues:         config.CryptoValues,
		convertCurrencies:    config.ConvertCurrencies,
	}
}

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
		"ILS": "₪",  // Israeli New Shekel symbol
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
		"GBP": "🦁", // Lion symbol for Great Britain
		"AUD": "🦘", // Kangaroo symbol for Australia
		"CHF": "🐄", // Cow symbol for Switzerland
		"CNY": "🐼", // Panda symbol for China
		"SEK": "🦌", // Deer symbol for Sweden
		"NZD": "🥝", // Kiwi bird symbol for New Zealand
		"ZAR": "🦁", // Lion symbol for South Africa
		"INR": "🐯", // Tiger symbol for India
		"RUB": "🐻", // Bear symbol for Russia
		"BRL": "🐆", // Jaguar symbol for Brazil
		"KRW": "🐯", // Tiger symbol for South Korea
		"IDR": "🐉", // Dragon symbol for Indonesia
		"MXN": "🦅", // Eagle symbol for Mexico
		"ARS": "🐆", // Jaguar symbol for Argentina
		"ILS": "💣", // Bomb symbol for Israel
		"MYR": "🐅", // Tiger symbol for Malaysia
		"PHP": "🦅", // Eagle symbol for the Philippines
		"PLN": "🦅", // Eagle symbol for Poland
		"THB": "🐘", // Elephant symbol for Thailand
		"TRY": "🐺", // Wolf symbol for Turkey
		"VND": "🐉", // Dragon symbol for Vietnam
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
	fmt.Println("🔎 Bank Informer script is starting...")

	fmt.Print("🔄 Fetching exchange rates for the following currencies:\n")
	for _, currency := range l.convertCurrencies {
		fmt.Printf("%s %s\n", fiatEmojis[currency], currency)
	}

	fmt.Print("💱 Crypto totals will be display in both crypto and the following fiat currency:\n")
	fmt.Printf("%s %s\n", fiatEmojis[l.cryptoFiatConversion], l.cryptoFiatConversion)

	fmt.Print("💱 Crypto values will be displayed for the following cryptocurrencies:\n")
	for _, crypto := range l.cryptoValues {
		fmt.Printf("%s\n", crypto)
	}

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
	fmt.Println("<--------- 🔐 Crypto Balances 🔐 --------->")
	for _, crypto := range l.cryptoValues {
		if balance, ok := balances[crypto]; ok {
			fiatValue := exchangeRates[l.cryptoFiatConversion][crypto]
			fmt.Printf("%s - %s @ %s%s = %s%s %s\n", crypto, formatFloat(balance), fiatSymbols[l.cryptoFiatConversion], formatFloat(fiatValue), fiatSymbols[l.cryptoFiatConversion], formatFloat(balance*fiatValue), l.cryptoFiatConversion)
		}
	}

	fmt.Println("\n<--------- 💰 Fiat Total Balances 💰 --------->")
	for _, fiat := range l.convertCurrencies {
		if balance, ok := fiatValues[fiat]; ok {
			fmt.Printf("%s %s - %s%s\n", fiatEmojis[fiat], fiat, fiatSymbols[fiat], formatFloat(balance))
		}
	}

	fmt.Print("\nFin.\n")
}

func formatFloat(num float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.2f", num)
}
