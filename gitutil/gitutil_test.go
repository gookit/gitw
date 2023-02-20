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
		args      args
		wantGroup string
		wantName  string
		wantErr   bool
	}{
		{
			args:      args{"my/repo"},
			wantGroup: "my",
			wantName:  "repo",
		},
		{
			args:      args{"my/repo-01"},
			wantGroup: "my",
			wantName:  "repo-01",
		},
	}
	for _, tt := range tests {
		gotGroup, gotName, err := gitutil.SplitPath(tt.args.repoPath)

		assert.Eq(t, tt.wantGroup, gotGroup)
		assert.Eq(t, tt.wantName, gotName)

		if tt.wantErr {
			assert.Err(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestIsRepoPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{
			path: "my/repo",
			want: true,
		},
		{
			path: "my/repo-01",
			want: true,
		},
		{
			path: "my/repo/sub01",
			want: false,
		},
		{
			path: "my-repo-01",
			want: false,
		},
	}

	for _, tt := range tests {
		assert.Eq(t, tt.want, gitutil.IsRepoPath(tt.path))
	}
}

func TestIsFullURL(t *testing.T) {
	tests := []struct {
		args string
		want bool
	}{
		{"inhere/gitw", false},
		{"github.com/inhere/gitw", false},
		{"https://github.com/inhere/gitw", true},
		{"http://github.com/inhere/gitw", true},
		{"git@github.com:inhere/gitw", true},
		{"ssh://git@github.com:inhere/gitw", true},
	}

	for _, tt := range tests {
		assert.Eq(t, tt.want, gitutil.IsFullURL(tt.args))
	}
}
