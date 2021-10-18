package chlog

const DefaultGroup = "Other"

// Rule struct
type Rule struct {
	// Name for group
	Name string
	// StartWiths message start withs string.
	StartWiths []string
	// Keywords message should contains there are keywords
	Keywords []string
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
		if len(rule.StartWiths) > 0 && hasOnePrefix(msg, rule.StartWiths) {
			return rule.Name
		}

		if len(rule.Keywords) > 0 && hasOneSub(msg, rule.Keywords) {
			return rule.Name
		}
	}

	return DefaultGroup
}

// SimpleMatchFunc for match group name.
var SimpleMatchFunc = func(msg string) string {
	if isFixMsg(msg) {
		return "Fixed"
	}

	if hasOnePrefix(msg, []string{"up", "add", "create"}) {
		return "Update"
	}

	if hasOnePrefix(msg, []string{"feat", "support", "new"}) {
		return "Feature"
	}
	return DefaultGroup
}
