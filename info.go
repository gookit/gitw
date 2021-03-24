package gitwrap

const (
	ProtoSsh  = "ssh"
	ProtoHttp = "http"

	DefaultRemoteName = "origin"
)

// RemoteInfo struct
// - http: "https://github.com/gookit/gitwrap.git"
// - git: "git@github.com:gookit/gitwrap.git"
type RemoteInfo struct {
	// the repo remote name and URL address
	Name, URL string

	// ---- details

	// Scheme the url scheme. eg: git, http, https
	Scheme string
	// Host name. eg: "github.com"
	Host string
	// the group, repo name
	Group, Repo string
	// Type string

	// ProtoType 'ssh' OR 'http'
	ProtoType string
}

func NewRemoteInfo(name, url string) *RemoteInfo {
	r := &RemoteInfo{
		Name: name,
		URL: url,
	}

	parseRemoteUrl(url, r)

	return r
}

func (r *RemoteInfo) Valid() bool {
	return r.URL != ""
}

func (r *RemoteInfo) Invalid() bool {
	return r.URL == ""
}

func (r *RemoteInfo) Path() string {
	return r.Group + "/" + r.Repo
}

func (r *RemoteInfo) String() string {
	return r.Name + " " + r.URL
}
