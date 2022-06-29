# Gitw

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/gitw?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/gitw)](https://github.com/gookit/gitw)
[![GoDoc](https://godoc.org/github.com/gookit/gitw?status.svg)](https://pkg.go.dev/github.com/gookit/gitw)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/gitw)](https://goreportcard.com/report/github.com/gookit/gitw)
[![Unit-Tests](https://github.com/gookit/gitw/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/gitw/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/gitw/badge.svg?branch=master)](https://coveralls.io/github/gookit/gitw?branch=master)

`gitw` - Git 命令包装器，生成 git 变更记录日志，获取 repo 信息和一些 git 命令工具。

> Github https://github.com/gookit/gitw

- 包装本地 git 命令
- 快速运行 git 命令
- 快速查询存储库信息
- 通过 git log 快速生成更改日志
  - 允许自定义生成配置，例如过滤、样式等

> **[EN-README](README.md)**

## 安装

> 需要: go 1.14+, git 2.x

```bash
go get github.com/gookit/gitw
```

## 使用

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

### 使用更多参数

示例，通过 `git log` 获取两个 sha 版本之间的提交日志

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

## 仓库信息

可以通过 gitw 在本地快速获取 git 存储库信息。

```go
repo := gitw.NewRepo("/path/to/my-repo")
```

### Branch 信息

```go
brInfo := repo.CurBranchInfo()

dump.Println(brInfo)
```

**Output**:

![one-remote-info](_examples/images/one-branch-info.png)

### Remote 信息

```go
rt := repo.DefaultRemoteInfo()

dump.Println(rt)
```

**Output**:

![one-remote-info](_examples/images/one-remote-info.png)

**仓库信息**:

```go
dump.Println(repo.Info())
```

**Output**:

![simple-repo-info](_examples/images/simple-repo-info.png)

## 生成变更日志

可以通过 `gitw/chlog` 包快速生成更新日志。

- 允许自定义生成配置 请看 [.github/changelog.yml](.github/changelog.yml)
- 可以设置过滤、分组、输出样式等

### 安装

```shell
go install github.com/gookit/gitw/cmd/chlog@latest
```

### 使用

运行 `chlog -h` 查看帮助:

![chlog-help](_examples/images/chlog-help.png)

**运行示例**:

```shell
chlog last head
chlog -c .github/changelog.yml last head
```

**Outputs**:

```text
## Change Log

### Update

- update: update some logic for git command run [96147fb](https://github.com/gookit/gitw/commit/96147fba43caf462a50bc97d7ed078dd0059e797)
- update: move RepoUrl config to MarkdownFormatter [8c861bf](https://github.com/gookit/gitw/commit/8c861bf05ae3576aba401692124df63372ae9ed7)

### Fixed

- fix: gen changelog error [1636761](https://github.com/gookit/gitw/commit/16367617bc364ce1022097e89313c7b09983981a)

### Other

- style: update some code logic [4a9f146](https://github.com/gookit/gitw/commit/4a9f14656b26a08b0cdd9c4f9cec9ae3bf5938b1)
- build(deps): bump github.com/gookit/color from 1.4.2 to 1.5.0 [037fa47](https://github.com/gookit/gitw/commit/037fa477954b630fe34ff7ceab51e6132db645cb)
- style: update examples and readme [8277389](https://github.com/gookit/gitw/commit/8277389817917e6b0cb97f3e5629f2c5034075e4)

```

### 在GitHub Action使用

`gitw/chlog` 可以直接在 GitHub Actions 中使用, 示例:

> Full script please see [.github/workflows/release.yml](.github/workflows/release.yml)

```yaml
# ...

    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Generate changelog
        run: |
          curl https://github.com/gookit/gitw/releases/latest/download/chlog-linux-amd64 -L -o /usr/local/bin/chlog
          chmod a+x /usr/local/bin/chlog
          chlog -c .github/changelog.yml -o changelog.md prev last 

```

### 在项目中使用

在项目代码中使用

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

### Commonly functions

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

## Refer

- https://github/phppkg/phpgit
- https://github.com/github/hub
- https://github.com/alibaba/git-repo-go

## LICENSE

[MIT](LICENSE)
