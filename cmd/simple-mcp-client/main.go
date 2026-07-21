package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	name := "World"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := client.NewStreamableHttpClient("http://localhost:3001/mcp")
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	if err := c.Start(ctx); err != nil {
		fmt.Printf("Failed to start client: %v\n", err)
		os.Exit(1)
	}

	serverInfo, err := c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "simple-mcp-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connected to %s v%s\n\n", serverInfo.ServerInfo.Name, serverInfo.ServerInfo.Version)

	toolsResult, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		fmt.Printf("Failed to list tools: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Available tools (%d):\n", len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	fmt.Printf("\nCalling hello tool with name=%q...\n\n", name)
	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "hello",
			Arguments: map[string]any{
				"name": name,
			},
		},
	})
	if err != nil {
		fmt.Printf("Failed to call tool: %v\n", err)
		os.Exit(1)
	}

	for _, content := range result.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			fmt.Printf("Response: %s\n", tc.Text)
		}
	}
}
