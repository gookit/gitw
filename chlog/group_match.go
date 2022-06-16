package chlog

import "github.com/gookit/goutil/strutil"

// DefaultGroup name
var DefaultGroup = "Other"

// DefaultMatcher for match group name.
var DefaultMatcher = NewDefaultMatcher()

// Rule struct
type Rule struct {
	// Name for group
	Name string `json:"name" yaml:"name"`
	// StartWiths message start withs string.
	StartWiths []string `json:"start_withs" yaml:"start_withs"`
	// Contains message should contain there are strings.
	Contains []string `json:"contains" yaml:"contains"`
}

// RuleMatcher struct
type RuleMatcher struct {
	// Names define group names and sort
	Names []string `json:"names" yaml:"names"`
	Rules []Rule   `json:"rules" yaml:"rules"`
}

// Match group name from log message.
func (m RuleMatcher) Match(msg string) string {
	for _, rule := range m.Rules {
		if len(rule.StartWiths) > 0 && strutil.HasOnePrefix(msg, rule.StartWiths) {
			return rule.Name
		}

		if len(rule.Contains) > 0 && strutil.HasOneSub(msg, rule.Contains) {
			return rule.Name
		}
	}

	return DefaultGroup
}

// NewDefaultMatcher instance
func NewDefaultMatcher() *RuleMatcher {
	return &RuleMatcher{
		Names: []string{"Feature", "Refactor", "Update", "Fixed", DefaultGroup},
		Rules: []Rule{
			{
				Name:       "Feature",
				StartWiths: []string{"feat", "new"},
				Contains:   []string{"feat:"},
			},
			{
				Name:       "Refactor",
				StartWiths: []string{"break", "refactor"},
				Contains:   []string{"refactor:"},
			},
			{
				Name:       "Update",
				StartWiths: []string{"up ", "update"},
				Contains:   []string{"up:", "update:"},
			},
			{
				Name:       "Fixed",
				StartWiths: []string{"bug", "close", "fix"},
				Contains:   []string{"fix:", "bug:"},
			},
		},
	}
}
