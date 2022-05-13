package gitwrap_test

import (
	"testing"

	"github.com/gookit/gitwrap"
	"github.com/gookit/goutil/dump"
	"github.com/stretchr/testify/assert"
)

func TestRepo_RemoteInfos(t *testing.T) {
	r := gitwrap.NewRepo("./")

	rs := r.AllRemoteInfos()
	assert.NoError(t, r.Err())
	assert.NotEmpty(t, rs)

	assert.True(t, r.HasRemote(gitwrap.DefaultRemoteName))
	assert.NotEmpty(t, r.RemoteNames())

	rt := r.DefaultRemoteInfo()
	assert.NotEmpty(t, rt)
	assert.True(t, rt.Valid())
	assert.False(t, rt.Invalid())
	assert.Equal(t, gitwrap.DefaultRemoteName, rt.Name)
	assert.Equal(t, "git@github.com:gookit/gitwrap.git", rt.GitURL())
	assert.Equal(t, "http://github.com/gookit/gitwrap", rt.URLOfHTTP())
	assert.Equal(t, "https://github.com/gookit/gitwrap", rt.URLOfHTTPS())

	rt = r.DefaultRemoteInfo(gitwrap.RemoteTypePush)
	assert.NotEmpty(t, rt)

	dump.P(r.Err(), rs)
}
