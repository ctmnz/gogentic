package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	OllamaURL    = "http://localhost:11434/api/generate"
	DefaultModel = "gemma4:12b-mlx" // Change to your preferred model
	SkillsDir    = "./skills"
)

// Global state for loaded skills
var activeSkills []string

func main() {
	fmt.Println("System initialized. Type 'learn [skillname]' to load a skill.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nUser > ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		// Master Skill Logic: Detection of "learn" command
		if strings.HasPrefix(strings.ToLower(input), "learn ") {
			skillName := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(input), "learn "))
			loadSkill(skillName)
			continue
		}

		// Default: Chat with injected skills
		response, err := chatWithOllama(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("\nAssistant > %s\n", response)
		}
	}
}

func loadSkill(name string) {
	path := filepath.Join(SkillsDir, name+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Skill '%s' not found in %s folder.\n", name, SkillsDir)
		return
	}

	content, _ := os.ReadFile(path)
	activeSkills = append(activeSkills, string(content))
	fmt.Printf("Successfully learned: %s\n", name)
}

func chatWithOllama(userInput string) (string, error) {
	// Construct System Prompt with injected skills
	systemPrompt := "You are a helpful assistant.\n"
	if len(activeSkills) > 0 {
		systemPrompt += "Active Skills:\n"
		for _, skill := range activeSkills {
			systemPrompt += "- " + skill + "\n"
		}
	}
	systemPrompt += "\nUser Input: " + userInput

	payload := map[string]interface{}{
		"model":  DefaultModel,
		"prompt": systemPrompt,
		"stream": false,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(OllamaURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return fmt.Sprintf("%v", result["response"]), nil
}
