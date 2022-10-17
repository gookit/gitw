package chlog_test

import (
	"testing"

	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil/testutil/assert"
)

func TestKeywordsFilter(t *testing.T) {
	fl := chlog.KeywordsFilter([]string{"format code", "action test"}, true)

	li := &chlog.LogItem{Msg: "chore: up some for action tests 3"}
	assert.False(t, fl(li))

	li = &chlog.LogItem{Msg: "chore: fix gh action script error"}
	assert.True(t, fl(li))
}
