package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type cooldownEntry struct {
	lastSentAt      time.Time
	suppressedCount int
}

// AlertDispatcher manages channel routing, cooldowns, and reliable deliveries
type AlertDispatcher struct {
	client       *http.Client
	channels     map[string]*AlertChannel
	cooldowns    map[string]*cooldownEntry
	deliveryLogs []DeliveryLog
	mu           sync.RWMutex
}

// NewAlertDispatcher creates an AlertDispatcher instance
func NewAlertDispatcher(client *http.Client) *AlertDispatcher {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &AlertDispatcher{
		client:       client,
		channels:     make(map[string]*AlertChannel),
		cooldowns:    make(map[string]*cooldownEntry),
		deliveryLogs: make([]DeliveryLog, 0, 100),
	}
}

// RegisterChannel creates or updates an alert destination
func (d *AlertDispatcher) RegisterChannel(ch AlertChannel) AlertChannel {
	d.mu.Lock()
	defer d.mu.Unlock()

	if ch.ID == "" {
		ch.ID = uuid.New().String()
	}
	if ch.ChannelType == "" {
		ch.ChannelType = DetectChannelType(ch.WebhookURL)
	}
	if ch.CooldownSeconds <= 0 {
		ch.CooldownSeconds = 300 // default 5 minutes
	}
	if len(ch.SubscribedEvents) == 0 {
		ch.SubscribedEvents = []string{
			EventBudgetWarning,
			EventBudgetExceeded,
			EventCircuitBreakerTripped,
			EventRunawayLoopPrevented,
			EventSpikeAnomaly,
		}
	}
	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = time.Now().UTC()
	}

	copied := ch
	d.channels[ch.ID] = &copied
	return copied
}

// DeleteChannel removes an alert channel
func (d *AlertDispatcher) DeleteChannel(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.channels[id]; exists {
		delete(d.channels, id)
		return true
	}
	return false
}

// GetChannels returns all channels, optionally filtered by tenant
func (d *AlertDispatcher) GetChannels(tenantID string) []AlertChannel {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []AlertChannel
	for _, ch := range d.channels {
		if tenantID == "" || tenantID == "all" || ch.TenantID == "" || ch.TenantID == "all" || ch.TenantID == tenantID {
			result = append(result, *ch)
		}
	}
	return result
}

// GetDeliveryLogs returns recent delivery audits
func (d *AlertDispatcher) GetDeliveryLogs(limit int) []DeliveryLog {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if limit <= 0 || limit > len(d.deliveryLogs) {
		limit = len(d.deliveryLogs)
	}

	result := make([]DeliveryLog, limit)
	startIdx := len(d.deliveryLogs) - limit
	copy(result, d.deliveryLogs[startIdx:])

	// Reverse so newest is first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Dispatch publishes an event to all subscribed channels with deduplication
func (d *AlertDispatcher) Dispatch(ctx context.Context, event NotificationEvent) {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.TriggeredAt.IsZero() {
		event.TriggeredAt = time.Now().UTC()
	}

	// 1. Check Fingerprint & Cooldown suppression
	fp := event.Fingerprint()
	suppressedCount := 0

	d.mu.Lock()
	entry, exists := d.cooldowns[fp]
	if exists {
		// Use default 300s cooldown for fingerprint check
		cooldownDuration := 300 * time.Second
		if time.Since(entry.lastSentAt) < cooldownDuration {
			// Suppress alert
			entry.suppressedCount++
			d.mu.Unlock()
			return
		}
		suppressedCount = entry.suppressedCount
		entry.lastSentAt = time.Now()
		entry.suppressedCount = 0
	} else {
		d.cooldowns[fp] = &cooldownEntry{
			lastSentAt:      time.Now(),
			suppressedCount: 0,
		}
	}

	// 2. Identify active target channels
	var targets []*AlertChannel
	for _, ch := range d.channels {
		if !ch.Enabled {
			continue
		}
		if ch.TenantID != "" && ch.TenantID != "all" && event.TenantID != "" && ch.TenantID != event.TenantID {
			continue
		}
		if isSubscribed(ch.SubscribedEvents, event.EventType) {
			targets = append(targets, ch)
		}
	}
	d.mu.Unlock()

	// 3. Deliver concurrently in background
	for _, target := range targets {
		go d.deliverWithRetry(context.Background(), target, event, suppressedCount)
	}
}

// TestChannel dispatches an immediate test card and returns delivery details
func (d *AlertDispatcher) TestChannel(ctx context.Context, ch AlertChannel) DeliveryLog {
	testEvent := NotificationEvent{
		ID:         uuid.New().String(),
		TenantID:   ch.TenantID,
		WorkflowID: "test-connectivity-flow",
		EventType:  EventTest,
		Severity:   "info",
		Title:      "通道连通性测试 (Connectivity Test)",
		Message:    fmt.Sprintf("您好！这是来自 AI Meter 的通道测试告警。当前通道 [%s]（%s）已成功建立连接，可正常接收生产事件。", ch.Name, ch.ChannelType),
		Metrics: map[string]interface{}{
			"channel_name": ch.Name,
			"channel_type": ch.ChannelType,
			"status":       "CONNECTED",
		},
		TriggeredAt: time.Now().UTC(),
	}

	return d.sendHTTP(ctx, &ch, testEvent, 0)
}

func (d *AlertDispatcher) deliverWithRetry(ctx context.Context, ch *AlertChannel, event NotificationEvent, suppressedCount int) {
	maxAttempts := 3
	var lastLog DeliveryLog

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		lastLog = d.sendHTTP(reqCtx, ch, event, suppressedCount)
		cancel()

		if lastLog.Success {
			return
		}

		if attempt < maxAttempts {
			// Exponential backoff: 500ms, 1000ms
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
		}
	}
}

