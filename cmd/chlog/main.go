package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/strutil"
	"gopkg.in/yaml.v3"
)

var opts = struct {
	noMerges   bool
	sha1, sha2 string
	configFile string
	outputFile string
}{}

func parseFlags() error {
	flag.BoolVar(&opts.noMerges, "no-merge", false, "don't collect merge commits")
	flag.StringVar(&opts.configFile, "config", "", "the YAML config file for generate changelog")
	flag.StringVar(&opts.outputFile, "output", "stdout", "the output file for generated changelog")

	flag.Usage = func() {
		showHelp(nil)
	}
	flag.Parse()

	args := flag.Args()
	aln := len(args)
	if aln == 0 {
		return errorx.Rawf("arguments sha1, sha2 is required")
	}

	if aln == 2 {
		opts.sha1, opts.sha2 = args[0], args[1]
	} else if strings.Contains(args[0], "...") {
		opts.sha1, opts.sha2 = strutil.MustCut(args[0], "...")
	}

	// check again
	if opts.sha1 == "" || opts.sha2 == "" {
		return errorx.Rawf("arguments: sha1, sha2 both is required")
	}

	return nil
}

var cfg = chlog.NewDefaultConfig()

// run: go run ./cmd/chlog
func main() {
	if err := parseFlags(); err != nil {
		showHelp(err)
		return
	}

	yml := fsutil.ReadExistFile(opts.configFile)
	if len(yml) > 0 {
		if err := yaml.Unmarshal([]byte(yml), cfg); err != nil {
			panic(err)
		}
	}

	dump.P(cfg)
	cl := chlog.NewWithConfig(cfg)
	cl.Formatter = &chlog.MarkdownFormatter{
		RepoURL: cfg.RepoURL,
	}
	// with some settings ...
	cl.WithConfigFn(func(c *chlog.Config) {
		c.GroupPrefix = "\n### "
		c.GroupSuffix = "\n"
	})

	// fetch git log
	var gitArgs []string
	if opts.noMerges {
		gitArgs = append(gitArgs, "--no-merges")
	}

	cl.FetchGitLog(opts.sha1, opts.sha2, gitArgs...)

	// do generate
	goutil.PanicIfErr(cl.Generate())

	// dump
	fmt.Println(cl.Changelog())
}

func showHelp(err error) {
	buf := new(bytes.Buffer)
	if err != nil {
		buf.WriteString("ERROR: " + err.Error())
		buf.WriteByte('\n')
		buf.WriteByte('\n')
	} else {
		buf.WriteString("Quick generate change log from git logs\n")
	}

	buf.WriteString("Usage: " + os.Args[0] + " [-options] sha1 sha2\n")
	buf.WriteString("Arguments:\n")
	buf.WriteString("  sha1 	  The old git sha version. allow: tag name, commit id\n")
	buf.WriteString("  sha2 	  The new git sha version. allow: tag name, commit id\n")
	buf.WriteString("Options:")
	fmt.Println(buf.String())
	flag.PrintDefaults()
	fmt.Println("Examples:")
	fmt.Println("  chlog v0.1.0 HEAD")
}
