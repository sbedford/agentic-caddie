package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// ToolHandler is implemented by each tool.
type ToolHandler func(ctx context.Context, input json.RawMessage) (string, error)

type Agent struct {
	client       anthropic.Client
	systemPrompt string
	tools        []anthropic.ToolUnionParam
	handlers     map[string]ToolHandler
}

func NewAgent(client anthropic.Client, systemPrompt string) *Agent {
	return &Agent{
		client:       client,
		systemPrompt: systemPrompt,
		handlers:     map[string]ToolHandler{},
	}
}

func (a *Agent) RegisterTool(def anthropic.ToolUnionParam, name string, h ToolHandler) {
	a.tools = append(a.tools, def)
	a.handlers[name] = h
}

// Run executes the full agent loop for a single user request.
// contextBlock is your game-model text, prepended to the user message.
func (a *Agent) Run(ctx context.Context, contextBlock, userMessage string) (string, error) {

	fmt.Printf("Received request [%f] \n", userMessage)

	fmt.Printf("Context block [%f] \n", contextBlock)

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewTextBlock(contextBlock + "\n\n" + userMessage),
		),
	}

	for {
		resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			//Model:     anthropic.ModelClaudeSonnet4_6,
			Model:     anthropic.ModelClaudeHaiku4_5,
			MaxTokens: 1024,
			System: []anthropic.TextBlockParam{
				{Text: a.systemPrompt},
			},
			Messages: messages,
			Tools:    a.tools,
		})
		if err != nil {
			return "", fmt.Errorf("messages.new: %w", err)
		}

		// Always append the model's turn before deciding what to do with it.
		messages = append(messages, resp.ToParam())

		if resp.StopReason != anthropic.StopReasonToolUse {
			// Final answer — concatenate any text blocks.
			var out string
			for _, block := range resp.Content {
				if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
					out += tb.Text
				}
			}
			return out, nil
		}

		// StopReason == tool_use: dispatch every tool_use block, collect results.
		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			tu, ok := block.AsAny().(anthropic.ToolUseBlock)
			if !ok {
				continue
			}

			handler, found := a.handlers[tu.Name]

			var resultText string
			isError := false

			if !found {
				resultText = fmt.Sprintf("Received request for invalid tool [%s]\n", tu.Name)
				isError = true
			} else {

				fmt.Printf("Calling tool [%v] input [%s]\n", tu.Name, tu.Input)

				out, err := handler(ctx, tu.Input)
				if err != nil {
					resultText = err.Error()
					isError = true
				} else {
					resultText = out
				}
			}
			fmt.Printf("Got response [%v] from tool [%v]\n", resultText, tu.Name)
			toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.ID, resultText, isError))
		}

		// Tool results go back in as a user message; loop continues.
		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}
}
