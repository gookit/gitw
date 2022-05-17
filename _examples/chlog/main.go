package main

import (
	"fmt"

	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil"
)

// run: go run ./_examples/chlog
func main() {
	cl := chlog.New()
	cl.Formatter = &chlog.MarkdownFormatter{
		RepoURL: "https://github.com/gookit/gitw",
	}

	// with some settings ...
	cl.WithConfigFn(func(c *chlog.Config) {
		c.GroupPrefix = "\n### "
		c.GroupSuffix = "\n"
	})

	// fetch git log
	cl.FetchGitLog("v0.1.0", "HEAD", "--no-merges")

	// do generate
	goutil.PanicIfErr(cl.Generate())

	// dump
	fmt.Println(cl.Changelog())
}
