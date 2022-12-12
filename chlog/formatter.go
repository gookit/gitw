package chlog

import (
	"fmt"
)

// Formatter interface
type Formatter interface {
	// MatchGroup from log msg
	MatchGroup(msg string) (group string)
	// Format the log item to line
	Format(li *LogItem) (group, fmtLine string)
}

// GroupMatcher interface
type GroupMatcher interface {
	// Match group from log msg(has been trimmed)
	Match(msg string) (group string)
}

// built-in formatters
const (
	FormatterSimple    = "simple"
	FormatterMarkdown  = "markdown"
	FormatterGhRelease = "ghr"
)

// SimpleFormatter struct
type SimpleFormatter struct {
	// GroupMatch group match handler.
	GroupMatch GroupMatcher
}

// MatchGroup from log msg
func (f *SimpleFormatter) MatchGroup(msg string) (group string) {
	if f.GroupMatch != nil {
		return f.GroupMatch.Match(msg)
	}

	return DefaultMatcher.Match(msg)
}

// Format the log item to line
func (f *SimpleFormatter) Format(li *LogItem) (group, fmtLine string) {
	fmtLine = " - "
	if li.HashID != "" {
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

// Format the log item to line
func (f *MarkdownFormatter) Format(li *LogItem) (group, fmtLine string) {
	group = f.MatchGroup(li.Msg)

	if li.HashID != "" {
		// full url.
		// eg: https://github.com/inhere/kite/commit/ebd90a304755218726df4eb398fd081c08d04b9a
		fmtLine = fmt.Sprintf("- %s [%s](%s/commit/%s)", li.Msg, li.AbbrevID(), f.RepoURL, li.HashID)
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
func (f *GHReleaseFormatter) Format(li *LogItem) (group, fmtLine string) {
	group = f.MatchGroup(li.Msg)

	if li.HashID != "" {
		// full url.
		// eg: https://github.com/inhere/kite/commit/ebd90a304755218726df4eb398fd081c08d04b9a
		fmtLine = fmt.Sprintf("- %s %s/commit/%s", li.Msg, f.RepoURL, li.HashID)
	} else {
		fmtLine = " - " + li.Msg
	}

	if user := li.Username(); user != "" {
		fmtLine += " by(@" + user + ")"
	}
	return
}
