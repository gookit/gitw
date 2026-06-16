package gitw

import (
	"net/url"
	"strings"

	"github.com/gookit/gitw/gitutil"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// ErrRemoteInfoNil error
var ErrRemoteInfoNil = errorx.Raw("the remote info data cannot be nil")

// ParseRemoteURL info to the RemoteInfo object.
func ParseRemoteURL(URL string, r *RemoteInfo) (err error) {
	if r == nil {
		return ErrRemoteInfoNil
	}

	var str string
	hasSfx := strings.HasSuffix(URL, ".git")

	// eg: "git@github.com:gookit/gitw.git"
	if gitutil.IsSSHProto(URL) {
		r.Proto = ProtoSSH
		r.Scheme = SchemeGIT

		if strings.HasPrefix(URL, "ssh://") && !isSCPStyleSSHURL(URL) {
			info, err := url.Parse(URL)
			if err != nil {
				return err
			}

			group, repo, err := splitRemotePath(info.Path)
			if err != nil {
				return err
			}

			r.Host, r.Group, r.Repo = info.Hostname(), group, repo
			r.Port = mathutil.SafeInt(info.Port())
			return nil
		}

		str = strings.TrimPrefix(strings.TrimPrefix(URL, "ssh://"), "git@")
		if hasSfx {
			str = str[0 : len(str)-4]
		}

		host, path, ok := strutil.Cut(str, ":")
		if !ok {
			return errorx.Rawf("invalid git URL: %s", URL)
		}

		nodes := strings.Split(path, "/")
		if len(nodes) > 2 && strutil.IsNumeric(nodes[0]) {
			r.Port = mathutil.SafeInt(nodes[0])
			path = strings.Join(nodes[1:], "/")
		}

		group, repo, err := splitRemotePath(path)
		if err != nil {
			return err
		}
		r.Host, r.Group, r.Repo = host, group, repo
		return nil
	}

	// http protocol
	str = URL
	if hasSfx {
		str = URL[0 : len(URL)-4]
	}

	// eg: "https://github.com/gookit/gitw.git"
	info, err := url.Parse(str)
	if err != nil {
		return err
	}

	group, repo, ok := strutil.Cut(strings.Trim(info.Path, "/"), "/")
	if !ok {
		return errorx.Rawf("invalid http URL path: %s", info.Path)
	}

	r.Proto = ProtoHTTP
	r.Scheme = info.Scheme
	r.Host, r.Group, r.Repo = info.Host, group, repo
	return nil
}

func splitRemotePath(path string) (group, repo string, err error) {
	path = strings.Trim(path, "/")
	if strings.HasSuffix(path, ".git") {
		path = path[0 : len(path)-4]
	}

	nodes := strings.Split(path, "/")
	if len(nodes) < 2 {
		return "", "", errorx.Rawf("invalid git URL path: %s", path)
	}

	return strings.Join(nodes[:len(nodes)-1], "/"), nodes[len(nodes)-1], nil
}

func isSCPStyleSSHURL(rawURL string) bool {
	str := strings.TrimPrefix(rawURL, "ssh://")
	firstColon := strings.IndexByte(str, ':')
	if firstColon < 0 {
		return false
	}

	firstSlash := strings.IndexByte(str, '/')
	if firstSlash >= 0 && firstSlash < firstColon {
		return false
	}

	pathPart := str[firstColon+1:]
	port, _, _ := strings.Cut(pathPart, "/")
	return !strutil.IsNumeric(port)
}

// ErrInvalidBrLine error
var ErrInvalidBrLine = errorx.Raw("invalid git branch line text")

// ParseBranchLine to BranchInfo data
//
// verbose:
//
//	False - only branch name
//	True  - get by `git br -v --all`
//	        format: * BRANCH_NAME  COMMIT_ID  COMMIT_MSG
func ParseBranchLine(line string, verbose bool) (*BranchInfo, error) {
	info := &BranchInfo{}
	line = strings.TrimSpace(line)

	if strings.HasPrefix(line, "*") {
		info.Current = true
		line = strings.Trim(line, "*\t ")
	}

	if line == "" {
		return nil, ErrInvalidBrLine
	}

	// at tag head. eg: `* （头指针在 v0.2.3 分离） 3c08adf chore: update readme add branch info docs`
	if strings.HasPrefix(line, "(") || strings.HasPrefix(line, "（") {
		return nil, ErrInvalidBrLine
	}

	if !verbose {
		info.SetName(line)
		return info, nil
	}

	// parse name
	nodes := strutil.SplitNTrimmed(line, " ", 2)
	if len(nodes) != 2 {
		return nil, ErrInvalidBrLine
	}

	info.SetName(nodes[0])

	// parse hash and message
	nodes = strutil.SplitNTrimmed(nodes[1], " ", 2)
	if len(nodes) != 2 {
		return nil, ErrInvalidBrLine
	}

	info.Hash, info.HashMsg = nodes[0], nodes[1]
	return info, nil
}

func isVerboseBranchLine(line string) bool {
	line = strings.Trim(line, " *\t\n\r\x0B")
	return strings.ContainsRune(line, ' ')
}
