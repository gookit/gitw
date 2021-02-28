# GitCmd

Git command wrapper library

> project is from https://github.com/github/hub

## Install

```bash
go get github.com/gookit/gitcmd
```

## Usage

```go
git := gitcmd.NewGit("log", "-2")
git.Run()
```
