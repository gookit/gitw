package gitw

import "fmt"

// remote type names
const (
	RemoteTypePush  = "push"
	RemoteTypeFetch = "fetch"
)

// RemoteInfos map. key is type name(see RemoteTypePush)
type RemoteInfos map[string]*RemoteInfo

// FetchInfo fetch remote info
func (rs RemoteInfos) FetchInfo() *RemoteInfo {
	return rs[RemoteTypeFetch]
}

// PushInfo push remote info
func (rs RemoteInfos) PushInfo() *RemoteInfo {
	return rs[RemoteTypePush]
}

// RemoteInfo struct
//
// - http: "https://github.com/gookit/gitw.git"
// - git: "git@github.com:gookit/gitw.git"
type RemoteInfo struct {
	// Name the repo remote name, default see DefaultRemoteName
	Name string
	// Type remote type. allow: push, fetch
	Type string
	// URL full git remote URL string.
	//
	// eg:
	// - http: "https://github.com/gookit/gitw.git"
	// - git: "git@github.com:gookit/gitw.git"
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

	if err := ParseRemoteURL(url, r); err != nil {
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

// GitURL build. eg: "git@github.com:gookit/gitw.git"
func (r *RemoteInfo) GitURL() string {
	return SchemeGIT + "@" + r.Host + ":" + r.Group + "/" + r.Repo + ".git"
}

// RawURLOfHTTP get remote url, if RemoteInfo.URL is git proto, build an HTTPS url.
func (r *RemoteInfo) RawURLOfHTTP() string {
	return r.URLOrBuild()
}

// URLOrBuild get remote HTTP url, if RemoteInfo.URL is git proto, build an HTTPS url.
func (r *RemoteInfo) URLOrBuild() string {
	if r.Proto == ProtoHTTP {
		return r.URL
	}
	return r.buildHTTPURL(true)
}

// URLOfHTTP build an HTTP url.
func (r *RemoteInfo) URLOfHTTP() string {
	return r.buildHTTPURL(false)
}

// URLOfHTTPS build an HTTPS url.
func (r *RemoteInfo) URLOfHTTPS() string {
	return r.buildHTTPURL(true)
}

// URLOfHTTPS build an HTTP(S) url.
func (r *RemoteInfo) buildHTTPURL(toHttps bool) string {
	schema := SchemeHTTP
	if toHttps {
		schema = SchemeHTTPS
	}

	return schema + "://" + r.Host + "/" + r.Group + "/" + r.Repo
}

// HTTPHost URL build. return like: https://github.com
func (r *RemoteInfo) HTTPHost() string {
	schema := r.Scheme
	if r.Proto != ProtoHTTP {
		schema = SchemeHTTPS
	}

	return schema + "://" + r.Host
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
