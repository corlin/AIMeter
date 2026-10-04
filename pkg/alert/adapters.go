package alert

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DetectChannelType detects the channel platform by examining the Webhook URL
func DetectChannelType(webhookURL string) string {
	lower := strings.ToLower(webhookURL)
	switch {
	case strings.Contains(lower, "feishu.cn") || strings.Contains(lower, "larksuite.com"):
		return ChannelFeishu
	case strings.Contains(lower, "dingtalk.com"):
		return ChannelDingTalk
	case strings.Contains(lower, "qyapi.weixin.qq.com") || strings.Contains(lower, "work.weixin.qq.com"):
		return ChannelWeCom
	case strings.Contains(lower, "slack.com"):
		return ChannelSlack
	default:
		return ChannelGenericJSON
	}
}

// FormatFeishuCard builds an interactive Feishu/Lark card payload
func FormatFeishuCard(event NotificationEvent, suppressedCount int) map[string]interface{} {
	headerColor := "blue"
	switch strings.ToLower(event.Severity) {
	case "critical":
		headerColor = "red"
	case "warning":
		headerColor = "orange"
	}

	elements := []map[string]interface{}{
		{
			"tag": "div",
			"text": map[string]interface{}{
				"tag": "lark_md",
				"content": fmt.Sprintf(
					"**级别**: %s | **租户**: `%s` | **工作流**: `%s`\n**时间**: %s",
					strings.ToUpper(event.Severity),
					event.TenantID,
					event.WorkflowID,
					event.TriggeredAt.Format("2006-01-02 15:04:05 UTC"),
				),
			},
		},
		{
			"tag": "hr",
		},
		{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": event.Message,
			},
		},
	}

	if len(event.Metrics) > 0 {
		var metricLines []string
		for k, v := range event.Metrics {
			metricLines = append(metricLines, fmt.Sprintf("• **%s**: %v", k, v))
		}
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": strings.Join(metricLines, "\n"),
			},
		})
	}

	if suppressedCount > 0 {
		elements = append(elements, map[string]interface{}{
			"tag": "note",
			"elements": []map[string]interface{}{
				{
					"tag":     "plain_text",
					"content": fmt.Sprintf("AI Meter 智能防刷保护：在过去静默周期内已自动聚合抑制 %d 次相同告警", suppressedCount),
				},
			},
		})
	}

	return map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"template": headerColor,
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": fmt.Sprintf("[AI Meter] %s", event.Title),
				},
			},
			"elements": elements,
		},
	}
}

// FormatDingTalkMarkdown builds a DingTalk bot Markdown payload
func FormatDingTalkMarkdown(event NotificationEvent, suppressedCount int) map[string]interface{} {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### [AI Meter 告警] %s\n\n", event.Title))
	sb.WriteString(fmt.Sprintf("- **级别**: %s\n", strings.ToUpper(event.Severity)))
	sb.WriteString(fmt.Sprintf("- **租户**: `%s`\n", event.TenantID))
	sb.WriteString(fmt.Sprintf("- **工作流**: `%s`\n", event.WorkflowID))
	sb.WriteString(fmt.Sprintf("- **时间**: %s\n\n", event.TriggeredAt.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString(fmt.Sprintf("> %s\n\n", event.Message))

	for k, v := range event.Metrics {
		sb.WriteString(fmt.Sprintf("- **%s**: %v\n", k, v))
	}

	if suppressedCount > 0 {
		sb.WriteString(fmt.Sprintf("\n*(智能防刷：已自动抑制 %d 次重复告警)*\n", suppressedCount))
	}

	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": fmt.Sprintf("[AI Meter] %s", event.Title),
			"text":  sb.String(),
		},
	}
}

// FormatWeComMarkdown builds a WeChat Work Markdown bot payload
func FormatWeComMarkdown(event NotificationEvent, suppressedCount int) map[string]interface{} {
	var sb strings.Builder
	colorTag := "info"
	if strings.ToLower(event.Severity) == "critical" {
		colorTag = "warning"
	}

	sb.WriteString(fmt.Sprintf("### <font color=\"%s\">[AI Meter 告警]</font> %s\n", colorTag, event.Title))
	sb.WriteString(fmt.Sprintf("> **级别**: <font color=\"comment\">%s</font>\n", strings.ToUpper(event.Severity)))
	sb.WriteString(fmt.Sprintf("> **租户**: `%s`\n", event.TenantID))
	sb.WriteString(fmt.Sprintf("> **工作流**: `%s`\n", event.WorkflowID))
	sb.WriteString(fmt.Sprintf("> **时间**: %s\n\n", event.TriggeredAt.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString(fmt.Sprintf("%s\n\n", event.Message))

	for k, v := range event.Metrics {
		sb.WriteString(fmt.Sprintf("• **%s**: %v\n", k, v))
	}

	if suppressedCount > 0 {
		sb.WriteString(fmt.Sprintf("\n> *(智能防刷：已自动聚合抑制 %d 次重复告警)*\n", suppressedCount))
	}

	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": sb.String(),
		},
	}
}

// FormatSlackPayload builds a Slack Block Kit payload
func FormatSlackPayload(event NotificationEvent, suppressedCount int) map[string]interface{} {
	color := "#3b82f6" // blue
	switch strings.ToLower(event.Severity) {
	case "critical":
		color = "#ef4444" // red
	case "warning":
		color = "#f59e0b" // orange
	}

	fields := []map[string]interface{}{
		{"title": "Tenant", "value": event.TenantID, "short": true},
		{"title": "Workflow", "value": event.WorkflowID, "short": true},
		{"title": "Severity", "value": strings.ToUpper(event.Severity), "short": true},
		{"title": "Time", "value": event.TriggeredAt.Format(time.RFC3339), "short": true},
	}

	for k, v := range event.Metrics {
		fields = append(fields, map[string]interface{}{
			"title": k,
			"value": fmt.Sprintf("%v", v),
			"short": true,
		})
	}

	footer := "AI Meter Control Plane"
	if suppressedCount > 0 {
		footer = fmt.Sprintf("AI Meter • Suppressed %d duplicate alerts", suppressedCount)
	}

	return map[string]interface{}{
		"text": fmt.Sprintf("[AI Meter] %s", event.Title),
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  event.Title,
				"text":   event.Message,
				"fields": fields,
				"footer": footer,
				"ts":     event.TriggeredAt.Unix(),
			},
		},
	}
}

// FormatGenericJSON serializes event and creates HMAC signature if secret provided
func FormatGenericJSON(event NotificationEvent, suppressedCount int, secret string) ([]byte, map[string]string) {
	payload := map[string]interface{}{
		"source":           "aimeter",
		"event":            event,
		"suppressed_count": suppressedCount,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	data, _ := json.Marshal(payload)
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "AIMeter-Webhook/1.0",
	}

	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(data)
		sig := hex.EncodeToString(mac.Sum(nil))
		headers["X-AIMeter-Signature"] = "sha256=" + sig
	}

	return data, headers
}
