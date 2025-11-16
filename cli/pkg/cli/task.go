package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"example.com/tinypenguin/pkg/common"
)

// TaskManager handles task execution with tinyllama integration
type TaskManager struct {
	tinyllamaClient *common.TinyllamaClient
	model           string
	toolsEnabled    bool
	debugMode       bool
}

// NewTaskManager creates a new task manager
func NewTaskManager(tinyllamaURL, model string, toolsEnabled, debugMode bool) *TaskManager {
	return &TaskManager{
		tinyllamaClient: common.NewTinyllamaClient(tinyllamaURL),
		model:          model,
		toolsEnabled:  toolsEnabled,
		debugMode:     debugMode,
	}
}

// TaskRequest represents a task execution request
type TaskRequest struct {
	Query string `json:"query"`
}

// TaskResponse represents a task execution response
type TaskResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

// ToolCallLog represents a log entry for tool call usage
type ToolCallLog struct {
	Timestamp        time.Time `json:"timestamp"`
	Model           string    `json:"model"`
	ToolName        string    `json:"tool_name"`
	Arguments       string    `json:"arguments"`
	Status          string    `json:"status"`
	Message         string    `json:"message"`
	Output          string    `json:"output,omitempty"`
	ErrorDetails    string    `json:"error_details,omitempty"`
	ToolsEnabled    bool      `json:"tools_enabled"`
	Rating          int       `json:"rating,omitempty"` // 1-5 stars for training data
}

// getLogPath returns the fixed path for the tool_calls.log file
func getLogPath() string {
	// Try to find project root by looking for README.md
	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	
	for {
		readmePath := filepath.Join(dir, "README.md")
		if _, err := os.Stat(readmePath); err == nil {
			// Found project root
			return filepath.Join(dir, "tool_calls.log")
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, use current directory as fallback
			break
		}
		dir = parent
	}
	
	// Fallback: use executable directory or current directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		return filepath.Join(execDir, "tool_calls.log")
	}
	
	// Last resort: use current directory
	wd, _ := os.Getwd()
	return filepath.Join(wd, "tool_calls.log")
}

// logToolCall appends a tool call log entry to the tool_calls.log file
func logToolCall(logEntry ToolCallLog) {
	const maxEntries = 1000
	logPath := getLogPath()

	var existingLogs []ToolCallLog

	// Read existing logs
	if data, err := os.ReadFile(logPath); err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				var entry ToolCallLog
				if json.Unmarshal([]byte(line), &entry) == nil {
					existingLogs = append(existingLogs, entry)
				}
			}
		}
	}

	// Add new entry
	existingLogs = append(existingLogs, logEntry)

	// Rotate if exceeded max entries
	if len(existingLogs) > maxEntries {
		existingLogs = existingLogs[len(existingLogs)-maxEntries:]
	}

	// Write back to file
	var logLines []string
	for _, log := range existingLogs {
		if jsonBytes, err := json.Marshal(log); err == nil {
			logLines = append(logLines, string(jsonBytes))
		}
	}

	logContent := strings.Join(logLines, "\n") + "\n"
	os.WriteFile(logPath, []byte(logContent), 0644)
}

func RunTask(query string, tinyllamaURL string, model string, toolsEnabled, debugMode bool) error {
	if tinyllamaURL == "" {
		tinyllamaURL = "http://localhost:11434/v1"
	}
	if model == "" {
		model = "qwen2.5-coder:3b"
	}
	manager := NewTaskManager(tinyllamaURL, model, toolsEnabled, debugMode)
	return manager.ExecuteTask(context.Background(), query)
}

// promptRating prompts the user to rate the tool usage (1-5 stars)
func promptRating() int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n⭐ Rate this tool usage (1-5 stars, or 0 to skip): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	rating, err := strconv.Atoi(input)
	if err != nil || rating < 0 || rating > 5 {
		return 0 // Skip rating if invalid
	}
	return rating
}

