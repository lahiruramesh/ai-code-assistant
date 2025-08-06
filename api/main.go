package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	// Check if the new CLI exists and run it
	if _, err := os.Stat("cmd/main.go"); err == nil {
		cmd := exec.Command("go", "run", "cmd/main.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to run CLI: %v", err)
		}
		return
	}

	log.Println("CLI not found. Please run: go run cmd/main.go")
}
