package chlog

import (
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
)

// Config struct
type Config struct {
	// Title string for formatted text. eg: "## Change Log"
	Title string `json:"title" yaml:"title"`
	// RepoURL repo URL address
	RepoURL string `json:"repo_url" yaml:"repo_url"`
	// Style name. allow: simple, markdown, ghr
	Style string `json:"style" yaml:"style"`
	// LogFormat built-in log format string.
	//
	// use on the `git log --pretty="format:%H"`.
	//
	// see consts LogFmt*, eg: LogFmtHs
	LogFormat string `json:"log_format" yaml:"log_format"`
	// GroupPrefix string. eg: '### '
	GroupPrefix string `yaml:"group_prefix"`
	// GroupPrefix string.
	GroupSuffix string `yaml:"group_suffix"`
	// NoGroup Not output group name line.
	NoGroup bool `yaml:"no_group"`
	// RmRepeat remove repeated log by message
	RmRepeat bool `json:"rm_repeat" yaml:"rm_repeat"`
	// Verbose show more information
	Verbose bool `json:"verbose" yaml:"verbose"`
	// Names define group names and sort
	Names []string `json:"names" yaml:"names"`
	// Rules for match group
	Rules []Rule `json:"rules" yaml:"rules"`
	// Filters for filtering
	Filters []maputil.Data `json:"filters" yaml:"filters"`
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

// Create Changelog
func (c *Config) Create() *Changelog {
	cl := NewWithConfig(c)

	return cl
}

// CreateFilters for Changelog
func (c *Config) CreateFilters() []ItemFilter {
	if len(c.Filters) == 0 {
		return nil
	}

	fls := make([]ItemFilter, 0, len(c.Filters))
	/*
	  - name: keywords
	    keywords: ['format code']
	    exclude: true
	*/
	for _, rule := range c.Filters {
		name := rule["name"]
		if name == "" {
			continue
		}

		switch name {
		case FilterMsgLen:
			ln := rule.Int("min_len")
			if ln <= 0 {
				continue
			}

			fls = append(fls, MsgLenFilter(ln))
		case FilterWordsLen:
			ln := rule.Int("min_len")
			if ln <= 0 {
				continue
			}

			fls = append(fls, WordsLenFilter(ln))
		case FilterKeyword:
			str := rule.Str("keyword")
			if len(str) <= 0 {
				continue
			}

			fls = append(fls, KeywordFilter(str, rule.Bool("exclude")))
		case FilterKeywords:
			str := rule.Str("keywords")
			ss := strutil.Split(str, ",")
			if len(ss) <= 0 {
				continue
			}

			fls = append(fls, KeywordsFilter(ss, rule.Bool("exclude")))
		}
	}

	return fls
}

// CreateFormatter for Changelog
func (c *Config) CreateFormatter() Formatter {
	sf := &SimpleFormatter{}
	ns := c.Names

	matcher := NewDefaultMatcher()
	if len(c.Rules) > 0 {
		if len(c.Names) == 0 {
			ns = maputil.Keys(c.Rules)
		}

		matcher = &RuleMatcher{Rules: c.Rules}
	}

	if len(ns) > 0 {
		matcher.Names = ns
	}

	c.Names = matcher.Names
	sf.GroupMatch = matcher

	switch c.Style {
	case FormatterMarkdown, "mkdown", "mkDown", "mkd", "md":
		return &MarkdownFormatter{
			RepoURL:         c.RepoURL,
			SimpleFormatter: *sf,
		}
	case FormatterGhRelease, "gh-release", "ghRelease", "gh":
		f := &GHReleaseFormatter{}
		f.RepoURL = c.RepoURL
		f.SimpleFormatter = *sf
		return f
	default:
		return sf
	}
}
