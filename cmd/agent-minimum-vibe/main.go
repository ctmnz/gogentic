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
	ollamaEndpoint := "http://192.168.86.247:11434/api/chat"
	modelName := "gemma4:12b-mlx"
	// modelName := "qwen3.5:9b-mlx"

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

		// Commands

		if strings.ToLower(userInput) == "exit" || strings.ToLower(userInput) == "quit" || strings.ToLower(userInput) == "/exit" || strings.ToLower(userInput) == "/quit" {
			fmt.Println("Hasta la vista!")
			break
		}

		// Clean up the memory
		if strings.ToLower(userInput) == "/clear" {
			messages = []Message{
				{
					Role:    "system",
					Content: "You are a helpful local AI agent. Keep your anwers concise.",
				},
			}
			continue
		}

		// That is working perfectly
		if strings.ToLower(userInput) == "/grilling" {
			messages = []Message{
				{
					Role:    "system",
					Content: "You are a helpful local AI agent. Keep your anwers concise.",
				},
				{
					Role:    "system",
					Content: "Interview me relentlessly about every aspect of this until we reach a shared understanding. Walk down each branch of the decision tree, resolving dependencies between decisions one-by-one. For each question, provide your recommended answer.Ask the questions one at a time, waiting for feedback on each question before continuing. Asking multiple questions at once is bewildering.If a *fact* can be found by exploring the environment (filesystem, tools, etc.), look it up rather than asking me. The *decisions*, though, are mine — put each one to me and wait for my answer.Do not act on it until I confirm we have reached a shared understanding.",
				},
			}
			continue
		}

		if strings.ToLower(userInput) == "/save" {
			continue
		}

		messages = append(messages, Message{Role: "user", Content: userInput})

		////// INNER LOOP
		for {
			reqData := ChatRequest{
				Model:    modelName,
				Messages: messages,
				Stream:   true,
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
			defer resp.Body.Close()

			decoder := json.NewDecoder(resp.Body)
			for {
				var chatResp ChatResponse
				if err := decoder.Decode(&chatResp); err == io.EOF {
					break
				} else if err != nil {
					log.Printf("Error decoding response: %v", err)
					break
				}

				if chatResp.Done {
					break
				}

				fmt.Print(chatResp.Message.Content)
				messages = append(messages, chatResp.Message)
			}
			fmt.Println()
			fmt.Printf("Took %v to answer!\n", time.Since(startTime))
			fmt.Printf("Context %v long!\n", len(messages))
			break
		}

	}
}
