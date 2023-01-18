package gitutil_test

import (
	"testing"

	"github.com/gookit/gitw/gitutil"
	"github.com/gookit/goutil/testutil/assert"
)

func TestSplitPath(t *testing.T) {
	type args struct {
		repoPath string
	}
	tests := []struct {
		name      string
		args      args
		wantGroup string
		wantName  string
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		gotGroup, gotName, err := gitutil.SplitPath(tt.args.repoPath)

		assert.Eq(t, tt.wantGroup, gotGroup)
		assert.Eq(t, tt.wantName, gotName)
		assert.Nil(t, err)
	}
}
