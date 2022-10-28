package gitw

import (
	"regexp"
	"strings"

	"github.com/gookit/goutil/strutil"
)

// StatusPattern string. eg: master...origin/master
const StatusPattern = `^([\w-]+)...([\w-]+)/(\w[\w/-]+)$`

var statusRegex = regexp.MustCompile(StatusPattern)

// StatusInfo struct
//
// by run: git status -bs -u
type StatusInfo struct {
	// Branch current branch name.
	Branch string
	// UpRemote current upstream remote name.
	UpRemote string
	// UpBranch current upstream remote branch name.
	UpBranch string

	fileNum int

	// Deleted files
	Deleted []string
	// Renamed files, contains RM(rename and modify) files
	Renamed []string
	// Modified files
	Modified []string
	// Unstacked new created files.
	Unstacked []string
}

// NewStatusInfo from string.
func NewStatusInfo(str string) *StatusInfo {
	si := &StatusInfo{}
	return si.FromString(str)
}

// FromString parse and load info
func (si *StatusInfo) FromString(str string) *StatusInfo {
	return si.FromLines(strings.Split(str, "\n"))
}

// FromLines parse and load info
func (si *StatusInfo) FromLines(lines []string) *StatusInfo {
	for _, line := range lines {
		line = strings.Trim(line, " \t")
		if len(line) == 0 {
			continue
		}

		// files
		mark, value := strutil.MustCut(line, " ")
		switch mark {
		case "##":
			ss := statusRegex.FindStringSubmatch(value)
			if len(ss) > 1 {
				si.Branch, si.UpRemote, si.UpBranch = ss[1], ss[2], ss[3]
			}
		case "D":
			si.fileNum++
			si.Deleted = append(si.Deleted, value)
		case "R":
			si.fileNum++
			si.Renamed = append(si.Renamed, value)
		case "M":
			si.fileNum++
			si.Modified = append(si.Modified, value)
		case "RM": // rename and modify
			si.fileNum++
			si.Renamed = append(si.Renamed, value)
		case "??":
			si.fileNum++
			si.Unstacked = append(si.Unstacked, value)
		}
	}
	return si
}

// FileNum in git status
func (si *StatusInfo) FileNum() int {
	return si.fileNum
}

// IsCleaned status in workspace
func (si *StatusInfo) IsCleaned() bool {
	return si.fileNum == 0
}
