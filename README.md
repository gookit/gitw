# Gitw

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/gitw?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/gitw)](https://github.com/gookit/gitw)
[![Go Reference](https://pkg.go.dev/badge/github.com/gookit/goutil.svg)](https://pkg.go.dev/github.com/gookit/goutil)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/gitw)](https://goreportcard.com/report/github.com/gookit/gitw)
[![Unit-Tests](https://github.com/gookit/gitw/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/gitw/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/gitw/badge.svg?branch=main)](https://coveralls.io/github/gookit/gitw?branch=main)

`gitw` - Git command wrapper, generate git changelog, fetch repo information and some git tools.

- Wrap local `git` commands 
- Quickly run `git` commands 
- Quickly query repository information 
  - Quick fetch status, remote, branch ... details
- Quickly generate version changelogs via `git log` 
  - Allow custom build configuration 
  - Allow custom build filtering , styles, etc 
  - can be used directly in GitHub Actions
- Support git-emoji code search and replace render

> **[中文说明](README.zh-CN.md)**

## Install

> required: go 1.18+, git 2.x

```bash
go get github.com/gookit/gitw
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/gookit/gitw"
)

func main() {
	// logTxt, err := gitw.ShowLogs("v1.0.2", "v1.0.3")
	logTxt := gitw.MustString(gitw.ShowLogs("v1.0.2", "v1.0.3"))
	fmt.Println(logTxt)

	// Local Branches
	brList := gitw.MustStrings(gitw.Branches())
	fmt.Println(brList)

	// custom create command

	logCmd := gitw.New("log", "-2")
	// git.Run()
	// txt, err := logCmd.Output()
	txt := logCmd.SafeOutput()

	fmt.Println(txt)
}
```

### With more arguments

Examples, get commit logs between two sha versions via `git log`

```go
	logCmd := gitw.Log("--reverse").
		Argf("--pretty=format:\"%s\"", c.cfg.LogFormat)

	if c.cfg.Verbose {
		logCmd.OnBeforeExec(gitw.PrintCmdline)
	}

	// add custom args. eg: "--no-merges"
	logCmd.AddArgs("--no-merges")

	// logCmd.Argf("%s...%s", "v0.1.0", "HEAD")
	if sha1 != "" && sha2 != "" {
		logCmd.Argf("%s...%s", sha1, sha2)
	}

	fmt.Println(logCmd.SafeOutput())
```

## Repository

You can quickly get a git repository information at local.

```go
repo := gitw.NewRepo("/path/to/my-repo")
```

### Status Information

```go
si := repo.StatusInfo()

dump.Println(si)
```

**Output**:

![repo-status-info](_examples/images/repo-status-info.png)

### Branch Information

```go
brInfo := repo.CurBranchInfo()

dump.Println(brInfo)
```

**Output**:

![one-remote-info](_examples/images/one-branch-info.png)

### Remote Information

```go
rt := repo.DefaultRemoteInfo()

dump.Println(rt)
```

**Output**:

![one-remote-info](_examples/images/one-remote-info.png)

### Repo Information

```go
dump.Println(repo.Info())
```

**Output**:

![simple-repo-info](_examples/images/simple-repo-info.png)

## Changelog

You can quickly generate changelog by `gitw/chlog` package.

- Allows custom build configuration. see [.github/changelog.yml](.github/changelog.yml)
- can set filtering, grouping, output styles, etc.

### Install

```shell
go install github.com/gookit/gitw/cmd/chlog@latest
```

### Usage

Please run `chlog -h` to see help:

![chlog-help](_examples/images/chlog-help.png)

**Generate changelog**:

```shell
chlog prev last
chlog last head
chlog -c .github/changelog.yml last head
```

**Outputs**:

![chlog-demo](_examples/images/chlog-demo.png)

### Use on action

Can use `gitw/chlog` on GitHub actions. It does not depend on the Go environment,
just download the binary files of the corresponding system.

Example:

> Full script please see [.github/workflows/release.yml](.github/workflows/release.yml)

```yaml
# ...

    steps:
      - name: Checkout
        uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Generate changelog
        run: |
          curl https://github.com/gookit/gitw/releases/latest/download/chlog-linux-amd64 -L -o /usr/local/bin/chlog
          chmod a+x /usr/local/bin/chlog
          chlog -c .github/changelog.yml -o changelog.md prev last 

```

### Use in code

```go
package main

import (
	"fmt"

	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil"
)

func main() {
	cl := chlog.New()
	cl.Formatter = &chlog.MarkdownFormatter{
		RepoURL: "https://github.com/gookit/gitw",
	}
	cl.WithConfig(func(c *chlog.Config) {
		// some settings ...
		c.Title = "## Change Log"
	})

	// fetch git log
	cl.FetchGitLog("v0.1.0", "HEAD", "--no-merges")

	// do generate
	goutil.PanicIfErr(cl.Generate())

	// dump
	fmt.Println(cl.Changelog())
}
```

## Commands

### Methods in `GitWrap`

Commands of git, more please see [pkg.go.dev](https://pkg.go.dev/github.com/gookit/gitw#GitWrap)

```go
func (gw *GitWrap) Add(args ...string) *GitWrap
func (gw *GitWrap) Annotate(args ...string) *GitWrap
func (gw *GitWrap) Apply(args ...string) *GitWrap
func (gw *GitWrap) Bisect(args ...string) *GitWrap
func (gw *GitWrap) Blame(args ...string) *GitWrap
func (gw *GitWrap) Branch(args ...string) *GitWrap
func (gw *GitWrap) Checkout(args ...string) *GitWrap
func (gw *GitWrap) CherryPick(args ...string) *GitWrap
func (gw *GitWrap) Clean(args ...string) *GitWrap
func (gw *GitWrap) Clone(args ...string) *GitWrap
func (gw *GitWrap) Commit(args ...string) *GitWrap
func (gw *GitWrap) Config(args ...string) *GitWrap
func (gw *GitWrap) Describe(args ...string) *GitWrap
func (gw *GitWrap) Diff(args ...string) *GitWrap
func (gw *GitWrap) Fetch(args ...string) *GitWrap
func (gw *GitWrap) Init(args ...string) *GitWrap
func (gw *GitWrap) Log(args ...string) *GitWrap
func (gw *GitWrap) Merge(args ...string) *GitWrap
func (gw *GitWrap) Pull(args ...string) *GitWrap
func (gw *GitWrap) Push(args ...string) *GitWrap
func (gw *GitWrap) Rebase(args ...string) *GitWrap
func (gw *GitWrap) Reflog(args ...string) *GitWrap
func (gw *GitWrap) Remote(args ...string) *GitWrap
// and more ...
```

### Commonly git functions

Git command functions of std:

```go
func Alias(name string) string
func AllVars() string
func Branches() ([]string, error)
func CommentChar(text string) (string, error)
func Config(name string) string
func ConfigAll(name string) ([]string, error)
func DataDir() (string, error)
func EditText(data string) string
func Editor() string
func GlobalConfig(name string) (string, error)
func HasDotGitDir(path string) bool
func HasFile(segments ...string) bool
func Head() (string, error)
func Quiet(args ...string) bool
func Ref(ref string) (string, error)
func RefList(a, b string) ([]string, error)
func Remotes() ([]string, error)
func SetGlobalConfig(name, value string) error
func SetWorkdir(dir string)
func ShowDiff(sha string) (string, error)
func ShowLogs(sha1, sha2 string) (string, error)
func Spawn(args ...string) error
func SymbolicFullName(name string) (string, error)
func SymbolicRef(ref string) (string, error)
func Tags(args ...string) ([]string, error)
func Var(name string) string
func Version() (string, error)
func Workdir() (string, error)
func WorkdirName() (string, error)
```

### Util functions

```go
func SetDebug()
func SetDebug(open bool)
func IsDebug() bool
func IsGitCmd(command string) bool
func IsGitCommand(command string) bool
func IsGitDir(dir string) bool
func ParseRemoteURL(URL string, r *RemoteInfo) (err error)
func MustString(s string, err error) string
func MustStrings(ss []string, err error) []string
func FirstLine(output string) string
func OutputLines(output string) []string
func EditText(data string) string
```

### Remote info

```go
// type RemoteInfo
func NewEmptyRemoteInfo(URL string) *RemoteInfo
func NewRemoteInfo(name, url, typ string) (*RemoteInfo, error)
func (r *RemoteInfo) GitURL() string
func (r *RemoteInfo) Invalid() bool
func (r *RemoteInfo) Path() string
func (r *RemoteInfo) RawURLOfHTTP() string
func (r *RemoteInfo) RepoPath() string
func (r *RemoteInfo) String() string
func (r *RemoteInfo) URLOfHTTP() string
func (r *RemoteInfo) URLOfHTTPS() string
func (r *RemoteInfo) Valid() bool
```

## Refers

- https://github/phppkg/phpgit
- https://github.com/github/hub
- https://github.com/alibaba/git-repo-go
- https://github.com/carloscuesta/gitmoji/blob/master/packages/gitmojis/src/gitmojis.json
- https://github.com/hooj0/git-emoji-guide

## LICENSE

[MIT](LICENSE)
