package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Start() {
	checkEnvFile()
}

func checkEnvFile() {
	_, err := os.Stat(".env")
	if os.IsNotExist(err) {
		promptUser()
	}
}

func promptUser() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("👋 Welcome to the Bank Informer app! It looks like you're running the app for the first time.\n❓We need to gather a few variables to get started. Would you like to proceed?\n(yes/no): ")

	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if strings.ToLower(text) == "yes" {
		createEnvFile()
	}
}

func createEnvFile() {
	file, err := os.Create(".env")
	if err != nil {
		fmt.Println("Error creating .env file:", err)
		return
	}
	defer file.Close()

	prompts := []struct {
		key, description string
	}{
		{"GROVE_PORTAL_APP_ID", "🌿 Enter your Grove Portal App ID. This is the portal application ID for the Grove Portal, used to fetch ERC20 and POKT wallet balances.\nYou can get a free Grove account at https://portal.grove.city/\n"},
		{"ETH_WALLET_ADDRESS", "💼 Enter your Ethereum Wallet Address. This is the address to fetch ERC20 token balances from.\n"},
		{"POKT_WALLET_ADDRESS", "🎒 Enter your POKT Wallet Address. This is the address to fetch the POKT balance from.\n"},
		{"CMC_API_KEY", "🔑 Enter the CoinMarketCap API Key. This is used to fetch exchange rates.\nYou can get a free API key from https://pro.coinmarketcap.com/\n"},
	}

	reader := bufio.NewReader(os.Stdin)
	for _, prompt := range prompts {
		clearConsole()

		fmt.Print(prompt.description)
		value, _ := reader.ReadString('\n')
		value = strings.ReplaceAll(value, "\n", "")
		_, err := file.WriteString(fmt.Sprintf("%s=%s\n", prompt.key, value))
		if err != nil {
			fmt.Println("Error writing to .env file:", err)
			return
		}
		os.Setenv(prompt.key, value)
	}

	clearConsole()

	fmt.Print("💱 Do you want to set optional currency variables?\nThese variables allow you to customize:\n- the fiat currency to convert the crypto balances to\n- the list of fiat currencies to fetch exchange rates for\n- the list of cryptocurrencies to display values for\n(yes/no): ")
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	if strings.ToLower(text) == "yes" {
		optionalPrompts := []struct {
			key, description string
		}{
			{"CRYPTO_FIAT_CONVERSION", "💱 Enter the fiat currency to convert the crypto balances to (default: USD):\n"},
			{"CONVERT_CURRENCIES", "🔄 Enter a comma-separated list of fiat currencies to fetch exchange rates for (default: USD):\n"},
			{"CRYPTO_VALUES", "💰 Enter a comma-separated list of cryptocurrencies to display values for (default: USDC,ETH,POKT):\n"},
		}

		for _, prompt := range optionalPrompts {
			clearConsole()
			fmt.Print(prompt.description)
			value, _ := reader.ReadString('\n')
			value = strings.ReplaceAll(value, "\n", "")
			if value != "" {
				_, err := file.WriteString(fmt.Sprintf("%s=%s\n", prompt.key, value))
				if err != nil {
					fmt.Println("Error writing to .env file:", err)
					return
				}
				os.Setenv(prompt.key, value)
			}
		}
	}

	clearConsole()
	fmt.Println(".env file has been created and populated.")
}

func clearConsole() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}