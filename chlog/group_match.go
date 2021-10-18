package chlog

const DefaultGroup = "Other"

type Group struct {
	Name string
}

// Rule struct
type Rule struct {
	Name string
}

// DefaultMatcher struct
type DefaultMatcher struct {
	StartWiths []string
	Keywords []string
}

// GroupMatch struct
type GroupMatch struct {
	// Names define group names and sort
	Names []string
	Rules []string
	// MatchHandler func
	MatchHandler func(msg string) string
}

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
