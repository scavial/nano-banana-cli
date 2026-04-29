// Package nano provides a client for interacting with the Nano cryptocurrency network.
package nano

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultNodeURL is the default Nano node RPC endpoint.
	DefaultNodeURL = "http://localhost:7076"
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client represents a Nano RPC client.
type Client struct {
	nodeURL    string
	httpClient *http.Client
}

// RPCRequest represents a generic Nano RPC request payload.
type RPCRequest struct {
	Action string `json:"action"`
	// Additional fields are embedded per-action via anonymous structs or map embedding.
}

// RPCError represents an error returned by the Nano RPC node.
type RPCError struct {
	Error string `json:"error"`
}

// NewClient creates a new Nano RPC client targeting the given node URL.
// If nodeURL is empty, DefaultNodeURL is used.
func NewClient(nodeURL string) *Client {
	if nodeURL == "" {
		nodeURL = DefaultNodeURL
	}
	return &Client{
		nodeURL: nodeURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// call sends a JSON-encoded request to the Nano node and decodes the response into dest.
func (c *Client) call(ctx context.Context, payload interface{}, dest interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("nano: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.nodeURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("nano: failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nano: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("nano: failed to read response body: %w", err)
	}

	// Check for node-level RPC errors before decoding into dest.
	var rpcErr RPCError
	if json.Unmarshal(respBody, &rpcErr) == nil && rpcErr.Error != "" {
		return fmt.Errorf("nano: RPC error: %s", rpcErr.Error)
	}

	if err := json.Unmarshal(respBody, dest); err != nil {
		return fmt.Errorf("nano: failed to decode response: %w", err)
	}
	return nil
}

// AccountBalanceResponse holds the balance information for a Nano account.
type AccountBalanceResponse struct {
	Balance    string `json:"balance"`
	Pending    string `json:"pending"`
	Receivable string `json:"receivable"`
}

// AccountBalance retrieves the balance of the given Nano account address.
func (c *Client) AccountBalance(ctx context.Context, account string) (*AccountBalanceResponse, error) {
	payload := struct {
		Action  string `json:"action"`
		Account string `json:"account"`
	}{
		Action:  "account_balance",
		Account: account,
	}

	var result AccountBalanceResponse
	if err := c.call(ctx, payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BlockCountResponse holds the current block count statistics from the node.
type BlockCountResponse struct {
	Count     string `json:"count"`
	Unchecked string `json:"unchecked"`
	Cemented  string `json:"cemented"`
}

// BlockCount retrieves the current block count from the Nano node.
func (c *Client) BlockCount(ctx context.Context) (*BlockCountResponse, error) {
	payload := struct {
		Action string `json:"action"`
	}{
		Action: "block_count",
	}

	var result BlockCountResponse
	if err := c.call(ctx, payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
