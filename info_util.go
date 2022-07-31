package gitw

import (
	"net/url"
	"strings"

	"github.com/gookit/goutil/errorx"
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
	if strings.HasPrefix(URL, "git@") {
		r.Proto = ProtoSSH
		if hasSfx {
			str = URL[4 : len(URL)-4]
		} else {
			str = URL[4:]
		}

		host, path, ok := strutil.Cut(str, ":")
		if !ok {
			return errorx.Rawf("invalid git URL: %s", URL)
		}

		group, repo, ok := strutil.Cut(path, "/")
		if !ok {
			return errorx.Rawf("invalid git URL path: %s", path)
		}

		r.Scheme = SchemeGIT
		r.Host, r.Group, r.Repo = host, group, repo
		return nil
	}

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

// ErrInvalidBrLine error
var ErrInvalidBrLine = errorx.Raw("invalid git branch line text")

// ParseBranchLine to BranchInfo data
//
// verbose:
// 	False - only branch name
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
