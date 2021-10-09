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

// RepoConfig struct
type RepoConfig struct {
	DefaultBranch string
	DefaultRemote string
}

// Repo struct
type Repo struct {
	gw *GitWrap
	// the repo dir
	dir string
	// config
	conf *RepoConfig
	// data cache
	cache map[string]interface{}
}

// NewRepo create Repo object
func NewRepo(dir string) *Repo {
	return &Repo{
		dir:   dir,
		cache: make(map[string]interface{}, 16),
		// init gw
		gw: New().WithWorkDir(dir),
	}
}

// Init run init for the repo dir.
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
	return r.gw
}
