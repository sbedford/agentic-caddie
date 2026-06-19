package agent

import (
	"log"

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
	log.Println("Input Tokens: ", u.InputTokens)
	log.Println("Output Tokens: ", u.OutputTokens)
	log.Println("Cache Create Tokens: ", u.CacheCreationInputTokens)
	log.Println("Cache Read Tokens: ", u.CacheReadInputTokens)
}
