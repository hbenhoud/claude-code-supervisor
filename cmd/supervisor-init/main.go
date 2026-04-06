package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hbenhoud/claude-code-supervisor/internal/hooks"
)

func main() {
	uninstall := flag.Bool("uninstall", false, "Remove Supervisor hooks from Claude Code settings")
	apiPort := flag.Int("api-port", 3001, "Supervisor API port for hook URLs")
	flag.Parse()

	if *uninstall {
		if err := hooks.Uninstall(); err != nil {
			log.Fatalf("Failed to uninstall hooks: %v", err)
		}
		fmt.Println("Supervisor hooks removed from ~/.claude/settings.json")
		os.Exit(0)
	}

	if err := hooks.Install(*apiPort); err != nil {
		log.Fatalf("Failed to install hooks: %v", err)
	}
	fmt.Println("Supervisor hooks installed in ~/.claude/settings.json")
	fmt.Printf("Hooks will POST to http://localhost:%d/api/events\n", *apiPort)
}
