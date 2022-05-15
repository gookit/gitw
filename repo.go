package gitw

import (
	"path"
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
)

const (
	cacheRemoteNames   = "rmtNames"
	cacheRemoteInfos   = "rmtInfos"
	cacheLastCommitID  = "lastCID"
	cacheCurrentBranch = "curBranch"
	cacheMaxTagVersion = "maxVersion"
)

// CmdBuilder struct
// type CmdBuilder struct {
// 	Dir string
// }

// RepoConfig struct
type RepoConfig struct {
	DefaultBranch string
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
	// remoteNames
	remoteNames []string
	// remoteInfosMp
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
	rt := r.loadRemoteInfos().DefaultRemoteInfo()
	if rt == nil {
		return nil
	}

	return &RepoInfo{
		Name: rt.Repo,
		Path: rt.Path(),
		Dir:  r.dir,
		URL:  rt.RawURLOfHTTP(),
		// more
		Branch:  r.CurrentBranch(),
		Version: r.TagLargest(),
		LastCID: r.LastAbbrevID(),
	}
}

// CurrentBranch return current branch name
func (r *Repo) CurrentBranch() string {
	brName := r.cache.Str(cacheCurrentBranch)
	if len(brName) > 0 {
		return brName
	}

	// cat .git/HEAD
	// OR
	// git symbolic-ref HEAD // out: refs/heads/fea_pref
	str, err := r.Cmd("symbolic-ref", "HEAD").Output()
	if err != nil {
		r.setErr(err)
		return ""
	}

	// eg: fea_pref
	brName = path.Base(FirstLine(str))
	r.cache.Set(cacheCurrentBranch, brName)

	return brName
}

// TagMax get max tag version of the repo
func (r *Repo) TagMax() string {
	return r.TagLargest()
}

// TagLargest get max tag version of the repo
func (r *Repo) TagLargest() string {
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

// TagSecondMax get second-largest tag of the repo
func (r *Repo) TagSecondMax() string {
	tags := r.TagsSortedByRefName()

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

	return OutputLines(str)
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

// RandomRemoteInfo get
func (r *Repo) RandomRemoteInfo(typ ...string) *RemoteInfo {
	r.loadRemoteInfos()

	if len(r.remoteNames) == 0 {
		return nil
	}
	return r.RemoteInfo(r.remoteNames[0], typ...)
}

// RemoteInfo get.
// if typ is empty, will return random type info.
func (r *Repo) RemoteInfo(remote string, typ ...string) *RemoteInfo {
	riMp := r.RemoteInfos(remote)
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
	dump.P(lines)
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

	dump.P(names, rmp)
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

// -------------------------------------------------
// helper methods
// -------------------------------------------------

// reset last error
func (r *Repo) setErr(err error) {
	if err != nil {
		r.err = nil
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
