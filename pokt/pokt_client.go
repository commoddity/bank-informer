package pokt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/commoddity/bank-informer/client"
)

const poktGrovePortalURL = "https://mainnet.rpc.grove.city/v1/%s/%s"

type Config struct {
	PortalAppID       string
	SecretKey         string
	POKTWalletAddress string
}

type Client struct {
	Config     Config
	url        string
	secretKey  string
	httpClient *http.Client
}

type queryBalanceOutput struct {
	Balance *big.Int `json:"balance"`
}

func NewClient(config Config, httpClient *http.Client) *Client {
	url := fmt.Sprintf(poktGrovePortalURL, config.PortalAppID, "v1/query/balance")

	return &Client{
		Config:     config,
		url:        url,
		secretKey:  config.SecretKey,
		httpClient: httpClient,
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

func ValidateSecretKey(key string) error {
	if len(key) != 32 {
		return fmt.Errorf("invalid Secret Key: %s", key)
	}
	if _, err := hex.DecodeString(key); err != nil {
		return fmt.Errorf("invalid Secret Key: %s", key)
	}
	return nil
}

func (p *Client) GetWalletBalance(balances map[string]float64) error {
	var balance *big.Int
	var err error
	var highestBalance *big.Int

	for i := 0; i < 5; i++ {
		balance, err = p.getPOKTWalletBalance(p.Config.POKTWalletAddress)
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

func (c *Client) getPOKTWalletBalance(address string) (*big.Int, error) {
	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	if c.secretKey != "" {
		header["Authorization"] = []string{c.secretKey}
	}

	params := map[string]any{
		"address": address,
	}

	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := client.Post[queryBalanceOutput](c.url, header, reqBody, c.httpClient)
	if err != nil {
		return nil, err
	}

	return resp.Balance, nil
}
