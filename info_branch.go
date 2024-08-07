package gitw

import (
	"strings"

	"github.com/gookit/gitw/brinfo"
	"github.com/gookit/goutil/basefn"
	"github.com/gookit/goutil/strutil"
)

// RemotePfxOnBranch prefix keywords
const RemotePfxOnBranch = "remotes/"

// BranchInfo for a git branch
type BranchInfo struct {
	// Current active branch
	Current bool
	// Name The full branch name. eg: fea_xx, remotes/origin/fea_xx
	Name string
	// Short only branch name. local branch is equals Name. eg: fea_xx
	Short string
	// Hash commit hash ID.
	Hash string
	// HashMsg commit hash message.
	HashMsg string
	// Alias name
	Alias string
	// Remote name. local branch is empty. eg: origin
	Remote string
}

// NewBranchInfo from branch line text
func NewBranchInfo(line string) (*BranchInfo, error) {
	return ParseBranchLine(line, isVerboseBranchLine(line))
}

// IsValid branch check
func (b *BranchInfo) IsValid() bool {
	return b.Name != ""
}

// IsRemoted branch check
func (b *BranchInfo) IsRemoted() bool {
	return strings.HasPrefix(b.Name, RemotePfxOnBranch)
}

// SetName for branch and parse
func (b *BranchInfo) SetName(name string) {
	b.Name = name
	b.ParseName()
}

// ParseName for get remote and short name.
func (b *BranchInfo) ParseName() *BranchInfo {
	b.Short = b.Name

	if b.IsRemoted() {
		// remove prefix "remotes"
		b.Remote, b.Short = strutil.QuietCut(b.Name[8:], "/")
	}
	return b
}

// branch types
const (
	BranchLocal  = "local"
	BranchRemote = "remote"
)

// BranchInfos for a git repo
type BranchInfos struct {
	parsed bool
	// last parse err
	err error
	// raw branch lines by git branch
	brLines []string

	current *BranchInfo
	// local branch list
	locales []*BranchInfo
	// all remote branch list
	remotes []*BranchInfo
}

// EmptyBranchInfos instance
func EmptyBranchInfos() *BranchInfos {
	return &BranchInfos{
		// locales: make(map[string]*BranchInfo),
		// remotes: make(map[string]*BranchInfo),
	}
}

// NewBranchInfos create
func NewBranchInfos(gitOut string) *BranchInfos {
	return &BranchInfos{
		brLines: strings.Split(strings.TrimSpace(gitOut), "\n"),
		// locales: make(map[string]*BranchInfo),
		// remotes: make(map[string]*BranchInfo),
	}
}

// Parse given branch lines
func (bs *BranchInfos) Parse() *BranchInfos {
	if len(bs.brLines) == 0 {
		return bs
	}

	if bs.parsed {
		return bs
	}

	bs.parsed = true
	verbose := isVerboseBranchLine(bs.brLines[0])

	for _, line := range bs.brLines {
		if len(line) == 0 {
			continue
		}

		// parse line
		info, err := ParseBranchLine(line, verbose)
		if err != nil {
			bs.err = err
			continue
		}

		// collect
		if info.IsRemoted() {
			bs.remotes = append(bs.remotes, info)
		} else {
			bs.locales = append(bs.locales, info)
			if info.Current {
				bs.current = info
			}
		}
	}

	return bs
}

// HasLocal branch check
func (bs *BranchInfos) HasLocal(branch string) bool {
	return bs.GetByName(branch) != nil
}

// HasRemote branch check
func (bs *BranchInfos) HasRemote(branch, remote string) bool {
	return bs.GetByName(branch, remote) != nil
}

// IsExists branch check
func (bs *BranchInfos) IsExists(branch string, remote ...string) bool {
	return bs.GetByName(branch, remote...) != nil
}

// GetByName find branch by name
func (bs *BranchInfos) GetByName(branch string, remote ...string) *BranchInfo {
	if len(remote) > 0 && remote[0] != "" {
		for _, info := range bs.remotes {
			if info.Remote == remote[0] && branch == info.Short {
				return info
			}
		}
		return nil
	}

	for _, info := range bs.locales {
		if branch == info.Short {
			return info
		}
	}
	return nil
}

