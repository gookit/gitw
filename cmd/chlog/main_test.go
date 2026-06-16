package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/gookit/gitw"
	"github.com/gookit/gitw/chlog"
	"github.com/gookit/goutil/x/assert"
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
