package waf

import (
	"strings"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ThreatDetector inspects prompt payloads using rule signatures and adversarial heuristics
type ThreatDetector struct {
	rules *RuleRegistry
}

// NewThreatDetector creates a new detector instance
func NewThreatDetector(rules *RuleRegistry) *ThreatDetector {
	return &ThreatDetector{
		rules: rules,
	}
}

// DetectVerdict represents internal detection output
type DetectVerdict struct {
	ThreatScore      float64
	ThreatCategory   domain.WAFThreatCategory
	Action           domain.WAFAction
	TriggeredRules   []string
	SanitizedPrompt  string
	EstimatedLossUSD float64
	BlockReason      string
}

// Inspect evaluates a raw prompt string and derives safety verdict
func (d *ThreatDetector) Inspect(prompt string, model string) DetectVerdict {
	matchedRules, ruleScore, dominantCategory := d.rules.MatchPrompt(prompt)

	ruleNames := make([]string, 0, len(matchedRules))
	for _, r := range matchedRules {
		ruleNames = append(ruleNames, r.Name)
	}

	finalScore := float64(ruleScore)

	// Heuristic 1: Denial-of-wallet token drain heuristics (extremely long payload > 30000 chars)
	if len(prompt) > 30000 {
		finalScore += 25.0
		if dominantCategory == "" || dominantCategory == domain.WAFThreatPromptInjection {
			dominantCategory = domain.WAFThreatDenialOfWallet
		}
		ruleNames = append(ruleNames, "Payload-Size-Anomalous-Denial-Of-Wallet")
	}

	// Heuristic 2: Repetitive word pattern check (common in infinite recursive thinking traps)
	words := strings.Fields(prompt)
	if len(words) > 50 {
		repeatedCount := 0
		for i := 1; i < len(words); i++ {
			if strings.EqualFold(words[i], words[i-1]) {
				repeatedCount++
			}
		}
		if repeatedCount > 15 {
			finalScore += 30.0
			dominantCategory = domain.WAFThreatDenialOfWallet
			ruleNames = append(ruleNames, "Repetitive-Stutter-Recursive-Trap")
		}
	}

	if finalScore > 100.0 {
		finalScore = 100.0
	}

	// Determine Action
	var action domain.WAFAction
	var blockReason string

	if finalScore >= 70.0 {
		action = domain.WAFActionBlock
		blockReason = "Critical adversarial risk detected (" + string(dominantCategory) + ")"
	} else if finalScore >= 35.0 {
		action = domain.WAFActionSanitize
	} else {
		action = domain.WAFActionAllow
	}

	// Calculate Estimated Financial Loss Avoided
	estimatedLoss := 0.05 // Baseline flagship call cost
	if dominantCategory == domain.WAFThreatDenialOfWallet {
		// Denial of wallet attempts to drain maximum output tokens (e.g. 16k tokens @ $0.03/1K = ~$0.48)
		estimatedLoss = 0.45
	} else if dominantCategory == domain.WAFThreatJailbreakDAN {
		estimatedLoss = 0.08
	}

	// Sanitize prompt if required (strip injection directives)
	sanitized := prompt
	if action == domain.WAFActionSanitize {
		sanitized = stripAdversarialDirectives(prompt)
	}

	return DetectVerdict{
		ThreatScore:      finalScore,
		ThreatCategory:   dominantCategory,
		Action:           action,
		TriggeredRules:   ruleNames,
		SanitizedPrompt:  sanitized,
		EstimatedLossUSD: estimatedLoss,
		BlockReason:      blockReason,
	}
}

// stripAdversarialDirectives performs soft sanitization
func stripAdversarialDirectives(prompt string) string {
	lower := strings.ToLower(prompt)
	for _, kw := range []string{
		"ignore previous instructions",
		"ignore all rules",
		"disregard instructions",
		"do anything now",
	} {
		if idx := strings.Index(lower, kw); idx != -1 {
			return strings.TrimSpace(prompt[:idx])
		}
	}
	return prompt
}
