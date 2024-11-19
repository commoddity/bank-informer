package eth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/commoddity/bank-informer/client"
)

const (
	ethGrovePortalURL = "https://eth-mainnet.rpc.grove.city/v1/%s"
)

var erc20TokenConfig = map[string]func(*JsonRPCRequest, string) float64{
	"USDC": func(requestBody *JsonRPCRequest, address string) float64 {
		requestBody.Method = "eth_call"
		requestBody.Params = json.RawMessage(fmt.Sprintf(`[{"to": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eb48", "data": "0x70a08231000000000000000000000000%s"}, "latest"]`, strings.TrimPrefix(address, "0x")))
		return 1e6
	},
	"ETH": func(requestBody *JsonRPCRequest, address string) float64 {
		requestBody.Method = "eth_getBalance"
		requestBody.Params = json.RawMessage(fmt.Sprintf(`["%s", "latest"]`, address))
		return 1e18
	},
	"WPOKT": func(requestBody *JsonRPCRequest, address string) float64 {
		requestBody.Method = "eth_call"
		requestBody.Params = json.RawMessage(fmt.Sprintf(`[{"to": "0x67F4C72a50f8Df6487720261E188F2abE83F57D7", "data": "0x70a08231000000000000000000000000%s"}, "latest"]`, strings.TrimPrefix(address, "0x")))
		return 1e6
	},
	"WBTC": func(requestBody *JsonRPCRequest, address string) float64 {
		requestBody.Method = "eth_call"
		requestBody.Params = json.RawMessage(fmt.Sprintf(`[{"to": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599", "data": "0x70a08231000000000000000000000000%s"}, "latest"]`, strings.TrimPrefix(address, "0x")))
		return 1e8
	},
}

type (
	JsonRPCRequest struct {
		Jsonrpc string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		Id      int             `json:"id"`
	}

	JsonRPCResponse struct {
		Id      int           `json:"id"`
		Jsonrpc string        `json:"jsonrpc"`
		Result  string        `json:"result"`
		Error   *JsonRPCError `json:"error,omitempty"`
	}

	JsonRPCError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data,omitempty"`
	}
)

// UnmarshalJSON unmarshals a JsonRPCResponse from JSON.
func (r *JsonRPCResponse) UnmarshalJSON(data []byte) error {
	type Alias JsonRPCResponse
	aux := &struct {
		Error interface{} `json:"error,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.Error.(type) {
	case nil:
		// If error is nil, set r.Error to nil
		r.Error = nil
	case map[string]interface{}:
		// Attempt to unmarshal into JsonRPCError
		var jsonRPCError JsonRPCError
		errData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(errData, &jsonRPCError); err != nil {
			return err
		}
		r.Error = &jsonRPCError
	case string:
		// If it's a string, set the error message
		r.Error = &JsonRPCError{Message: v}
	default:
		// Handle unexpected types
		return fmt.Errorf("unexpected error type: %T", v)
	}

	return nil
}

type Config struct {
	PortalAppID      string
	SecretKey        string
	ETHWalletAddress string
	HTTPClient       *http.Client
}

type Client struct {
	url          string
	secretKey    string
	config       Config
	httpClient   *http.Client
	progressChan chan string
	mutex        *sync.Mutex
	waitGroup    *sync.WaitGroup
}

func NewClient(config Config, httpClient *http.Client, progressChan chan string, mutex *sync.Mutex, waitGroup *sync.WaitGroup) *Client {
	url := fmt.Sprintf(ethGrovePortalURL, config.PortalAppID)

	return &Client{
		url:          url,
		secretKey:    config.SecretKey,
		config:       config,
		httpClient:   httpClient,
		progressChan: progressChan,
		mutex:        mutex,
		waitGroup:    waitGroup,
	}
}

func ValidateETHWalletAddress(address string) error {
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return fmt.Errorf("invalid Ethereum wallet address: %s", address)
	}
	return nil
}

func (c *Client) GetETHWalletBalances(balances map[string]float64) error {
	errorChan := make(chan error, len(balances))

	for token := range balances {
		if _, ok := erc20TokenConfig[token]; !ok {
			continue
		}

		c.waitGroup.Add(1)
		go func(token string) {
			defer c.waitGroup.Done()

			balance, err := c.getETHWalletBalance(token)
			if err != nil {
				errorChan <- err
				return
			}

			c.mutex.Lock()
			balances[token] = balance
			c.mutex.Unlock()

			c.progressChan <- token
		}(token)
	}

	c.waitGroup.Wait()
	close(errorChan)

	// Check if there were any errors
	if len(errorChan) > 0 {
		return <-errorChan
	}

	return nil
}

func (c *Client) getETHWalletBalance(erc20Token string) (float64, error) {
	const maxRetries = 5
	var lastErr error

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	if c.secretKey != "" {
		header["Authorization"] = []string{c.secretKey}
	}

	reqBody, roundValue, err := c.getJsonRPCRequest(erc20Token)
	if err != nil {
		return 0, err
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := client.Post[JsonRPCResponse](c.url, header, reqBody, c.httpClient)
		if err != nil {
			lastErr = err
			continue
		}

		erc20WalletBalance, err := c.decodeHexToFloat64(resp.Result)
		if err != nil {
			lastErr = err
			continue
		}

		return erc20WalletBalance / roundValue, nil
	}

	return 0, fmt.Errorf("failed to get wallet balance after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) getJsonRPCRequest(token string) ([]byte, float64, error) {
	requestBody := &JsonRPCRequest{Jsonrpc: "2.0", Id: 1}

	var roundValue float64

	if getConfigFunc, ok := erc20TokenConfig[token]; ok {
		roundValue = getConfigFunc(requestBody, c.config.ETHWalletAddress)
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, 0, err
	}

	return jsonData, roundValue, nil
}

func (c *Client) decodeHexToFloat64(hexValue string) (float64, error) {
	// Check if the hexValue is empty
	if hexValue == "" {
		return 0, fmt.Errorf("empty result field")
	}

	// Remove the "0x" prefix before parsing
	hexValue = strings.TrimPrefix(hexValue, "0x")

	value, err := strconv.ParseInt(hexValue, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hex value: %w", err)
	}

	return float64(value), nil
}
