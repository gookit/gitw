package gitw

import (
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/sysutil/cmdr"
)

const (
	cacheRemoteNames   = "rmtNames"
	cacheRemoteInfos   = "rmtInfos"
	cacheLastCommitID  = "lastCID"
	cacheCurrentBranch = "curBranch"
	cacheMaxTagVersion = "maxVersion"
	cacheUpstreamPath  = "upstreamTo"
)

// RepoConfig struct
type RepoConfig struct {
	// DefaultBranch name, default is DefaultBranchName
	DefaultBranch string
	// DefaultRemote name, default is DefaultRemoteName
	DefaultRemote string
}

func newDefaultCfg() *RepoConfig {
	return &RepoConfig{
		DefaultBranch: DefaultBranchName,
		DefaultRemote: DefaultRemoteName,
	}
}

// Repo struct
type Repo struct {
	gw *GitWrap
	// the repo dir
	dir string
	// save last error
	err error
	// config
	cfg *RepoConfig

	// status info
	statusInfo *StatusInfo

	// branch infos for the repo
	branchInfos *BranchInfos

	// remoteNames
	remoteNames []string
	// remoteInfosMp
	//
	// Example:
	// 	{origin: {fetch: remote info, push: remote info}}
	remoteInfosMp map[string]RemoteInfos

	// cache some information of the repo
	cache maputil.Data
}

// NewRepo create Repo object
func NewRepo(dir string) *Repo {
	return &Repo{
		dir: dir,
		cfg: newDefaultCfg(),
		// init gw
		gw: NewWithWorkdir(dir),
		// cache some information
		cache: make(maputil.Data, 8),
	}
}

// WithFn new repo self config func
func (r *Repo) WithFn(fn func(r *Repo)) *Repo {
	fn(r)
	return r
}

// WithConfig new repo config
func (r *Repo) WithConfig(cfg *RepoConfig) *Repo {
	r.cfg = cfg
	return r
}

// WithConfigFn new repo config func
func (r *Repo) WithConfigFn(fn func(cfg *RepoConfig)) *Repo {
	fn(r.cfg)
	return r
}

// PrintCmdOnExec settings.
func (r *Repo) PrintCmdOnExec() *Repo {
	r.gw.BeforeExec = PrintCmdline
	return r
}

// SetDryRun settings.
func (r *Repo) SetDryRun(dr bool) *Repo {
	r.gw.DryRun = dr
	return r
}

// Init run git init for the repo dir.
func (r *Repo) Init() error {
	return r.gw.Init().Run()
}

// IsInited is init git repo dir
func (r *Repo) IsInited() bool {
	return r.gw.IsGitRepo()
}

// Info get repo information
func (r *Repo) Info() *RepoInfo {
	ri := &RepoInfo{
		Dir:  r.dir,
		Name: fsutil.Name(r.dir),
		// more
		Branch:   r.CurBranchName(),
		Version:  r.LargestTag(),
		LastHash: r.LastAbbrevID(),
		Upstream: r.UpstreamPath(),
	}

	rt := r.loadRemoteInfos().FirstRemoteInfo()
	if rt == nil {
		return ri
	}

	ri.Name = rt.Repo
	ri.Path = rt.Path()
	ri.URL = rt.URLOrBuild()

	remotes := make(map[string]string)
	for name, infos := range r.remoteInfosMp {
		remotes[name] = infos.FetchInfo().URL
	}

	ri.Remotes = remotes
	return ri
}

// FetchAll fetch all remote branches
func (r *Repo) FetchAll(args ...string) error {
	return r.gw.Cmd("fetch", "--all").AddArgs(args).Run()
}

// -------------------------------------------------
// repo tags
// -------------------------------------------------

// ShaHead keywords
const ShaHead = "HEAD"

// some special keywords for match tag
const (
	TagLast = "last"
	TagPrev = "prev"
	TagHead = "head"
)

// enum type value constants for fetch tags
const (
	RefNameTagType int = iota
	CreatorDateTagType
	DescribeTagType
)

// AutoMatchTag by given sha or tag name
func (r *Repo) AutoMatchTag(sha string) string {
	return r.AutoMatchTagByType(sha, RefNameTagType)
}

// AutoMatchTagByType by given sha or tag name.
func (r *Repo) AutoMatchTagByType(sha string, tagType int) string {
	switch strings.ToLower(sha) {
	case TagLast:
		return r.LargestTagByTagType(tagType)
	case TagPrev:
		return r.TagSecondMaxByTagType(tagType)
	case TagHead:
		return ShaHead
	default:
		return sha
	}
}

// MaxTag get max tag version of the repo
func (r *Repo) MaxTag() string {
	return r.LargestTag()
}

