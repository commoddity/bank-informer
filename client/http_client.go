package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var errResponseNotOK error = errors.New("Response not OK")

type retryTransport struct {
	underlying http.RoundTripper
	retries    int
}

func New() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &retryTransport{
			underlying: http.DefaultTransport,
			retries:    3,
		},
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.underlying
	if rt == nil {
		rt = http.DefaultTransport
	}

	var resp *http.Response
	var err error

	// Cache request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i <= t.retries; i++ {
		// Recreate body reader
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = rt.RoundTrip(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if i < t.retries {
			time.Sleep(time.Duration(i*i) * 100 * time.Millisecond)
		}
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Generic HTTP GET request
func Get[T any](endpoint string, header http.Header, httpClient *http.Client) (T, error) {
	var data T

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return data, err
	}

	// Set headers
	req.Header = header

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		code := resp.StatusCode
		text := http.StatusText(code)
		return data, fmt.Errorf("%s. %d %s", errResponseNotOK, code, text)
	}

	// Decode response body
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// Generic HTTP POST request
func Post[T any](endpoint string, header http.Header, postData []byte, httpClient *http.Client) (T, error) {
	var data T

	// Create a new request
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(postData))
	if err != nil {
		return data, err
	}

	// Set headers
	req.Header = header

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := resp.StatusCode
		text := http.StatusText(code)
		return data, fmt.Errorf("%s. %d %s", errResponseNotOK, code, text)
	}

	// Decode response body
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}
