package mcp

import (
	"context"
	"testing"
	"time"
)

func TestValidateServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ServerConfig
		wantErr bool
	}{
		{
			name: "valid npx config",
			config: &ServerConfig{
				ID:      "test",
				Name:    "Test Server",
				Type:    SpawnerTypeNPX,
				Command: "@modelcontextprotocol/server-memory",
				Enabled: true,
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid executable config",
			config: &ServerConfig{
				ID:      "test",
				Name:    "Test Server",
				Type:    SpawnerTypeExecutable,
				Command: "node",
				Args:    []string{"server.js"},
				Enabled: true,
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			config: &ServerConfig{
				Name:    "Test Server",
				Type:    SpawnerTypeNPX,
				Command: "@modelcontextprotocol/server-memory",
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			config: &ServerConfig{
				ID:      "test",
				Type:    SpawnerTypeNPX,
				Command: "@modelcontextprotocol/server-memory",
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			config: &ServerConfig{
				ID:      "test",
				Name:    "Test Server",
				Type:    "invalid",
				Command: "@modelcontextprotocol/server-memory",
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "npx missing command and args",
			config: &ServerConfig{
				ID:      "test",
				Name:    "Test Server",
				Type:    SpawnerTypeNPX,
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "executable missing command",
			config: &ServerConfig{
				ID:      "test",
				Name:    "Test Server",
				Type:    SpawnerTypeExecutable,
				Enabled: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetNPXPackageCommand(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		wantCommand string
		wantArgs    []string
	}{
		{
			name:        "known package - filesystem",
			packageName: "filesystem",
			wantCommand: "npx",
			wantArgs:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
		},
		{
			name:        "known package - memory",
			packageName: "memory",
			wantCommand: "npx",
			wantArgs:    []string{"-y", "@modelcontextprotocol/server-memory"},
		},
		{
			name:        "full package name",
			packageName: "@modelcontextprotocol/server-custom",
			wantCommand: "npx",
			wantArgs:    []string{"-y", "@modelcontextprotocol/server-custom"},
		},
		{
			name:        "unknown package",
			packageName: "unknown",
			wantCommand: "npx",
			wantArgs:    []string{"-y", "@modelcontextprotocol/server-unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCommand, gotArgs := GetNPXPackageCommand(tt.packageName)
			if gotCommand != tt.wantCommand {
				t.Errorf("GetNPXPackageCommand() command = %v, want %v", gotCommand, tt.wantCommand)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("GetNPXPackageCommand() args length = %v, want %v", len(gotArgs), len(tt.wantArgs))
				return
			}
			for i, arg := range gotArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("GetNPXPackageCommand() args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestServerManagerBasic(t *testing.T) {
	manager := NewServerManager()

	if manager == nil {
		t.Fatal("NewServerManager() returned nil")
	}

	// Test with invalid config
	invalidConfig := &ServerConfig{
		ID:   "test",
		Type: "invalid",
	}

	_, err := manager.SpawnServer(context.Background(), invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid spawner type")
	}
}

func TestNPXSpawner(t *testing.T) {
	spawner := NewNPXSpawner()

	if spawner == nil {
		t.Fatal("NewNPXSpawner() returned nil")
	}

	// Test with nil transport
	if spawner.IsRunning(nil) {
		t.Error("Expected IsRunning to return false for nil transport")
	}

	// Test getting active servers
	activeServers := spawner.GetActiveServers()
	if activeServers == nil {
		t.Error("Expected GetActiveServers to return non-nil map")
	}

	if len(activeServers) != 0 {
		t.Error("Expected GetActiveServers to return empty map initially")
	}
}

func TestExecutableSpawner(t *testing.T) {
	spawner := NewExecutableSpawner()

	if spawner == nil {
		t.Fatal("NewExecutableSpawner() returned nil")
	}

	// Test with nil transport
	if spawner.IsRunning(nil) {
		t.Error("Expected IsRunning to return false for nil transport")
	}

	// Test getting active servers
	activeServers := spawner.GetActiveServers()
	if activeServers == nil {
		t.Error("Expected GetActiveServers to return non-nil map")
	}

	if len(activeServers) != 0 {
		t.Error("Expected GetActiveServers to return empty map initially")
	}
}