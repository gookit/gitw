package gitwrap_test

import (
	"testing"

	"github.com/gookit/gitwrap"
	"github.com/gookit/goutil/dump"
	"github.com/stretchr/testify/assert"
)

var repo = gitwrap.NewRepo("./").WithFn(func(r *gitwrap.Repo) {
	r.Git().BeforeExec = gitwrap.PrintCmdline
})

func TestRepo_Info(t *testing.T) {
	info := repo.Info()
	dump.P(info)

	assert.Nil(t, repo.Err())
	assert.NotNil(t, info)
	assert.Equal(t, "gitwrap", info.Name)
}

func TestRepo_RemoteInfos(t *testing.T) {
	rs := repo.AllRemoteInfos()
	dump.P(rs)

	assert.NoError(t, repo.Err())
	assert.NotEmpty(t, rs)

	assert.True(t, repo.HasRemote(gitwrap.DefaultRemoteName))
	assert.NotEmpty(t, repo.RemoteNames())

	rt := repo.DefaultRemoteInfo()
	dump.P(rt)
	assert.NotEmpty(t, rt)
	assert.True(t, rt.Valid())
	assert.False(t, rt.Invalid())
	assert.Equal(t, gitwrap.DefaultRemoteName, rt.Name)
	assert.Equal(t, "git@github.com:gookit/gitwrap.git", rt.GitURL())
	assert.Equal(t, "http://github.com/gookit/gitwrap", rt.URLOfHTTP())
	assert.Equal(t, "https://github.com/gookit/gitwrap", rt.URLOfHTTPS())

	rt = repo.DefaultRemoteInfo(gitwrap.RemoteTypePush)
	assert.NotEmpty(t, rt)
}
