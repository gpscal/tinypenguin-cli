package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ToolCallLog represents the structure from tool_calls.log (both old and new format)
type ToolCallLog struct {
	Timestamp     string `json:"timestamp"`
	Model         string `json:"model"`
	UserQuery     string `json:"user_query,omitempty"`     // New field - may be empty in old logs
	ModelResponse string `json:"model_response,omitempty"` // New field - may be empty in old logs
	ToolName      string `json:"tool_name"`
	Arguments     string `json:"arguments"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	Output        string `json:"output,omitempty"`
	ErrorDetails  string `json:"error_details,omitempty"`
	ToolsEnabled  bool   `json:"tools_enabled"`
	Rating        int    `json:"rating,omitempty"`
}

// ModelResponse represents the parsed model response structure
type ModelResponse struct {
	Role     string     `json:"role"`
	Content  string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in the model response
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// FineTuningExample represents the format needed for Qwen fine-tuning
type FineTuningExample struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string      `json:"role"`
	Content string      `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run convert_logs_for_finetuning.go <tool_calls.log> [output.jsonl] [--min-rating N]")
		fmt.Println("Converts tool_calls.log entries to Qwen fine-tuning format")
		fmt.Println("Options:")
		fmt.Println("  --min-rating N    Only include examples with rating >= N (default: 3)")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := "finetuning_data.jsonl"
	minRating := 3

	// Parse arguments
	for i, arg := range os.Args {
		if arg == "--min-rating" && i+1 < len(os.Args) {
			fmt.Sscanf(os.Args[i+1], "%d", &minRating)
		}
		if i == 2 && !strings.HasPrefix(arg, "--") {
			outputFile = arg
		}
	}

	// Open input file
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Open output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	converted := 0
	skipped := 0
	oldFormat := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var logEntry ToolCallLog
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse line %d: %v\n", lineNum, err)
			skipped++
			continue
		}

		// Skip low-rated entries
		if logEntry.Rating > 0 && logEntry.Rating < minRating {
			skipped++
			continue
		}

		// Create fine-tuning example
		example, err := createFineTuningExample(logEntry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to create example from line %d: %v\n", lineNum, err)
			skipped++
			continue
		}

		if example == nil {
			// Old format without user_query - skip or reconstruct
			oldFormat++
			if logEntry.UserQuery == "" {
				// Try to reconstruct from tool call
				example = reconstructExample(logEntry)
				if example == nil {
					skipped++
					continue
				}
			} else {
				skipped++
				continue
			}
		}

		// Write as JSONL
		jsonData, err := json.Marshal(example)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to marshal example on line %d: %v\n", lineNum, err)
			skipped++
			continue
		}

		writer.WriteString(string(jsonData) + "\n")
		converted++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Conversion complete!\n")
	fmt.Printf("  ✅ Converted: %d examples\n", converted)
	fmt.Printf("  ⚠️  Skipped: %d entries\n", skipped)
	fmt.Printf("  📝 Old format (reconstructed): %d entries\n", oldFormat)
	fmt.Printf("  📄 Output file: %s\n", outputFile)
	fmt.Printf("  ⭐ Minimum rating filter: %d+\n", minRating)
}

