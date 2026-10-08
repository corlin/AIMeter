package alert_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/alert"
)

func TestDetectChannelType(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://open.feishu.cn/open-apis/bot/v2/hook/xxx", alert.ChannelFeishu},
		{"https://open.larksuite.com/open-apis/bot/v2/hook/xxx", alert.ChannelFeishu},
		{"https://oapi.dingtalk.com/robot/send?access_token=xxx", alert.ChannelDingTalk},
		{"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx", alert.ChannelWeCom},
		{"https://hooks.slack.com/services/T00/B00/X00", alert.ChannelSlack},
		{"https://api.mycorp.com/v1/alerts", alert.ChannelGenericJSON},
	}

	for _, tt := range tests {
		actual := alert.DetectChannelType(tt.url)
		if actual != tt.expected {
			t.Errorf("URL %s expected %s, got %s", tt.url, tt.expected, actual)
		}
	}
}

func TestDispatch_FeishuCardFormat(t *testing.T) {
	payloadCh := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		payloadCh <- body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": 0, "msg": "success"}`))
	}))
	defer server.Close()

	dispatcher := alert.NewAlertDispatcher(nil)
	dispatcher.RegisterChannel(alert.AlertChannel{
		ID:          "ch-feishu",
		TenantID:    "org-test",
		Name:        "FinOps SRE Feishu Group",
		ChannelType: alert.ChannelFeishu,
		WebhookURL:  server.URL,
		Enabled:     true,
	})

	event := alert.NotificationEvent{
		TenantID:   "org-test",
		WorkflowID: "customer-service-agent",
		EventType:  alert.EventBudgetExceeded,
		Severity:   "critical",
		Title:      "月度预算已被 100% 耗尽",
		Message:    "租户 org-test 已达到 $500.00 月度预算限额，后续调用将自动拦截。",
		Metrics: map[string]interface{}{
			"spent_usd": 502.40,
			"limit_usd": 500.00,
		},
		TriggeredAt: time.Now(),
	}

	dispatcher.Dispatch(context.Background(), event)

	var receivedPayload map[string]interface{}
	select {
	case body := <-payloadCh:
		_ = json.Unmarshal(body, &receivedPayload)
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for Feishu alert delivery")
	}

	if receivedPayload["msg_type"] != "interactive" {
		t.Fatalf("Expected msg_type 'interactive', got %v", receivedPayload["msg_type"])
	}
	card, ok := receivedPayload["card"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected card object in payload: %v", receivedPayload)
	}
	header, ok := card["header"].(map[string]interface{})
	if !ok || header["template"] != "red" {
		t.Errorf("Expected red critical header template, got %v", header)
	}
}

func TestDispatch_DeduplicationCooldown(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := alert.NewAlertDispatcher(nil)
	dispatcher.RegisterChannel(alert.AlertChannel{
		ID:          "ch-dedup",
		TenantID:    "org-cooldown",
		Name:        "Test Cooldown",
		ChannelType: alert.ChannelGenericJSON,
		WebhookURL:  server.URL,
		Enabled:     true,
	})

	event := alert.NotificationEvent{
		TenantID:    "org-cooldown",
		WorkflowID:  "data-indexer",
		EventType:   alert.EventCircuitBreakerTripped,
		Severity:    "critical",
		Title:       "熔断器触发跳闸",
		Message:     "Workflow data-indexer tripped circuit breaker due to high failure rate.",
		TriggeredAt: time.Now(),
	}

	// 1. First dispatch: should succeed
	dispatcher.Dispatch(context.Background(), event)
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatalf("Expected 1 call, got %d", atomic.LoadInt32(&callCount))
	}

	// 2. Immediate duplicate dispatch: should be suppressed
	dispatcher.Dispatch(context.Background(), event)
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected still 1 call due to cooldown suppression, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestTestChannelConnectivity(t *testing.T) {
	receivedHeader := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-AIMeter-Signature")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	dispatcher := alert.NewAlertDispatcher(nil)
	ch := alert.AlertChannel{
		ID:          "ch-test",
		Name:        "Generic Secured Webhook",
		ChannelType: alert.ChannelGenericJSON,
		WebhookURL:  server.URL,
		Secret:      "my-secret-key-123",
		Enabled:     true,
	}

	log := dispatcher.TestChannel(context.Background(), ch)

	if !log.Success {
		t.Fatalf("Expected test channel to succeed, got error: %s", log.ErrorMessage)
	}
	if log.HTTPStatus != http.StatusOK {
		t.Errorf("Expected HTTP 200, got %d", log.HTTPStatus)
	}
	if !strings.HasPrefix(receivedHeader, "sha256=") {
		t.Errorf("Expected HMAC signature header, got %s", receivedHeader)
	}
}
