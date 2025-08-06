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

var erc20TokenConfig = map[string]func(*JsonRPCRequest, string) float64{
	"USDC": func(requestBody *JsonRPCRequest, address string) float64 {
		requestBody.Method = "eth_call"
		requestBody.Params = json.RawMessage(fmt.Sprintf(`[{"to": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eb48", "data": "0x70a08231000000000000000000000000%s"}, "latest"]`, strings.TrimPrefix(address, "0x")))
		return 1e6
	},
	"USDT": func(requestBody *JsonRPCRequest, address string) float64 {
		requestBody.Method = "eth_call"
		requestBody.Params = json.RawMessage(fmt.Sprintf(`[{"to": "0xdAC17F958D2ee523a2206206994597C13D831ec7", "data": "0x70a08231000000000000000000000000%s"}, "latest"]`, strings.TrimPrefix(address, "0x")))
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

	JsonRPCBatchResponse []JsonRPCResponse

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
	PathApiUrl       string
	PathApiKey       string
	ETHWalletAddress string
	HttpClient       *http.Client
}

type Client struct {
	url          string
	pathAPIKey   string
	config       Config
	httpClient   *http.Client
	progressChan chan string
	mutex        *sync.Mutex
	waitGroup    *sync.WaitGroup
}

func NewClient(config Config, progressChan chan string, mutex *sync.Mutex, waitGroup *sync.WaitGroup) *Client {
	return &Client{
		url:          config.PathApiUrl,
		pathAPIKey:   config.PathApiKey,
		config:       config,
		httpClient:   config.HttpClient,
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
	// Prepare batch request for all tokens
	var batchRequest []JsonRPCRequest
	tokenIDMap := make(map[int]string)
	idCounter := 1

	for token := range balances {
		if _, ok := erc20TokenConfig[token]; !ok {
			continue
		}

		reqBody := JsonRPCRequest{Jsonrpc: "2.0", Id: idCounter}
		if getConfigFunc, ok := erc20TokenConfig[token]; ok {
			getConfigFunc(&reqBody, c.config.ETHWalletAddress)
		}

		batchRequest = append(batchRequest, reqBody)
		tokenIDMap[idCounter] = token
		idCounter++
	}

	if len(batchRequest) == 0 {
		return nil
	}

	// Execute batch request
	batchResponse, err := c.executeBatchRequest(batchRequest)
	if err != nil {
		return err
	}

	// Process responses and update balances
	for _, response := range batchResponse {
		token, exists := tokenIDMap[response.Id]
		if !exists {
			continue
		}

		if response.Error != nil {
			return fmt.Errorf("error for token %s: %s", token, response.Error.Message)
		}

		erc20WalletBalance, err := c.decodeHexToFloat64(response.Result)
		if err != nil {
			return fmt.Errorf("failed to decode balance for token %s: %w", token, err)
		}

		// Get the round value for this token
		roundValue := c.getRoundValueForToken(token)
		balances[token] = erc20WalletBalance / roundValue

		c.progressChan <- token
	}

	return nil
}

func (c *Client) executeBatchRequest(batchRequest []JsonRPCRequest) (JsonRPCBatchResponse, error) {
	const maxRetries = 5
	var lastErr error

	header := http.Header{
		"Content-Type":      []string{"application/json"},
		"Target-Service-Id": []string{"eth"},
		"Authorization":     []string{c.pathAPIKey},
	}

	jsonData, err := json.Marshal(batchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := client.Post[JsonRPCBatchResponse](c.url, header, jsonData, c.httpClient)
		if err != nil {
			lastErr = err
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("failed to execute batch request after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) getRoundValueForToken(token string) float64 {
	if getConfigFunc, ok := erc20TokenConfig[token]; ok {
		// Create a dummy request to get the round value
		dummyRequest := &JsonRPCRequest{}
		return getConfigFunc(dummyRequest, c.config.ETHWalletAddress)
	}
	return 1.0
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
