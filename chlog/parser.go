package chlog

import (
	"strings"

	"github.com/gookit/goutil"
)

// Sep consts for parse git log
const Sep = " | "

// see https://devhints.io/git-log-format
// see https://git-scm.com/docs/pretty-formats
const (
	// LogFmtHs - %n new line
	// id, msg
	LogFmtHs = "%H | %s"
	// LogFmtHsa id, msg, author
	LogFmtHsa = "%H | %s | %an"
	// LogFmtHsc id, msg, committer
	LogFmtHsc = "%H | %s | %cn"
	// LogFmtHsd id, msg, author date
	LogFmtHsd = "%H | %s | %ai"
	// LogFmtHsd1 id, msg, commit date
	LogFmtHsd1 = "%H | %s | %ci"
)

// LineParser interface define
type LineParser interface {
	Parse(line string, c *Changelog) *LogItem
}

// LineParseFunc func define
type LineParseFunc func(line string, c *Changelog) *LogItem

// Parse log line to log item
func (f LineParseFunc) Parse(line string, c *Changelog) *LogItem {
	return f(line, c)
}

// BuiltInParser struct
var BuiltInParser = LineParseFunc(func(line string, c *Changelog) *LogItem {
	li := &LogItem{}
	switch c.cfg.LogFormat {
	case LogFmtHs:
		ss := strings.SplitN(line, Sep, 2)
		if len(ss) < 2 {
			return nil
		}

		li.HashID, li.Msg = ss[0], ss[1]
	case LogFmtHsa:
		ss := strings.SplitN(line, Sep, 3)
		if len(ss) < 3 {
			return nil
		}

		li.HashID, li.Msg, li.Author = ss[0], ss[1], ss[2]
	case LogFmtHsc:
		ss := strings.SplitN(line, Sep, 3)
		if len(ss) < 3 {
			return nil
		}

		li.HashID, li.Msg, li.Committer = ss[0], ss[1], ss[2]
	case LogFmtHsd, LogFmtHsd1:
		ss := strings.SplitN(line, Sep, 3)
		if len(ss) < 3 {
			return nil
		}

		li.HashID, li.Msg, li.Date = ss[0], ss[1], ss[2]
	default:
		goutil.Panicf("unsupported log format '%s'", c.cfg.LogFormat)
	}

	return li
})
