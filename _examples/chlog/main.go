package main

import (
	"fmt"

	"github.com/gookit/gitwrap/chlog"
	"github.com/gookit/goutil"
)

func main() {
	cl := chlog.New()
	cl.WithConfig(func(c *chlog.Changelog) {
		// some settings ...
		c.Formatter = &chlog.MarkdownFormatter{
			RepoURL: "https://github.com/gookit/gitwrap",
		}
	})

	// fetch git log
	cl.FetchGitLog("v0.1.0", "HEAD", "--no-merges")

	// do generate
	goutil.PanicIfErr(cl.Generate())

	// dump
	fmt.Println(cl.Changelog())
}