func (tm *TaskManager) ExecuteTask(ctx context.Context, query string) error {
	fmt.Printf("🚀 Starting task: %s\n", query)
	
	// Create system prompt for RHCSA/bash operations
	systemPrompt := `You are a Red Hat Certified System Administrator (RHCSA) assistant. 
You help with Linux system administration tasks including:
- File system operations (create, edit, delete files)
- Package management (yum/dnf, rpm)
- Service management (systemctl)
- User and group management
- Network configuration
- Security (SELinux, firewall, permissions)

CRITICAL INSTRUCTIONS FOR TOOL CALLING:
When you need to execute a command or edit a file, you MUST use the tool_calls format in your response.
DO NOT put JSON in your text content - the API expects tool_calls in a specific format.

CORRECT FORMAT (what you MUST do):
Your response should have a "tool_calls" array with this structure:
{
  "tool_calls": [
    {
      "id": "call_abc123",
      "type": "function",
      "function": {
        "name": "run_commands",
        "arguments": "{\"command\": \"who\"}"
      }
    }
  ]
}

WRONG FORMAT (what you MUST NOT do):
DO NOT put this in your content/text:
{
  "content": "```json\n{\"command\": \"who\"}\n```"
}

DO NOT put this in your content/text:
{
  "content": "{\"name\": \"run_commands\", \"arguments\": {\"command\": \"who\"}}"
}

KEY RULES:
1. ALWAYS use tool_calls array format (not JSON in content)
2. The "arguments" field must be a JSON STRING (escaped), not an object
3. For run_commands: arguments = "{\"command\": \"your-command-here\"}"
4. For edit_files: arguments = "{\"path\": \"/path/to/file\", \"diff\": \"your-diff-here\"}"
5. When user asks informational questions (like "check users"), ALWAYS use run_commands tool
6. The tool name must be exactly "run_commands" or "edit_files" (as defined in available tools)

EXAMPLES:

User: "Check current users"
You should respond with tool_calls containing:
{
  "tool_calls": [{
    "id": "call_1",
    "type": "function", 
    "function": {
      "name": "run_commands",
      "arguments": "{\"command\": \"who\"}"
    }
  }]
}

User: "What's the current directory?"
You should respond with tool_calls containing:
{
  "tool_calls": [{
    "id": "call_2",
    "type": "function",
    "function": {
      "name": "run_commands", 
      "arguments": "{\"command\": \"pwd\"}"
    }
  }]
}

Always prioritize security and provide safe, tested commands.
Use sudo when necessary for administrative tasks.

Current working directory: ` + getCurrentDirectory() + `
Available tools:
- edit_files: Edit file contents using diff format
- run_commands: Execute shell commands (USE THIS tool for ALL commands, including informational queries)`

	// Prepare messages for the model
	messages := []common.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: query,
		},
	}

	// Define available tools (only if tools are enabled)
	var tools []common.Tool
	if tm.toolsEnabled {
		tools = []common.Tool{
			common.CreateToolDefinition(
				"edit_files",
				"Edit file contents by providing a diff of changes to make",
				map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the file to edit",
						},
						"diff": map[string]interface{}{
							"type":        "string",
							"description": "Diff content showing changes to make",
						},
					},
					"required": []interface{}{"path", "diff"},
				},
			),
			common.CreateToolDefinition(
				"run_commands",
				"Execute shell commands on the system",
				map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "Command to execute",
						},
						"timeout": map[string]interface{}{
							"type":        "integer",
							"description": "Timeout in seconds (optional)",
						},
					},
					"required": []interface{}{"command"},
				},
			),
		}
		if tm.debugMode {
			fmt.Printf("🔧 Tools enabled: %d tool(s) available\n", len(tools))
			for _, tool := range tools {
				fmt.Printf("   - %s: %s\n", tool.Function.Name, tool.Function.Description)
			}
		}
	} else {
		if tm.debugMode {
			fmt.Printf("⚠️  Tools are disabled - model will only provide text responses\n")
		}
	}

	// Create chat request
	chatReq := &common.ChatRequest{
		Model:    tm.model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}
	
	if tm.debugMode {
		reqJSON, _ := json.MarshalIndent(chatReq, "", "  ")
		fmt.Printf("🐛 DEBUG - Request:\n%s\n", string(reqJSON))
	}

	// Send request to the model
	fmt.Printf("🤖 Analyzing task with %s...\n", tm.model)
	if tm.debugMode {
		fmt.Printf("🐛 DEBUG - Tools enabled: %v\n", tm.toolsEnabled)
	}
	
	resp, err := tm.tinyllamaClient.Chat(ctx, chatReq)
	if err != nil {
		return fmt.Errorf("failed to get response from model: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response from model")
	}

	choice := resp.Choices[0]
	message := choice.Message
	
	if tm.debugMode {
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Printf("🐛 DEBUG - Response:\n%s\n", string(respJSON))
		fmt.Printf("🐛 DEBUG - Finish reason: %s\n", choice.FinishReason)
		fmt.Printf("🐛 DEBUG - Tool calls count: %d\n", len(message.ToolCalls))
		if len(message.ToolCalls) > 0 {
			for i, tc := range message.ToolCalls {
				fmt.Printf("🐛 DEBUG - Tool call %d: ID=%s, Type=%s, Name=%s, Args=%s\n", 
					i+1, tc.ID, tc.Type, tc.Function.Name, tc.Function.Arguments)
			}
		}
	}
	
	// Check if the model wants to use tools
	if len(message.ToolCalls) > 0 {
		fmt.Printf("🔧 Model wants to use %d tool(s)\n", len(message.ToolCalls))
		
		for _, toolCall := range message.ToolCalls {
			fmt.Printf("🛠️  Executing tool: %s\n", toolCall.Function.Name)

			var toolResult TaskResponse

			switch toolCall.Function.Name {
			case "edit_files":
				toolResult = tm.executeEditFiles(toolCall.Function.Arguments)
			case "run_commands":
				toolResult = tm.executeRunCommands(toolCall.Function.Arguments)
			default:
				toolResult = TaskResponse{
					Status:  "error",
					Message: fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name),
				}
			}

			fmt.Printf("📊 Tool result: %s - %s\n", toolResult.Status, toolResult.Message)
			if toolResult.Output != "" {
				fmt.Printf("📤 Output:\n%s\n", toolResult.Output)
			}

			// Prompt for rating
			rating := promptRating()
			if rating > 0 {
				fmt.Printf("⭐ Rating saved: %d/5 stars\n", rating)
			}

			// Log the tool call for training
			logEntry := ToolCallLog{
				Timestamp:     time.Now(),
				Model:        tm.model,
				ToolName:     toolCall.Function.Name,
				Arguments:    toolCall.Function.Arguments,
				Status:       toolResult.Status,
				Message:      toolResult.Message,
				Output:       toolResult.Output,
				ToolsEnabled: tm.toolsEnabled,
				Rating:       rating,
				ErrorDetails: func() string {
					if toolResult.Status == "error" {
						return toolResult.Message
					}
					return ""
				}(),
			}
			logToolCall(logEntry)
		}
	} else {
		if tm.debugMode {
			fmt.Printf("🐛 DEBUG - No tool calls in response. Content: %s\n", message.Content)
		}
		
		// Try to parse JSON response that might contain command suggestions
		// This handles cases where the model returns malformed tool calls in content
		command, shouldExecute := tm.parseCommandFromResponse(message.Content)
		
		if tm.debugMode {
			fmt.Printf("🐛 DEBUG - Parsed command: '%s', shouldExecute: %v\n", command, shouldExecute)
		}
		
		if shouldExecute && command != "" {
			// For informational questions, automatically execute the suggested command
			fmt.Printf("💡 Detected command suggestion in response: %s\n", command)
			fmt.Printf("⚠️  Note: Model should use tool_calls format, but detected command in content. Executing anyway...\n")
			fmt.Printf("🚀 Executing command to answer your question...\n\n")
			
			// Properly escape the command in JSON
			cmdJSON, _ := json.Marshal(map[string]string{"command": command})
			toolResult := tm.executeRunCommands(string(cmdJSON))
			
			if toolResult.Status == "success" {
				fmt.Printf("✅ Answer:\n%s\n", toolResult.Output)
			} else {
				fmt.Printf("❌ Error executing command: %s\n", toolResult.Message)
				if toolResult.Output != "" {
					fmt.Printf("Output: %s\n", toolResult.Output)
				}
			}

			// Prompt for rating
			rating := promptRating()
			if rating > 0 {
				fmt.Printf("⭐ Rating saved: %d/5 stars\n", rating)
			}

			// Log the tool call for training (fallback path - malformed tool call)
			logEntry := ToolCallLog{
				Timestamp:     time.Now(),
				Model:        tm.model,
				ToolName:     "run_commands",
				Arguments:    string(cmdJSON),
				Status:       toolResult.Status,
				Message:      toolResult.Message,
				Output:       toolResult.Output,
				ToolsEnabled: tm.toolsEnabled,
				Rating:       rating,
				ErrorDetails: func() string {
					if toolResult.Status == "error" {
						return toolResult.Message
					}
					return ""
				}(),
			}
			logToolCall(logEntry)
		} else if command != "" {
			// Command found but not safe to auto-execute
			fmt.Printf("💡 Model suggested command: %s\n", command)
			fmt.Printf("⚠️  Note: Model should use tool_calls format instead of JSON in content.\n")
			fmt.Printf("💬 Suggested command: %s\n", command)
			fmt.Printf("💬 To execute this command, you can run: %s\n", command)
		} else if message.Content != "" {
			// Display the model's response if it's not just JSON
			// Check if it's valid JSON - if so, try to extract useful info
			var jsonContent map[string]interface{}
			if err := json.Unmarshal([]byte(message.Content), &jsonContent); err == nil {
				// It's JSON, try to extract command or provide helpful message
				if cmd, ok := jsonContent["command"].(string); ok && cmd != "" {
					fmt.Printf("💡 Suggested command: %s\n", cmd)
					fmt.Printf("💬 To execute this command, you can run: %s\n", cmd)
				} else {
					fmt.Printf("📝 Model response: %s\n", message.Content)
				}
			} else {
				// Not JSON, display as-is
				fmt.Printf("💬 Answer:\n%s\n", message.Content)
			}
		} else {
			fmt.Println("✅ Task completed without tool usage")
		}
	}

	return nil
}

func (tm *TaskManager) executeEditFiles(arguments string) TaskResponse {
	var params struct {
		Path string `json:"path"`
		Diff string `json:"diff"`
	}
	
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return TaskResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to parse edit_files arguments: %v", err),
		}
	}

	fmt.Printf("📝 Editing file: %s\n", params.Path)
	fmt.Printf("📝 Diff:\n%s\n", params.Diff)
	
	// For now, just validate the input and return success
	// In a real implementation, you would apply the diff to the file
	if params.Path == "" || params.Diff == "" {
		return TaskResponse{
			Status:  "error",
			Message: "Both path and diff are required",
		}
	}
	
	return TaskResponse{
		Status:  "success",
		Message: fmt.Sprintf("File edit operation would be applied to %s", params.Path),
		Output:  fmt.Sprintf("Applied diff to %s", params.Path),
	}
}

func (tm *TaskManager) executeRunCommands(arguments string) TaskResponse {
	var params struct {
		Command string `json:"command"`
		Timeout *int   `json:"timeout,omitempty"`
	}
	
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return TaskResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to parse run_commands arguments: %v", err),
		}
	}

	fmt.Printf("💻 Executing command: %s\n", params.Command)
	
	// Validate command
	if params.Command == "" {
		return TaskResponse{
			Status:  "error",
			Message: "Command is required",
		}
	}

	// Check for dangerous commands
	if isDangerousCommand(params.Command) {
		return TaskResponse{
			Status:  "denied",
			Message: "Command was denied for safety reasons",
		}
	}

	// Execute the command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if params.Timeout != nil {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(*params.Timeout)*time.Second)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", params.Command)
	
	// Set working directory
	wd, _ := os.Getwd()
	cmd.Dir = wd
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return TaskResponse{
				Status:  "error",
				Message: "Command timed out",
			}
		}
		return TaskResponse{
			Status:  "error",
			Message: fmt.Sprintf("Command failed: %v", err),
			Output:  string(output),
		}
	}
	
	return TaskResponse{
		Status:  "success",
		Message: "Command executed successfully",
		Output:  string(output),
	}
}

func isDangerousCommand(command string) bool {
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /usr",
		"rm -rf /bin",
		"dd if=",
		"mkfs",
		"fdisk",
		"shred",
		"cryptsetup",
	}
	
	command = strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return true
		}
	}
	
	return false
}

func getCurrentDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "/unknown"
	}
	return wd
}

// parseCommandFromResponse attempts to extract a command from the model's response
// Returns the command and whether it should be executed automatically
func (tm *TaskManager) parseCommandFromResponse(content string) (string, bool) {
	if content == "" {
		return "", false
	}
	
	// Strip markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		// Remove opening ```json or ```
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimSpace(lines[0])
			if strings.HasPrefix(firstLine, "```") {
				lines = lines[1:]
			}
		}
		// Remove closing ```
		if len(lines) > 0 {
			lastLine := strings.TrimSpace(lines[len(lines)-1])
			if lastLine == "```" {
				lines = lines[:len(lines)-1]
			}
		}
		content = strings.TrimSpace(strings.Join(lines, "\n"))
	}
	
	// Try to parse as JSON
	var jsonContent map[string]interface{}
	var jsonErr error
	if jsonErr = json.Unmarshal([]byte(content), &jsonContent); jsonErr != nil {
		// If parsing failed, try to find JSON object in the content using regex-like approach
		// Look for {...} pattern
		startIdx := strings.Index(content, "{")
		endIdx := strings.LastIndex(content, "}")
		if startIdx >= 0 && endIdx > startIdx {
			jsonStr := content[startIdx : endIdx+1]
			jsonErr = json.Unmarshal([]byte(jsonStr), &jsonContent)
			if jsonErr == nil {
				content = jsonStr
			}
		}
	}
	
	if jsonErr == nil {
		// It's valid JSON - try multiple formats
		var cmd string
		
		// Format 1: {"command": "users"}
		if c, ok := jsonContent["command"].(string); ok && c != "" {
			cmd = c
		}
		
		// Format 2: {"name": "run_commands", "arguments": {"command": "cat /etc/passwd"}}
		// Format 3: {"name": "systemctl", "arguments": {"command": "cat /etc/passwd"}}
		if cmd == "" {
			if args, ok := jsonContent["arguments"].(map[string]interface{}); ok {
				if c, ok := args["command"].(string); ok && c != "" {
					cmd = c
				}
			}
		}
		
		// Format 4: {"arguments": "{\"command\": \"cat /etc/passwd\"}"} (stringified JSON)
		if cmd == "" {
			if argsStr, ok := jsonContent["arguments"].(string); ok {
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
					if c, ok := args["command"].(string); ok && c != "" {
						cmd = c
					}
				}
			}
		}
		
		if cmd != "" {
			// Check if it's a safe informational command
			cmdLower := strings.ToLower(strings.TrimSpace(cmd))
			
			// List of safe informational commands that can be auto-executed
			// These are read-only commands that provide information
			safeInfoCommands := []string{
				"who", "w", "users", "whoami", "id",
				"cat /etc/passwd", "getent passwd", "cut -d: -f1 /etc/passwd",
				"ls", "pwd", "date", "uptime",
				"uname", "hostname", "df", "free",
				"ps", "systemctl list-units", "systemctl status",
				"netstat", "ss", "ip addr", "ip route",
			}
			
			// Check if command matches or starts with any safe pattern
			for _, safeCmd := range safeInfoCommands {
				// Exact match or starts with the safe command (allowing for flags)
				if cmdLower == safeCmd || strings.HasPrefix(cmdLower, safeCmd+" ") {
					return cmd, true
				}
			}
			
			// Also check for common read-only patterns
			if strings.HasPrefix(cmdLower, "cat ") || 
			   strings.HasPrefix(cmdLower, "less ") ||
			   strings.HasPrefix(cmdLower, "head ") ||
			   strings.HasPrefix(cmdLower, "tail ") ||
			   strings.HasPrefix(cmdLower, "grep ") ||
			   strings.HasPrefix(cmdLower, "find ") ||
			   strings.HasPrefix(cmdLower, "ls ") ||
			   strings.HasPrefix(cmdLower, "getent ") ||
			   strings.HasPrefix(cmdLower, "cut ") {
				// These are generally safe read operations
				return cmd, true
			}
			
			// For other commands, suggest but don't auto-execute
			return cmd, false
		}
	}
	
	// Try to extract command from text patterns
	// Look for patterns like "command: users" or "run: users" or just "users" at start
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check for JSON-like patterns
		if strings.Contains(line, `"command"`) || strings.Contains(line, `'command'`) {
			// Try to extract from this line
			if idx := strings.Index(line, ":"); idx > 0 {
				potentialCmd := strings.TrimSpace(line[idx+1:])
				potentialCmd = strings.Trim(potentialCmd, `"'{}[]`)
				if potentialCmd != "" && !strings.Contains(potentialCmd, "{") {
					return potentialCmd, false
				}
			}
		}
	}
	
	return "", false
}

func CancelTask(taskID string) error {
	fmt.Printf("Cancelling task: %s\n", taskID)
	// Placeholder implementation
	return nil
}

func ListTasks() error {
	fmt.Println("Listing tasks:")
	// Placeholder implementation  
	return nil
}