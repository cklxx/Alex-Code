package mcp

import (
	"testing"
	"time"
)

func TestGetDefaultMCPConfig(t *testing.T) {
	config := GetDefaultMCPConfig()

	if config == nil {
		t.Fatal("GetDefaultMCPConfig() returned nil")
	}

	if !config.Enabled {
		t.Error("Expected default config to be enabled")
	}

	if config.GlobalTimeout <= 0 {
		t.Error("Expected positive global timeout")
	}

	if config.RefreshInterval <= 0 {
		t.Error("Expected positive refresh interval")
	}

	if config.Security == nil {
		t.Error("Expected security config to be non-nil")
	}

	if config.Logging == nil {
		t.Error("Expected logging config to be non-nil")
	}

	if config.Servers == nil {
		t.Error("Expected servers map to be non-nil")
	}
}

func TestMCPConfigAddServerConfig(t *testing.T) {
	config := GetDefaultMCPConfig()

	serverConfig := &ServerConfig{
		ID:      "test",
		Name:    "Test Server",
		Type:    SpawnerTypeNPX,
		Command: "@modelcontextprotocol/server-memory",
		Enabled: true,
		Timeout: 30 * time.Second,
	}

	err := config.AddServerConfig(serverConfig)
	if err != nil {
		t.Fatalf("Failed to add server config: %v", err)
	}

	retrieved, exists := config.GetServerConfig("test")
	if !exists {
		t.Error("Expected server config to exist after adding")
	}

	if retrieved.ID != serverConfig.ID {
		t.Errorf("Expected server ID %s, got %s", serverConfig.ID, retrieved.ID)
	}

	if retrieved.Name != serverConfig.Name {
		t.Errorf("Expected server name %s, got %s", serverConfig.Name, retrieved.Name)
	}
}

func TestMCPConfigRemoveServerConfig(t *testing.T) {
	config := GetDefaultMCPConfig()

	serverConfig := &ServerConfig{
		ID:      "test",
		Name:    "Test Server",
		Type:    SpawnerTypeNPX,
		Command: "@modelcontextprotocol/server-memory",
		Enabled: true,
		Timeout: 30 * time.Second,
	}

	// Add server config
	err := config.AddServerConfig(serverConfig)
	if err != nil {
		t.Fatalf("Failed to add server config: %v", err)
	}

	// Remove server config
	config.RemoveServerConfig("test")

	// Verify removal
	_, exists := config.GetServerConfig("test")
	if exists {
		t.Error("Expected server config to not exist after removal")
	}
}

func TestMCPConfigListServerConfigs(t *testing.T) {
	config := GetDefaultMCPConfig()

	// Initially empty
	configs := config.ListServerConfigs()
	if len(configs) != 0 {
		t.Error("Expected empty server configs list initially")
	}

	// Add server configs
	serverConfig1 := &ServerConfig{
		ID:      "test1",
		Name:    "Test Server 1",
		Type:    SpawnerTypeNPX,
		Command: "@modelcontextprotocol/server-memory",
		Enabled: true,
		Timeout: 30 * time.Second,
	}

	serverConfig2 := &ServerConfig{
		ID:      "test2",
		Name:    "Test Server 2",
		Type:    SpawnerTypeNPX,
		Command: "@modelcontextprotocol/server-filesystem",
		Enabled: false,
		Timeout: 30 * time.Second,
	}

	config.AddServerConfig(serverConfig1)
	config.AddServerConfig(serverConfig2)

	// Test ListServerConfigs
	configs = config.ListServerConfigs()
	if len(configs) != 2 {
		t.Errorf("Expected 2 server configs, got %d", len(configs))
	}

	// Test GetEnabledServers
	enabledConfigs := config.GetEnabledServers()
	if len(enabledConfigs) != 1 {
		t.Errorf("Expected 1 enabled server config, got %d", len(enabledConfigs))
	}

	if enabledConfigs[0].ID != "test1" {
		t.Errorf("Expected enabled server ID 'test1', got %s", enabledConfigs[0].ID)
	}
}

