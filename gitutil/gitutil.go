package gitutil

import (
	"regexp"
	"strings"

	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/strutil"
)

// SplitPath split path to group and name.
func SplitPath(repoPath string) (group, name string, err error) {
	group, name = strutil.TrimCut(repoPath, "/")

	if strutil.HasEmpty(group, name) {
		err = errorx.Raw("invalid git repo path, must be as GROUP/NAME")
	}
	return
}

var repoPathReg = regexp.MustCompile(`^[\w-]+/[\w-]+$`)

// IsRepoPath string. should match GROUP/NAME
func IsRepoPath(path string) bool {
	return repoPathReg.MatchString(path)
}

// ParseCommitTopic for git commit message
func ParseCommitTopic(msg string) []string {
	return nil // TODO
}

// ResolveGhURL string
func ResolveGhURL(s string) (string, bool) {
	if strings.HasPrefix(s, githubHost) {
		return "https://" + s, true
	}
	return s, false
}

// IsFullURL quick and simple check input is URL string
func IsFullURL(s string) bool {
	if strings.HasPrefix(s, "http:") || strings.HasPrefix(s, "https:") {
		return true
	}

	if strings.HasPrefix(s, "ssh:") || strings.HasPrefix(s, "git@") {
		return true
	}
	return false
}
