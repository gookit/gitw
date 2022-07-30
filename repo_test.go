package gitw_test

import (
	"testing"

	"github.com/gookit/gitw"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/sysutil"
	"github.com/gookit/slog"
	"github.com/stretchr/testify/assert"
)

var repo = gitw.NewRepo("./").WithFn(func(r *gitw.Repo) {
	r.Git().BeforeExec = gitw.PrintCmdline
})

func TestMain(m *testing.M) {
	slog.Println("workdir", sysutil.Workdir())
	m.Run()
}

func TestRepo_RemoteInfos(t *testing.T) {
	rs := repo.AllRemoteInfos()
	dump.P(rs)

	assert.NoError(t, repo.Err())
	assert.NotEmpty(t, rs)

	assert.True(t, repo.HasRemote(gitw.DefaultRemoteName))
	assert.NotEmpty(t, repo.RemoteNames())
}

func TestRepo_DefaultRemoteInfo(t *testing.T) {
	rt := repo.DefaultRemoteInfo()
	dump.P(rt)

	assert.NotEmpty(t, rt)
	assert.True(t, rt.Valid())
	assert.False(t, rt.Invalid())
	assert.Equal(t, gitw.DefaultRemoteName, rt.Name)
	assert.Equal(t, "git@github.com:gookit/gitw.git", rt.GitURL())
	assert.Equal(t, "http://github.com/gookit/gitw", rt.URLOfHTTP())
	assert.Equal(t, "https://github.com/gookit/gitw", rt.URLOfHTTPS())

	rt = repo.RandomRemoteInfo(gitw.RemoteTypePush)
	assert.NotEmpty(t, rt)
}

func TestRepo_AutoMatchTag(t *testing.T) {
	assert.Equal(t, "HEAD", repo.AutoMatchTag("head"))
	assert.Equal(t, "541fb9d", repo.AutoMatchTag("541fb9d"))
}

func TestRepo_BranchInfos(t *testing.T) {
	bs := repo.BranchInfos()
	assert.NotEmpty(t, bs)
	dump.P(bs.BrLines())

	assert.NotEmpty(t, repo.SearchBranches("main", gitw.BrSearchAll))

	cur := repo.CurBranchInfo()
	if cur != nil {
		assert.NotEmpty(t, cur)
		assert.Equal(t, "main", cur.Name)
		assert.NotEmpty(t, repo.BranchInfo("main"))
	}
}

func TestRepo_Info(t *testing.T) {
	info := repo.Info()
	dump.P(info)

	assert.Nil(t, repo.Err())
	assert.NotNil(t, info)
	assert.Equal(t, "gitw", info.Name)
}

func TestRepo_AutoMatchTagByTagType(t *testing.T) {
	assert.Equal(t, "HEAD", repo.AutoMatchTagByTagType("head", 0))
	assert.Equal(t, "541fb9d", repo.AutoMatchTagByTagType("541fb9d", 0))
}

func TestRepo_TagsSortedByCreatordate(t *testing.T) {
	tags := repo.TagsSortedByCreatordate()
	dump.P(tags)
	assert.NotEmpty(t, tags)
}

func TestRepo_TagByDescribe(t *testing.T) {
	tags := repo.TagByDescribe("")
	dump.P(tags)
	assert.NotEmpty(t, tags)
}
