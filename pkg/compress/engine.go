package compress

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/corlin/AIMeter/pkg/domain"
)

var (
	// Regex for multi-newlines and spaces
	reMultiNewlines = regexp.MustCompile(`\n{3,}`)
	reMultiSpaces   = regexp.MustCompile(`[ \t]{2,}`)
	reCodeBlock     = regexp.MustCompile("(?s)```.*?```")

	// Low-entropy conversational fluff patterns (Chinese & English)
	lowEntropyPhrases = []string{
		"好的，我明白了。", "好的，我知道了。", "收到，没问题。", "非常感谢您的提问！",
		"如果您有任何其他问题，请随时告诉我。", "作为人工智能助手，", "希望以上内容对您有所帮助。",
		"Sure, I can help you with that.", "Sure thing!", "Got it!",
		"As an AI language model,", "Please let me know if you have any questions.",
		"I hope this helps!", "Thank you for asking!", "No problem at all.",
	}
)

// Engine performs high-throughput, low-latency semantic prompt compression
type Engine struct{}

// NewEngine creates a new prompt compression engine
func NewEngine() *Engine {
	return &Engine{}
}

// CompressMessages compresses a list of chat completion messages
func (e *Engine) CompressMessages(messages []domain.ChatMessage, policy domain.PromptCompressionPolicy) domain.PromptCompressionResult {
	start := time.Now()

	// Default threshold safety
	threshold := policy.MinTokenThreshold
	if threshold <= 0 {
		threshold = 300
	}
	preserveRecent := policy.PreserveRecentTurns
	if preserveRecent <= 0 {
		preserveRecent = 2
	}
	mode := strings.ToLower(policy.Mode)
	if mode == "" {
		mode = "balanced"
	}

	// Calculate initial tokens
	origTokens := 0
	for _, m := range messages {
		text := extractMessageContent(m.Content)
		origTokens += EstimateTokens(text)
	}

	// Short prompt fail-safe: skip if under minimum threshold
	if origTokens < threshold || !policy.Enabled {
		return domain.PromptCompressionResult{
			OriginalTokens:   origTokens,
			CompressedTokens: origTokens,
			SavedTokens:      0,
			CompressionRatio: 0.0,
			DurationMs:       float64(time.Since(start).Microseconds()) / 1000.0,
			Messages:         messages,
		}
	}

	// Step 1: Pre-process and consolidate consecutive system messages
	consolidated := consolidateSystemMessages(messages)

	totalMsgs := len(consolidated)
	compressedMsgs := make([]domain.ChatMessage, 0, totalMsgs)

	// Threshold index for recent turns: keep the last 2*preserveRecent messages verbatim
	recentCutoffIdx := totalMsgs - (preserveRecent * 2)
	if recentCutoffIdx < 0 {
		recentCutoffIdx = 0
	}

	for idx, msg := range consolidated {
		rawText := extractMessageContent(msg.Content)
		isRecent := idx >= recentCutoffIdx
		isSystem := strings.EqualFold(msg.Role, "system")

		var newText string

		if isSystem {
			// System prompt: apply Stage 1 structural dehydration only to preserve exact directives
			newText = e.dehydrateText(rawText, policy.PreserveCodeBlocks)
		} else if isRecent {
			// Recent focus interaction: apply Stage 1 structural dehydration only
			newText = e.dehydrateText(rawText, policy.PreserveCodeBlocks)
		} else {
			// Historical turns: apply Stage 1 and Stage 2 according to mode
			newText = e.dehydrateText(rawText, policy.PreserveCodeBlocks)
			if mode == "balanced" || mode == "aggressive" {
				newText = e.pruneFluff(newText, mode == "aggressive")
			}
		}

		// Avoid adding empty assistant responses
		if strings.TrimSpace(newText) == "" && strings.EqualFold(msg.Role, "assistant") {
			continue
		}

		newMsg := msg
		newMsg.Content = newText
		compressedMsgs = append(compressedMsgs, newMsg)
	}

	// Calculate compressed tokens
	compTokens := 0
	for _, m := range compressedMsgs {
		text := extractMessageContent(m.Content)
		compTokens += EstimateTokens(text)
	}

	savedTokens := origTokens - compTokens
	if savedTokens < 0 {
		savedTokens = 0
		compTokens = origTokens
		compressedMsgs = messages
	}

	ratio := 0.0
	if origTokens > 0 {
		ratio = float64(savedTokens) / float64(origTokens) * 100.0
	}

	return domain.PromptCompressionResult{
		OriginalTokens:   origTokens,
		CompressedTokens: compTokens,
		SavedTokens:      savedTokens,
		CompressionRatio: ratio,
		DurationMs:       float64(time.Since(start).Microseconds()) / 1000.0,
		Messages:         compressedMsgs,
	}
}

