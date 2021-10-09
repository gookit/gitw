# GitWrap

Git command wrapper, and some extra git tools package.

> Github https://github.com/gookit/gitwrap

## Install

> required: go 1.14+, git 2.x

```bash
go get github.com/gookit/gitwrap
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/gookit/gitwrap"
)

func main() {
	// logTxt, err := gitwrap.Log("v1.0.2", "v1.0.3")
	logTxt := gitwrap.MustString(gitwrap.Log("v1.0.2", "v1.0.3"))
	fmt.Println(logTxt)

	// LocalBranches
	brList := gitwrap.MustStrings(gitwrap.LocalBranches())
	fmt.Println(brList)
	
	// custom create command

	logCmd := gitwrap.New("log", "-2")
	// git.Run()
	// txt, err := logCmd.Output()
	txt := logCmd.SafeOutput()

	fmt.Println(txt)
}
```

## Functions

```go
func Alias(name string) string
func CommentChar(text string) (string, error)
func Config(name string) string
func ConfigAll(name string) ([]string, error)
func Dir() (string, error)
func Editor() string
func GlobalConfig(name string) (string, error)
func HasFile(segments ...string) bool
func Head() (string, error)
func IsGitCommand(command string) bool
func IsGitDir(dir string) bool
func LocalBranches() ([]string, error)
func Log(sha1, sha2 string) (string, error)
func ParseURL(rawURL string) (u *url.URL, err error)
func Quiet(args ...string) bool
func Ref(ref string) (string, error)
func RefList(a, b string) ([]string, error)
func Remotes() ([]string, error)
func Run(args ...string) error
func SetDebug()
func SetGlobalConfig(name, value string) error
func Show(sha string) (string, error)
func Spawn(args ...string) error
func SymbolicFullName(name string) (string, error)
func SymbolicRef(ref string) (string, error)
func Var(name string) string
func Version() (string, error)
func WorkdirName() (string, error)
```

Util functions:

```go
func MustString(s string, err error) string
func MustStrings(ss []string, err error) []string
func EditText(data string) string
```

## Refer

- https://github/phppkg/phpgit
- https://github.com/github/hub
- https://github.com/alibaba/git-repo-go
