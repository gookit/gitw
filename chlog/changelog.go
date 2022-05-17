package chlog

import (
	"errors"
	"io"
	"strings"

	"github.com/gookit/gitw"
	"github.com/gookit/goutil/strutil"
)

// ErrEmptyLogText error
var ErrEmptyLogText = errors.New("empty git log text for parse")

// LogItem struct
type LogItem struct {
	HashID    string // %H %h
	ParentID  string // %P %p
	Msg       string // %s
	Date      string // %ci
	Author    string // %an
	Committer string // %cn
}

// AbbrevID get abbrev commit ID
func (l *LogItem) AbbrevID() string {
	if l.HashID == "" {
		return ""
	}

	return strutil.Substr(l.HashID, 0, 7)
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
	cfg *Config
	// handle mark
	parsed, generated bool
	// the generated change log text
	changelog string
	// The git log output. eg: `git log --pretty="format:%H"`
	// see https://devhints.io/git-log-format
	// and https://git-scm.com/docs/pretty-formats
	logText string
	// the parsed log items
	logItems []*LogItem
	// the formatted lines by formatter
	//	{group: [line, line, ...], ...}
	formatted map[string][]string
	// the valid commit log count after parse and formatted.
	logCount int
	// LineParser can custom log line parser
	LineParser LineParser
	// ItemFilters The parsed log item filters
	ItemFilters []ItemFilter
	// Formatter The item formatter. format each item to string
	Formatter Formatter
}

// New object
func New() *Changelog {
	return &Changelog{
		cfg: NewDefaultConfig(),
	}
}

// NewWithGitLog new object with git log output text
func NewWithGitLog(gitLogOut string) *Changelog {
	cl := New()
	cl.SetLogText(gitLogOut)
	return cl
}

// NewWithConfig object
func NewWithConfig(cfg *Config) *Changelog {
	return &Changelog{
		cfg: cfg,
	}
}

// WithFn config the object
func (c *Changelog) WithFn(fn func(c *Changelog)) *Changelog {
	fn(c)
	return c
}

// WithConfig with new config object
func (c *Changelog) WithConfig(cfg *Config) *Changelog {
	c.cfg = cfg
	return c
}

// WithConfigFn config the object
func (c *Changelog) WithConfigFn(fn func(cfg *Config)) *Changelog {
	fn(c.cfg)
	return c
}

// SetLogText by git log
func (c *Changelog) SetLogText(gitLogOut string) {
	c.logText = gitLogOut
}

// FetchGitLog fetch log data by git log
func (c *Changelog) FetchGitLog(sha1, sha2 string, moreArgs ...string) *Changelog {
	logCmd := gitw.Log("--reverse").
		Argf("--pretty=format:\"%s\"", c.cfg.LogFormat)

	if c.cfg.Verbose {
		logCmd.OnBeforeExec(gitw.PrintCmdline)
	}

	// add custom args. eg: "--no-merges"
	logCmd.AddArgs(moreArgs)

	// logCmd.Argf("%s...%s", "v0.1.0", "HEAD")
	if sha1 != "" && sha2 != "" {
		logCmd.Argf("%s...%s", sha1, sha2)
	}

	c.SetLogText(logCmd.SafeOutput())
	return c
}

// prepare something
func (c *Changelog) prepare() {
	if c.Formatter == nil {
		c.Formatter = c.cfg.CreateFormatter()
	}

	c.ItemFilters = c.cfg.CreateFilters()
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
	c.prepare()

	str := strings.TrimSpace(c.logText)
	if str == "" {
		return ErrEmptyLogText
	}

	// ensure parser exists
	if c.LineParser == nil {
		c.LineParser = BuiltInParser
	}

	parser := c.LineParser
	msgIDMap := make(map[string]int)

	for _, line := range strings.Split(str, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		line = strings.Trim(line, "\"' ")
		if line == "" {
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
		if c.cfg.RmRepeat {
			msgID := strutil.Md5(li.Msg)
			if _, ok := msgIDMap[msgID]; ok {
				continue
			}

			msgIDMap[msgID] = 1
		}

		c.logItems = append(c.logItems, li)
	}

	return
}

func (c *Changelog) applyItemFilters(li *LogItem) bool {
	for _, filter := range c.ItemFilters {
		if !filter.Handle(li) {
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
	if c.cfg.Title != "" {
		outLines = append(outLines, c.cfg.Title)
	}

	// use sorted names for-each
	for _, grpName := range c.cfg.Names {
		list := c.formatted[grpName]
		if len(list) == 0 {
			continue
		}

		// if only one group, not render group name.
		if groupCount > 1 {
			outLines = append(outLines, c.cfg.GroupPrefix+grpName+c.cfg.GroupSuffix)
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
	c.formatted = make(map[string][]string, len(c.cfg.Names))

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

// WriteTo changelog to the writer
func (c *Changelog) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, c.changelog)
	return int64(n), err
}

// Changelog get generated change log string
func (c *Changelog) Changelog() string {
	return c.changelog
}

// Formatted get formatted change log line
// func (c *Changelog) Formatted() []string {
// 	return c.formatted
// }

// Config get
func (c *Changelog) Config() *Config {
	return c.cfg
}

// LogCount get
func (c *Changelog) LogCount() int {
	return c.logCount
}
