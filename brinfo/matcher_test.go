package brinfo_test

import (
	"testing"

	"github.com/gookit/gitw/brinfo"
	"github.com/gookit/goutil/testutil/assert"
)

func TestGlobMatch_Match(t *testing.T) {
	m := brinfo.NewMatcher("fea*")
	assert.True(t, m.Match("fea-1"))
	assert.True(t, m.Match("fea_dev"))
	assert.False(t, m.Match("fix_2"))

	m = brinfo.NewMatcher("fix", "prefix")
	assert.False(t, m.Match("fea-1"))
	assert.False(t, m.Match("fea_dev"))
	assert.True(t, m.Match("fix_2"))

	m = brinfo.NewMatcher(`reg:^ma\w+`)
	assert.True(t, m.Match("main"))
	assert.True(t, m.Match("master"))
	assert.False(t, m.Match("x-main"))

	m = brinfo.NewGlobMatch("*new*")
	assert.True(t, m.Match("fea/new_br001"))
}

func TestNewMulti(t *testing.T) {
	m := brinfo.NewMulti(
		brinfo.NewGlobMatch("fea*"),
		brinfo.NewPrefixMatch("fix"),
		brinfo.NewContainsMatch("main"),
		brinfo.NewSuffixMatch("-dev"),
	)

	assert.True(t, m.Match("fea-1"))
	assert.True(t, m.Match("fea_dev"))
	assert.True(t, m.Match("fix_2"))
	assert.True(t, m.Match("main"))
	assert.True(t, m.Match("some/fea-dev"))

	m = brinfo.QuickMulti("start:fix", "end:-dev")
	assert.True(t, m.Match("fix_23"))
	assert.True(t, m.Match("fea23-dev"))

	m.WithMode(brinfo.MatchAll)

	assert.False(t, m.Match("fix_23"))
	assert.False(t, m.Match("fea23-dev"))
	assert.True(t, m.Match("fix-23-dev"))
}
