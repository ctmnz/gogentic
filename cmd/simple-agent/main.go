package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ollamaEndpoint = "http://localhost:11434/api/chat"
	modelName      = "gemma4:12b-mlx"
	mcpEndpoint    = "http://localhost:3001/mcp"
)

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  any        `json:"parameters"`
}

type ToolCall struct {
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func connectMCP(ctx context.Context) (*client.Client, []Tool, error) {
	c, err := client.NewStreamableHttpClient(mcpEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	if err := c.Start(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to start MCP client: %w", err)
	}

	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "simple-agent",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	toolsResult, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools: %w", err)
	}

	var tools []Tool
	for _, t := range toolsResult.Tools {
		tools = append(tools, Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	return c, tools, nil
}

func callMCPTool(ctx context.Context, c *client.Client, name string, args map[string]any) (string, error) {
	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
	if err != nil {
		return "", fmt.Errorf("tool call failed: %w", err)
	}

	var texts []string
	for _, content := range result.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			texts = append(texts, tc.Text)
		}
	}
	return strings.Join(texts, "\n"), nil
}

func main() {
	initCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("Connecting to MCP server...")
	mcpClient, tools, err := connectMCP(initCtx)
	if err != nil {
		fmt.Printf("Failed to connect to MCP server: %v\n", err)
		os.Exit(1)
	}
	defer mcpClient.Close()

	fmt.Printf("Loaded %d tool(s) from MCP server:\n", len(tools))
	for _, t := range tools {
		fmt.Printf("  - %s: %s\n", t.Function.Name, t.Function.Description)
	}

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful local AI agent with access to external tools. Keep your answers concise. Use the available tools when the user asks you to greet someone or says hello.",
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\nReady! Type your message (or 'exit' to quit).")

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}
		if strings.ToLower(userInput) == "exit" || strings.ToLower(userInput) == "quit" {
			fmt.Println("Bye!")
			break
		}

		messages = append(messages, Message{Role: "user", Content: userInput})

		for {
			reqData := ChatRequest{
				Model:    modelName,
				Messages: messages,
				Tools:    tools,
				Stream:   false,
			}
			jsonData, err := json.Marshal(reqData)
			if err != nil {
				log.Printf("Error marshaling request: %v", err)
				break
			}

			startTime := time.Now()
			resp, err := http.Post(ollamaEndpoint, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error calling Ollama: %v", err)
				break
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("Error reading response: %v", err)
				break
			}

			var chatResp ChatResponse
			if err := json.Unmarshal(body, &chatResp); err != nil {
				log.Printf("Error parsing response: %v\nBody: %s", err, string(body))
				break
			}

			responseMsg := chatResp.Message
			messages = append(messages, responseMsg)

			if len(responseMsg.ToolCalls) == 0 {
				if responseMsg.Content != "" {
					fmt.Printf("Agent: %s\n", responseMsg.Content)
				}
				fmt.Printf("(took %v)\n", time.Since(startTime))
				break
			}

			fmt.Printf("[Calling tools...]\n")
			for _, tc := range responseMsg.ToolCalls {
				fmt.Printf("  -> %s(%v)\n", tc.Function.Name, tc.Function.Arguments)

				result, err := callMCPTool(context.Background(), mcpClient, tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					log.Printf("Error calling MCP tool: %v", err)
					result = fmt.Sprintf("Error: %v", err)
				}

				fmt.Printf("  <- %s\n", result)
				messages = append(messages, Message{
					Role:    "tool",
					Content: result,
				})
			}
		}
	}
}
