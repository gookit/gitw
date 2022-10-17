package gitw_test

import (
	"testing"

	"github.com/gookit/gitw"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/testutil/assert"
)

func TestTags(t *testing.T) {
	ts, err := gitw.Tags()
	assert.NoErr(t, err)
	dump.P(ts)

	ts, err = gitw.Tags("-n", "--sort=-version:refname")
	assert.NoErr(t, err)
	dump.P(ts)
}
