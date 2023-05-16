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
		{"my/repo", true},
		{"my/repo-01", true},
		{"my/repo/sub01", false},
		{"my-repo-01", false},
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

func TestIsBranchName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"master", true},
		{"dev", true},
		{"dev-01", true},
		{"dev_01", true},
		{"dev/01", true},
		{"dev-01/02", true},
		{"dev_01/02", true},
		{"dev/01-02", true},
		{"dev/01_02", true},
		{"dev/01-02/03", true},
		{"-master", false},
		{"dev-", false},
	}

	for _, tt := range tests {
		assert.Eq(t, tt.want, gitutil.IsBranchName(tt.name))
	}
}
