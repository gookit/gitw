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
	// LastHash last commit hash value
	LastHash string
	Branch   string
	Version  string
	// Upstream remote name
	Upstream string
	// Remotes name and url mapping.
	Remotes map[string]string
}
