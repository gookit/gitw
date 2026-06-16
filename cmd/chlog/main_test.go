package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/gookit/gitw"
	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/goutil/x/ccolor"
)

func TestGenerateReturnsErrorOnMissingLastTag(t *testing.T) {
	workdir := initGitRepoWithoutTags(t)
	oldRepo, oldOpts, oldCfg := repo, opts, cfg
	t.Cleanup(func() {
		repo, opts, cfg = oldRepo, oldOpts, oldCfg
	})

	repo = gitw.NewRepo(workdir)
	opts.sha1 = gitw.TagLast
	opts.sha2 = gitw.TagHead
	cfg = chlog.NewDefaultConfig()

	err := generate(chlog.NewWithConfig(cfg))
	if !assert.Err(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "no git tags found")
}

func TestGeneratePrintsRefsWithShortHash(t *testing.T) {
	workdir := initGitRepoWithoutTags(t)
	runGit(t, workdir, "tag", "v0.1.0")
	runGit(t, workdir, "commit", "--allow-empty", "-m", "fix: second commit")
	chdir(t, workdir)

	oldRepo, oldOpts, oldCfg := repo, opts, cfg
	t.Cleanup(func() {
		repo, opts, cfg = oldRepo, oldOpts, oldCfg
	})

	repo = gitw.NewRepo(workdir)
	opts.sha1 = gitw.TagLast
	opts.sha2 = gitw.TagHead
	cfg = chlog.NewDefaultConfig()

	var out bytes.Buffer
	ccolor.SetOutput(&out)
	t.Cleanup(func() {
		ccolor.SetOutput(os.Stdout)
	})

	err := generate(chlog.NewWithConfig(cfg))
	assert.NoErr(t, err)

	tagHash := shortRev(t, workdir, "v0.1.0")
	headHash := shortRev(t, workdir, "HEAD")
	assert.Contains(t, out.String(), "Generate changelog: v0.1.0("+tagHash+") to HEAD("+headHash+")")
}

func TestGenerateAllIncludesRootCommit(t *testing.T) {
	workdir := initGitRepoWithoutTags(t)
	runGit(t, workdir, "commit", "--allow-empty", "-m", "fix: second commit")
	chdir(t, workdir)

	oldRepo, oldOpts, oldCfg := repo, opts, cfg
	t.Cleanup(func() {
		repo, opts, cfg = oldRepo, oldOpts, oldCfg
	})

	repo = gitw.NewRepo(workdir)
	opts.sha1 = "all"
	opts.sha2 = gitw.TagHead
	cfg = chlog.NewDefaultConfig()

	cl := chlog.NewWithConfig(cfg)
	err := generate(cl)
	assert.NoErr(t, err)
	assert.Contains(t, cl.Changelog(), "feat: initial commit")
	assert.Contains(t, cl.Changelog(), "fix: second commit")
}

func TestLoadConfigSkipsMissingRemote(t *testing.T) {
	workdir := initGitRepoWithoutTags(t)
	oldRepo, oldOpts, oldCfg := repo, opts, cfg
	t.Cleanup(func() {
		repo, opts, cfg = oldRepo, oldOpts, oldCfg
	})

	repo = gitw.NewRepo(workdir)
	opts.configFile = ""
	opts.style = ""
	opts.verbose = false
	cfg = chlog.NewDefaultConfig()

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("loadConfig should not panic without remote: %v", err)
		}
	}()
	loadConfig()
	assert.Eq(t, "", cfg.RepoURL)
}

func initGitRepoWithoutTags(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "tester@example.com")
	runGit(t, dir, "config", "user.name", "tester")

	err := os.WriteFile(dir+"/README.md", []byte("test"), 0644)
	assert.NoErr(t, err)

	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "feat: initial commit")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func shortRev(t *testing.T, dir, rev string) string {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--short", rev)
	cmd.Dir = dir
	out, err := cmd.Output()
	assert.NoErr(t, err)
	return string(bytes.TrimSpace(out))
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	wd, err := os.Getwd()
	assert.NoErr(t, err)
	assert.NoErr(t, os.Chdir(dir))
	t.Cleanup(func() {
		assert.NoErr(t, os.Chdir(wd))
	})
}
