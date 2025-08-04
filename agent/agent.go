package agent

import (
	"agent/tools"
	"context"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
)

type Agent struct {
	client     *api.Client
	getUserMsg func() (string, bool)
	tools      api.Tools
}

func NewAgent(
	client *api.Client,
	getUserMsg func() (string, bool),
	tools api.Tools,
) *Agent {
	return &Agent{
		client:     client,
		getUserMsg: getUserMsg,
		tools:      tools,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []api.Message{}

	fmt.Println("Chat with Ollma Cognito (use 'ctrl + c' to quit )")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMsg()
		if !ok {
			break
		}

		userMsg := api.Message{
			Role:    "user",
			Content: userInput,
		}

		conversation = append(conversation, userMsg)

		message, err := a.RunInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message)

		if len(message.ToolCalls) > 0 {
			fmt.Printf("\u001b[96m[Tool Calls]\u001b[0m: %d tool(s) to execute\n", len(message.ToolCalls))

			for _, toolCall := range message.ToolCalls {
				fmt.Printf("\u001b[96m[Executing]\u001b[0m: %s\n", toolCall.Function.Name)

				result, err := tools.ExecuteToolCall(toolCall)
				if err != nil {
					result = fmt.Sprintf("Error executing tool %s: %v", toolCall.Function.Name, err)
					fmt.Printf("\u001b[91m[Error]\u001b[0m: %s\n", result)
				} else {
					fmt.Printf("\u001b[92m[Success]\u001b[0m: Tool executed successfully\n")
				}

				toolResultMsg := api.Message{
					Role:      "tool",
					Content:   result,
					ToolCalls: []api.ToolCall{toolCall},
				}
				conversation = append(conversation, toolResultMsg)
			}

			followUpMessage, err := a.RunInference(ctx, conversation)
			if err != nil {
				return err
			}
			conversation = append(conversation, followUpMessage)

			fmt.Printf("\u001b[93mOllama\u001b[0m: %s\n", followUpMessage.Content)
		} else {
			fmt.Printf("\u001b[93mOllama\u001b[0m: %s\n", message.Content)
		}
	}

	return nil
}

func (a *Agent) RunInference(ctx context.Context, converstation []api.Message) (api.Message, error) {
	req := &api.ChatRequest{
		Model:    "cogito:14b",
		Messages: converstation,
		Tools:    a.tools,
	}

	var res api.Message
	respFunc := func(resp api.ChatResponse) error {
		if resp.Message.Content != "" {
			res.Role = resp.Message.Role
			res.Content += resp.Message.Content
			fmt.Print(resp.Message.Content) // Stream the content as it arrives
		}

		// Handle tool calls
		if len(resp.Message.ToolCalls) > 0 {
			res.Role = resp.Message.Role
			res.ToolCalls = append(res.ToolCalls, resp.Message.ToolCalls...)
		}

		return nil
	}

	err := a.client.Chat(ctx, req, respFunc)
	if err != nil {
		log.Fatal(err)
	}

	// Add a newline after streaming content if there was content
	if res.Content != "" {
		fmt.Println()
	}

	return res, nil
}