// LargestTag get max tag version of the repo
func (r *Repo) LargestTag() string {
	tagVer := r.cache.Str(cacheMaxTagVersion)
	if len(tagVer) > 0 {
		return tagVer
	}

	tags := r.TagsSortedByRefName()

	if len(tags) > 0 {
		r.cache.Set(cacheMaxTagVersion, tags[0])
		return tags[0]
	}
	return ""
}

// LargestTagByTagType get max tag version of the repo by tag_type
func (r *Repo) LargestTagByTagType(tagType int) string {
	tagVer := r.cache.Str(cacheMaxTagVersion)
	if len(tagVer) > 0 {
		return tagVer
	}

	tags := make([]string, 0, 2)
	switch tagType {
	case CreatorDateTagType:
		tags = append(tags, r.TagsSortedByCreatorDate()...)
	case DescribeTagType:
		tags = append(tags, r.TagByDescribe(""))
	default:
		tags = append(tags, r.TagsSortedByRefName()...)
	}

	if len(tags) > 0 {
		r.cache.Set(cacheMaxTagVersion, tags[0])
		return tags[0]
	}
	return ""
}

// PrevMaxTag get second-largest tag of the repo
func (r *Repo) PrevMaxTag() string {
	return r.TagSecondMax()
}

// TagSecondMax get second-largest tag of the repo
func (r *Repo) TagSecondMax() string {
	tags := r.TagsSortedByRefName()

	if len(tags) > 1 {
		return tags[1]
	}
	return ""
}

// TagSecondMaxByTagType  get second-largest tag of the repo by tag_type
func (r *Repo) TagSecondMaxByTagType(tagType int) string {
	tags := make([]string, 0, 2)
	switch tagType {
	case CreatorDateTagType:
		tags = append(tags, r.TagsSortedByCreatorDate()...)
	case DescribeTagType:
		current := r.TagByDescribe("")
		if len(current) != 0 {
			tags = append(tags, current, r.TagByDescribe(current))
		} else {
			tags = append(tags, current)
		}
	default:
		tags = append(tags, r.TagsSortedByRefName()...)
	}

	if len(tags) > 1 {
		return tags[1]
	}
	return ""
}

// TagsSortedByRefName get repo tags list
func (r *Repo) TagsSortedByRefName() []string {
	str, err := r.gw.Tag("-l", "--sort=-version:refname").Output()
	if err != nil {
		r.setErr(err)
		return nil
	}

	return cmdr.OutputLines(str)
}

// TagsSortedByCreatorDate get repo tags list by creator date sort
func (r *Repo) TagsSortedByCreatorDate() []string {
	str, err := r.gw.
		Tag("-l", "--sort=-creatordate", "--format=%(refname:strip=2)").
		Output()

	if err != nil {
		r.setErr(err)
		return nil
	}
	return cmdr.OutputLines(str)
}

// TagByDescribe get tag by describe command. if current not empty, will exclude it.
func (r *Repo) TagByDescribe(current string) (ver string) {
	var err error
	if len(current) == 0 {
		ver, err = r.gw.Describe("--tags", "--abbrev=0").Output()
	} else {
		ver, err = r.gw.
			Describe("--tags", "--abbrev=0").
			Argf("tags/%s^", current).
			Output()
	}

	if err != nil {
		r.setErr(err)
		return ""
	}
	return cmdr.FirstLine(ver)
}

// Tags get repo tags list
func (r *Repo) Tags() []string {
	ss, err := r.gw.Tag("-l").OutputLines()
	if err != nil {
		r.setErr(err)
		return nil
	}

	return ss
}

// -------------------------------------------------
// repo git log
// -------------------------------------------------

// LastAbbrevID get last abbrev commit ID, len is 7
func (r *Repo) LastAbbrevID() string {
	cid := r.LastCommitID()
	if cid == "" {
		return ""
	}

	return strutil.Substr(cid, 0, 7)
}

// LastCommitID value
func (r *Repo) LastCommitID() string {
	lastCID := r.cache.Str(cacheLastCommitID)
	if len(lastCID) > 0 {
		return lastCID
	}

	// by: git log -1 --format='%H'
	lastCID, err := r.gw.Log("-1", "--format=%H").Output()
	if err != nil {
		r.setErr(err)
		return ""
	}

	r.cache.Set(cacheLastCommitID, lastCID)
	return lastCID
}

// -------------------------------------------------
// repo status
// -------------------------------------------------

// StatusInfo get status info of the repo
func (r *Repo) StatusInfo() *StatusInfo {
	if r.statusInfo == nil {
		r.statusInfo = &StatusInfo{}
		lines, err := r.gw.Status("-bs", "-u").OutputLines()
		if err != nil {
			r.setErr(err)
			return nil
		}

		r.statusInfo.FromLines(lines)
	}
	return r.statusInfo
}

// -------------------------------------------------
// repo branch
// -------------------------------------------------