// flags for search branches
const (
	BrSearchLocal  uint8 = 1
	BrSearchRemote uint8 = 1 << 1
	BrSearchAll          = BrSearchLocal | BrSearchRemote
)

// Search branches by name.
//
// TIP: recommend use `SearchV2()` for search branches.
//
// Usage:
//
//	Search("fea", BrSearchLocal)
//	Search("fea", BrSearchAll)
//	// search on remotes
//	Search("fea", BrSearchRemote)
//	// search on remotes and remote name must be equals "origin"
//	Search("origin:fea", BrSearchRemote)
func (bs *BranchInfos) Search(name string, flag uint8) []*BranchInfo {
	var list []*BranchInfo

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return list
	}

	var remote string
	// "remote name" - search on the remote
	if strings.Contains(name, ":") {
		remote, name = strutil.MustCut(name, ":")
	}

	if remote == "" && flag&BrSearchLocal == BrSearchLocal {
		for _, info := range bs.locales {
			if strings.Contains(info.Short, name) {
				list = append(list, info)
			}
		}
	}

	if flag&BrSearchRemote == BrSearchRemote {
		for _, info := range bs.remotes {
			if strings.Contains(info.Short, name) {
				if remote == "" {
					list = append(list, info)
				} else if remote == info.Remote {
					list = append(list, info)
				}
			}
		}
	}

	return list
}

// SearchOpt for search branches
type SearchOpt struct {
	// Flag search flag, default is BrSearchLocal.
	Flag  uint8
	Limit int
	// Remote name, on which remote to search.
	Remote string
	// Before search callback, return false to skip.
	Before func(bi *BranchInfo) bool
}

// SearchV2 search branches by matcher and hook func.
//
// Usage:
//
//	SearchV2(brinfo.NewContainsMatch("fea"), &SearchOpt{})
//	// use multi matcher
//	SearchV2(brinfo.QuickMulti("start:fea","glob:fea*"), &SearchOpt{})
func (bs *BranchInfos) SearchV2(matcher brinfo.BranchMatcher, opt *SearchOpt) []*BranchInfo {
	if opt == nil {
		opt = &SearchOpt{Limit: 10}
	}
	if opt.Flag == 0 {
		opt.Flag = basefn.OrValue(opt.Remote == "", BrSearchLocal, BrSearchRemote)
	}

	var list []*BranchInfo

	if opt.Flag&BrSearchLocal == BrSearchLocal {
		for _, info := range bs.locales {
			if opt.Before != nil && !opt.Before(info) {
				continue
			}

			if matcher.Match(info.Short) {
				list = append(list, info)
				if opt.Limit > 0 && len(list) >= opt.Limit {
					break
				}
			}
		}
	}

	if opt.Flag&BrSearchRemote == BrSearchRemote {
		for _, info := range bs.remotes {
			if opt.Remote != "" && opt.Remote != info.Remote {
				continue
			}

			if opt.Before != nil && !opt.Before(info) {
				continue
			}

			if matcher.Match(info.Short) {
				list = append(list, info)
				if opt.Limit > 0 && len(list) >= opt.Limit {
					break
				}
			}
		}
	}

	return list
}

// BrLines get
func (bs *BranchInfos) BrLines() []string {
	return bs.brLines
}

// LastErr get
func (bs *BranchInfos) LastErr() error {
	return bs.err
}

// SetBrLines for parse.
func (bs *BranchInfos) SetBrLines(brLines []string) {
	bs.brLines = brLines
}

// Current branch
func (bs *BranchInfos) Current() *BranchInfo {
	return bs.current
}

// Locales branches
func (bs *BranchInfos) Locales() []*BranchInfo {
	return bs.locales
}

// Remotes branch infos get
//
// if remote="", will return all remote branches
func (bs *BranchInfos) Remotes(remote string) []*BranchInfo {
	if remote == "" {
		return bs.remotes
	}

	ls := make([]*BranchInfo, 0)
	for _, info := range bs.remotes {
		if info.Remote == remote {
			ls = append(ls, info)
		}
	}
	return ls
}

// All branches list
func (bs *BranchInfos) All() []*BranchInfo {
	ls := make([]*BranchInfo, 0, len(bs.locales)+len(bs.remotes))
	for _, info := range bs.locales {
		ls = append(ls, info)
	}

	for _, info := range bs.remotes {
		ls = append(ls, info)
	}
	return ls
}
