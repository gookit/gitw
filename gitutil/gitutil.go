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

// IsFullURL quick and simple check input is ssh/http URL
func IsFullURL(s string) bool {
	if IsHTTPProto(s) {
		return true
	}
	return IsSSHProto(s)
}

// IsHTTPProto check input is HTTP URL
func IsHTTPProto(s string) bool {
	return strings.HasPrefix(s, "http:") || strings.HasPrefix(s, "https:")
}

// IsSSHProto check input is ssh URL
func IsSSHProto(s string) bool {
	return strings.HasPrefix(s, "ssh:") || strings.HasPrefix(s, "git@")
}

// regex which validates that the git branch name is correct
var brNameReg = regexp.MustCompile(`^[a-zA-Z0-9]+([/_-][a-zA-Z0-9]+)*$`)

// IsBranchName validate branch name
func IsBranchName(name string) bool {
	return brNameReg.MatchString(name)
}

// FormatVersion string. eg: v1.2.0 -> 1.2.0
func FormatVersion(ver string) (string, bool) {
	ver = strings.TrimLeft(ver, "vV")
	if strutil.IsVersion(ver) {
		return ver, true
	}
	return "", false
}

// IsValidVersion check
func IsValidVersion(ver string) bool {
	ver = strings.TrimLeft(ver, "vV")
	return strutil.IsVersion(ver)
}

// NextVersion build. eg: v1.2.0 -> v1.2.1
func NextVersion(ver string) string {
	if len(ver) == 0 {
		return "v0.0.1"
	}

	ver = strings.TrimLeft(ver, "vV")
	nodes := strings.Split(ver, ".")
	if len(nodes) == 1 {
		return ver + ".0.1"
	}

	for i := len(nodes) - 1; i > 0; i-- {
		num, err := strutil.ToInt(nodes[i])
		if err != nil {
			continue
		}
		nodes[i] = strutil.SafeString(num + 1)
		break
	}

	return strings.Join(nodes, ".")
}
