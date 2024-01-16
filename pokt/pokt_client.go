package pokt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/pokt-foundation/pocket-go/provider"
)

const poktGrovePortalURL = "https://mainnet.rpc.grove.city/v1/%s"

type Config struct {
	PortalAppID       string
	POKTWalletAddress string
}

type Client struct {
	Config   Config
	Provider *provider.Provider
}

func NewClient(config Config, httpClient *http.Client) *Client {
	url := fmt.Sprintf(poktGrovePortalURL, config.PortalAppID)

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
	var highestBalance *big.Int

	for i := 0; i < 3; i++ {
		balance, err = p.Provider.GetBalance(p.Config.POKTWalletAddress, nil)
		if err != nil {
			return err
		}

		// If it's the first iteration or the current balance is higher than the highest, update the highest balance
		if i == 0 || balance.Cmp(highestBalance) > 0 {
			highestBalance = balance
		}
	}

	// Convert balance to float64 and divide by 1e6 to get the correct value
	balanceFloat := new(big.Float).SetInt(balance)
	balanceFloat.Quo(balanceFloat, big.NewFloat(1e6))
	balanceValue, _ := balanceFloat.Float64()

	// Modify the passed map with the balance
	balances["POKT"] = balanceValue

	return nil
}
