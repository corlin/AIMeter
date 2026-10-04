package alert

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Supported Channel Types
const (
	ChannelFeishu      = "feishu"
	ChannelDingTalk    = "dingtalk"
	ChannelWeCom       = "wecom"
	ChannelSlack       = "slack"
	ChannelGenericJSON = "generic_json"
)

// Standard Event Types
const (
	EventBudgetWarning         = "BUDGET_WARNING"
	EventBudgetExceeded        = "BUDGET_EXCEEDED"
	EventCircuitBreakerTripped = "CIRCUIT_BREAKER_TRIPPED"
	EventRunawayLoopPrevented  = "RUNAWAY_LOOP_PREVENTED"
	EventSpikeAnomaly          = "SPIKE_ANOMALY"
	EventTest                  = "TEST_ALERT"
)

// AlertChannel defines a notification destination
type AlertChannel struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Name             string    `json:"name"`
	ChannelType      string    `json:"channel_type"` // feishu, dingtalk, wecom, slack, generic_json
	WebhookURL       string    `json:"webhook_url"`
	Secret           string    `json:"secret,omitempty"` // HMAC signing secret or token
	SubscribedEvents []string  `json:"subscribed_events"`
	CooldownSeconds  int       `json:"cooldown_seconds"` // Default 300s
	Enabled          bool      `json:"enabled"`
	CreatedAt        time.Time `json:"created_at"`
}

// NotificationEvent represents an alert to be published to channels
type NotificationEvent struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	WorkflowID  string                 `json:"workflow_id,omitempty"`
	EventType   string                 `json:"event_type"`
	Severity    string                 `json:"severity"` // "critical", "warning", "info"
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
	TriggeredAt time.Time              `json:"triggered_at"`
}

// Fingerprint generates a deduplication key for alert cooldown
func (e *NotificationEvent) Fingerprint() string {
	raw := fmt.Sprintf("%s:%s:%s", e.TenantID, e.EventType, e.WorkflowID)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:8])
}

// DeliveryLog captures the outcome of an alert dispatch
type DeliveryLog struct {
	ID           string    `json:"id"`
	ChannelID    string    `json:"channel_id"`
	ChannelName  string    `json:"channel_name"`
	ChannelType  string    `json:"channel_type"`
	EventID      string    `json:"event_id"`
	EventType    string    `json:"event_type"`
	Success      bool      `json:"success"`
	HTTPStatus   int       `json:"http_status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	LatencyMs    int       `json:"latency_ms"`
	DeliveredAt  time.Time `json:"delivered_at"`
}
