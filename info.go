package gitwrap

import "fmt"

const (
	ProtoSsh  = "ssh"
	ProtoHttp = "http"

	SchemeGit   = "git"
	SchemeHttp  = "https"
	SchemeHttps = "https"

	DefaultBranchName = "master"
	DefaultRemoteName = "origin"
)

// Info struct
// type Info struct {
// 	gw *GitWrap
// }

// RepoInfo struct
type RepoInfo struct {
	Name string
	Path string

	Dir string
	URL string
}

// remote type names
const (
	RemoteTypePush  = "push"
	RemoteTypeFetch = "fetch"
)

// RemoteInfos map. key is type name(see RemoteTypePush)
type RemoteInfos map[string]*RemoteInfo

// RemoteInfo struct
// - http: "https://github.com/gookit/gitwrap.git"
// - git: "git@github.com:gookit/gitwrap.git"
type RemoteInfo struct {
	// Name the repo remote name, default see DefaultRemoteName
	Name string
	// Type remote type. allow: push, fetch
	Type string
	// URL full git remote URL string.
	//
	// eg:
	// - http: "https://github.com/gookit/gitwrap.git"
	// - git: "git@github.com:gookit/gitwrap.git"
	URL string

	// ---- details

	// Scheme the url scheme. eg: git, http, https
	Scheme string
	// Host name. eg: "github.com"
	Host string
	// the group, repo name
	Group, Repo string

	// Proto the type 'ssh' OR 'http'
	Proto string
}

// NewRemoteInfo create
func NewRemoteInfo(name, url, typ string) (*RemoteInfo, error) {
	r := &RemoteInfo{
		Name: name,
		URL:  url,
		Type: typ,
	}

	err := ParseRemoteUrl(url, r)

	if err != nil {
		return nil, err
	}
	return r, nil
}

// NewEmptyRemoteInfo only with URL string.
func NewEmptyRemoteInfo(URL string) *RemoteInfo {
	return &RemoteInfo{
		Name: DefaultRemoteName,
		URL:  URL,
		Type: RemoteTypePush,
	}
}

// Valid check
func (r *RemoteInfo) Valid() bool {
	return r.URL != ""
}

// Invalid check
func (r *RemoteInfo) Invalid() bool {
	return r.URL == ""
}

// GitUrl build. eg: "git@github.com:gookit/gitwrap.git"
func (r *RemoteInfo) GitUrl() string {
	return SchemeGit + "@" + r.Host + ":" + r.Group + "/" + r.Repo + ".git"
}

func (r *RemoteInfo) HttpUrl() string {
	return SchemeHttp + "//" + r.Host + "/" + r.Group + "/" + r.Repo
}

// HttpsUrl build
func (r *RemoteInfo) HttpsUrl() string {
	return SchemeHttps + "//" + r.Host + "/" + r.Group + "/" + r.Repo
}

// Path string
func (r *RemoteInfo) Path() string {
	return r.RepoPath()
}

// RepoPath string
func (r *RemoteInfo) RepoPath() string {
	return r.Group + "/" + r.Repo
}

// String remote info to string.
func (r *RemoteInfo) String() string {
	return fmt.Sprintf("%s  %s (%s)", r.Name, r.URL, r.Type)
}