// dehydrateText performs Stage 1 zero-loss structural dehydration
func (e *Engine) dehydrateText(text string, preserveCodeBlocks bool) string {
	if text == "" {
		return ""
	}

	var codePlaceholders []string
	if preserveCodeBlocks {
		// Protect code blocks with placeholders
		text = reCodeBlock.ReplaceAllStringFunc(text, func(match string) string {
			idx := len(codePlaceholders)
			codePlaceholders = append(codePlaceholders, match)
			return fmt.Sprintf("__AIMETER_CODE_BLOCK_%d__", idx)
		})
	}

	// 1. Compress 3+ consecutive newlines to 2 newlines
	text = reMultiNewlines.ReplaceAllString(text, "\n\n")

	// 2. Compress horizontal spaces and tabs
	text = reMultiSpaces.ReplaceAllString(text, " ")

	// 3. Compress repetitive punctuation
	text = compressRepeatedPunctuation(text)

	// 4. Trim blank spaces around lines
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	text = strings.Join(lines, "\n")

	// Restore code blocks
	if preserveCodeBlocks && len(codePlaceholders) > 0 {
		for i, code := range codePlaceholders {
			text = strings.Replace(text, fmt.Sprintf("__AIMETER_CODE_BLOCK_%d__", i), code, 1)
		}
	}

	return strings.TrimSpace(text)
}

// pruneFluff performs Stage 2 conversational low-entropy fluff pruning
func (e *Engine) pruneFluff(text string, aggressive bool) string {
	for _, phrase := range lowEntropyPhrases {
		text = strings.ReplaceAll(text, phrase, "")
	}

	if aggressive {
		// In aggressive mode, strip standalone pleasantries
		lines := strings.Split(text, "\n")
		var cleanLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) < 15 {
				lower := strings.ToLower(trimmed)
				if lower == "ok" || lower == "okay" || lower == "好的" || lower == "收到" || lower == "明白" || lower == "yes" {
					continue
				}
			}
			cleanLines = append(cleanLines, line)
		}
		text = strings.Join(cleanLines, "\n")
	}

	return strings.TrimSpace(text)
}

func extractMessageContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		// Multimodal content
		var sb strings.Builder
		for _, part := range v {
			if m, ok := part.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					sb.WriteString(text)
					sb.WriteString("\n")
				}
			}
		}
		return sb.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func consolidateSystemMessages(messages []domain.ChatMessage) []domain.ChatMessage {
	if len(messages) <= 1 {
		return messages
	}

	var result []domain.ChatMessage
	var systemBuilder strings.Builder
	hasSystem := false

	for _, msg := range messages {
		if strings.EqualFold(msg.Role, "system") {
			hasSystem = true
			text := extractMessageContent(msg.Content)
			if systemBuilder.Len() > 0 {
				systemBuilder.WriteString("\n\n")
			}
			systemBuilder.WriteString(text)
		} else {
			if hasSystem {
				result = append(result, domain.ChatMessage{
					Role:    "system",
					Content: systemBuilder.String(),
				})
				hasSystem = false
				systemBuilder.Reset()
			}
			result = append(result, msg)
		}
	}

	if hasSystem && systemBuilder.Len() > 0 {
		result = append([]domain.ChatMessage{{
			Role:    "system",
			Content: systemBuilder.String(),
		}}, result...)
	}

	return result
}

func compressRepeatedPunctuation(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	var lastRune rune
	count := 0
	punctMap := map[rune]bool{
		'!': true, '！': true,
		'?': true, '？': true,
		'.': true, '。': true,
		',': true, '，': true,
	}

	for _, r := range s {
		if punctMap[r] {
			if r == lastRune {
				count++
				// Allow up to 3 dots for ellipsis (...), but only 1 for others
				maxAllowed := 1
				if r == '.' || r == '。' {
					maxAllowed = 3
				}
				if count <= maxAllowed {
					sb.WriteRune(r)
				}
			} else {
				lastRune = r
				count = 1
				sb.WriteRune(r)
			}
		} else {
			lastRune = r
			count = 0
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// EstimateTokens calculates approximate token counts for mixed CJK and Western text
func EstimateTokens(content string) int {
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

	westernTokens := (nonCjkChars + 3) / 4
	total := cjkCount + westernTokens

	if total == 0 && len(content) > 0 {
		return 1
	}
	return total
}