func (r *Repo) HasBranch(branch string, remote ...string) bool {
	return r.loadBranchInfos().branchInfos.IsExists(branch, remote...)
}

func (r *Repo) HasRemoteBranch(branch, remote string) bool {
	return r.loadBranchInfos().branchInfos.HasRemote(branch, remote)
}

func (r *Repo) HasLocalBranch(branch string) bool {
	return r.loadBranchInfos().branchInfos.HasLocal(branch)
}

// BranchInfos get branch infos of the repo
func (r *Repo) BranchInfos() *BranchInfos {
	return r.loadBranchInfos().branchInfos
}

// ReloadBranches reload branch infos of the repo
func (r *Repo) ReloadBranches() *BranchInfos {
	r.branchInfos = nil
	return r.loadBranchInfos().branchInfos
}

// CurBranchInfo get current branch info of the repo
func (r *Repo) CurBranchInfo() *BranchInfo {
	return r.loadBranchInfos().branchInfos.Current()
}

// BranchInfo find branch info by name, if remote is empty, find local branch
func (r *Repo) BranchInfo(branch string, remote ...string) *BranchInfo {
	return r.loadBranchInfos().branchInfos.GetByName(branch, remote...)
}

// SearchBranches search branch infos by name
func (r *Repo) SearchBranches(name string, flag int) []*BranchInfo {
	return r.loadBranchInfos().branchInfos.Search(name, flag)
}

// load branch infos
func (r *Repo) loadBranchInfos() *Repo {
	// has loaded
	if r.branchInfos != nil {
		return r
	}

	str, err := r.gw.Branch("-v", "--all").Output()
	if err != nil {
		r.setErr(err)
		r.branchInfos = EmptyBranchInfos()
		return r
	}

	r.branchInfos = NewBranchInfos(str).Parse()
	return r
}

// HeadBranchName return current branch name
func (r *Repo) HeadBranchName() string { return r.CurBranchName() }

// CurBranchName return current branch name
func (r *Repo) CurBranchName() string {
	brName := r.cache.Str(cacheCurrentBranch)
	if len(brName) > 0 {
		return brName
	}

	// 	cat .git/HEAD
	// OR
	// 	git branch --show-current // on high version git
	// OR
	// 	git symbolic-ref HEAD // out: refs/heads/fea_pref
	// 	git symbolic-ref --short -q HEAD // on checkout tag, run will error
	// Or
	// 	git rev-parse --abbrev-ref -q HEAD // on init project, will error

	str := r.gw.Branch("--show-current").SafeOutput()
	if len(str) == 0 {
		str, r.err = r.gw.RevParse("--abbrev-ref", "-q", "HEAD").Output()
		if r.err != nil {
			return ""
		}
	}

	// eg: fea_pref
	brName = cmdr.FirstLine(str)
	r.cache.Set(cacheCurrentBranch, brName)
	return brName
}

// SetUpstreamTo set the branch upstream remote branch.
// If `localBranch` is empty, will use `branch` as `localBranch`
//
// CMD:
//
//	git branch --set-upstream-to=<remote>/<branch> <local_branch>
func (r *Repo) SetUpstreamTo(remote, branch string, localBranch ...string) error {
	localBr := branch
	if len(localBranch) > 0 {
		localBr = localBranch[0]
	}

	return r.gw.Cmd("branch").
		Argf("--set-upstream-to=%s/%s", remote, branch).
		AddArg(localBr).
		Run()
}

// -------------------------------------------------
// repo remote
// -------------------------------------------------

// HasRemote check
func (r *Repo) HasRemote(name string) bool {
	return arrutil.StringsHas(r.RemoteNames(), name)
}

// RemoteNames get
func (r *Repo) RemoteNames() []string {
	return r.loadRemoteInfos().remoteNames
}

// RemoteLines get like: {origin: url, other: url}
func (r *Repo) RemoteLines() map[string]string {
	remotes := make(map[string]string)
	for name, infos := range r.loadRemoteInfos().remoteInfosMp {
		remotes[name] = infos.FetchInfo().URL
	}

	return remotes
}

// UpstreamPath get current upstream remote and branch.
// Returns like: origin/main
//
// CMD:
//
//	git rev-parse --abbrev-ref @{u}
func (r *Repo) UpstreamPath() string {
	path := r.cache.Str(cacheUpstreamPath)

	// RUN: git rev-parse --abbrev-ref @{u}
	if path == "" {
		path = r.Git().RevParse("--abbrev-ref", "@{u}").SafeOutput()
		r.cache.Set(cacheUpstreamPath, strings.TrimSpace(path))
	}

	return path
}

// UpstreamRemote get current upstream remote name.
func (r *Repo) UpstreamRemote() string {
	return strutil.OrHandle(r.UpstreamPath(), func(s string) string {
		remote, _ := strutil.QuietCut(s, "/")
		return remote
	})
}