func TestMCPConfigValidateConfig(t *testing.T) {
	// Test valid config
	config := GetDefaultMCPConfig()
	err := config.ValidateConfig()
	if err != nil {
		t.Errorf("Expected valid config to pass validation: %v", err)
	}

	// Test invalid global timeout
	config.GlobalTimeout = 0
	err = config.ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail for zero global timeout")
	}

	// Reset and test invalid refresh interval
	config = GetDefaultMCPConfig()
	config.RefreshInterval = 0
	err = config.ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail for zero refresh interval")
	}

	// Reset and test invalid security config
	config = GetDefaultMCPConfig()
	config.Security.MaxProcesses = 0
	err = config.ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail for zero max processes")
	}

	// Reset and test invalid logging level
	config = GetDefaultMCPConfig()
	config.Logging.Level = "invalid"
	err = config.ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail for invalid log level")
	}
}

func TestMCPConfigJSONSerialization(t *testing.T) {
	config := GetDefaultMCPConfig()

	// Add a server config
	serverConfig := &ServerConfig{
		ID:      "test",
		Name:    "Test Server",
		Type:    SpawnerTypeNPX,
		Command: "@modelcontextprotocol/server-memory",
		Enabled: true,
		Timeout: 30 * time.Second,
	}
	config.AddServerConfig(serverConfig)

	// Test JSON serialization
	jsonData, err := config.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize config to JSON: %v", err)
	}

	// Test JSON deserialization
	var newConfig MCPConfig
	err = newConfig.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to deserialize config from JSON: %v", err)
	}

	// Verify deserialized config
	if newConfig.Enabled != config.Enabled {
		t.Error("Enabled flag not preserved in serialization")
	}

	if newConfig.GlobalTimeout != config.GlobalTimeout {
		t.Error("GlobalTimeout not preserved in serialization")
	}

	if len(newConfig.Servers) != len(config.Servers) {
		t.Error("Server configs not preserved in serialization")
	}

	retrievedServer, exists := newConfig.GetServerConfig("test")
	if !exists {
		t.Error("Server config not preserved in serialization")
	}

	if retrievedServer.ID != serverConfig.ID {
		t.Error("Server ID not preserved in serialization")
	}
}

func TestMCPConfigClone(t *testing.T) {
	config := GetDefaultMCPConfig()

	// Add a server config
	serverConfig := &ServerConfig{
		ID:      "test",
		Name:    "Test Server",
		Type:    SpawnerTypeNPX,
		Command: "@modelcontextprotocol/server-memory",
		Enabled: true,
		Timeout: 30 * time.Second,
	}
	config.AddServerConfig(serverConfig)

	// Test cloning
	clone := config.Clone()
	if clone == nil {
		t.Fatal("Clone() returned nil")
	}

	// Verify clone is independent
	if clone == config {
		t.Error("Clone returned same instance")
	}

	// Verify clone has same values
	if clone.Enabled != config.Enabled {
		t.Error("Enabled flag not preserved in clone")
	}

	if clone.GlobalTimeout != config.GlobalTimeout {
		t.Error("GlobalTimeout not preserved in clone")
	}

	if len(clone.Servers) != len(config.Servers) {
		t.Error("Server configs not preserved in clone")
	}

	// Modify original and verify clone is unaffected
	config.Enabled = false
	if clone.Enabled == false {
		t.Error("Clone was affected by modification to original")
	}
}

func TestAddCommonServerConfig(t *testing.T) {
	config := GetDefaultMCPConfig()

	// Test adding known common server
	err := config.AddCommonServerConfig("memory")
	if err != nil {
		t.Fatalf("Failed to add common server config: %v", err)
	}

	retrieved, exists := config.GetServerConfig("memory")
	if !exists {
		t.Error("Expected common server config to exist after adding")
	}

	if retrieved.Name != "Memory Server" {
		t.Errorf("Expected server name 'Memory Server', got %s", retrieved.Name)
	}

	// Test adding unknown common server
	err = config.AddCommonServerConfig("unknown")
	if err == nil {
		t.Error("Expected error for unknown common server config")
	}
}

func TestListCommonServerConfigs(t *testing.T) {
	commonConfigs := ListCommonServerConfigs()

	if len(commonConfigs) == 0 {
		t.Error("Expected non-empty list of common server configs")
	}

	// Check for expected common configs
	expectedConfigs := []string{"memory", "filesystem", "github"}
	for _, expected := range expectedConfigs {
		found := false
		for _, actual := range commonConfigs {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected common config %s not found", expected)
		}
	}
}