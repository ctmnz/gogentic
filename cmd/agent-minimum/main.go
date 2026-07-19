package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Main structs

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// Tools structs

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolCall struct {
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Main function

func main() {

	ollamaEndpoint := "http://localhost:11434/api/chat"
	modelName := "gemma4:12b"

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful local AI agent. Keep your anwers concise.",
		},
	}

	fmt.Printf("Thee program is running! Started at  %v\n", time.Now())

	scanner := bufio.NewScanner(os.Stdin)

	// Main loop

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
			fmt.Println("Hasta la vista!")
			break
		}

		messages = append(messages, Message{Role: "user", Content: userInput})

		////// INNER LOOP
		for {
			reqData := ChatRequest{
				Model:    modelName,
				Messages: messages,
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
				log.Printf("Error calling Ollama API: %v", err)
				break
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				break
			}
			var chatResp ChatResponse
			if err := json.Unmarshal(body, &chatResp); err != nil {
				log.Printf("Error unmarhaling response: %v\n", err)
				break
			}
			responseMsg := chatResp.Message
			messages = append(messages, responseMsg)

			fmt.Printf("Agent: %s\n", responseMsg.Content)
			fmt.Printf("Took %v to answer!\n", time.Since(startTime))
			break
		}

	}

}