func createFineTuningExample(logEntry ToolCallLog) (*FineTuningExample, error) {
	// Check if we have the new format with user_query and model_response
	if logEntry.UserQuery == "" || logEntry.ModelResponse == "" {
		return nil, fmt.Errorf("missing user_query or model_response")
	}

	// Parse the model response
	var modelResp ModelResponse
	if err := json.Unmarshal([]byte(logEntry.ModelResponse), &modelResp); err != nil {
		// Try to parse as a Message directly
		var msg Message
		if err2 := json.Unmarshal([]byte(logEntry.ModelResponse), &msg); err2 != nil {
			return nil, fmt.Errorf("failed to parse model_response: %v", err)
		}
		modelResp.Role = msg.Role
		modelResp.Content = msg.Content
		modelResp.ToolCalls = msg.ToolCalls
	}

	// Build the messages array
	messages := []Message{
		{
			Role:    "user",
			Content: logEntry.UserQuery,
		},
	}

	// Add assistant message with tool calls
	assistantMsg := Message{
		Role:    "assistant",
		Content: modelResp.Content,
	}

	// Convert tool calls to the format expected by Qwen
	if len(modelResp.ToolCalls) > 0 {
		assistantMsg.ToolCalls = modelResp.ToolCalls
	} else {
		// Reconstruct tool call from log entry
		toolCall := ToolCall{
			ID:   "call_1",
			Type: "function",
		}
		toolCall.Function.Name = logEntry.ToolName
		toolCall.Function.Arguments = logEntry.Arguments
		assistantMsg.ToolCalls = []ToolCall{toolCall}
	}

	messages = append(messages, assistantMsg)

	// Optionally add tool result as a follow-up message
	if logEntry.Status == "success" && logEntry.Output != "" {
		toolResultMsg := Message{
			Role:    "tool",
			Content: fmt.Sprintf("Tool execution result:\nStatus: %s\nOutput: %s", logEntry.Status, logEntry.Output),
		}
		messages = append(messages, toolResultMsg)
	}

	return &FineTuningExample{
		Messages: messages,
	}, nil
}

func reconstructExample(logEntry ToolCallLog) *FineTuningExample {
	// Reconstruct user query from tool call (best effort)
	userQuery := reconstructUserQuery(logEntry)

	// Create assistant response with tool call
	assistantResponse := createAssistantResponse(logEntry)

	// Build messages
	messages := []Message{
		{
			Role:    "user",
			Content: userQuery,
		},
		{
			Role:    "assistant",
			Content: assistantResponse,
		},
	}

	// Add tool call to assistant message
	toolCall := ToolCall{
		ID:   "call_1",
		Type: "function",
	}
	toolCall.Function.Name = logEntry.ToolName
	toolCall.Function.Arguments = logEntry.Arguments

	messages[1].ToolCalls = []ToolCall{toolCall}

	// Add tool result if available
	if logEntry.Status == "success" && logEntry.Output != "" {
		toolResultMsg := Message{
			Role:    "tool",
			Content: fmt.Sprintf("Tool execution result:\nStatus: %s\nOutput: %s", logEntry.Status, logEntry.Output),
		}
		messages = append(messages, toolResultMsg)
	}

	return &FineTuningExample{
		Messages: messages,
	}
}

func reconstructUserQuery(logEntry ToolCallLog) string {
	// Try to infer the user query from the tool call
	var args map[string]interface{}
	json.Unmarshal([]byte(logEntry.Arguments), &args)

	switch logEntry.ToolName {
	case "run_commands":
		if cmd, ok := args["command"].(string); ok {
			// Try to make it more natural
			if strings.HasPrefix(cmd, "who") || strings.HasPrefix(cmd, "w ") {
				return "Check current users"
			}
			if strings.HasPrefix(cmd, "pwd") {
				return "What's the current directory?"
			}
			if strings.HasPrefix(cmd, "ls") {
				return "List files in current directory"
			}
			if strings.HasPrefix(cmd, "ps") {
				return "Show running processes"
			}
			return fmt.Sprintf("Execute: %s", cmd)
		}
		return "Execute a command"
	case "edit_files":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Edit file: %s", path)
		}
		return "Edit a file"
	default:
		return fmt.Sprintf("Use tool: %s", logEntry.ToolName)
	}
}

func createAssistantResponse(logEntry ToolCallLog) string {
	// Create a natural assistant response
	response := fmt.Sprintf("I'll help you with that. Let me use the %s tool.", logEntry.ToolName)
	
	// Format tool call
	toolCallJSON := fmt.Sprintf(`{"id": "call_1", "type": "function", "function": {"name": "%s", "arguments": %s}}`, 
		logEntry.ToolName, logEntry.Arguments)
	
	response += fmt.Sprintf(`\n\n<tool_call>\n%s\n</tool_call>`, toolCallJSON)
	
	// Add result if available
	if logEntry.Status == "success" && logEntry.Output != "" {
		response += fmt.Sprintf(`\n\nTool execution completed successfully:\n%s`, logEntry.Output)
	} else if logEntry.Status == "error" {
		response += fmt.Sprintf(`\n\nTool execution failed: %s`, logEntry.Message)
	}
	
	return response
}
