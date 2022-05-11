package main

import (
	"flag"
	"fmt"

	"github.com/gookit/gitwrap/chlog"
	"github.com/gookit/goutil"
)

var opts = struct {
	noMerges                 bool
	sha1Version, sha2Version string
	configFile               string
	outputFile               string
}{}

func parseFlags() {
	flag.BoolVar(&opts.noMerges, "no-merge", false, "don't collect merge commits")
	flag.StringVar(&opts.sha1Version, "sha1", "", "the old git sha version. allow: tag name, commit id")
	flag.StringVar(&opts.sha2Version, "sha2", "", "the new git sha version. allow: tag name, commit id")
	flag.StringVar(&opts.configFile, "config", "", "the config file for generate changelog")
	flag.StringVar(&opts.outputFile, "output", "stdout", "the output file for generated changelog")

	flag.Parse()
}

func main() {
	parseFlags()

	cl := chlog.New()
	// with some settings ...
	cl.WithConfig(func(c *chlog.Changelog) {
		c.GroupPrefix = "\n### "
		c.GroupSuffix = "\n"
		c.Formatter = &chlog.MarkdownFormatter{
			RepoURL: "https://github.com/gookit/gitwrap",
		}
	})

	// fetch git log
	var gitArgs []string
	if opts.noMerges {
		gitArgs = append(gitArgs, "--no-merges")
	}

	cl.FetchGitLog(opts.sha1Version, opts.sha2Version, gitArgs...)

	// do generate
	goutil.PanicIfErr(cl.Generate())

	// dump
	fmt.Println(cl.Changelog())
}
