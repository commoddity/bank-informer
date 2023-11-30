package pokt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/pokt-foundation/pocket-go/provider"
)

const poktURL = "https://mainnet.gateway.pokt.network/v1/%s"

type Config struct {
	PortalAppID       string
	POKTWalletAddress string
}

type Client struct {
	Config   Config
	Provider *provider.Provider
}

func NewClient(config Config, httpClient *http.Client) *Client {
	url := fmt.Sprintf(poktURL, config.PortalAppID)

	return &Client{
		Config:   config,
		Provider: provider.NewProvider(url, []string{url}),
	}
}

func ValidatePortalAppID(id string) error {
	if len(id) != 8 && len(id) != 24 {
		return fmt.Errorf("invalid Portal App ID: %s", id)
	}
	if _, err := hex.DecodeString(id); err != nil {
		return fmt.Errorf("invalid Portal App ID: %s", id)
	}
	return nil
}

func (p *Client) GetWalletBalance(balances map[string]float64) error {
	var balance *big.Int
	var err error
	retries := 0
	for balance == nil && retries < 5 {
		balance, err = p.Provider.GetBalance(p.Config.POKTWalletAddress, nil)
		if err != nil {
			return err
		}
		retries++
	}

	// Convert balance to float64 and divide by 1e6 to get the correct value
	balanceFloat := new(big.Float).SetInt(balance)
	balanceFloat.Quo(balanceFloat, big.NewFloat(1e6))
	balanceValue, _ := balanceFloat.Float64()

	// Modify the passed map with the balance
	balances["POKT"] = balanceValue

	return nil
}
