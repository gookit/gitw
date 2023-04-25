package gitw_test

import (
	"strings"
	"testing"

	"github.com/gookit/gitw"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/testutil/assert"
)

func TestNewRemoteInfo(t *testing.T) {
	URL := "https://github.com/gookit/gitw"

	rt, err := gitw.NewRemoteInfo("origin", URL, gitw.RemoteTypePush)
	assert.NoErr(t, err)
	assert.True(t, rt.Valid())
	assert.False(t, rt.Invalid())
	assert.Eq(t, "origin", rt.Name)
	assert.Eq(t, gitw.RemoteTypePush, rt.Type)
	assert.Eq(t, "github.com", rt.Host)
	assert.Eq(t, "gookit/gitw", rt.RepoPath())
	assert.Eq(t, gitw.SchemeHTTPS, rt.Scheme)
	assert.Eq(t, gitw.ProtoHTTP, rt.Proto)
	assert.Eq(t, rt.URL, rt.RawURLOfHTTP())

	URL = "git@github.com:gookit/gitw.git"
	rt, err = gitw.NewRemoteInfo("origin", URL, gitw.RemoteTypePush)
	assert.NoErr(t, err)
	assert.Eq(t, "github.com", rt.Host)
	assert.Eq(t, "gookit/gitw", rt.Path())
	assert.Eq(t, gitw.SchemeGIT, rt.Scheme)
	assert.Eq(t, gitw.ProtoSSH, rt.Proto)
	assert.Eq(t, "https://github.com/gookit/gitw", rt.RawURLOfHTTP())
}

func TestParseBranchLine_simple(t *testing.T) {
	info, err := gitw.ParseBranchLine("* ", false)
	assert.Err(t, err)

	info, err = gitw.ParseBranchLine("* (HEAD)", false)
	assert.Err(t, err)

	info, err = gitw.ParseBranchLine("* fea/new_br001", false)
	assert.NoErr(t, err)

	assert.True(t, info.Current)
	assert.True(t, info.IsValid())
	assert.False(t, info.IsRemoted())
	assert.Eq(t, "", info.Remote)
	assert.Eq(t, "fea/new_br001", info.Name)
	assert.Eq(t, "fea/new_br001", info.Short)

	info, err = gitw.ParseBranchLine("  remotes/source/my_new_br ", false)
	assert.NoErr(t, err)

	assert.False(t, info.Current)
	assert.True(t, info.IsValid())
	assert.True(t, info.IsRemoted())
	assert.Eq(t, "source", info.Remote)
	assert.Eq(t, "remotes/source/my_new_br", info.Name)
	assert.Eq(t, "my_new_br", info.Short)
}

func TestParseBranchLine_verbose(t *testing.T) {
	info, err := gitw.ParseBranchLine("* fea/new_br001              73j824d the message 001", true)
	assert.NoErr(t, err)

	assert.True(t, info.Current)
	assert.True(t, info.IsValid())
	assert.False(t, info.IsRemoted())
	assert.Eq(t, "", info.Remote)
	assert.Eq(t, "fea/new_br001", info.Name)
	assert.Eq(t, "fea/new_br001", info.Short)
	assert.Eq(t, "73j824d", info.Hash)
	assert.Eq(t, "the message 001", info.HashMsg)

	info, err = gitw.ParseBranchLine("  remotes/source/my_new_br   6fb8dcd the message 003 ", true)
	assert.NoErr(t, err)
	dump.P(info)

	assert.False(t, info.Current)
	assert.True(t, info.IsValid())
	assert.True(t, info.IsRemoted())
	assert.Eq(t, "source", info.Remote)
	assert.Eq(t, "remotes/source/my_new_br", info.Name)
	assert.Eq(t, "my_new_br", info.Short)
	assert.Eq(t, "6fb8dcd", info.Hash)
	assert.Eq(t, "the message 003", info.HashMsg)

	info, err = gitw.ParseBranchLine("* （头指针在 v0.2.3 分离） 3c08adf chore: update readme add branch info docs", true)
	assert.Err(t, err)
	info, err = gitw.ParseBranchLine("* (HEAD detached at pull/29/merge)                                    62f3455 Merge cfc79b748e176c1c9e266c8bc413c87fe974acef into c9503c2aef993a2cf582d90c137deda53c9bca68", true)
	assert.Err(t, err)
}

func TestBranchInfo_parse_simple(t *testing.T) {
	gitOut := `
  fea/new_br001
* master
  my_new_br 
  remotes/origin/my_new_br 
  remotes/source/my_new_br 
`
	bis := gitw.NewBranchInfos(gitOut)
	bis.Parse()
	// dump.P(bis)

	assert.NoErr(t, bis.LastErr())
	assert.NotEmpty(t, bis.Current())
	assert.NotEmpty(t, bis.Locales())
	assert.NotEmpty(t, bis.Remotes(""))
	assert.Eq(t, "master", bis.Current().Name)
}

func TestBranchInfo_parse_invalid(t *testing.T) {
	gitOut := `
  fea/new_br001
* (HEAD)
  my_new_br 
  remotes/origin/my_new_br 
`
	bis := gitw.NewBranchInfos(gitOut)
	bis.Parse()
	// dump.P(bis)

	assert.Err(t, bis.LastErr())
	assert.Nil(t, bis.Current())
	assert.NotEmpty(t, bis.Locales())
	assert.NotEmpty(t, bis.Remotes("origin"))
}

func TestBranchInfo_parse_verbose(t *testing.T) {
	gitOut := `
  fea/new_br001              73j824d the message 001
* master                     7r60d4f the message 002
  my_new_br                  6fb8dcd the message 003
  remotes/origin/my_new_br   6fb8dcd the message 003
  remotes/source/my_new_br   6fb8dcd the message 003
`

	bis := gitw.EmptyBranchInfos()
	bis.SetBrLines(strings.Split(strings.TrimSpace(gitOut), "\n"))
	bis.Parse()
	// dump.P(bis)

	assert.NoErr(t, bis.LastErr())
	assert.NotEmpty(t, bis.Current())
	assert.NotEmpty(t, bis.Locales())
	assert.NotEmpty(t, bis.Remotes(""))
	assert.Eq(t, "master", bis.Current().Name)

	// search
	rets := bis.Search("new", gitw.BrSearchLocal)
	assert.NotEmpty(t, rets)
	assert.Len(t, rets, 2)

	// search
	rets = bis.Search("new", gitw.BrSearchRemote)
	assert.NotEmpty(t, rets)
	assert.Len(t, rets, 2)

	// search
	rets = bis.Search("origin:new", gitw.BrSearchRemote)
	assert.NotEmpty(t, rets)
	assert.Len(t, rets, 1)
	assert.True(t, rets[0].IsRemoted())
	assert.Eq(t, "origin", rets[0].Remote)
}

func TestStatusInfo_FromLines(t *testing.T) {
	text := `
## master...origin/fea/master
 RM app/Common/GitLocal/GitFlow.php -> app/Common/GitLocal/GitFactory.php
  M app/Common/GitLocal/GitHub.php
 ?? app/Common/GitLocal/GitConst.php
  D tmp/delete-some.file
`
	si := gitw.NewStatusInfo(text)

	dump.P(si)
	assert.Eq(t, "master", si.Branch)
	assert.Eq(t, "origin", si.UpRemote)
	assert.Eq(t, "fea/master", si.UpBranch)
	assert.False(t, si.IsCleaned())
	assert.Gt(t, si.FileNum(), 2)
}
