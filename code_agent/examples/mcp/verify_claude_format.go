package main

import (
	"encoding/json"
	"fmt"
	"os"

	"code_agent/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify_claude_format.go <config-file>")
		os.Exit(1)
	}

	configPath := os.Args[1]

	fmt.Printf("Loading MCP configuration from: %s\n\n", configPath)

	cfg, err := config.LoadMCP(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Configuration loaded successfully!\n\n")
	fmt.Printf("Enabled: %v\n", cfg.Enabled)
	fmt.Printf("Number of servers: %d\n\n", len(cfg.Servers))

	for name, srv := range cfg.Servers {
		fmt.Printf("Server: %s\n", name)
		fmt.Printf("  Type: %s\n", srv.Type)
		fmt.Printf("  Command: %s\n", srv.Command)
		if len(srv.Args) > 0 {
			fmt.Printf("  Args: %v\n", srv.Args)
		}
		if len(srv.Env) > 0 {
			fmt.Printf("  Env vars:\n")
			for k, v := range srv.Env {
				fmt.Printf("    %s=%s\n", k, v)
			}
		}
		if srv.URL != "" {
			fmt.Printf("  URL: %s\n", srv.URL)
		}
		fmt.Println()
	}

	fmt.Println("Internal representation as JSON:")
	jsonBytes, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(jsonBytes))
}
