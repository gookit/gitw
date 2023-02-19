package gitutil

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	cachedSSHCfg SSHConfig
	protocolReg  = regexp.MustCompile("^[a-zA-Z_+-]+://")
)

const githubHost = "github.com"

// URLParser struct
type URLParser struct {
	SSHConfig SSHConfig
}

// Parse parse raw url
func (p *URLParser) Parse(rawURL string) (u *url.URL, err error) {
	if !protocolReg.MatchString(rawURL) &&
		strings.Contains(rawURL, ":") &&
		// not a Windows path
		!strings.Contains(rawURL, "\\") {
		rawURL = "ssh://" + strings.Replace(rawURL, ":", "/", 1)
	}

	u, err = url.Parse(rawURL)
	if err != nil {
		return
	}

	if u.Scheme == "git+ssh" {
		u.Scheme = "ssh"
	}

	if u.Scheme != "ssh" {
		return
	}

	if strings.HasPrefix(u.Path, "//") {
		u.Path = strings.TrimPrefix(u.Path, "/")
	}

	if idx := strings.Index(u.Host, ":"); idx >= 0 {
		u.Host = u.Host[0:idx]
	}

	sshHost := p.SSHConfig[u.Host]
	// ignore replacing host that fixes for limited network
	// https://help.github.com/articles/using-ssh-over-the-https-port
	ignoredHost := u.Host == githubHost && sshHost == "ssh.github.com"
	if !ignoredHost && sshHost != "" {
		u.Host = sshHost
	}

	return
}

// ParseURL parse raw url
func ParseURL(rawURL string) (u *url.URL, err error) {
	if cachedSSHCfg == nil {
		cachedSSHCfg = newSSHConfigReader().Read()
	}

	p := &URLParser{cachedSSHCfg}
	return p.Parse(rawURL)
}
