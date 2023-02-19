package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/gookit/color"
	"github.com/gookit/gitw"
	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
)

// Version number
var Version = "1.0.2"

var opts = struct {
	verbose bool
	// with git merges log
	withMerges bool

	workdir    string
	excludes   string
	configFile string

	outputFile string
	sha1, sha2 string

	style   string
	tagType int
}{}

var cfg = chlog.NewDefaultConfig()
var repo = gitw.NewRepo("./")
var cmd = cflag.New(func(c *cflag.CFlags) {
	c.Version = Version
	c.Desc = "Quick generate change log from git logs"
})

// quick run:
//
//	go run ./cmd/chlog
//	go run ./cmd/chlog -h
//
// install to GOPATH/bin:
//
//	go install ./cmd/chlog
func main() {
	configCmd()

	cmd.MustParse(nil)
}

func configCmd() {
	cmd.BoolVar(&opts.verbose, "verbose", false, "show more information;;v")
	cmd.BoolVar(&opts.withMerges, "with-merge", false, "collect git merge commits")
	cmd.StringVar(&opts.workdir, "workdir", "", "workdir for run, default is current workdir")
	cmd.StringVar(&opts.configFile, "config", "", "the YAML config file for generate changelog;;c")
	cmd.StringVar(&opts.outputFile, "output", "stdout", "the output file for generated changelog;;o")
	cmd.StringVar(&opts.excludes, "exclude", "", "exclude commit by keywords, multi split by comma")
	cmd.StringVar(&opts.style, "style", "", "the output contents format style\nallow: simple, markdown(mkdown,md), ghr(gh-release.gh);;s")
	cmd.IntVar(&opts.tagType, "tag-type", 0, `get git tag name by tag type.
Allowed:
0 ref-name sort(<cyan>default</>)
1 creator date sort
2 describe command;;t`)

	cmd.AddArg("sha1", "The old git sha version. allow: tag name, commit id", true, nil)
	cmd.AddArg("sha2", "The new git sha version. allow: tag name, commit id", false, nil)

	cmd.Func = handle
	cmd.Example = `
  {{cmd}} v0.1.0 HEAD
  {{cmd}} prev last
  {{cmd}} prev...last
  {{cmd}} --exclude 'action tests,script error' prev last
  {{cmd}} -c .github/changelog.yml last HEAD
  {{cmd}} -c .github/changelog.yml -o changelog.md last HEAD
`
}

func checkInput(c *cflag.CFlags) error {
	opts.sha1 = c.Arg("sha1").String()
	opts.sha2 = c.Arg("sha2").String()

	if strings.Contains(opts.sha1, "...") {
		opts.sha1, opts.sha2 = strutil.MustCut(opts.sha1, "...")
	}

	// check again
	if opts.sha2 == "" {
		return errorx.Rawf("arguments: sha1, sha2 both is required")
	}

	if opts.workdir != "" {
		cliutil.Infoln("try change workdir to", opts.workdir)
		return os.Chdir(opts.workdir)
	}

	return nil
}

func handle(c *cflag.CFlags) error {
	if err := checkInput(c); err != nil {
		return err
	}

	// load config
	loadConfig()

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
		return err
	}

	// dump change logs to file
	outputTo(cl, opts.outputFile)
	return nil
}

func loadConfig() {
	yml := fsutil.ReadExistFile(opts.configFile)
	if len(yml) > 0 {
		if err := yaml.Unmarshal(yml, cfg); err != nil {
			panic(err)
		}
	}

	if cfg.RepoURL == "" {
		cfg.RepoURL = repo.DefaultRemoteInfo().URLOfHTTPS()
	}

	if opts.style != "" {
		cfg.Style = opts.style
	}

	if opts.verbose {
		cfg.Verbose = true
		cliutil.Cyanln("Changelog Config:")
		dump.NoLoc(cfg)
		fmt.Println()
	}
}

func generate(cl *chlog.Changelog) error {
	// fetch git logs
	var gitArgs []string
	if !opts.withMerges {
		gitArgs = append(gitArgs, "--no-merges")
	}

	sha1 := repo.AutoMatchTagByType(opts.sha1, opts.tagType)
	sha2 := repo.AutoMatchTagByType(opts.sha2, opts.tagType)
	cliutil.Infof("Generate changelog: %s to %s\n", sha1, sha2)

	cl.FetchGitLog(sha1, sha2, gitArgs...)

	// do generate
	return cl.Generate()
}

func outputTo(cl *chlog.Changelog, outFile string) {
	if outFile == "stdout" {
		fmt.Println(cl.Changelog())
		return
	}

	f, err := fsutil.QuickOpenFile(outFile)
	if err != nil {
		cliutil.Errorln("open the output file error:", err)
		return
	}

	defer f.Close()
	_, err = cl.WriteTo(f)
	if err != nil {
		cliutil.Errorln("write to output file error:", err)
		return
	}

	color.Success.Println("OK. Changelog written to:", outFile)
}
