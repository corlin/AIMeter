package dlp

import (
	"strings"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestScanner_EntityDetections(t *testing.T) {
	scanner := NewScanner()

	policy := &domain.DLPPolicy{
		Enabled:       true,
		DefaultAction: domain.DLPActionMask,
		EntityActions: map[string]domain.DLPAction{
			string(domain.DLPEntityAPIKey): domain.DLPActionBlock,
		},
		CustomKeywords: []string{"ProjectX_Secret"},
	}

	tests := []struct {
		name         string
		input        string
		expectedType domain.DLPEntityType
		expectAction domain.DLPAction
	}{
		{
			name:         "Email Detection",
			input:        "Please contact user at alice.smith@example.com for access",
			expectedType: domain.DLPEntityEmail,
			expectAction: domain.DLPActionMask,
		},
		{
			name:         "China Phone Detection",
			input:        "Urgent callback requested to +8613812345678 immediately",
			expectedType: domain.DLPEntityPhone,
			expectAction: domain.DLPActionMask,
		},
		{
			name:         "China Phone Hyphenated",
			input:        "Office mobile 139-8765-4321 on duty",
			expectedType: domain.DLPEntityPhone,
			expectAction: domain.DLPActionMask,
		},
		{
			name:         "China Resident ID Card",
			input:        "Resident identity code 110101199003072398 verified",
			expectedType: domain.DLPEntityIDCard,
			expectAction: domain.DLPActionMask,
		},
		{
			name:         "OpenAI API Key Detection",
			input:        "Leaked credential sk-proj-1234567890abcdefghijklmnopqrstuvwxyz",
			expectedType: domain.DLPEntityAPIKey,
			expectAction: domain.DLPActionBlock,
		},
		{
			name:         "AWS Access Key ID",
			input:        "Cloud storage key AKIAIOSFODNN7EXAMPLE configured",
			expectedType: domain.DLPEntityAPIKey,
			expectAction: domain.DLPActionBlock,
		},
		{
			name:         "JWT Token Detection",
			input:        "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedType: domain.DLPEntityJWTToken,
			expectAction: domain.DLPActionMask,
		},
		{
			name:         "Private Network IP",
			input:        "Internal server host 192.168.1.105:8080 online",
			expectedType: domain.DLPEntityPrivateIP,
			expectAction: domain.DLPActionMask,
		},
		{
			name:         "Database Connection String",
			input:        "Connecting via postgres://admin:secret123@db.prod.internal:5432/finance",
			expectedType: domain.DLPEntityConnectionString,
			expectAction: domain.DLPActionMask,
		},
		{
			name:         "Custom Keyword",
			input:        "Confidential report contains ProjectX_Secret details",
			expectedType: domain.DLPEntityCustomKeyword,
			expectAction: domain.DLPActionMask,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := scanner.ScanAndRemediate(tc.input, policy)
			if !res.HasViolations {
				t.Fatalf("expected violation for %q, but got none", tc.input)
			}
			found := false
			for _, ent := range res.DetectedEntities {
				if ent.Type == tc.expectedType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected entity type %s in results: %+v", tc.expectedType, res.DetectedEntities)
			}
			if res.ActionTaken != tc.expectAction {
				t.Errorf("expected action %s, got %s", tc.expectAction, res.ActionTaken)
			}
		})
	}
}

func TestScanner_BankCardLuhn(t *testing.T) {
	scanner := NewScanner()
	policy := &domain.DLPPolicy{
		Enabled:       true,
		DefaultAction: domain.DLPActionMask,
	}

	// Valid Luhn Visa card: 4532015112830366 (checksum mod 10 = 0)
	validVisa := "Pay via card 4532015112830366 now"
	resValid := scanner.ScanAndRemediate(validVisa, policy)
	if !resValid.HasViolations {
		t.Fatalf("expected valid card to be detected")
	}
	if len(resValid.DetectedEntities) != 1 || resValid.DetectedEntities[0].Type != domain.DLPEntityBankCard {
		t.Errorf("expected bank card entity, got %+v", resValid.DetectedEntities)
	}

	// Invalid card (checksum failure): 4532015112830367
	invalidCard := "Failed card number 4532015112830367 entered"
	resInvalid := scanner.ScanAndRemediate(invalidCard, policy)
	for _, ent := range resInvalid.DetectedEntities {
		if ent.Type == domain.DLPEntityBankCard {
			t.Errorf("invalid Luhn card should NOT be flagged as bank card, but was detected: %s", ent.RawText)
		}
	}
}

