package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

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

type AgentResponse struct {
	Usage    TokenUsage
	Response string
	Err      error
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

func cleanJSONString(input string) string {
	// Trim surrounding whitespace and newlines
	cleaned := strings.TrimSpace(input)

	// Remove the leading markdown code block markers
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")

	// Remove the trailing markdown code block marker
	cleaned = strings.TrimSuffix(cleaned, "```")

	// Trim again in case there are remaining spaces/newlines inside the code blocks
	return strings.TrimSpace(cleaned)
}

// Run executes the full agent loop for a single user request.
// contextBlock is your game-model text, prepended to the user message.
func (a *Agent) Run(ctx context.Context, contextBlock, userMessage string) AgentResponse {

	agentTokens := TokenUsage{}

	log.Println("Received request: ", userMessage)
	log.Println("Context block: \n", contextBlock)

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
			return AgentResponse{
				Response: "",
				Err:      fmt.Errorf("messages.new: %w", err),
				Usage:    agentTokens,
			}
		}

		agentTokens.AddUsage(resp.Usage)

		messages = append(messages, resp.ToParam())

		if resp.StopReason != anthropic.StopReasonToolUse {
			// Final answer — concatenate any text blocks.
			var out string
			for _, block := range resp.Content {
				if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
					out += tb.Text
				}
			}

			return AgentResponse{
				Response: cleanJSONString("{" + out),
				Err:      nil,
				Usage:    agentTokens,
			}
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

				log.Println("Calling tool [", tu.Name, "]")

				out, err := handler(ctx, tu.Input)
				if err != nil {
					resultText = err.Error()
					isError = true
				} else {
					resultText = out
				}
			}
			log.Println("Got response [", resultText, "] from tool [", tu.Name, "]")
			toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.ID, resultText, isError))
		}

		// Tool results go back in as a user message; loop continues.
		messages = append(messages, anthropic.NewUserMessage(toolResults...))

		// Prefill the assistant response to enforce JSON-only output
		messages = append(messages, anthropic.MessageParam{
			Role: anthropic.MessageParamRole(anthropic.MessageParamRoleAssistant),
			Content: []anthropic.ContentBlockParamUnion{
				anthropic.NewTextBlock("{"),
			},
		})
	}
}
