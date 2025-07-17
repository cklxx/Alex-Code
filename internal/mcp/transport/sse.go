package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"alex/internal/mcp/protocol"
)

// SSETransport implements MCP transport over Server-Sent Events
type SSETransport struct {
	endpoint    string
	client      *http.Client
	messagesCh  chan []byte
	errorsCh    chan error
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	connected   bool
	headers     map[string]string
	requestID   int64
	pendingReqs map[int64]chan *protocol.JSONRPCResponse
}

// SSETransportConfig represents configuration for SSE transport
type SSETransportConfig struct {
	Endpoint    string
	Headers     map[string]string
	Timeout     time.Duration
	RetryDelay  time.Duration
	MaxRetries  int
}

// NewSSETransport creates a new SSE transport instance
func NewSSETransport(config *SSETransportConfig) *SSETransport {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &SSETransport{
		endpoint:    config.Endpoint,
		client:      client,
		messagesCh:  make(chan []byte, 100),
		errorsCh:    make(chan error, 10),
		headers:     config.Headers,
		pendingReqs: make(map[int64]chan *protocol.JSONRPCResponse),
	}
}

// Connect establishes the SSE connection
func (t *SSETransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connected {
		return nil
	}

	t.ctx, t.cancel = context.WithCancel(ctx)

	// Create SSE connection for server-to-client messages
	go t.startSSEConnection()

	t.connected = true
	return nil
}

// Disconnect closes the SSE connection
func (t *SSETransport) Disconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil
	}

	if t.cancel != nil {
		t.cancel()
	}

	t.connected = false
	return nil
}

// SendRequest sends a JSON-RPC request via HTTP POST
func (t *SSETransport) SendRequest(req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	t.mu.RLock()
	connected := t.connected
	t.mu.RUnlock()

	if !connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Serialize request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(t.ctx, "POST", t.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}

	// For requests with ID, set up response channel
	var responseCh chan *protocol.JSONRPCResponse
	if req.ID != nil {
		if id, ok := req.ID.(int64); ok {
			responseCh = make(chan *protocol.JSONRPCResponse, 1)
			t.mu.Lock()
			t.pendingReqs[id] = responseCh
			t.mu.Unlock()

			defer func() {
				t.mu.Lock()
				delete(t.pendingReqs, id)
				t.mu.Unlock()
			}()
		}
	}

	// Send HTTP request
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// For notifications (no ID), return immediately
	if req.ID == nil {
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, nil
	}

	// For requests with ID, wait for response via SSE
	if responseCh != nil {
		select {
		case response := <-responseCh:
			return response, nil
		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("request timeout")
		case <-t.ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		}
	}

	return nil, fmt.Errorf("no response channel for request")
}

// SendNotification sends a JSON-RPC notification
func (t *SSETransport) SendNotification(notification *protocol.JSONRPCNotification) error {
	t.mu.RLock()
	connected := t.connected
	t.mu.RUnlock()

	if !connected {
		return fmt.Errorf("transport not connected")
	}

	// Serialize notification
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(t.ctx, "POST", t.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}

	// Send HTTP request
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// ReceiveMessages returns a channel for receiving messages
func (t *SSETransport) ReceiveMessages() <-chan []byte {
	return t.messagesCh
}

// ReceiveErrors returns a channel for receiving errors
func (t *SSETransport) ReceiveErrors() <-chan error {
	return t.errorsCh
}

// startSSEConnection establishes and maintains the SSE connection
func (t *SSETransport) startSSEConnection() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			if err := t.connectSSE(); err != nil {
				t.errorsCh <- fmt.Errorf("SSE connection failed: %w", err)
				select {
				case <-t.ctx.Done():
					return
				case <-time.After(1 * time.Second):
					continue
				}
			}
		}
	}
}

// connectSSE establishes the SSE connection
func (t *SSETransport) connectSSE() error {
	// Create SSE endpoint URL (typically /sse or /events)
	sseEndpoint := strings.TrimSuffix(t.endpoint, "/messages") + "/sse"

	req, err := http.NewRequestWithContext(t.ctx, "GET", sseEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	// Set SSE headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected SSE status code: %d", resp.StatusCode)
	}

	// Read SSE stream
	scanner := bufio.NewScanner(resp.Body)
	var eventData strings.Builder

	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return nil
		default:
		}

		line := scanner.Text()

		// Handle SSE format
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "" {
				continue
			}
			eventData.WriteString(data)
			eventData.WriteString("\n")
		} else if line == "" {
			// End of event
			if eventData.Len() > 0 {
				eventStr := strings.TrimSpace(eventData.String())
				if eventStr != "" {
					t.handleSSEMessage([]byte(eventStr))
				}
				eventData.Reset()
			}
		}
	}

	return scanner.Err()
}

// handleSSEMessage processes incoming SSE messages
func (t *SSETransport) handleSSEMessage(data []byte) {
	// Try to parse as JSON-RPC response first
	if protocol.IsResponse(data) {
		var response protocol.JSONRPCResponse
		if err := json.Unmarshal(data, &response); err == nil {
			t.handleResponse(&response)
			return
		}
	}

	// Try to parse as JSON-RPC notification
	if protocol.IsNotification(data) {
		select {
		case t.messagesCh <- data:
		case <-t.ctx.Done():
			return
		}
		return
	}

	// Forward raw message
	select {
	case t.messagesCh <- data:
	case <-t.ctx.Done():
		return
	}
}

// handleResponse handles JSON-RPC responses
func (t *SSETransport) handleResponse(response *protocol.JSONRPCResponse) {
	if response.ID == nil {
		return
	}

	id, ok := response.ID.(float64)
	if !ok {
		return
	}

	t.mu.Lock()
	responseCh, exists := t.pendingReqs[int64(id)]
	t.mu.Unlock()

	if exists {
		select {
		case responseCh <- response:
		case <-t.ctx.Done():
			return
		}
	}
}

// IsConnected returns the connection status
func (t *SSETransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// NextRequestID generates a new request ID
func (t *SSETransport) NextRequestID() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requestID++
	return t.requestID
}