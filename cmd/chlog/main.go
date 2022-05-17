package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gookit/color"
	"github.com/gookit/gitw"
	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
	"gopkg.in/yaml.v3"
)

var opts = struct {
	verbose bool
	// with git merges log
	withMerges bool

	sha1, sha2 string
	excludes   string
	configFile string
	outputFile string
}{}

func parseFlags() error {
	flag.BoolVar(&opts.verbose, "verbose", false, "show more information")
	flag.BoolVar(&opts.withMerges, "with-merge", false, "collect git merge commits")
	flag.StringVar(&opts.configFile, "config", "", "the YAML config file for generate changelog")
	flag.StringVar(&opts.outputFile, "output", "stdout", "the output file for generated changelog")
	flag.StringVar(&opts.excludes, "exclude", "", "exclude commit by keywords, multi split by comma")

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
		if err := yaml.Unmarshal(yml, cfg); err != nil {
			panic(err)
		}
	}

	if opts.verbose {
		cfg.Verbose = true
		color.Cyanln("Changelog Config:")
		dump.NoLoc(cfg)
	}

	// with some settings ...
	if len(opts.excludes) > 0 {
		cfg.Filters = append(cfg.Filters, maputil.Data{
			"name":     chlog.FilterKeywords,
			"keywords": opts.excludes,
			"exclude":  "true",
		})
	}

	// create
	cl := chlog.NewWithConfig(cfg)

	// generate
	err := generate(cl)
	if err != nil {
		color.Errorln("Generate error: ", err)
		return
	}

	// dump
	outputTo(cl, opts.outputFile)
}

func generate(cl *chlog.Changelog) error {

	// fetch git log
	var gitArgs []string
	if !opts.withMerges {
		gitArgs = append(gitArgs, "--no-merges")
	}

	sha1, sha2 := matchShaVal(opts.sha1), matchShaVal(opts.sha2)
	color.Infof("Generate changelog: %s to %s\n", sha1, sha2)

	cl.FetchGitLog(sha1, sha2, gitArgs...)

	// do generate
	return cl.Generate()
}

var repo = gitw.NewRepo("./")

func matchShaVal(sha string) string {
	name := strings.ToLower(sha)
	if name == "last" {
		return repo.LargestTag()
	}

	if name == "prev" {
		return repo.PrevMaxTag()
	}

	if name == "head" {
		return "HEAD"
	}

	return sha
}

func outputTo(cl *chlog.Changelog, outFile string) {
	if outFile == "stdout" {
		fmt.Println(cl.Changelog())
		return
	}

	f, err := fsutil.QuickOpenFile(outFile)
	if err != nil {
		color.Errorln("open the output file error:", err)
		return
	}

	defer f.Close()
	_, err = cl.WriteTo(f)
	if err != nil {
		color.Errorln("write to output file error:", err)
		return
	}

	color.Success.Println("OK. Changelog written to:", outFile)
}

func showHelp(err error) {
	buf := new(bytes.Buffer)
	if err != nil {
		buf.WriteString("ERROR: " + err.Error())
		buf.WriteByte('\n')
	} else {
		buf.WriteString(color.Cyan.Render("Quick generate change log from git logs\n"))
	}

	binName := path.Base(os.Args[0])
	buf.WriteByte('\n')
	buf.WriteString("Usage: " + binName + " [-options] sha1 sha2\n")
	buf.WriteString(color.Comment.Render("Arguments:\n"))
	buf.WriteString("  sha1 	  The old git sha version. allow: tag name, commit id\n")
	buf.WriteString("  sha2 	  The new git sha version. allow: tag name, commit id\n")
	buf.WriteString(color.Comment.Render("Options:"))
	fmt.Println(buf.String())
	flag.PrintDefaults()
	fmt.Println(color.Comment.Render("Examples:"))
	fmt.Printf("  %s v0.1.0 HEAD\n", binName)
	fmt.Printf("  %s prev last\n", binName)
	fmt.Printf("  %s -exclude 'action tests,script error' prev last\n", binName)
	fmt.Printf("  %s -config .github/changelog.yml last HEAD\n", binName)
	fmt.Printf("  %s -config .github/changelog.yml -output changelog.md last HEAD\n", binName)
}