// UpstreamBranch get current upstream branch name.
func (r *Repo) UpstreamBranch() string {
	return strutil.OrHandle(r.UpstreamPath(), func(s string) string {
		_, branch := strutil.QuietCut(s, "/")
		return branch
	})
}

// RemoteInfos get by remote name
func (r *Repo) RemoteInfos(remote string) RemoteInfos {
	r.loadRemoteInfos()

	if len(r.remoteInfosMp) == 0 {
		return nil
	}
	return r.remoteInfosMp[remote]
}

// DefaultRemoteInfo get
func (r *Repo) DefaultRemoteInfo(typ ...string) *RemoteInfo {
	return r.RemoteInfo(r.cfg.DefaultRemote, typ...)
}

// FirstRemoteInfo get
func (r *Repo) FirstRemoteInfo(typ ...string) *RemoteInfo {
	return r.RandomRemoteInfo(typ...)
}

// RandomRemoteInfo get
func (r *Repo) RandomRemoteInfo(typ ...string) *RemoteInfo {
	r.loadRemoteInfos()

	if len(r.remoteNames) == 0 {
		return nil
	}
	return r.RemoteInfo(r.remoteNames[0], typ...)
}

// RemoteInfo get by remote name and type.
//
// - If remote is empty, will return default remote
// - If typ is empty, will return random type info.
//
// Usage:
//
//	ri := RemoteInfo("origin")
//	ri = RemoteInfo("origin", "push")
func (r *Repo) RemoteInfo(remote string, typ ...string) *RemoteInfo {
	riMp := r.RemoteInfos(strutil.OrElse(remote, r.cfg.DefaultRemote))
	if len(riMp) == 0 {
		return nil
	}

	if len(typ) > 0 {
		return riMp[typ[0]]
	}

	// get random type info
	for _, info := range riMp {
		return info
	}
	return nil // should never happen
}

// AllRemoteInfos get
func (r *Repo) AllRemoteInfos() map[string]RemoteInfos {
	return r.loadRemoteInfos().remoteInfosMp
}

// AllRemoteInfos get
func (r *Repo) loadRemoteInfos() *Repo {
	// has loaded
	if len(r.remoteNames) > 0 {
		return r
	}

	str, err := r.gw.Remote("-v").Output()
	if err != nil {
		r.setErr(err)
		return r
	}

	// origin  https://github.com/gookit/gitw.git (fetch)
	// origin  https://github.com/gookit/gitw.git (push)
	rmp := make(map[string]RemoteInfos, 2)
	str = strings.ReplaceAll(strings.TrimSpace(str), "\t", " ")

	names := make([]string, 0, 2)
	lines := strings.Split(str, "\n")

	for _, line := range lines {
		// origin https://github.com/gookit/gitw (push)
		ss := strutil.SplitN(line, " ", 3)
		if len(ss) < 3 {
			r.setErr(errorx.Rawf("invalid remote line: %s", line))
			continue
		}

		name, url, typ := ss[0], ss[1], ss[2]
		typ = strings.Trim(typ, "()")

		// create instance
		ri, err := NewRemoteInfo(name, url, typ)
		if err != nil {
			r.setErr(err)
			continue
		}

		rs, ok := rmp[name]
		if !ok {
			rs = make(RemoteInfos, 2)
		}

		// add
		rs[typ] = ri
		rmp[name] = rs
		if !arrutil.StringsHas(names, name) {
			names = append(names, name)
		}
	}

	if len(names) > 0 {
		r.remoteNames = names
		r.remoteInfosMp = rmp
	}
	return r
}

// reset last error
// func (r *Repo) resetErr() {
// 	r.err = nil
// }

// ReadConfig contents from REPO/.git/config
func (r *Repo) ReadConfig() []byte {
	return fsutil.GetContents(fsutil.JoinPaths(r.dir, GitDir, ConfFile))
}

// ReadHEAD contents from REPO/.git/HEAD
func (r *Repo) ReadHEAD() []byte {
	return fsutil.GetContents(fsutil.JoinPaths(r.dir, GitDir, HeadFile))
}

// -------------------------------------------------
// helper methods
// -------------------------------------------------

// reset last error
func (r *Repo) setErr(err error) {
	if err != nil {
		r.err = err
	}
}

// Err get last error
func (r *Repo) Err() error {
	return r.err
}

// Dir get repo dir
func (r *Repo) Dir() string {
	return r.dir
}

// Git get git wrapper
func (r *Repo) Git() *GitWrap {
	return r.gw
}

// Cmd new git command wrapper
func (r *Repo) Cmd(name string, args ...string) *GitWrap {
	return r.gw.Cmd(name, args...)
}

// QuickRun git command
func (r *Repo) QuickRun(cmd string, args ...string) error {
	return r.gw.Cmd(cmd, args...).Run()
}
