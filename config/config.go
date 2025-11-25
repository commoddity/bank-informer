package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	bankInformerFolder = "bank-informer"
	configFilename     = ".bankinformer.config.yaml"
	dbPath             = "db"
	csvFilename        = "crypto_values.csv"

	defaultCryptoFiatConversion = "USD"
	defaultConvertCurrencies    = "USD"
	defaultCryptoValues         = "USDC,ETH,POKT"

	// RPCURLTemplate is the template for building Pocket Network API URLs
	// Format: https://<SUBDOMAIN>.api.pocket.network
	RPCURLTemplate = "https://%s.api.pocket.network"
)

var (
	homeDir, _ = os.UserHomeDir()
	ConfigPath = filepath.Join(homeDir, bankInformerFolder, configFilename)
	DBPath     = filepath.Join(homeDir, bankInformerFolder, dbPath)
	CSVPath    = filepath.Join(homeDir, bankInformerFolder, csvFilename)
)

// Config represents the configuration settings for the Bank Informer service.
type Config struct {
	RpcApiKey            string   `yaml:"rpc_api_key"`            // optional
	EthWalletAddress     string   `yaml:"eth_wallet_address"`     // required
	PoktWalletAddress    string   `yaml:"pokt_wallet_address"`    // required
	CMCAPIKey            string   `yaml:"cmc_api_key"`            // required
	PoktExchangeAmount   int64    `yaml:"pokt_exchange_amount"`   // optional
	CryptoFiatConversion string   `yaml:"crypto_fiat_conversion"` // optional, defaults to "USD"
	ConvertCurrencies    []string `yaml:"convert_currencies"`     // optional, defaults to "USD"
	CryptoValues         []string `yaml:"crypto_values"`          // optional, defaults to "USDC,ETH,POKT"
}

// LoadConfig loads the Bank Informer configuration from a YAML file,
// assigns default values for optional fields, and validates required fields.
func LoadConfig() (*Config, error) {
	file, err := os.Open(ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Validate required fields and assign defaults for optional ones.
	if err := config.validateAndSetDefaults(); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateAndSetDefaults checks that all required fields are provided,
// and assigns default values to any missing optional fields.
func (c *Config) validateAndSetDefaults() error {
	if c.EthWalletAddress == "" {
		return fmt.Errorf("missing required field: eth_wallet_address")
	}
	if c.PoktWalletAddress == "" {
		return fmt.Errorf("missing required field: pokt_wallet_address")
	}
	if c.CMCAPIKey == "" {
		return fmt.Errorf("missing required field: cmc_api_key")
	}
	if c.CryptoFiatConversion == "" {
		c.CryptoFiatConversion = defaultCryptoFiatConversion
	}
	if len(c.ConvertCurrencies) == 0 {
		c.ConvertCurrencies = []string{defaultConvertCurrencies}
	}
	if len(c.CryptoValues) == 0 {
		c.CryptoValues = []string{defaultCryptoValues}
	}
	return nil
}
