package dlp

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

var (
	// Pre-compiled regex patterns for PII & Secrets
	reEmail        = regexp.MustCompile(`(?i)\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	rePhoneCN      = regexp.MustCompile(`\b(?:(?:\+?86)|(?:\(\+?86\)))?(?:1[3-9]\d{9})\b`)
	rePhoneHyphen  = regexp.MustCompile(`\b1[3-9]\d{1}-\d{4}-\d{4}\b`)
	reIDCardCN     = regexp.MustCompile(`\b[1-9]\d{5}(?:18|19|20)\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01])\d{3}[\dXx]\b`)
	reBankCard     = regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9]{2})[0-9]{12}|3[47][0-9]{13}|62[0-9]{14,17})\b`)
	reAPIKeyOpenAI = regexp.MustCompile(`\bsk-(?:proj-)?[a-zA-Z0-9_-]{20,}\b`)
	reAPIKeyAWS    = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	reJWT          = regexp.MustCompile(`\bey[a-zA-Z0-9_-]{15,}\.ey[a-zA-Z0-9_-]{15,}\.[a-zA-Z0-9_-]{10,}\b`)
	rePrivateIP    = regexp.MustCompile(`\b(?:10\.\d{1,3}\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3})\b`)
	reConnString   = regexp.MustCompile(`(?i)\b(?:postgres|mysql|mongodb|redis):\/\/[^\s"']+\b`)
)

// Scanner implements high-performance zero-external-dependency PII detection
type Scanner struct{}

// NewScanner creates a new scanner instance
func NewScanner() *Scanner {
	return &Scanner{}
}

// Luhn algorithm verification for credit card numbers
func isValidLuhn(number string) bool {
	digits := make([]int, 0, len(number))
	for _, ch := range number {
		if ch >= '0' && ch <= '9' {
			digits = append(digits, int(ch-'0'))
		}
	}
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}

	checksum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		val := digits[i]
		if double {
			val *= 2
			if val > 9 {
				val -= 9
			}
		}
		checksum += val
		double = !double
	}
	return checksum%10 == 0
}

