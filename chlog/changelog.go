package chlog

import (
	"errors"
	"strings"

	"github.com/gookit/gitw"
	"github.com/gookit/goutil/maputil"
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

// Config struct
type Config struct {
	// Title string for formatted text. eg: "## Change Log"
	Title string `json:"title" yaml:"title"`
	// RepoURL repo URL address
	RepoURL string `json:"repo_url" yaml:"repo_url"`
	// LogFormat string on call git log.
	LogFormat string `json:"log_format" yaml:"log_format"`
	// GroupPrefix string. eg: '### '
	GroupPrefix string `yaml:"group_prefix"`
	// GroupPrefix string.
	GroupSuffix string `yaml:"group_suffix"`
	// NoGroup Not output group name line.
	NoGroup bool `yaml:"no_group"`
	// RmRepeat remove repeated log by message
	RmRepeat bool `yaml:"rm_repeat"`
	// Names define group names and sort
	Names []string `json:"names" yaml:"names"`
	// Rules for match group
	Rules []Rule `json:"rules" yaml:"rules"`
	// Filters for filtering
	Filters []maputil.SMap `json:"filters" yaml:"filters"`
}

// NewDefaultConfig instance
func NewDefaultConfig() *Config {
	return &Config{
		Title:       "## Change Log",
		RmRepeat:    true,
		LogFormat:   LogFmtHs,
		GroupPrefix: "\n### ",
		GroupSuffix: "\n",
	}
}

// Changelog struct
type Changelog struct {
	cfg *Config

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
	// LogFormat built-in log format string on the `git log --pretty="format:%H"`.
	// see consts LogFmt*
	LogFormat string
	// Title string for formatted text. eg: "## Change Log"
	Title string
	// RepoURL repo URL address
	RepoURL string
	// GroupPrefix string. eg: '### '
	GroupPrefix string
	// GroupPrefix string.
	GroupSuffix string
	// NoGroup Not output group name line.
	NoGroup bool
	// RmRepeat remove repeated log by message
	RmRepeat bool
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
	logCmd := gitw.New("log", "--reverse")
	logCmd.Addf("--pretty=format:\"%s\"", c.LogFormat)
	// logCmd.Add("--no-merges")
	logCmd.Add(moreArgs...) // add custom args

	// logCmd.Addf("%s...%s", "v0.1.0", "HEAD")
	if sha1 != "" && sha2 != "" {
		logCmd.Addf("%s...%s", sha1, sha2)
	}

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
		if c.RmRepeat {
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
	if c.Title != "" {
		outLines = append(outLines, c.Title)
	}

	for grpName, list := range c.formatted {
		// if only one group, not render group name.
		if groupCount > 1 {
			outLines = append(outLines, c.GroupPrefix+grpName+c.GroupSuffix)
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
