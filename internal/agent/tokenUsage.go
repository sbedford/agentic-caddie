package agent

import (
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

type TokenUsage struct {
	TotalInputTokens  int64
	TotalOutputTokens int64

	TotalCacheCreationInputTokens int64
	TotalCacheReadInputTokens     int64
}

func (t *TokenUsage) AddUsage(u anthropic.Usage) {
	t.TotalInputTokens += u.InputTokens
	t.TotalOutputTokens += u.OutputTokens
	t.TotalCacheCreationInputTokens += u.CacheCreationInputTokens
	t.TotalCacheReadInputTokens += u.CacheReadInputTokens
	t.logUsage(u)
}

func (t *TokenUsage) logUsage(u anthropic.Usage) {
	fmt.Printf("Input Tokens: %d\n", u.InputTokens)
	fmt.Printf("Output Tokens: %d\n", u.OutputTokens)
	fmt.Printf("Cache Create Tokens: %d\n", u.CacheCreationInputTokens)
	fmt.Printf("Cache Read Tokens: %d\n", u.CacheReadInputTokens)
}
