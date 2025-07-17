package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"alex/internal/mcp/transport"
)

// SpawnerType represents different types of server spawners
type SpawnerType string

const (
	SpawnerTypeNPX        SpawnerType = "npx"
	SpawnerTypeExecutable SpawnerType = "executable"
	SpawnerTypeDocker     SpawnerType = "docker"
)

// ServerConfig represents configuration for an MCP server
type ServerConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        SpawnerType       `json:"type"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	WorkDir     string            `json:"workDir"`
	AutoStart   bool              `json:"autoStart"`
	AutoRestart bool              `json:"autoRestart"`
	Timeout     time.Duration     `json:"timeout"`
	Enabled     bool              `json:"enabled"`
}

// Spawner interface for different server spawning strategies
type Spawner interface {
	Spawn(ctx context.Context, config *ServerConfig) (*transport.StdioTransport, error)
	Stop(ctx context.Context, transport *transport.StdioTransport) error
	IsRunning(transport *transport.StdioTransport) bool
}

// NPXSpawner implements spawning MCP servers via npx
type NPXSpawner struct {
	mu           sync.RWMutex
	activeServers map[string]*transport.StdioTransport
}

// NewNPXSpawner creates a new NPX spawner
func NewNPXSpawner() *NPXSpawner {
	return &NPXSpawner{
		activeServers: make(map[string]*transport.StdioTransport),
	}
}

// Spawn spawns an MCP server using npx
func (s *NPXSpawner) Spawn(ctx context.Context, config *ServerConfig) (*transport.StdioTransport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if server is already running
	if existing, exists := s.activeServers[config.ID]; exists {
		if existing.IsConnected() {
			return existing, nil
		}
		// Clean up dead connection
		delete(s.activeServers, config.ID)
	}

	// Validate NPX availability
	if err := s.validateNPX(); err != nil {
		return nil, fmt.Errorf("npx validation failed: %w", err)
	}

	// Prepare command and args
	var command string
	var args []string

	switch config.Type {
	case SpawnerTypeNPX:
		command = "npx"
		args = append([]string{"-y"}, config.Args...)
		if config.Command != "" {
			args = append(args, config.Command)
		}
	case SpawnerTypeExecutable:
		command = config.Command
		args = config.Args
	default:
		return nil, fmt.Errorf("unsupported spawner type: %s", config.Type)
	}

	// Prepare environment
	env := os.Environ()
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set working directory
	workDir := config.WorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// Create transport configuration
	transportConfig := &transport.StdioTransportConfig{
		Command: command,
		Args:    args,
		Env:     env,
		WorkDir: workDir,
	}

	// Create and connect transport
	stdioTransport := transport.NewStdioTransport(transportConfig)
	if err := stdioTransport.ConnectWithConfig(ctx, transportConfig); err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	// Store active server
	s.activeServers[config.ID] = stdioTransport

	return stdioTransport, nil
}

// Stop stops an MCP server
func (s *NPXSpawner) Stop(ctx context.Context, transport *transport.StdioTransport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if transport == nil {
		return nil
	}

	// Find and remove from active servers
	for id, t := range s.activeServers {
		if t == transport {
			delete(s.activeServers, id)
			break
		}
	}

	return transport.Disconnect()
}

// IsRunning checks if a transport is still running
func (s *NPXSpawner) IsRunning(transport *transport.StdioTransport) bool {
	if transport == nil {
		return false
	}
	return transport.IsConnected()
}

// validateNPX validates that npx is available
func (s *NPXSpawner) validateNPX() error {
	cmd := exec.Command("npx", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npx not found or not executable: %w", err)
	}
	return nil
}

// GetActiveServers returns a copy of active servers
func (s *NPXSpawner) GetActiveServers() map[string]*transport.StdioTransport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*transport.StdioTransport)
	for id, transport := range s.activeServers {
		result[id] = transport
	}
	return result
}

// ExecutableSpawner implements spawning MCP servers via local executables
type ExecutableSpawner struct {
	mu           sync.RWMutex
	activeServers map[string]*transport.StdioTransport
}

// NewExecutableSpawner creates a new executable spawner
func NewExecutableSpawner() *ExecutableSpawner {
	return &ExecutableSpawner{
		activeServers: make(map[string]*transport.StdioTransport),
	}
}

// Spawn spawns an MCP server using a local executable
func (s *ExecutableSpawner) Spawn(ctx context.Context, config *ServerConfig) (*transport.StdioTransport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if server is already running
	if existing, exists := s.activeServers[config.ID]; exists {
		if existing.IsConnected() {
			return existing, nil
		}
		// Clean up dead connection
		delete(s.activeServers, config.ID)
	}

	// Validate executable
	if err := s.validateExecutable(config.Command); err != nil {
		return nil, fmt.Errorf("executable validation failed: %w", err)
	}

	// Prepare environment
	env := os.Environ()
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set working directory
	workDir := config.WorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// Create transport configuration
	transportConfig := &transport.StdioTransportConfig{
		Command: config.Command,
		Args:    config.Args,
		Env:     env,
		WorkDir: workDir,
	}

	// Create and connect transport
	stdioTransport := transport.NewStdioTransport(transportConfig)
	if err := stdioTransport.ConnectWithConfig(ctx, transportConfig); err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	// Store active server
	s.activeServers[config.ID] = stdioTransport

	return stdioTransport, nil
}

// Stop stops an MCP server
func (s *ExecutableSpawner) Stop(ctx context.Context, transport *transport.StdioTransport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if transport == nil {
		return nil
	}

	// Find and remove from active servers
	for id, t := range s.activeServers {
		if t == transport {
			delete(s.activeServers, id)
			break
		}
	}

	return transport.Disconnect()
}

// IsRunning checks if a transport is still running
func (s *ExecutableSpawner) IsRunning(transport *transport.StdioTransport) bool {
	if transport == nil {
		return false
	}
	return transport.IsConnected()
}

// validateExecutable validates that the executable exists and is executable
func (s *ExecutableSpawner) validateExecutable(command string) error {
	// Check if it's an absolute path
	if filepath.IsAbs(command) {
		if _, err := os.Stat(command); err != nil {
			return fmt.Errorf("executable not found: %s", command)
		}
		return nil
	}

	// Check if it's in PATH
	if _, err := exec.LookPath(command); err != nil {
		return fmt.Errorf("executable not found in PATH: %s", command)
	}

	return nil
}

// GetActiveServers returns a copy of active servers
func (s *ExecutableSpawner) GetActiveServers() map[string]*transport.StdioTransport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*transport.StdioTransport)
	for id, transport := range s.activeServers {
		result[id] = transport
	}
	return result
}

// ServerManager manages multiple MCP server spawners
type ServerManager struct {
	spawners map[SpawnerType]Spawner
	mu       sync.RWMutex
}

// NewServerManager creates a new server manager
func NewServerManager() *ServerManager {
	return &ServerManager{
		spawners: map[SpawnerType]Spawner{
			SpawnerTypeNPX:        NewNPXSpawner(),
			SpawnerTypeExecutable: NewExecutableSpawner(),
		},
	}
}

// SpawnServer spawns an MCP server based on its configuration
func (m *ServerManager) SpawnServer(ctx context.Context, config *ServerConfig) (*transport.StdioTransport, error) {
	m.mu.RLock()
	spawner, exists := m.spawners[config.Type]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported spawner type: %s", config.Type)
	}

	return spawner.Spawn(ctx, config)
}

// StopServer stops an MCP server
func (m *ServerManager) StopServer(ctx context.Context, config *ServerConfig, transport *transport.StdioTransport) error {
	m.mu.RLock()
	spawner, exists := m.spawners[config.Type]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unsupported spawner type: %s", config.Type)
	}

	return spawner.Stop(ctx, transport)
}

// IsServerRunning checks if a server is running
func (m *ServerManager) IsServerRunning(config *ServerConfig, transport *transport.StdioTransport) bool {
	m.mu.RLock()
	spawner, exists := m.spawners[config.Type]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	return spawner.IsRunning(transport)
}

// GetNPXPackageCommand generates the npx command for common MCP packages
func GetNPXPackageCommand(packageName string) (string, []string) {
	// Common MCP server packages
	packages := map[string]string{
		"filesystem": "@modelcontextprotocol/server-filesystem",
		"memory":     "@modelcontextprotocol/server-memory",
		"github":     "@modelcontextprotocol/server-github",
		"gitlab":     "@modelcontextprotocol/server-gitlab",
		"sqlite":     "@modelcontextprotocol/server-sqlite",
		"postgres":   "@modelcontextprotocol/server-postgres",
		"brave":      "@modelcontextprotocol/server-brave-search",
		"youtube":    "@modelcontextprotocol/server-youtube-transcript",
		"puppeteer":  "@modelcontextprotocol/server-puppeteer",
		"docker":     "@modelcontextprotocol/server-docker",
		"kubernetes": "@modelcontextprotocol/server-kubernetes",
	}

	if fullPackage, exists := packages[packageName]; exists {
		return "npx", []string{"-y", fullPackage}
	}

	// If not a known package, assume it's a full package name
	if strings.Contains(packageName, "/") {
		return "npx", []string{"-y", packageName}
	}

	// Default to adding the MCP prefix
	return "npx", []string{"-y", "@modelcontextprotocol/server-" + packageName}
}

// ValidateServerConfig validates an MCP server configuration
func ValidateServerConfig(config *ServerConfig) error {
	if config.ID == "" {
		return fmt.Errorf("server ID is required")
	}

	if config.Name == "" {
		return fmt.Errorf("server name is required")
	}

	switch config.Type {
	case SpawnerTypeNPX:
		if len(config.Args) == 0 && config.Command == "" {
			return fmt.Errorf("NPX spawner requires either command or args")
		}
	case SpawnerTypeExecutable:
		if config.Command == "" {
			return fmt.Errorf("executable spawner requires command")
		}
	default:
		return fmt.Errorf("unsupported spawner type: %s", config.Type)
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	return nil
}