package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"alex/internal/mcp/protocol"
)

// StdioTransport implements MCP transport over standard input/output
type StdioTransport struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	messagesCh  chan []byte
	errorsCh    chan error
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	connected   bool
	requestID   int64
	pendingReqs map[int64]chan *protocol.JSONRPCResponse
	config      *StdioTransportConfig
}

// StdioTransportConfig represents configuration for stdio transport
type StdioTransportConfig struct {
	Command string
	Args    []string
	Env     []string
	WorkDir string
}

// NewStdioTransport creates a new stdio transport instance
func NewStdioTransport(config *StdioTransportConfig) *StdioTransport {
	return &StdioTransport{
		messagesCh:  make(chan []byte, 100),
		errorsCh:    make(chan error, 10),
		pendingReqs: make(map[int64]chan *protocol.JSONRPCResponse),
		config:      config,
	}
}

// Connect establishes the stdio connection by starting the MCP server process
func (t *StdioTransport) Connect(ctx context.Context) error {
	return t.ConnectWithConfig(ctx, t.config)
}

// ConnectWithConfig establishes the stdio connection with configuration
func (t *StdioTransport) ConnectWithConfig(ctx context.Context, config *StdioTransportConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connected {
		return nil
	}

	if config == nil {
		config = t.config
	}
	if config == nil {
		return fmt.Errorf("no configuration provided")
	}

	t.ctx, t.cancel = context.WithCancel(ctx)

	// Create command
	t.cmd = exec.CommandContext(t.ctx, config.Command, config.Args...)
	t.cmd.Env = config.Env
	if config.WorkDir != "" {
		t.cmd.Dir = config.WorkDir
	}

	// Set up pipes
	stdin, err := t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	t.stdin = stdin

	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	t.stdout = stdout

	stderr, err := t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	t.stderr = stderr

	// Start the process
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Start reading from stdout and stderr
	go t.readStdout()
	go t.readStderr()

	// Monitor process
	go t.monitorProcess()

	t.connected = true
	return nil
}

// Disconnect closes the stdio connection
func (t *StdioTransport) Disconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil
	}

	if t.cancel != nil {
		t.cancel()
	}

	if t.stdin != nil {
		t.stdin.Close()
	}

	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}

	t.connected = false
	return nil
}

// SendRequest sends a JSON-RPC request via stdin
func (t *StdioTransport) SendRequest(req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	t.mu.RLock()
	connected := t.connected
	stdin := t.stdin
	t.mu.RUnlock()

	if !connected || stdin == nil {
		return nil, fmt.Errorf("transport not connected")
	}

	// Serialize request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
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

	// Send request
	if _, err := stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// For notifications (no ID), return immediately
	if req.ID == nil {
		return nil, nil
	}

	// For requests with ID, wait for response
	if responseCh != nil {
		select {
		case response := <-responseCh:
			return response, nil
		case <-t.ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		}
	}

	return nil, fmt.Errorf("no response channel for request")
}

// SendNotification sends a JSON-RPC notification via stdin
func (t *StdioTransport) SendNotification(notification *protocol.JSONRPCNotification) error {
	t.mu.RLock()
	connected := t.connected
	stdin := t.stdin
	t.mu.RUnlock()

	if !connected || stdin == nil {
		return fmt.Errorf("transport not connected")
	}

	// Serialize notification
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Send notification
	if _, err := stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

// ReceiveMessages returns a channel for receiving messages
func (t *StdioTransport) ReceiveMessages() <-chan []byte {
	return t.messagesCh
}

// ReceiveErrors returns a channel for receiving errors
func (t *StdioTransport) ReceiveErrors() <-chan error {
	return t.errorsCh
}

// readStdout reads messages from stdout
func (t *StdioTransport) readStdout() {
	defer func() {
		t.mu.Lock()
		if t.stdout != nil {
			t.stdout.Close()
		}
		t.mu.Unlock()
	}()

	scanner := bufio.NewScanner(t.stdout)
	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Copy the line to avoid scanner buffer reuse issues
		message := make([]byte, len(line))
		copy(message, line)

		t.handleMessage(message)
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errorsCh <- fmt.Errorf("stdout read error: %w", err):
		case <-t.ctx.Done():
		}
	}
}

// readStderr reads error messages from stderr
func (t *StdioTransport) readStderr() {
	defer func() {
		t.mu.Lock()
		if t.stderr != nil {
			t.stderr.Close()
		}
		t.mu.Unlock()
	}()

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line != "" {
			select {
			case t.errorsCh <- fmt.Errorf("stderr: %s", line):
			case <-t.ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errorsCh <- fmt.Errorf("stderr read error: %w", err):
		case <-t.ctx.Done():
		}
	}
}

// monitorProcess monitors the MCP server process
func (t *StdioTransport) monitorProcess() {
	if t.cmd == nil {
		return
	}

	err := t.cmd.Wait()
	if err != nil {
		select {
		case t.errorsCh <- fmt.Errorf("MCP server process exited: %w", err):
		case <-t.ctx.Done():
		}
	}

	t.mu.Lock()
	t.connected = false
	t.mu.Unlock()
}

// handleMessage processes incoming messages
func (t *StdioTransport) handleMessage(data []byte) {
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
func (t *StdioTransport) handleResponse(response *protocol.JSONRPCResponse) {
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
func (t *StdioTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// NextRequestID generates a new request ID
func (t *StdioTransport) NextRequestID() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requestID++
	return t.requestID
}