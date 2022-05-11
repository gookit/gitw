package chlog

import "github.com/gookit/goutil/strutil"

// DefaultGroup name
var DefaultGroup = "Other"

// Rule struct
type Rule struct {
	// Name for group
	Name string
	// StartWiths message start withs string.
	StartWiths []string
	// Contains message should contain there are strings.
	Contains []string
}

// RuleMatcher struct
type RuleMatcher struct {
	// Names define group names and sort
	Names []string
	Rules []Rule
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

// DefaultMatcher for match group name.
var DefaultMatcher = &RuleMatcher{
	Names: []string{"Feature", "Update", "Fixed"},
	Rules: []Rule{
		{
			Name:       "Feature",
			StartWiths: []string{"feat", "new"},
			Contains:   []string{"feature"},
		},
		{
			Name:       "Update",
			StartWiths: []string{"up:", "update"},
			Contains:   []string{" update"},
		},
		{
			Name:       "Fixed",
			StartWiths: []string{"bug", "close", "fix"},
			Contains:   []string{"fix:", "bug:"},
		},
	},
}
