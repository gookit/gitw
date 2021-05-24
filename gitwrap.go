// Package gitwrap is library warp git commands. code is refer from github/hub
package gitwrap

import "os"

var debug = isDebugFromEnv()

// SetDebug mode
func SetDebug() {
	debug = true
}

func isDebugFromEnv() bool {
	return os.Getenv("GIT_CMD_VERBOSE") != ""
}

// CmdBuilder struct
type CmdBuilder struct {
	Dir string
}

type RepoConfig struct {

	DefaultBranch string
	DefaultRemote string
}

// Repo struct
type Repo struct {
	gw GitWrap
	// the repo dir
	dir string
	// config
	conf *RepoConfig
	// data cache
	cache map[string]interface{}
}

func NewRepo(dir string) *Repo {
	return &Repo{
		dir:   dir,
		cache: make(map[string]interface{}, 16),
	}
}

func (r *Repo) Init() error {
	return r.gw.SubCmd("init").Run()
}

func (r *Repo) Info() {
	// TODO
}

func (r *Repo) RemoteInfos() {
	// TODO
}

func (r *Repo) DefaultRemoteInfo() *RemoteInfo {
	// TODO
	return nil
}

func (r *Repo) RemoteInfo(name string) *RemoteInfo {
	// TODO
	return nil
}

func (r *Repo) Dir() string {
	return r.dir
}

func (r *Repo) Git() *GitWrap {
	return New().WithWorkDir(r.dir)
}
