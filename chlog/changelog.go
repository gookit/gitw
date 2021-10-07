package chlog

import (
	"errors"
	"strings"

	"github.com/gookit/goutil"
	"github.com/gookit/goutil/strutil"
)

const (
	Sep = " | "

	DefaultGroup = "Other"
)

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

func (f LineParseFunc) Parse(line string, c *Changelog) *LogItem {
	return f(line, c)
}

// LogItem struct
type LogItem struct {
	HashId    string // %H %h
	ParentId  string // %P %p
	Msg       string // %s
	Date      string // %ci
	Author    string // %an
	Committer string // %cn
}

// Changelog struct
type Changelog struct {
	parsed, generated bool
	// the generated change log text
	changelog string
	// The git log output. eg: `git log --pretty="format:%H"`
	// see https://devhints.io/git-log-format
	// and https://git-scm.com/docs/pretty-formats
	logText string
	// built-in log format string on the `git log --pretty="format:%H"`.
	// see consts LogFmt*
	LogFormat string
	// the parsed log items
	logItems []*LogItem
	// the formatted lines by formatter
	formatted []string
	// the valid commit log count after parse and formatted.
	logCount int
	// LineFilters The log line filters
	LineFilters []func(line string) bool
	// LineParser can custom log line parser
	LineParser LineParser
	// ItemFilters The parsed log item filters
	ItemFilters []func(li *LogItem) bool
	// The item formatter. format each item to string
	formatter func(li *LogItem) string
	// Title string for formatted text. eg: "## Change Log"
	Title string
	// RepoURL repo URL address
	RepoURL string
	// NoGroup Not output group name line.
	NoGroup bool
	// RmRepeat remove repeated log by message
	RmRepeat bool
}

// New object with git log output text
func New(gitLogOut string) *Changelog {
	return &Changelog{
		logText: gitLogOut,
	}
}

// NewEmpty object
func NewEmpty() *Changelog {
	return &Changelog{}
}

func (c *Changelog) Load(gitLogOut string) {
	c.logText = gitLogOut
}

var ErrEmptyLogText = errors.New("empty git log text for parse")

// Parse the loaded git log text
func (c *Changelog) Parse() (err error) {
	if c.parsed {
		return
	}

	c.parsed = true

	str := strings.TrimSpace(c.logText)
	if str == "" {
		return ErrEmptyLogText
	}

	if c.LineParser == nil {
		c.LineParser = BuiltInParser
	}

	parser := c.LineParser

	for _, line := range strings.Split(str, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		line = strings.Trim(line, "\"' ")
		if line == "" {
			continue
		}

		// line filters
		if !c.applyLineFilters(line) {
			continue
		}

		// parse
		li := parser.Parse(line, c)

		// item filter

		// remove repeat msg item
		if c.RmRepeat {
			msgId := strutil.Md5(li.Msg)
		}
	}


	return
}

func (c *Changelog) applyLineFilters(line string) bool {
	for _, filter := range c.LineFilters {
		if !filter(line) {
			return false
		}
	}
	return true
}

// Generate the changelog by parsed log items
func (c *Changelog) Generate() *Changelog {
	if c.generated {
		return c
	}

	c.generated = true

	// ensure parse
	c.Parse()

	return c
}

// BuiltInParser struct
var  BuiltInParser = LineParseFunc(func(line string, c *Changelog) *LogItem {
	li := &LogItem{}
	switch c.LogFormat {
	case LogFmtHs:
		ss := strings.SplitN(line, Sep, 2)

		li.HashId, li.Msg = ss[0], ss[1]
	case LogFmtHsa:
		ss := strings.SplitN(line, Sep, 3)
		li.HashId, li.Msg, li.Author = ss[0], ss[1], ss[2]
	case LogFmtHsc:
		ss := strings.SplitN(line, Sep, 3)
		li.HashId, li.Msg, li.Committer = ss[0], ss[1], ss[2]

	case LogFmtHsd,LogFmtHsd1:
		ss := strings.SplitN(line, Sep, 3)
		li.HashId, li.Msg, li.Date = ss[0], ss[1], ss[2]
	default:
		goutil.Panicf("unsupported log format '%s'", c.LogFormat)
	}

	return li
})
