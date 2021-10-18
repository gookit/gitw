package chlog

import (
	"fmt"
	"strings"
)

// Formatter interface
type Formatter interface {
	// MatchGroup from log msg
	MatchGroup(msg string) (group string)
	// Format the log item to line
	Format(li *LogItem) (group, fmtLine string)
}

// SimpleFormatter struct
type SimpleFormatter struct {}

// MatchGroup from log msg
func (f *SimpleFormatter) MatchGroup(msg string) (group string) {
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

// Format the log item to line
func (f *SimpleFormatter) Format(li *LogItem) (group, fmtLine string) {
	fmtLine = " - "
	if li.HashId != "" {
		fmtLine += li.AbbrevID() + " "
	}

	group = f.MatchGroup(li.Msg)

	fmtLine += li.Msg
	if user := li.Username(); user != "" {
		fmtLine += " by(" + user + ")"
	}

	return
}

// MarkdownFormatter struct
type MarkdownFormatter struct {
	SimpleFormatter
	// RepoURL git repo remote URL address
	RepoURL string
}

// MatchGroup from log msg
func (f *MarkdownFormatter) MatchGroup(msg string) (group string) {
	group = f.SimpleFormatter.MatchGroup(msg)

	return "\n### " + group + "\n"
}

// Format the log item to line
func (f *MarkdownFormatter) Format(li *LogItem) (group, fmtLine string)  {
	group = f.MatchGroup(li.Msg)

	if li.HashId != "" {
		// full url.
		// eg: https://github.com/inhere/kite/commit/ebd90a304755218726df4eb398fd081c08d04b9a
		fmtLine = fmt.Sprintf("- %s [%s](%s/commit/%s)", li.Msg, li.AbbrevID(), f.RepoURL, li.HashId)
	} else {
		fmtLine = " - " + li.Msg
	}

	if user := li.Username(); user != "" {
		fmtLine += " by(" + user + ")"
	}
	return
}

// GHReleaseFormatter struct
type GHReleaseFormatter struct {
	MarkdownFormatter
}

// Format the log item to line
func (f *GHReleaseFormatter) Format(li *LogItem) (group, fmtLine string)  {
	group = f.MatchGroup(li.Msg)

	if li.HashId != "" {
		// full url.
		// eg: https://github.com/inhere/kite/commit/ebd90a304755218726df4eb398fd081c08d04b9a
		fmtLine = fmt.Sprintf("- %s %s/commit/%s", li.Msg, f.RepoURL, li.HashId)
	} else {
		fmtLine = " - " + li.Msg
	}

	if user := li.Username(); user != "" {
		fmtLine += " by(@" + user + ")"
	}
	return
}

func isFixMsg(msg string) bool {
	if hasOnePrefix(msg, []string{"bug", "close", "fix"}) {
		return true
	}

	return strings.Contains(msg, " fix")
}

// TODO use strutil.HasOneSub
func hasOneSub(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// TODO use strutil.HasOnePrefix
func hasOnePrefix(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.HasPrefix(s, sub) {
			return true
		}
	}
	return false
}
