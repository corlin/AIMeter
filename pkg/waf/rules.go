package waf

import (
	"regexp"
	"strings"
	"sync"

	"github.com/corlin/AIMeter/pkg/domain"
)

// CompiledRule wraps domain rule with compiled regexes
type CompiledRule struct {
	Rule    domain.WAFRule
	Regexes []*regexp.Regexp
}

// RuleRegistry manages detection patterns with thread-safety
type RuleRegistry struct {
	mu    sync.RWMutex
	rules map[string]*CompiledRule
}

// NewRuleRegistry initializes the registry
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]*CompiledRule),
	}
}

// RegisterRule compiles regexes and registers a rule
func (r *RuleRegistry) RegisterRule(rule domain.WAFRule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	compiled := make([]*regexp.Regexp, 0, len(rule.Patterns))
	for _, p := range rule.Patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}

	r.rules[strings.ToLower(rule.ID)] = &CompiledRule{
		Rule:    rule,
		Regexes: compiled,
	}
}

// GetRule returns a rule by ID
func (r *RuleRegistry) GetRule(id string) (domain.WAFRule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cr, exists := r.rules[strings.ToLower(id)]
	if !exists {
		return domain.WAFRule{}, false
	}
	return cr.Rule, true
}

// ListRules returns all rules in the registry
func (r *RuleRegistry) ListRules() []domain.WAFRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.WAFRule, 0, len(r.rules))
	for _, cr := range r.rules {
		res = append(res, cr.Rule)
	}
	return res
}

// MatchPrompt checks prompt against all enabled rules and returns matched rules and score
func (r *RuleRegistry) MatchPrompt(prompt string) ([]domain.WAFRule, int, domain.WAFThreatCategory) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []domain.WAFRule
	totalScore := 0
	categoryScores := make(map[domain.WAFThreatCategory]int)

	for _, cr := range r.rules {
		if !cr.Rule.Enabled {
			continue
		}
		for _, re := range cr.Regexes {
			if re.MatchString(prompt) {
				matched = append(matched, cr.Rule)
				totalScore += cr.Rule.ThreatScore
				categoryScores[cr.Rule.Category] += cr.Rule.ThreatScore
				break
			}
		}
	}

	// Determine dominant category
	dominantCategory := domain.WAFThreatPromptInjection
	maxCatScore := 0
	for cat, s := range categoryScores {
		if s > maxCatScore {
			maxCatScore = s
			dominantCategory = cat
		}
	}

	if totalScore > 100 {
		totalScore = 100
	}

	return matched, totalScore, dominantCategory
}