// ScanAndRemediate performs 2-phase scanning and applies configured remediation (Block / Mask / Audit)
func (s *Scanner) ScanAndRemediate(text string, policy *domain.DLPPolicy) domain.DLPScanResult {
	start := time.Now()

	if policy == nil || !policy.Enabled || len(text) == 0 {
		return domain.DLPScanResult{
			HasViolations:    false,
			ActionTaken:      domain.DLPActionAudit,
			DetectedEntities: []domain.DLPDetectedEntity{},
			SanitizedText:    text,
			PlaceholderVault: make(map[string]string),
			ScanDurationUs:   time.Since(start).Microseconds(),
		}
	}

	var candidates []domain.DLPDetectedEntity

	// Fast pre-filter checks
	hasAt := strings.Contains(text, "@")
	hasDigit := strings.ContainsAny(text, "0123456789")
	hasSk := strings.Contains(text, "sk-")
	hasAkia := strings.Contains(text, "AKIA")
	hasEy := strings.Contains(text, "ey")
	hasProto := strings.Contains(text, "://")

	// 1. Email
	if hasAt {
		matches := reEmail.FindAllStringIndex(text, -1)
		for _, m := range matches {
			raw := text[m[0]:m[1]]
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityEmail,
				RawText:  raw,
				StartIdx: m[0],
				EndIdx:   m[1],
			})
		}
	}

	// 2. Phone Numbers
	if hasDigit {
		for _, re := range []*regexp.Regexp{rePhoneCN, rePhoneHyphen} {
			matches := re.FindAllStringIndex(text, -1)
			for _, m := range matches {
				raw := text[m[0]:m[1]]
				candidates = append(candidates, domain.DLPDetectedEntity{
					Type:     domain.DLPEntityPhone,
					RawText:  raw,
					StartIdx: m[0],
					EndIdx:   m[1],
				})
			}
		}

		// 3. ID Card (China)
		idMatches := reIDCardCN.FindAllStringIndex(text, -1)
		for _, m := range idMatches {
			raw := text[m[0]:m[1]]
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityIDCard,
				RawText:  raw,
				StartIdx: m[0],
				EndIdx:   m[1],
			})
		}

		// 4. Bank Card (with Luhn)
		cardMatches := reBankCard.FindAllStringIndex(text, -1)
		for _, m := range cardMatches {
			raw := text[m[0]:m[1]]
			if isValidLuhn(raw) {
				candidates = append(candidates, domain.DLPDetectedEntity{
					Type:     domain.DLPEntityBankCard,
					RawText:  raw,
					StartIdx: m[0],
					EndIdx:   m[1],
				})
			}
		}

		// 5. Private IP
		ipMatches := rePrivateIP.FindAllStringIndex(text, -1)
		for _, m := range ipMatches {
			raw := text[m[0]:m[1]]
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityPrivateIP,
				RawText:  raw,
				StartIdx: m[0],
				EndIdx:   m[1],
			})
		}
	}

	// 6. API Keys
	if hasSk {
		matches := reAPIKeyOpenAI.FindAllStringIndex(text, -1)
		for _, m := range matches {
			raw := text[m[0]:m[1]]
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityAPIKey,
				RawText:  raw,
				StartIdx: m[0],
				EndIdx:   m[1],
			})
		}
	}
	if hasAkia {
		matches := reAPIKeyAWS.FindAllStringIndex(text, -1)
		for _, m := range matches {
			raw := text[m[0]:m[1]]
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityAPIKey,
				RawText:  raw,
				StartIdx: m[0],
				EndIdx:   m[1],
			})
		}
	}

	// 7. JWT Tokens
	if hasEy {
		matches := reJWT.FindAllStringIndex(text, -1)
		for _, m := range matches {
			raw := text[m[0]:m[1]]
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityJWTToken,
				RawText:  raw,
				StartIdx: m[0],
				EndIdx:   m[1],
			})
		}
	}

	// 8. Connection Strings
	if hasProto {
		matches := reConnString.FindAllStringIndex(text, -1)
		for _, m := range matches {
			raw := text[m[0]:m[1]]
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityConnectionString,
				RawText:  raw,
				StartIdx: m[0],
				EndIdx:   m[1],
			})
		}
	}

	// 9. Custom Keywords
	for _, kw := range policy.CustomKeywords {
		kw = strings.TrimSpace(kw)
		if len(kw) == 0 {
			continue
		}
		idx := 0
		for {
			pos := strings.Index(strings.ToLower(text[idx:]), strings.ToLower(kw))
			if pos == -1 {
				break
			}
			startIdx := idx + pos
			endIdx := startIdx + len(kw)
			candidates = append(candidates, domain.DLPDetectedEntity{
				Type:     domain.DLPEntityCustomKeyword,
				RawText:  text[startIdx:endIdx],
				StartIdx: startIdx,
				EndIdx:   endIdx,
			})
			idx = endIdx
		}
	}

	if len(candidates) == 0 {
		return domain.DLPScanResult{
			HasViolations:    false,
			ActionTaken:      domain.DLPActionAudit,
			DetectedEntities: []domain.DLPDetectedEntity{},
			SanitizedText:    text,
			PlaceholderVault: make(map[string]string),
			ScanDurationUs:   time.Since(start).Microseconds(),
		}
	}

	// De-duplicate & Sort candidates by StartIdx
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].StartIdx < candidates[j].StartIdx
	})

	filtered := make([]domain.DLPDetectedEntity, 0, len(candidates))
	lastEnd := 0
	for _, c := range candidates {
		if c.StartIdx >= lastEnd {
			filtered = append(filtered, c)
			lastEnd = c.EndIdx
		}
	}

	// Assign Actions per entity type
	overallAction := domain.DLPActionAudit
	vault := make(map[string]string)
	counters := make(map[string]int)

	for i := range filtered {
		entity := &filtered[i]
		action := policy.DefaultAction
		if specific, exists := policy.EntityActions[string(entity.Type)]; exists && specific != "" {
			action = specific
		}
		entity.ActionTaken = action

		// Escalate overall action: Block > Mask > Audit
		if action == domain.DLPActionBlock {
			overallAction = domain.DLPActionBlock
		} else if action == domain.DLPActionMask && overallAction != domain.DLPActionBlock {
			overallAction = domain.DLPActionMask
		}

		counters[string(entity.Type)]++
		placeholder := fmt.Sprintf("[AIMETER_%s_%d]", strings.ToUpper(string(entity.Type)), counters[string(entity.Type)])
		entity.MaskedPlaceholder = placeholder
		vault[placeholder] = entity.RawText
	}

	// If overall action is BLOCK, no need to rewrite text, caller will block
	if overallAction == domain.DLPActionBlock {
		return domain.DLPScanResult{
			HasViolations:    true,
			ActionTaken:      domain.DLPActionBlock,
			DetectedEntities: filtered,
			SanitizedText:    text,
			PlaceholderVault: vault,
			ScanDurationUs:   time.Since(start).Microseconds(),
		}
	}

	// If overall action is MASK, rebuild sanitized text
	sanitizedText := text
	if overallAction == domain.DLPActionMask {
		var sb strings.Builder
		cursor := 0
		for _, e := range filtered {
			if e.ActionTaken == domain.DLPActionMask {
				sb.WriteString(text[cursor:e.StartIdx])
				sb.WriteString(e.MaskedPlaceholder)
				cursor = e.EndIdx
			}
		}
		sb.WriteString(text[cursor:])
		sanitizedText = sb.String()
	}

	return domain.DLPScanResult{
		HasViolations:    true,
		ActionTaken:      overallAction,
		DetectedEntities: filtered,
		SanitizedText:    sanitizedText,
		PlaceholderVault: vault,
		ScanDurationUs:   time.Since(start).Microseconds(),
	}
}

// UnmaskText reverses masked placeholders back to original values in LLM response
func (s *Scanner) UnmaskText(text string, vault map[string]string) string {
	if len(vault) == 0 || len(text) == 0 {
		return text
	}
	result := text
	for placeholder, raw := range vault {
		result = strings.ReplaceAll(result, placeholder, raw)
	}
	return result
}
