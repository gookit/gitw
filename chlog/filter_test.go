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

func TestWordsLenFilter(t *testing.T) {
	fl := chlog.WordsLenFilter(3)

	// 英文测试：4个单词，应该通过
	li := &chlog.LogItem{Msg: "fix: update config file"}
	assert.True(t, fl(li))

	// 英文测试：2个单词，应该被过滤
	li = &chlog.LogItem{Msg: "fix bug"}
	assert.False(t, fl(li))

	// 中文测试：超过3个字符，应该通过
	li = &chlog.LogItem{Msg: "修复了一个重要的配置问题"}
	assert.True(t, fl(li))

	// 中文测试：少于3个字符，应该被过滤
	li = &chlog.LogItem{Msg: "修bug"}
	assert.False(t, fl(li))

	// 中英混合测试
	li = &chlog.LogItem{Msg: "fix 修复了问题"}
	assert.True(t, fl(li))
}
