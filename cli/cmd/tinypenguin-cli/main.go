package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"example.com/tinypenguin/pkg/cli"
)

// getDefaultModel returns the default model from environment or fallback
func getDefaultModel() string {
	if model := os.Getenv("MODEL"); model != "" {
		return model
	}
	return "qwen2.5-coder:3b"
}

// getDefaultURL returns the default URL from environment or fallback
func getDefaultURL() string {
	if url := os.Getenv("TINYLLAMA_URL"); url != "" {
		return url
	}
	return "http://localhost:11434/v1"
}

var (
	tinyllamaURL *string
	model        *string
	taskID       *string
	toolsEnabled *bool
	debugMode    *bool
)

func init() {
	// Load .env file if it exists (ignore errors if file doesn't exist)
	_ = godotenv.Load()
	
	// Initialize flags with defaults from environment variables
	tinyllamaURL = flag.String("url", getDefaultURL(), "API URL (Ollama compatible)")
	model = flag.String("model", getDefaultModel(), "Model name to use")
	taskID = flag.String("task-id", "", "Task ID for cancel/list operations")
	toolsEnabled = flag.Bool("tools", true, "Enable tool calling (default: true)")
	debugMode = flag.Bool("debug", false, "Enable debug output to diagnose tool calling issues")
}

func main() {
	flag.Parse()
	
	if len(flag.Args()) == 0 {
		fmt.Println("tinypenguin-cli - A CLI tool for AI-powered system administration")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  tinypenguin-cli [flags] <command> [args...]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  run <query>    - Run a task with the given query")
		fmt.Println("  cancel <id>    - Cancel a task by ID")
		fmt.Println("  list           - List all tasks")
		fmt.Println("")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  tinypenguin-cli run \"Create a new user named john\"")
		fmt.Println("  tinypenguin-cli run \"Install nginx package\"")
		fmt.Println("  tinypenguin-cli run \"Create a bash script to backup files\"")
		fmt.Println("  tinypenguin-cli --tools=false run \"Just provide advice\"")
		fmt.Println("  tinypenguin-cli --debug run \"Check current users\"")
		return
	}
	
	command := flag.Arg(0)
	
	switch command {
	case "run":
		if len(flag.Args()) < 2 {
			log.Fatal("run command requires a query argument")
		}
		query := flag.Arg(1)
		if err := cli.RunTask(query, *tinyllamaURL, *model, *toolsEnabled, *debugMode); err != nil {
			log.Fatalf("Failed to run task: %v", err)
		}
		
	case "cancel":
		if *taskID == "" {
			log.Fatal("cancel command requires --task-id flag")
		}
		if err := cli.CancelTask(*taskID); err != nil {
			log.Fatalf("Failed to cancel task: %v", err)
		}
		
	case "list":
		if err := cli.ListTasks(); err != nil {
			log.Fatalf("Failed to list tasks: %v", err)
		}
		
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}