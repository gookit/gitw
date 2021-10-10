package chlog

import (
	"errors"
	"strings"

	"github.com/gookit/gitwrap"
	"github.com/gookit/goutil/strutil"
)

var ErrEmptyLogText = errors.New("empty git log text for parse")

// LogItem struct
type LogItem struct {
	HashId    string // %H %h
	ParentId  string // %P %p
	Msg       string // %s
	Date      string // %ci
	Author    string // %an
	Committer string // %cn
	RepoUrl   string // Changelog.RepoUrl
}

// AbbrevID get abbrev commit ID
func (l *LogItem) AbbrevID() string {
	if l.HashId == "" {
		return ""
	}

	return strutil.Substr(l.HashId, 0, 7)
}

// Username get commit username.
func (l *LogItem) Username() string {
	if l.Author != "" {
		return l.Author
	}
	return l.Committer
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
	//	{group: [line, line, ...], ...}
	formatted map[string][]string
	// the valid commit log count after parse and formatted.
	logCount int
	// LineFilters The log line filters
	LineFilters []func(line string) bool
	// LineParser can custom log line parser
	LineParser LineParser
	// ItemFilters The parsed log item filters
	ItemFilters []func(li *LogItem) bool
	// Formatter The item formatter. format each item to string
	Formatter Formatter
	// Title string for formatted text. eg: "## Change Log"
	Title string
	// RepoURL repo URL address
	RepoURL string
	// NoGroup Not output group name line.
	NoGroup bool
	// RmRepeat remove repeated log by message
	RmRepeat bool
}

// NewWithGitLog new object with git log output text
func NewWithGitLog(gitLogOut string) *Changelog {
	cl := New()
	cl.SetLogText(gitLogOut)
	return cl
}

// New object
func New() *Changelog {
	return &Changelog{
		// init some settings
		Title:     "## Change Log",
		LogFormat: LogFmtHs,
		RmRepeat:  true,
	}
}

// WithConfig config the object
func (c *Changelog) WithConfig(fn func(c *Changelog)) *Changelog {
	fn(c)
	return c
}

// SetLogText by git log
func (c *Changelog) SetLogText(gitLogOut string) {
	c.logText = gitLogOut
}

// FetchGitLog by git log
func (c *Changelog) FetchGitLog(sha1, sha2 string, moreArgs ...string) *Changelog {
	logCmd := gitwrap.New("log", "--reverse")
	logCmd.Addf("--pretty=format:\"%s\"", c.LogFormat)
	// logCmd.Add("--no-merges")
	logCmd.Add(moreArgs...) // add custom args
	// logCmd.Addf("%s...%s", "v0.1.0", "HEAD")
	logCmd.Addf("%s...%s", sha1, sha2)

	c.SetLogText(logCmd.SafeOutput())
	return c
}

// -------------------------------------------------------------------
// parse git log
// -------------------------------------------------------------------

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

	// ensure parser exists
	if c.LineParser == nil {
		c.LineParser = BuiltInParser
	}

	parser := c.LineParser

	msgIdMap := make(map[string]int)
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

		// parse line
		li := parser.Parse(line, c)
		if li == nil {
			continue
		}

		// item filters
		if !c.applyItemFilters(li) {
			continue
		}

		// remove repeat msg item
		if c.RmRepeat {
			msgId := strutil.Md5(li.Msg)
			if _, ok := msgIdMap[msgId]; ok {
				continue
			}

			msgIdMap[msgId] = 1
		}

		c.logItems = append(c.logItems, li)
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

func (c *Changelog) applyItemFilters(li *LogItem) bool {
	for _, filter := range c.ItemFilters {
		if !filter(li) {
			return false
		}
	}
	return true
}

// -------------------------------------------------------------------
// generate and export
// -------------------------------------------------------------------

// Generate the changelog by parsed log items
func (c *Changelog) Generate() (err error) {
	if c.generated {
		return
	}

	c.generated = true

	// ensure parse
	if err = c.Parse(); err != nil {
		return err
	}

	// format parsed items
	groupNames := c.formatLogItems()
	groupCount := len(groupNames)

	var outLines []string
	// first add title
	if c.Title != "" {
		outLines = append(outLines, c.Title)
	}

	for grpName, list := range c.formatted {
		// only one group, not render group name.
		if groupCount > 1 {
			outLines = append(outLines, grpName)
		}

		outLines = append(outLines, strings.Join(list, "\n"))
	}

	c.changelog = strings.Join(outLines, "\n")
	return
}

func (c *Changelog) formatLogItems() map[string]int {
	if c.Formatter == nil {
		c.Formatter = &SimpleFormatter{}
	}

	// init field
	c.formatted = make(map[string][]string)

	groupMap := make(map[string]int, len(c.logItems))
	for _, li := range c.logItems {
		group, fmtLine := c.Formatter.Format(li)
		// invalid line
		if fmtLine == "" {
			continue
		}

		if group == "" {
			group = DefaultGroup
		}

		c.logCount++
		groupMap[group] = 1

		if list, ok := c.formatted[group]; ok {
			c.formatted[group] = append(list, fmtLine)
		} else {
			c.formatted[group] = []string{fmtLine}
		}
	}

	return groupMap
}

// String get generated change log string
func (c *Changelog) String() string {
	return c.changelog
}

// Changelog get generated change log string
func (c *Changelog) Changelog() string {
	return c.changelog
}

// Formatted get formatted change log line
// func (c *Changelog) Formatted() []string {
// 	return c.formatted
// }

// LogCount get
func (c *Changelog) LogCount() int {
	return c.logCount
}
