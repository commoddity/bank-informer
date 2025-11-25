package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/commoddity/bank-informer/config"
)

// Define the field names as constants in snake_case, as sourced from .bankinformer.config.yaml and config.go.
const (
	KeyRpcApiKey            = "rpc_api_key"
	KeyEthWalletAddress     = "eth_wallet_address"
	KeyPoktWalletAddress    = "pokt_wallet_address"
	KeyCmcApiKey            = "cmc_api_key"
	KeyCryptoFiatConversion = "crypto_fiat_conversion"
	KeyConvertCurrencies    = "convert_currencies"
	KeyCryptoValues         = "crypto_values"
)

func Start() {
	checkConfigFile()
}

func checkConfigFile() {
	_, err := os.Stat(config.ConfigPath)
	if os.IsNotExist(err) {
		promptUser()
	}
}

func promptUser() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("👋 Welcome to the Bank Informer app! It looks like you're running the app for the first time.\n❓ We need to gather a few variables to create your YAML configuration file.\nWould you like to proceed? (yes/no): ")

	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if strings.ToLower(text) == "yes" {
		createConfigFile()
	}
}

func createConfigFile() {
	// Ensure the configuration directory exists.
	configDir := filepath.Dir(config.ConfigPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		fmt.Println("🚫 Error creating configuration directory:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	var cfg config.Config

	// Required configuration prompts using constant field names.
	requiredPrompts := []struct {
		field  string
		prompt string
	}{
		{KeyEthWalletAddress, "💼 Enter your Ethereum Wallet Address: "},
		{KeyPoktWalletAddress, "🎒 Enter your POKT Wallet Address: "},
		{KeyCmcApiKey, "🔑 Enter the CoinMarketCap API KEY: "},
	}

	for _, p := range requiredPrompts {
		clearConsole()
		fmt.Print(p.prompt)
		value, _ := reader.ReadString('\n')
		value = strings.TrimSpace(value)
		switch p.field {
		case KeyEthWalletAddress:
			cfg.EthWalletAddress = value
		case KeyPoktWalletAddress:
			cfg.PoktWalletAddress = value
		case KeyCmcApiKey:
			cfg.CMCAPIKey = value
		}
	}

	// Optional RPC API key prompt
	clearConsole()
	fmt.Print("🔑 Enter your RPC API KEY (optional, for Pocket Network API authentication): ")
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value != "" {
		cfg.RpcApiKey = value
	}

	// Optional configuration prompts using constant field names.
	clearConsole()
	fmt.Print("💱 Do you want to set optional currency variables?\nThese variables allow you to customize:\n- the fiat currency to convert crypto balances to\n- the list of fiat currencies to fetch exchange rates for\n- the list of cryptocurrencies to display values for\n(yes/no): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if strings.ToLower(text) == "yes" {
		optionalPrompts := []struct {
			field  string
			prompt string
		}{
			{KeyCryptoFiatConversion, "💱 Enter the fiat currency to convert crypto balances to (default: USD): "},
			{KeyConvertCurrencies, "🔄 Enter a comma-separated list of fiat currencies (default: USD): "},
			{KeyCryptoValues, "💰 Enter a comma-separated list of cryptocurrencies (default: USDC,ETH,POKT): "},
		}

		for _, p := range optionalPrompts {
			clearConsole()
			fmt.Print(p.prompt)
			value, _ := reader.ReadString('\n')
			value = strings.TrimSpace(value)
			if value != "" {
				switch p.field {
				case KeyCryptoFiatConversion:
					cfg.CryptoFiatConversion = value
				case KeyConvertCurrencies:
					parts := strings.Split(value, ",")
					for i, s := range parts {
						parts[i] = strings.TrimSpace(s)
					}
					cfg.ConvertCurrencies = parts
				case KeyCryptoValues:
					parts := strings.Split(value, ",")
					for i, s := range parts {
						parts[i] = strings.TrimSpace(s)
					}
					cfg.CryptoValues = parts
				}
			}
		}
	}

	clearConsole()
	fmt.Println("Creating YAML configuration file...")

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		fmt.Println("🚫 Error marshaling YAML:", err)
		return
	}

	err = os.WriteFile(config.ConfigPath, data, 0600)
	if err != nil {
		fmt.Println("🚫 Error writing YAML configuration file:", err)
		return
	}

	fmt.Println("YAML configuration file has been created and populated at", config.ConfigPath)
}

func clearConsole() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
