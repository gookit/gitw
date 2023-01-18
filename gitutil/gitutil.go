package gitutil

import (
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
