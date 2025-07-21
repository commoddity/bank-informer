package pokt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/commoddity/bank-informer/client"
)

type Config struct {
	PathApiUrl         string
	PathApiKey         string
	POKTWalletAddress  string
	PoktExchangeAmount int64
	HttpClient         *http.Client
}

type Client struct {
	Config       Config
	baseUrl      string
	pathAPIKey   string
	httpClient   *http.Client
	progressChan chan string
	mutex        *sync.Mutex
	waitGroup    *sync.WaitGroup
}

type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type queryBalanceOutput struct {
	Balances   []Balance `json:"balances"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

func NewClient(config Config, progressChan chan string, mutex *sync.Mutex, waitGroup *sync.WaitGroup) *Client {
	baseUrl := fmt.Sprintf("%s/cosmos/bank/v1beta1/balances", config.PathApiUrl)

	return &Client{
		Config:       config,
		baseUrl:      baseUrl,
		pathAPIKey:   config.PathApiKey,
		httpClient:   config.HttpClient,
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
	var successfulAttempts int

	// Create a channel to receive balance results
	balanceChan := make(chan *big.Int, 5)
	errorChan := make(chan error, 5)

	for i := 0; i < 5; i++ {
		c.waitGroup.Add(1)
		go func() {
			defer c.waitGroup.Done()
			var balance *big.Int
			var err error
			for attempt := 0; attempt < 5; attempt++ {
				balance, err = c.getPOKTWalletBalance(c.Config.POKTWalletAddress)
				if err == nil {
					balanceChan <- balance
					return
				}
			}
			errorChan <- err
		}()
	}

	// Wait for all goroutines to finish
	c.waitGroup.Wait()
	close(balanceChan)
	close(errorChan)

	// Process the balance results
	for balance = range balanceChan {
		successfulAttempts++
		// If it's the first iteration or the current balance is higher than the highest, update the highest balance
		if highestBalance == nil || balance.Cmp(highestBalance) > 0 {
			highestBalance = balance
		}
	}

	// If there were no successful attempts, return an error
	if successfulAttempts == 0 {
		return <-errorChan
	}

	// Convert balance to float64 and divide by 1e6 to get the correct value
	balanceFloat := new(big.Float).SetInt(highestBalance)
	balanceFloat.Quo(balanceFloat, big.NewFloat(1e6))
	balanceValue, _ := balanceFloat.Float64()

	// Add exchange amount if configured
	if c.Config.PoktExchangeAmount > 0 {
		balanceValue += float64(c.Config.PoktExchangeAmount)
	}

	c.progressChan <- "POKT"

	// Modify the passed map with the balance
	c.mutex.Lock()
	balances["POKT"] = balanceValue
	c.mutex.Unlock()

	return nil
}

func (c *Client) getPOKTWalletBalance(address string) (*big.Int, error) {
	url := fmt.Sprintf("%s/%s", c.baseUrl, address)

	header := http.Header{
		"Target-Service-Id": []string{"pocket"},
		"Authorization":     []string{c.pathAPIKey},
	}

	resp, err := client.Get[queryBalanceOutput](url, header, c.httpClient)
	if err != nil {
		return nil, err
	}

	// Find the upokt balance in the balances array
	for _, balance := range resp.Balances {
		if balance.Denom == "upokt" {
			amount := new(big.Int)
			amount, ok := amount.SetString(balance.Amount, 10)
			if !ok {
				return nil, fmt.Errorf("failed to parse balance amount: %s", balance.Amount)
			}
			return amount, nil
		}
	}

	return nil, fmt.Errorf("upokt balance not found")
}
