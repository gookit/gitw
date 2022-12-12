package chlog_test

import (
	"testing"

	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil/testutil/assert"
)

func TestRuleMatcher_Match(t *testing.T) {
	line := ":sparkles: feat(dump): dump 支持 []byte 作为字符串打印和新增更多新选项"
	m := chlog.DefaultMatcher.Match(line)
	assert.Equal(t, "Feature", m)

	line = ":necktie: up(str): 更新字节工具方法并添加新的哈希工具方法"
	m = chlog.DefaultMatcher.Match(line)
	assert.Equal(t, "Update", m)
}
