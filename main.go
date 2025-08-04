package main

import (
	a "agent/agent"
	"agent/tools"
	"bufio"
	"context"
	"log"
	"os"

	"github.com/ollama/ollama/api"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	getUserMsg := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	tools := tools.GetAllTools()
	agent := a.NewAgent(client, getUserMsg, tools)
	err = agent.Run(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
