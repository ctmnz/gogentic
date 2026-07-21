package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	log.Printf("[API] GET /hello?name=%s", name)
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/hello?name=" + url.QueryEscape(name))
	if err != nil {
		log.Printf("[API] GET /hello?name=%s failed: %v", name, err)
		return mcp.NewToolResultErrorFromErr("failed to call hello API", err), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API] Failed to read response: %v", err)
		return mcp.NewToolResultErrorFromErr("failed to read response", err), nil
	}

	log.Printf("[API] GET /hello?name=%s -> %s (%v)", name, string(body), time.Since(start))
	return mcp.NewToolResultText(string(body)), nil
}

func main() {
	s := server.NewMCPServer(
		"Hello API",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	tool := mcp.NewTool("hello",
		mcp.WithDescription("Say hello to someone via the hello API"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)

	s.AddTool(tool, helloHandler)

	httpServer := server.NewStreamableHTTPServer(s)
	fmt.Println("MCP server running on :3001/mcp")
	if err := httpServer.Start(":3001"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
