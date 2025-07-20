package main

import (
	"context"
	"fmt"
	"time"

	"alex/internal/config"
	"alex/internal/tools/mcp"
	"alex/internal/tools/mcp/protocol"
	"alex/internal/tools/builtin"
)

// convertToMCPConfig converts config.MCPConfig to mcp.MCPConfig
func convertToMCPConfig(configMCP *config.MCPConfig) *mcp.MCPConfig {
	mcpConfig := &mcp.MCPConfig{
		Enabled:         configMCP.Enabled,
		Servers:         make(map[string]*mcp.ServerConfig),
		GlobalTimeout:   configMCP.GlobalTimeout,
		AutoRefresh:     configMCP.AutoRefresh,
		RefreshInterval: configMCP.RefreshInterval,
	}

	// Convert security config
	if configMCP.Security != nil {
		mcpConfig.Security = &mcp.SecurityConfig{
			AllowedCommands:     configMCP.Security.AllowedCommands,
			BlockedCommands:     configMCP.Security.BlockedCommands,
			AllowedPackages:     configMCP.Security.AllowedPackages,
			BlockedPackages:     configMCP.Security.BlockedPackages,
			RequireConfirmation: configMCP.Security.RequireConfirmation,
			SandboxMode:         configMCP.Security.SandboxMode,
			MaxProcesses:        configMCP.Security.MaxProcesses,
			MaxMemoryMB:         configMCP.Security.MaxMemoryMB,
			AllowedEnvironment:  configMCP.Security.AllowedEnvironment,
			RestrictedPaths:     configMCP.Security.RestrictedPaths,
		}
	}

	// Convert logging config
	if configMCP.Logging != nil {
		mcpConfig.Logging = &mcp.LoggingConfig{
			Level:        configMCP.Logging.Level,
			LogRequests:  configMCP.Logging.LogRequests,
			LogResponses: configMCP.Logging.LogResponses,
			LogFile:      configMCP.Logging.LogFile,
		}
	}

	// Convert server configs
	for id, server := range configMCP.Servers {
		mcpConfig.Servers[id] = &mcp.ServerConfig{
			ID:          server.ID,
			Name:        server.Name,
			Type:        mcp.SpawnerType(server.Type),
			Command:     server.Command,
			Args:        server.Args,
			Env:         server.Env,
			WorkDir:     server.WorkDir,
			AutoStart:   server.AutoStart,
			AutoRestart: server.AutoRestart,
			Timeout:     server.Timeout,
			Enabled:     server.Enabled,
		}
	}

	return mcpConfig
}

func main() {
	fmt.Println("MCP Demo for Alex")
	fmt.Println("==================")

	// Create configuration manager
	configManager, err := config.NewManager()
	if err != nil {
		fmt.Printf("Failed to create config manager: %v\n", err)
		return
	}

	// Get MCP configuration from config manager
	configMCP := configManager.GetMCPConfig()
	fmt.Printf("MCP Enabled: %v\n", configMCP.Enabled)
	fmt.Printf("Global Timeout: %v\n", configMCP.GlobalTimeout)

	// Convert config.MCPConfig to mcp.MCPConfig
	mcpConfig := convertToMCPConfig(configMCP)

	// Add a memory server for demo (disabled to avoid requiring npx)
	memoryServer := &mcp.ServerConfig{
		ID:          "memory-demo",
		Name:        "Memory Server Demo",
		Type:        mcp.SpawnerTypeNPX,
		Command:     "@modelcontextprotocol/server-memory",
		Args:        []string{},
		Env:         make(map[string]string),
		AutoStart:   false, // Disabled for demo
		AutoRestart: true,
		Timeout:     30 * time.Second,
		Enabled:     false, // Disabled for demo
	}

	err = mcpConfig.AddServerConfig(memoryServer)
	if err != nil {
		fmt.Printf("Failed to add memory server: %v\n", err)
		return
	}

	// Create MCP manager
	mcpManager := mcp.NewManager(mcpConfig)

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start MCP manager (won't start disabled servers)
	fmt.Println("\nStarting MCP Manager...")
	if err := mcpManager.Start(ctx); err != nil {
		fmt.Printf("Failed to start MCP manager: %v\n", err)
		// Continue demo even if servers fail to start
	} else {
		fmt.Println("MCP Manager started successfully")
	}

	// Demonstrate configuration
	fmt.Println("\nMCP Configuration:")
	fmt.Printf("- Enabled: %v\n", mcpConfig.Enabled)
	fmt.Printf("- Auto Refresh: %v\n", mcpConfig.AutoRefresh)
	fmt.Printf("- Security Sandbox: %v\n", mcpConfig.Security.SandboxMode)
	fmt.Printf("- Max Processes: %d\n", mcpConfig.Security.MaxProcesses)
	fmt.Printf("- Logging Level: %s\n", mcpConfig.Logging.Level)

	// List configured servers
	fmt.Println("\nConfigured MCP Servers:")
	for _, server := range mcpConfig.ListServerConfigs() {
		fmt.Printf("- %s (%s): %s\n", server.Name, server.ID, server.Type)
		fmt.Printf("  Command: %s\n", server.Command)
		fmt.Printf("  Enabled: %v\n", server.Enabled)
		fmt.Printf("  Auto Start: %v\n", server.AutoStart)
	}

	// List common server configurations
	fmt.Println("\nAvailable Common Server Configs:")
	for _, name := range mcp.ListCommonServerConfigs() {
		fmt.Printf("- %s\n", name)
	}

	// Demonstrate tool integration
	fmt.Println("\nTool Integration:")
	builtinTools := builtin.GetAllBuiltinTools()
	fmt.Printf("Built-in tools: %d\n", len(builtinTools))

	// Integrate MCP tools (would normally include tools from running servers)
	allTools := mcpManager.IntegrateWithBuiltinTools(builtinTools)
	fmt.Printf("Total tools after MCP integration: %d\n", len(allTools))

	// List first few tools
	fmt.Println("\nFirst 5 tools:")
	for i, tool := range allTools {
		if i >= 5 {
			break
		}
		fmt.Printf("- %s: %s\n", tool.Name(), tool.Description())
	}

	// Demonstrate JSON-RPC protocol
	fmt.Println("\nJSON-RPC Protocol Demo:")
	request := protocol.NewRequest(1, "tools/list", nil)
	fmt.Printf("Sample request: %+v\n", request)

	response := protocol.NewResponse(1, map[string]interface{}{
		"tools": []interface{}{
			map[string]interface{}{
				"name":        "demo_tool",
				"description": "A demo tool",
			},
		},
	})
	fmt.Printf("Sample response: %+v\n", response)

	// Demonstrate spawner capabilities
	fmt.Println("\nSpawner Demo:")
	command, args := mcp.GetNPXPackageCommand("memory")
	fmt.Printf("Memory server command: %s %v\n", command, args)

	command, args = mcp.GetNPXPackageCommand("filesystem")
	fmt.Printf("Filesystem server command: %s %v\n", command, args)

	// Stop MCP manager
	fmt.Println("\nStopping MCP Manager...")
	_ = mcpManager.Stop()

	fmt.Println("\nMCP Demo completed successfully!")
}