func (d *AlertDispatcher) sendHTTP(ctx context.Context, ch *AlertChannel, event NotificationEvent, suppressedCount int) DeliveryLog {
	start := time.Now()
	channelType := ch.ChannelType
	if channelType == "" {
		channelType = DetectChannelType(ch.WebhookURL)
	}

	var bodyBytes []byte
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	switch channelType {
	case ChannelFeishu:
		payload := FormatFeishuCard(event, suppressedCount)
		bodyBytes, _ = json.Marshal(payload)
	case ChannelDingTalk:
		payload := FormatDingTalkMarkdown(event, suppressedCount)
		bodyBytes, _ = json.Marshal(payload)
	case ChannelWeCom:
		payload := FormatWeComMarkdown(event, suppressedCount)
		bodyBytes, _ = json.Marshal(payload)
	case ChannelSlack:
		payload := FormatSlackPayload(event, suppressedCount)
		bodyBytes, _ = json.Marshal(payload)
	default:
		bodyBytes, headers = FormatGenericJSON(event, suppressedCount, ch.Secret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ch.WebhookURL, bytes.NewReader(bodyBytes))
	logEntry := DeliveryLog{
		ID:          uuid.New().String(),
		ChannelID:   ch.ID,
		ChannelName: ch.Name,
		ChannelType: channelType,
		EventID:     event.ID,
		EventType:   event.EventType,
		DeliveredAt: time.Now().UTC(),
	}

	if err != nil {
		logEntry.Success = false
		logEntry.ErrorMessage = fmt.Sprintf("Failed to construct HTTP request: %v", err)
		logEntry.LatencyMs = int(time.Since(start).Milliseconds())
		d.recordLog(logEntry)
		return logEntry
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	logEntry.LatencyMs = int(time.Since(start).Milliseconds())

	if err != nil {
		logEntry.Success = false
		logEntry.ErrorMessage = err.Error()
		d.recordLog(logEntry)
		return logEntry
	}
	defer resp.Body.Close()

	logEntry.HTTPStatus = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logEntry.Success = true
	} else {
		logEntry.Success = false
		bodySample, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		logEntry.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodySample))
	}

	d.recordLog(logEntry)
	return logEntry
}

func (d *AlertDispatcher) recordLog(log DeliveryLog) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.deliveryLogs) >= 100 {
		d.deliveryLogs = d.deliveryLogs[1:]
	}
	d.deliveryLogs = append(d.deliveryLogs, log)
}

func isSubscribed(events []string, targetEvent string) bool {
	for _, ev := range events {
		if ev == "*" || ev == targetEvent {
			return true
		}
	}
	return false
}
