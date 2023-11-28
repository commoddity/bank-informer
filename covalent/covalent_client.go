package covalent

import (
	"fmt"
	"net/http"

	"github.com/commoddity/bank-informer/client"
)

const covalentProxyURL = "https://index.portal.pokt.network/eth-mainnet/address/%s/balances_v2/"

type Config struct {
	APIKey           string
	EthWalletAddress string
	EthURL           string
}

type Client struct {
	Config     Config
	HTTPClient *http.Client
}

type EthWalletBalance struct {
	Data struct {
		Address       string `json:"address"`
		UpdatedAt     string `json:"updated_at"`
		NextUpdateAt  string `json:"next_update_at"`
		QuoteCurrency string `json:"quote_currency"`
		ChainID       int    `json:"chain_id"`
		ChainName     string `json:"chain_name"`
		Items         []struct {
			ContractDecimals     int      `json:"contract_decimals"`
			ContractName         string   `json:"contract_name"`
			ContractTickerSymbol string   `json:"contract_ticker_symbol"`
			ContractAddress      string   `json:"contract_address"`
			SupportsErc          []string `json:"supports_erc"`
			LogoURL              string   `json:"logo_url"`
			ContractDisplayName  string   `json:"contract_display_name"`
			LogoUrls             struct {
				TokenLogoURL    string `json:"token_logo_url"`
				ProtocolLogoURL string `json:"protocol_logo_url"`
				ChainLogoURL    string `json:"chain_logo_url"`
			} `json:"logo_urls"`
			LastTransferredAt string  `json:"last_transferred_at"`
			NativeToken       bool    `json:"native_token"`
			Type              string  `json:"type"`
			IsSpam            bool    `json:"is_spam"`
			Balance           string  `json:"balance"`
			Balance24H        string  `json:"balance_24h"`
			QuoteRate         float64 `json:"quote_rate"`
			QuoteRate24H      float64 `json:"quote_rate_24h"`
			Quote             float64 `json:"quote"`
			PrettyQuote       string  `json:"pretty_quote"`
			Quote24H          float64 `json:"quote_24h"`
			PrettyQuote24H    string  `json:"pretty_quote_24h"`
			ProtocolMetadata  string  `json:"protocol_metadata"`
			NftData           string  `json:"nft_data"`
		} `json:"items"`
		Pagination string `json:"pagination"`
	} `json:"data"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
	ErrorCode    string `json:"error_code"`
}

func NewClient(config Config, httpClient *http.Client) *Client {
	config.EthURL = fmt.Sprintf(covalentProxyURL, config.EthWalletAddress)

	return &Client{
		Config:     config,
		HTTPClient: httpClient,
	}
}

func (c *Client) GetEthWalletBalance(balances map[string]float64) error {
	header := http.Header{}
	header.Set("x-api-key", c.Config.APIKey)

	ethWalletBalance, err := client.Get[EthWalletBalance](c.Config.EthURL, header, c.HTTPClient)
	if err != nil {
		return err
	}

	for _, item := range ethWalletBalance.Data.Items {
		balances[item.ContractTickerSymbol] = (item.Quote / item.QuoteRate)
	}

	return nil
}
