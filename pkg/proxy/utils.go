package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/gin-gonic/gin"
)

func (h *ProxyHandler) resolveTargetURL(provider, headerTarget string) string {
	if headerTarget != "" {
		return fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(headerTarget, "/"))
	}
	targetBase := "https://api.openai.com"
	if provider == "openai" && os.Getenv("AIMETER_UPSTREAM_OPENAI_URL") != "" {
		targetBase = os.Getenv("AIMETER_UPSTREAM_OPENAI_URL")
	} else if provider == "vllm" && os.Getenv("AIMETER_UPSTREAM_VLLM_URL") != "" {
		targetBase = os.Getenv("AIMETER_UPSTREAM_VLLM_URL")
	} else if provider == "ollama" && os.Getenv("AIMETER_UPSTREAM_OLLAMA_URL") != "" {
		targetBase = os.Getenv("AIMETER_UPSTREAM_OLLAMA_URL")
	} else if base, exists := h.defaultUpstreams[provider]; exists {
		targetBase = base
	}
	return fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(targetBase, "/"))
}

// NewProxyHandler creates a new ProxyHandler instance

func extractPromptText(payload map[string]interface{}) string {
	if msgs, ok := payload["messages"].([]domain.ChatMessage); ok && len(msgs) > 0 {
		var sb strings.Builder
		for _, m := range msgs {
			if str, ok := m.Content.(string); ok && str != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(str)
			}
		}
		if sb.Len() > 0 {
			return sb.String()
		}
	}

	rawMsgs, ok := payload["messages"].([]interface{})
	if !ok || len(rawMsgs) == 0 {
		if promptStr, ok := payload["prompt"].(string); ok {
			return promptStr
		}
		return ""
	}

	var sb strings.Builder
	for _, m := range rawMsgs {
		if msgMap, ok := m.(map[string]interface{}); ok {
			if content, ok := msgMap["content"].(string); ok {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(content)
			}
		} else if chatMsg, ok := m.(domain.ChatMessage); ok {
			if str, ok := chatMsg.Content.(string); ok {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(str)
			}
		}
	}
	return sb.String()
}

func renderJSONFromCache(c *gin.Context, entry *domain.CacheEntry, model string) {
	if len(entry.ResponseJSON) > 0 {
		c.Data(http.StatusOK, "application/json; charset=utf-8", entry.ResponseJSON)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      "chatcmpl-cached-" + entry.ID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []gin.H{
			{
				"index": 0,
				"message": gin.H{
					"role":    "assistant",
					"content": entry.ResponseText,
				},
				"finish_reason": "stop",
			},
		},
		"usage": gin.H{
			"prompt_tokens":     entry.InputTokens,
			"completion_tokens": entry.OutputTokens,
			"total_tokens":      entry.InputTokens + entry.OutputTokens,
		},
	})
}

func renderStreamFromCache(c *gin.Context, entry *domain.CacheEntry, model string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		renderJSONFromCache(c, entry, model)
		return
	}

	created := time.Now().Unix()
	chunkID := "chatcmpl-cached-" + entry.ID

	roleChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"\"},\"finish_reason\":null}]}\n\n",
		chunkID, created, model)
	_, _ = c.Writer.Write([]byte(roleChunk))
	flusher.Flush()

	escapedText, _ := json.Marshal(entry.ResponseText)
	contentChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{\"content\":%s},\"finish_reason\":null}]}\n\n",
		chunkID, created, model, string(escapedText))
	_, _ = c.Writer.Write([]byte(contentChunk))
	flusher.Flush()

	finishChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n",
		chunkID, created, model)
	_, _ = c.Writer.Write([]byte(finishChunk))
	flusher.Flush()

	usageChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[],\"usage\":{\"prompt_tokens\":%d,\"completion_tokens\":%d,\"total_tokens\":%d}}\n\n",
		chunkID, created, model, entry.InputTokens, entry.OutputTokens, entry.InputTokens+entry.OutputTokens)
	_, _ = c.Writer.Write([]byte(usageChunk))

	_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
}

// HandleChatCompletions handles POST /v1/chat/completions (OpenAI standard compatible)
