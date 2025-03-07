package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/glamour"
	"gopkg.in/yaml.v3"
)

const (
	defaultWidth = 120
)

// Config structure to hold API configuration
var config struct {
	API struct {
		URL string `yaml:"url"`
		Key string `yaml:"key"`
	}
}

func init() {
	// Load the configuration from the YAML file
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		os.Exit(1)
	}
	// Try config in faraday subfolder first, then executable directory
	configFilePath := filepath.Join(filepath.Dir(exePath), "faraday", "config.yaml")
	configFile, err := os.Open(configFilePath)
	if err != nil {
		// Fallback to executable directory
		configFilePath = filepath.Join(filepath.Dir(exePath), "config.yaml")
		configFile, err = os.Open(configFilePath)
		if err != nil {
			fmt.Printf("Error opening config file: %v\n", err)
			os.Exit(1)
		}
	}
	defer configFile.Close()

	yamlDecoder := yaml.NewDecoder(configFile)
	err = yamlDecoder.Decode(&config)
	if err != nil {
		fmt.Printf("Error decoding config file: %v\n", err)
		os.Exit(1)
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Messages []Message `json:"messages"`
}

func main() {
	// Define command line flags
	flag.Parse()

	// Join all command-line arguments as the prompt
	args := flag.Args()
	prompt := strings.Join(args, " ")

	if prompt == "" {
		fmt.Println("Usage: Please provide a prompt as a command-line argument.")
		fmt.Println("Example: $ faraday Your prompt here")
		return
	}

	var contextFilePath string

	// Count the occurrences of '@'
	atCount := strings.Count(prompt, "@")
	if atCount > 1 {
		fmt.Println("Error: Multiple '@' symbols detected. Please provide only one context file.")
		return
	}

	// Check for @file syntax in the prompt
	if atCount == 1 {
		parts := strings.SplitN(prompt, "@", 2)
		prompt = strings.TrimSpace(parts[0])
		contextFilePath = strings.TrimSpace(parts[1])
	}

	s := spinner.New(spinner.CharSets[37], 60*time.Millisecond)
	s.Suffix = " Thinking ... "
	s.Start()
	time.Sleep(1 * time.Second)

	// Call AI service
	response, err := callAIService(prompt, contextFilePath)
	s.Stop()

	if err != nil {
		fmt.Printf("Error calling AI service: %v\n", err)
		return
	}

	// Create a new renderer.
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(defaultWidth),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating renderer: %s\n", err)
	}

	// Render markdown.
	md, err := r.Render(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering markdown: %s\n", err)
	}

	// fmt.Print(md)

	chunks := strings.Split(string(md), "\n")

	// Write markdown to stdout with delay
	for _, chunk := range chunks {
		fmt.Fprintf(os.Stdout, "%s\n", chunk)
		time.Sleep(120 * time.Millisecond)
	}
}

func callAIService(prompt string, contextFilePath string) (string, error) {
	// Initialize the request body with the user prompt
	reqBody := RequestBody{
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	// Check if context file path is set and valid
	if contextFilePath != "" {
		info, err := os.Stat(contextFilePath)
		if err != nil || info.IsDir() {
			return "", fmt.Errorf("context file path '%s' is not a file or does not exist", contextFilePath)
		}

		// Read the context file
		contextContent, err := os.ReadFile(contextFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read context file: %w", err)
		}

		// Add context content to request body
		reqBody.Messages = append(reqBody.Messages, Message{Role: "system", Content: string(contextContent)})
	}

	// Marshal the request body to JSON
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequest("POST", config.API.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.API.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response (%w)", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse the JSON response
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON response (%w)", err)
	}

	// Extract the full response
	var fullResponse string = ""
	if message, ok := jsonResponse["choices"]; ok {
		if len(message.([]interface{})) > 0 {
			if content, ok := message.([]interface{})[0].(map[string]interface{}); ok {
				if text, ok := content["message"].(map[string]interface{}); ok {
					if textContent, ok := text["content"].(string); ok {
						fullResponse = textContent
					}
				}
			}
		}
	}

	return fullResponse, nil
}
