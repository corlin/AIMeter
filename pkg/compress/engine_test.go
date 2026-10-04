package compress

import (
	"strings"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestCompressMessages_ShortPromptFailSafe(t *testing.T) {
	engine := NewEngine()
	msgs := []domain.ChatMessage{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello, how are you?"},
	}

	policy := domain.PromptCompressionPolicy{
		Enabled:           true,
		Mode:              "balanced",
		MinTokenThreshold: 300,
	}

	res := engine.CompressMessages(msgs, policy)
	assert.Equal(t, 0, res.SavedTokens)
	assert.Equal(t, res.OriginalTokens, res.CompressedTokens)
	assert.Equal(t, 2, len(res.Messages))
}

func TestCompressMessages_StructuralDehydrationAndCodePreservation(t *testing.T) {
	engine := NewEngine()

	codeSnippet := "```python\ndef hello():\n    # Preserve exact indentation\n    print('world')\n```"
	fluffyContent := "Here is the code you requested:\n\n\n\n" + codeSnippet + "\n\n\n\nIs there anything else I can help with?????"

	// Construct large message to exceed threshold
	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("This is a detailed background context document for the AI model to read.\n\n\n")
	}
	sb.WriteString(fluffyContent)

	msgs := []domain.ChatMessage{
		{Role: "system", Content: "You are an expert coder. Please write clean code."},
		{Role: "user", Content: sb.String()},
	}

	policy := domain.PromptCompressionPolicy{
		Enabled:            true,
		Mode:               "balanced",
		MinTokenThreshold:  50,
		PreserveCodeBlocks: true,
	}

	res := engine.CompressMessages(msgs, policy)
	assert.Greater(t, res.SavedTokens, 0)
	assert.Greater(t, res.CompressionRatio, 0.0)

	// Ensure code block is strictly preserved
	compressedContent := extractMessageContent(res.Messages[1].Content)
	assert.Contains(t, compressedContent, codeSnippet)
	assert.NotContains(t, compressedContent, "\n\n\n\n")
	assert.NotContains(t, compressedContent, "?????")
	assert.Contains(t, compressedContent, "?")
}

func TestCompressMessages_ContextAdaptivePruning(t *testing.T) {
	engine := NewEngine()

	// 6 turns of conversation (3 past, 3 recent)
	var longContext strings.Builder
	for i := 0; i < 15; i++ {
		longContext.WriteString("Very long background document with lots of tokens to exceed the min threshold. ")
	}

	msgs := []domain.ChatMessage{
		{Role: "system", Content: "You are an assistant."},
		{Role: "user", Content: "Turn 1: " + longContext.String() + "\n好的，我明白了。"},
		{Role: "assistant", Content: "Turn 1 Reply: " + longContext.String() + "\n如果您有任何其他问题，请随时告诉我。"},
		{Role: "user", Content: "Turn 2: " + longContext.String()},
		{Role: "assistant", Content: "Turn 2 Reply: " + longContext.String()},
		{Role: "user", Content: "Turn 3 Recent: What is the weather?"},
		{Role: "assistant", Content: "Turn 3 Recent Reply: It is sunny."},
	}

	policy := domain.PromptCompressionPolicy{
		Enabled:             true,
		Mode:                "balanced",
		MinTokenThreshold:   100,
		PreserveRecentTurns: 1, // Keep only Turn 3 completely verbatim
	}

	res := engine.CompressMessages(msgs, policy)
	assert.Greater(t, res.SavedTokens, 0)
	assert.Less(t, res.DurationMs, 50.0) // sub-millisecond in release, allow headroom under -race

	turn1Content := extractMessageContent(res.Messages[1].Content)
	assert.NotContains(t, turn1Content, "好的，我明白了。")

	turn1Assistant := extractMessageContent(res.Messages[2].Content)
	assert.NotContains(t, turn1Assistant, "如果您有任何其他问题，请随时告诉我。")
}
