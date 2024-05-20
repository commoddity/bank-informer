package pokt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/commoddity/bank-informer/client"
)

type Config struct {
	PortalAppID       string
	SecretKey         string
	POKTWalletAddress string
}

type Client struct {
	Config       Config
	url          string
	secretKey    string
	httpClient   *http.Client
	progressChan chan string
	mutex        *sync.Mutex
	waitGroup    *sync.WaitGroup
}

type queryBalanceOutput struct {
	Balance *big.Int `json:"balance"`
}

func NewClient(config Config, httpClient *http.Client, progressChan chan string, mutex *sync.Mutex, waitGroup *sync.WaitGroup) *Client {
	url := fmt.Sprintf("https://mainnet.rpc.grove.city/v1/%s/v1/query/balance", config.PortalAppID)

	return &Client{
		Config:       config,
		url:          url,
		secretKey:    config.SecretKey,
		httpClient:   httpClient,
		progressChan: progressChan,
		mutex:        mutex,
		waitGroup:    waitGroup,
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

func (c *Client) GetWalletBalance(balances map[string]float64) error {
	var balance *big.Int
	var highestBalance *big.Int

	// Create a channel to receive balance results
	balanceChan := make(chan *big.Int, 5)
	errorChan := make(chan error, 5)

	for i := 0; i < 5; i++ {
		c.waitGroup.Add(1)
		go func() {
			defer c.waitGroup.Done()
			balance, err := c.getPOKTWalletBalance(c.Config.POKTWalletAddress)
			if err != nil {
				errorChan <- err
				return
			}
			balanceChan <- balance
		}()
	}

	// Wait for all goroutines to finish
	c.waitGroup.Wait()
	close(balanceChan)
	close(errorChan)

	// Check if there were any errors
	if len(errorChan) > 0 {
		return <-errorChan
	}

	// Process the balance results
	for balance = range balanceChan {
		// If it's the first iteration or the current balance is higher than the highest, update the highest balance
		if highestBalance == nil || balance.Cmp(highestBalance) > 0 {
			highestBalance = balance
		}
	}

	// Convert balance to float64 and divide by 1e6 to get the correct value
	balanceFloat := new(big.Float).SetInt(highestBalance)
	balanceFloat.Quo(balanceFloat, big.NewFloat(1e6))
	balanceValue, _ := balanceFloat.Float64()

	c.progressChan <- "POKT"

	// Modify the passed map with the balance
	c.mutex.Lock()
	balances["POKT"] = balanceValue
	c.mutex.Unlock()

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
		fmt.Println("ERROR HERE", err)
		return nil, err
	}

	return resp.Balance, nil
}
