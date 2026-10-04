package proxy

import (
	"encoding/json"
	"strings"
	"unicode"
)

// EstimateDeltaTokens performs high-speed, zero-lock heuristic token counting
// on incoming SSE content chunks for both Asian (CJK) and Western languages.
func EstimateDeltaTokens(content string) int {
	if len(content) == 0 {
		return 0
	}

	var cjkCount int
	var nonCjkChars int

	for _, r := range content {
		if unicode.Is(unicode.Han, r) ||
			unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Hangul) {
			cjkCount++
		} else {
			nonCjkChars++
		}
	}

	// CJK: ~1.0 token per character in modern BPE (DeepSeek/Qwen/Llama3 tokenizers)
	// Western: ~3.8 characters per token
	westernTokens := (nonCjkChars + 3) / 4
	total := cjkCount + westernTokens

	if total == 0 && len(content) > 0 {
		return 1
	}
	return total
}

// SSEChunkData represents minimal chunk structure to extract delta content
type SSEChunkData struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *OpenAIUsage `json:"usage"`
}

// ExtractDeltaFromSSELine extracts incremental content from an SSE payload line
func ExtractDeltaFromSSELine(jsonBytes []byte) (string, *OpenAIUsage, bool) {
	if len(jsonBytes) == 0 || string(jsonBytes) == "[DONE]" {
		return "", nil, false
	}

	var chunk SSEChunkData
	if err := json.Unmarshal(jsonBytes, &chunk); err != nil {
		return "", nil, false
	}

	var content string
	if len(chunk.Choices) > 0 {
		content = chunk.Choices[0].Delta.Content
	}

	return content, chunk.Usage, true
}

// BuildTerminationSSEChunks generates graceful closing SSE messages
func BuildTerminationSSEChunks(customNotice string) [][]byte {
	notice := customNotice
	if strings.TrimSpace(notice) == "" {
		notice = "\n\n[AI Meter: Generation capped: single-request token budget exceeded]"
	}

	// 1. Notice chunk with the informative message
	noticeChunk := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"delta": map[string]string{
					"content": notice,
				},
				"index": 0,
			},
		},
	}
	noticeBytes, _ := json.Marshal(noticeChunk)

	// 2. Final finish chunk with finish_reason: "budget_exceeded"
	finishReason := "budget_exceeded"
	finishChunk := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"delta":         map[string]interface{}{},
				"finish_reason": finishReason,
				"index":         0,
			},
		},
	}
	finishBytes, _ := json.Marshal(finishChunk)

	return [][]byte{
		[]byte("data: " + string(noticeBytes) + "\n\n"),
		[]byte("data: " + string(finishBytes) + "\n\n"),
		[]byte("data: [DONE]\n\n"),
	}
}
