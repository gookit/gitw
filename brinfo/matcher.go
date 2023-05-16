package brinfo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gookit/goutil/strutil"
)

// BranchMatcher interface
type BranchMatcher interface {
	fmt.Stringer
	// Match branch name, no remote prefix
	Match(branch string) bool
}

// ---------- matcher implement ----------

// ContainsMatch handle contains matching
type ContainsMatch struct {
	pattern string
}

// NewContainsMatch create a new contains matcher
func NewContainsMatch(pattern string) BranchMatcher {
	return &ContainsMatch{pattern: pattern}
}

// Match branch name by contains pattern
func (c *ContainsMatch) Match(branch string) bool {
	return strings.Contains(branch, c.pattern)
}

// String get string
func (c *ContainsMatch) String() string {
	return "contains: " + c.pattern
}

// PrefixMatch handle prefix matching
type PrefixMatch struct {
	pattern string
}

// NewPrefixMatch create a new prefix matcher
func NewPrefixMatch(pattern string) BranchMatcher {
	return &PrefixMatch{pattern: pattern}
}

// Match branch name by prefix pattern
func (p *PrefixMatch) Match(branch string) bool {
	return strings.HasPrefix(branch, p.pattern)
}

// String get string
func (p *PrefixMatch) String() string {
	return "prefix: " + p.pattern
}

// SuffixMatch handle suffix matching
type SuffixMatch struct {
	pattern string
}

// NewSuffixMatch create a new suffix matcher
func NewSuffixMatch(pattern string) BranchMatcher {
	return &SuffixMatch{pattern: pattern}
}

// Match branch name by suffix pattern
func (s *SuffixMatch) Match(branch string) bool {
	return strings.HasSuffix(branch, s.pattern)
}

// String get string
func (s *SuffixMatch) String() string {
	return "suffix: " + s.pattern
}

// GlobMatch handle glob matching
type GlobMatch struct {
	pattern string
}

// NewGlobMatch create a new glob matcher
func NewGlobMatch(pattern string) BranchMatcher {
	return &GlobMatch{pattern: pattern}
}

// Match branch name by glob pattern
func (g *GlobMatch) Match(branch string) bool {
	return strutil.GlobMatch(g.pattern, branch)
}

// String get string
func (g *GlobMatch) String() string {
	return "glob: " + g.pattern
}

// RegexMatch handle regex matching
type RegexMatch struct {
	pattern string
	regex   *regexp.Regexp
}

// NewRegexMatch create a new regex matcher
func NewRegexMatch(pattern string) BranchMatcher {
	return &RegexMatch{
		pattern: pattern,
		regex:   regexp.MustCompile(pattern),
	}
}

// Match branch name by regex pattern
func (r *RegexMatch) Match(branch string) bool {
	return r.regex.MatchString(branch)
}

// String get string
func (r *RegexMatch) String() string {
	return "regex: " + r.pattern
}

// NewMatcher create a branch matcher by type and pattern
func NewMatcher(pattern string, typ ...string) BranchMatcher {
	var typName string
	if len(typ) > 0 {
		typName = typ[0]
	} else if strings.Contains(pattern, ":") {
		typName, pattern = strutil.TrimCut(pattern, ":")
	}

	switch typName {
	case "contains", "contain", "has":
		return NewContainsMatch(pattern)
	case "prefix", "start", "pfx":
		return NewPrefixMatch(pattern)
	case "suffix", "end", "sfx":
		return NewSuffixMatch(pattern)
	case "regex", "regexp", "reg", "re":
		return NewRegexMatch(pattern)
	case "glob", "pattern", "pat":
		return NewGlobMatch(pattern)
	default: // default is glob mode.
		return NewGlobMatch(pattern)
	}
}

// NewBranchMatcher create a new branch matcher
func NewBranchMatcher(pattern string, regex bool) BranchMatcher {
	if regex {
		return NewRegexMatch(pattern)
	}
	return NewGlobMatch(pattern)
}

// ---------- multi-matcher wrapper ----------

const (
	// MatchAny match any one as success(default)
	MatchAny = iota
	// MatchAll match all as success
	MatchAll
)

// MultiMatcher match branch name by multi matcher
type MultiMatcher struct {
	// match mode. default is MatchAny
	mode uint8
	// matchers list
	matchers []BranchMatcher
}

// NewMulti create a multi matcher by matchers
func NewMulti(ms ...BranchMatcher) *MultiMatcher {
	return &MultiMatcher{matchers: ms}
}

// QuickMulti quick create a multi matcher by type and patterns
//
// Usage:
//
//	m := QuickMulti("contains:feat", "prefix:fix", "suffix:bug")
//	m := QuickMulti("contains:feat", "prefix:fix", "suffix:bug").WithMode(MatchAll)
func QuickMulti(typWithPatterns ...string) *MultiMatcher {
	matchers := make([]BranchMatcher, len(typWithPatterns))

	for i, typWithPattern := range typWithPatterns {
		matchers[i] = NewMatcher(typWithPattern)
	}
	return NewMulti(matchers...)
}

// WithMode set mode
func (m *MultiMatcher) WithMode(mode uint8) *MultiMatcher {
	m.mode = mode
	return m
}

// Len of multi matcher
func (m *MultiMatcher) Len() int {
	return len(m.matchers)
}

// IsEmpty check
func (m *MultiMatcher) IsEmpty() bool {
	return len(m.matchers) == 0
}

// Add matcher to multi matcher
func (m *MultiMatcher) Add(ms ...BranchMatcher) {
	m.matchers = append(m.matchers, ms...)
}

// Match branch name by multi matcher
func (m *MultiMatcher) Match(branch string) bool {
	// match one
	if m.mode == MatchAny {
		for _, matcher := range m.matchers {
			if matcher.Match(branch) {
				return true
			}
		}
		return false
	}

	// match all
	for _, matcher := range m.matchers {
		if !matcher.Match(branch) {
			return false
		}
	}
	return true
}

// String get string
func (m *MultiMatcher) String() string {
	ss := make([]string, len(m.matchers))
	for i, matcher := range m.matchers {
		ss[i] = matcher.String()
	}
	return "multi: " + strings.Join(ss, ", ")
}
