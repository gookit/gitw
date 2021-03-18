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

// Repo struct
type Repo struct {
	// the repo dir
	dir string
}

func NewRepo(dir string) *Repo {
	return &Repo{dir: dir}
}

func (r *Repo) Init() {
	// TODO
}

func (r *Repo) Info()  {
	// TODO
}

func (r *Repo) Dir() string {
	return r.dir
}
