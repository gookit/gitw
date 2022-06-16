package gitw

// some consts for remote info
const (
	ProtoSSH  = "ssh"
	ProtoHTTP = "http"

	SchemeGIT   = "git"
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
)

// RepoInfo struct
type RepoInfo struct {
	Name string
	Path string
	Dir  string
	URL  string
	// LastCID value
	LastCID string
	Branch  string
	Version string
}