func TestScanner_BidirectionalMaskingAndUnmask(t *testing.T) {
	scanner := NewScanner()
	policy := &domain.DLPPolicy{
		Enabled:       true,
		DefaultAction: domain.DLPActionMask,
	}

	rawInput := "Hello, my phone is 13800138000 and email is support@aimeter.dev. Please confirm!"
	res := scanner.ScanAndRemediate(rawInput, policy)

	if !res.HasViolations {
		t.Fatalf("expected violations detected")
	}
	if res.ActionTaken != domain.DLPActionMask {
		t.Fatalf("expected MASK action, got %s", res.ActionTaken)
	}

	// Verify sanitized text does NOT contain sensitive strings
	if strings.Contains(res.SanitizedText, "13800138000") {
		t.Errorf("phone was not masked: %s", res.SanitizedText)
	}
	if strings.Contains(res.SanitizedText, "support@aimeter.dev") {
		t.Errorf("email was not masked: %s", res.SanitizedText)
	}

	// Verify placeholders exist
	if !strings.Contains(res.SanitizedText, "[AIMETER_PHONE_1]") {
		t.Errorf("missing phone placeholder: %s", res.SanitizedText)
	}
	if !strings.Contains(res.SanitizedText, "[AIMETER_EMAIL_1]") {
		t.Errorf("missing email placeholder: %s", res.SanitizedText)
	}

	// Simulate LLM returning response that references the placeholders
	llmOutput := "We have registered your phone [AIMETER_PHONE_1] and sent an activation email to [AIMETER_EMAIL_1]."
	restored := scanner.UnmaskText(llmOutput, res.PlaceholderVault)

	expectedRestored := "We have registered your phone 13800138000 and sent an activation email to support@aimeter.dev."
	if restored != expectedRestored {
		t.Errorf("unmask failed!\nExpected: %s\nGot:      %s", expectedRestored, restored)
	}
}

func TestManager_ConcurrentSimulationAndLifecycle(t *testing.T) {
	mgr := NewManager("")

	policy := domain.DLPPolicy{
		ID:            "pol_corp",
		TenantID:      "tenant-corp",
		Name:          "Corporate Privacy Policy",
		Enabled:       true,
		DefaultAction: domain.DLPActionMask,
		EntityActions: map[string]domain.DLPAction{
			string(domain.DLPEntityAPIKey): domain.DLPActionBlock,
		},
	}
	mgr.SetPolicy(policy)

	// Verify retrieval
	got := mgr.GetPolicy("tenant-corp")
	if got.Name != "Corporate Privacy Policy" {
		t.Fatalf("policy retrieval mismatch: %+v", got)
	}

	// Concurrently run message scans and simulations
	var wg sync.WaitGroup
	workers := 10
	iterations := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				msgs := []domain.ChatMessage{
					{Role: "user", Content: "Contact me at user@corp.io or 13912345678"},
				}
				sanitized, vault, blocked, res := mgr.ScanMessages("tenant-corp", "req-123", msgs)
				if blocked {
					t.Errorf("worker %d: unexpected block", workerID)
				}
				if !res.HasViolations || len(vault) < 2 {
					t.Errorf("worker %d: expected violations and vault populated, got: %+v", workerID, res)
				}
				contentStr, _ := sanitized[0].Content.(string)
				if strings.Contains(contentStr, "user@corp.io") {
					t.Errorf("worker %d: message was not masked", workerID)
				}

				// Run simulate
				simRes := mgr.Simulate(domain.DLPSimulateRequest{
					TenantID: "tenant-corp",
					Text:     "Leaked token sk-proj-1234567890abcdef12345678",
				})
				if simRes.ActionTaken != domain.DLPActionBlock {
					t.Errorf("worker %d: expected block for simulated sk- key, got %s", workerID, simRes.ActionTaken)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify stats
	stats := mgr.GetStats("tenant-corp")
	if stats.TotalScans == 0 || stats.TotalViolations == 0 {
		t.Fatalf("expected non-zero stats, got: %+v", stats)
	}

	// Verify audit logs
	logs := mgr.ListAuditLogs("tenant-corp", 50)
	if len(logs) == 0 {
		t.Fatalf("expected audit logs recorded, got empty")
	}
}
